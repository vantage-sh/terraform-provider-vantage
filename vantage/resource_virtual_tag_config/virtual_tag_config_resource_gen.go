// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_virtual_tag_config

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func VirtualTagConfigResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"backfill_until": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The earliest month the VirtualTagConfig should be backfilled to.",
				MarkdownDescription: "The earliest month the VirtualTagConfig should be backfilled to.",
			},
			"created_by_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the User who created the VirtualTagConfig.",
				MarkdownDescription: "The token of the User who created the VirtualTagConfig.",
			},
			"key": schema.StringAttribute{
				Required:            true,
				Description:         "The key of the VirtualTagConfig.",
				MarkdownDescription: "The key of the VirtualTagConfig.",
			},
			"overridable": schema.BoolAttribute{
				Required:            true,
				Description:         "Whether the VirtualTagConfig can override a provider-supplied tag on a matching Cost.",
				MarkdownDescription: "Whether the VirtualTagConfig can override a provider-supplied tag on a matching Cost.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the VirtualTagConfig.",
				MarkdownDescription: "The token of the VirtualTagConfig.",
			},
			"values": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"business_metric_token": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The token of an associated business metric.",
							MarkdownDescription: "The token of an associated business metric.",
						},
						"cost_metric": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"aggregation": schema.SingleNestedAttribute{
									Attributes: map[string]schema.Attribute{
										"tag": schema.StringAttribute{
											Required:            true,
											Description:         "The tag to aggregate on.",
											MarkdownDescription: "The tag to aggregate on.",
										},
									},
									CustomType: AggregationType{
										ObjectType: types.ObjectType{
											AttrTypes: AggregationValue{}.AttributeTypes(ctx),
										},
									},
									Required: true,
								},
								"filter": schema.StringAttribute{
									Required:            true,
									Description:         "The filter VQL for the cost metric.",
									MarkdownDescription: "The filter VQL for the cost metric.",
								},
							},
							CustomType: CostMetricType{
								ObjectType: types.ObjectType{
									AttrTypes: CostMetricValue{}.AttributeTypes(ctx),
								},
							},
							Optional: true,
							Computed: true,
						},
						"filter": schema.StringAttribute{
							Required:            true,
							Description:         "The filter query language to apply to the value. Additional documentation available at https://docs.vantage.sh/vql.",
							MarkdownDescription: "The filter query language to apply to the value. Additional documentation available at https://docs.vantage.sh/vql.",
						},
						"name": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The name of the value.",
							MarkdownDescription: "The name of the value.",
						},
					},
					CustomType: ValuesType{
						ObjectType: types.ObjectType{
							AttrTypes: ValuesValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "Values for the VirtualTagConfig, with match precedence determined by order in the list.",
				MarkdownDescription: "Values for the VirtualTagConfig, with match precedence determined by order in the list.",
			},
		},
	}
}

type VirtualTagConfigModel struct {
	BackfillUntil  types.String `tfsdk:"backfill_until"`
	CreatedByToken types.String `tfsdk:"created_by_token"`
	Key            types.String `tfsdk:"key"`
	Overridable    types.Bool   `tfsdk:"overridable"`
	Token          types.String `tfsdk:"token"`
	Values         types.List   `tfsdk:"values"`
}

var _ basetypes.ObjectTypable = ValuesType{}

type ValuesType struct {
	basetypes.ObjectType
}

func (t ValuesType) Equal(o attr.Type) bool {
	other, ok := o.(ValuesType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ValuesType) String() string {
	return "ValuesType"
}

func (t ValuesType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	businessMetricTokenAttribute, ok := attributes["business_metric_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`business_metric_token is missing from object`)

		return nil, diags
	}

	businessMetricTokenVal, ok := businessMetricTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`business_metric_token expected to be basetypes.StringValue, was: %T`, businessMetricTokenAttribute))
	}

	costMetricAttribute, ok := attributes["cost_metric"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`cost_metric is missing from object`)

		return nil, diags
	}

	costMetricVal, ok := costMetricAttribute.(basetypes.ObjectValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`cost_metric expected to be basetypes.ObjectValue, was: %T`, costMetricAttribute))
	}

	filterAttribute, ok := attributes["filter"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`filter is missing from object`)

		return nil, diags
	}

	filterVal, ok := filterAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`filter expected to be basetypes.StringValue, was: %T`, filterAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return nil, diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return ValuesValue{
		BusinessMetricToken: businessMetricTokenVal,
		CostMetric:          costMetricVal,
		Filter:              filterVal,
		Name:                nameVal,
		state:               attr.ValueStateKnown,
	}, diags
}

func NewValuesValueNull() ValuesValue {
	return ValuesValue{
		state: attr.ValueStateNull,
	}
}

func NewValuesValueUnknown() ValuesValue {
	return ValuesValue{
		state: attr.ValueStateUnknown,
	}
}

func NewValuesValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (ValuesValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing ValuesValue Attribute Value",
				"While creating a ValuesValue value, a missing attribute value was detected. "+
					"A ValuesValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ValuesValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid ValuesValue Attribute Type",
				"While creating a ValuesValue value, an invalid attribute value was detected. "+
					"A ValuesValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ValuesValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("ValuesValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra ValuesValue Attribute Value",
				"While creating a ValuesValue value, an extra attribute value was detected. "+
					"A ValuesValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra ValuesValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewValuesValueUnknown(), diags
	}

	businessMetricTokenAttribute, ok := attributes["business_metric_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`business_metric_token is missing from object`)

		return NewValuesValueUnknown(), diags
	}

	businessMetricTokenVal, ok := businessMetricTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`business_metric_token expected to be basetypes.StringValue, was: %T`, businessMetricTokenAttribute))
	}

	costMetricAttribute, ok := attributes["cost_metric"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`cost_metric is missing from object`)

		return NewValuesValueUnknown(), diags
	}

	costMetricVal, ok := costMetricAttribute.(basetypes.ObjectValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`cost_metric expected to be basetypes.ObjectValue, was: %T`, costMetricAttribute))
	}

	filterAttribute, ok := attributes["filter"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`filter is missing from object`)

		return NewValuesValueUnknown(), diags
	}

	filterVal, ok := filterAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`filter expected to be basetypes.StringValue, was: %T`, filterAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewValuesValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	if diags.HasError() {
		return NewValuesValueUnknown(), diags
	}

	return ValuesValue{
		BusinessMetricToken: businessMetricTokenVal,
		CostMetric:          costMetricVal,
		Filter:              filterVal,
		Name:                nameVal,
		state:               attr.ValueStateKnown,
	}, diags
}

func NewValuesValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) ValuesValue {
	object, diags := NewValuesValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewValuesValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t ValuesType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewValuesValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewValuesValueUnknown(), nil
	}

	if in.IsNull() {
		return NewValuesValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewValuesValueMust(ValuesValue{}.AttributeTypes(ctx), attributes), nil
}

func (t ValuesType) ValueType(ctx context.Context) attr.Value {
	return ValuesValue{}
}

var _ basetypes.ObjectValuable = ValuesValue{}

type ValuesValue struct {
	BusinessMetricToken basetypes.StringValue `tfsdk:"business_metric_token"`
	CostMetric          basetypes.ObjectValue `tfsdk:"cost_metric"`
	Filter              basetypes.StringValue `tfsdk:"filter"`
	Name                basetypes.StringValue `tfsdk:"name"`
	state               attr.ValueState
}

func (v ValuesValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 4)

	var val tftypes.Value
	var err error

	attrTypes["business_metric_token"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["cost_metric"] = basetypes.ObjectType{
		AttrTypes: CostMetricValue{}.AttributeTypes(ctx),
	}.TerraformType(ctx)
	attrTypes["filter"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 4)

		val, err = v.BusinessMetricToken.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["business_metric_token"] = val

		val, err = v.CostMetric.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["cost_metric"] = val

		val, err = v.Filter.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["filter"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v ValuesValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ValuesValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ValuesValue) String() string {
	return "ValuesValue"
}

func (v ValuesValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	var costMetric basetypes.ObjectValue

	if v.CostMetric.IsNull() {
		costMetric = types.ObjectNull(
			CostMetricValue{}.AttributeTypes(ctx),
		)
	}

	if v.CostMetric.IsUnknown() {
		costMetric = types.ObjectUnknown(
			CostMetricValue{}.AttributeTypes(ctx),
		)
	}

	if !v.CostMetric.IsNull() && !v.CostMetric.IsUnknown() {
		costMetric = types.ObjectValueMust(
			CostMetricValue{}.AttributeTypes(ctx),
			v.CostMetric.Attributes(),
		)
	}

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"business_metric_token": basetypes.StringType{},
			"cost_metric": basetypes.ObjectType{
				AttrTypes: CostMetricValue{}.AttributeTypes(ctx),
			},
			"filter": basetypes.StringType{},
			"name":   basetypes.StringType{},
		},
		map[string]attr.Value{
			"business_metric_token": v.BusinessMetricToken,
			"cost_metric":           costMetric,
			"filter":                v.Filter,
			"name":                  v.Name,
		})

	return objVal, diags
}

func (v ValuesValue) Equal(o attr.Value) bool {
	other, ok := o.(ValuesValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.BusinessMetricToken.Equal(other.BusinessMetricToken) {
		return false
	}

	if !v.CostMetric.Equal(other.CostMetric) {
		return false
	}

	if !v.Filter.Equal(other.Filter) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	return true
}

func (v ValuesValue) Type(ctx context.Context) attr.Type {
	return ValuesType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v ValuesValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"business_metric_token": basetypes.StringType{},
		"cost_metric": basetypes.ObjectType{
			AttrTypes: CostMetricValue{}.AttributeTypes(ctx),
		},
		"filter": basetypes.StringType{},
		"name":   basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = CostMetricType{}

type CostMetricType struct {
	basetypes.ObjectType
}

func (t CostMetricType) Equal(o attr.Type) bool {
	other, ok := o.(CostMetricType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t CostMetricType) String() string {
	return "CostMetricType"
}

func (t CostMetricType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	aggregationAttribute, ok := attributes["aggregation"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`aggregation is missing from object`)

		return nil, diags
	}

	aggregationVal, ok := aggregationAttribute.(basetypes.ObjectValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`aggregation expected to be basetypes.ObjectValue, was: %T`, aggregationAttribute))
	}

	filterAttribute, ok := attributes["filter"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`filter is missing from object`)

		return nil, diags
	}

	filterVal, ok := filterAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`filter expected to be basetypes.StringValue, was: %T`, filterAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return CostMetricValue{
		Aggregation: aggregationVal,
		Filter:      filterVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewCostMetricValueNull() CostMetricValue {
	return CostMetricValue{
		state: attr.ValueStateNull,
	}
}

func NewCostMetricValueUnknown() CostMetricValue {
	return CostMetricValue{
		state: attr.ValueStateUnknown,
	}
}

func NewCostMetricValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (CostMetricValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing CostMetricValue Attribute Value",
				"While creating a CostMetricValue value, a missing attribute value was detected. "+
					"A CostMetricValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("CostMetricValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid CostMetricValue Attribute Type",
				"While creating a CostMetricValue value, an invalid attribute value was detected. "+
					"A CostMetricValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("CostMetricValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("CostMetricValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra CostMetricValue Attribute Value",
				"While creating a CostMetricValue value, an extra attribute value was detected. "+
					"A CostMetricValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra CostMetricValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewCostMetricValueUnknown(), diags
	}

	aggregationAttribute, ok := attributes["aggregation"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`aggregation is missing from object`)

		return NewCostMetricValueUnknown(), diags
	}

	aggregationVal, ok := aggregationAttribute.(basetypes.ObjectValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`aggregation expected to be basetypes.ObjectValue, was: %T`, aggregationAttribute))
	}

	filterAttribute, ok := attributes["filter"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`filter is missing from object`)

		return NewCostMetricValueUnknown(), diags
	}

	filterVal, ok := filterAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`filter expected to be basetypes.StringValue, was: %T`, filterAttribute))
	}

	if diags.HasError() {
		return NewCostMetricValueUnknown(), diags
	}

	return CostMetricValue{
		Aggregation: aggregationVal,
		Filter:      filterVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewCostMetricValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) CostMetricValue {
	object, diags := NewCostMetricValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewCostMetricValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t CostMetricType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewCostMetricValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewCostMetricValueUnknown(), nil
	}

	if in.IsNull() {
		return NewCostMetricValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewCostMetricValueMust(CostMetricValue{}.AttributeTypes(ctx), attributes), nil
}

func (t CostMetricType) ValueType(ctx context.Context) attr.Value {
	return CostMetricValue{}
}

var _ basetypes.ObjectValuable = CostMetricValue{}

type CostMetricValue struct {
	Aggregation basetypes.ObjectValue `tfsdk:"aggregation"`
	Filter      basetypes.StringValue `tfsdk:"filter"`
	state       attr.ValueState
}

func (v CostMetricValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["aggregation"] = basetypes.ObjectType{
		AttrTypes: AggregationValue{}.AttributeTypes(ctx),
	}.TerraformType(ctx)
	attrTypes["filter"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.Aggregation.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["aggregation"] = val

		val, err = v.Filter.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["filter"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v CostMetricValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v CostMetricValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v CostMetricValue) String() string {
	return "CostMetricValue"
}

func (v CostMetricValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	var aggregation basetypes.ObjectValue

	if v.Aggregation.IsNull() {
		aggregation = types.ObjectNull(
			AggregationValue{}.AttributeTypes(ctx),
		)
	}

	if v.Aggregation.IsUnknown() {
		aggregation = types.ObjectUnknown(
			AggregationValue{}.AttributeTypes(ctx),
		)
	}

	if !v.Aggregation.IsNull() && !v.Aggregation.IsUnknown() {
		aggregation = types.ObjectValueMust(
			AggregationValue{}.AttributeTypes(ctx),
			v.Aggregation.Attributes(),
		)
	}

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"aggregation": basetypes.ObjectType{
				AttrTypes: AggregationValue{}.AttributeTypes(ctx),
			},
			"filter": basetypes.StringType{},
		},
		map[string]attr.Value{
			"aggregation": aggregation,
			"filter":      v.Filter,
		})

	return objVal, diags
}

func (v CostMetricValue) Equal(o attr.Value) bool {
	other, ok := o.(CostMetricValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Aggregation.Equal(other.Aggregation) {
		return false
	}

	if !v.Filter.Equal(other.Filter) {
		return false
	}

	return true
}

func (v CostMetricValue) Type(ctx context.Context) attr.Type {
	return CostMetricType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v CostMetricValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"aggregation": basetypes.ObjectType{
			AttrTypes: AggregationValue{}.AttributeTypes(ctx),
		},
		"filter": basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = AggregationType{}

type AggregationType struct {
	basetypes.ObjectType
}

func (t AggregationType) Equal(o attr.Type) bool {
	other, ok := o.(AggregationType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t AggregationType) String() string {
	return "AggregationType"
}

func (t AggregationType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	tagAttribute, ok := attributes["tag"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`tag is missing from object`)

		return nil, diags
	}

	tagVal, ok := tagAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`tag expected to be basetypes.StringValue, was: %T`, tagAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return AggregationValue{
		Tag:   tagVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewAggregationValueNull() AggregationValue {
	return AggregationValue{
		state: attr.ValueStateNull,
	}
}

func NewAggregationValueUnknown() AggregationValue {
	return AggregationValue{
		state: attr.ValueStateUnknown,
	}
}

func NewAggregationValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (AggregationValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing AggregationValue Attribute Value",
				"While creating a AggregationValue value, a missing attribute value was detected. "+
					"A AggregationValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AggregationValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid AggregationValue Attribute Type",
				"While creating a AggregationValue value, an invalid attribute value was detected. "+
					"A AggregationValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AggregationValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("AggregationValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra AggregationValue Attribute Value",
				"While creating a AggregationValue value, an extra attribute value was detected. "+
					"A AggregationValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra AggregationValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewAggregationValueUnknown(), diags
	}

	tagAttribute, ok := attributes["tag"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`tag is missing from object`)

		return NewAggregationValueUnknown(), diags
	}

	tagVal, ok := tagAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`tag expected to be basetypes.StringValue, was: %T`, tagAttribute))
	}

	if diags.HasError() {
		return NewAggregationValueUnknown(), diags
	}

	return AggregationValue{
		Tag:   tagVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewAggregationValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) AggregationValue {
	object, diags := NewAggregationValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewAggregationValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t AggregationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewAggregationValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewAggregationValueUnknown(), nil
	}

	if in.IsNull() {
		return NewAggregationValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewAggregationValueMust(AggregationValue{}.AttributeTypes(ctx), attributes), nil
}

func (t AggregationType) ValueType(ctx context.Context) attr.Value {
	return AggregationValue{}
}

var _ basetypes.ObjectValuable = AggregationValue{}

type AggregationValue struct {
	Tag   basetypes.StringValue `tfsdk:"tag"`
	state attr.ValueState
}

func (v AggregationValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 1)

	var val tftypes.Value
	var err error

	attrTypes["tag"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 1)

		val, err = v.Tag.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["tag"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v AggregationValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v AggregationValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v AggregationValue) String() string {
	return "AggregationValue"
}

func (v AggregationValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"tag": basetypes.StringType{},
		},
		map[string]attr.Value{
			"tag": v.Tag,
		})

	return objVal, diags
}

func (v AggregationValue) Equal(o attr.Value) bool {
	other, ok := o.(AggregationValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Tag.Equal(other.Tag) {
		return false
	}

	return true
}

func (v AggregationValue) Type(ctx context.Context) attr.Type {
	return AggregationType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v AggregationValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"tag": basetypes.StringType{},
	}
}
