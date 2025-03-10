package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_dashboard"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type dashboardModel resource_dashboard.DashboardModel

func (m *dashboardModel) applyPayload(ctx context.Context, payload *modelsv2.Dashboard) diag.Diagnostics {
	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.DateBin = types.StringValue(payload.DateBin)
	m.DateInterval = types.StringValue(payload.DateInterval)
	m.EndDate = types.StringValue(payload.EndDate)

	saved_filters, diag := types.ListValueFrom(ctx, types.StringType, payload.SavedFilterTokens)
	if diag.HasError() {
		return diag
	}
	m.SavedFilterTokens = saved_filters

	m.StartDate = types.StringValue(payload.StartDate)
	m.Title = types.StringValue(payload.Title)
	m.Token = types.StringValue(payload.Token)
	m.UpdatedAt = types.StringValue(payload.UpdatedAt)

	tfWidgets := make([]basetypes.ObjectValue, 0, len(payload.Widgets))
	for _, widget := range payload.Widgets {
		tfWidget := resource_dashboard.WidgetsValue{
			Title:           types.StringValue(widget.Title),
			WidgetableToken: types.StringValue(widget.WidgetableToken),
		}

		if widget.Settings != nil {
			s := resource_dashboard.SettingsValue{
				DisplayType: types.StringValue(widget.Settings.DisplayType),
			}

			sObj, diag := s.ToObjectValue(ctx)
			if diag.HasError() {
				return diag
			}

			tfWidget.Settings = sObj
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

	payload := &modelsv2.CreateDashboard{
		DateBin:           m.DateBin.ValueString(),
		SavedFilterTokens: fromStringsValue(savedFilterTokens),
		Title:             m.Title.ValueStringPointer(),
		Widgets:           widgets,
		WorkspaceToken:    m.WorkspaceToken.ValueString(),
	}

	if m.DateInterval.ValueString() == "" || m.DateInterval.ValueString() == "custom" {
		payload.StartDate = m.StartDate.ValueString()
		payload.EndDate = m.EndDate.ValueString()
	} else {
		payload.DateInterval = m.DateInterval.ValueStringPointer()
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

	payload := &modelsv2.UpdateDashboard{
		DateBin:           m.DateBin.ValueString(),
		SavedFilterTokens: fromStringsValue(savedFilterTokens),
		Title:             m.Title.ValueString(),
		Widgets:           widgets,
		WorkspaceToken:    m.WorkspaceToken.ValueString(),
	}

	if m.DateInterval.ValueString() == "" || m.DateInterval.ValueString() == "custom" {
		payload.StartDate = m.StartDate.ValueString()
		payload.EndDate = m.EndDate.ValueStringPointer()
	} else {
		payload.DateInterval = m.DateInterval.ValueString()
	}

	return payload
}
