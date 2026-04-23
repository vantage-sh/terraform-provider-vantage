package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
