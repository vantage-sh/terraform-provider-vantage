package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_report_forecast"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type reportForecastModel resource_report_forecast.ReportForecastModel

func (m *reportForecastModel) applyPayload(ctx context.Context, src *modelsv2.ReportForecast) diag.Diagnostics {
	m.Token = types.StringValue(src.Token)
	m.Id = types.StringValue(src.Token)
	m.Title = types.StringValue(src.Title)
	m.CostReportToken = types.StringValue(src.CostReportToken)
	m.CreatedAt = types.StringValue(src.CreatedAt)
	m.UpdatedAt = types.StringValue(src.UpdatedAt)
	m.CreatedByToken = types.StringPointerValue(src.CreatedByToken)
	// Keep cleared optional strings as "" so nullableStringPlanModifier stays stable.
	m.BusinessMetricToken = nullableStringFromAPI(src.BusinessMetricToken)
	m.IsDefault = types.BoolValue(src.IsDefault)

	if m.SetAsDefault.IsNull() || m.SetAsDefault.IsUnknown() {
		m.SetAsDefault = types.BoolValue(src.IsDefault)
	}

	tokens := src.ScenarioModelTokens
	if tokens == nil {
		tokens = []string{}
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, tokens)
	if diags.HasError() {
		return diags
	}
	m.ScenarioModelTokens = list
	return nil
}

func (m *reportForecastModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateReportForecast {
	dst := &modelsv2.CreateReportForecast{
		Title:           m.Title.ValueStringPointer(),
		CostReportToken: m.CostReportToken.ValueStringPointer(),
	}

	if !m.BusinessMetricToken.IsNull() && !m.BusinessMetricToken.IsUnknown() {
		// Empty string pointer keeps the JSON key present under omitempty so the API clears.
		dst.BusinessMetricToken = m.BusinessMetricToken.ValueStringPointer()
	}
	if !m.SetAsDefault.IsNull() && !m.SetAsDefault.IsUnknown() {
		dst.SetAsDefault = m.SetAsDefault.ValueBoolPointer()
	}

	if !m.ScenarioModelTokens.IsNull() && !m.ScenarioModelTokens.IsUnknown() {
		tokens := []string{}
		diags.Append(m.ScenarioModelTokens.ElementsAs(ctx, &tokens, false)...)
		if diags.HasError() {
			return nil
		}
		dst.ScenarioModelTokens = tokens
	} else {
		dst.ScenarioModelTokens = []string{}
	}

	return dst
}

func (m *reportForecastModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateReportForecast {
	dst := &modelsv2.UpdateReportForecast{}

	if !m.Title.IsNull() && !m.Title.IsUnknown() {
		dst.Title = m.Title.ValueString()
	}
	if !m.BusinessMetricToken.IsNull() && !m.BusinessMetricToken.IsUnknown() {
		// Empty string pointer keeps the JSON key present under omitempty so the API clears.
		dst.BusinessMetricToken = m.BusinessMetricToken.ValueStringPointer()
	}
	if !m.SetAsDefault.IsNull() && !m.SetAsDefault.IsUnknown() {
		dst.SetAsDefault = m.SetAsDefault.ValueBoolPointer()
	}

	if !m.ScenarioModelTokens.IsNull() && !m.ScenarioModelTokens.IsUnknown() {
		tokens := []string{}
		diags.Append(m.ScenarioModelTokens.ElementsAs(ctx, &tokens, false)...)
		if diags.HasError() {
			return nil
		}
		dst.ScenarioModelTokens = tokens
	}

	return dst
}
