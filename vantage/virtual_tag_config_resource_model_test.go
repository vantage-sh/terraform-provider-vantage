package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

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
	for _, field := range []string{"date_ranges", "percentages", "label_transforms"} {
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
