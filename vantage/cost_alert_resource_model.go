package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_cost_alerts"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_cost_alert"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type costAlertModel resource_cost_alert.CostAlertModel
type costAlertDataSourceValue datasource_cost_alerts.CostAlertsValue

func (m *costAlertModel) applyPayload(ctx context.Context, payload *modelsv2.CostAlert) diag.Diagnostics {
	var diag diag.Diagnostics
	var list types.List

	m.Token = types.StringValue(payload.Token)
	m.Title = types.StringValue(payload.Title)
	m.Interval = types.StringValue(payload.Interval)
	m.UnitType = types.StringValue(payload.UnitType)
	m.Threshold = types.Float64Value(payload.Threshold)
	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.UpdatedAt = types.StringValue(payload.UpdatedAt)

	list, diag = types.ListValueFrom(ctx, types.StringType, payload.EmailRecipients)
	if diag.HasError() {
		return diag
	}
	m.EmailRecipients = list

	list, diag = types.ListValueFrom(ctx, types.StringType, payload.SlackChannels)
	if diag.HasError() {
		return diag
	}
	m.SlackChannels = list

	list, diag = types.ListValueFrom(ctx, types.StringType, payload.TeamsChannels)
	if diag.HasError() {
		return diag
	}
	m.TeamsChannels = list

	list, diag = types.ListValueFrom(ctx, types.StringType, payload.ReportTokens)
	if diag.HasError() {
		return diag
	}
	m.ReportTokens = list

	return diag
}

func (m *costAlertModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateCostAlert {
	var threshold *float32
	if !m.Threshold.IsNull() && !m.Threshold.IsUnknown() {
		parsedFloat := float32(m.Threshold.ValueFloat64())
		threshold = &parsedFloat
	}

	return &modelsv2.CreateCostAlert{
		WorkspaceToken:  m.WorkspaceToken.ValueStringPointer(),
		Title:           m.Title.ValueStringPointer(),
		Interval:        m.Interval.ValueStringPointer(),
		Threshold:       threshold,
		UnitType:        m.UnitType.ValueStringPointer(),
		EmailRecipients: terraformListToStrings(ctx, m.EmailRecipients, diags),
		SlackChannels:   terraformListToStrings(ctx, m.SlackChannels, diags),
		TeamsChannels:   terraformListToStrings(ctx, m.TeamsChannels, diags),
		ReportTokens:    terraformListToStrings(ctx, m.ReportTokens, diags),
	}
}

func (m *costAlertModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateCostAlert {
	return &modelsv2.UpdateCostAlert{
		Title:           m.Title.ValueString(),
		Interval:        m.Interval.ValueString(),
		Threshold:       float32(m.Threshold.ValueFloat64()),
		UnitType:        m.UnitType.ValueString(),
		EmailRecipients: terraformListToStrings(ctx, m.EmailRecipients, diags),
		SlackChannels:   terraformListToStrings(ctx, m.SlackChannels, diags),
		TeamsChannels:   terraformListToStrings(ctx, m.TeamsChannels, diags),
		ReportTokens:    terraformListToStrings(ctx, m.ReportTokens, diags),
	}
}

func terraformListToStrings(ctx context.Context, tfList types.List, diags *diag.Diagnostics) []string {
	if tfList.IsNull() || tfList.IsUnknown() {
		return nil
	}
	var out []string
	diagnostics := tfList.ElementsAs(ctx, &out, false)
	diags.Append(diagnostics...)
	return out
}
