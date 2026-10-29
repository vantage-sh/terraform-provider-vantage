package vantage

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func attachmentLabelFiltersToAPI(ctx context.Context, legacy types.List, canonical types.Map, diags *diag.Diagnostics) map[string][]string {
	filters := labelFiltersToAPI(ctx, canonical, diags)
	if legacy.IsNull() || legacy.IsUnknown() {
		return filters
	}
	var labels []string
	diags.Append(legacy.ElementsAs(ctx, &labels, false)...)
	if len(labels) == 0 {
		return filters
	}
	if native, exists := filters["label"]; exists {
		if !sameLabelSelections(labels, native) {
			diags.AddError("Conflicting label filters", "label_filter and label_filters[\"label\"] must select the same values when both are configured.")
			return nil
		}
		return filters
	}
	filters["label"] = labels
	return filters
}

func attachmentLabelFiltersFromAPI(ctx context.Context, legacy types.List, canonical types.Map, configuredLegacy types.List, configuredCanonical types.Map, diags *diag.Diagnostics) (types.List, types.Map) {
	filters := labelFiltersToAPI(ctx, canonical, diags)
	if _, exists := filters["label"]; !exists && !legacy.IsNull() && !legacy.IsUnknown() && len(legacy.Elements()) > 0 {
		var labels []string
		diags.Append(legacy.ElementsAs(ctx, &labels, false)...)
		filters["label"] = labels
	}
	legacy = types.ListNull(types.StringType)
	if labels, exists := filters["label"]; exists {
		value, d := types.ListValueFrom(ctx, types.StringType, labels)
		diags.Append(d...)
		legacy = value
	}
	if !configuredLegacy.IsNull() && !configuredLegacy.IsUnknown() {
		var configuredLabels []string
		diags.Append(configuredLegacy.ElementsAs(ctx, &configuredLabels, false)...)
		if len(configuredLabels) == 0 || sameLabelSelections(configuredLabels, filters["label"]) {
			legacy = configuredLegacy
		}
	}
	if len(filters) == 0 && (configuredCanonical.IsNull() || configuredCanonical.IsUnknown()) {
		return legacy, emptyLabelFiltersMap()
	}
	value, d := types.MapValueFrom(ctx, types.ListType{ElemType: types.StringType}, filters)
	diags.Append(d...)
	return legacy, value
}

func attachmentLabelFilterInputs(ctx context.Context, planned types.List, configured types.List, prior types.List, tokenKey string, diags *diag.Diagnostics) types.List {
	if planned.IsNull() || planned.IsUnknown() {
		return planned
	}
	configuredByToken := attachmentFilterAttributesByToken(ctx, configured, tokenKey, diags)
	priorByToken := attachmentFilterAttributesByToken(ctx, prior, tokenKey, diags)
	values := make([]attr.Value, 0, len(planned.Elements()))
	var elementType types.ObjectType
	for _, value := range planned.Elements() {
		object, d := value.(basetypes.ObjectValuable).ToObjectValue(ctx)
		diags.Append(d...)
		attributes := object.Attributes()
		elementType = types.ObjectType{AttrTypes: object.AttributeTypes(ctx)}
		configuration, exists := configuredByToken[attributes[tokenKey].(types.String).ValueString()]
		if exists && configuration["label_filters"].IsNull() && configuration["label_filter"].IsNull() {
			previous := priorByToken[attributes[tokenKey].(types.String).ValueString()]
			attributes["label_filter"] = types.ListNull(types.StringType)
			attributes["label_filters"] = emptyLabelFiltersMap()
			if previous != nil {
				if !previous["label_filters"].IsNull() && !previous["label_filters"].IsUnknown() {
					attributes["label_filters"] = previous["label_filters"]
				} else {
					attributes["label_filter"] = previous["label_filter"]
				}
			}
		} else if exists {
			if configuration["label_filters"].IsNull() {
				attributes["label_filters"] = emptyLabelFiltersMap()
			}
			if configuration["label_filter"].IsNull() {
				attributes["label_filter"] = types.ListNull(types.StringType)
			}
		} else if !attributes["label_filters"].IsNull() && !attributes["label_filters"].IsUnknown() {
			attributes["label_filter"] = types.ListNull(types.StringType)
		}
		object, d = types.ObjectValue(elementType.AttrTypes, attributes)
		diags.Append(d...)
		values = append(values, object)
	}
	if len(values) == 0 {
		return planned
	}
	result, d := types.ListValue(elementType, values)
	diags.Append(d...)
	return result
}

func attachmentFilterAttributesByToken(ctx context.Context, values types.List, tokenKey string, diags *diag.Diagnostics) map[string]map[string]attr.Value {
	result := map[string]map[string]attr.Value{}
	for _, value := range values.Elements() {
		object, d := value.(basetypes.ObjectValuable).ToObjectValue(ctx)
		diags.Append(d...)
		attributes := object.Attributes()
		result[attributes[tokenKey].(types.String).ValueString()] = attributes
	}
	return result
}

func sameLabelSelections(left, right []string) bool {
	left, right = slices.Clone(left), slices.Clone(right)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(slices.Compact(left), slices.Compact(right))
}
