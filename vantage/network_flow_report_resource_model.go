package vantage

import (
	"context"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_network_flow_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type networkFlowReportResourceModel resource_network_flow_report.NetworkFlowReportModel

func (m *networkFlowReportResourceModel) applyPayload(ctx context.Context, payload *modelsv2.NetworkFlowReport) diag.Diagnostics {

	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.CreatedByToken = types.StringValue(payload.CreatedByToken)
	m.DateInterval = types.StringValue(payload.DateInterval)
	m.Default = types.BoolValue(payload.Default)
	m.EndDate = types.StringValue(payload.EndDate)
	m.Filter = types.StringValue(payload.Filter)
	m.FlowDirection = types.StringValue(payload.FlowDirection)
	m.FlowWeight = types.StringValue(payload.FlowWeight)
	m.StartDate = types.StringValue(payload.StartDate)
	m.Title = types.StringValue(payload.Title)
	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.WorkspaceToken = types.StringValue(payload.WorkspaceToken)

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

func (m *networkFlowReportResourceModel) toCreateModel(ctx context.Context) *modelsv2.CreateNetworkFlowReport {
	var groupings []string

	if !m.Groupings.IsNull() && !m.Groupings.IsUnknown() {
		m.Groupings.ElementsAs(ctx, &groupings, false)
	}

	m.Groupings.Elements()
	dst := &modelsv2.CreateNetworkFlowReport{
		Title:          m.Title.ValueStringPointer(),
		Filter:         m.Filter.ValueString(),
		FlowDirection:  m.FlowDirection.ValueString(),
		FlowWeight:     m.FlowWeight.ValueString(),
		WorkspaceToken: m.WorkspaceToken.ValueStringPointer(),
		DateInterval:   m.DateInterval.ValueString(),
		Groupings:      groupings,
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

func (m *networkFlowReportResourceModel) toUpdateModel(ctx context.Context) *modelsv2.UpdateNetworkFlowReport {
	var groupings []string

	if !m.Groupings.IsNull() && !m.Groupings.IsUnknown() {
		m.Groupings.ElementsAs(ctx, &groupings, false)
	}

	m.Groupings.Elements()
	dst := &modelsv2.UpdateNetworkFlowReport{
		Title:         m.Title.ValueString(),
		Filter:        m.Filter.ValueString(),
		FlowDirection: m.FlowDirection.ValueString(),
		FlowWeight:    m.FlowWeight.ValueString(),
		DateInterval:  m.DateInterval.ValueString(),
		Groupings:     groupings,
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
