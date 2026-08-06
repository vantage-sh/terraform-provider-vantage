package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestVirtualTagConfigValueDiff(t *testing.T) {
	value := func(token, name string) *virtualTagConfigValueModel {
		return &virtualTagConfigValueModel{
			BusinessMetricToken: types.StringNull(),
			CostMetric:          resource_virtual_tag_config.NewCostMetricValueNull(),
			DateRanges:          types.ListNull(types.StringType),
			DisplayName:         types.StringNull(),
			Filter:              types.StringValue("costs.provider = 'aws'"),
			LabelKey:            types.StringNull(),
			LabelTransforms:     types.ListNull(types.StringType),
			LabelValues:         types.ListNull(types.StringType),
			Name:                types.StringValue(name),
			Percentages:         types.ListNull(types.StringType),
			Token:               types.StringValue(token),
		}
	}
	a := value("vtv_a", "a")
	b := value("vtv_b", "b")
	planned := func(name string) *virtualTagConfigValueModel {
		v := value("", name)
		v.BusinessMetricToken = types.StringUnknown()
		v.CostMetric = resource_virtual_tag_config.NewCostMetricValueUnknown()
		v.DateRanges = types.ListUnknown(types.StringType)
		v.DisplayName = types.StringUnknown()
		v.LabelKey = types.StringUnknown()
		v.LabelTransforms = types.ListUnknown(types.StringType)
		v.LabelValues = types.ListUnknown(types.StringType)
		v.Percentages = types.ListUnknown(types.StringType)
		v.Token = types.StringUnknown()
		return v
	}

	t.Run("append", func(t *testing.T) {
		changes := diffVirtualTagConfigValues(
			[]*virtualTagConfigValueModel{planned("a"), planned("b"), planned("c")},
			[]*virtualTagConfigValueModel{a, b},
		)
		if changes.requiresParentUpdate || len(changes.creates) != 1 {
			t.Fatalf("changes = %#v; want one granular create", changes)
		}
	})

	t.Run("update", func(t *testing.T) {
		changes := diffVirtualTagConfigValues(
			[]*virtualTagConfigValueModel{planned("a"), planned("updated")},
			[]*virtualTagConfigValueModel{a, b},
		)
		if changes.requiresParentUpdate || len(changes.updates) != 1 {
			t.Fatalf("changes = %#v; want one granular update", changes)
		}
	})

	t.Run("delete", func(t *testing.T) {
		changes := diffVirtualTagConfigValues([]*virtualTagConfigValueModel{planned("b")}, []*virtualTagConfigValueModel{a, b})
		if changes.requiresParentUpdate || len(changes.deletes) != 1 {
			t.Fatalf("changes = %#v; want one granular delete", changes)
		}
	})

	t.Run("reorder", func(t *testing.T) {
		changes := diffVirtualTagConfigValues(
			[]*virtualTagConfigValueModel{planned("b"), planned("a")},
			[]*virtualTagConfigValueModel{a, b},
		)
		if !changes.requiresParentUpdate {
			t.Fatal("reorder should require the parent update endpoint")
		}
	})

	t.Run("clear omitted patch field", func(t *testing.T) {
		stateValue := value("vtv_a", "a")
		stateValue.LabelKey = types.StringValue("team")
		planValue := planned("a")
		planValue.LabelKey = types.StringNull()
		changes := diffVirtualTagConfigValues([]*virtualTagConfigValueModel{planValue}, []*virtualTagConfigValueModel{stateValue})
		if !changes.requiresParentUpdate {
			t.Fatal("clearing label_key should require the parent update endpoint")
		}
	})
}

// Regression for ENG-2415: applyPayload must emit known-empty lists (not null)
// for Optional+Computed nested lists when the API response has nil/empty fields.
// Terraform treats null != [] for list attributes, so a planned known-empty
// list paired with a null read-back fails the post-apply consistency check.
func TestVirtualTagConfig_ApplyPayload_NilNestedListsAreKnownEmpty(t *testing.T) {
	ctx := context.Background()

	name := "value-0"
	filter := "costs.provider = 'aws'"
	createdBy := "usr_test"

	payload := &modelsv2.VirtualTagConfig{
		Token:          "vtag_test",
		Key:            "test-key",
		Overridable:    true,
		BackfillUntil:  "2025-01-01",
		CreatedByToken: &createdBy,
		CollapsedTagKeys: []*modelsv2.VirtualTagConfigCollapsedTagKey{
			{
				Key:       "environment",
				Providers: nil,
			},
		},
		Values: []*modelsv2.VirtualTagConfigValue{
			{
				Token:           "vtag_val_test0",
				Name:            &name,
				Filter:          &filter,
				DateRanges:      nil,
				Percentages:     nil,
				LabelTransforms: nil,
			},
		},
	}

	m := &virtualTagConfigModel{}
	if diags := m.applyPayload(ctx, payload); diags.HasError() {
		t.Fatalf("applyPayload returned errors: %v", diags)
	}

	if m.CollapsedTagKeys.IsNull() {
		t.Fatalf("CollapsedTagKeys is null; want known list")
	}
	ctkElements := m.CollapsedTagKeys.Elements()
	if len(ctkElements) != 1 {
		t.Fatalf("CollapsedTagKeys has %d elements; want 1", len(ctkElements))
	}
	ctk, ok := ctkElements[0].(resource_virtual_tag_config.CollapsedTagKeysValue)
	if !ok {
		t.Fatalf("CollapsedTagKeys[0] is %T; want CollapsedTagKeysValue", ctkElements[0])
	}
	if ctk.Providers.IsNull() {
		t.Errorf("collapsed_tag_keys[0].providers is null; want known-empty list")
	}
	if !ctk.Providers.IsNull() && len(ctk.Providers.Elements()) != 0 {
		t.Errorf("collapsed_tag_keys[0].providers has %d elements; want 0",
			len(ctk.Providers.Elements()))
	}

	if m.Values.IsNull() {
		t.Fatalf("Values is null; want known list")
	}
	valueElements := m.Values.Elements()
	if len(valueElements) != 1 {
		t.Fatalf("Values has %d elements; want 1", len(valueElements))
	}
	valueObj, ok := valueElements[0].(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("Values[0] is %T; want basetypes.ObjectValue", valueElements[0])
	}
	attrs := valueObj.Attributes()
	tokenAttr, ok := attrs["token"].(basetypes.StringValue)
	if !ok {
		t.Fatalf("values[0].token is %T; want basetypes.StringValue", attrs["token"])
	}
	if got := tokenAttr.ValueString(); got != "vtag_val_test0" {
		t.Errorf("values[0].token = %q; want %q", got, "vtag_val_test0")
	}
	for _, field := range []string{"date_ranges", "percentages", "label_transforms", "label_values"} {
		raw, exists := attrs[field]
		if !exists {
			t.Errorf("values[0].%s is missing", field)
			continue
		}
		listVal, ok := raw.(basetypes.ListValue)
		if !ok {
			t.Errorf("values[0].%s is %T; want basetypes.ListValue", field, raw)
			continue
		}
		if listVal.IsNull() {
			t.Errorf("values[0].%s is null; want known-empty list", field)
			continue
		}
		if len(listVal.Elements()) != 0 {
			t.Errorf("values[0].%s has %d elements; want 0", field, len(listVal.Elements()))
		}
	}
}

// Verifies populated nested lists round-trip through applyPayload unchanged.
func TestVirtualTagConfig_ApplyPayload_PopulatedNestedLists(t *testing.T) {
	ctx := context.Background()

	name := "value-0"
	filter := "costs.provider = 'aws'"
	createdBy := "usr_test"
	startDate := "2024-01-01"
	endDate := "2024-03-31"

	payload := &modelsv2.VirtualTagConfig{
		Token:          "vtag_test",
		Key:            "test-key",
		Overridable:    true,
		BackfillUntil:  "2025-01-01",
		CreatedByToken: &createdBy,
		CollapsedTagKeys: []*modelsv2.VirtualTagConfigCollapsedTagKey{
			{
				Key:       "project",
				Providers: []string{"aws", "gcp"},
			},
		},
		Values: []*modelsv2.VirtualTagConfigValue{
			{
				Token:  "vtag_val_test0",
				Name:   &name,
				Filter: &filter,
				DateRanges: []*modelsv2.VirtualTagConfigValueDateRange{
					{StartDate: &startDate, EndDate: &endDate},
				},
			},
		},
	}

	m := &virtualTagConfigModel{}
	if diags := m.applyPayload(ctx, payload); diags.HasError() {
		t.Fatalf("applyPayload returned errors: %v", diags)
	}

	ctk := m.CollapsedTagKeys.Elements()[0].(resource_virtual_tag_config.CollapsedTagKeysValue)
	if ctk.Providers.IsNull() {
		t.Fatalf("providers is null; want [\"aws\",\"gcp\"]")
	}
	if got := len(ctk.Providers.Elements()); got != 2 {
		t.Errorf("providers has %d elements; want 2", got)
	}

	valueObj := m.Values.Elements()[0].(basetypes.ObjectValue)
	dr := valueObj.Attributes()["date_ranges"].(basetypes.ListValue)
	if dr.IsNull() {
		t.Fatalf("date_ranges is null; want 1 element")
	}
	if got := len(dr.Elements()); got != 1 {
		t.Errorf("date_ranges has %d elements; want 1", got)
	}
}
