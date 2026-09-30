package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_dashboard"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestDashboardModel_applyPayload_kpiSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	kpiCalculation := "sum"
	kpiType := "cost"

	model := &dashboardModel{}
	diags := model.applyPayload(ctx, &modelsv2.Dashboard{
		Title:          "kpi-dashboard",
		Token:          "dshbrd_test",
		WorkspaceToken: "wrkspc_test",
		Widgets: []*modelsv2.DashboardWidget{
			{
				Title:           "Spend KPI",
				WidgetableToken: "rprt_test",
				Settings: &modelsv2.DashboardWidgetSettings{
					DisplayType:    "kpi",
					KpiCalculation: &kpiCalculation,
					KpiType:        &kpiType,
				},
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("applyPayload diagnostics: %v", diags.Errors())
	}

	var widgets []resource_dashboard.WidgetsValue
	if d := model.Widgets.ElementsAs(ctx, &widgets, false); d.HasError() {
		t.Fatalf("ElementsAs diagnostics: %v", d.Errors())
	}
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}

	settings, d := resource_dashboard.SettingsType{}.ValueFromObject(ctx, widgets[0].Settings)
	if d.HasError() {
		t.Fatalf("ValueFromObject diagnostics: %v", d.Errors())
	}
	settingsVal, ok := settings.(resource_dashboard.SettingsValue)
	if !ok {
		t.Fatalf("expected SettingsValue, got %T", settings)
	}
	if got := settingsVal.DisplayType.ValueString(); got != "kpi" {
		t.Fatalf("display_type = %q, want kpi", got)
	}
	if got := settingsVal.KpiCalculation.ValueString(); got != "sum" {
		t.Fatalf("kpi_calculation = %q, want sum", got)
	}
	if got := settingsVal.KpiType.ValueString(); got != "cost" {
		t.Fatalf("kpi_type = %q, want cost", got)
	}
	if !settingsVal.KpiUsageUnit.IsNull() {
		t.Fatalf("kpi_usage_unit = %v, want null", settingsVal.KpiUsageUnit)
	}
}

func TestDashboardModel_toCreate_kpiSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settingsAttrTypes := resource_dashboard.SettingsValue{}.AttributeTypes(ctx)
	settingsVal, d := resource_dashboard.NewSettingsValue(settingsAttrTypes, map[string]attr.Value{
		"display_type":    types.StringValue("kpi"),
		"kpi_calculation": types.StringValue("average"),
		"kpi_type":        types.StringValue("usage"),
		"kpi_usage_unit":  types.StringValue("GB"),
	})
	if d.HasError() {
		t.Fatalf("NewSettingsValue diagnostics: %v", d.Errors())
	}
	settingsObj, d := settingsVal.ToObjectValue(ctx)
	if d.HasError() {
		t.Fatalf("ToObjectValue diagnostics: %v", d.Errors())
	}

	widgetAttrTypes := resource_dashboard.WidgetsValue{}.AttributeTypes(ctx)
	widgetVal, d := resource_dashboard.NewWidgetsValue(widgetAttrTypes, map[string]attr.Value{
		"settings":         settingsObj,
		"title":            types.StringValue("Usage KPI"),
		"widgetable_token": types.StringValue("rprt_test"),
	})
	if d.HasError() {
		t.Fatalf("NewWidgetsValue diagnostics: %v", d.Errors())
	}

	widgets, d := types.ListValueFrom(ctx, resource_dashboard.WidgetsValue{}.Type(ctx), []resource_dashboard.WidgetsValue{widgetVal})
	if d.HasError() {
		t.Fatalf("ListValueFrom diagnostics: %v", d.Errors())
	}

	model := &dashboardModel{
		Title:          types.StringValue("kpi-dashboard"),
		WorkspaceToken: types.StringValue("wrkspc_test"),
		Widgets:        widgets,
	}

	var diags diag.Diagnostics
	payload := model.toCreate(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("toCreate diagnostics: %v", diags.Errors())
	}
	if len(payload.Widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(payload.Widgets))
	}
	settings := payload.Widgets[0].Settings
	if settings == nil {
		t.Fatal("expected settings")
	}
	if got := *settings.DisplayType; got != "kpi" {
		t.Fatalf("display_type = %q, want kpi", got)
	}
	if got := settings.KpiCalculation; got != "average" {
		t.Fatalf("kpi_calculation = %q, want average", got)
	}
	if got := settings.KpiType; got != "usage" {
		t.Fatalf("kpi_type = %q, want usage", got)
	}
	if got := settings.KpiUsageUnit; got != "GB" {
		t.Fatalf("kpi_usage_unit = %q, want GB", got)
	}
}

func TestDashboardModel_toUpdate_kpiSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settingsAttrTypes := resource_dashboard.SettingsValue{}.AttributeTypes(ctx)
	settingsVal, d := resource_dashboard.NewSettingsValue(settingsAttrTypes, map[string]attr.Value{
		"display_type":    types.StringValue("kpi"),
		"kpi_calculation": types.StringValue("sum"),
		"kpi_type":        types.StringValue("cost"),
		"kpi_usage_unit":  types.StringNull(),
	})
	if d.HasError() {
		t.Fatalf("NewSettingsValue diagnostics: %v", d.Errors())
	}
	settingsObj, d := settingsVal.ToObjectValue(ctx)
	if d.HasError() {
		t.Fatalf("ToObjectValue diagnostics: %v", d.Errors())
	}

	widgetAttrTypes := resource_dashboard.WidgetsValue{}.AttributeTypes(ctx)
	widgetVal, d := resource_dashboard.NewWidgetsValue(widgetAttrTypes, map[string]attr.Value{
		"settings":         settingsObj,
		"title":            types.StringValue("Spend KPI"),
		"widgetable_token": types.StringValue("rprt_test"),
	})
	if d.HasError() {
		t.Fatalf("NewWidgetsValue diagnostics: %v", d.Errors())
	}

	widgets, d := types.ListValueFrom(ctx, resource_dashboard.WidgetsValue{}.Type(ctx), []resource_dashboard.WidgetsValue{widgetVal})
	if d.HasError() {
		t.Fatalf("ListValueFrom diagnostics: %v", d.Errors())
	}

	model := &dashboardModel{
		Title:   types.StringValue("kpi-dashboard"),
		Widgets: widgets,
	}

	var diags diag.Diagnostics
	payload := model.toUpdate(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("toUpdate diagnostics: %v", diags.Errors())
	}
	if len(payload.Widgets) != 1 || payload.Widgets[0].Settings == nil {
		t.Fatal("expected widget settings")
	}
	settings := payload.Widgets[0].Settings
	if got := *settings.DisplayType; got != "kpi" {
		t.Fatalf("display_type = %q, want kpi", got)
	}
	if got := settings.KpiCalculation; got != "sum" {
		t.Fatalf("kpi_calculation = %q, want sum", got)
	}
	if got := settings.KpiType; got != "cost" {
		t.Fatalf("kpi_type = %q, want cost", got)
	}
	if got := settings.KpiUsageUnit; got != "" {
		t.Fatalf("kpi_usage_unit = %q, want empty", got)
	}
}
