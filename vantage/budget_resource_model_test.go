package vantage

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestBudgetPeriodCadenceCreateMapping(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := testBudgetPeriodCadenceValue(map[string]attr.Value{
		"starts_at":      types.StringValue("2024-01-22"),
		"interval_count": types.Int64Value(2),
		"interval_unit":  types.StringValue("week"),
	})
	if diagnostics.HasError() {
		t.Fatalf("building cadence value: %v", diagnostics)
	}

	model := toCreateModel(context.Background(), &diag.Diagnostics{}, budgetModel{
		Name:          types.StringValue("Test Budget"),
		PeriodCadence: cadence,
	})

	if model.PeriodCadence == nil {
		t.Fatal("expected period cadence in create model")
	}
	if got := model.PeriodCadence.StartsAt.String(); got != "2024-01-22" {
		t.Errorf("starts_at = %q, want %q", got, "2024-01-22")
	}
	if got := model.PeriodCadence.IntervalCount; got != 2 {
		t.Errorf("interval_count = %d, want 2", got)
	}
	if got := model.PeriodCadence.IntervalUnit; got != "week" {
		t.Errorf("interval_unit = %q, want %q", got, "week")
	}
}

func TestBudgetConfigRequiresCadenceStart(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := testBudgetPeriodCadenceValue(map[string]attr.Value{
		"starts_at":      types.StringNull(),
		"interval_count": types.Int64Value(1),
		"interval_unit":  types.StringValue("month"),
	})
	if diagnostics.HasError() {
		t.Fatalf("building cadence value: %v", diagnostics)
	}

	validateBudgetConfig(budgetModel{
		PeriodCadence: cadence,
	}, &diagnostics)

	if !diagnostics.HasError() {
		t.Fatal("expected a configured cadence without starts_at to fail validation")
	}
}

func TestBudgetConfigRejectsCadenceForCompoundBudget(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := testBudgetPeriodCadenceValue(map[string]attr.Value{
		"starts_at":      types.StringValue("2024-01-22"),
		"interval_count": types.Int64Value(1),
		"interval_unit":  types.StringValue("month"),
	})
	if diagnostics.HasError() {
		t.Fatalf("building cadence value: %v", diagnostics)
	}

	children, listDiagnostics := types.ListValueFrom(
		context.Background(),
		types.StringType,
		[]string{"bdgt_child"},
	)
	if listDiagnostics.HasError() {
		t.Fatalf("building child budget list: %v", listDiagnostics)
	}

	validateBudgetConfig(budgetModel{
		PeriodCadence:     cadence,
		ChildBudgetTokens: children,
	}, &diagnostics)

	if !diagnostics.HasError() {
		t.Fatal("expected period_cadence with child_budget_tokens to fail validation")
	}
}

func TestBudgetConfigDefersUnknownCadenceStart(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := testBudgetPeriodCadenceValue(map[string]attr.Value{
		"starts_at":      types.StringUnknown(),
		"interval_count": types.Int64Value(1),
		"interval_unit":  types.StringValue("month"),
	})
	if diagnostics.HasError() {
		t.Fatalf("building cadence value: %v", diagnostics)
	}

	validateBudgetConfig(budgetModel{PeriodCadence: cadence}, &diagnostics)

	if diagnostics.HasError() {
		t.Fatalf("unknown starts_at should defer validation: %v", diagnostics)
	}
}

func TestBudgetConfigRejectsUnitWithoutUsageType(t *testing.T) {
	t.Parallel()

	var diagnostics diag.Diagnostics
	validateBudgetConfig(budgetModel{
		Type: types.StringValue("cost"),
		Unit: types.StringValue("GB-Hours"),
	}, &diagnostics)

	if !diagnostics.HasError() {
		t.Fatal("expected unit with cost budget type to fail validation")
	}
}

// A block that sets only some fields must still send them. The interval fields
// carry omitempty, so leaving them at zero omits them from the request instead
// of overwriting the cadence with placeholders.
func TestBudgetPeriodCadenceSendsPartialConfig(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := testBudgetPeriodCadenceValue(map[string]attr.Value{
		"starts_at":      types.StringValue("2024-01-22"),
		"interval_count": types.Int64Unknown(),
		"interval_unit":  types.StringUnknown(),
	})
	if diagnostics.HasError() {
		t.Fatalf("building cadence value: %v", diagnostics)
	}

	model := toCreateModel(context.Background(), &diag.Diagnostics{}, budgetModel{
		Name:          types.StringValue("Test Budget"),
		PeriodCadence: cadence,
	})
	if model.PeriodCadence == nil {
		t.Fatal("expected the configured period cadence to be sent")
	}
	if got := model.PeriodCadence.StartsAt.String(); got != "2024-01-22" {
		t.Errorf("starts_at = %q, want %q", got, "2024-01-22")
	}

	encoded, err := json.Marshal(model.PeriodCadence)
	if err != nil {
		t.Fatalf("marshalling cadence: %v", err)
	}
	for _, field := range []string{"interval_count", "interval_unit"} {
		if strings.Contains(string(encoded), field) {
			t.Errorf("request contains %s, want it omitted: %s", field, encoded)
		}
	}
}

func TestBudgetPeriodCadenceUpdateOmitsDerivedCadence(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := testBudgetPeriodCadenceValue(map[string]attr.Value{
		"starts_at":      types.StringValue("2024-01-01"),
		"interval_count": types.Int64Value(1),
		"interval_unit":  types.StringValue("month"),
	})
	if diagnostics.HasError() {
		t.Fatalf("building cadence value: %v", diagnostics)
	}

	model := toUpdateModel(context.Background(), &diag.Diagnostics{}, budgetModel{
		Name:          types.StringValue("Test Budget"),
		PeriodCadence: cadence,
	}, resource_budget.NewPeriodCadenceValueNull())
	if model.PeriodCadence != nil {
		t.Fatalf("expected derived period cadence to be omitted from update, got %+v", model.PeriodCadence)
	}
}

func TestBudgetPeriodCadenceResponseMapping(t *testing.T) {
	t.Parallel()

	startsAt := "2024-01-22"
	value, diagnostics := periodCadenceFromPayload(&modelsv2.PeriodCadence{
		StartsAt:      &startsAt,
		IntervalCount: 2,
		IntervalUnit:  "week",
	})
	if diagnostics.HasError() {
		t.Fatalf("mapping cadence payload: %v", diagnostics)
	}

	objectValue, diagnostics := value.ToObjectValue(context.Background())
	if diagnostics.HasError() {
		t.Fatalf("converting cadence payload: %v", diagnostics)
	}
	attributes := objectValue.Attributes()
	if got := attributes["starts_at"].(types.String).ValueString(); got != startsAt {
		t.Errorf("starts_at = %q, want %q", got, startsAt)
	}
	if got := attributes["interval_count"].(types.Int64).ValueInt64(); got != 2 {
		t.Errorf("interval_count = %d, want 2", got)
	}
	if got := attributes["interval_unit"].(types.String).ValueString(); got != "week" {
		t.Errorf("interval_unit = %q, want %q", got, "week")
	}
}

func TestBudgetTypeAndUnitCreateMapping(t *testing.T) {
	t.Parallel()

	model := toCreateModel(context.Background(), &diag.Diagnostics{}, budgetModel{
		Name: types.StringValue("Test Budget"),
		Type: types.StringValue("usage"),
		Unit: types.StringValue("GB-Hours"),
	})

	if got := model.Type; got != "usage" {
		t.Errorf("type = %q, want %q", got, "usage")
	}
	if got := model.Unit; got != "GB-Hours" {
		t.Errorf("unit = %q, want %q", got, "GB-Hours")
	}
}

func TestBudgetTypeAndUnitUpdateMapping(t *testing.T) {
	t.Parallel()

	model := toUpdateModel(context.Background(), &diag.Diagnostics{}, budgetModel{
		Name: types.StringValue("Test Budget"),
		Type: types.StringValue("usage"),
		Unit: types.StringValue("GB-Hours"),
	}, resource_budget.NewPeriodCadenceValueNull())

	if got := model.Type; got != "usage" {
		t.Errorf("type = %q, want %q", got, "usage")
	}
	if model.Unit == nil {
		t.Fatal("expected unit to be sent")
	}
	if got := *model.Unit; got != "GB-Hours" {
		t.Errorf("unit = %q, want %q", got, "GB-Hours")
	}
}

func TestBudgetTypeChangeToCostOmitsStaleUsageUnit(t *testing.T) {
	t.Parallel()

	model := toUpdateModel(context.Background(), &diag.Diagnostics{}, budgetModel{
		Name: types.StringValue("Test Budget"),
		Type: types.StringValue("cost"),
		Unit: types.StringValue("GB-Hours"),
	}, resource_budget.NewPeriodCadenceValueNull())

	if got := model.Type; got != "cost" {
		t.Errorf("type = %q, want %q", got, "cost")
	}
	if model.Unit != nil {
		t.Fatalf("expected stale usage unit to be omitted, got %q", *model.Unit)
	}
}

func TestBudgetTypeAndUnitResponseMapping(t *testing.T) {
	t.Parallel()

	unit := "GB-Hours"
	var model budgetModel
	diagnostics := applyBudgetPayload(context.Background(), false, &modelsv2.Budget{
		Token:          "bdgt_test",
		CreatedAt:      "2026-01-01T00:00:00Z",
		WorkspaceToken: "wrkspc_test",
		Type:           "usage",
		Unit:           &unit,
	}, &model)

	if diagnostics.HasError() {
		t.Fatalf("mapping budget payload: %v", diagnostics)
	}
	if got := model.Type.ValueString(); got != "usage" {
		t.Errorf("type = %q, want %q", got, "usage")
	}
	if got := model.Unit.ValueString(); got != unit {
		t.Errorf("unit = %q, want %q", got, unit)
	}
}

func testBudgetPeriodCadenceValue(attributes map[string]attr.Value) (resource_budget.PeriodCadenceValue, diag.Diagnostics) {
	if _, ok := attributes["starts_at"]; !ok {
		attributes["starts_at"] = types.StringUnknown()
	}
	if _, ok := attributes["interval_count"]; !ok {
		attributes["interval_count"] = types.Int64Unknown()
	}
	if _, ok := attributes["interval_unit"]; !ok {
		attributes["interval_unit"] = types.StringUnknown()
	}

	return resource_budget.NewPeriodCadenceValue(periodCadenceAttrTypes, attributes)
}
