package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_dashboard"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type dashboardModel resource_dashboard.DashboardModel

func (m *dashboardModel) applyPayload(ctx context.Context, payload *modelsv2.Dashboard) diag.Diagnostics {
	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.DateBin = types.StringPointerValue(payload.DateBin)
	tflog.Debug(ctx, fmt.Sprintf("DateInterval is not null %v", payload.DateInterval))
	if payload.DateInterval != nil && *payload.DateInterval != "" {
		m.DateInterval = types.StringPointerValue(payload.DateInterval)
	} else {
		m.DateInterval = types.StringNull()
	}

	if payload.DateInterval != nil && *payload.DateInterval == "custom" {
		m.StartDate = ptrStringOrEmpty(payload.StartDate)
		m.EndDate = ptrStringOrEmpty(payload.EndDate)
	} else {
		// Keep these aligned with schema defaults (empty string) when date_interval isn't custom.
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

	tfWidgets := make([]basetypes.ObjectValue, 0, len(payload.Widgets))
	for _, widget := range payload.Widgets {
		// Build settings object
		var settingsObj basetypes.ObjectValue
		settingsAttrTypes := map[string]attr.Type{
			"display_type": types.StringType,
		}
		
		if widget.Settings != nil {
			settingsAttrs := map[string]attr.Value{
				"display_type": types.StringValue(widget.Settings.DisplayType),
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
			// Create null settings object when not provided
			settingsObj = types.ObjectNull(settingsAttrTypes)
		}

		// Build widget using proper constructor with attribute types and values
		widgetAttrTypes := map[string]attr.Type{
			"settings": types.ObjectType{AttrTypes: settingsAttrTypes},
			"title":            types.StringType,
			"widgetable_token": types.StringType,
		}
		widgetAttrs := map[string]attr.Value{
			"settings":         settingsObj,
			"title":            types.StringValue(widget.Title),
			"widgetable_token": types.StringValue(widget.WidgetableToken),
		}

		tfWidget, diag := resource_dashboard.NewWidgetsValue(widgetAttrTypes, widgetAttrs)
		if diag.HasError() {
			return diag
		}

		tfValue, diag := tfWidget.ToObjectValue(ctx)
		if diag.HasError() {
			return diag
		}

		tfWidgets = append(tfWidgets, tfValue)
	}

	attrTypes := resource_dashboard.WidgetsValue{}.AttributeTypes(ctx)
	widgets, diag := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attrTypes}, tfWidgets)
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
				WidgetableToken: w.WidgetableToken.ValueStringPointer(),
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

				widget.Settings = &modelsv2.CreateDashboardWidgetsItems0Settings{
					DisplayType: tfSettingsTyped.DisplayType.ValueStringPointer(),
				}
			}

			widgets = append(widgets, widget)
		}
	}
	tflog.Debug(ctx, fmt.Sprintf("DateInterval create %s", m.DateInterval.ValueString()))
	payload := &modelsv2.CreateDashboard{
		DateBin:           m.DateBin.ValueString(),
		SavedFilterTokens: fromStringsValue(savedFilterTokens),
		Title:             m.Title.ValueStringPointer(),
		Widgets:           widgets,
		WorkspaceToken:    m.WorkspaceToken.ValueString(),
		DateInterval:      m.DateInterval.ValueString(),
	}

	if (m.DateInterval.ValueString() == "" && !m.StartDate.IsNull() && !m.EndDate.IsNull()) || m.DateInterval.ValueString() == "custom" {
		// if m.DateInterval.ValueString() == "custom" {
		payload.StartDate = m.StartDate.ValueString()
		payload.EndDate = m.EndDate.ValueString()
	}
	// else {
	// 	payload.DateInterval = m.DateInterval.ValueString()
	// }

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
				WidgetableToken: w.WidgetableToken.ValueStringPointer(),
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

				widget.Settings = &modelsv2.UpdateDashboardWidgetsItems0Settings{
					DisplayType: tfSettingsTyped.DisplayType.ValueStringPointer(),
				}
			}

			widgets = append(widgets, widget)
		}
	}
	var dateInterval string
	if !m.DateInterval.IsNull() {
		dateInterval = m.DateInterval.ValueString()
	}

	payload := &modelsv2.UpdateDashboard{
		DateBin:           m.DateBin.ValueString(),
		SavedFilterTokens: fromStringsValue(savedFilterTokens),
		Title:             m.Title.ValueString(),
		Widgets:           widgets,
		WorkspaceToken:    m.WorkspaceToken.ValueString(),
		DateInterval:      dateInterval,
	}

	if m.DateInterval.ValueString() == "custom" && !m.StartDate.IsNull() && !m.EndDate.IsNull() {
		payload.StartDate = m.StartDate.ValueString()
		payload.EndDate = m.EndDate.ValueString()
	}

	return payload
}
