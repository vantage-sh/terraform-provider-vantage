package vantage

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_budget_alert"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

type budgetAlertModel resource_budget_alert.BudgetAlertModel

// durationInDaysToState maps the nullable integer the API returns onto the
// string the create/update payloads accept. The API returns null for alerts
// that span the full month, which requests express as an empty value.
func durationInDaysToState(src *int32) types.String {
	if src == nil {
		return types.StringValue("")
	}
	return types.StringValue(strconv.FormatInt(int64(*src), 10))
}

// stringListOrEmpty converts a Terraform list into a slice, defaulting to an
// empty slice so the API receives an array rather than null.
func stringListOrEmpty(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	values := terraformListToStrings(ctx, list, diags)
	if values == nil {
		return []string{}
	}
	return values
}

// stringListOrNil distinguishes an unset recipient list from an empty one. A
// list that is absent from config becomes nil, which reaches the API as null and
// leaves the field untouched. An explicitly empty list must survive as an empty
// array, which the API applies as a clear.
func stringListOrNil(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	values := terraformListToStrings(ctx, list, diags)
	if values == nil {
		return []string{}
	}
	return values
}

func (m *budgetAlertModel) applyPayload(ctx context.Context, payload *modelsv2.BudgetAlert) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.DurationInDays = durationInDaysToState(payload.DurationInDays)
	m.IntegrationProvider = types.StringPointerValue(payload.IntegrationProvider)
	m.PeriodToTrack = types.StringPointerValue(payload.PeriodToTrack)
	m.Threshold = types.Int64Value(int64(payload.Threshold))
	m.UserToken = types.StringPointerValue(payload.UserToken)
	m.WorkspaceToken = types.StringPointerValue(payload.WorkspaceToken)

	for _, list := range []struct {
		dst *types.List
		src []string
	}{
		{&m.BudgetTokens, payload.BudgetTokens},
		{&m.RecipientChannels, payload.RecipientChannels},
		{&m.RecipientEmails, payload.RecipientEmails},
		{&m.UserTokens, payload.UserTokens},
	} {
		value, d := types.ListValueFrom(ctx, types.StringType, list.src)
		if d.HasError() {
			diags.Append(d...)
			return diags
		}
		*list.dst = value
	}

	return diags
}

func (m *budgetAlertModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateBudgetAlert {
	threshold := int32(m.Threshold.ValueInt64())
	durationInDays := m.DurationInDays.ValueString()

	payload := &modelsv2.CreateBudgetAlert{
		BudgetTokens:      stringListOrEmpty(ctx, m.BudgetTokens, diags),
		DurationInDays:    &durationInDays,
		Threshold:         &threshold,
		RecipientChannels: stringListOrNil(ctx, m.RecipientChannels, diags),
		RecipientEmails:   stringListOrNil(ctx, m.RecipientEmails, diags),
		UserTokens:        stringListOrNil(ctx, m.UserTokens, diags),
	}

	if !m.PeriodToTrack.IsNull() && !m.PeriodToTrack.IsUnknown() {
		payload.PeriodToTrack = m.PeriodToTrack.ValueString()
	}

	if !m.WorkspaceToken.IsNull() && !m.WorkspaceToken.IsUnknown() {
		payload.WorkspaceToken = m.WorkspaceToken.ValueString()
	}

	return payload
}

func (m *budgetAlertModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateBudgetAlert {
	payload := &modelsv2.UpdateBudgetAlert{
		BudgetTokens:      stringListOrEmpty(ctx, m.BudgetTokens, diags),
		DurationInDays:    m.DurationInDays.ValueString(),
		Threshold:         int32(m.Threshold.ValueInt64()),
		RecipientChannels: stringListOrNil(ctx, m.RecipientChannels, diags),
		RecipientEmails:   stringListOrNil(ctx, m.RecipientEmails, diags),
		UserTokens:        stringListOrNil(ctx, m.UserTokens, diags),
	}

	if !m.PeriodToTrack.IsNull() && !m.PeriodToTrack.IsUnknown() {
		payload.PeriodToTrack = m.PeriodToTrack.ValueString()
	}

	if !m.WorkspaceToken.IsNull() && !m.WorkspaceToken.IsUnknown() {
		payload.WorkspaceToken = m.WorkspaceToken.ValueString()
	}

	return payload
}
