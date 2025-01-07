package vantage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	resource_budget "github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

// let the budget, budget performance and budget period models
// defined in the resource_budget package be the common models for
// both the data source and the resource
type budgetModel resource_budget.BudgetModel
type budgetPerformanceModel resource_budget.PerformanceValue
type budgetPeriodResourceModel struct {
	Amount  types.Float64 `tfsdk:"amount"`
	EndAt   types.String  `tfsdk:"end_at"`
	StartAt types.String  `tfsdk:"start_at"`
}

type budgetPeriodDataSourceModel struct {
	Amount  types.String `tfsdk:"amount"`
	EndAt   types.String `tfsdk:"end_at"`
	StartAt types.String `tfsdk:"start_at"`
}

// toCreateModel and toUpdateModel can be further refactored.
func toCreateModel(ctx context.Context, diags *diag.Diagnostics, src budgetModel) *modelsv2.CreateBudget {
	dst := &modelsv2.CreateBudget{
		Name:            src.Name.ValueStringPointer(),
		CostReportToken: src.CostReportToken.ValueString(),
		WorkspaceToken:  src.WorkspaceToken.ValueString(),
	}

    if !src.ChildBudgetTokens.IsNull() && !src.ChildBudgetTokens.IsUnknown() {
        childBudgetTokens := []string{}
        src.ChildBudgetTokens.ElementsAs(ctx, &childBudgetTokens, false)
        dst.ChildBudgetTokens = childBudgetTokens
    }

	if !src.Periods.IsNull() && !src.Periods.IsUnknown() {
		periods := make([]*budgetPeriodResourceModel, 0, len(src.Periods.Elements()))
		if diag := src.Periods.ElementsAs(ctx, &periods, false); diag.HasError() {
			diags.Append(diag...)
			return nil
		}

		dstValues := make([]*modelsv2.CreateBudgetPeriodsItems0, 0, len(periods))
		for _, p := range periods {
			periodItem := &modelsv2.CreateBudgetPeriodsItems0{
				Amount: p.Amount.ValueFloat64Pointer(),
			}

			if p.EndAt.String() != "<unknown>" {
				endAt, err := time.Parse("2006-01-02", strings.ReplaceAll(p.EndAt.String(), "\"", ""))
				if err != nil {
					diags.AddError("parsing error", fmt.Sprintf("failed to parse end_at: %s", err))
					return nil
				}

				ea := strfmt.Date(endAt)
				periodItem.EndAt = &ea
			}

			startAt, err := time.Parse("2006-01-02", strings.ReplaceAll(p.StartAt.String(), "\"", ""))
			if err != nil {
				diags.AddError("parsing error", fmt.Sprintf("failed to parse start_at: %s", err))
				return nil
			}

			if !startAt.IsZero() {
				sa := strfmt.Date(startAt)
				periodItem.StartAt = &sa
			}
			dstValues = append(dstValues, periodItem)
		}
		dst.Periods = dstValues
	}
	return dst
}

func toUpdateModel(ctx context.Context, diags *diag.Diagnostics, src budgetModel) *modelsv2.UpdateBudget {
	dst := &modelsv2.UpdateBudget{
		Name:            src.Name.ValueString(),
		CostReportToken: src.CostReportToken.ValueString(),
	}

	if !src.ChildBudgetTokens.IsNull() && !src.ChildBudgetTokens.IsUnknown() {
		childBudgetTokens := []string{}
		src.ChildBudgetTokens.ElementsAs(ctx, &childBudgetTokens, false)
		dst.ChildBudgetTokens = childBudgetTokens
	}

	if !src.Periods.IsNull() && !src.Periods.IsUnknown() {
		periods := make([]*budgetPeriodResourceModel, 0, len(src.Periods.Elements()))
		if diag := src.Periods.ElementsAs(ctx, &periods, false); diag.HasError() {
			diags.Append(diag...)
			return nil
		}

		dstValues := make([]*modelsv2.UpdateBudgetPeriodsItems0, 0, len(periods))
		for _, p := range periods {
			periodItem := &modelsv2.UpdateBudgetPeriodsItems0{
				Amount: p.Amount.ValueFloat64Pointer(),
			}

			if p.EndAt.String() != "<unknown>" {
				endAt, err := time.Parse("2006-01-02", strings.ReplaceAll(p.EndAt.String(), "\"", ""))
				if err != nil {
					diags.AddError("parsing error", fmt.Sprintf("failed to parse end_at: %s", err))
					return nil
				}

				ea := strfmt.Date(endAt)
				periodItem.EndAt = &ea
			}

			startAt, err := time.Parse("2006-01-02", strings.ReplaceAll(p.StartAt.String(), "\"", ""))
			if err != nil {
				diags.AddError("parsing error", fmt.Sprintf("failed to parse start_at: %s", err))
				return nil
			}

			if !startAt.IsZero() {
				sa := strfmt.Date(startAt)
				periodItem.StartAt = &sa
			}
			dstValues = append(dstValues, periodItem)
		}
		dst.Periods = dstValues
	}

	return dst
}

func applyBudgetPayload(ctx context.Context, isDataSource bool, src *modelsv2.Budget, dst *budgetModel) diag.Diagnostics {
	dst.Token = types.StringValue(src.Token)
	dst.CreatedAt = types.StringValue(src.CreatedAt)
	dst.CreatedByToken = types.StringValue(src.CreatedByToken)
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

    if src.ChildBudgetTokens != nil {
        childBudgetTokens, diag := types.ListValueFrom(ctx, types.StringType, src.ChildBudgetTokens)
        if diag.HasError() {
            return diag
        }
        dst.ChildBudgetTokens = childBudgetTokens
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
		if isDataSource {
			periods := make([]budgetPeriodDataSourceModel, 0, len(src.Periods))
			for _, p := range src.Periods {
				period := budgetPeriodDataSourceModel{
					Amount:  types.StringValue(p.Amount),
					EndAt:   types.StringValue(p.EndAt),
					StartAt: types.StringValue(p.StartAt),
				}
				periods = append(periods, period)
			}

			attrTypes := map[string]attr.Type{
				"amount":   types.StringType,
				"end_at":   types.StringType,
				"start_at": types.StringType,
			}

			l, d := types.ListValueFrom(
				ctx,
				types.ObjectType{AttrTypes: attrTypes},
				periods,
			)
			if d.HasError() {
				return d
			}
			dst.Periods = l
		} else {
			periods := make([]budgetPeriodResourceModel, 0, len(src.Periods))
			for _, p := range src.Periods {
				amt, _ := strconv.ParseFloat(p.Amount, 64)

				period := budgetPeriodResourceModel{
					Amount:  types.Float64Value(amt),
					EndAt:   types.StringValue(p.EndAt),
					StartAt: types.StringValue(p.StartAt),
				}
				periods = append(periods, period)
			}

			attrTypes := map[string]attr.Type{
				"amount":   types.Float64Type,
				"end_at":   types.StringType,
				"start_at": types.StringType,
			}
			tflog.Debug(ctx, fmt.Sprintf("andy 2 isDataSource: %v, periods %v, attrTypes %v", isDataSource, periods, attrTypes))
			l, d := types.ListValueFrom(
				ctx,
				types.ObjectType{AttrTypes: attrTypes},
				periods,
			)
			if d.HasError() {
				return d
			}
			dst.Periods = l
		}

	}
	return nil
}
