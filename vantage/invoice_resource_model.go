package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_invoice"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type invoiceModel resource_invoice.InvoiceModel

func (m *invoiceModel) applyPayload(ctx context.Context, payload *modelsv2.Invoice) diag.Diagnostics {
	if payload.AccountName != "" {
		m.AccountName = types.StringValue(payload.AccountName)
	} else {
		m.AccountName = types.StringNull()
	}

	m.AccountToken = types.StringValue(payload.AccountToken)
	m.BillingPeriodEnd = types.StringValue(payload.BillingPeriodEnd)
	m.BillingPeriodStart = types.StringValue(payload.BillingPeriodStart)

	if payload.CreatedAt != "" {
		m.CreatedAt = types.StringValue(payload.CreatedAt)
	} else {
		m.CreatedAt = types.StringNull()
	}

	if payload.InvoiceNumber != "" {
		m.InvoiceNumber = types.StringValue(payload.InvoiceNumber)
	} else {
		m.InvoiceNumber = types.StringNull()
	}

	if payload.MspAccountToken != "" {
		m.MspAccountToken = types.StringValue(payload.MspAccountToken)
	} else {
		m.MspAccountToken = types.StringNull()
	}

	if payload.Status != "" {
		m.Status = types.StringValue(payload.Status)
	} else {
		m.Status = types.StringNull()
	}

	if payload.Token != "" {
		m.Token = types.StringValue(payload.Token)
	} else {
		m.Token = types.StringNull()
	}

	if payload.Total != "" {
		m.Total = types.StringValue(payload.Total)
	} else {
		m.Total = types.StringNull()
	}

	if payload.UpdatedAt != "" {
		m.UpdatedAt = types.StringValue(payload.UpdatedAt)
	} else {
		m.UpdatedAt = types.StringNull()
	}

	return nil
}

func (m *invoiceModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateInvoice {
	accountToken := m.AccountToken.ValueString()
	billingPeriodEnd := m.BillingPeriodEnd.ValueString()
	billingPeriodStart := m.BillingPeriodStart.ValueString()
	
	return &modelsv2.CreateInvoice{
		AccountToken:        &accountToken,
		BillingPeriodEnd:    &billingPeriodEnd,
		BillingPeriodStart:  &billingPeriodStart,
	}
}
