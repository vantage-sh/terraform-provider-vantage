package vantage

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_cost_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type costReportModel resource_cost_report.CostReportModel

func (m *costReportModel) applyPayload(ctx context.Context, payload *modelsv2.CostReport) diag.Diagnostics {
	m.Title = types.StringValue(payload.Title)
	m.FolderToken = types.StringValue(payload.FolderToken)
	m.Filter = types.StringValue(payload.Filter)
	m.Groupings = types.StringValue(payload.Groupings)

	saved_filters, diag := types.ListValueFrom(ctx, types.StringType, payload.SavedFilterTokens)
	if diag.HasError() {
		return diag
	}
	m.SavedFilterTokens = saved_filters

	m.WorkspaceToken = types.StringValue(payload.WorkspaceToken)
	m.PreviousPeriodStartDate = types.StringValue(payload.PreviousPeriodStartDate)
	m.PreviousPeriodEndDate = types.StringValue(payload.PreviousPeriodEndDate)
	m.ChartType = types.StringValue(payload.ChartType)
	m.StartDate = types.StringValue(payload.StartDate)
	m.EndDate = types.StringValue(payload.EndDate)

	tflog.Debug(ctx, fmt.Sprintf("DateInterval is not null %s", payload.DateInterval))
	if payload.DateInterval != "" {
		m.DateInterval = types.StringValue(payload.DateInterval)
	} else {
		m.DateInterval = types.StringNull()
	}

	if payload.DateInterval == "custom" {
		m.StartDate = types.StringValue(payload.StartDate)
		m.EndDate = types.StringValue(payload.EndDate)
	}

	m.DateBin = types.StringValue(payload.DateBin)

	m.Token = types.StringValue(payload.Token)

	return nil
}

func (m *costReportModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateCostReport {
	savedFilterTokens := []types.String{}
	if !m.SavedFilterTokens.IsNull() && !m.SavedFilterTokens.IsUnknown() {
		savedFilterTokens = make([]types.String, 0, len(m.SavedFilterTokens.Elements()))
		if diag := m.SavedFilterTokens.ElementsAs(ctx, &savedFilterTokens, false); diag.HasError() {
			diags.Append(diag...)
			return nil
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("DateInterval create %s", m.DateInterval.ValueString()))
	payload := &modelsv2.CreateCostReport{
		Title:               m.Title.ValueStringPointer(),
		FolderToken:         m.FolderToken.ValueString(),
		Filter:              m.Filter.ValueString(),
		Groupings:           m.Groupings.ValueString(),
		SavedFilterTokens:   fromStringsValue(savedFilterTokens),
		WorkspaceToken:      m.WorkspaceToken.ValueString(),
		DateInterval:        m.DateInterval.ValueString(),
		DateBin:             m.DateBin.ValueStringPointer(),
	}

	if m.DateInterval.ValueString() == "custom" {
		payload.StartDate = m.StartDate.ValueString()
	    payload.EndDate = m.EndDate.ValueStringPointer()
		payload.DateInterval = "custom"
	} else {
		payload.DateInterval = m.DateInterval.ValueString()
	}

	return payload
}

func (m *costReportModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateCostReport {
	savedFilterTokens := []types.String{}
	if !m.SavedFilterTokens.IsNull() && !m.SavedFilterTokens.IsUnknown() {
		savedFilterTokens = make([]types.String, 0, len(m.SavedFilterTokens.Elements()))
		if diag := m.SavedFilterTokens.ElementsAs(ctx, &savedFilterTokens, false); diag.HasError() {
			diags.Append(diag...)
			return nil
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("DateInterval update %s", m.DateInterval.ValueString()))
	payload := &modelsv2.UpdateCostReport{
		Title:               m.Title.ValueString(),
		FolderToken:         m.FolderToken.ValueString(),
		Filter:              m.Filter.ValueString(),
		Groupings:           m.Groupings.ValueString(),
		StartDate:           m.StartDate.ValueString(),
		EndDate:             m.EndDate.ValueString(),
		PreviousPeriodStartDate: m.PreviousPeriodStartDate.ValueString(),
		PreviousPeriodEndDate:   m.PreviousPeriodEndDate.ValueString(),
		DateInterval:        m.DateInterval.ValueString(),
		ChartType:           m.ChartType.ValueStringPointer(),
		DateBin:             m.DateBin.ValueStringPointer(),
		SavedFilterTokens:   fromStringsValue(savedFilterTokens),
	}

	return payload
}
