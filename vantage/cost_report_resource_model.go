package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_cost_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type costReportModel resource_cost_report.CostReportModel

func (m *costReportModel) applyPayload(ctx context.Context, payload *modelsv2.CostReport, isDataSource bool) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.Title = types.StringValue(payload.Title)
	m.Filter = types.StringPointerValue(payload.Filter)
	m.FolderToken = types.StringPointerValue(payload.FolderToken)
	m.WorkspaceToken = types.StringValue(payload.WorkspaceToken)
	m.Groupings = ptrStringOrEmpty(payload.Groupings)
	m.StartDate = types.StringPointerValue(payload.StartDate)
	m.EndDate = types.StringPointerValue(payload.EndDate)
	m.PreviousPeriodStartDate = types.StringPointerValue(payload.PreviousPeriodStartDate)
	m.PreviousPeriodEndDate = types.StringPointerValue(payload.PreviousPeriodEndDate)
	m.DateInterval = types.StringValue(payload.DateInterval)
	m.ChartType = types.StringValue(payload.ChartType)
	m.DateBin = types.StringValue(payload.DateBin)
	m.CreatedAt = types.StringValue(payload.CreatedAt)

	savedFilterTokensValue, d := types.ListValueFrom(ctx, types.StringType, payload.SavedFilterTokens)
	if d.HasError() {
		diags.Append(d...)
		return diags
	}
	m.SavedFilterTokens = savedFilterTokensValue

	// Handle nested objects - settings, chart_settings, business_metric_tokens_with_metadata
	// For now, set these to null if not provided in the plan
	// These fields are new and the old resource didn't support them
	if m.Settings.IsNull() || m.Settings.IsUnknown() {
		m.Settings = resource_cost_report.NewSettingsValueNull()
	}

	if m.ChartSettings.IsNull() || m.ChartSettings.IsUnknown() {
		m.ChartSettings = resource_cost_report.NewChartSettingsValueNull()
	}

	if m.BusinessMetricTokensWithMetadata.IsNull() || m.BusinessMetricTokensWithMetadata.IsUnknown() {
		m.BusinessMetricTokensWithMetadata = types.ListNull(types.ObjectType{
			AttrTypes: resource_cost_report.BusinessMetricTokensWithMetadataValue{}.AttributeTypes(ctx),
		})
	}

	return diags
}

func (m *costReportModel) toCreateModel(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateCostReport {
	create := &modelsv2.CreateCostReport{
		Title:  m.Title.ValueStringPointer(),
		Filter: m.Filter.ValueString(),
	}

	// Optional fields
	if !m.FolderToken.IsNull() && !m.FolderToken.IsUnknown() {
		create.FolderToken = m.FolderToken.ValueString()
	}

	if !m.WorkspaceToken.IsNull() && !m.WorkspaceToken.IsUnknown() {
		create.WorkspaceToken = m.WorkspaceToken.ValueString()
	}

	if !m.Groupings.IsNull() && !m.Groupings.IsUnknown() {
		create.Groupings = m.Groupings.ValueString()
	}

	if !m.StartDate.IsNull() && !m.StartDate.IsUnknown() {
		create.StartDate = m.StartDate.ValueString()
	}

	if !m.EndDate.IsNull() && !m.EndDate.IsUnknown() {
		create.EndDate = m.EndDate.ValueStringPointer()
	}

	if !m.DateInterval.IsNull() && !m.DateInterval.IsUnknown() {
		create.DateInterval = m.DateInterval.ValueString()
	}

	if !m.PreviousPeriodStartDate.IsNull() && !m.PreviousPeriodStartDate.IsUnknown() {
		create.PreviousPeriodStartDate = m.PreviousPeriodStartDate.ValueString()
	}

	if !m.PreviousPeriodEndDate.IsNull() && !m.PreviousPeriodEndDate.IsUnknown() {
		create.PreviousPeriodEndDate = m.PreviousPeriodEndDate.ValueStringPointer()
	}

	if !m.ChartType.IsNull() && !m.ChartType.IsUnknown() {
		create.ChartType = m.ChartType.ValueStringPointer()
	}

	if !m.DateBin.IsNull() && !m.DateBin.IsUnknown() {
		create.DateBin = m.DateBin.ValueStringPointer()
	}

	// Handle saved filter tokens - default to empty array per AGENTS.md guidelines
	if !m.SavedFilterTokens.IsNull() && !m.SavedFilterTokens.IsUnknown() {
		sft := make([]types.String, 0, len(m.SavedFilterTokens.Elements()))
		m.SavedFilterTokens.ElementsAs(ctx, &sft, false)
		create.SavedFilterTokens = fromStringsValue(sft)
	} else {
		create.SavedFilterTokens = []string{}
	}

	return create
}

func (m *costReportModel) toUpdateModel(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateCostReport {
	update := &modelsv2.UpdateCostReport{
		Title:  m.Title.ValueString(),
		Filter: m.Filter.ValueString(),
	}

	// Optional fields
	if !m.FolderToken.IsNull() && !m.FolderToken.IsUnknown() {
		update.FolderToken = m.FolderToken.ValueString()
	}

	if !m.Groupings.IsNull() && !m.Groupings.IsUnknown() {
		update.Groupings = m.Groupings.ValueString()
	}

	if !m.PreviousPeriodStartDate.IsNull() && !m.PreviousPeriodStartDate.IsUnknown() {
		update.PreviousPeriodStartDate = m.PreviousPeriodStartDate.ValueString()
	}

	if !m.PreviousPeriodEndDate.IsNull() && !m.PreviousPeriodEndDate.IsUnknown() {
		update.PreviousPeriodEndDate = m.PreviousPeriodEndDate.ValueString()
	}

	if !m.ChartType.IsNull() && !m.ChartType.IsUnknown() {
		update.ChartType = m.ChartType.ValueStringPointer()
	}

	if !m.DateBin.IsNull() && !m.DateBin.IsUnknown() {
		update.DateBin = m.DateBin.ValueStringPointer()
	}

	// Handle date interval logic
	if m.DateInterval.ValueString() == "custom" {
		update.StartDate = m.StartDate.ValueString()
		update.EndDate = m.EndDate.ValueString()
		update.DateInterval = "custom"
	} else if !m.DateInterval.IsNull() && !m.DateInterval.IsUnknown() {
		update.DateInterval = m.DateInterval.ValueString()
	}

	// Handle saved filter tokens - default to empty array per AGENTS.md guidelines
	if !m.SavedFilterTokens.IsNull() && !m.SavedFilterTokens.IsUnknown() {
		sft := make([]types.String, 0, len(m.SavedFilterTokens.Elements()))
		m.SavedFilterTokens.ElementsAs(ctx, &sft, false)
		update.SavedFilterTokens = fromStringsValue(sft)
	} else {
		update.SavedFilterTokens = []string{}
	}

	return update
}
