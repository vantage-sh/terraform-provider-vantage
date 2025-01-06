// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_budget

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

func BudgetResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"budget_alert_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "The tokens of the BudgetAlerts associated with the Budget.",
				MarkdownDescription: "The tokens of the BudgetAlerts associated with the Budget.",
			},
			"child_budget_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The tokens of any child Budgets when creating a hierarchical Budget.",
				MarkdownDescription: "The tokens of any child Budgets when creating a hierarchical Budget.",
			},
			"cost_report_token": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The CostReport token. Ignored for hierarchical Budgets.",
				MarkdownDescription: "The CostReport token. Ignored for hierarchical Budgets.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
			},
			"created_by_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the Creator of the Budget.",
				MarkdownDescription: "The token of the Creator of the Budget.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the Budget.",
				MarkdownDescription: "The name of the Budget.",
			},
			"performance": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"actual": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
							MarkdownDescription: "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
						},
						"amount": schema.StringAttribute{
							Computed:            true,
							Description:         "The amount of the Budget Period as a string to ensure precision.",
							MarkdownDescription: "The amount of the Budget Period as a string to ensure precision.",
						},
						"date": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
							MarkdownDescription: "The date and time, in UTC, the Budget was created. ISO 8601 Formatted.",
						},
					},
					CustomType: PerformanceType{
						ObjectType: types.ObjectType{
							AttrTypes: PerformanceValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed:            true,
				Description:         "The historical performance of the Budget.",
				MarkdownDescription: "The historical performance of the Budget.",
			},
			"periods": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"amount": schema.Float64Attribute{
							Required:            true,
							Description:         "The amount of the period.",
							MarkdownDescription: "The amount of the period.",
						},
						"end_at": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The end date of the period.",
							MarkdownDescription: "The end date of the period.",
						},
						"start_at": schema.StringAttribute{
							Required:            true,
							Description:         "The start date of the period.",
							MarkdownDescription: "The start date of the period.",
						},
					},
					CustomType: PeriodsType{
						ObjectType: types.ObjectType{
							AttrTypes: PeriodsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "The periods for the Budget. The start_at and end_at must be iso8601 formatted e.g. YYYY-MM-DD. Ignored for hierarchical Budgets.",
				MarkdownDescription: "The periods for the Budget. The start_at and end_at must be iso8601 formatted e.g. YYYY-MM-DD. Ignored for hierarchical Budgets.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the budget",
				MarkdownDescription: "The token of the budget",
			},
			"user_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the User who created this Budget.",
				MarkdownDescription: "The token for the User who created this Budget.",
			},
			"workspace_token": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The token of the Workspace to add the Budget to.",
				MarkdownDescription: "The token of the Workspace to add the Budget to.",
			},
		},
	}
}

type BudgetModel struct {
	BudgetAlertTokens types.List   `tfsdk:"budget_alert_tokens"`
	ChildBudgetTokens types.List   `tfsdk:"child_budget_tokens"`
	CostReportToken   types.String `tfsdk:"cost_report_token"`
	CreatedAt         types.String `tfsdk:"created_at"`
	CreatedByToken    types.String `tfsdk:"created_by_token"`
	Name              types.String `tfsdk:"name"`
	Performance       types.List   `tfsdk:"performance"`
	Periods           types.List   `tfsdk:"periods"`
	Token             types.String `tfsdk:"token"`
	UserToken         types.String `tfsdk:"user_token"`
	WorkspaceToken    types.String `tfsdk:"workspace_token"`
}

var _ basetypes.ObjectTypable = PerformanceType{}

type PerformanceType struct {
	basetypes.ObjectType
}

func (t PerformanceType) Equal(o attr.Type) bool {
	other, ok := o.(PerformanceType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t PerformanceType) String() string {
	return "PerformanceType"
}

func (t PerformanceType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	actualAttribute, ok := attributes["actual"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`actual is missing from object`)

		return nil, diags
	}

	actualVal, ok := actualAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`actual expected to be basetypes.StringValue, was: %T`, actualAttribute))
	}

	amountAttribute, ok := attributes["amount"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`amount is missing from object`)

		return nil, diags
	}

	amountVal, ok := amountAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`amount expected to be basetypes.StringValue, was: %T`, amountAttribute))
	}

	dateAttribute, ok := attributes["date"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`date is missing from object`)

		return nil, diags
	}

	dateVal, ok := dateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`date expected to be basetypes.StringValue, was: %T`, dateAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return PerformanceValue{
		Actual: actualVal,
		Amount: amountVal,
		Date:   dateVal,
		state:  attr.ValueStateKnown,
	}, diags
}

func NewPerformanceValueNull() PerformanceValue {
	return PerformanceValue{
		state: attr.ValueStateNull,
	}
}

func NewPerformanceValueUnknown() PerformanceValue {
	return PerformanceValue{
		state: attr.ValueStateUnknown,
	}
}

func NewPerformanceValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (PerformanceValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing PerformanceValue Attribute Value",
				"While creating a PerformanceValue value, a missing attribute value was detected. "+
					"A PerformanceValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("PerformanceValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid PerformanceValue Attribute Type",
				"While creating a PerformanceValue value, an invalid attribute value was detected. "+
					"A PerformanceValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("PerformanceValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("PerformanceValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra PerformanceValue Attribute Value",
				"While creating a PerformanceValue value, an extra attribute value was detected. "+
					"A PerformanceValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra PerformanceValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewPerformanceValueUnknown(), diags
	}

	actualAttribute, ok := attributes["actual"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`actual is missing from object`)

		return NewPerformanceValueUnknown(), diags
	}

	actualVal, ok := actualAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`actual expected to be basetypes.StringValue, was: %T`, actualAttribute))
	}

	amountAttribute, ok := attributes["amount"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`amount is missing from object`)

		return NewPerformanceValueUnknown(), diags
	}

	amountVal, ok := amountAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`amount expected to be basetypes.StringValue, was: %T`, amountAttribute))
	}

	dateAttribute, ok := attributes["date"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`date is missing from object`)

		return NewPerformanceValueUnknown(), diags
	}

	dateVal, ok := dateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`date expected to be basetypes.StringValue, was: %T`, dateAttribute))
	}

	if diags.HasError() {
		return NewPerformanceValueUnknown(), diags
	}

	return PerformanceValue{
		Actual: actualVal,
		Amount: amountVal,
		Date:   dateVal,
		state:  attr.ValueStateKnown,
	}, diags
}

func NewPerformanceValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) PerformanceValue {
	object, diags := NewPerformanceValue(attributeTypes, attributes)

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

		panic("NewPerformanceValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t PerformanceType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewPerformanceValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewPerformanceValueUnknown(), nil
	}

	if in.IsNull() {
		return NewPerformanceValueNull(), nil
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

	return NewPerformanceValueMust(PerformanceValue{}.AttributeTypes(ctx), attributes), nil
}

func (t PerformanceType) ValueType(ctx context.Context) attr.Value {
	return PerformanceValue{}
}

var _ basetypes.ObjectValuable = PerformanceValue{}

type PerformanceValue struct {
	Actual basetypes.StringValue `tfsdk:"actual"`
	Amount basetypes.StringValue `tfsdk:"amount"`
	Date   basetypes.StringValue `tfsdk:"date"`
	state  attr.ValueState
}

func (v PerformanceValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)

	var val tftypes.Value
	var err error

	attrTypes["actual"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["amount"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["date"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err = v.Actual.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["actual"] = val

		val, err = v.Amount.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["amount"] = val

		val, err = v.Date.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["date"] = val

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

func (v PerformanceValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v PerformanceValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v PerformanceValue) String() string {
	return "PerformanceValue"
}

func (v PerformanceValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"actual": basetypes.StringType{},
			"amount": basetypes.StringType{},
			"date":   basetypes.StringType{},
		},
		map[string]attr.Value{
			"actual": v.Actual,
			"amount": v.Amount,
			"date":   v.Date,
		})

	return objVal, diags
}

func (v PerformanceValue) Equal(o attr.Value) bool {
	other, ok := o.(PerformanceValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Actual.Equal(other.Actual) {
		return false
	}

	if !v.Amount.Equal(other.Amount) {
		return false
	}

	if !v.Date.Equal(other.Date) {
		return false
	}

	return true
}

func (v PerformanceValue) Type(ctx context.Context) attr.Type {
	return PerformanceType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v PerformanceValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"actual": basetypes.StringType{},
		"amount": basetypes.StringType{},
		"date":   basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = PeriodsType{}

type PeriodsType struct {
	basetypes.ObjectType
}

func (t PeriodsType) Equal(o attr.Type) bool {
	other, ok := o.(PeriodsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t PeriodsType) String() string {
	return "PeriodsType"
}

func (t PeriodsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	amountAttribute, ok := attributes["amount"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`amount is missing from object`)

		return nil, diags
	}

	amountVal, ok := amountAttribute.(basetypes.Float64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`amount expected to be basetypes.Float64Value, was: %T`, amountAttribute))
	}

	endAtAttribute, ok := attributes["end_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`end_at is missing from object`)

		return nil, diags
	}

	endAtVal, ok := endAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`end_at expected to be basetypes.StringValue, was: %T`, endAtAttribute))
	}

	startAtAttribute, ok := attributes["start_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`start_at is missing from object`)

		return nil, diags
	}

	startAtVal, ok := startAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`start_at expected to be basetypes.StringValue, was: %T`, startAtAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return PeriodsValue{
		Amount:  amountVal,
		EndAt:   endAtVal,
		StartAt: startAtVal,
		state:   attr.ValueStateKnown,
	}, diags
}

func NewPeriodsValueNull() PeriodsValue {
	return PeriodsValue{
		state: attr.ValueStateNull,
	}
}

func NewPeriodsValueUnknown() PeriodsValue {
	return PeriodsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewPeriodsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (PeriodsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing PeriodsValue Attribute Value",
				"While creating a PeriodsValue value, a missing attribute value was detected. "+
					"A PeriodsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("PeriodsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid PeriodsValue Attribute Type",
				"While creating a PeriodsValue value, an invalid attribute value was detected. "+
					"A PeriodsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("PeriodsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("PeriodsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra PeriodsValue Attribute Value",
				"While creating a PeriodsValue value, an extra attribute value was detected. "+
					"A PeriodsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra PeriodsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewPeriodsValueUnknown(), diags
	}

	amountAttribute, ok := attributes["amount"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`amount is missing from object`)

		return NewPeriodsValueUnknown(), diags
	}

	amountVal, ok := amountAttribute.(basetypes.Float64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`amount expected to be basetypes.Float64Value, was: %T`, amountAttribute))
	}

	endAtAttribute, ok := attributes["end_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`end_at is missing from object`)

		return NewPeriodsValueUnknown(), diags
	}

	endAtVal, ok := endAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`end_at expected to be basetypes.StringValue, was: %T`, endAtAttribute))
	}

	startAtAttribute, ok := attributes["start_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`start_at is missing from object`)

		return NewPeriodsValueUnknown(), diags
	}

	startAtVal, ok := startAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`start_at expected to be basetypes.StringValue, was: %T`, startAtAttribute))
	}

	if diags.HasError() {
		return NewPeriodsValueUnknown(), diags
	}

	return PeriodsValue{
		Amount:  amountVal,
		EndAt:   endAtVal,
		StartAt: startAtVal,
		state:   attr.ValueStateKnown,
	}, diags
}

func NewPeriodsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) PeriodsValue {
	object, diags := NewPeriodsValue(attributeTypes, attributes)

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

		panic("NewPeriodsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t PeriodsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewPeriodsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewPeriodsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewPeriodsValueNull(), nil
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

	return NewPeriodsValueMust(PeriodsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t PeriodsType) ValueType(ctx context.Context) attr.Value {
	return PeriodsValue{}
}

var _ basetypes.ObjectValuable = PeriodsValue{}

type PeriodsValue struct {
	Amount  basetypes.Float64Value `tfsdk:"amount"`
	EndAt   basetypes.StringValue  `tfsdk:"end_at"`
	StartAt basetypes.StringValue  `tfsdk:"start_at"`
	state   attr.ValueState
}

func (v PeriodsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)

	var val tftypes.Value
	var err error

	attrTypes["amount"] = basetypes.Float64Type{}.TerraformType(ctx)
	attrTypes["end_at"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["start_at"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err = v.Amount.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["amount"] = val

		val, err = v.EndAt.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["end_at"] = val

		val, err = v.StartAt.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["start_at"] = val

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

func (v PeriodsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v PeriodsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v PeriodsValue) String() string {
	return "PeriodsValue"
}

func (v PeriodsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"amount":   basetypes.Float64Type{},
			"end_at":   basetypes.StringType{},
			"start_at": basetypes.StringType{},
		},
		map[string]attr.Value{
			"amount":   v.Amount,
			"end_at":   v.EndAt,
			"start_at": v.StartAt,
		})

	return objVal, diags
}

func (v PeriodsValue) Equal(o attr.Value) bool {
	other, ok := o.(PeriodsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Amount.Equal(other.Amount) {
		return false
	}

	if !v.EndAt.Equal(other.EndAt) {
		return false
	}

	if !v.StartAt.Equal(other.StartAt) {
		return false
	}

	return true
}

func (v PeriodsValue) Type(ctx context.Context) attr.Type {
	return PeriodsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v PeriodsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"amount":   basetypes.Float64Type{},
		"end_at":   basetypes.StringType{},
		"start_at": basetypes.StringType{},
	}
}
