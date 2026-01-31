package vantage

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_billing_profile"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type billingProfileModel resource_billing_profile.BillingProfileModel

func (m *billingProfileModel) applyPayload(ctx context.Context, payload *modelsv2.BillingProfile) diag.Diagnostics {
	var diags diag.Diagnostics

	// Handle Banking Information Attributes
	if payload.BankingInformationAttributes != nil {
		// Handle SecureData
		var secureDataAttr attr.Value
		if payload.BankingInformationAttributes.SecureData != nil {
			secureDataAttrs := map[string]attr.Value{
				"account_number": types.StringValue(payload.BankingInformationAttributes.SecureData.AccountNumber),
				"routing_number": types.StringValue(payload.BankingInformationAttributes.SecureData.RoutingNumber),
				"iban":          types.StringValue(payload.BankingInformationAttributes.SecureData.Iban),
				"swift_bic":     types.StringValue(payload.BankingInformationAttributes.SecureData.SwiftBic),
			}
			secureDataObj, diag := types.ObjectValue(map[string]attr.Type{
				"account_number": types.StringType,
				"routing_number": types.StringType,
				"iban":          types.StringType,
				"swift_bic":     types.StringType,
			}, secureDataAttrs)
			if diag.HasError() {
				diags.Append(diag...)
				return diags
			}
			secureDataAttr = secureDataObj
		} else {
			secureDataAttr = types.ObjectNull(map[string]attr.Type{
				"account_number": types.StringType,
				"routing_number": types.StringType,
				"iban":          types.StringType,
				"swift_bic":     types.StringType,
			})
		}
		
		// Use the generated constructor to create the proper value with state
		bankingAttrTypes := map[string]attr.Type{
			"bank_name":        types.StringType,
			"beneficiary_name": types.StringType,
			"tax_id":           types.StringType,
			"token":            types.StringType,
			"secure_data":      types.ObjectType{AttrTypes: map[string]attr.Type{
				"account_number": types.StringType,
				"routing_number": types.StringType,
				"iban":          types.StringType,
				"swift_bic":     types.StringType,
			}},
		}
		
		bankingAttrs := map[string]attr.Value{
			"bank_name":        types.StringValue(payload.BankingInformationAttributes.BankName),
			"beneficiary_name": types.StringValue(payload.BankingInformationAttributes.BeneficiaryName),
			"tax_id":           types.StringValue(payload.BankingInformationAttributes.TaxID),
			"token":            types.StringValue(payload.BankingInformationAttributes.Token),
			"secure_data":      secureDataAttr,
		}
		
		bankingInfo, diag := resource_billing_profile.NewBankingInformationAttributesValue(bankingAttrTypes, bankingAttrs)
		if diag.HasError() {
			diags.Append(diag...)
			return diags
		}

		m.BankingInformationAttributes = bankingInfo
	}
	// Note: When API doesn't return banking_information_attributes, we preserve
	// the existing planned values by not modifying m.BankingInformationAttributes

	// Handle Billing Information Attributes
	if payload.BillingInformationAttributes != nil {
		billingEmails := []attr.Value{}
		if payload.BillingInformationAttributes.BillingEmail != nil {
			for _, email := range payload.BillingInformationAttributes.BillingEmail {
				billingEmails = append(billingEmails, types.StringValue(email))
			}
		}

		billingEmailsList, diag := types.ListValue(types.StringType, billingEmails)
		if diag.HasError() {
			diags.Append(diag...)
			return diags
		}

		// Use the generated constructor to create the proper value with state
		attrTypes := map[string]attr.Type{
			"address_line_1": types.StringType,
			"address_line_2": types.StringType,
			"billing_email":  types.ListType{ElemType: types.StringType},
			"city":           types.StringType,
			"company_name":   types.StringType,
			"country_code":   types.StringType,
			"postal_code":    types.StringType,
			"state":          types.StringType,
			"token":          types.StringType,
		}
		
		attrs := map[string]attr.Value{
			"address_line_1": types.StringValue(payload.BillingInformationAttributes.AddressLine1),
			"address_line_2": types.StringValue(payload.BillingInformationAttributes.AddressLine2),
			"billing_email":  billingEmailsList,
			"city":           types.StringValue(payload.BillingInformationAttributes.City),
			"company_name":   types.StringValue(payload.BillingInformationAttributes.CompanyName),
			"country_code":   types.StringValue(payload.BillingInformationAttributes.CountryCode),
			"postal_code":    types.StringValue(payload.BillingInformationAttributes.PostalCode),
			"state":          types.StringValue(payload.BillingInformationAttributes.State),
			"token":          types.StringValue(payload.BillingInformationAttributes.Token),
		}
		
		billingInfo, diag := resource_billing_profile.NewBillingInformationAttributesValue(attrTypes, attrs)
		if diag.HasError() {
			diags.Append(diag...)
			return diags
		}

		m.BillingInformationAttributes = billingInfo
	}
	// Note: When API doesn't return billing_information_attributes, we preserve
	// the existing planned values by not modifying m.BillingInformationAttributes

	// Handle Business Information Attributes
	if payload.BusinessInformationAttributes != nil {
		// Handle metadata
		var metadataAttr attr.Value
		if payload.BusinessInformationAttributes.Metadata != nil {
			customFieldsList := []attr.Value{}
			if payload.BusinessInformationAttributes.Metadata.CustomFields != nil {
				for _, field := range payload.BusinessInformationAttributes.Metadata.CustomFields {
					fieldAttrs := map[string]attr.Value{
						"name":  types.StringValue(field.Name),
						"value": types.StringValue(field.Value),
					}
					fieldObj, diag := types.ObjectValue(
						map[string]attr.Type{
							"name":  types.StringType,
							"value": types.StringType,
						},
						fieldAttrs,
					)
					if diag.HasError() {
						diags.Append(diag...)
						return diags
					}
					customFieldsList = append(customFieldsList, fieldObj)
				}
			}
			
			customFieldsListValue, diag := types.ListValue(
				types.ObjectType{AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				}},
				customFieldsList,
			)
			if diag.HasError() {
				diags.Append(diag...)
				return diags
			}
			
			metadataObj, diag := types.ObjectValue(
				map[string]attr.Type{
					"custom_fields": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
					}}},
				},
				map[string]attr.Value{
					"custom_fields": customFieldsListValue,
				},
			)
			if diag.HasError() {
				diags.Append(diag...)
				return diags
			}
			metadataAttr = metadataObj
		} else {
			metadataAttr = types.ObjectNull(map[string]attr.Type{
				"custom_fields": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				}}},
			})
		}
		
		// Use the generated constructor to create the proper value with state
		businessAttrTypes := map[string]attr.Type{
			"token":    types.StringType,
			"metadata": types.ObjectType{AttrTypes: map[string]attr.Type{
				"custom_fields": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				}}},
			}},
		}
		
		businessAttrs := map[string]attr.Value{
			"token":    types.StringValue(payload.BusinessInformationAttributes.Token),
			"metadata": metadataAttr,
		}
		
		businessInfo, diag := resource_billing_profile.NewBusinessInformationAttributesValue(businessAttrTypes, businessAttrs)
		if diag.HasError() {
			diags.Append(diag...)
			return diags
		}

		m.BusinessInformationAttributes = businessInfo
	}
	// Note: When API doesn't return business_information_attributes, we preserve
	// the existing planned values by not modifying m.BusinessInformationAttributes

	// Handle Invoice Adjustment Attributes
	if payload.InvoiceAdjustmentAttributes != nil {
		// Build the adjustment items list using the generated types
		adjustmentItemsList := []attr.Value{}
		if payload.InvoiceAdjustmentAttributes.AdjustmentItems != nil {
			for _, item := range payload.InvoiceAdjustmentAttributes.AdjustmentItems {
				// Parse amount string to float64
				var amountVal basetypes.Float64Value
				if item.Amount != "" {
					if amount, err := strconv.ParseFloat(item.Amount, 64); err == nil {
						amountVal = types.Float64Value(amount)
					} else {
						amountVal = types.Float64Null()
					}
				} else {
					amountVal = types.Float64Null()
				}

				// Use the generated constructor for proper type matching
				itemVal, d := resource_billing_profile.NewAdjustmentItemsValue(
					resource_billing_profile.AdjustmentItemsValue{}.AttributeTypes(ctx),
					map[string]attr.Value{
						"adjustment_type":  types.StringValue(item.AdjustmentType),
						"amount":           amountVal,
						"calculation_type": types.StringValue(item.CalculationType),
						"name":             types.StringValue(item.Name),
					},
				)
				if d.HasError() {
					diags.Append(d...)
					return diags
				}

				// Keep as AdjustmentItemsValue (don't convert to ObjectValue)
				// The custom type is required for proper list element matching
				adjustmentItemsList = append(adjustmentItemsList, itemVal)
			}
		}

		// Use the generated type for the list element type
		adjustmentItemsListValue, d := types.ListValue(
			resource_billing_profile.AdjustmentItemsType{
				basetypes.ObjectType{AttrTypes: resource_billing_profile.AdjustmentItemsValue{}.AttributeTypes(ctx)},
			},
			adjustmentItemsList,
		)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}

		invoiceAdjInfo, d := resource_billing_profile.NewInvoiceAdjustmentAttributesValue(
			resource_billing_profile.InvoiceAdjustmentAttributesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"adjustment_items": adjustmentItemsListValue,
				"token":            types.StringValue(payload.InvoiceAdjustmentAttributes.Token),
			},
		)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}

		m.InvoiceAdjustmentAttributes = invoiceAdjInfo
	}
	// Note: When API doesn't return invoice_adjustment_attributes, we preserve
	// the existing planned values by not modifying m.InvoiceAdjustmentAttributes

	// Handle simple attributes
	m.CreatedAt = types.StringPointerValue(&payload.CreatedAt)
	m.Id = types.StringPointerValue(&payload.Token)
	m.ManagedAccountsCount = types.StringPointerValue(&payload.ManagedAccountsCount)
	m.Nickname = types.StringValue(payload.Nickname)
	m.Token = types.StringPointerValue(&payload.Token)
	m.UpdatedAt = types.StringPointerValue(&payload.UpdatedAt)

	return diags
}

func (m *billingProfileModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateBillingProfile {
	nickname := m.Nickname.ValueString()
	body := &modelsv2.CreateBillingProfile{
		Nickname: &nickname,
	}

	// Handle nested banking information attributes
	if !m.BankingInformationAttributes.IsNull() && !m.BankingInformationAttributes.IsUnknown() {
		bankingInfo := &modelsv2.CreateBillingProfileBankingInformationAttributes{}
		
		if !m.BankingInformationAttributes.BankName.IsNull() {
			bankingInfo.BankName = m.BankingInformationAttributes.BankName.ValueString()
		}
		if !m.BankingInformationAttributes.BeneficiaryName.IsNull() {
			bankingInfo.BeneficiaryName = m.BankingInformationAttributes.BeneficiaryName.ValueString()
		}
		if !m.BankingInformationAttributes.TaxId.IsNull() {
			bankingInfo.TaxID = m.BankingInformationAttributes.TaxId.ValueString()
		}
		
		// Handle secure data if present
		if !m.BankingInformationAttributes.SecureData.IsNull() && !m.BankingInformationAttributes.SecureData.IsUnknown() {
			secureDataMap := m.BankingInformationAttributes.SecureData.Attributes()
			secureData := &modelsv2.CreateBillingProfileBankingInformationAttributesSecureData{}
			
			if accountNumber, ok := secureDataMap["account_number"]; ok && !accountNumber.IsNull() {
				val := accountNumber.(types.String).ValueString()
				secureData.AccountNumber = val
			}
			if routingNumber, ok := secureDataMap["routing_number"]; ok && !routingNumber.IsNull() {
				val := routingNumber.(types.String).ValueString()
				secureData.RoutingNumber = val
			}
			if iban, ok := secureDataMap["iban"]; ok && !iban.IsNull() {
				val := iban.(types.String).ValueString()
				secureData.Iban = val
			}
			if swiftBic, ok := secureDataMap["swift_bic"]; ok && !swiftBic.IsNull() {
				val := swiftBic.(types.String).ValueString()
				secureData.SwiftBic = val
			}
			
			bankingInfo.SecureData = secureData
		}
		
		body.BankingInformationAttributes = bankingInfo
	}

	// Handle nested billing information attributes
	if !m.BillingInformationAttributes.IsNull() && !m.BillingInformationAttributes.IsUnknown() {
		billingInfo := &modelsv2.CreateBillingProfileBillingInformationAttributes{}
		
		if !m.BillingInformationAttributes.AddressLine1.IsNull() {
			billingInfo.AddressLine1 = m.BillingInformationAttributes.AddressLine1.ValueString()
		}
		if !m.BillingInformationAttributes.AddressLine2.IsNull() {
			billingInfo.AddressLine2 = m.BillingInformationAttributes.AddressLine2.ValueString()
		}
		if !m.BillingInformationAttributes.City.IsNull() {
			billingInfo.City = m.BillingInformationAttributes.City.ValueString()
		}
		if !m.BillingInformationAttributes.CompanyName.IsNull() {
			billingInfo.CompanyName = m.BillingInformationAttributes.CompanyName.ValueString()
		}
		if !m.BillingInformationAttributes.CountryCode.IsNull() {
			billingInfo.CountryCode = m.BillingInformationAttributes.CountryCode.ValueString()
		}
		if !m.BillingInformationAttributes.PostalCode.IsNull() {
			billingInfo.PostalCode = m.BillingInformationAttributes.PostalCode.ValueString()
		}
		if !m.BillingInformationAttributes.State.IsNull() {
			billingInfo.State = m.BillingInformationAttributes.State.ValueString()
		}
		if !m.BillingInformationAttributes.BillingEmail.IsNull() {
			emailTypes := []types.String{}
			if diag := m.BillingInformationAttributes.BillingEmail.ElementsAs(ctx, &emailTypes, false); !diag.HasError() {
				emails := []string{}
				for _, email := range emailTypes {
					emails = append(emails, email.ValueString())
				}
				billingInfo.BillingEmail = emails
			}
		}
		
		body.BillingInformationAttributes = billingInfo
	}

	// Handle nested business information attributes
	if !m.BusinessInformationAttributes.IsNull() && !m.BusinessInformationAttributes.IsUnknown() {
		businessInfo := &modelsv2.CreateBillingProfileBusinessInformationAttributes{}
		
		// Handle metadata if present
		if !m.BusinessInformationAttributes.Metadata.IsNull() && !m.BusinessInformationAttributes.Metadata.IsUnknown() {
			metadataMap := m.BusinessInformationAttributes.Metadata.Attributes()
			metadata := &modelsv2.CreateBillingProfileBusinessInformationAttributesMetadata{}
			
			if customFieldsAttr, ok := metadataMap["custom_fields"]; ok && !customFieldsAttr.IsNull() && !customFieldsAttr.IsUnknown() {
				customFieldsList := customFieldsAttr.(types.List)
				customFields := []*modelsv2.CreateBillingProfileBusinessInformationAttributesMetadataCustomFieldsItems0{}
				
				var tfCustomFields []resource_billing_profile.CustomFieldsValue
				if diag := customFieldsList.ElementsAs(ctx, &tfCustomFields, false); !diag.HasError() {
					for _, field := range tfCustomFields {
						customField := &modelsv2.CreateBillingProfileBusinessInformationAttributesMetadataCustomFieldsItems0{
							Name:  field.Name.ValueString(),
							Value: field.Value.ValueString(),
						}
						customFields = append(customFields, customField)
					}
					metadata.CustomFields = customFields
				}
			}
			
			businessInfo.Metadata = metadata
		}
		
		body.BusinessInformationAttributes = businessInfo
	}

	// Handle nested invoice adjustment attributes
	if !m.InvoiceAdjustmentAttributes.IsNull() && !m.InvoiceAdjustmentAttributes.IsUnknown() {
		invoiceAdjInfo := &modelsv2.CreateBillingProfileInvoiceAdjustmentAttributes{}

		if !m.InvoiceAdjustmentAttributes.AdjustmentItems.IsNull() && !m.InvoiceAdjustmentAttributes.AdjustmentItems.IsUnknown() {
			var tfAdjustmentItems []resource_billing_profile.AdjustmentItemsValue
			diags.Append(m.InvoiceAdjustmentAttributes.AdjustmentItems.ElementsAs(ctx, &tfAdjustmentItems, false)...)
			if diags.HasError() {
				return nil
			}

			adjustmentItems := []*modelsv2.CreateBillingProfileInvoiceAdjustmentAttributesAdjustmentItemsItems0{}
			for _, item := range tfAdjustmentItems {
				adjustmentType := item.AdjustmentType.ValueString()
				amount := item.Amount.ValueFloat64()
				calculationType := item.CalculationType.ValueString()
				name := item.Name.ValueString()

				adjustmentItem := &modelsv2.CreateBillingProfileInvoiceAdjustmentAttributesAdjustmentItemsItems0{
					AdjustmentType:  &adjustmentType,
					Amount:          &amount,
					CalculationType: &calculationType,
					Name:            &name,
				}
				adjustmentItems = append(adjustmentItems, adjustmentItem)
			}
			invoiceAdjInfo.AdjustmentItems = adjustmentItems
		}

		body.InvoiceAdjustmentAttributes = invoiceAdjInfo
	}

	return body
}

func (m *billingProfileModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateBillingProfile {
	body := &modelsv2.UpdateBillingProfile{
		Nickname: m.Nickname.ValueString(),
	}

	// Handle nested banking information attributes
	if !m.BankingInformationAttributes.IsNull() && !m.BankingInformationAttributes.IsUnknown() {
		bankingInfo := &modelsv2.UpdateBillingProfileBankingInformationAttributes{}
		
		if !m.BankingInformationAttributes.BankName.IsNull() {
			bankingInfo.BankName = m.BankingInformationAttributes.BankName.ValueString()
		}
		if !m.BankingInformationAttributes.BeneficiaryName.IsNull() {
			bankingInfo.BeneficiaryName = m.BankingInformationAttributes.BeneficiaryName.ValueString()
		}
		if !m.BankingInformationAttributes.TaxId.IsNull() {
			bankingInfo.TaxID = m.BankingInformationAttributes.TaxId.ValueString()
		}
		
		// Handle secure data if present
		if !m.BankingInformationAttributes.SecureData.IsNull() && !m.BankingInformationAttributes.SecureData.IsUnknown() {
			secureDataMap := m.BankingInformationAttributes.SecureData.Attributes()
			secureData := &modelsv2.UpdateBillingProfileBankingInformationAttributesSecureData{}
			
			if accountNumber, ok := secureDataMap["account_number"]; ok && !accountNumber.IsNull() {
				val := accountNumber.(types.String).ValueString()
				secureData.AccountNumber = val
			}
			if routingNumber, ok := secureDataMap["routing_number"]; ok && !routingNumber.IsNull() {
				val := routingNumber.(types.String).ValueString()
				secureData.RoutingNumber = val
			}
			if iban, ok := secureDataMap["iban"]; ok && !iban.IsNull() {
				val := iban.(types.String).ValueString()
				secureData.Iban = val
			}
			if swiftBic, ok := secureDataMap["swift_bic"]; ok && !swiftBic.IsNull() {
				val := swiftBic.(types.String).ValueString()
				secureData.SwiftBic = val
			}
			
			bankingInfo.SecureData = secureData
		}
		
		body.BankingInformationAttributes = bankingInfo
	}

	// Handle nested billing information attributes
	if !m.BillingInformationAttributes.IsNull() && !m.BillingInformationAttributes.IsUnknown() {
		billingInfo := &modelsv2.UpdateBillingProfileBillingInformationAttributes{}
		
		if !m.BillingInformationAttributes.AddressLine1.IsNull() {
			billingInfo.AddressLine1 = m.BillingInformationAttributes.AddressLine1.ValueString()
		}
		if !m.BillingInformationAttributes.AddressLine2.IsNull() {
			billingInfo.AddressLine2 = m.BillingInformationAttributes.AddressLine2.ValueString()
		}
		if !m.BillingInformationAttributes.City.IsNull() {
			billingInfo.City = m.BillingInformationAttributes.City.ValueString()
		}
		if !m.BillingInformationAttributes.CompanyName.IsNull() {
			billingInfo.CompanyName = m.BillingInformationAttributes.CompanyName.ValueString()
		}
		if !m.BillingInformationAttributes.CountryCode.IsNull() {
			billingInfo.CountryCode = m.BillingInformationAttributes.CountryCode.ValueString()
		}
		if !m.BillingInformationAttributes.PostalCode.IsNull() {
			billingInfo.PostalCode = m.BillingInformationAttributes.PostalCode.ValueString()
		}
		if !m.BillingInformationAttributes.State.IsNull() {
			billingInfo.State = m.BillingInformationAttributes.State.ValueString()
		}
		if !m.BillingInformationAttributes.BillingEmail.IsNull() {
			emailTypes := []types.String{}
			if diag := m.BillingInformationAttributes.BillingEmail.ElementsAs(ctx, &emailTypes, false); !diag.HasError() {
				emails := []string{}
				for _, email := range emailTypes {
					emails = append(emails, email.ValueString())
				}
				billingInfo.BillingEmail = emails
			}
		}
		
		body.BillingInformationAttributes = billingInfo
	}

	// Handle nested business information attributes
	if !m.BusinessInformationAttributes.IsNull() && !m.BusinessInformationAttributes.IsUnknown() {
		businessInfo := &modelsv2.UpdateBillingProfileBusinessInformationAttributes{}
		
		// Handle metadata if present
		if !m.BusinessInformationAttributes.Metadata.IsNull() && !m.BusinessInformationAttributes.Metadata.IsUnknown() {
			metadataMap := m.BusinessInformationAttributes.Metadata.Attributes()
			metadata := &modelsv2.UpdateBillingProfileBusinessInformationAttributesMetadata{}
			
			if customFieldsAttr, ok := metadataMap["custom_fields"]; ok && !customFieldsAttr.IsNull() && !customFieldsAttr.IsUnknown() {
				customFieldsList := customFieldsAttr.(types.List)
				customFields := []*modelsv2.UpdateBillingProfileBusinessInformationAttributesMetadataCustomFieldsItems0{}
				
				var tfCustomFields []resource_billing_profile.CustomFieldsValue
				if diag := customFieldsList.ElementsAs(ctx, &tfCustomFields, false); !diag.HasError() {
					for _, field := range tfCustomFields {
						customField := &modelsv2.UpdateBillingProfileBusinessInformationAttributesMetadataCustomFieldsItems0{
							Name:  field.Name.ValueString(),
							Value: field.Value.ValueString(),
						}
						customFields = append(customFields, customField)
					}
					metadata.CustomFields = customFields
				}
			}
			
			businessInfo.Metadata = metadata
		}
		
		body.BusinessInformationAttributes = businessInfo
	}

	// Handle nested invoice adjustment attributes
	if !m.InvoiceAdjustmentAttributes.IsNull() && !m.InvoiceAdjustmentAttributes.IsUnknown() {
		invoiceAdjInfo := &modelsv2.UpdateBillingProfileInvoiceAdjustmentAttributes{}

		if !m.InvoiceAdjustmentAttributes.AdjustmentItems.IsNull() && !m.InvoiceAdjustmentAttributes.AdjustmentItems.IsUnknown() {
			var tfAdjustmentItems []resource_billing_profile.AdjustmentItemsValue
			diags.Append(m.InvoiceAdjustmentAttributes.AdjustmentItems.ElementsAs(ctx, &tfAdjustmentItems, false)...)
			if diags.HasError() {
				return nil
			}

			adjustmentItems := []*modelsv2.UpdateBillingProfileInvoiceAdjustmentAttributesAdjustmentItemsItems0{}
			for _, item := range tfAdjustmentItems {
				adjustmentType := item.AdjustmentType.ValueString()
				amount := item.Amount.ValueFloat64()
				calculationType := item.CalculationType.ValueString()
				name := item.Name.ValueString()

				adjustmentItem := &modelsv2.UpdateBillingProfileInvoiceAdjustmentAttributesAdjustmentItemsItems0{
					AdjustmentType:  &adjustmentType,
					Amount:          &amount,
					CalculationType: &calculationType,
					Name:            &name,
				}
				adjustmentItems = append(adjustmentItems, adjustmentItem)
			}
			invoiceAdjInfo.AdjustmentItems = adjustmentItems
		}

		body.InvoiceAdjustmentAttributes = invoiceAdjInfo
	}

	return body
}
