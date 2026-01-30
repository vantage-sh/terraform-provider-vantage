package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_billing_profiles"
	billingprofilesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/billing_profiles"
)

var _ datasource.DataSource = (*billingProfilesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*billingProfilesDataSource)(nil)

func NewBillingProfilesDataSource() datasource.DataSource {
	return &billingProfilesDataSource{}
}

type billingProfilesDataSource struct {
	client *Client
}

func (d *billingProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_profiles"
}

func (d *billingProfilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_billing_profiles.BillingProfilesDataSourceSchema(ctx)
}

func (d *billingProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *Client, got something else.",
		)
		return
	}

	d.client = client
}

func (d *billingProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_billing_profiles.BillingProfilesModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to get billing profiles
	params := billingprofilesv2.NewGetBillingProfilesParams()
	result, err := d.client.V2.BillingProfiles.GetBillingProfiles(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Billing Profiles",
			err.Error(),
		)
		return
	}

	// Convert API response to Terraform model
	var billingProfilesList []attr.Value
	for _, bp := range result.Payload.BillingProfiles {
		
		// Handle Banking Information Attributes
		var bankingInfoAttr attr.Value
		if bp.BankingInformationAttributes != nil {
			// Handle secure data
			var secureDataAttr attr.Value
			if bp.BankingInformationAttributes.SecureData != nil {
				secureDataAttrs := map[string]attr.Value{
					"account_number": types.StringValue(bp.BankingInformationAttributes.SecureData.AccountNumber),
					"routing_number": types.StringValue(bp.BankingInformationAttributes.SecureData.RoutingNumber),
					"iban":          types.StringValue(bp.BankingInformationAttributes.SecureData.Iban),
					"swift_bic":     types.StringValue(bp.BankingInformationAttributes.SecureData.SwiftBic),
				}
				secureDataObj, diag := types.ObjectValue(map[string]attr.Type{
					"account_number": types.StringType,
					"routing_number": types.StringType,
					"iban":          types.StringType,
					"swift_bic":     types.StringType,
				}, secureDataAttrs)
				if diag.HasError() {
					resp.Diagnostics.Append(diag...)
					return
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
			
			bankingAttrs := map[string]attr.Value{
				"token":            types.StringValue(bp.BankingInformationAttributes.Token),
				"bank_name":        types.StringValue(bp.BankingInformationAttributes.BankName),
				"beneficiary_name": types.StringValue(bp.BankingInformationAttributes.BeneficiaryName),
				"tax_id":           types.StringValue(bp.BankingInformationAttributes.TaxID),
				"secure_data":      secureDataAttr,
			}
			bankingObj, diag := types.ObjectValue(map[string]attr.Type{
				"token":            types.StringType,
				"bank_name":        types.StringType,
				"beneficiary_name": types.StringType,
				"tax_id":           types.StringType,
				"secure_data":      types.ObjectType{AttrTypes: map[string]attr.Type{
					"account_number": types.StringType,
					"routing_number": types.StringType,
					"iban":          types.StringType,
					"swift_bic":     types.StringType,
				}},
			}, bankingAttrs)
			if diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}
			bankingInfoAttr = bankingObj
		} else {
			bankingInfoAttr = types.ObjectNull(map[string]attr.Type{
				"token":            types.StringType,
				"bank_name":        types.StringType,
				"beneficiary_name": types.StringType,
				"tax_id":           types.StringType,
				"secure_data":      types.ObjectType{AttrTypes: map[string]attr.Type{
					"account_number": types.StringType,
					"routing_number": types.StringType,
					"iban":          types.StringType,
					"swift_bic":     types.StringType,
				}},
			})
		}
		
		// Handle Billing Information Attributes
		var billingInfoAttr attr.Value
		if bp.BillingInformationAttributes != nil {
			billingEmails := []attr.Value{}
			if bp.BillingInformationAttributes.BillingEmail != nil {
				for _, email := range bp.BillingInformationAttributes.BillingEmail {
					billingEmails = append(billingEmails, types.StringValue(email))
				}
			}
			billingEmailsList, diag := types.ListValue(types.StringType, billingEmails)
			if diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}
			
			billingAttrs := map[string]attr.Value{
				"token":          types.StringValue(bp.BillingInformationAttributes.Token),
				"company_name":   types.StringValue(bp.BillingInformationAttributes.CompanyName),
				"country_code":   types.StringValue(bp.BillingInformationAttributes.CountryCode),
				"address_line_1": types.StringValue(bp.BillingInformationAttributes.AddressLine1),
				"address_line_2": types.StringValue(bp.BillingInformationAttributes.AddressLine2),
				"city":           types.StringValue(bp.BillingInformationAttributes.City),
				"state":          types.StringValue(bp.BillingInformationAttributes.State),
				"postal_code":    types.StringValue(bp.BillingInformationAttributes.PostalCode),
				"billing_email":  billingEmailsList,
			}
			billingObj, diag := types.ObjectValue(map[string]attr.Type{
				"token":          types.StringType,
				"company_name":   types.StringType,
				"country_code":   types.StringType,
				"address_line_1": types.StringType,
				"address_line_2": types.StringType,
				"city":           types.StringType,
				"state":          types.StringType,
				"postal_code":    types.StringType,
				"billing_email":  types.ListType{ElemType: types.StringType},
			}, billingAttrs)
			if diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}
			billingInfoAttr = billingObj
		} else {
			billingInfoAttr = types.ObjectNull(map[string]attr.Type{
				"token":          types.StringType,
				"company_name":   types.StringType,
				"country_code":   types.StringType,
				"address_line_1": types.StringType,
				"address_line_2": types.StringType,
				"city":           types.StringType,
				"state":          types.StringType,
				"postal_code":    types.StringType,
				"billing_email":  types.ListType{ElemType: types.StringType},
			})
		}
		
		// Handle Business Information Attributes  
		var businessInfoAttr attr.Value
		if bp.BusinessInformationAttributes != nil {
			var metadataAttr attr.Value
			if bp.BusinessInformationAttributes.Metadata != nil {
				customFieldsList := []attr.Value{}
				if bp.BusinessInformationAttributes.Metadata.CustomFields != nil {
					for _, field := range bp.BusinessInformationAttributes.Metadata.CustomFields {
						fieldValue, diag := datasource_billing_profiles.NewCustomFieldsValue(
							datasource_billing_profiles.CustomFieldsValue{}.AttributeTypes(ctx),
							map[string]attr.Value{
								"name":  types.StringValue(field.Name),
								"value": types.StringValue(field.Value),
							},
						)
						if diag.HasError() {
							resp.Diagnostics.Append(diag...)
							return
						}
						customFieldsList = append(customFieldsList, fieldValue)
					}
				}
				
				customFieldsListValue, diag := types.ListValue(
					datasource_billing_profiles.CustomFieldsType{
						ObjectType: types.ObjectType{
							AttrTypes: datasource_billing_profiles.CustomFieldsValue{}.AttributeTypes(ctx),
						},
					},
					customFieldsList,
				)
				if diag.HasError() {
					resp.Diagnostics.Append(diag...)
					return
				}
				
				metadataObj, diag := types.ObjectValue(
					map[string]attr.Type{
						"custom_fields": types.ListType{ElemType: datasource_billing_profiles.CustomFieldsType{
							ObjectType: types.ObjectType{
								AttrTypes: datasource_billing_profiles.CustomFieldsValue{}.AttributeTypes(ctx),
							},
						}},
					},
					map[string]attr.Value{
						"custom_fields": customFieldsListValue,
					},
				)
				if diag.HasError() {
					resp.Diagnostics.Append(diag...)
					return
				}
				metadataAttr = metadataObj
			} else {
				metadataAttr = types.ObjectNull(map[string]attr.Type{
					"custom_fields": types.ListType{ElemType: datasource_billing_profiles.CustomFieldsType{
						ObjectType: types.ObjectType{
							AttrTypes: datasource_billing_profiles.CustomFieldsValue{}.AttributeTypes(ctx),
						},
					}},
				})
			}
			
			businessAttrs := map[string]attr.Value{
				"token":    types.StringValue(bp.BusinessInformationAttributes.Token),
				"metadata": metadataAttr,
			}
			businessObj, diag := types.ObjectValue(map[string]attr.Type{
				"token":    types.StringType,
				"metadata": types.ObjectType{AttrTypes: map[string]attr.Type{
					"custom_fields": types.ListType{ElemType: datasource_billing_profiles.CustomFieldsType{
						ObjectType: types.ObjectType{
							AttrTypes: datasource_billing_profiles.CustomFieldsValue{}.AttributeTypes(ctx),
						},
					}},
				}},
			}, businessAttrs)
			if diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}
			businessInfoAttr = businessObj
		} else {
			businessInfoAttr = types.ObjectNull(map[string]attr.Type{
				"token":    types.StringType,
				"metadata": types.ObjectType{AttrTypes: map[string]attr.Type{
					"custom_fields": types.ListType{ElemType: datasource_billing_profiles.CustomFieldsType{
						ObjectType: types.ObjectType{
							AttrTypes: datasource_billing_profiles.CustomFieldsValue{}.AttributeTypes(ctx),
						},
					}},
				}},
			})
		}

		// Handle Invoice Adjustment Attributes
		// Note: In data source schema, adjustment_items is a SingleNestedAttribute (object), not a list
		adjustmentItemsAttrTypes := map[string]attr.Type{
			"adjustment_type":  types.StringType,
			"amount":           types.StringType,
			"calculation_type": types.StringType,
			"name":             types.StringType,
		}
		invoiceAdjAttrTypes := map[string]attr.Type{
			"adjustment_items": types.ObjectType{AttrTypes: adjustmentItemsAttrTypes},
			"token":            types.StringType,
		}

		var invoiceAdjAttr attr.Value
		if bp.InvoiceAdjustmentAttributes != nil {
			// Use the first adjustment item if available (data source schema only supports single object)
			var adjustmentItemsAttr attr.Value
			if bp.InvoiceAdjustmentAttributes.AdjustmentItems != nil && len(bp.InvoiceAdjustmentAttributes.AdjustmentItems) > 0 {
				item := bp.InvoiceAdjustmentAttributes.AdjustmentItems[0]
				itemAttrs := map[string]attr.Value{
					"adjustment_type":  types.StringValue(item.AdjustmentType),
					"amount":           types.StringValue(item.Amount),
					"calculation_type": types.StringValue(item.CalculationType),
					"name":             types.StringValue(item.Name),
				}
				itemObj, diag := types.ObjectValue(adjustmentItemsAttrTypes, itemAttrs)
				if diag.HasError() {
					resp.Diagnostics.Append(diag...)
					return
				}
				adjustmentItemsAttr = itemObj
			} else {
				adjustmentItemsAttr = types.ObjectNull(adjustmentItemsAttrTypes)
			}

			invoiceAdjAttrs := map[string]attr.Value{
				"adjustment_items": adjustmentItemsAttr,
				"token":            types.StringValue(bp.InvoiceAdjustmentAttributes.Token),
			}
			invoiceAdjObj, diag := types.ObjectValue(invoiceAdjAttrTypes, invoiceAdjAttrs)
			if diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}
			invoiceAdjAttr = invoiceAdjObj
		} else {
			invoiceAdjAttr = types.ObjectNull(invoiceAdjAttrTypes)
		}

		// Create a billing profile value using the generated type
		bpValue, diag := datasource_billing_profiles.NewBillingProfilesValue(
			datasource_billing_profiles.BillingProfilesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"id":                               types.StringPointerValue(&bp.Token),
				"token":                            types.StringPointerValue(&bp.Token),
				"nickname":                         types.StringValue(bp.Nickname),
				"created_at":                       types.StringPointerValue(&bp.CreatedAt),
				"updated_at":                       types.StringPointerValue(&bp.UpdatedAt),
				"managed_accounts_count":           types.StringPointerValue(&bp.ManagedAccountsCount),
				"banking_information_attributes":   bankingInfoAttr,
				"billing_information_attributes":   billingInfoAttr,
				"business_information_attributes":  businessInfoAttr,
				"invoice_adjustment_attributes":    invoiceAdjAttr,
			},
		)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		billingProfilesList = append(billingProfilesList, bpValue)
	}
	
	// Create the list of billing profiles
	billingProfilesListValue, diag := types.ListValue(
		datasource_billing_profiles.BillingProfilesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_billing_profiles.BillingProfilesValue{}.AttributeTypes(ctx),
			},
		},
		billingProfilesList,
	)
	
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.BillingProfiles = billingProfilesListValue

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
