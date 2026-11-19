package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_custom_calculation"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type customCalculationResourceModel resource_custom_calculation.CustomCalculationModel

func (m *customCalculationResourceModel) toCreateModel(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateCustomCalculation {
	scale := int32(m.UnitScale.ValueInt64())
	return &modelsv2.CreateCustomCalculation{
		Title:            m.Title.ValueStringPointer(),
		Description:      m.Description.ValueStringPointer(),
		DisplayAs:        m.DisplayAs.ValueStringPointer(),
		Operation:        m.Operation.ValueStringPointer(),
		UnitScale:        &scale,
		Filter:           calculationFilterRequest(ctx, m.Filter, diags),
		CostReportTokens: calculationTokenRequest(ctx, m.CostReportTokens, diags),
	}
}

func (m *customCalculationResourceModel) toUpdateModel(ctx context.Context, prior customCalculationResourceModel, configuredReports types.Set, diags *diag.Diagnostics) *modelsv2.UpdateCustomCalculation {
	update := &modelsv2.UpdateCustomCalculation{}
	if !m.Title.Equal(prior.Title) {
		update.Title = m.Title.ValueString()
	}
	if !m.Description.Equal(prior.Description) {
		description := m.Description.ValueString()
		update.Description = &description // Core clears a description with an explicit empty string.
	}
	if !m.DisplayAs.Equal(prior.DisplayAs) {
		update.DisplayAs = m.DisplayAs.ValueString()
	}
	if !m.Operation.Equal(prior.Operation) {
		update.Operation = m.Operation.ValueString()
	}
	if !m.UnitScale.Equal(prior.UnitScale) {
		scale := int32(m.UnitScale.ValueInt64())
		update.UnitScale = &scale
	}
	if !m.Filter.Equal(prior.Filter) {
		update.Filter = calculationFilterRequest(ctx, m.Filter, diags)
	}
	if !configuredReports.Equal(prior.CostReportTokens) {
		update.CostReportTokens = calculationTokenRequest(ctx, configuredReports, diags)
	}
	return update
}

func (m *customCalculationResourceModel) applyPayload(ctx context.Context, payload *modelsv2.CustomCalculation) diag.Diagnostics {
	var diags diag.Diagnostics
	m.Id = types.StringValue(payload.Token)
	m.Token = types.StringValue(payload.Token)
	m.Title = types.StringValue(payload.Title)
	m.Description = types.StringPointerValue(payload.Description)
	m.DisplayAs = types.StringValue(payload.DisplayAs)
	m.Operation = types.StringValue(payload.Operation)
	m.UnitScale = types.Int64Value(int64(payload.UnitScale))
	m.CostReportTokens = calculationTokenState(ctx, payload.CostReportTokens, &diags)
	m.Filter = calculationFilterState(ctx, payload.Filter, &diags)
	return diags
}

func calculationTokenRequest(ctx context.Context, value types.Set, diags *diag.Diagnostics) []string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	// A known empty set must encode as [] to clear existing attachments.
	tokens := make([]string, 0, len(value.Elements()))
	diags.Append(value.ElementsAs(ctx, &tokens, false)...)
	return tokens
}

func calculationTokenState(ctx context.Context, tokens []string, diags *diag.Diagnostics) types.Set {
	result, d := types.SetValueFrom(ctx, types.StringType, tokens)
	diags.Append(d...)
	return result
}

func calculationFilterRequest(ctx context.Context, value resource_custom_calculation.FilterValue, diags *diag.Diagnostics) *modelsv2.CustomCalculationFilter {
	return &modelsv2.CustomCalculationFilter{
		Type:                 value.FilterType.ValueStringPointer(),
		Filter:               value.Filter.ValueStringPointer(),
		FilterDenominator:    value.FilterDenominator.ValueStringPointer(),
		FilterNumerator:      value.FilterNumerator.ValueStringPointer(),
		UsageUnit:            usageUnitRequest(ctx, value.UsageUnit, diags),
		UsageUnitDenominator: usageUnitRequest(ctx, value.UsageUnitDenominator, diags),
		UsageUnitNumerator:   usageUnitRequest(ctx, value.UsageUnitNumerator, diags),
	}
}

func usageUnitRequest(ctx context.Context, value types.List, diags *diag.Diagnostics) interface{} {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	var units []string
	diags.Append(value.ElementsAs(ctx, &units, false)...)
	return units
}

func calculationFilterState(ctx context.Context, value *modelsv2.CustomCalculationFilter, diags *diag.Diagnostics) resource_custom_calculation.FilterValue {
	if value == nil {
		return resource_custom_calculation.NewFilterValueNull()
	}
	fields := map[string]attr.Value{
		"type":                   types.StringPointerValue(value.Type),
		"filter":                 types.StringPointerValue(value.Filter),
		"filter_denominator":     types.StringPointerValue(value.FilterDenominator),
		"filter_numerator":       types.StringPointerValue(value.FilterNumerator),
		"usage_unit":             usageUnitState(ctx, value.UsageUnit, diags),
		"usage_unit_denominator": usageUnitState(ctx, value.UsageUnitDenominator, diags),
		"usage_unit_numerator":   usageUnitState(ctx, value.UsageUnitNumerator, diags),
	}
	result, d := resource_custom_calculation.NewFilterValue(resource_custom_calculation.FilterValue{}.AttributeTypes(ctx), fields)
	diags.Append(d...)
	return result
}

func usageUnitState(ctx context.Context, value interface{}, diags *diag.Diagnostics) types.List {
	// The SDK decodes this union into nil, string, or []interface{}.
	var units []string
	switch value := value.(type) {
	case nil:
		return types.ListNull(types.StringType)
	case string:
		units = []string{value}
	case []interface{}:
		units = make([]string, 0, len(value))
		for _, item := range value {
			unit, ok := item.(string)
			if !ok {
				diags.AddError("Invalid usage unit response", fmt.Sprintf("Expected a string, got %T.", item))
				return types.ListNull(types.StringType)
			}
			units = append(units, unit)
		}
	default:
		diags.AddError("Invalid usage unit response", fmt.Sprintf("Expected a string or list, got %T.", value))
		return types.ListNull(types.StringType)
	}
	result, d := types.ListValueFrom(ctx, types.StringType, units)
	diags.Append(d...)
	return result
}
