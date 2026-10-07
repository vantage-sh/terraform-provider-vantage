package vantage

import (
	"context"
	"testing"

	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_dashboards"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestDashboardDataSourceValueFromPayload_kpiSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	kpiCalculation := "average"
	kpiType := "usage"
	kpiUsageUnit := "GB"
	dateBin := "day"
	dateInterval := "this_month"

	value, diags := dashboardDataSourceValueFromPayload(ctx, &modelsv2.Dashboard{
		CreatedAt:         "2026-01-01T00:00:00Z",
		DateBin:           &dateBin,
		DateInterval:      &dateInterval,
		SavedFilterTokens: []string{},
		Title:             "kpi-dashboard",
		Token:             "dshbrd_test",
		UpdatedAt:         "2026-01-02T00:00:00Z",
		WorkspaceToken:    "wrkspc_test",
		Widgets: []*modelsv2.DashboardWidget{
			{
				Title:           "Usage KPI",
				WidgetableToken: "rprt_test",
				Settings: &modelsv2.DashboardWidgetSettings{
					DisplayType:    "kpi",
					KpiCalculation: &kpiCalculation,
					KpiType:        &kpiType,
					KpiUsageUnit:   &kpiUsageUnit,
				},
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("dashboardDataSourceValueFromPayload diagnostics: %v", diags.Errors())
	}

	var widgets []datasource_dashboards.WidgetsValue
	if d := value.Widgets.ElementsAs(ctx, &widgets, false); d.HasError() {
		t.Fatalf("ElementsAs diagnostics: %v", d.Errors())
	}
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}

	settings, d := datasource_dashboards.SettingsType{}.ValueFromObject(ctx, widgets[0].Settings)
	if d.HasError() {
		t.Fatalf("ValueFromObject diagnostics: %v", d.Errors())
	}
	settingsVal, ok := settings.(datasource_dashboards.SettingsValue)
	if !ok {
		t.Fatalf("expected SettingsValue, got %T", settings)
	}
	if got := settingsVal.DisplayType.ValueString(); got != "kpi" {
		t.Fatalf("display_type = %q, want kpi", got)
	}
	if got := settingsVal.KpiCalculation.ValueString(); got != "average" {
		t.Fatalf("kpi_calculation = %q, want average", got)
	}
	if got := settingsVal.KpiType.ValueString(); got != "usage" {
		t.Fatalf("kpi_type = %q, want usage", got)
	}
	if got := settingsVal.KpiUsageUnit.ValueString(); got != "GB" {
		t.Fatalf("kpi_usage_unit = %q, want GB", got)
	}
	if !settingsVal.Grid.IsNull() {
		t.Fatalf("grid = %v, want null", settingsVal.Grid)
	}
	if got := value.UpdatedAt.ValueString(); got != "2026-01-02T00:00:00Z" {
		t.Fatalf("updated_at = %q, want 2026-01-02T00:00:00Z", got)
	}
}

func TestDashboardDataSourceValueFromPayload_gridSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dateBin := "day"
	dateInterval := "this_month"

	value, diags := dashboardDataSourceValueFromPayload(ctx, &modelsv2.Dashboard{
		CreatedAt:         "2026-01-01T00:00:00Z",
		DateBin:           &dateBin,
		DateInterval:      &dateInterval,
		SavedFilterTokens: []string{},
		Title:             "grid-dashboard",
		Token:             "dshbrd_test",
		UpdatedAt:         "2026-01-02T00:00:00Z",
		WorkspaceToken:    "wrkspc_test",
		Widgets: []*modelsv2.DashboardWidget{
			{
				Title:           "Spend Chart",
				Token:           "dshbrd_wdgt_test",
				WidgetableToken: "rprt_test",
				Settings: &modelsv2.DashboardWidgetSettings{
					DisplayType: "chart",
					Grid: &modelsv2.DashboardWidgetGridLayout{
						X: 3,
						Y: 1,
						W: 9,
						H: 2,
					},
				},
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("dashboardDataSourceValueFromPayload diagnostics: %v", diags.Errors())
	}

	var widgets []datasource_dashboards.WidgetsValue
	if d := value.Widgets.ElementsAs(ctx, &widgets, false); d.HasError() {
		t.Fatalf("ElementsAs diagnostics: %v", d.Errors())
	}
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}

	settings, d := datasource_dashboards.SettingsType{}.ValueFromObject(ctx, widgets[0].Settings)
	if d.HasError() {
		t.Fatalf("ValueFromObject diagnostics: %v", d.Errors())
	}
	settingsVal, ok := settings.(datasource_dashboards.SettingsValue)
	if !ok {
		t.Fatalf("expected SettingsValue, got %T", settings)
	}
	grid, d := datasource_dashboards.GridType{}.ValueFromObject(ctx, settingsVal.Grid)
	if d.HasError() {
		t.Fatalf("grid ValueFromObject diagnostics: %v", d.Errors())
	}
	gridVal, ok := grid.(datasource_dashboards.GridValue)
	if !ok {
		t.Fatalf("expected GridValue, got %T", grid)
	}
	if got := gridVal.X.ValueInt64(); got != 3 {
		t.Fatalf("grid.x = %d, want 3", got)
	}
	if got := gridVal.Y.ValueInt64(); got != 1 {
		t.Fatalf("grid.y = %d, want 1", got)
	}
	if got := gridVal.W.ValueInt64(); got != 9 {
		t.Fatalf("grid.w = %d, want 9", got)
	}
	if got := gridVal.H.ValueInt64(); got != 2 {
		t.Fatalf("grid.h = %d, want 2", got)
	}
}

func TestDashboardDataSourceValueFromPayload_freeTextWidget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dateBin := "day"
	dateInterval := "this_month"
	content := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "Hello"},
				},
			},
		},
	}

	value, diags := dashboardDataSourceValueFromPayload(ctx, &modelsv2.Dashboard{
		CreatedAt:         "2026-01-01T00:00:00Z",
		DateBin:           &dateBin,
		DateInterval:      &dateInterval,
		SavedFilterTokens: []string{},
		Title:             "free-text-dashboard",
		Token:             "dshbrd_test",
		UpdatedAt:         "2026-01-02T00:00:00Z",
		WorkspaceToken:    "wrkspc_test",
		Widgets: []*modelsv2.DashboardWidget{
			{
				Title:          "Notes",
				Token:          "dshbrd_wdgt_test",
				WidgetableType: "free_text",
				Content:        content,
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("dashboardDataSourceValueFromPayload diagnostics: %v", diags.Errors())
	}

	var widgets []datasource_dashboards.WidgetsValue
	if d := value.Widgets.ElementsAs(ctx, &widgets, false); d.HasError() {
		t.Fatalf("ElementsAs diagnostics: %v", d.Errors())
	}
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if got := widgets[0].WidgetableType.ValueString(); got != "free_text" {
		t.Fatalf("widgetable_type = %q, want free_text", got)
	}
	if !widgets[0].WidgetableToken.IsNull() {
		t.Fatalf("widgetable_token = %v, want null", widgets[0].WidgetableToken)
	}
	if widgets[0].Content.IsNull() || widgets[0].Content.ValueString() == "" {
		t.Fatal("expected content")
	}
	if !widgets[0].Settings.IsNull() {
		t.Fatalf("settings = %v, want null", widgets[0].Settings)
	}
}
