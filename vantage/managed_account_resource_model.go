package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_managed_account"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type managedAccountModel resource_managed_account.ManagedAccountModel

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

	return nil
}

func (m *managedAccountModel) toCreateModel(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateManagedAccount {
	dst := &modelsv2.CreateManagedAccount{
		Name:         m.Name.ValueStringPointer(),
		ContactEmail: m.ContactEmail.ValueStringPointer(),
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
	}

	if !m.BillingRuleTokens.IsNull() && !m.BillingRuleTokens.IsUnknown() {
		billingRuleTokens := []string{}
		m.BillingRuleTokens.ElementsAs(ctx, &billingRuleTokens, false)
		dst.BillingRuleTokens = billingRuleTokens
	}

	return dst
}
