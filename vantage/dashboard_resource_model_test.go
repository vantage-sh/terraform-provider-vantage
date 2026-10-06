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
	if !settingsVal.Grid.IsNull() {
		t.Fatalf("grid = %v, want null", settingsVal.Grid)
	}
}

func TestDashboardModel_applyPayload_gridSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := &dashboardModel{}
	diags := model.applyPayload(ctx, &modelsv2.Dashboard{
		Title:          "grid-dashboard",
		Token:          "dshbrd_test",
		WorkspaceToken: "wrkspc_test",
		Widgets: []*modelsv2.DashboardWidget{
			{
				Title:           "Spend Chart",
				Token:           "dshbrd_wdgt_test",
				WidgetableToken: "rprt_test",
				Settings: &modelsv2.DashboardWidgetSettings{
					DisplayType: "chart",
					Grid: &modelsv2.DashboardWidgetGridLayout{
						X: 0,
						Y: 2,
						W: 6,
						H: 4,
					},
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
	grid, d := resource_dashboard.GridType{}.ValueFromObject(ctx, settingsVal.Grid)
	if d.HasError() {
		t.Fatalf("grid ValueFromObject diagnostics: %v", d.Errors())
	}
	gridVal, ok := grid.(resource_dashboard.GridValue)
	if !ok {
		t.Fatalf("expected GridValue, got %T", grid)
	}
	if got := gridVal.X.ValueInt64(); got != 0 {
		t.Fatalf("grid.x = %d, want 0", got)
	}
	if got := gridVal.Y.ValueInt64(); got != 2 {
		t.Fatalf("grid.y = %d, want 2", got)
	}
	if got := gridVal.W.ValueInt64(); got != 6 {
		t.Fatalf("grid.w = %d, want 6", got)
	}
	if got := gridVal.H.ValueInt64(); got != 4 {
		t.Fatalf("grid.h = %d, want 4", got)
	}
}

func TestDashboardModel_toCreate_kpiSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settingsAttrTypes := resource_dashboard.SettingsValue{}.AttributeTypes(ctx)
	settingsVal, d := resource_dashboard.NewSettingsValue(settingsAttrTypes, map[string]attr.Value{
		"display_type":    types.StringValue("kpi"),
		"grid":            types.ObjectNull(resource_dashboard.GridValue{}.AttributeTypes(ctx)),
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
		"content":          types.StringNull(),
		"settings":         settingsObj,
		"title":            types.StringValue("Usage KPI"),
		"token":            types.StringNull(),
		"widgetable_token": types.StringValue("rprt_test"),
		"widgetable_type":  types.StringNull(),
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
	if settings.Grid != nil {
		t.Fatalf("grid = %#v, want nil", settings.Grid)
	}
}

func TestDashboardModel_toCreate_gridSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	gridAttrTypes := resource_dashboard.GridValue{}.AttributeTypes(ctx)
	gridVal, d := resource_dashboard.NewGridValue(gridAttrTypes, map[string]attr.Value{
		"x": types.Int64Value(0),
		"y": types.Int64Value(1),
		"w": types.Int64Value(12),
		"h": types.Int64Value(3),
	})
	if d.HasError() {
		t.Fatalf("NewGridValue diagnostics: %v", d.Errors())
	}
	gridObj, d := gridVal.ToObjectValue(ctx)
	if d.HasError() {
		t.Fatalf("grid ToObjectValue diagnostics: %v", d.Errors())
	}

	settingsAttrTypes := resource_dashboard.SettingsValue{}.AttributeTypes(ctx)
	settingsVal, d := resource_dashboard.NewSettingsValue(settingsAttrTypes, map[string]attr.Value{
		"display_type":    types.StringValue("table"),
		"grid":            gridObj,
		"kpi_calculation": types.StringNull(),
		"kpi_type":        types.StringNull(),
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
		"content":          types.StringNull(),
		"settings":         settingsObj,
		"title":            types.StringValue("Table Widget"),
		"token":            types.StringNull(),
		"widgetable_token": types.StringValue("rprt_test"),
		"widgetable_type":  types.StringNull(),
	})
	if d.HasError() {
		t.Fatalf("NewWidgetsValue diagnostics: %v", d.Errors())
	}

	widgets, d := types.ListValueFrom(ctx, resource_dashboard.WidgetsValue{}.Type(ctx), []resource_dashboard.WidgetsValue{widgetVal})
	if d.HasError() {
		t.Fatalf("ListValueFrom diagnostics: %v", d.Errors())
	}

	model := &dashboardModel{
		Title:          types.StringValue("grid-dashboard"),
		WorkspaceToken: types.StringValue("wrkspc_test"),
		Widgets:        widgets,
	}

	var diags diag.Diagnostics
	payload := model.toCreate(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("toCreate diagnostics: %v", diags.Errors())
	}
	if len(payload.Widgets) != 1 || payload.Widgets[0].Settings == nil || payload.Widgets[0].Settings.Grid == nil {
		t.Fatal("expected widget grid settings")
	}
	grid := payload.Widgets[0].Settings.Grid
	if got := *grid.X; got != 0 {
		t.Fatalf("grid.x = %d, want 0", got)
	}
	if got := *grid.Y; got != 1 {
		t.Fatalf("grid.y = %d, want 1", got)
	}
	if got := *grid.W; got != 12 {
		t.Fatalf("grid.w = %d, want 12", got)
	}
	if got := *grid.H; got != 3 {
		t.Fatalf("grid.h = %d, want 3", got)
	}
}

func TestDashboardModel_toUpdate_kpiSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settingsAttrTypes := resource_dashboard.SettingsValue{}.AttributeTypes(ctx)
	settingsVal, d := resource_dashboard.NewSettingsValue(settingsAttrTypes, map[string]attr.Value{
		"display_type":    types.StringValue("kpi"),
		"grid":            types.ObjectNull(resource_dashboard.GridValue{}.AttributeTypes(ctx)),
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
		"content":          types.StringNull(),
		"settings":         settingsObj,
		"title":            types.StringValue("Spend KPI"),
		"token":            types.StringNull(),
		"widgetable_token": types.StringValue("rprt_test"),
		"widgetable_type":  types.StringNull(),
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
	if settings.Grid != nil {
		t.Fatalf("grid = %#v, want nil", settings.Grid)
	}
}

func TestDashboardModel_applyPayload_freeTextWidget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
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

	model := &dashboardModel{}
	diags := model.applyPayload(ctx, &modelsv2.Dashboard{
		Title:          "free-text-dashboard",
		Token:          "dshbrd_test",
		WorkspaceToken: "wrkspc_test",
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
		t.Fatalf("applyPayload diagnostics: %v", diags.Errors())
	}

	var widgets []resource_dashboard.WidgetsValue
	if d := model.Widgets.ElementsAs(ctx, &widgets, false); d.HasError() {
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

func TestDashboardModel_toCreate_freeTextWidget(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	contentJSON := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello"}]}]}`

	widgetAttrTypes := resource_dashboard.WidgetsValue{}.AttributeTypes(ctx)
	widgetVal, d := resource_dashboard.NewWidgetsValue(widgetAttrTypes, map[string]attr.Value{
		"content":          types.StringValue(contentJSON),
		"settings":         types.ObjectNull(resource_dashboard.SettingsValue{}.AttributeTypes(ctx)),
		"title":            types.StringValue("Notes"),
		"token":            types.StringNull(),
		"widgetable_token": types.StringNull(),
		"widgetable_type":  types.StringValue("free_text"),
	})
	if d.HasError() {
		t.Fatalf("NewWidgetsValue diagnostics: %v", d.Errors())
	}

	widgets, d := types.ListValueFrom(ctx, resource_dashboard.WidgetsValue{}.Type(ctx), []resource_dashboard.WidgetsValue{widgetVal})
	if d.HasError() {
		t.Fatalf("ListValueFrom diagnostics: %v", d.Errors())
	}

	model := &dashboardModel{
		Title:          types.StringValue("free-text-dashboard"),
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
	widget := payload.Widgets[0]
	if got := widget.WidgetableType; got != "free_text" {
		t.Fatalf("widgetable_type = %q, want free_text", got)
	}
	if got := widget.WidgetableToken; got != "" {
		t.Fatalf("widgetable_token = %q, want empty", got)
	}
	if widget.Content == nil || widget.Content.Type == nil || *widget.Content.Type != "doc" {
		t.Fatalf("content = %#v, want TipTap doc", widget.Content)
	}
	if widget.Settings != nil {
		t.Fatalf("settings = %#v, want nil", widget.Settings)
	}
}

func TestDashboardModel_toCreate_invalidFreeTextContent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	widgetAttrTypes := resource_dashboard.WidgetsValue{}.AttributeTypes(ctx)
	widgetVal, d := resource_dashboard.NewWidgetsValue(widgetAttrTypes, map[string]attr.Value{
		"content":          types.StringValue(`{"content":[]}`),
		"settings":         types.ObjectNull(resource_dashboard.SettingsValue{}.AttributeTypes(ctx)),
		"title":            types.StringValue("Notes"),
		"token":            types.StringNull(),
		"widgetable_token": types.StringNull(),
		"widgetable_type":  types.StringValue("free_text"),
	})
	if d.HasError() {
		t.Fatalf("NewWidgetsValue diagnostics: %v", d.Errors())
	}

	widgets, d := types.ListValueFrom(ctx, resource_dashboard.WidgetsValue{}.Type(ctx), []resource_dashboard.WidgetsValue{widgetVal})
	if d.HasError() {
		t.Fatalf("ListValueFrom diagnostics: %v", d.Errors())
	}

	model := &dashboardModel{
		Title:          types.StringValue("free-text-dashboard"),
		WorkspaceToken: types.StringValue("wrkspc_test"),
		Widgets:        widgets,
	}

	var diags diag.Diagnostics
	payload := model.toCreate(ctx, &diags)
	if !diags.HasError() {
		t.Fatal("expected diagnostics for invalid content")
	}
	if payload != nil {
		t.Fatalf("expected nil payload, got %#v", payload)
	}
}
