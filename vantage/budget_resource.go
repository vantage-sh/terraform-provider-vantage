package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget"
	budgetsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/budgets"
)

var (
	_ resource.Resource                = (*budgetResource)(nil)
	_ resource.ResourceWithConfigure   = (*budgetResource)(nil)
	_ resource.ResourceWithImportState = (*budgetResource)(nil)
)

func NewBudgetResource() resource.Resource {
	return &budgetResource{}
}

type budgetResource struct {
	client *Client
}

func (r *budgetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}
func (r *budgetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_budget"
}

func (r *budgetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_budget.BudgetResourceSchema(ctx)
	attrs := s.GetAttributes()
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		Description:         attrs["token"].GetDescription(),
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	resp.Schema = s
}

func (r *budgetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data budgetModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the planned periods value to preserve empty lists
	plannedPeriods := data.Periods

	params := budgetsv2.NewCreateBudgetParams().WithCreateBudget(toCreateModel(ctx, &resp.Diagnostics, data))
	out, err := r.client.V2.Budgets.CreateBudget(params, r.client.Auth)

	if err != nil {
		if e, ok := err.(*budgetsv2.CreateBudgetBadRequest); ok {
			handleBadRequest("Create Budget", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Budget", &resp.Diagnostics, err)
		return
	}

	tflog.Debug(ctx, "applyBudgetPayload create")
	diag := applyBudgetPayload(ctx, false, out.Payload, &data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// If the plan had an explicit empty list for periods, preserve it
	// This prevents inconsistent state when the API returns default periods
	if !plannedPeriods.IsNull() && !plannedPeriods.IsUnknown() && len(plannedPeriods.Elements()) == 0 {
		data.Periods = plannedPeriods
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data budgetModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the prior state's periods value to preserve empty lists
	priorPeriods := data.Periods

	fBool := false

	params := budgetsv2.NewGetBudgetParams().WithBudgetToken(data.Token.ValueString()).WithIncludePerformance(&fBool)
	out, err := r.client.V2.Budgets.GetBudget(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*budgetsv2.GetBudgetNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Get Budget", &resp.Diagnostics, err)
		return
	}
	tflog.Debug(ctx, "applyBudgetPayload read")
	diag := applyBudgetPayload(ctx, false, out.Payload, &data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// If the prior state had an explicit empty list for periods, preserve it
	// This prevents inconsistent state when the API returns default periods
	if !priorPeriods.IsNull() && !priorPeriods.IsUnknown() && len(priorPeriods.Elements()) == 0 {
		data.Periods = priorPeriods
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *budgetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data budgetModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the planned periods value to preserve empty lists
	plannedPeriods := data.Periods

	params := budgetsv2.NewUpdateBudgetParams().WithUpdateBudget(toUpdateModel(ctx, &resp.Diagnostics, data)).WithBudgetToken(data.Token.ValueString())
	out, err := r.client.V2.Budgets.UpdateBudget(params, r.client.Auth)

	if err != nil {
		if e, ok := err.(*budgetsv2.UpdateBudgetBadRequest); ok {
			handleBadRequest("Update Budget", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update Budget", &resp.Diagnostics, err)
		return
	}
	tflog.Debug(ctx, "applyBudgetPayload update")
	diag := applyBudgetPayload(ctx, false, out.Payload, &data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// If the plan had an explicit empty list for periods, preserve it
	// This prevents inconsistent state when the API returns default periods
	if !plannedPeriods.IsNull() && !plannedPeriods.IsUnknown() && len(plannedPeriods.Elements()) == 0 {
		data.Periods = plannedPeriods
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data budgetModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetsv2.NewDeleteBudgetParams().WithBudgetToken(data.Token.ValueString())
	_, err := r.client.V2.Budgets.DeleteBudget(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*budgetsv2.DeleteBudgetNotFound); ok {
			handleBadRequest("Delete Budget", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Delete Budget", &resp.Diagnostics, err)
		return
	}

}
