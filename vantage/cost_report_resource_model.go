package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

	// Handle chart_settings from API payload.
	// Only populate when the user has configured chart_settings to avoid drift
	// on plans where the block is omitted from the config.
	if !m.ChartSettings.IsNull() && !m.ChartSettings.IsUnknown() && payload.ChartSettings != nil {
		xAxisDimension, d := types.ListValueFrom(ctx, types.StringType, payload.ChartSettings.XAxisDimension)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		csValue, d := resource_cost_report.NewChartSettingsValue(
			resource_cost_report.ChartSettingsValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"x_axis_dimension": xAxisDimension,
				"y_axis_dimension": types.StringValue(payload.ChartSettings.YAxisDimension),
			},
		)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		m.ChartSettings = csValue
	} else if m.ChartSettings.IsNull() || m.ChartSettings.IsUnknown() {
		m.ChartSettings = resource_cost_report.NewChartSettingsValueNull()
	}

	// Handle settings from API payload.
	// Only populate when the user has configured settings (not null/unknown) to
	// avoid drift on plans where settings is omitted from the config. For
	// Optional+Computed nested objects, Terraform cannot reconcile state values
	// with an absent config block, causing perpetual "(known after apply)" drift.
	if !m.Settings.IsNull() && !m.Settings.IsUnknown() && payload.Settings != nil {
		settingsValue, d := resource_cost_report.NewSettingsValue(
			resource_cost_report.SettingsValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"aggregate_by":         types.StringPointerValue(payload.Settings.AggregateBy),
				"amortize":             types.BoolPointerValue(payload.Settings.Amortize),
				"include_credits":      types.BoolPointerValue(payload.Settings.IncludeCredits),
				"include_discounts":    types.BoolPointerValue(payload.Settings.IncludeDiscounts),
				"include_refunds":      types.BoolPointerValue(payload.Settings.IncludeRefunds),
				"include_tax":          types.BoolPointerValue(payload.Settings.IncludeTax),
				"show_previous_period": types.BoolPointerValue(payload.Settings.ShowPreviousPeriod),
				"unallocated":          types.BoolPointerValue(payload.Settings.Unallocated),
			},
		)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		m.Settings = settingsValue
	} else if m.Settings.IsNull() || m.Settings.IsUnknown() {
		m.Settings = resource_cost_report.NewSettingsValueNull()
	}

	// Handle business_metric_tokens_with_metadata from API payload.
	// Same conditional approach as settings to avoid drift.
	bmtElemType := types.ObjectType{
		AttrTypes: resource_cost_report.BusinessMetricTokensWithMetadataValue{}.AttributeTypes(ctx),
	}
	if !m.BusinessMetricTokensWithMetadata.IsNull() && !m.BusinessMetricTokensWithMetadata.IsUnknown() {
		bmtValues := make([]attr.Value, 0, len(payload.BusinessMetricTokensWithMetadata))
		for _, bmt := range payload.BusinessMetricTokensWithMetadata {
			labelFilter, d := types.ListValueFrom(ctx, types.StringType, bmt.LabelFilter)
			if d.HasError() {
				diags.Append(d...)
				return diags
			}
			objVal, d := types.ObjectValue(
				resource_cost_report.BusinessMetricTokensWithMetadataValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"business_metric_token": types.StringValue(bmt.BusinessMetricToken),
					"label_filter":          labelFilter,
					"unit_scale":            types.StringValue(bmt.UnitScale),
				},
			)
			if d.HasError() {
				diags.Append(d...)
				return diags
			}
			bmtValues = append(bmtValues, objVal)
		}
		bmtList, d := types.ListValue(bmtElemType, bmtValues)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		m.BusinessMetricTokensWithMetadata = bmtList
	} else {
		m.BusinessMetricTokensWithMetadata = types.ListNull(bmtElemType)
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

	// Handle chart_settings
	if !m.ChartSettings.IsNull() && !m.ChartSettings.IsUnknown() {
		cs := &modelsv2.CreateCostReportChartSettings{}
		if !m.ChartSettings.XAxisDimension.IsNull() && !m.ChartSettings.XAxisDimension.IsUnknown() {
			items := []string{}
			d := m.ChartSettings.XAxisDimension.ElementsAs(ctx, &items, false)
			diags.Append(d...)
			cs.XAxisDimension = items
		}
		if !m.ChartSettings.YAxisDimension.IsNull() && !m.ChartSettings.YAxisDimension.IsUnknown() {
			cs.YAxisDimension = m.ChartSettings.YAxisDimension.ValueString()
		}
		create.ChartSettings = cs
	}

	// Handle settings
	if !m.Settings.IsNull() && !m.Settings.IsUnknown() {
		s := &modelsv2.CreateCostReportSettings{}
		if !m.Settings.AggregateBy.IsNull() && !m.Settings.AggregateBy.IsUnknown() {
			s.AggregateBy = m.Settings.AggregateBy.ValueStringPointer()
		}
		if !m.Settings.Amortize.IsNull() && !m.Settings.Amortize.IsUnknown() {
			s.Amortize = m.Settings.Amortize.ValueBoolPointer()
		}
		if !m.Settings.IncludeCredits.IsNull() && !m.Settings.IncludeCredits.IsUnknown() {
			s.IncludeCredits = m.Settings.IncludeCredits.ValueBoolPointer()
		}
		if !m.Settings.IncludeDiscounts.IsNull() && !m.Settings.IncludeDiscounts.IsUnknown() {
			s.IncludeDiscounts = m.Settings.IncludeDiscounts.ValueBoolPointer()
		}
		if !m.Settings.IncludeRefunds.IsNull() && !m.Settings.IncludeRefunds.IsUnknown() {
			s.IncludeRefunds = m.Settings.IncludeRefunds.ValueBoolPointer()
		}
		if !m.Settings.IncludeTax.IsNull() && !m.Settings.IncludeTax.IsUnknown() {
			s.IncludeTax = m.Settings.IncludeTax.ValueBoolPointer()
		}
		if !m.Settings.ShowPreviousPeriod.IsNull() && !m.Settings.ShowPreviousPeriod.IsUnknown() {
			s.ShowPreviousPeriod = m.Settings.ShowPreviousPeriod.ValueBoolPointer()
		}
		if !m.Settings.Unallocated.IsNull() && !m.Settings.Unallocated.IsUnknown() {
			s.Unallocated = m.Settings.Unallocated.ValueBoolPointer()
		}
		create.Settings = s
	}

	// Handle business_metric_tokens_with_metadata
	if !m.BusinessMetricTokensWithMetadata.IsNull() && !m.BusinessMetricTokensWithMetadata.IsUnknown() {
		bmtItems := make([]*modelsv2.CreateCostReportBusinessMetricTokensWithMetadataItems0, 0)
		for _, elem := range m.BusinessMetricTokensWithMetadata.Elements() {
			obj := elem.(types.Object)
			attrs := obj.Attributes()
			item := &modelsv2.CreateCostReportBusinessMetricTokensWithMetadataItems0{}
			if token, ok := attrs["business_metric_token"]; ok && !token.IsNull() && !token.IsUnknown() {
				item.BusinessMetricToken = token.(types.String).ValueStringPointer()
			}
			if unitScale, ok := attrs["unit_scale"]; ok && !unitScale.IsNull() && !unitScale.IsUnknown() {
				item.UnitScale = unitScale.(types.String).ValueStringPointer()
			}
			if lf, ok := attrs["label_filter"]; ok && !lf.IsNull() && !lf.IsUnknown() {
				lfList := lf.(types.List)
				items := []string{}
				d := lfList.ElementsAs(ctx, &items, false)
				diags.Append(d...)
				item.LabelFilter = items
			} else {
				item.LabelFilter = []string{}
			}
			bmtItems = append(bmtItems, item)
		}
		create.BusinessMetricTokensWithMetadata = bmtItems
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

	// Always send groupings so it can be cleared (empty string clears it)
	update.Groupings = m.Groupings.ValueString()

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

	// Handle chart_settings
	if !m.ChartSettings.IsNull() && !m.ChartSettings.IsUnknown() {
		cs := &modelsv2.UpdateCostReportChartSettings{}
		if !m.ChartSettings.XAxisDimension.IsNull() && !m.ChartSettings.XAxisDimension.IsUnknown() {
			items := []string{}
			d := m.ChartSettings.XAxisDimension.ElementsAs(ctx, &items, false)
			diags.Append(d...)
			cs.XAxisDimension = items
		}
		if !m.ChartSettings.YAxisDimension.IsNull() && !m.ChartSettings.YAxisDimension.IsUnknown() {
			cs.YAxisDimension = m.ChartSettings.YAxisDimension.ValueString()
		}
		update.ChartSettings = cs
	}

	// Handle settings (UpdateCostReportSettings uses value types, not pointers)
	if !m.Settings.IsNull() && !m.Settings.IsUnknown() {
		s := &modelsv2.UpdateCostReportSettings{}
		if !m.Settings.AggregateBy.IsNull() && !m.Settings.AggregateBy.IsUnknown() {
			s.AggregateBy = m.Settings.AggregateBy.ValueString()
		}
		if !m.Settings.Amortize.IsNull() && !m.Settings.Amortize.IsUnknown() {
			s.Amortize = m.Settings.Amortize.ValueBool()
		}
		if !m.Settings.IncludeCredits.IsNull() && !m.Settings.IncludeCredits.IsUnknown() {
			s.IncludeCredits = m.Settings.IncludeCredits.ValueBool()
		}
		if !m.Settings.IncludeDiscounts.IsNull() && !m.Settings.IncludeDiscounts.IsUnknown() {
			s.IncludeDiscounts = m.Settings.IncludeDiscounts.ValueBool()
		}
		if !m.Settings.IncludeRefunds.IsNull() && !m.Settings.IncludeRefunds.IsUnknown() {
			s.IncludeRefunds = m.Settings.IncludeRefunds.ValueBool()
		}
		if !m.Settings.IncludeTax.IsNull() && !m.Settings.IncludeTax.IsUnknown() {
			s.IncludeTax = m.Settings.IncludeTax.ValueBool()
		}
		if !m.Settings.ShowPreviousPeriod.IsNull() && !m.Settings.ShowPreviousPeriod.IsUnknown() {
			s.ShowPreviousPeriod = m.Settings.ShowPreviousPeriod.ValueBool()
		}
		if !m.Settings.Unallocated.IsNull() && !m.Settings.Unallocated.IsUnknown() {
			s.Unallocated = m.Settings.Unallocated.ValueBool()
		}
		update.Settings = s
	}

	// Handle business_metric_tokens_with_metadata
	if !m.BusinessMetricTokensWithMetadata.IsNull() && !m.BusinessMetricTokensWithMetadata.IsUnknown() {
		bmtItems := make([]*modelsv2.UpdateCostReportBusinessMetricTokensWithMetadataItems0, 0)
		for _, elem := range m.BusinessMetricTokensWithMetadata.Elements() {
			obj := elem.(types.Object)
			attrs := obj.Attributes()
			item := &modelsv2.UpdateCostReportBusinessMetricTokensWithMetadataItems0{}
			if token, ok := attrs["business_metric_token"]; ok && !token.IsNull() && !token.IsUnknown() {
				item.BusinessMetricToken = token.(types.String).ValueStringPointer()
			}
			if unitScale, ok := attrs["unit_scale"]; ok && !unitScale.IsNull() && !unitScale.IsUnknown() {
				item.UnitScale = unitScale.(types.String).ValueStringPointer()
			}
			if lf, ok := attrs["label_filter"]; ok && !lf.IsNull() && !lf.IsUnknown() {
				lfList := lf.(types.List)
				items := []string{}
				d := lfList.ElementsAs(ctx, &items, false)
				diags.Append(d...)
				item.LabelFilter = items
			} else {
				item.LabelFilter = []string{}
			}
			bmtItems = append(bmtItems, item)
		}
		update.BusinessMetricTokensWithMetadata = bmtItems
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
