package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_dashboard"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type dashboardModel resource_dashboard.DashboardModel

func (m *dashboardModel) applyPayload(ctx context.Context, payload *modelsv2.Dashboard) diag.Diagnostics {
	planStartDate := m.StartDate
	planEndDate := m.EndDate

	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.DateBin = types.StringPointerValue(payload.DateBin)
	if payload.DateInterval != nil && *payload.DateInterval != "" {
		m.DateInterval = types.StringPointerValue(payload.DateInterval)
	} else {
		m.DateInterval = types.StringNull()
	}

	if payload.DateInterval != nil && *payload.DateInterval == "custom" {
		m.StartDate = ptrStringOrEmpty(payload.StartDate)
		m.EndDate = ptrStringOrEmpty(payload.EndDate)
	} else if !planStartDate.IsNull() && !planStartDate.IsUnknown() && planStartDate.ValueString() != "" &&
		!planEndDate.IsNull() && !planEndDate.IsUnknown() && planEndDate.ValueString() != "" {
		m.StartDate = planStartDate
		m.EndDate = planEndDate
	} else {
		m.StartDate = types.StringValue("")
		m.EndDate = types.StringValue("")
	}

	saved_filters, diag := types.ListValueFrom(ctx, types.StringType, payload.SavedFilterTokens)
	if diag.HasError() {
		return diag
	}
	m.SavedFilterTokens = saved_filters

	m.Title = types.StringValue(payload.Title)
	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)

	tfWidgets := make([]resource_dashboard.WidgetsValue, 0, len(payload.Widgets))
	settingsAttrTypes := resource_dashboard.SettingsValue{}.AttributeTypes(ctx)
	widgetAttrTypes := resource_dashboard.WidgetsValue{}.AttributeTypes(ctx)
	for _, widget := range payload.Widgets {
		var settingsObj basetypes.ObjectValue

		if widget.Settings != nil {
			settingsAttrs := map[string]attr.Value{
				"display_type":    types.StringValue(widget.Settings.DisplayType),
				"grid":            dashboardWidgetGridObject(ctx, widget.Settings.Grid),
				"kpi_calculation": types.StringPointerValue(widget.Settings.KpiCalculation),
				"kpi_type":        types.StringPointerValue(widget.Settings.KpiType),
				"kpi_usage_unit":  types.StringPointerValue(widget.Settings.KpiUsageUnit),
			}
			settingsVal, diag := resource_dashboard.NewSettingsValue(settingsAttrTypes, settingsAttrs)
			if diag.HasError() {
				return diag
			}
			settingsObj, diag = settingsVal.ToObjectValue(ctx)
			if diag.HasError() {
				return diag
			}
		} else {
			settingsObj = types.ObjectNull(settingsAttrTypes)
		}

		widgetAttrs := map[string]attr.Value{
			"settings":         settingsObj,
			"title":            types.StringValue(widget.Title),
			"token":            types.StringValue(widget.Token),
			"widgetable_token": stringValueOrNull(widget.WidgetableToken),
		}

		tfWidget, diag := resource_dashboard.NewWidgetsValue(widgetAttrTypes, widgetAttrs)
		if diag.HasError() {
			return diag
		}

		tfWidgets = append(tfWidgets, tfWidget)
	}

	widgets, diag := types.ListValueFrom(ctx, resource_dashboard.WidgetsValue{}.Type(ctx), tfWidgets)
	if diag.HasError() {
		return diag
	}

	m.Widgets = widgets
	m.WorkspaceToken = types.StringValue(payload.WorkspaceToken)

	return nil
}

func (m *dashboardModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateDashboard {
	savedFilterTokens := []types.String{}
	if !m.SavedFilterTokens.IsNull() && !m.SavedFilterTokens.IsUnknown() {
		savedFilterTokens = make([]types.String, 0, len(m.SavedFilterTokens.Elements()))
		if diag := m.SavedFilterTokens.ElementsAs(ctx, &savedFilterTokens, false); diag.HasError() {
			diags.Append(diag...)
			return nil
		}
	}

	widgets := []*modelsv2.CreateDashboardWidgetsItems0{}
	if !m.Widgets.IsNull() && !m.Widgets.IsUnknown() {
		tfWidgets := make([]resource_dashboard.WidgetsValue, 0, len(m.Widgets.Elements()))
		if diag := m.Widgets.ElementsAs(ctx, &tfWidgets, false); diag.HasError() {
			diags.Append(diag...)
			return nil
		}
		for _, w := range tfWidgets {
			widget := &modelsv2.CreateDashboardWidgetsItems0{
				WidgetableToken: w.WidgetableToken.ValueString(),
				Title:           w.Title.ValueString(),
			}

			if !w.Settings.IsNull() && !w.Settings.IsUnknown() {
				tfSettings, diag := resource_dashboard.SettingsType{}.ValueFromObject(ctx, w.Settings)
				if diag.HasError() {
					diags.Append(diag...)
					return nil
				}

				tfSettingsTyped, ok := tfSettings.(resource_dashboard.SettingsValue)
				if !ok {
					diags.AddError("Error converting widgets", "Error converting widgets")
					return nil
				}

				widget.Settings = createDashboardWidgetSettings(ctx, tfSettingsTyped, diags)
				if diags.HasError() {
					return nil
				}
			}

			widgets = append(widgets, widget)
		}
	}
	payload := &modelsv2.CreateDashboard{
		DateBin:           m.DateBin.ValueString(),
		SavedFilterTokens: fromStringsValue(savedFilterTokens),
		Title:             m.Title.ValueStringPointer(),
		Widgets:           widgets,
		WorkspaceToken:    m.WorkspaceToken.ValueString(),
		DateInterval:      m.DateInterval.ValueString(),
	}

	if !m.StartDate.IsNull() && !m.StartDate.IsUnknown() && m.StartDate.ValueString() != "" &&
		!m.EndDate.IsNull() && !m.EndDate.IsUnknown() && m.EndDate.ValueString() != "" {
		payload.StartDate = m.StartDate.ValueString()
		payload.EndDate = m.EndDate.ValueString()
		payload.DateInterval = "custom"
	}

	return payload
}

func (m *dashboardModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateDashboard {
	savedFilterTokens := []types.String{}
	if !m.SavedFilterTokens.IsNull() && !m.SavedFilterTokens.IsUnknown() {
		savedFilterTokens = make([]types.String, 0, len(m.SavedFilterTokens.Elements()))
		diags.Append(m.SavedFilterTokens.ElementsAs(ctx, &savedFilterTokens, false)...)
		if diags.HasError() {
			return nil
		}
	}

	widgets := []*modelsv2.UpdateDashboardWidgetsItems0{}
	if !m.Widgets.IsNull() && !m.Widgets.IsUnknown() {
		tfWidgets := make([]resource_dashboard.WidgetsValue, 0, len(m.Widgets.Elements()))
		if diag := m.Widgets.ElementsAs(ctx, &tfWidgets, false); diag.HasError() {
			diags.Append(diag...)
			return nil
		}
		for _, w := range tfWidgets {
			widget := &modelsv2.UpdateDashboardWidgetsItems0{
				WidgetableToken: w.WidgetableToken.ValueString(),
				Title:           w.Title.ValueString(),
			}

			if !w.Settings.IsNull() && !w.Settings.IsUnknown() {
				tfSettings, diag := resource_dashboard.SettingsType{}.ValueFromObject(ctx, w.Settings)
				if diag.HasError() {
					diags.Append(diag...)
					return nil
				}

				tfSettingsTyped, ok := tfSettings.(resource_dashboard.SettingsValue)
				if !ok {
					diags.AddError("Error converting widgets", "Error converting widgets")
					return nil
				}

				widget.Settings = updateDashboardWidgetSettings(ctx, tfSettingsTyped, diags)
				if diags.HasError() {
					return nil
				}
			}

			widgets = append(widgets, widget)
		}
	}

	// date_interval is a pointer with omitempty in the regenerated SDK. Always
	// send a non-nil value so clearing the attribute still emits "" to the API,
	// matching the previous non-omitempty string field behavior.
	dateInterval := ""
	if !m.DateInterval.IsNull() && !m.DateInterval.IsUnknown() {
		dateInterval = m.DateInterval.ValueString()
	}

	payload := &modelsv2.UpdateDashboard{
		DateBin:           m.DateBin.ValueString(),
		SavedFilterTokens: fromStringsValue(savedFilterTokens),
		Title:             m.Title.ValueString(),
		Widgets:           widgets,
		WorkspaceToken:    m.WorkspaceToken.ValueString(),
		DateInterval:      &dateInterval,
	}

	if !m.StartDate.IsNull() && !m.StartDate.IsUnknown() && m.StartDate.ValueString() != "" &&
		!m.EndDate.IsNull() && !m.EndDate.IsUnknown() && m.EndDate.ValueString() != "" {
		payload.StartDate = m.StartDate.ValueStringPointer()
		payload.EndDate = m.EndDate.ValueStringPointer()
		customInterval := "custom"
		payload.DateInterval = &customInterval
	}

	return payload
}

func createDashboardWidgetSettings(ctx context.Context, s resource_dashboard.SettingsValue, diags *diag.Diagnostics) *modelsv2.CreateDashboardWidgetsItems0Settings {
	settings := &modelsv2.CreateDashboardWidgetsItems0Settings{
		DisplayType: s.DisplayType.ValueStringPointer(),
	}
	if !s.KpiCalculation.IsNull() && !s.KpiCalculation.IsUnknown() {
		settings.KpiCalculation = s.KpiCalculation.ValueString()
	}
	if !s.KpiType.IsNull() && !s.KpiType.IsUnknown() {
		settings.KpiType = s.KpiType.ValueString()
	}
	if !s.KpiUsageUnit.IsNull() && !s.KpiUsageUnit.IsUnknown() {
		settings.KpiUsageUnit = s.KpiUsageUnit.ValueString()
	}
	settings.Grid = createDashboardWidgetGrid(ctx, s.Grid, diags)
	if diags.HasError() {
		return nil
	}
	return settings
}

func updateDashboardWidgetSettings(ctx context.Context, s resource_dashboard.SettingsValue, diags *diag.Diagnostics) *modelsv2.UpdateDashboardWidgetsItems0Settings {
	settings := &modelsv2.UpdateDashboardWidgetsItems0Settings{
		DisplayType: s.DisplayType.ValueStringPointer(),
	}
	if !s.KpiCalculation.IsNull() && !s.KpiCalculation.IsUnknown() {
		settings.KpiCalculation = s.KpiCalculation.ValueString()
	}
	if !s.KpiType.IsNull() && !s.KpiType.IsUnknown() {
		settings.KpiType = s.KpiType.ValueString()
	}
	if !s.KpiUsageUnit.IsNull() && !s.KpiUsageUnit.IsUnknown() {
		settings.KpiUsageUnit = s.KpiUsageUnit.ValueString()
	}
	settings.Grid = updateDashboardWidgetGrid(ctx, s.Grid, diags)
	if diags.HasError() {
		return nil
	}
	return settings
}

func createDashboardWidgetGrid(ctx context.Context, gridObj basetypes.ObjectValue, diags *diag.Diagnostics) *modelsv2.CreateDashboardWidgetsItems0SettingsGrid {
	if gridObj.IsNull() || gridObj.IsUnknown() {
		return nil
	}

	gridValuable, d := resource_dashboard.GridType{}.ValueFromObject(ctx, gridObj)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	grid, ok := gridValuable.(resource_dashboard.GridValue)
	if !ok {
		diags.AddError("Error converting widgets", "Error converting widget grid settings")
		return nil
	}

	x := int32(grid.X.ValueInt64())
	y := int32(grid.Y.ValueInt64())
	w := int32(grid.W.ValueInt64())
	h := int32(grid.H.ValueInt64())
	return &modelsv2.CreateDashboardWidgetsItems0SettingsGrid{
		X: &x,
		Y: &y,
		W: &w,
		H: &h,
	}
}

func updateDashboardWidgetGrid(ctx context.Context, gridObj basetypes.ObjectValue, diags *diag.Diagnostics) *modelsv2.UpdateDashboardWidgetsItems0SettingsGrid {
	if gridObj.IsNull() || gridObj.IsUnknown() {
		return nil
	}

	gridValuable, d := resource_dashboard.GridType{}.ValueFromObject(ctx, gridObj)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	grid, ok := gridValuable.(resource_dashboard.GridValue)
	if !ok {
		diags.AddError("Error converting widgets", "Error converting widget grid settings")
		return nil
	}

	x := int32(grid.X.ValueInt64())
	y := int32(grid.Y.ValueInt64())
	w := int32(grid.W.ValueInt64())
	h := int32(grid.H.ValueInt64())
	return &modelsv2.UpdateDashboardWidgetsItems0SettingsGrid{
		X: &x,
		Y: &y,
		W: &w,
		H: &h,
	}
}

func dashboardWidgetGridObject(ctx context.Context, grid *modelsv2.DashboardWidgetGridLayout) types.Object {
	gridAttrTypes := resource_dashboard.GridValue{}.AttributeTypes(ctx)
	if grid == nil {
		return types.ObjectNull(gridAttrTypes)
	}

	gridVal, diags := resource_dashboard.NewGridValue(gridAttrTypes, map[string]attr.Value{
		"x": types.Int64Value(int64(grid.X)),
		"y": types.Int64Value(int64(grid.Y)),
		"w": types.Int64Value(int64(grid.W)),
		"h": types.Int64Value(int64(grid.H)),
	})
	if diags.HasError() {
		return types.ObjectNull(gridAttrTypes)
	}
	obj, diags := gridVal.ToObjectValue(ctx)
	if diags.HasError() {
		return types.ObjectNull(gridAttrTypes)
	}
	return obj
}

func stringValueOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
