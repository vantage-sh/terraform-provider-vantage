// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package datasource_managed_accounts

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

func ManagedAccountsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"managed_accounts": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"access_credential_tokens": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "The tokens for the Access Credentials assigned to the Managed Account.",
							MarkdownDescription: "The tokens for the Access Credentials assigned to the Managed Account.",
						},
						"billing_rule_tokens": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "The tokens for the Billing Rules assigned to the Managed Account.",
							MarkdownDescription: "The tokens for the Billing Rules assigned to the Managed Account.",
						},
						"contact_email": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"parent_account_token": schema.StringAttribute{
							Computed:            true,
							Description:         "The token for the parent Account.",
							MarkdownDescription: "The token for the parent Account.",
						},
						"token": schema.StringAttribute{
							Computed: true,
						},
					},
					CustomType: ManagedAccountsType{
						ObjectType: types.ObjectType{
							AttrTypes: ManagedAccountsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed: true,
			},
		},
	}
}

type ManagedAccountsModel struct {
	ManagedAccounts types.List `tfsdk:"managed_accounts"`
}

var _ basetypes.ObjectTypable = ManagedAccountsType{}

type ManagedAccountsType struct {
	basetypes.ObjectType
}

func (t ManagedAccountsType) Equal(o attr.Type) bool {
	other, ok := o.(ManagedAccountsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ManagedAccountsType) String() string {
	return "ManagedAccountsType"
}

func (t ManagedAccountsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	accessCredentialTokensAttribute, ok := attributes["access_credential_tokens"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`access_credential_tokens is missing from object`)

		return nil, diags
	}

	accessCredentialTokensVal, ok := accessCredentialTokensAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`access_credential_tokens expected to be basetypes.ListValue, was: %T`, accessCredentialTokensAttribute))
	}

	billingRuleTokensAttribute, ok := attributes["billing_rule_tokens"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`billing_rule_tokens is missing from object`)

		return nil, diags
	}

	billingRuleTokensVal, ok := billingRuleTokensAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`billing_rule_tokens expected to be basetypes.ListValue, was: %T`, billingRuleTokensAttribute))
	}

	contactEmailAttribute, ok := attributes["contact_email"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`contact_email is missing from object`)

		return nil, diags
	}

	contactEmailVal, ok := contactEmailAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`contact_email expected to be basetypes.StringValue, was: %T`, contactEmailAttribute))
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

	parentAccountTokenAttribute, ok := attributes["parent_account_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`parent_account_token is missing from object`)

		return nil, diags
	}

	parentAccountTokenVal, ok := parentAccountTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`parent_account_token expected to be basetypes.StringValue, was: %T`, parentAccountTokenAttribute))
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

	if diags.HasError() {
		return nil, diags
	}

	return ManagedAccountsValue{
		AccessCredentialTokens: accessCredentialTokensVal,
		BillingRuleTokens:      billingRuleTokensVal,
		ContactEmail:           contactEmailVal,
		Name:                   nameVal,
		ParentAccountToken:     parentAccountTokenVal,
		Token:                  tokenVal,
		state:                  attr.ValueStateKnown,
	}, diags
}

func NewManagedAccountsValueNull() ManagedAccountsValue {
	return ManagedAccountsValue{
		state: attr.ValueStateNull,
	}
}

func NewManagedAccountsValueUnknown() ManagedAccountsValue {
	return ManagedAccountsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewManagedAccountsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (ManagedAccountsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing ManagedAccountsValue Attribute Value",
				"While creating a ManagedAccountsValue value, a missing attribute value was detected. "+
					"A ManagedAccountsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ManagedAccountsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid ManagedAccountsValue Attribute Type",
				"While creating a ManagedAccountsValue value, an invalid attribute value was detected. "+
					"A ManagedAccountsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ManagedAccountsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("ManagedAccountsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra ManagedAccountsValue Attribute Value",
				"While creating a ManagedAccountsValue value, an extra attribute value was detected. "+
					"A ManagedAccountsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra ManagedAccountsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewManagedAccountsValueUnknown(), diags
	}

	accessCredentialTokensAttribute, ok := attributes["access_credential_tokens"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`access_credential_tokens is missing from object`)

		return NewManagedAccountsValueUnknown(), diags
	}

	accessCredentialTokensVal, ok := accessCredentialTokensAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`access_credential_tokens expected to be basetypes.ListValue, was: %T`, accessCredentialTokensAttribute))
	}

	billingRuleTokensAttribute, ok := attributes["billing_rule_tokens"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`billing_rule_tokens is missing from object`)

		return NewManagedAccountsValueUnknown(), diags
	}

	billingRuleTokensVal, ok := billingRuleTokensAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`billing_rule_tokens expected to be basetypes.ListValue, was: %T`, billingRuleTokensAttribute))
	}

	contactEmailAttribute, ok := attributes["contact_email"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`contact_email is missing from object`)

		return NewManagedAccountsValueUnknown(), diags
	}

	contactEmailVal, ok := contactEmailAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`contact_email expected to be basetypes.StringValue, was: %T`, contactEmailAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewManagedAccountsValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	parentAccountTokenAttribute, ok := attributes["parent_account_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`parent_account_token is missing from object`)

		return NewManagedAccountsValueUnknown(), diags
	}

	parentAccountTokenVal, ok := parentAccountTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`parent_account_token expected to be basetypes.StringValue, was: %T`, parentAccountTokenAttribute))
	}

	tokenAttribute, ok := attributes["token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`token is missing from object`)

		return NewManagedAccountsValueUnknown(), diags
	}

	tokenVal, ok := tokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`token expected to be basetypes.StringValue, was: %T`, tokenAttribute))
	}

	if diags.HasError() {
		return NewManagedAccountsValueUnknown(), diags
	}

	return ManagedAccountsValue{
		AccessCredentialTokens: accessCredentialTokensVal,
		BillingRuleTokens:      billingRuleTokensVal,
		ContactEmail:           contactEmailVal,
		Name:                   nameVal,
		ParentAccountToken:     parentAccountTokenVal,
		Token:                  tokenVal,
		state:                  attr.ValueStateKnown,
	}, diags
}

func NewManagedAccountsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) ManagedAccountsValue {
	object, diags := NewManagedAccountsValue(attributeTypes, attributes)

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

		panic("NewManagedAccountsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t ManagedAccountsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewManagedAccountsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewManagedAccountsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewManagedAccountsValueNull(), nil
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

	return NewManagedAccountsValueMust(ManagedAccountsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t ManagedAccountsType) ValueType(ctx context.Context) attr.Value {
	return ManagedAccountsValue{}
}

var _ basetypes.ObjectValuable = ManagedAccountsValue{}

type ManagedAccountsValue struct {
	AccessCredentialTokens basetypes.ListValue   `tfsdk:"access_credential_tokens"`
	BillingRuleTokens      basetypes.ListValue   `tfsdk:"billing_rule_tokens"`
	ContactEmail           basetypes.StringValue `tfsdk:"contact_email"`
	Name                   basetypes.StringValue `tfsdk:"name"`
	ParentAccountToken     basetypes.StringValue `tfsdk:"parent_account_token"`
	Token                  basetypes.StringValue `tfsdk:"token"`
	state                  attr.ValueState
}

func (v ManagedAccountsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 6)

	var val tftypes.Value
	var err error

	attrTypes["access_credential_tokens"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["billing_rule_tokens"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["contact_email"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["parent_account_token"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["token"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 6)

		val, err = v.AccessCredentialTokens.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["access_credential_tokens"] = val

		val, err = v.BillingRuleTokens.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["billing_rule_tokens"] = val

		val, err = v.ContactEmail.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["contact_email"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		val, err = v.ParentAccountToken.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["parent_account_token"] = val

		val, err = v.Token.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["token"] = val

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

func (v ManagedAccountsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ManagedAccountsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ManagedAccountsValue) String() string {
	return "ManagedAccountsValue"
}

func (v ManagedAccountsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	accessCredentialTokensVal, d := types.ListValue(types.StringType, v.AccessCredentialTokens.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"access_credential_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
			"billing_rule_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
			"contact_email":        basetypes.StringType{},
			"name":                 basetypes.StringType{},
			"parent_account_token": basetypes.StringType{},
			"token":                basetypes.StringType{},
		}), diags
	}

	billingRuleTokensVal, d := types.ListValue(types.StringType, v.BillingRuleTokens.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"access_credential_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
			"billing_rule_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
			"contact_email":        basetypes.StringType{},
			"name":                 basetypes.StringType{},
			"parent_account_token": basetypes.StringType{},
			"token":                basetypes.StringType{},
		}), diags
	}

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"access_credential_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
			"billing_rule_tokens": basetypes.ListType{
				ElemType: types.StringType,
			},
			"contact_email":        basetypes.StringType{},
			"name":                 basetypes.StringType{},
			"parent_account_token": basetypes.StringType{},
			"token":                basetypes.StringType{},
		},
		map[string]attr.Value{
			"access_credential_tokens": accessCredentialTokensVal,
			"billing_rule_tokens":      billingRuleTokensVal,
			"contact_email":            v.ContactEmail,
			"name":                     v.Name,
			"parent_account_token":     v.ParentAccountToken,
			"token":                    v.Token,
		})

	return objVal, diags
}

func (v ManagedAccountsValue) Equal(o attr.Value) bool {
	other, ok := o.(ManagedAccountsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.AccessCredentialTokens.Equal(other.AccessCredentialTokens) {
		return false
	}

	if !v.BillingRuleTokens.Equal(other.BillingRuleTokens) {
		return false
	}

	if !v.ContactEmail.Equal(other.ContactEmail) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	if !v.ParentAccountToken.Equal(other.ParentAccountToken) {
		return false
	}

	if !v.Token.Equal(other.Token) {
		return false
	}

	return true
}

func (v ManagedAccountsValue) Type(ctx context.Context) attr.Type {
	return ManagedAccountsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v ManagedAccountsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"access_credential_tokens": basetypes.ListType{
			ElemType: types.StringType,
		},
		"billing_rule_tokens": basetypes.ListType{
			ElemType: types.StringType,
		},
		"contact_email":        basetypes.StringType{},
		"name":                 basetypes.StringType{},
		"parent_account_token": basetypes.StringType{},
		"token":                basetypes.StringType{},
	}
}
