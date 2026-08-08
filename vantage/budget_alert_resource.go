package vantage

import (
	"context"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget_alert"
	budgetalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/budget_alerts"
)

var (
	_ resource.Resource                = (*budgetAlertResource)(nil)
	_ resource.ResourceWithConfigure   = (*budgetAlertResource)(nil)
	_ resource.ResourceWithImportState = (*budgetAlertResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*budgetAlertResource)(nil)
)

func NewBudgetAlertResource() resource.Resource {
	return &budgetAlertResource{}
}

type budgetAlertResource struct {
	client *Client
}

func (r *budgetAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *budgetAlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_budget_alert"
}

func (r *budgetAlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_budget_alert.BudgetAlertResourceSchema(ctx)
	attrs := s.GetAttributes()

	// These values are fixed once the alert exists. Without a plan modifier
	// Terraform plans every computed attribute as "known after apply" on any
	// change, which buries the one attribute that really changes.
	for _, name := range []string{"created_at", "id", "token", "user_token"} {
		s.Attributes[name] = schema.StringAttribute{
			Computed:            true,
			Description:         attrs[name].GetDescription(),
			MarkdownDescription: attrs[name].GetMarkdownDescription(),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		}
	}

	// The API defaults period_to_track and keeps the value afterwards, so the
	// prior value stands when the configuration does not set one.
	s.Attributes["period_to_track"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         attrs["period_to_track"].GetDescription(),
		MarkdownDescription: attrs["period_to_track"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// The API returns these tokens in its own order, so they are sets rather
	// than the generated lists. A list would make Terraform compare positions
	// and reject the applied value as inconsistent with the plan.
	s.Attributes["budget_tokens"] = schema.SetAttribute{
		ElementType:         types.StringType,
		Required:            true,
		Description:         attrs["budget_tokens"].GetDescription(),
		MarkdownDescription: attrs["budget_tokens"].GetMarkdownDescription(),
	}

	// The API accepts an empty recipient_channels and clears the channels with
	// it. A configuration that names no channels keeps whatever the alert holds,
	// so clearing them means setting the attribute to an empty set.
	s.Attributes["recipient_channels"] = schema.SetAttribute{
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Description:         attrs["recipient_channels"].GetDescription(),
		MarkdownDescription: attrs["recipient_channels"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
	}

	// user_tokens behaves differently. The API rejects both an empty array and a
	// null with "user_tokens is empty", so an update can never take the last
	// user off an alert. Emptying the set replaces the alert instead.
	userTokensDescription := attrs["user_tokens"].GetDescription() +
		" Emptying this set replaces the alert, because the API cannot remove the last user from an existing one."
	s.Attributes["user_tokens"] = schema.SetAttribute{
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Description:         userTokensDescription,
		MarkdownDescription: userTokensDescription,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
			setplanmodifier.RequiresReplaceIf(
				func(_ context.Context, req planmodifier.SetRequest, resp *setplanmodifier.RequiresReplaceIfFuncResponse) {
					resp.RequiresReplace = budgetAlertUserTokensCleared(req.ConfigValue, req.StateValue)
				},
				"Replaces the alert when user_tokens becomes empty.",
				"Replaces the alert when `user_tokens` becomes empty.",
			),
		},
	}

	// The API types duration_in_days as a string on write and as an integer on
	// read, so the generated string attribute is dropped and the field declared
	// here as an integer instead. Leaving it out tracks the full month.
	//
	// The update body omits an empty duration, so the API cannot move an alert
	// back to the full month. Clearing the field replaces the alert instead.
	//
	// The validator keeps zero out of the configuration. The API reports a
	// full month as an absent duration, which reads back as null, so a
	// configured zero would never match the value the apply produces.
	durationDescription := "The number of days from the start or end of the month to trigger the alert if the threshold is reached, between 1 and 31. Omit to track the full month. Changing this to or from an omitted value replaces the alert."
	s.Attributes["duration_in_days"] = schema.Int64Attribute{
		Optional:            true,
		Description:         durationDescription,
		MarkdownDescription: durationDescription,
		Validators: []validator.Int64{
			int64validator.Between(1, 31),
		},
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.RequiresReplaceIf(
				func(_ context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
					resp.RequiresReplace = req.ConfigValue.IsNull() && !req.StateValue.IsNull()
				},
				"Replaces the alert when duration_in_days is removed from the configuration.",
				"Replaces the alert when `duration_in_days` is removed from the configuration.",
			),
		},
	}

	resp.Schema = s
}

// budgetAlertUserTokensCleared reports whether a plan takes the last user off an
// existing alert. Only an explicit empty set does that. A set the configuration
// leaves out keeps the prior value instead.
func budgetAlertUserTokensCleared(configValue, stateValue types.Set) bool {
	if configValue.IsNull() || configValue.IsUnknown() {
		return false
	}
	if stateValue.IsNull() || stateValue.IsUnknown() {
		return false
	}
	return len(configValue.Elements()) == 0 && len(stateValue.Elements()) > 0
}

// budgetAlertDerivedAttributes lists the computed attributes the API derives
// from another attribute of the same alert, each paired with its source.
var budgetAlertDerivedAttributes = []struct {
	attribute string
	source    string
}{
	// The workspace of an alert is the workspace of the budgets it watches.
	{attribute: "workspace_token", source: "budget_tokens"},
	// The integration is the one behind the channels the alert posts to.
	{attribute: "integration_provider", source: "recipient_channels"},
}

// ModifyPlan keeps a derived value in the plan while the attribute it comes
// from is unchanged. Without this, an unrelated edit plans them as "known after
// apply" and hides the change the practitioner asked for.
//
// This runs here rather than in an attribute plan modifier because attribute
// modifiers run in schema order. integration_provider is modified before
// recipient_channels, so it would read a source that is still unknown.
func (r *budgetAlertResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// A create has no prior value to keep, and a destroy has no plan to fill.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	for _, derived := range budgetAlertDerivedAttributes {
		attribute := path.Root(derived.attribute)
		source := path.Root(derived.source)

		var planned types.String
		resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, attribute, &planned)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Something already decided this value.
		if !planned.IsUnknown() {
			continue
		}

		var plannedSource, priorSource types.Set
		resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, source, &plannedSource)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, source, &priorSource)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// The source is still being decided, or it moved. Either way the API
		// settles this value.
		if plannedSource.IsUnknown() || !plannedSource.Equal(priorSource) {
			continue
		}

		var prior types.String
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, attribute, &prior)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, attribute, prior)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *budgetAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *budgetAlertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetalertsv2.NewCreateBudgetAlertParams()
	out, err := r.client.V2.BudgetAlerts.CreateBudgetAlert(params, r.client.Auth, withBudgetAlertBody(input))
	if err != nil {
		if e, ok := err.(*budgetalertsv2.CreateBudgetAlertBadRequest); ok {
			handleBadRequest("Create Budget Alert", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Budget Alert", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *budgetAlertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetalertsv2.NewGetBudgetAlertParams().WithBudgetAlertToken(data.Token.ValueString())
	out, err := r.client.V2.BudgetAlerts.GetBudgetAlert(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*budgetalertsv2.GetBudgetAlertNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Get Budget Alert", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *budgetAlertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetalertsv2.NewUpdateBudgetAlertParams().
		WithBudgetAlertToken(data.Token.ValueString())

	out, err := r.client.V2.BudgetAlerts.UpdateBudgetAlert(params, r.client.Auth, withBudgetAlertBody(input))
	if err != nil {
		handleError("Update Budget Alert", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *budgetAlertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetalertsv2.NewDeleteBudgetAlertParams().WithBudgetAlertToken(data.Token.ValueString())
	_, err := r.client.V2.BudgetAlerts.DeleteBudgetAlert(params, r.client.Auth)
	if err != nil {
		handleError("Delete Budget Alert", &resp.Diagnostics, err)
	}
}

func (r *budgetAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

// withBudgetAlertBody sends the given value as the request body, in place of the
// generated model.
//
// The generated create and update models tag every array without omitempty, so
// they can only send an empty array or a null. The API rejects both for
// user_tokens and needs the key to be absent, so the body is written from a type
// that can leave a field out. Everything else about the call stays generated.
func withBudgetAlertBody(body any) budgetalertsv2.ClientOption {
	return func(op *runtime.ClientOperation) {
		params := op.Params
		op.Params = runtime.ClientRequestWriterFunc(func(req runtime.ClientRequest, reg strfmt.Registry) error {
			if params != nil {
				if err := params.WriteToRequest(req, reg); err != nil {
					return err
				}
			}
			return req.SetBodyParam(body)
		})
	}
}
