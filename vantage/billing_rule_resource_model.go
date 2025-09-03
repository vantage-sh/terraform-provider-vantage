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
	Percentage     types.String `tfsdk:"percentage"`
	Amount         types.String `tfsdk:"amount"`
	ApplyToAll     types.Bool   `tfsdk:"apply_to_all"`
	EndDate        types.String `tfsdk:"end_date"`
	StartDate      types.String `tfsdk:"start_date"`
	Category       types.String `tfsdk:"category"`
	ChargeType     types.String `tfsdk:"charge_type"`
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByToken types.String `tfsdk:"created_by_token"`
	Service        types.String `tfsdk:"service"`
	StartPeriod    types.String `tfsdk:"start_period"`
	SubCategory    types.String `tfsdk:"sub_category"`
	SqlQuery       types.String `tfsdk:"sql_query"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	Type           types.String `tfsdk:"type"`
}

type billingRuleModel resource_billing_rule.BillingRuleModel

func (m *billingRuleModel) toDatasourceModel() datasourceBillingRuleModel {
	percentage := strconv.FormatFloat(m.Percentage.ValueFloat64(), 'g', -1, 64)
	amount := strconv.FormatFloat(m.Amount.ValueFloat64(), 'g', -1, 64)

	return datasourceBillingRuleModel{
		Percentage:     types.StringValue(percentage),
		Amount:         types.StringValue(amount),
		ApplyToAll:     m.ApplyToAll,
		EndDate:        m.EndDate,
		StartDate:      m.StartDate,
		Category:       m.Category,
		ChargeType:     m.ChargeType,
		CreatedAt:      m.CreatedAt,
		CreatedByToken: m.CreatedByToken,
		Service:        m.Service,
		StartPeriod:    m.StartPeriod,
		SubCategory:    m.SubCategory,
		SqlQuery:       m.SqlQuery,
		Title:          m.Title,
		Token:          m.Token,
		Type:           m.Type,
	}
}

func (m *billingRuleModel) applyPayload(ctx context.Context, payload *modelsv2.BillingRule) diag.Diagnostics {

	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.Title = types.StringValue(payload.Title)
	if payload.Percentage != "" {
		rate, err := strconv.ParseFloat(payload.Percentage, 64)
		if err != nil {
			d := diag.Diagnostics{}
			d.AddError("error converting rate to int", err.Error())
			return d
		}

		m.Percentage = types.Float64Value(rate)
	} else {
		m.Percentage = types.Float64Value(0.0)
	}

	if payload.Amount != "" {
		amount, err := strconv.ParseFloat(payload.Amount, 64)
		if err != nil {
			d := diag.Diagnostics{}
			d.AddError("error converting rate to int", err.Error())
			return d
		}
		m.Amount = types.Float64Value(amount)
	} else {
		m.Amount = types.Float64Value(0.0)
	}

	m.ApplyToAll = types.BoolValue(payload.ApplyToAll)

	if payload.EndDate != "" {
		m.EndDate = types.StringValue(payload.EndDate)
	} else {
		m.EndDate = types.StringNull()
	}

	if payload.StartDate != "" {
		m.StartDate = types.StringValue(payload.StartDate)
	} else {
		m.StartDate = types.StringNull()
	}

	if payload.Category != "" {
		m.Category = types.StringValue(payload.Category)
	} else {
		m.Category = types.StringNull()
	}

	if payload.ChargeType != "" {
		m.ChargeType = types.StringValue(payload.ChargeType)
	} else {
		m.ChargeType = types.StringNull()
	}

	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.CreatedByToken = types.StringValue(payload.CreatedByToken)

	if payload.Service != "" {
		m.Service = types.StringValue(payload.Service)
	} else {
		m.Service = types.StringNull()
	}

	if payload.StartPeriod != "" {
		m.StartPeriod = types.StringValue(payload.StartPeriod)
	} else {
		m.StartPeriod = types.StringNull()
	}

	if payload.SubCategory != "" {
		m.SubCategory = types.StringValue(payload.SubCategory)
	} else {
		m.SubCategory = types.StringNull()
	}

	if payload.SQLQuery != "" {
		m.SqlQuery = types.StringValue(payload.SQLQuery)
	} else {
		m.SqlQuery = types.StringNull()
	}

	m.Title = types.StringValue(payload.Title)
	m.Token = types.StringValue(payload.Token)
	m.Type = types.StringValue(payload.Type)

	return nil
}

func (m *billingRuleModel) toCreateModel(_ context.Context, _ *diag.Diagnostics) *modelsv2.CreateBillingRule {
	return &modelsv2.CreateBillingRule{
		Percentage:  m.Percentage.ValueFloat64(),
		Amount:      m.Amount.ValueFloat64(),
		ApplyToAll:  m.ApplyToAll.ValueBool(),
		StartDate:   m.StartDate.ValueString(),
		EndDate:     m.EndDate.ValueString(),
		Category:    m.Category.ValueString(),
		ChargeType:  m.ChargeType.ValueString(),
		Service:     m.Service.ValueString(),
		StartPeriod: m.StartPeriod.ValueString(),
		SubCategory: m.SubCategory.ValueString(),
		SQLQuery:    m.SqlQuery.ValueString(),
		Title:       m.Title.ValueStringPointer(),
		Type:        m.Type.ValueStringPointer(),
	}
}

func (m *billingRuleModel) toUpdateModel(_ context.Context, _ *diag.Diagnostics) *modelsv2.UpdateBillingRule {

	return &modelsv2.UpdateBillingRule{
		Percentage:  m.Percentage.ValueFloat64(),
		Amount:      m.Amount.ValueFloat64(),
		ApplyToAll:  m.ApplyToAll.ValueBool(),
		StartDate:   m.StartDate.ValueString(),
		EndDate:     m.EndDate.ValueString(),
		Category:    m.Category.ValueString(),
		ChargeType:  m.ChargeType.ValueString(),
		Service:     m.Service.ValueString(),
		StartPeriod: m.StartPeriod.ValueString(),
		SubCategory: m.SubCategory.ValueString(),
		SQLQuery:    m.SqlQuery.ValueString(),
		Title:       m.Title.ValueString(),
	}
}
