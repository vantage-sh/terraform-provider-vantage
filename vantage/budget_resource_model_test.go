package vantage

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestBudgetPeriodCadenceCreateMapping(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := types.ObjectValue(periodCadenceAttrTypes, map[string]attr.Value{
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

func TestBudgetPeriodCadenceUpdateClearsStartsAt(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := types.ObjectValue(periodCadenceAttrTypes, map[string]attr.Value{
		"starts_at":      types.StringValue(""),
		"interval_count": types.Int64Value(1),
		"interval_unit":  types.StringValue("month"),
	})
	if diagnostics.HasError() {
		t.Fatalf("building cadence value: %v", diagnostics)
	}

	model := toUpdateModel(context.Background(), &diag.Diagnostics{}, budgetModel{
		Name:          types.StringValue("Test Budget"),
		PeriodCadence: cadence,
	}, cadence)

	payload, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("marshaling update model: %v", err)
	}

	if !strings.Contains(string(payload), `"starts_at":null`) {
		t.Errorf("payload = %s, want starts_at:null", payload)
	}
}

func TestBudgetPeriodCadenceSkipsUnknownNestedFields(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := types.ObjectValue(periodCadenceAttrTypes, map[string]attr.Value{
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
	if model.PeriodCadence != nil {
		t.Fatalf("expected no period cadence while nested fields are unknown, got %+v", model.PeriodCadence)
	}
}

func TestBudgetPeriodCadenceUpdateOmitsDerivedCadence(t *testing.T) {
	t.Parallel()

	cadence, diagnostics := types.ObjectValue(periodCadenceAttrTypes, map[string]attr.Value{
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
	}, types.ObjectNull(periodCadenceAttrTypes))
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

	attributes := value.Attributes()
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
