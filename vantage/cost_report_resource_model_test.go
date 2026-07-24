package vantage

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_cost_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

// TestApplyPayload_ClearsStaleChartSettingsAndSettings verifies that when
// state has non-null `chart_settings` and `settings` but the API payload
// returns nil for both, applyPayload resets the state values to null instead
// of silently retaining stale data.
//
// Regression test for the drift bug where only two branches existed:
//   - populate (required state non-null AND payload non-nil)
//   - null-out (required state null/unknown)
//
// Leaving the (state non-null, payload nil) case unhandled, which let stale
// state perpetuate across Read calls.
func TestApplyPayload_ClearsStaleChartSettingsAndSettings(t *testing.T) {
	ctx := context.Background()

	xAxis, d := types.ListValue(types.StringType, []attr.Value{types.StringValue("date")})
	if d.HasError() {
		t.Fatalf("unexpected diags building xAxis: %v", d)
	}
	csValue, d := resource_cost_report.NewChartSettingsValue(
		resource_cost_report.ChartSettingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"x_axis_dimension": xAxis,
			"y_axis_dimension": types.StringValue("cost"),
		},
	)
	if d.HasError() {
		t.Fatalf("unexpected diags building chart_settings: %v", d)
	}

	sValue, d := resource_cost_report.NewSettingsValue(
		resource_cost_report.SettingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"aggregate_by":         types.StringValue("cost"),
			"amortize":             types.BoolValue(true),
			"complete_period":      types.BoolValue(true),
			"include_credits":      types.BoolValue(true),
			"include_discounts":    types.BoolValue(true),
			"include_refunds":      types.BoolValue(true),
			"include_tax":          types.BoolValue(true),
			"show_previous_period": types.BoolValue(true),
			"unallocated":          types.BoolValue(true),
		},
	)
	if d.HasError() {
		t.Fatalf("unexpected diags building settings: %v", d)
	}

	model := &costReportModel{
		ChartSettings: csValue,
		Settings:      sValue,
	}

	payload := &modelsv2.CostReport{
		Token:             "rprt_test",
		Title:             "test",
		WorkspaceToken:    "wrkspc_test",
		DateInterval:      "last_7_days",
		ChartType:         "bar",
		DateBin:           "day",
		CreatedAt:         "2025-01-01T00:00:00Z",
		SavedFilterTokens: []string{},
		ChartSettings:     nil,
		Settings:          nil,
	}

	diags := model.applyPayload(ctx, payload)
	if diags.HasError() {
		t.Fatalf("applyPayload returned errors: %v", diags)
	}

	if !model.ChartSettings.IsNull() {
		t.Errorf("expected ChartSettings to be null after API returned nil, got %#v", model.ChartSettings)
	}
	if !model.Settings.IsNull() {
		t.Errorf("expected Settings to be null after API returned nil, got %#v", model.Settings)
	}
}

// TestApplyPayload_PopulatesChartSettingsAndSettingsFromAPI verifies the
// positive path: when both state has the blocks configured and the API
// returns values, applyPayload maps the API values into state.
func TestApplyPayload_PopulatesChartSettingsAndSettingsFromAPI(t *testing.T) {
	ctx := context.Background()

	xAxis, d := types.ListValue(types.StringType, []attr.Value{types.StringValue("date")})
	if d.HasError() {
		t.Fatalf("unexpected diags building xAxis: %v", d)
	}
	csValue, d := resource_cost_report.NewChartSettingsValue(
		resource_cost_report.ChartSettingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"x_axis_dimension": xAxis,
			"y_axis_dimension": types.StringValue("cost"),
		},
	)
	if d.HasError() {
		t.Fatalf("unexpected diags building chart_settings: %v", d)
	}

	sValue, d := resource_cost_report.NewSettingsValue(
		resource_cost_report.SettingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"aggregate_by":         types.StringValue("cost"),
			"amortize":             types.BoolValue(false),
			"complete_period":      types.BoolValue(false),
			"include_credits":      types.BoolValue(false),
			"include_discounts":    types.BoolValue(false),
			"include_refunds":      types.BoolValue(false),
			"include_tax":          types.BoolValue(false),
			"show_previous_period": types.BoolValue(false),
			"unallocated":          types.BoolValue(false),
		},
	)
	if d.HasError() {
		t.Fatalf("unexpected diags building settings: %v", d)
	}

	model := &costReportModel{
		ChartSettings: csValue,
		Settings:      sValue,
	}

	aggregateBy := "usage"
	amortize := true
	completePeriod := true
	includeCredits := true

	payload := &modelsv2.CostReport{
		Token:             "rprt_test",
		Title:             "test",
		WorkspaceToken:    "wrkspc_test",
		DateInterval:      "last_7_days",
		ChartType:         "bar",
		DateBin:           "day",
		CreatedAt:         "2025-01-01T00:00:00Z",
		SavedFilterTokens: []string{},
		ChartSettings: &modelsv2.ChartSettings{
			XAxisDimension: []string{"service"},
			YAxisDimension: "usage",
		},
		Settings: &modelsv2.CostReportSettings{
			AggregateBy:    &aggregateBy,
			Amortize:       &amortize,
			CompletePeriod: &completePeriod,
			IncludeCredits: &includeCredits,
		},
	}

	diags := model.applyPayload(ctx, payload)
	if diags.HasError() {
		t.Fatalf("applyPayload returned errors: %v", diags)
	}

	if model.ChartSettings.IsNull() || model.ChartSettings.IsUnknown() {
		t.Fatalf("expected ChartSettings populated, got null/unknown")
	}
	if got, want := model.ChartSettings.YAxisDimension.ValueString(), "usage"; got != want {
		t.Errorf("ChartSettings.YAxisDimension = %q, want %q", got, want)
	}

	if model.Settings.IsNull() || model.Settings.IsUnknown() {
		t.Fatalf("expected Settings populated, got null/unknown")
	}
	if got, want := model.Settings.AggregateBy.ValueString(), "usage"; got != want {
		t.Errorf("Settings.AggregateBy = %q, want %q", got, want)
	}
	if !model.Settings.Amortize.ValueBool() {
		t.Errorf("Settings.Amortize = false, want true")
	}
	if !model.Settings.CompletePeriod.ValueBool() {
		t.Errorf("Settings.CompletePeriod = false, want true")
	}
}

// TestApplyPayload_NullStateStaysNull verifies that when state has
// chart_settings/settings null (config omitted the block), applyPayload keeps
// them null even when the API reports default values, to avoid capturing
// server-side defaults and generating drift.
func TestApplyPayload_NullStateStaysNull(t *testing.T) {
	ctx := context.Background()

	model := &costReportModel{
		ChartSettings: resource_cost_report.NewChartSettingsValueNull(),
		Settings:      resource_cost_report.NewSettingsValueNull(),
	}

	aggregateBy := "cost"

	payload := &modelsv2.CostReport{
		Token:             "rprt_test",
		Title:             "test",
		WorkspaceToken:    "wrkspc_test",
		DateInterval:      "last_7_days",
		ChartType:         "bar",
		DateBin:           "day",
		CreatedAt:         "2025-01-01T00:00:00Z",
		SavedFilterTokens: []string{},
		ChartSettings: &modelsv2.ChartSettings{
			XAxisDimension: []string{"date"},
			YAxisDimension: "cost",
		},
		Settings: &modelsv2.CostReportSettings{
			AggregateBy: &aggregateBy,
		},
	}

	diags := model.applyPayload(ctx, payload)
	if diags.HasError() {
		t.Fatalf("applyPayload returned errors: %v", diags)
	}

	if !model.ChartSettings.IsNull() {
		t.Errorf("expected ChartSettings to remain null when config omits the block")
	}
	if !model.Settings.IsNull() {
		t.Errorf("expected Settings to remain null when config omits the block")
	}
}

func TestCostReportModel_toUpdateModelMapsDefaultForecast(t *testing.T) {
	ctx := context.Background()
	defaultForecast, d := resource_cost_report.NewDefaultForecastValue(
		resource_cost_report.DefaultForecastValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"kind":                  types.StringValue("report_forecast"),
			"report_forecast_token": types.StringValue("rprt_fcst_test"),
		},
	)
	if d.HasError() {
		t.Fatalf("unexpected diags building default_forecast: %v", d)
	}

	model := &costReportModel{
		BusinessMetricTokensWithMetadata: types.ListNull(types.ObjectType{
			AttrTypes: resource_cost_report.BusinessMetricTokensWithMetadataValue{}.AttributeTypes(ctx),
		}),
		ChartSettings:   resource_cost_report.NewChartSettingsValueNull(),
		DefaultForecast: defaultForecast,
		Settings:        resource_cost_report.NewSettingsValueNull(),
	}

	var diags diag.Diagnostics
	update := model.toUpdateModel(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("toUpdateModel returned errors: %v", diags)
	}
	if update.DefaultForecast == nil {
		t.Fatal("expected DefaultForecast in update payload")
	}
	if got, want := *update.DefaultForecast.Kind, "report_forecast"; got != want {
		t.Errorf("DefaultForecast.Kind = %q, want %q", got, want)
	}
	if got, want := update.DefaultForecast.ReportForecastToken, "rprt_fcst_test"; got != want {
		t.Errorf("DefaultForecast.ReportForecastToken = %q, want %q", got, want)
	}
	if update.BusinessMetricTokensWithMetadata == nil || len(update.BusinessMetricTokensWithMetadata) != 0 {
		t.Errorf("BusinessMetricTokensWithMetadata = %#v, want non-nil empty slice", update.BusinessMetricTokensWithMetadata)
	}
}

func TestCostReportModelMapsBusinessMetricMetadata(t *testing.T) {
	ctx := context.Background()
	teamLabels, d := types.ListValueFrom(ctx, types.StringType, []string{"platform", "finops"})
	if d.HasError() {
		t.Fatalf("unexpected diags building team labels: %v", d)
	}
	labelFilters, d := types.MapValue(
		types.ListType{ElemType: types.StringType},
		map[string]attr.Value{"team": teamLabels},
	)
	if d.HasError() {
		t.Fatalf("unexpected diags building label_filters: %v", d)
	}
	labelFilter, d := types.ListValueFrom(ctx, types.StringType, []string{})
	if d.HasError() {
		t.Fatalf("unexpected diags building label_filter: %v", d)
	}
	businessMetric, d := types.ObjectValue(
		resource_cost_report.BusinessMetricTokensWithMetadataValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"business_metric_token": types.StringValue("bsnss_mtrc_test"),
			"calculation_type":      types.StringValue("gross_margin"),
			"label":                 types.StringValue("Platform Gross Margin"),
			"label_filter":          labelFilter,
			"label_filters":         labelFilters,
			"unit_scale":            types.StringValue("per_hundred"),
		},
	)
	if d.HasError() {
		t.Fatalf("unexpected diags building business metric metadata: %v", d)
	}
	businessMetrics, d := types.ListValue(
		types.ObjectType{
			AttrTypes: resource_cost_report.BusinessMetricTokensWithMetadataValue{}.AttributeTypes(ctx),
		},
		[]attr.Value{businessMetric},
	)
	if d.HasError() {
		t.Fatalf("unexpected diags building business metric list: %v", d)
	}

	model := &costReportModel{
		BusinessMetricTokensWithMetadata: businessMetrics,
		ChartSettings:                    resource_cost_report.NewChartSettingsValueNull(),
		DefaultForecast:                  resource_cost_report.NewDefaultForecastValueNull(),
		Settings:                         resource_cost_report.NewSettingsValueNull(),
	}
	var createDiags diag.Diagnostics
	create := model.toCreateModel(ctx, &createDiags)
	if createDiags.HasError() {
		t.Fatalf("toCreateModel returned errors: %v", createDiags)
	}
	assertCreateBusinessMetricMetadata(t, create.BusinessMetricTokensWithMetadata)

	var updateDiags diag.Diagnostics
	update := model.toUpdateModel(ctx, &updateDiags)
	if updateDiags.HasError() {
		t.Fatalf("toUpdateModel returned errors: %v", updateDiags)
	}
	if len(update.BusinessMetricTokensWithMetadata) != 1 {
		t.Fatalf("update business metric count = %d, want 1", len(update.BusinessMetricTokensWithMetadata))
	}
	updateItem := update.BusinessMetricTokensWithMetadata[0]
	if got, want := *updateItem.CalculationType, "gross_margin"; got != want {
		t.Errorf("update CalculationType = %q, want %q", got, want)
	}
	if got, want := updateItem.Label, "Platform Gross Margin"; got != want {
		t.Errorf("update Label = %q, want %q", got, want)
	}
	if got, want := updateItem.LabelFilters["team"], []string{"platform", "finops"}; !slices.Equal(got, want) {
		t.Errorf("update LabelFilters[team] = %v, want %v", got, want)
	}
}

func assertCreateBusinessMetricMetadata(
	t *testing.T,
	items []*modelsv2.CreateCostReportBusinessMetricTokensWithMetadataItems0,
) {
	t.Helper()
	if len(items) != 1 {
		t.Fatalf("create business metric count = %d, want 1", len(items))
	}
	item := items[0]
	if got, want := *item.CalculationType, "gross_margin"; got != want {
		t.Errorf("create CalculationType = %q, want %q", got, want)
	}
	if got, want := item.Label, "Platform Gross Margin"; got != want {
		t.Errorf("create Label = %q, want %q", got, want)
	}
	if got, want := item.LabelFilters["team"], []string{"platform", "finops"}; !slices.Equal(got, want) {
		t.Errorf("create LabelFilters[team] = %v, want %v", got, want)
	}
}
