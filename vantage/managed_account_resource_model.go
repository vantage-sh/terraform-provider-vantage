package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_managed_accounts"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_managed_account"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type managedAccountModel resource_managed_account.ManagedAccountModel

// managedAccountDataSourceModel is used for the data source and includes
// all read-only fields that are not part of the resource model.
type managedAccountDataSourceModel struct {
	AccessCredentialTokens        types.List   `tfsdk:"access_credential_tokens"`
	BillingInformationAttributes  types.Object `tfsdk:"billing_information_attributes"`
	BillingRuleTokens             types.List   `tfsdk:"billing_rule_tokens"`
	BusinessInformationAttributes types.Object `tfsdk:"business_information_attributes"`
	ContactEmail                  types.String `tfsdk:"contact_email"`
	EmailDomain                   types.String `tfsdk:"email_domain"`
	Id                            types.String `tfsdk:"id"`
	MspBillingProfileToken        types.String `tfsdk:"msp_billing_profile_token"`
	Name                          types.String `tfsdk:"name"`
	ParentAccountToken            types.String `tfsdk:"parent_account_token"`
	Token                         types.String `tfsdk:"token"`
}

// applyPayloadDataSource populates the data source model from the API response.
func (m *managedAccountDataSourceModel) applyPayloadDataSource(ctx context.Context, payload *modelsv2.ManagedAccount) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Name = types.StringValue(payload.Name)
	m.ContactEmail = types.StringValue(payload.ContactEmail)
	m.Id = types.StringValue(payload.Token)
	m.Token = types.StringValue(payload.Token)
	m.ParentAccountToken = types.StringValue(payload.ParentAccountToken)
	m.EmailDomain = types.StringValue(payload.EmailDomain)
	m.MspBillingProfileToken = types.StringValue(payload.MspBillingProfileToken)

	// Handle access_credential_tokens
	if payload.AccessCredentialTokens != nil {
		accessCredentialTokens, d := types.ListValueFrom(ctx, types.StringType, payload.AccessCredentialTokens)
		diags.Append(d...)
		m.AccessCredentialTokens = accessCredentialTokens
	} else {
		m.AccessCredentialTokens = types.ListNull(types.StringType)
	}

	// Handle billing_rule_tokens
	if payload.BillingRuleTokens != nil {
		billingRuleTokens, d := types.ListValueFrom(ctx, types.StringType, payload.BillingRuleTokens)
		diags.Append(d...)
		m.BillingRuleTokens = billingRuleTokens
	} else {
		m.BillingRuleTokens = types.ListNull(types.StringType)
	}

	// Handle billing_information_attributes
	if payload.BillingInformationAttributes != nil {
		billingInfo, d := buildBillingInformationObject(ctx, payload.BillingInformationAttributes)
		diags.Append(d...)
		m.BillingInformationAttributes = billingInfo
	} else {
		m.BillingInformationAttributes = types.ObjectNull(
			datasource_managed_accounts.BillingInformationAttributesValue{}.AttributeTypes(ctx),
		)
	}

	// Handle business_information_attributes
	if payload.BusinessInformationAttributes != nil {
		businessInfo, d := buildBusinessInformationObject(ctx, payload.BusinessInformationAttributes)
		diags.Append(d...)
		m.BusinessInformationAttributes = businessInfo
	} else {
		m.BusinessInformationAttributes = types.ObjectNull(
			datasource_managed_accounts.BusinessInformationAttributesValue{}.AttributeTypes(ctx),
		)
	}

	return diags
}

// buildBillingInformationObject converts API billing information to a Terraform object.
func buildBillingInformationObject(ctx context.Context, info *modelsv2.BillingInformation) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	billingEmailList := types.ListNull(types.StringType)
	if info.BillingEmail != nil {
		var d diag.Diagnostics
		billingEmailList, d = types.ListValueFrom(ctx, types.StringType, info.BillingEmail)
		diags.Append(d...)
	}

	attrTypes := datasource_managed_accounts.BillingInformationAttributesValue{}.AttributeTypes(ctx)
	attrValues := map[string]attr.Value{
		"address_line_1": types.StringPointerValue(info.AddressLine1),
		"address_line_2": types.StringPointerValue(info.AddressLine2),
		"billing_email":  billingEmailList,
		"city":           types.StringPointerValue(info.City),
		"company_name":   types.StringPointerValue(info.CompanyName),
		"country_code":   types.StringPointerValue(info.CountryCode),
		"postal_code":    types.StringPointerValue(info.PostalCode),
		"state":          types.StringPointerValue(info.State),
		"token":          types.StringValue(info.Token),
	}

	billingInfoValue, d := datasource_managed_accounts.NewBillingInformationAttributesValue(attrTypes, attrValues)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(attrTypes), diags
	}

	obj, d := billingInfoValue.ToObjectValue(ctx)
	diags.Append(d...)
	return obj, diags
}

// buildBusinessInformationObject converts API business information to a Terraform object.
func buildBusinessInformationObject(ctx context.Context, info *modelsv2.BusinessInformation) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Build metadata value
	metadataValue := datasource_managed_accounts.NewMetadataValueNull()
	if info.Metadata != nil {
		var d diag.Diagnostics
		metadataValue, d = buildMetadataValue(ctx, info.Metadata)
		diags.Append(d...)
	}

	// Convert metadata value to object for use in attributes
	metadataObj, d := metadataValue.ToObjectValue(ctx)
	diags.Append(d...)

	attrTypes := datasource_managed_accounts.BusinessInformationAttributesValue{}.AttributeTypes(ctx)
	attrValues := map[string]attr.Value{
		"metadata": metadataObj,
		"token":    types.StringValue(info.Token),
	}

	businessInfoValue, d := datasource_managed_accounts.NewBusinessInformationAttributesValue(attrTypes, attrValues)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(attrTypes), diags
	}

	obj, d := businessInfoValue.ToObjectValue(ctx)
	diags.Append(d...)
	return obj, diags
}

// buildMetadataValue converts API metadata to a MetadataValue.
func buildMetadataValue(ctx context.Context, metadata *modelsv2.BusinessInformationMetadata) (datasource_managed_accounts.MetadataValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Build custom_fields list using the proper CustomFieldsType
	customFieldsElemType := datasource_managed_accounts.CustomFieldsValue{}.Type(ctx)
	customFieldsListType := types.ListType{ElemType: customFieldsElemType}
	customFieldsList := types.ListNull(customFieldsElemType)

	if metadata.CustomFields != nil && len(metadata.CustomFields) > 0 {
		customFieldsValues := make([]attr.Value, 0, len(metadata.CustomFields))
		for _, cf := range metadata.CustomFields {
			cfAttrTypes := datasource_managed_accounts.CustomFieldsValue{}.AttributeTypes(ctx)
			cfAttrValues := map[string]attr.Value{
				"name":  types.StringValue(cf.Name),
				"value": types.StringPointerValue(cf.Value),
			}
			cfValue, d := datasource_managed_accounts.NewCustomFieldsValue(cfAttrTypes, cfAttrValues)
			diags.Append(d...)
			if diags.HasError() {
				return datasource_managed_accounts.NewMetadataValueNull(), diags
			}

			// Use CustomFieldsValue directly - it implements attr.Value
			customFieldsValues = append(customFieldsValues, cfValue)
		}
		var d diag.Diagnostics
		customFieldsList, d = types.ListValue(customFieldsElemType, customFieldsValues)
		diags.Append(d...)
	}

	attrTypes := map[string]attr.Type{
		"custom_fields": customFieldsListType,
	}
	attrValues := map[string]attr.Value{
		"custom_fields": customFieldsList,
	}

	metadataValue, d := datasource_managed_accounts.NewMetadataValue(attrTypes, attrValues)
	diags.Append(d...)
	return metadataValue, diags
}

func (m *managedAccountModel) applyPayload(ctx context.Context, payload *modelsv2.ManagedAccount, isDataSource bool) diag.Diagnostics {
	m.Name = types.StringValue(payload.Name)
	m.ContactEmail = types.StringValue(payload.ContactEmail)
	m.Id = types.StringValue(payload.Token) // Set ID to the token for resource identification
	// Only apply access_credential_tokens from API response for data sources (reads/imports)
	// For resource operations, keep the planned value to avoid delegation/mixed credential issues. i.e. The endpoint will attempt to return all child access credentials and not just the parent access credentials that are being delegated.
	if isDataSource && payload.AccessCredentialTokens != nil {
		accessCredentialTokens, diag := types.ListValueFrom(ctx, types.StringType, payload.AccessCredentialTokens)
		if diag.HasError() {
			return diag
		}
		m.AccessCredentialTokens = accessCredentialTokens
	} else if !isDataSource && m.AccessCredentialTokens.IsUnknown() {
		m.AccessCredentialTokens = types.ListNull(types.StringType)
	}

	// Only apply billing_rule_tokens from API response for data sources (reads/imports)
	// For resource operations, keep the planned value to avoid order/consistency issues
	if isDataSource && payload.BillingRuleTokens != nil {
		billingRuleTokens, diag := types.ListValueFrom(ctx, types.StringType, payload.BillingRuleTokens)
		if diag.HasError() {
			return diag
		}
		m.BillingRuleTokens = billingRuleTokens
	} else if !isDataSource && m.BillingRuleTokens.IsUnknown() {
		m.BillingRuleTokens = types.ListNull(types.StringType)
	}

	m.ParentAccountToken = types.StringValue(payload.ParentAccountToken)
	m.Token = types.StringValue(payload.Token)
	m.EmailDomain = types.StringValue(payload.EmailDomain)

	return nil
}

func (m *managedAccountModel) toCreateModel(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateManagedAccount {
	dst := &modelsv2.CreateManagedAccount{
		Name:         m.Name.ValueStringPointer(),
		ContactEmail: m.ContactEmail.ValueStringPointer(),
	}

	if !m.EmailDomain.IsNull() && !m.EmailDomain.IsUnknown() {
		dst.EmailDomain = m.EmailDomain.ValueString()
	}

	if !m.AccessCredentialTokens.IsNull() && !m.AccessCredentialTokens.IsUnknown() {
		accessCredentialTokens := []string{}
		m.AccessCredentialTokens.ElementsAs(ctx, &accessCredentialTokens, false)
		dst.AccessCredentialTokens = accessCredentialTokens
	}

	if !m.BillingRuleTokens.IsNull() && !m.BillingRuleTokens.IsUnknown() {
		billingRuleTokens := []string{}
		m.BillingRuleTokens.ElementsAs(ctx, &billingRuleTokens, false)
		dst.BillingRuleTokens = billingRuleTokens
	}

	return dst
}

func (m *managedAccountModel) toUpdateModel(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateManagedAccount {
	dst := &modelsv2.UpdateManagedAccount{}

	if !m.Name.IsNull() {
		dst.Name = m.Name.ValueString()
	}

	if !m.ContactEmail.IsNull() {
		dst.ContactEmail = m.ContactEmail.ValueString()
	}

	if !m.AccessCredentialTokens.IsNull() && !m.AccessCredentialTokens.IsUnknown() {
		accessCredentialTokens := []string{}
		m.AccessCredentialTokens.ElementsAs(ctx, &accessCredentialTokens, false)
		dst.AccessCredentialTokens = accessCredentialTokens
	} else {
		dst.AccessCredentialTokens = []string{}
	}

	if !m.BillingRuleTokens.IsNull() && !m.BillingRuleTokens.IsUnknown() {
		billingRuleTokens := []string{}
		m.BillingRuleTokens.ElementsAs(ctx, &billingRuleTokens, false)
		dst.BillingRuleTokens = billingRuleTokens
	} else {
		dst.BillingRuleTokens = []string{}
	}

	if !m.EmailDomain.IsNull() && !m.EmailDomain.IsUnknown() {
		dst.EmailDomain = m.EmailDomain.ValueString()
	}

	return dst
}
