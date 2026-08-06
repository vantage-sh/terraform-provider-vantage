package vantage

import (
	"context"
	"slices"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type virtualTagConfigModel resource_virtual_tag_config.VirtualTagConfigModel

type virtualTagConfigValueModel struct {
	BusinessMetricToken types.String                                `tfsdk:"business_metric_token"`
	CostMetric          resource_virtual_tag_config.CostMetricValue `tfsdk:"cost_metric"`
	DateRanges          types.List                                  `tfsdk:"date_ranges"`
	DisplayName         types.String                                `tfsdk:"display_name"`
	Filter              types.String                                `tfsdk:"filter"`
	LabelKey            types.String                                `tfsdk:"label_key"`
	LabelTransforms     types.List                                  `tfsdk:"label_transforms"`
	LabelValues         types.List                                  `tfsdk:"label_values"`
	Name                types.String                                `tfsdk:"name"`
	Percentages         types.List                                  `tfsdk:"percentages"`
	Token               types.String                                `tfsdk:"token"`
}

// Intermediate types for shared conversion logic between Create and Update operations.
// The API generates separate types for each operation, but the data extraction from
// Terraform state is identical.

type collapsedTagKeyData struct {
	Filter    *string
	Key       *string
	Providers []string
}

type percentageData struct {
	Pct   float32
	Value *string
}

type dateRangeData struct {
	StartDate string
	EndDate   string
}

type labelTransformData struct {
	Type      string
	Delimiter *string
	Index     *int32
	Template  *string
}

type aggregationData struct {
	Tag *string
}

type costMetricData struct {
	Filter      *string
	Aggregation *aggregationData
}

type valueData struct {
	Name                string
	DisplayName         string
	Filter              *string
	BusinessMetricToken string
	LabelKey            string
	LabelValues         []string
	CostMetric          *costMetricData
	DateRanges          []dateRangeData
	LabelTransforms     []labelTransformData
	Percentages         []percentageData
}

type virtualTagConfigValueChanges struct {
	creates              []*virtualTagConfigValueModel
	updates              []*virtualTagConfigValueModel
	deletes              []*virtualTagConfigValueModel
	requiresParentUpdate bool
}

func (m *virtualTagConfigModel) parentFieldsEqual(other *virtualTagConfigModel) bool {
	return plannedValueEqual(m.Key, other.Key) &&
		plannedValueEqual(m.Overridable, other.Overridable) &&
		plannedValueEqual(m.BackfillUntil, other.BackfillUntil) &&
		plannedValueEqual(m.CollapsedTagKeys, other.CollapsedTagKeys)
}

func (v *virtualTagConfigValueModel) equal(other *virtualTagConfigValueModel) bool {
	return plannedValueEqual(v.BusinessMetricToken, other.BusinessMetricToken) &&
		plannedValueEqual(v.CostMetric, other.CostMetric) &&
		plannedValueEqual(v.DateRanges, other.DateRanges) &&
		plannedValueEqual(v.DisplayName, other.DisplayName) &&
		plannedValueEqual(v.Filter, other.Filter) &&
		plannedValueEqual(v.LabelKey, other.LabelKey) &&
		plannedValueEqual(v.LabelTransforms, other.LabelTransforms) &&
		plannedValueEqual(v.LabelValues, other.LabelValues) &&
		plannedValueEqual(v.Name, other.Name) &&
		plannedValueEqual(v.Percentages, other.Percentages)
}

func (v *virtualTagConfigValueModel) valueType() string {
	if !v.BusinessMetricToken.IsNull() && !v.BusinessMetricToken.IsUnknown() && v.BusinessMetricToken.ValueString() != "" {
		return "business_metric"
	}
	if !v.CostMetric.IsNull() && !v.CostMetric.IsUnknown() {
		return "cost_metric"
	}
	if !v.Percentages.IsNull() && !v.Percentages.IsUnknown() && len(v.Percentages.Elements()) > 0 {
		return "percentages"
	}
	if !v.Name.IsNull() && !v.Name.IsUnknown() && v.Name.ValueString() != "" {
		return "name"
	}
	return ""
}

func (v *virtualTagConfigValueModel) fillUnknownsFrom(state *virtualTagConfigValueModel) {
	if v.Filter.IsUnknown() {
		v.Filter = state.Filter
	}
	if valueType := v.valueType(); valueType != "" && valueType != state.valueType() {
		return
	}
	if v.DisplayName.IsUnknown() {
		v.DisplayName = state.DisplayName
	}
}

func plannedValueEqual(plan, state attr.Value) bool {
	return plan.IsUnknown() || plan.Equal(state)
}

func (v *virtualTagConfigValueModel) requiresParentUpdateFrom(state *virtualTagConfigValueModel) bool {
	return (state.valueType() != "" && v.valueType() == "") ||
		clearsList(v.DateRanges, state.DateRanges) ||
		clearsList(v.LabelTransforms, state.LabelTransforms) ||
		clearsList(v.LabelValues, state.LabelValues) ||
		(!state.LabelKey.IsNull() && state.LabelKey.ValueString() != "" &&
			(v.LabelKey.IsNull() || v.LabelKey.IsUnknown() || v.LabelKey.ValueString() == ""))
}

func clearsList(plan, state types.List) bool {
	return !state.IsNull() && !state.IsUnknown() && len(state.Elements()) > 0 &&
		(plan.IsNull() || plan.IsUnknown() || len(plan.Elements()) == 0)
}

func diffVirtualTagConfigValues(plan, state []*virtualTagConfigValueModel) virtualTagConfigValueChanges {
	changes := virtualTagConfigValueChanges{}
	if !resolveVirtualTagConfigValueTokens(plan, state) {
		changes.requiresParentUpdate = true
		return changes
	}
	stateByToken := make(map[string]*virtualTagConfigValueModel, len(state))
	for _, value := range state {
		token := value.Token.ValueString()
		if token == "" {
			changes.requiresParentUpdate = true
			return changes
		}
		stateByToken[token] = value
	}

	seen := make(map[string]bool, len(plan))
	planTokens := make([]string, 0, len(plan))
	for _, value := range plan {
		token := value.Token.ValueString()
		stateValue, exists := stateByToken[token]
		if token == "" || !exists {
			changes.creates = append(changes.creates, value)
			continue
		}
		value.fillUnknownsFrom(stateValue)
		if len(changes.creates) > 0 || seen[token] {
			changes.requiresParentUpdate = true
			return changes
		}
		seen[token] = true
		planTokens = append(planTokens, token)
		if value.requiresParentUpdateFrom(stateValue) {
			changes.requiresParentUpdate = true
			return changes
		}
		if !value.equal(stateValue) {
			changes.updates = append(changes.updates, value)
		}
	}

	stateTokens := make([]string, 0, len(state))
	for _, value := range state {
		token := value.Token.ValueString()
		if seen[token] {
			stateTokens = append(stateTokens, token)
		} else {
			changes.deletes = append(changes.deletes, value)
		}
	}
	changes.requiresParentUpdate = !slices.Equal(planTokens, stateTokens)
	return changes
}

func resolveVirtualTagConfigValueTokens(plan, state []*virtualTagConfigValueModel) bool {
	usedState := make([]bool, len(state))
	for _, planValue := range plan {
		token := planValue.Token.ValueString()
		if token == "" {
			continue
		}
		found := false
		for i, stateValue := range state {
			if token == stateValue.Token.ValueString() {
				usedState[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, planValue := range plan {
		if planValue.Token.ValueString() != "" {
			continue
		}
		match := -1
		for i, stateValue := range state {
			if !usedState[i] && planValue.equal(stateValue) {
				if match != -1 {
					match = -1
					break
				}
				match = i
			}
		}
		if match != -1 {
			planValue.Token = state[match].Token
			usedState[match] = true
		}
	}

	unmatchedPlan := make([]*virtualTagConfigValueModel, 0)
	for _, value := range plan {
		if value.Token.ValueString() == "" {
			unmatchedPlan = append(unmatchedPlan, value)
		}
	}
	unmatchedState := make([]*virtualTagConfigValueModel, 0)
	for i, value := range state {
		if !usedState[i] {
			unmatchedState = append(unmatchedState, value)
		}
	}
	if len(unmatchedPlan) > 0 && len(unmatchedState) > 0 {
		if len(unmatchedPlan) != len(unmatchedState) {
			return false
		}
		for i := range unmatchedPlan {
			unmatchedPlan[i].Token = unmatchedState[i].Token
		}
	}
	return true
}

func buildCostMetricFromPayload(ctx context.Context, cm *modelsv2.VirtualTagConfigValueCostMetric) (basetypes.ObjectValue, diag.Diagnostics) {
	tfAggregation := types.ObjectNull(resource_virtual_tag_config.AggregationValue{}.AttributeTypes(ctx))
	if cm.Aggregation != nil {
		aggregation, d := resource_virtual_tag_config.NewAggregationValue(
			resource_virtual_tag_config.AggregationValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"tag": types.StringPointerValue(cm.Aggregation.Tag),
			},
		)
		if d.HasError() {
			return basetypes.ObjectValue{}, d
		}
		tfAggregation, d = aggregation.ToObjectValue(ctx)
		if d.HasError() {
			return basetypes.ObjectValue{}, d
		}
	}

	costMetric, d := resource_virtual_tag_config.NewCostMetricValue(
		resource_virtual_tag_config.CostMetricValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"filter":      types.StringPointerValue(cm.Filter),
			"aggregation": tfAggregation,
		},
	)
	if d.HasError() {
		return basetypes.ObjectValue{}, d
	}
	return costMetric.ToObjectValue(ctx)
}

func buildPercentagesFromPayload(ctx context.Context, percentages []*modelsv2.VirtualTagConfigValuePercentage) (basetypes.ListValue, diag.Diagnostics) {
	tfPercentages := make([]resource_virtual_tag_config.PercentagesValue, 0, len(percentages))
	for _, p := range percentages {
		pv, d := resource_virtual_tag_config.NewPercentagesValue(
			resource_virtual_tag_config.PercentagesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"pct":   types.Float64Value(p.Pct),
				"value": types.StringValue(p.Value),
			},
		)
		if d.HasError() {
			return basetypes.ListValue{}, d
		}
		tfPercentages = append(tfPercentages, pv)
	}
	return types.ListValueFrom(ctx, resource_virtual_tag_config.PercentagesValue{}.Type(ctx), tfPercentages)
}

func buildLabelTransformsFromPayload(ctx context.Context, labelTransforms []*modelsv2.VirtualTagConfigValueLabelTransform) (basetypes.ListValue, diag.Diagnostics) {
	tfLabelTransforms := make([]resource_virtual_tag_config.LabelTransformsValue, 0, len(labelTransforms))
	for _, lt := range labelTransforms {
		var indexValue attr.Value = types.Int64Null()
		if lt.Index != nil {
			indexValue = types.Int64Value(int64(*lt.Index))
		}
		ltv, d := resource_virtual_tag_config.NewLabelTransformsValue(
			resource_virtual_tag_config.LabelTransformsValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"type":      types.StringValue(lt.Type),
				"delimiter": types.StringPointerValue(lt.Delimiter),
				"index":     indexValue,
				"template":  types.StringPointerValue(lt.Template),
			},
		)
		if d.HasError() {
			return basetypes.ListValue{}, d
		}
		tfLabelTransforms = append(tfLabelTransforms, ltv)
	}
	return types.ListValueFrom(ctx, resource_virtual_tag_config.LabelTransformsValue{}.Type(ctx), tfLabelTransforms)
}

func buildDateRangesFromPayload(ctx context.Context, dateRanges []*modelsv2.VirtualTagConfigValueDateRange) (basetypes.ListValue, diag.Diagnostics) {
	tfDateRanges := make([]resource_virtual_tag_config.DateRangesValue, 0, len(dateRanges))
	for _, dr := range dateRanges {
		drv, d := resource_virtual_tag_config.NewDateRangesValue(
			resource_virtual_tag_config.DateRangesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"start_date": types.StringPointerValue(dr.StartDate),
				"end_date":   types.StringPointerValue(dr.EndDate),
			},
		)
		if d.HasError() {
			return basetypes.ListValue{}, d
		}
		tfDateRanges = append(tfDateRanges, drv)
	}
	return types.ListValueFrom(ctx, resource_virtual_tag_config.DateRangesValue{}.Type(ctx), tfDateRanges)
}

func buildValueFromPayload(ctx context.Context, v *modelsv2.VirtualTagConfigValue) (basetypes.ObjectValue, diag.Diagnostics) {
	// For optional string fields, use null if empty, otherwise use the value
	var nameValue attr.Value
	if v.Name == nil || *v.Name == "" {
		nameValue = types.StringNull()
	} else {
		nameValue = types.StringValue(*v.Name)
	}

	var displayNameValue attr.Value
	if v.DisplayName == nil || *v.DisplayName == "" {
		displayNameValue = types.StringNull()
	} else {
		displayNameValue = types.StringValue(*v.DisplayName)
	}

	var businessMetricTokenValue attr.Value
	if v.BusinessMetricToken == nil || *v.BusinessMetricToken == "" {
		businessMetricTokenValue = types.StringNull()
	} else {
		businessMetricTokenValue = types.StringValue(*v.BusinessMetricToken)
	}

	// Build cost_metric value
	var costMetricValue attr.Value = types.ObjectNull(resource_virtual_tag_config.CostMetricValue{}.AttributeTypes(ctx))
	if v.CostMetric != nil {
		costMetric, d := buildCostMetricFromPayload(ctx, v.CostMetric)
		if d.HasError() {
			return basetypes.ObjectValue{}, d
		}
		costMetricValue = costMetric
	}

	// Always build percentages/date_ranges/label_transforms as known lists (empty
	// when the API returns no entries). Emitting types.ListNull here would
	// mismatch a planned known-empty list and trigger "Provider produced
	// inconsistent result after apply" — these attributes are Optional+Computed
	// lists, so Terraform treats null and [] as distinct values.
	percentagesValue, d := buildPercentagesFromPayload(ctx, v.Percentages)
	if d.HasError() {
		return basetypes.ObjectValue{}, d
	}

	dateRangesValue, d := buildDateRangesFromPayload(ctx, v.DateRanges)
	if d.HasError() {
		return basetypes.ObjectValue{}, d
	}

	labelTransformsValue, d := buildLabelTransformsFromPayload(ctx, v.LabelTransforms)
	if d.HasError() {
		return basetypes.ObjectValue{}, d
	}

	labelValues := v.LabelValues
	if labelValues == nil {
		labelValues = []string{}
	}
	labelValuesValue, d := types.ListValueFrom(ctx, types.StringType, labelValues)
	if d.HasError() {
		return basetypes.ObjectValue{}, d
	}

	// Use the constructor to properly set the state field
	value, diags := resource_virtual_tag_config.NewValuesValue(
		resource_virtual_tag_config.ValuesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"name":                  nameValue,
			"filter":                types.StringPointerValue(v.Filter),
			"business_metric_token": businessMetricTokenValue,
			"cost_metric":           costMetricValue,
			"date_ranges":           dateRangesValue,
			"display_name":          displayNameValue,
			"label_key":             types.StringPointerValue(v.LabelKey),
			"label_transforms":      labelTransformsValue,
			"label_values":          labelValuesValue,
			"percentages":           percentagesValue,
			"token":                 types.StringValue(v.Token),
		},
	)
	if diags.HasError() {
		return basetypes.ObjectValue{}, diags
	}

	return value.ToObjectValue(ctx)
}

func (m *virtualTagConfigModel) applyPayload(ctx context.Context, payload *modelsv2.VirtualTagConfig) diag.Diagnostics {
	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.Key = types.StringValue(payload.Key)
	m.Overridable = types.BoolValue(payload.Overridable)
	m.BackfillUntil = types.StringValue(payload.BackfillUntil)
	m.CreatedByToken = types.StringPointerValue(payload.CreatedByToken)

	tfCollapsedTagKeys := make([]resource_virtual_tag_config.CollapsedTagKeysValue, 0, len(payload.CollapsedTagKeys))
	for _, c := range payload.CollapsedTagKeys {
		// Coalesce a nil Providers slice to an empty slice so ListValueFrom emits
		// a known-empty list instead of null. Schema declares providers as
		// Optional+Computed; null vs [] would fail post-apply consistency checks.
		providers := c.Providers
		if providers == nil {
			providers = []string{}
		}
		tfProviders, diag := types.ListValueFrom(ctx, types.StringType, providers)
		if diag.HasError() {
			return diag
		}
		collapsedTagKey, diag := resource_virtual_tag_config.NewCollapsedTagKeysValue(
			resource_virtual_tag_config.CollapsedTagKeysValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"filter":    types.StringPointerValue(c.Filter),
				"key":       types.StringValue(c.Key),
				"providers": tfProviders,
			},
		)
		if diag.HasError() {
			return diag
		}
		tfCollapsedTagKeys = append(tfCollapsedTagKeys, collapsedTagKey)
	}
	tfCollapsedTagKeysValue, diag := types.ListValueFrom(ctx, resource_virtual_tag_config.CollapsedTagKeysValue{}.Type(ctx), tfCollapsedTagKeys)
	if diag.HasError() {
		return diag
	}
	m.CollapsedTagKeys = tfCollapsedTagKeysValue

	if payload.Values != nil {
		tfValues := make([]basetypes.ObjectValue, 0, len(payload.Values))
		for _, v := range payload.Values {
			tfValue, diag := buildValueFromPayload(ctx, v)
			if diag.HasError() {
				return diag
			}
			tfValues = append(tfValues, tfValue)
		}
		values, diag := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: resource_virtual_tag_config.ValuesValue{}.AttributeTypes(ctx)},
			tfValues,
		)
		if diag.HasError() {
			return diag
		}
		m.Values = values
	}

	return nil
}

func (m *virtualTagConfigModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateVirtualTagConfig {
	model := &modelsv2.CreateVirtualTagConfig{
		Key:         m.Key.ValueStringPointer(),
		Overridable: m.Overridable.ValueBoolPointer(),
	}

	backfillUntil := m.backfillUntilFromTf(diags)
	if diags.HasError() {
		return nil
	}
	model.BackfillUntil = *backfillUntil

	if collapsedTagKeys := m.collapsedTagKeysFromTf(ctx, diags); collapsedTagKeys != nil {
		if diags.HasError() {
			return nil
		}
		model.CollapsedTagKeys = make([]*modelsv2.CreateVirtualTagConfigCollapsedTagKeysItems0, 0, len(collapsedTagKeys))
		for _, c := range collapsedTagKeys {
			item := &modelsv2.CreateVirtualTagConfigCollapsedTagKeysItems0{
				Key:       c.Key,
				Providers: c.Providers,
			}
			if c.Filter != nil {
				item.Filter = *c.Filter
			}
			model.CollapsedTagKeys = append(model.CollapsedTagKeys, item)
		}
	}

	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		tfValues := m.valuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}
		model.Values = make([]*modelsv2.CreateVirtualTagConfigValuesItems0, 0, len(tfValues))
		for _, v := range tfValues {
			data := v.toValueData(ctx, diags)
			if diags.HasError() {
				return nil
			}

			value := &modelsv2.CreateVirtualTagConfigValuesItems0{
				Name:                data.Name,
				DisplayName:         data.DisplayName,
				Filter:              data.Filter,
				BusinessMetricToken: data.BusinessMetricToken,
				LabelKey:            data.LabelKey,
				LabelValues:         data.LabelValues,
			}

			if data.CostMetric != nil {
				value.CostMetric = &modelsv2.CreateVirtualTagConfigValuesItems0CostMetric{
					Filter: data.CostMetric.Filter,
				}
				if data.CostMetric.Aggregation != nil {
					value.CostMetric.Aggregation = &modelsv2.CreateVirtualTagConfigValuesItems0CostMetricAggregation{
						Tag: data.CostMetric.Aggregation.Tag,
					}
				}
			}

			if len(data.Percentages) > 0 {
				value.Percentages = make([]*modelsv2.CreateVirtualTagConfigValuesItems0PercentagesItems0, 0, len(data.Percentages))
				for _, p := range data.Percentages {
					pct := p.Pct
					value.Percentages = append(value.Percentages, &modelsv2.CreateVirtualTagConfigValuesItems0PercentagesItems0{
						Pct:   &pct,
						Value: p.Value,
					})
				}
			}

			if len(data.DateRanges) > 0 {
				value.DateRanges = make([]*modelsv2.CreateVirtualTagConfigValuesItems0DateRangesItems0, 0, len(data.DateRanges))
				for _, dr := range data.DateRanges {
					value.DateRanges = append(value.DateRanges, &modelsv2.CreateVirtualTagConfigValuesItems0DateRangesItems0{
						StartDate: dr.StartDate,
						EndDate:   dr.EndDate,
					})
				}
			}

			if len(data.LabelTransforms) > 0 {
				value.LabelTransforms = make([]*modelsv2.CreateVirtualTagConfigValuesItems0LabelTransformsItems0, 0, len(data.LabelTransforms))
				for _, lt := range data.LabelTransforms {
					ltType := lt.Type
					value.LabelTransforms = append(value.LabelTransforms, &modelsv2.CreateVirtualTagConfigValuesItems0LabelTransformsItems0{
						Type:      &ltType,
						Delimiter: lt.Delimiter,
						Index:     lt.Index,
						Template:  lt.Template,
					})
				}
			}

			model.Values = append(model.Values, value)
		}
	}

	return model
}

func (m *virtualTagConfigModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateVirtualTagConfig {
	if m.Token.IsNull() || m.Token.IsUnknown() {
		diags.AddError("virtual_tag_config_token is required", "")
		return nil
	}

	model := &modelsv2.UpdateVirtualTagConfig{}

	if !m.Key.IsNull() {
		model.Key = m.Key.ValueString()
	}

	if !m.Overridable.IsNull() {
		model.Overridable = m.Overridable.ValueBoolPointer()
	}

	if !m.BackfillUntil.IsNull() {
		model.BackfillUntil = m.backfillUntilFromTf(diags)
		if diags.HasError() {
			return nil
		}
	}

	if collapsedTagKeys := m.collapsedTagKeysFromTf(ctx, diags); collapsedTagKeys != nil {
		if diags.HasError() {
			return nil
		}
		model.CollapsedTagKeys = make([]*modelsv2.UpdateVirtualTagConfigCollapsedTagKeysItems0, 0, len(collapsedTagKeys))
		for _, c := range collapsedTagKeys {
			item := &modelsv2.UpdateVirtualTagConfigCollapsedTagKeysItems0{
				Key:       c.Key,
				Providers: c.Providers,
			}
			if c.Filter != nil {
				item.Filter = *c.Filter
			}
			model.CollapsedTagKeys = append(model.CollapsedTagKeys, item)
		}
	}

	if !m.Values.IsNull() && !m.Values.IsUnknown() {
		tfValues := m.valuesFromTf(ctx, diags)
		if diags.HasError() {
			return nil
		}

		model.Values = make([]*modelsv2.UpdateVirtualTagConfigValuesItems0, 0, len(tfValues))
		for _, v := range tfValues {
			data := v.toValueData(ctx, diags)
			if diags.HasError() {
				return nil
			}

			value := &modelsv2.UpdateVirtualTagConfigValuesItems0{
				Name:                data.Name,
				DisplayName:         data.DisplayName,
				Filter:              data.Filter,
				BusinessMetricToken: data.BusinessMetricToken,
				LabelKey:            data.LabelKey,
				LabelValues:         data.LabelValues,
			}

			if data.CostMetric != nil {
				value.CostMetric = &modelsv2.UpdateVirtualTagConfigValuesItems0CostMetric{
					Filter: data.CostMetric.Filter,
				}
				if data.CostMetric.Aggregation != nil {
					value.CostMetric.Aggregation = &modelsv2.UpdateVirtualTagConfigValuesItems0CostMetricAggregation{
						Tag: data.CostMetric.Aggregation.Tag,
					}
				}
			}

			if len(data.Percentages) > 0 {
				value.Percentages = make([]*modelsv2.UpdateVirtualTagConfigValuesItems0PercentagesItems0, 0, len(data.Percentages))
				for _, p := range data.Percentages {
					pct := p.Pct
					value.Percentages = append(value.Percentages, &modelsv2.UpdateVirtualTagConfigValuesItems0PercentagesItems0{
						Pct:   &pct,
						Value: p.Value,
					})
				}
			}

			if len(data.DateRanges) > 0 {
				value.DateRanges = make([]*modelsv2.UpdateVirtualTagConfigValuesItems0DateRangesItems0, 0, len(data.DateRanges))
				for _, dr := range data.DateRanges {
					value.DateRanges = append(value.DateRanges, &modelsv2.UpdateVirtualTagConfigValuesItems0DateRangesItems0{
						StartDate: dr.StartDate,
						EndDate:   dr.EndDate,
					})
				}
			}

			if len(data.LabelTransforms) > 0 {
				value.LabelTransforms = make([]*modelsv2.UpdateVirtualTagConfigValuesItems0LabelTransformsItems0, 0, len(data.LabelTransforms))
				for _, lt := range data.LabelTransforms {
					ltType := lt.Type
					value.LabelTransforms = append(value.LabelTransforms, &modelsv2.UpdateVirtualTagConfigValuesItems0LabelTransformsItems0{
						Type:      &ltType,
						Delimiter: lt.Delimiter,
						Index:     lt.Index,
						Template:  lt.Template,
					})
				}
			}

			model.Values = append(model.Values, value)
		}
	}

	return model
}

func (m *virtualTagConfigModel) backfillUntilFromTf(diags *diag.Diagnostics) *strfmt.Date {
	date := strfmt.Date{}
	if err := date.UnmarshalText([]byte(m.BackfillUntil.ValueString())); err != nil {
		diags.AddError("Unable to parse backfill_until", err.Error())
	}
	return &date
}

func (m *virtualTagConfigModel) valuesFromTf(ctx context.Context, diags *diag.Diagnostics) []*virtualTagConfigValueModel {
	values := make([]*virtualTagConfigValueModel, 0, len(m.Values.Elements()))
	if diag := m.Values.ElementsAs(ctx, &values, false); diag.HasError() {
		diags.Append(diag...)
		return nil
	}
	return values
}

// collapsedTagKeysFromTf extracts collapsed tag keys from Terraform state into an intermediate format.
func (m *virtualTagConfigModel) collapsedTagKeysFromTf(ctx context.Context, diags *diag.Diagnostics) []collapsedTagKeyData {
	if m.CollapsedTagKeys.IsNull() || m.CollapsedTagKeys.IsUnknown() {
		return nil
	}

	tfCollapsedTagKeys := make([]resource_virtual_tag_config.CollapsedTagKeysValue, 0, len(m.CollapsedTagKeys.Elements()))
	if d := m.CollapsedTagKeys.ElementsAs(ctx, &tfCollapsedTagKeys, false); d.HasError() {
		diags.Append(d...)
		return nil
	}

	result := make([]collapsedTagKeyData, 0, len(tfCollapsedTagKeys))
	for _, c := range tfCollapsedTagKeys {
		var providers []string
		// Only extract providers if it's not null or unknown
		if !c.Providers.IsNull() && !c.Providers.IsUnknown() {
			providers = make([]string, 0, len(c.Providers.Elements()))
			if d := c.Providers.ElementsAs(ctx, &providers, false); d.HasError() {
				diags.Append(d...)
				return nil
			}
		}
		result = append(result, collapsedTagKeyData{
			Filter:    c.Filter.ValueStringPointer(),
			Key:       c.Key.ValueStringPointer(),
			Providers: providers,
		})
	}
	return result
}

// toValueData extracts a single value's data from Terraform state into an intermediate format.
func (v *virtualTagConfigValueModel) toValueData(ctx context.Context, diags *diag.Diagnostics) *valueData {
	data := &valueData{
		Name:                v.Name.ValueString(),
		DisplayName:         v.DisplayName.ValueString(),
		Filter:              v.Filter.ValueStringPointer(),
		BusinessMetricToken: v.BusinessMetricToken.ValueString(),
		LabelKey:            v.LabelKey.ValueString(),
	}

	if !v.LabelValues.IsNull() && !v.LabelValues.IsUnknown() {
		labelValues := make([]string, 0, len(v.LabelValues.Elements()))
		if d := v.LabelValues.ElementsAs(ctx, &labelValues, false); d.HasError() {
			diags.Append(d...)
			return nil
		}
		data.LabelValues = labelValues
	}

	if !v.CostMetric.IsNull() && !v.CostMetric.IsUnknown() {
		data.CostMetric = &costMetricData{
			Filter: v.CostMetric.Filter.ValueStringPointer(),
		}

		if !v.CostMetric.Aggregation.IsNull() && !v.CostMetric.Aggregation.IsUnknown() {
			// Access the tag attribute directly from the ObjectValue
			if tagAttr, ok := v.CostMetric.Aggregation.Attributes()["tag"].(basetypes.StringValue); ok {
				data.CostMetric.Aggregation = &aggregationData{
					Tag: tagAttr.ValueStringPointer(),
				}
			}
		}
	}

	if !v.Percentages.IsNull() && !v.Percentages.IsUnknown() {
		tfPercentages := make([]resource_virtual_tag_config.PercentagesValue, 0, len(v.Percentages.Elements()))
		if d := v.Percentages.ElementsAs(ctx, &tfPercentages, false); d.HasError() {
			diags.Append(d...)
			return nil
		}
		data.Percentages = make([]percentageData, 0, len(tfPercentages))
		for _, p := range tfPercentages {
			data.Percentages = append(data.Percentages, percentageData{
				Pct:   float32(p.Pct.ValueFloat64()),
				Value: p.Value.ValueStringPointer(),
			})
		}
	}

	if !v.DateRanges.IsNull() && !v.DateRanges.IsUnknown() {
		tfDateRanges := make([]resource_virtual_tag_config.DateRangesValue, 0, len(v.DateRanges.Elements()))
		if d := v.DateRanges.ElementsAs(ctx, &tfDateRanges, false); d.HasError() {
			diags.Append(d...)
			return nil
		}
		data.DateRanges = make([]dateRangeData, 0, len(tfDateRanges))
		for _, dr := range tfDateRanges {
			data.DateRanges = append(data.DateRanges, dateRangeData{
				StartDate: dr.StartDate.ValueString(),
				EndDate:   dr.EndDate.ValueString(),
			})
		}
	}

	if !v.LabelTransforms.IsNull() && !v.LabelTransforms.IsUnknown() {
		tfLabelTransforms := make([]resource_virtual_tag_config.LabelTransformsValue, 0, len(v.LabelTransforms.Elements()))
		if d := v.LabelTransforms.ElementsAs(ctx, &tfLabelTransforms, false); d.HasError() {
			diags.Append(d...)
			return nil
		}
		data.LabelTransforms = make([]labelTransformData, 0, len(tfLabelTransforms))
		for _, lt := range tfLabelTransforms {
			lt := lt
			item := labelTransformData{
				Type:      lt.LabelTransformsType.ValueString(),
				Delimiter: lt.Delimiter.ValueStringPointer(),
				Template:  lt.Template.ValueStringPointer(),
			}
			if !lt.Index.IsNull() && !lt.Index.IsUnknown() {
				idx := int32(lt.Index.ValueInt64())
				item.Index = &idx
			}
			data.LabelTransforms = append(data.LabelTransforms, item)
		}
	}

	return data
}

func (v *virtualTagConfigValueModel) toCreateValue(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateVirtualTagConfigValue {
	data := v.toValueData(ctx, diags)
	if diags.HasError() {
		return nil
	}

	value := &modelsv2.CreateVirtualTagConfigValue{
		BusinessMetricToken: data.BusinessMetricToken,
		Filter:              data.Filter,
		LabelKey:            data.LabelKey,
		LabelValues:         data.LabelValues,
		Name:                data.Name,
	}
	if data.CostMetric != nil || len(data.Percentages) > 0 {
		value.DisplayName = &data.DisplayName
	}
	if data.CostMetric != nil {
		value.CostMetric = &modelsv2.CreateVirtualTagConfigValueCostMetric{Filter: data.CostMetric.Filter}
		if data.CostMetric.Aggregation != nil {
			value.CostMetric.Aggregation = &modelsv2.CreateVirtualTagConfigValueCostMetricAggregation{
				Tag: data.CostMetric.Aggregation.Tag,
			}
		}
	}
	for _, percentage := range data.Percentages {
		pct := percentage.Pct
		value.Percentages = append(value.Percentages, &modelsv2.CreateVirtualTagConfigValuePercentagesItems0{
			Pct:   &pct,
			Value: percentage.Value,
		})
	}
	for _, dateRange := range data.DateRanges {
		item := &modelsv2.CreateVirtualTagConfigValueDateRangesItems0{}
		if dateRange.StartDate != "" {
			item.StartDate = &dateRange.StartDate
		}
		if dateRange.EndDate != "" {
			item.EndDate = &dateRange.EndDate
		}
		value.DateRanges = append(value.DateRanges, item)
	}
	for _, transform := range data.LabelTransforms {
		transformType := transform.Type
		value.LabelTransforms = append(value.LabelTransforms, &modelsv2.CreateVirtualTagConfigValueLabelTransformsItems0{
			Type:      &transformType,
			Delimiter: transform.Delimiter,
			Index:     transform.Index,
			Template:  transform.Template,
		})
	}
	return value
}

func (v *virtualTagConfigValueModel) toUpdateValue(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateVirtualTagConfigValue {
	data := v.toValueData(ctx, diags)
	if diags.HasError() {
		return nil
	}

	value := &modelsv2.UpdateVirtualTagConfigValue{
		BusinessMetricToken: data.BusinessMetricToken,
		LabelKey:            data.LabelKey,
		LabelValues:         data.LabelValues,
		Name:                data.Name,
	}
	if data.Filter != nil {
		value.Filter = *data.Filter
	}
	if data.CostMetric != nil || len(data.Percentages) > 0 {
		value.DisplayName = &data.DisplayName
	}
	if data.CostMetric != nil {
		value.CostMetric = &modelsv2.UpdateVirtualTagConfigValueCostMetric{Filter: data.CostMetric.Filter}
		if data.CostMetric.Aggregation != nil {
			value.CostMetric.Aggregation = &modelsv2.UpdateVirtualTagConfigValueCostMetricAggregation{
				Tag: data.CostMetric.Aggregation.Tag,
			}
		}
	}
	for _, percentage := range data.Percentages {
		pct := percentage.Pct
		value.Percentages = append(value.Percentages, &modelsv2.UpdateVirtualTagConfigValuePercentagesItems0{
			Pct:   &pct,
			Value: percentage.Value,
		})
	}
	for _, dateRange := range data.DateRanges {
		item := &modelsv2.UpdateVirtualTagConfigValueDateRangesItems0{}
		if dateRange.StartDate != "" {
			item.StartDate = &dateRange.StartDate
		}
		if dateRange.EndDate != "" {
			item.EndDate = &dateRange.EndDate
		}
		value.DateRanges = append(value.DateRanges, item)
	}
	for _, transform := range data.LabelTransforms {
		transformType := transform.Type
		value.LabelTransforms = append(value.LabelTransforms, &modelsv2.UpdateVirtualTagConfigValueLabelTransformsItems0{
			Type:      &transformType,
			Delimiter: transform.Delimiter,
			Index:     transform.Index,
			Template:  transform.Template,
		})
	}
	return value
}
