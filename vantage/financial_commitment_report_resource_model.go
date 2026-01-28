package vantage

import (
	"context"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_financial_commitment_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type financialCommitmentReportModel resource_financial_commitment_report.FinancialCommitmentReportModel

func (m *financialCommitmentReportModel) applyPayload(ctx context.Context, payload *modelsv2.FinancialCommitmentReport) diag.Diagnostics {
	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.UserToken = types.StringValue(payload.UserToken)
	m.Filter = types.StringValue(payload.Filter)
	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.Title = types.StringValue(payload.Title)
	m.WorkspaceToken = types.StringValue(payload.WorkspaceToken)
	m.DateBucket = types.StringValue(payload.DateBucket)
	m.DateInterval = types.StringValue(payload.DateInterval)
	m.Default = types.BoolValue(payload.Default)
	m.OnDemandCostsScope = types.StringValue(payload.OnDemandCostsScope)
	m.StartDate = types.StringValue(payload.StartDate)
	m.EndDate = types.StringValue(payload.EndDate)

	// Handle groupings - strings.Split on empty string returns [""], not []
	// so we need to handle that case explicitly
	var groupings []string
	if payload.Groupings != "" {
		groupings = strings.Split(payload.Groupings, ",")
	} else {
		groupings = []string{}
	}

	var d diag.Diagnostics
	m.Groupings, d = types.ListValueFrom(ctx, types.StringType, groupings)

	if d.HasError() {
		return d
	}

	return nil
}

func (m *financialCommitmentReportModel) toCreateModel(ctx context.Context) *modelsv2.CreateFinancialCommitmentReport {

	var groupings []string

	if !m.Groupings.IsNull() && !m.Groupings.IsUnknown() {
		m.Groupings.ElementsAs(ctx, &groupings, false)
	}

	m.Groupings.Elements()
	dst := &modelsv2.CreateFinancialCommitmentReport{
		Title:              m.Title.ValueStringPointer(),
		Filter:             m.Filter.ValueString(),
		WorkspaceToken:     m.WorkspaceToken.ValueStringPointer(),
		DateBucket:         m.DateBucket.ValueString(),
		DateInterval:       m.DateInterval.ValueString(),
		Groupings:          groupings,
		OnDemandCostsScope: m.OnDemandCostsScope.ValueString(),
	}

	if m.DateInterval.ValueString() == "custom" {
		endDate := strfmt.Date{}
		if m.EndDate.ValueString() != "" {
			endDate.UnmarshalText([]byte(m.EndDate.ValueString()))
		}
		startDate := strfmt.Date{}
		if m.StartDate.ValueString() != "" {
			startDate.UnmarshalText([]byte(m.StartDate.ValueString()))
		}
		dst.EndDate = endDate
		dst.StartDate = startDate
	}
	return dst
}

func (m *financialCommitmentReportModel) toUpdateModel(ctx context.Context) *modelsv2.UpdateFinancialCommitmentReport {

	var groupings []string

	if !m.Groupings.IsNull() && !m.Groupings.IsUnknown() {
		m.Groupings.ElementsAs(ctx, &groupings, false)
	}

	m.Groupings.Elements()
	dst := &modelsv2.UpdateFinancialCommitmentReport{
		Title:              m.Title.ValueString(),
		Filter:             m.Filter.ValueString(),
		DateBucket:         m.DateBucket.ValueString(),
		DateInterval:       m.DateInterval.ValueString(),
		Groupings:          groupings,
		OnDemandCostsScope: m.OnDemandCostsScope.ValueString(),
	}

	if m.DateInterval.ValueString() == "custom" {
		endDate := strfmt.Date{}
		if m.EndDate.ValueString() != "" {
			endDate.UnmarshalText([]byte(m.EndDate.ValueString()))
		}
		startDate := strfmt.Date{}
		if m.StartDate.ValueString() != "" {
			startDate.UnmarshalText([]byte(m.StartDate.ValueString()))
		}
		dst.EndDate = endDate
		dst.StartDate = startDate
	}
	return dst
}
