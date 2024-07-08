package vantage

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_billing_rule"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type datasourceBillingRuleModel struct {
	AdjustedRate   types.String `tfsdk:"adjusted_rate"`
	Amount         types.String `tfsdk:"amount"`
	Category       types.String `tfsdk:"category"`
	ChargeType     types.String `tfsdk:"charge_type"`
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByToken types.String `tfsdk:"created_by_token"`
	Service        types.String `tfsdk:"service"`
	StartPeriod    types.String `tfsdk:"start_period"`
	SubCategory    types.String `tfsdk:"sub_category"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	Type           types.String `tfsdk:"type"`
}

type billingRuleModel resource_billing_rule.BillingRuleModel

func (m *billingRuleModel) toDatasourceModel() datasourceBillingRuleModel {
	adjustedRate := strconv.FormatFloat(m.AdjustedRate.ValueFloat64(), 'g', -1, 64)
	amount := strconv.FormatFloat(m.Amount.ValueFloat64(), 'g', -1, 64)

	return datasourceBillingRuleModel{
		AdjustedRate:   types.StringValue(adjustedRate),
		Amount:         types.StringValue(amount),
		Category:       m.Category,
		ChargeType:     m.ChargeType,
		CreatedAt:      m.CreatedAt,
		CreatedByToken: m.CreatedByToken,
		Service:        m.Service,
		StartPeriod:    m.StartPeriod,
		SubCategory:    m.SubCategory,
		Title:          m.Title,
		Token:          m.Token,
		Type:           m.Type,
	}
}

func (m *billingRuleModel) applyPayload(_ context.Context, payload *modelsv2.BillingRule) diag.Diagnostics {
	m.Token = types.StringValue(payload.Token)
	m.Title = types.StringValue(payload.Title)
	if payload.AdjustedRate != "" {
		rate, err := strconv.ParseFloat(payload.AdjustedRate, 64)
		if err != nil {
			d := diag.Diagnostics{}
			d.AddError("error converting rate to int", err.Error())
			return d
		}

		m.AdjustedRate = types.Float64Value(rate)
	}

	if payload.Amount != "" {
		amount, err := strconv.ParseFloat(payload.Amount, 64)
		if err != nil {
			d := diag.Diagnostics{}
			d.AddError("error converting rate to int", err.Error())
			return d
		}
		m.Amount = types.Float64Value(amount)
	}

	if payload.Category != "" {
		m.Category = types.StringValue(payload.Category)
	}
	if payload.ChargeType != "" {
		m.ChargeType = types.StringValue(payload.ChargeType)
	}

	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.CreatedByToken = types.StringValue(payload.CreatedByToken)

	if payload.Service != "" {
		m.Service = types.StringValue(payload.Service)
	}

	if payload.StartPeriod != "" {
		m.StartPeriod = types.StringValue(payload.StartPeriod)
	}

	if payload.SubCategory != "" {
		m.SubCategory = types.StringValue(payload.SubCategory)
	}
	m.Title = types.StringValue(payload.Title)
	m.Token = types.StringValue(payload.Token)
	m.Type = types.StringValue(payload.Type)

	return nil
}

func (m *billingRuleModel) toCreateModel(_ context.Context, _ *diag.Diagnostics) *modelsv2.CreateBillingRule {
	return &modelsv2.CreateBillingRule{
		AdjustedRate: m.AdjustedRate.ValueFloat64Pointer(),
		Amount:       m.Amount.ValueFloat64Pointer(),
		Category:     m.Category.ValueStringPointer(),
		ChargeType:   m.ChargeType.ValueStringPointer(),
		Service:      m.Service.ValueStringPointer(),
		StartPeriod:  m.StartPeriod.ValueStringPointer(),
		SubCategory:  m.SubCategory.ValueStringPointer(),
		Title:        m.Title.ValueStringPointer(),
		Type:         m.Type.ValueStringPointer(),
	}
}

func (m *billingRuleModel) toUpdateModel(_ context.Context, _ *diag.Diagnostics) *modelsv2.UpdateBillingRule {

	return &modelsv2.UpdateBillingRule{
		AdjustedRate: m.AdjustedRate.ValueFloat64(),
		Amount:       m.Amount.ValueFloat64(),
		Category:     m.Category.ValueString(),
		ChargeType:   m.ChargeType.ValueString(),
		Service:      m.Service.ValueString(),
		StartPeriod:  m.StartPeriod.ValueString(),
		SubCategory:  m.SubCategory.ValueString(),
		Title:        m.Title.ValueString(),
	}
}
