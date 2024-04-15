package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resource_budget "github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

// let the budget model defined in the resource be the common model for
// both the data source and the resource
type budgetModel resource_budget.BudgetModel

type budgetPerformanceModel struct {
	Actual types.String `tfsdk:"actual"`
	Amount types.String `tfsdk:"amount"`
	Date   types.String `tfsdk:"budget"`
}

type budgetPeriodModel struct {
	Amount  types.String `tfsdk:"amount"`
	EndAt   types.String `tfsdk:"end_at"`
	StartAt types.String `tfsdk:"start_at"`
}

// type budgetResourceModel struct {}
func applyBudgetPayload(ctx context.Context, src *modelsv2.Budget, dst *budgetModel) diag.Diagnostics {
	dst.Token = types.StringValue(src.Token)
	dst.CreatedAt = types.StringValue(src.CreatedAt)
	dst.Name = types.StringValue(src.Name)
	dst.UserToken = types.StringValue(src.UserToken)
	dst.WorkspaceToken = types.StringValue(src.WorkspaceToken)
	dst.CostReportToken = types.StringValue(src.CostReportToken)

	if src.BudgetAlertTokens != nil {
		budgetAlertTokens, diag := types.ListValueFrom(ctx, types.StringType, src.BudgetAlertTokens)
		if diag.HasError() {
			return diag
		}
		dst.BudgetAlertTokens = budgetAlertTokens
	}

	if src.Performance != nil {
		perfs := make([]budgetPerformanceModel, 0, len(src.Performance))
		for _, p := range src.Performance {
			performance := budgetPerformanceModel{
				Actual: types.StringValue(p.Actual),
				Amount: types.StringValue(p.Amount),
				Date:   types.StringValue(p.Date),
			}
			perfs = append(perfs, performance)
		}

		l, d := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: map[string]attr.Type{
				"actual": types.StringType,
				"amount": types.StringType,
				"date":   types.StringType,
			}},
			perfs,
		)
		if d.HasError() {
			return d
		}
		dst.Performance = l
	}

	if src.Periods != nil {
		periods := make([]budgetPeriodModel, 0, len(src.Periods))
		for _, p := range src.Periods {
			period := budgetPeriodModel{
				Amount:  types.StringValue(p.Amount),
				EndAt:   types.StringValue(p.EndAt),
				StartAt: types.StringValue(p.StartAt),
			}
			periods = append(periods, period)
		}

		l, d := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: map[string]attr.Type{
				"amount": types.StringType,
				"end":    types.StringType,
				"start":  types.StringType,
			}},
			periods,
		)
		if d.HasError() {
			return d
		}
		dst.Periods = l
	}
	return nil
}
