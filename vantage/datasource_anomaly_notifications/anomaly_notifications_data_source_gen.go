// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package datasource_anomaly_notifications

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func AnomalyNotificationsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"anomaly_notifications": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cost_report_token": schema.StringAttribute{
							Computed:            true,
							Description:         "The token for the CostReport the AnomalyNotification is associated with.",
							MarkdownDescription: "The token for the CostReport the AnomalyNotification is associated with.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time, in UTC, the AnomalyNotification was created. ISO 8601 Formatted.",
							MarkdownDescription: "The date and time, in UTC, the AnomalyNotification was created. ISO 8601 Formatted.",
						},
						"recipient_channels": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "The channels that the notification is sent to.",
							MarkdownDescription: "The channels that the notification is sent to.",
						},
						"threshold": schema.Int64Attribute{
							Computed:            true,
							Description:         "The threshold amount that must be met for the notification to fire.",
							MarkdownDescription: "The threshold amount that must be met for the notification to fire.",
						},
						"token": schema.StringAttribute{
							Computed: true,
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							Description:         "The date and time, in UTC, the AnomalyNotification was last updated at. ISO 8601 Formatted.",
							MarkdownDescription: "The date and time, in UTC, the AnomalyNotification was last updated at. ISO 8601 Formatted.",
						},
						"user_tokens": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "The tokens of the users that receive the notification.",
							MarkdownDescription: "The tokens of the users that receive the notification.",
						},
					},
					CustomType: AnomalyNotificationsType{
						ObjectType: types.ObjectType{
							AttrTypes: AnomalyNotificationsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed: true,
			},
		},
	}
}

type AnomalyNotificationsModel struct {
	AnomalyNotifications types.List `tfsdk:"anomaly_notifications"`
}

var _ basetypes.ObjectTypable = AnomalyNotificationsType{}

type AnomalyNotificationsType struct {
	basetypes.ObjectType
}

func (t AnomalyNotificationsType) Equal(o attr.Type) bool {
	other, ok := o.(AnomalyNotificationsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t AnomalyNotificationsType) String() string {
	return "AnomalyNotificationsType"
}

func (t AnomalyNotificationsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	costReportTokenAttribute, ok := attributes["cost_report_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`cost_report_token is missing from object`)

		return nil, diags
	}

	costReportTokenVal, ok := costReportTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`cost_report_token expected to be basetypes.StringValue, was: %T`, costReportTokenAttribute))
	}

	createdAtAttribute, ok := attributes["created_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`created_at is missing from object`)

		return nil, diags
	}

	createdAtVal, ok := createdAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`created_at expected to be basetypes.StringValue, was: %T`, createdAtAttribute))
	}

	recipientChannelsAttribute, ok := attributes["recipient_channels"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`recipient_channels is missing from object`)

		return nil, diags
	}

	recipientChannelsVal, ok := recipientChannelsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`recipient_channels expected to be basetypes.ListValue, was: %T`, recipientChannelsAttribute))
	}

	thresholdAttribute, ok := attributes["threshold"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`threshold is missing from object`)

		return nil, diags
	}

	thresholdVal, ok := thresholdAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`threshold expected to be basetypes.Int64Value, was: %T`, thresholdAttribute))
	}

	tokenAttribute, ok := attributes["token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`token is missing from object`)

		return nil, diags
	}

	tokenVal, ok := tokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`token expected to be basetypes.StringValue, was: %T`, tokenAttribute))
	}

	updatedAtAttribute, ok := attributes["updated_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`updated_at is missing from object`)

		return nil, diags
	}

	updatedAtVal, ok := updatedAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`updated_at expected to be basetypes.StringValue, was: %T`, updatedAtAttribute))
	}

	userTokensAttribute, ok := attributes["user_tokens"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`user_tokens is missing from object`)

		return nil, diags
	}

	userTokensVal, ok := userTokensAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`user_tokens expected to be basetypes.ListValue, was: %T`, userTokensAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return AnomalyNotificationsValue{
		CostReportToken:   costReportTokenVal,
		CreatedAt:         createdAtVal,
		RecipientChannels: recipientChannelsVal,
		Threshold:         thresholdVal,
		Token:             tokenVal,
		UpdatedAt:         updatedAtVal,
		UserTokens:        userTokensVal,
		state:             attr.ValueStateKnown,
	}, diags
}

func NewAnomalyNotificationsValueNull() AnomalyNotificationsValue {
	return AnomalyNotificationsValue{
		state: attr.ValueStateNull,
	}
}

func NewAnomalyNotificationsValueUnknown() AnomalyNotificationsValue {
	return AnomalyNotificationsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewAnomalyNotificationsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (AnomalyNotificationsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing AnomalyNotificationsValue Attribute Value",
				"While creating a AnomalyNotificationsValue value, a missing attribute value was detected. "+
					"A AnomalyNotificationsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AnomalyNotificationsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid AnomalyNotificationsValue Attribute Type",
				"While creating a AnomalyNotificationsValue value, an invalid attribute value was detected. "+
					"A AnomalyNotificationsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AnomalyNotificationsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("AnomalyNotificationsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra AnomalyNotificationsValue Attribute Value",
				"While creating a AnomalyNotificationsValue value, an extra attribute value was detected. "+
					"A AnomalyNotificationsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra AnomalyNotificationsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewAnomalyNotificationsValueUnknown(), diags
	}

	costReportTokenAttribute, ok := attributes["cost_report_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`cost_report_token is missing from object`)

		return NewAnomalyNotificationsValueUnknown(), diags
	}

	costReportTokenVal, ok := costReportTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`cost_report_token expected to be basetypes.StringValue, was: %T`, costReportTokenAttribute))
	}

	createdAtAttribute, ok := attributes["created_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`created_at is missing from object`)

		return NewAnomalyNotificationsValueUnknown(), diags
	}

	createdAtVal, ok := createdAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`created_at expected to be basetypes.StringValue, was: %T`, createdAtAttribute))
	}

	recipientChannelsAttribute, ok := attributes["recipient_channels"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`recipient_channels is missing from object`)

		return NewAnomalyNotificationsValueUnknown(), diags
	}

	recipientChannelsVal, ok := recipientChannelsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`recipient_channels expected to be basetypes.ListValue, was: %T`, recipientChannelsAttribute))
	}

	thresholdAttribute, ok := attributes["threshold"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`threshold is missing from object`)

		return NewAnomalyNotificationsValueUnknown(), diags
	}

	thresholdVal, ok := thresholdAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`threshold expected to be basetypes.Int64Value, was: %T`, thresholdAttribute))
	}

	tokenAttribute, ok := attributes["token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`token is missing from object`)

		return NewAnomalyNotificationsValueUnknown(), diags
	}

	tokenVal, ok := tokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`token expected to be basetypes.StringValue, was: %T`, tokenAttribute))
	}

	updatedAtAttribute, ok := attributes["updated_at"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`updated_at is missing from object`)

		return NewAnomalyNotificationsValueUnknown(), diags
	}

	updatedAtVal, ok := updatedAtAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`updated_at expected to be basetypes.StringValue, was: %T`, updatedAtAttribute))
	}

	userTokensAttribute, ok := attributes["user_tokens"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`user_tokens is missing from object`)

		return NewAnomalyNotificationsValueUnknown(), diags
	}

	userTokensVal, ok := userTokensAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`user_tokens expected to be basetypes.ListValue, was: %T`, userTokensAttribute))
	}

	if diags.HasError() {
		return NewAnomalyNotificationsValueUnknown(), diags
	}

	return AnomalyNotificationsValue{
		CostReportToken:   costReportTokenVal,
		CreatedAt:         createdAtVal,
		RecipientChannels: recipientChannelsVal,
		Threshold:         thresholdVal,
		Token:             tokenVal,
		UpdatedAt:         updatedAtVal,
		UserTokens:        userTokensVal,
		state:             attr.ValueStateKnown,
	}, diags
}

func NewAnomalyNotificationsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) AnomalyNotificationsValue {
	object, diags := NewAnomalyNotificationsValue(attributeTypes, attributes)

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

		panic("NewAnomalyNotificationsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t AnomalyNotificationsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewAnomalyNotificationsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewAnomalyNotificationsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewAnomalyNotificationsValueNull(), nil
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

	return NewAnomalyNotificationsValueMust(AnomalyNotificationsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t AnomalyNotificationsType) ValueType(ctx context.Context) attr.Value {
	return AnomalyNotificationsValue{}
}

var _ basetypes.ObjectValuable = AnomalyNotificationsValue{}

type AnomalyNotificationsValue struct {
	CostReportToken   basetypes.StringValue `tfsdk:"cost_report_token"`
	CreatedAt         basetypes.StringValue `tfsdk:"created_at"`
	RecipientChannels basetypes.ListValue   `tfsdk:"recipient_channels"`
	Threshold         basetypes.Int64Value  `tfsdk:"threshold"`
	Token             basetypes.StringValue `tfsdk:"token"`
	UpdatedAt         basetypes.StringValue `tfsdk:"updated_at"`
	UserTokens        basetypes.ListValue   `tfsdk:"user_tokens"`
	state             attr.ValueState
}

func (v AnomalyNotificationsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 7)

	var val tftypes.Value
	var err error

	attrTypes["cost_report_token"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["created_at"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["recipient_channels"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["threshold"] = basetypes.Int64Type{}.TerraformType(ctx)
	attrTypes["token"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["updated_at"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["user_tokens"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 7)

		val, err = v.CostReportToken.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["cost_report_token"] = val

		val, err = v.CreatedAt.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["created_at"] = val

		val, err = v.RecipientChannels.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["recipient_channels"] = val

		val, err = v.Threshold.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["threshold"] = val

		val, err = v.Token.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["token"] = val

		val, err = v.UpdatedAt.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["updated_at"] = val

		val, err = v.UserTokens.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["user_tokens"] = val

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

func (v AnomalyNotificationsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v AnomalyNotificationsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v AnomalyNotificationsValue) String() string {
	return "AnomalyNotificationsValue"
}

func (v AnomalyNotificationsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	recipientChannelsVal, d := types.ListValue(types.StringType, v.RecipientChannels.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"cost_report_token": basetypes.StringType{},
			"created_at":        basetypes.StringType{},
			"recipient_channels": basetypes.ListType{
				ElemType: types.StringType,
			},
			"threshold":  basetypes.Int64Type{},
			"token":      basetypes.StringType{},
			"updated_at": basetypes.StringType{},
			"user_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	userTokensVal, d := types.ListValue(types.StringType, v.UserTokens.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"cost_report_token": basetypes.StringType{},
			"created_at":        basetypes.StringType{},
			"recipient_channels": basetypes.ListType{
				ElemType: types.StringType,
			},
			"threshold":  basetypes.Int64Type{},
			"token":      basetypes.StringType{},
			"updated_at": basetypes.StringType{},
			"user_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"cost_report_token": basetypes.StringType{},
			"created_at":        basetypes.StringType{},
			"recipient_channels": basetypes.ListType{
				ElemType: types.StringType,
			},
			"threshold":  basetypes.Int64Type{},
			"token":      basetypes.StringType{},
			"updated_at": basetypes.StringType{},
			"user_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
		},
		map[string]attr.Value{
			"cost_report_token":  v.CostReportToken,
			"created_at":         v.CreatedAt,
			"recipient_channels": recipientChannelsVal,
			"threshold":          v.Threshold,
			"token":              v.Token,
			"updated_at":         v.UpdatedAt,
			"user_tokens":        userTokensVal,
		})

	return objVal, diags
}

func (v AnomalyNotificationsValue) Equal(o attr.Value) bool {
	other, ok := o.(AnomalyNotificationsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.CostReportToken.Equal(other.CostReportToken) {
		return false
	}

	if !v.CreatedAt.Equal(other.CreatedAt) {
		return false
	}

	if !v.RecipientChannels.Equal(other.RecipientChannels) {
		return false
	}

	if !v.Threshold.Equal(other.Threshold) {
		return false
	}

	if !v.Token.Equal(other.Token) {
		return false
	}

	if !v.UpdatedAt.Equal(other.UpdatedAt) {
		return false
	}

	if !v.UserTokens.Equal(other.UserTokens) {
		return false
	}

	return true
}

func (v AnomalyNotificationsValue) Type(ctx context.Context) attr.Type {
	return AnomalyNotificationsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v AnomalyNotificationsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"cost_report_token": basetypes.StringType{},
		"created_at":        basetypes.StringType{},
		"recipient_channels": basetypes.ListType{
			ElemType: types.StringType,
		},
		"threshold":  basetypes.Int64Type{},
		"token":      basetypes.StringType{},
		"updated_at": basetypes.StringType{},
		"user_tokens": basetypes.ListType{
			ElemType: types.StringType,
		},
	}
}