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
	if got := value.UpdatedAt.ValueString(); got != "2026-01-02T00:00:00Z" {
		t.Fatalf("updated_at = %q, want 2026-01-02T00:00:00Z", got)
	}
}
