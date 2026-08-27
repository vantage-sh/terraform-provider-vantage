package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget"
	budgetsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/budgets"
)

var (
	_ resource.Resource                   = (*budgetResource)(nil)
	_ resource.ResourceWithConfigure      = (*budgetResource)(nil)
	_ resource.ResourceWithImportState    = (*budgetResource)(nil)
	_ resource.ResourceWithValidateConfig = (*budgetResource)(nil)
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
	s.Attributes["period_cadence"] = schema.SingleNestedAttribute{
		Optional:            true,
		Computed:            true,
		Description:         "The interval cadence for standard Budget periods. Requires the flexible_budget_periods feature. Changing a configured cadence replaces the Budget; removing the block stops managing it but does not clear the API cadence.",
		MarkdownDescription: "The interval cadence for standard Budget periods. Requires the `flexible_budget_periods` feature. Changing a configured cadence replaces the Budget; removing the block stops managing it but does not clear the API cadence.",
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplaceIfConfigured(),
		},
		Attributes: map[string]schema.Attribute{
			"starts_at": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The required anchor date for configured budget period intervals. ISO 8601 date (YYYY-MM-DD).",
				MarkdownDescription: "The required anchor date for configured budget period intervals. ISO 8601 date (`YYYY-MM-DD`).",
			},
			"interval_count": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The number of interval units per budget period.",
				MarkdownDescription: "The number of interval units per budget period.",
			},
			"interval_unit": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The unit for budget period intervals. One of: day, week, month, year.",
				MarkdownDescription: "The unit for budget period intervals. One of: day, week, month, year.",
				Validators: []validator.String{
					stringvalidator.OneOf("day", "week", "month", "year"),
				},
			},
		},
	}
	resp.Schema = s
}

func (r *budgetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config budgetModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateBudgetConfig(config, &resp.Diagnostics)
}

func validateBudgetConfig(config budgetModel, diagnostics *diag.Diagnostics) {
	if config.PeriodCadence.IsNull() || config.PeriodCadence.IsUnknown() {
		return
	}

	startsAt, ok := config.PeriodCadence.Attributes()["starts_at"].(types.String)
	if !ok {
		diagnostics.AddAttributeError(
			path.Root("period_cadence").AtName("starts_at"),
			"Invalid Budget Period Cadence",
			"period_cadence.starts_at must be a string.",
		)
		return
	}
	if !startsAt.IsUnknown() && (startsAt.IsNull() || startsAt.ValueString() == "") {
		diagnostics.AddAttributeError(
			path.Root("period_cadence").AtName("starts_at"),
			"Missing Budget Period Cadence Start",
			"period_cadence.starts_at must be set to a non-empty ISO 8601 date when period_cadence is configured.",
		)
	}

	if !config.ChildBudgetTokens.IsNull() &&
		!config.ChildBudgetTokens.IsUnknown() &&
		len(config.ChildBudgetTokens.Elements()) > 0 {
		diagnostics.AddAttributeError(
			path.Root("period_cadence"),
			"Period Cadence Is Not Supported for Compound Budgets",
			"period_cadence cannot be configured together with child_budget_tokens. Period cadence is only supported for standard Budgets.",
		)
	}
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

	// Save the current state periods value to preserve empty lists
	statePeriods := data.Periods

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

	// If the previous state had an explicit empty list for periods, preserve it
	// The API returns default periods even when empty was specified
	if !statePeriods.IsNull() && !statePeriods.IsUnknown() && len(statePeriods.Elements()) == 0 {
		data.Periods = statePeriods
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *budgetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *budgetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data budgetModel
	var config budgetModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the planned periods value to preserve empty lists
	plannedPeriods := data.Periods

	params := budgetsv2.NewUpdateBudgetParams().WithUpdateBudget(toUpdateModel(ctx, &resp.Diagnostics, data, config.PeriodCadence)).WithBudgetToken(data.Token.ValueString())
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

	// Save data into Terraform state
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
