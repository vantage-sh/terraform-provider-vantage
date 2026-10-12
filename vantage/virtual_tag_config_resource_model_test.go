package vantage

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
			LabelFilters:        types.MapNull(types.ListType{ElemType: types.StringType}),
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
		v.LabelFilters = types.MapUnknown(types.ListType{ElemType: types.StringType})
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

	t.Run("clear unknown list", func(t *testing.T) {
		stateValue := value("vtv_a", "a")
		stateValue.LabelTransforms = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("split")})
		changes := diffVirtualTagConfigValues(
			[]*virtualTagConfigValueModel{planned("a")},
			[]*virtualTagConfigValueModel{stateValue},
		)
		if !changes.requiresParentUpdate {
			t.Fatal("an omitted non-empty list should require the parent update endpoint")
		}
	})

	t.Run("clear active type", func(t *testing.T) {
		planValue := planned("a")
		planValue.Name = types.StringNull()
		changes := diffVirtualTagConfigValues([]*virtualTagConfigValueModel{planValue}, []*virtualTagConfigValueModel{a})
		if !changes.requiresParentUpdate {
			t.Fatal("clearing the active value type should require the parent update endpoint")
		}
	})

	t.Run("preserve unknown display name", func(t *testing.T) {
		percentages := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("allocation")})
		stateValue := value("vtv_a", "")
		stateValue.Name = types.StringNull()
		stateValue.Percentages = percentages
		stateValue.DisplayName = types.StringValue("Allocation")
		planValue := planned("")
		planValue.Name = types.StringUnknown()
		planValue.Percentages = percentages
		planValue.Filter = types.StringValue("costs.provider = 'gcp'")

		changes := diffVirtualTagConfigValues([]*virtualTagConfigValueModel{planValue}, []*virtualTagConfigValueModel{stateValue})
		if changes.requiresParentUpdate || len(changes.updates) != 1 {
			t.Fatalf("changes = %#v; want one granular update", changes)
		}
		if got := planValue.DisplayName.ValueString(); got != "Allocation" {
			t.Fatalf("display name = %q; want preserved state value", got)
		}
	})

	t.Run("do not preserve old type during type change", func(t *testing.T) {
		planValue := planned("")
		planValue.Name = types.StringUnknown()
		planValue.BusinessMetricToken = types.StringValue("bmet_test")
		changes := diffVirtualTagConfigValues([]*virtualTagConfigValueModel{planValue}, []*virtualTagConfigValueModel{a})
		if changes.requiresParentUpdate || len(changes.updates) != 1 {
			t.Fatalf("changes = %#v; want one granular type update", changes)
		}
		if !planValue.Name.IsUnknown() {
			t.Fatal("old name type should not be copied into a business metric update")
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
		Hidden:         true,
		Overridable:    true,
		Preferred:      true,
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
	if m.Hidden.IsNull() || m.Hidden.IsUnknown() || !m.Hidden.ValueBool() {
		t.Errorf("Hidden = %s; want true", m.Hidden)
	}
	if m.Preferred.IsNull() || m.Preferred.IsUnknown() || !m.Preferred.ValueBool() {
		t.Errorf("Preferred = %s; want true", m.Preferred)
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
	labelFilters, exists := attrs["label_filters"]
	if !exists {
		t.Fatal("values[0].label_filters is missing")
	}
	labelFiltersMap, ok := labelFilters.(basetypes.MapValue)
	if !ok {
		t.Fatalf("values[0].label_filters is %T; want basetypes.MapValue", labelFilters)
	}
	if !labelFiltersMap.IsNull() {
		t.Fatal("values[0].label_filters is set; want null when the API omits it")
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

func TestVirtualTagConfig_LabelFilters(t *testing.T) {
	ctx := context.Background()
	filters := map[string][]string{
		"edition": {"enterprise", "standard"},
		"region":  {"us"},
	}
	filterValue, d := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, filters)
	if d.HasError() {
		t.Fatalf("building label_filters: %v", d)
	}

	t.Run("read keeps configuration order", func(t *testing.T) {
		configured, d := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, map[string][]string{
			"team": {"platform", "data"},
		})
		if d.HasError() {
			t.Fatalf("building label_filters: %v", d)
		}
		m := &virtualTagConfigModel{
			resource_virtual_tag_config.VirtualTagConfigModel{
				Values: labelFilterValuesList(t, ctx, configured),
			},
		}
		name := "allocation"
		filter := "costs.provider = 'gcp'"
		labelKey := "team"
		payload := &modelsv2.VirtualTagConfig{
			Token:         "vtag_test",
			Key:           "test-key",
			Overridable:   true,
			BackfillUntil: "2025-01-01",
			Values: []*modelsv2.VirtualTagConfigValue{
				{
					Token:    "vtag_val_test0",
					Name:     &name,
					Filter:   &filter,
					LabelKey: &labelKey,
					LabelFilters: map[string][]string{
						"team": {"data", "platform", "finops"},
					},
				},
			},
		}
		if diags := m.applyPayload(ctx, payload); diags.HasError() {
			t.Fatalf("applyPayload returned errors: %v", diags)
		}
		got := labelFiltersFromState(t, ctx, m.Values.Elements()[0])
		want := map[string][]string{"team": {"platform", "data", "finops"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("label_filters = %#v; want %#v", got, want)
		}
	})

	t.Run("read", func(t *testing.T) {
		name := "allocation"
		filter := "costs.provider = 'gcp'"
		labelKey := "team"
		payload := &modelsv2.VirtualTagConfig{
			Token:         "vtag_test",
			Key:           "test-key",
			Overridable:   true,
			BackfillUntil: "2025-01-01",
			Values: []*modelsv2.VirtualTagConfigValue{
				{
					Token:        "vtag_val_test0",
					Name:         &name,
					Filter:       &filter,
					LabelKey:     &labelKey,
					LabelFilters: filters,
				},
			},
		}
		m := &virtualTagConfigModel{}
		if diags := m.applyPayload(ctx, payload); diags.HasError() {
			t.Fatalf("applyPayload returned errors: %v", diags)
		}
		got := labelFiltersFromState(t, ctx, m.Values.Elements()[0])
		if !reflect.DeepEqual(got, filters) {
			t.Fatalf("label_filters = %#v; want %#v", got, filters)
		}
	})

	t.Run("write", func(t *testing.T) {
		model := labelFilterValueModel(filterValue)
		var diags diag.Diagnostics
		created := model.toCreateValue(ctx, &diags)
		updated := model.toUpdateValue(ctx, &diags)
		if diags.HasError() {
			t.Fatalf("conversion errors: %v", diags)
		}
		if !reflect.DeepEqual(created.LabelFilters, filters) {
			t.Fatalf("create label_filters = %#v; want %#v", created.LabelFilters, filters)
		}
		if !reflect.DeepEqual(updated.LabelFilters, filters) {
			t.Fatalf("update label_filters = %#v; want %#v", updated.LabelFilters, filters)
		}

		parent := virtualTagConfigTestModel(ctx, "test-key", false)
		parent.Overridable = types.BoolValue(true)
		parent.BackfillUntil = types.StringValue("2025-01-01")
		parent.Values = labelFilterValuesList(t, ctx, filterValue)
		state := virtualTagConfigTestState(t, ctx, virtualTagConfigTestSchema(ctx), parent)
		var roundTrip virtualTagConfigModel
		if roundTripDiags := state.Get(ctx, &roundTrip); roundTripDiags.HasError() {
			t.Fatalf("reading round-trip state: %v", roundTripDiags)
		}
		createPayload := roundTrip.toCreate(ctx, &diags)
		updatePayload := roundTrip.toUpdate(ctx, &diags)
		if diags.HasError() {
			t.Fatalf("parent conversion errors: %v", diags)
		}
		if !reflect.DeepEqual(createPayload.Values[0].LabelFilters, filters) {
			t.Fatalf("parent create label_filters = %#v; want %#v", createPayload.Values[0].LabelFilters, filters)
		}
		if !reflect.DeepEqual(updatePayload.Values[0].LabelFilters, filters) {
			t.Fatalf("parent update label_filters = %#v; want %#v", updatePayload.Values[0].LabelFilters, filters)
		}
	})

	t.Run("omit sends null", func(t *testing.T) {
		model := labelFilterValueModel(types.MapNull(types.ListType{ElemType: types.StringType}))
		var diags diag.Diagnostics
		created := model.toCreateValue(ctx, &diags)
		updated := model.toUpdateValue(ctx, &diags)
		if diags.HasError() {
			t.Fatalf("conversion errors: %v", diags)
		}
		if created.LabelFilters != nil || updated.LabelFilters != nil {
			t.Fatalf("omitted label_filters = %#v / %#v; want nil", created.LabelFilters, updated.LabelFilters)
		}
	})

	t.Run("change uses value update", func(t *testing.T) {
		stateValue := labelFilterValueModel(filterValue)
		stateValue.Token = types.StringValue("vtv_a")
		replacement, d := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, map[string][]string{
			"edition": {"on-demand"},
		})
		if d.HasError() {
			t.Fatalf("building replacement label_filters: %v", d)
		}
		planValue := labelFilterValueModel(replacement)
		planValue.Token = types.StringValue("vtv_a")
		changes := diffVirtualTagConfigValues([]*virtualTagConfigValueModel{planValue}, []*virtualTagConfigValueModel{stateValue})
		if changes.requiresParentUpdate || len(changes.updates) != 1 {
			t.Fatalf("changes = %#v; want one value update", changes)
		}
	})

	t.Run("clear uses parent update", func(t *testing.T) {
		stateValue := labelFilterValueModel(filterValue)
		stateValue.Token = types.StringValue("vtv_a")
		planValue := labelFilterValueModel(types.MapNull(types.ListType{ElemType: types.StringType}))
		planValue.Token = types.StringValue("vtv_a")
		changes := diffVirtualTagConfigValues([]*virtualTagConfigValueModel{planValue}, []*virtualTagConfigValueModel{stateValue})
		if !changes.requiresParentUpdate {
			t.Fatal("clearing label_filters should require the parent update endpoint")
		}
	})
}

func labelFilterValueModel(labelFilters types.Map) *virtualTagConfigValueModel {
	return &virtualTagConfigValueModel{
		BusinessMetricToken: types.StringValue("bsnss_mtrc_test"),
		CostMetric:          resource_virtual_tag_config.NewCostMetricValueNull(),
		DateRanges:          types.ListNull(types.StringType),
		DisplayName:         types.StringNull(),
		Filter:              types.StringValue("costs.provider = 'gcp'"),
		LabelFilters:        labelFilters,
		LabelKey:            types.StringValue("team"),
		LabelTransforms:     types.ListNull(types.StringType),
		LabelValues:         types.ListNull(types.StringType),
		Name:                types.StringNull(),
		Percentages:         types.ListNull(types.StringType),
		Token:               types.StringNull(),
	}
}

func labelFilterValuesList(t *testing.T, ctx context.Context, labelFilters types.Map) types.List {
	t.Helper()
	model := labelFilterValueModel(labelFilters)
	attrs := map[string]attr.Value{
		"business_metric_token": model.BusinessMetricToken,
		"cost_metric":           types.ObjectNull(resource_virtual_tag_config.CostMetricValue{}.AttributeTypes(ctx)),
		"date_ranges":           types.ListValueMust(resource_virtual_tag_config.DateRangesValue{}.Type(ctx), []attr.Value{}),
		"display_name":          model.DisplayName,
		"filter":                model.Filter,
		"label_filters":         model.LabelFilters,
		"label_key":             model.LabelKey,
		"label_transforms":      types.ListValueMust(resource_virtual_tag_config.LabelTransformsValue{}.Type(ctx), []attr.Value{}),
		"label_values":          types.ListValueMust(types.StringType, []attr.Value{}),
		"name":                  model.Name,
		"percentages":           types.ListValueMust(resource_virtual_tag_config.PercentagesValue{}.Type(ctx), []attr.Value{}),
		"token":                 types.StringValue("vtag_val_test0"),
	}
	value := resource_virtual_tag_config.NewValuesValueMust(
		resource_virtual_tag_config.ValuesValue{}.AttributeTypes(ctx),
		attrs,
	)
	list, diags := types.ListValueFrom(ctx, resource_virtual_tag_config.ValuesValue{}.Type(ctx), []resource_virtual_tag_config.ValuesValue{value})
	if diags.HasError() {
		t.Fatalf("building values list: %v", diags)
	}
	return list
}

func labelFiltersFromState(t *testing.T, ctx context.Context, value attr.Value) map[string][]string {
	t.Helper()
	obj, ok := value.(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("value is %T; want basetypes.ObjectValue", value)
	}
	raw, ok := obj.Attributes()["label_filters"].(basetypes.MapValue)
	if !ok {
		t.Fatalf("label_filters is %T; want basetypes.MapValue", obj.Attributes()["label_filters"])
	}
	got := map[string][]string{}
	if d := raw.ElementsAs(ctx, &got, false); d.HasError() {
		t.Fatalf("reading label_filters: %v", d)
	}
	return got
}
