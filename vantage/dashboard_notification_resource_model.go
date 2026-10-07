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

	// Omit unset sibling recipient lists (nil) so the API can derive them; send
	// an empty slice only when the config explicitly clears the list.
	payload.UserTokens = stringListOrNil(ctx, m.UserTokens, &diags)
	payload.RecipientEmails = stringListOrNil(ctx, m.RecipientEmails, &diags)

	return payload, diags
}

func toUpdateDashboardNotification(
	ctx context.Context,
	m *resource_dashboard_notification.DashboardNotificationModel,
) (*modelsv2.UpdateDashboardNotification, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := &modelsv2.UpdateDashboardNotification{
		Title:           m.Title.ValueString(),
		DashboardToken:  m.DashboardToken.ValueString(),
		Frequency:       m.Frequency.ValueString(),
		UserTokens:      stringListOrNil(ctx, m.UserTokens, &diags),
		RecipientEmails: stringListOrNil(ctx, m.RecipientEmails, &diags),
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
	data.WorkspaceToken = types.StringValue(payload.WorkspaceToken)

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

	return diags
}
