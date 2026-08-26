package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

	// The update payload omits duration_in_days when it is empty, so an existing
	// alert can only be switched to the full month by recreating it.
	s.Attributes["duration_in_days"] = schema.StringAttribute{
		Required:            true,
		Description:         attrs["duration_in_days"].GetDescription(),
		MarkdownDescription: attrs["duration_in_days"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplaceIf(
				func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
					resp.RequiresReplace = req.ConfigValue.ValueString() == "" && req.StateValue.ValueString() != ""
				},
				"Recreates the alert when duration_in_days is cleared to cover the full month.",
				"Recreates the alert when `duration_in_days` is cleared to cover the full month.",
			),
		},
	}

	s.Attributes["period_to_track"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         attrs["period_to_track"].GetDescription(),
		MarkdownDescription: attrs["period_to_track"].GetMarkdownDescription(),
		Validators: []validator.String{
			stringvalidator.OneOf("start_of_the_month", "end_of_the_month"),
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
		handleError("Read Budget Alert", &resp.Diagnostics, err)
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
	if _, err := r.client.V2.BudgetAlerts.DeleteBudgetAlert(params, r.client.Auth); err != nil {
		handleError("Delete Budget Alert", &resp.Diagnostics, err)
	}
}

func (r *budgetAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}
