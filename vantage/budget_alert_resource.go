package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget_alert"
	budgetalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/budget_alerts"
)

var (
	_ resource.Resource                = (*budgetAlertResource)(nil)
	_ resource.ResourceWithConfigure   = (*budgetAlertResource)(nil)
	_ resource.ResourceWithImportState = (*budgetAlertResource)(nil)
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

	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		Description:         attrs["token"].GetDescription(),
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// The API types duration_in_days as a string on write and as an integer on
	// read, so the generated string attribute is dropped and the field declared
	// here as an integer instead. Leaving it out tracks the full month.
	//
	// The update body omits an empty duration, so the API cannot move an alert
	// back to the full month. Clearing the field replaces the alert instead.
	durationDescription := "The number of days from the start or end of the month to trigger the alert if the threshold is reached. Omit to track the full month. Changing this to or from an omitted value replaces the alert."
	s.Attributes["duration_in_days"] = schema.Int64Attribute{
		Optional:            true,
		Description:         durationDescription,
		MarkdownDescription: durationDescription,
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

	params := budgetalertsv2.NewCreateBudgetAlertParams().WithCreateBudgetAlert(input)
	out, err := r.client.V2.BudgetAlerts.CreateBudgetAlert(params, r.client.Auth)
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
		WithBudgetAlertToken(data.Token.ValueString()).
		WithUpdateBudgetAlert(input)

	out, err := r.client.V2.BudgetAlerts.UpdateBudgetAlert(params, r.client.Auth)
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
