package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_dashboard_notification"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func toCreateDashboardNotification(
	ctx context.Context,
	m *resource_dashboard_notification.DashboardNotificationModel,
) (*modelsv2.CreateDashboardNotification, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := &modelsv2.CreateDashboardNotification{
		Title:          m.Title.ValueStringPointer(),
		DashboardToken: m.DashboardToken.ValueStringPointer(),
		Frequency:      m.Frequency.ValueStringPointer(),
	}

	if !m.WorkspaceToken.IsNull() && !m.WorkspaceToken.IsUnknown() {
		payload.WorkspaceToken = m.WorkspaceToken.ValueString()
	}

	userTokens, d := stringSliceFromList(ctx, m.UserTokens)
	diags.Append(d...)
	payload.UserTokens = userTokens

	recipientEmails, d := stringSliceFromList(ctx, m.RecipientEmails)
	diags.Append(d...)
	payload.RecipientEmails = recipientEmails

	return payload, diags
}

func toUpdateDashboardNotification(
	ctx context.Context,
	m *resource_dashboard_notification.DashboardNotificationModel,
) (*modelsv2.UpdateDashboardNotification, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := &modelsv2.UpdateDashboardNotification{
		Title:          m.Title.ValueString(),
		DashboardToken: m.DashboardToken.ValueString(),
		Frequency:      m.Frequency.ValueString(),
	}

	if !m.UserTokens.IsNull() && !m.UserTokens.IsUnknown() {
		userTokens, d := stringSliceFromList(ctx, m.UserTokens)
		diags.Append(d...)
		payload.UserTokens = userTokens
	}

	if !m.RecipientEmails.IsNull() && !m.RecipientEmails.IsUnknown() {
		recipientEmails, d := stringSliceFromList(ctx, m.RecipientEmails)
		diags.Append(d...)
		payload.RecipientEmails = recipientEmails
	}

	return payload, diags
}

func applyDashboardNotificationPayload(
	ctx context.Context,
	payload *modelsv2.DashboardNotification,
	data *resource_dashboard_notification.DashboardNotificationModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Token = types.StringValue(payload.Token)
	data.Id = types.StringValue(payload.Token)
	data.Title = types.StringValue(payload.Title)
	data.DashboardToken = types.StringValue(payload.DashboardToken)
	data.Frequency = types.StringValue(payload.Frequency)

	if payload.UserTokens != nil {
		list, d := types.ListValueFrom(ctx, types.StringType, payload.UserTokens)
		diags.Append(d...)
		if d.HasError() {
			return diags
		}
		data.UserTokens = list
	} else {
		data.UserTokens = types.ListValueMust(types.StringType, nil)
	}

	if payload.RecipientEmails != nil {
		list, d := types.ListValueFrom(ctx, types.StringType, payload.RecipientEmails)
		diags.Append(d...)
		if d.HasError() {
			return diags
		}
		data.RecipientEmails = list
	} else {
		data.RecipientEmails = types.ListValueMust(types.StringType, nil)
	}

	// workspace_token is create-only and not returned by the API; preserve plan/state.

	return diags
}

func stringSliceFromList(ctx context.Context, value types.List) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := []string{}
	if value.IsNull() || value.IsUnknown() {
		return items, diags
	}

	diags.Append(value.ElementsAs(ctx, &items, false)...)
	return items, diags
}
