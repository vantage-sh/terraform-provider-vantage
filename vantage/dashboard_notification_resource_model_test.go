package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_dashboard_notification"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestDashboardNotificationModel_toCreateAndApply(t *testing.T) {
	ctx := context.Background()
	userTokens, diags := types.ListValueFrom(ctx, types.StringType, []string{"usr_1"})
	if diags.HasError() {
		t.Fatalf("user tokens: %v", diags)
	}

	model := &resource_dashboard_notification.DashboardNotificationModel{
		Title:          types.StringValue("Weekly Dashboard"),
		DashboardToken: types.StringValue("dshbrd_1"),
		Frequency:      types.StringValue("weekly"),
		WorkspaceToken: types.StringValue("wrkspc_1"),
		UserTokens:     userTokens,
	}

	create, diags := toCreateDashboardNotification(ctx, model)
	if diags.HasError() {
		t.Fatalf("toCreate: %v", diags)
	}
	if create.Title == nil || *create.Title != "Weekly Dashboard" {
		t.Fatalf("unexpected title: %#v", create.Title)
	}
	if create.DashboardToken == nil || *create.DashboardToken != "dshbrd_1" {
		t.Fatalf("unexpected dashboard token: %#v", create.DashboardToken)
	}
	if create.Frequency == nil || *create.Frequency != "weekly" {
		t.Fatalf("unexpected frequency: %#v", create.Frequency)
	}
	if create.WorkspaceToken != "wrkspc_1" {
		t.Fatalf("unexpected workspace token: %q", create.WorkspaceToken)
	}
	if len(create.UserTokens) != 1 || create.UserTokens[0] != "usr_1" {
		t.Fatalf("unexpected user tokens: %#v", create.UserTokens)
	}
	if create.RecipientEmails != nil {
		t.Fatalf("omitted recipient emails should be nil, got %#v", create.RecipientEmails)
	}

	payload := &modelsv2.DashboardNotification{
		Token:           "dbnotif_1",
		Title:           "Weekly Dashboard",
		DashboardToken:  "dshbrd_1",
		Frequency:       "weekly",
		WorkspaceToken:  "wrkspc_1",
		UserTokens:      []string{"usr_1"},
		RecipientEmails: []string{"user@example.com"},
	}

	diags = applyDashboardNotificationPayload(ctx, payload, model)
	if diags.HasError() {
		t.Fatalf("applyPayload: %v", diags)
	}
	if model.Token.ValueString() != "dbnotif_1" || model.Id.ValueString() != "dbnotif_1" {
		t.Fatalf("token/id not applied: token=%s id=%s", model.Token.ValueString(), model.Id.ValueString())
	}
	if model.WorkspaceToken.ValueString() != "wrkspc_1" {
		t.Fatalf("workspace token not applied, got %q", model.WorkspaceToken.ValueString())
	}
	if model.RecipientEmails.IsNull() || len(model.RecipientEmails.Elements()) != 1 {
		t.Fatalf("recipient emails not applied: %#v", model.RecipientEmails)
	}
}

func TestDashboardNotificationModel_toUpdate(t *testing.T) {
	ctx := context.Background()
	userTokens, diags := types.ListValueFrom(ctx, types.StringType, []string{"usr_1", "usr_2"})
	if diags.HasError() {
		t.Fatalf("user tokens: %v", diags)
	}
	emails, diags := types.ListValueFrom(ctx, types.StringType, []string{"a@example.com"})
	if diags.HasError() {
		t.Fatalf("emails: %v", diags)
	}

	model := &resource_dashboard_notification.DashboardNotificationModel{
		Title:           types.StringValue("Updated"),
		DashboardToken:  types.StringValue("dshbrd_2"),
		Frequency:       types.StringValue("monthly"),
		UserTokens:      userTokens,
		RecipientEmails: emails,
	}

	update, diags := toUpdateDashboardNotification(ctx, model)
	if diags.HasError() {
		t.Fatalf("toUpdate: %v", diags)
	}
	if update.Title != "Updated" || update.DashboardToken != "dshbrd_2" || update.Frequency != "monthly" {
		t.Fatalf("unexpected update fields: %#v", update)
	}
	if len(update.UserTokens) != 2 {
		t.Fatalf("expected 2 user tokens, got %#v", update.UserTokens)
	}
	if len(update.RecipientEmails) != 1 {
		t.Fatalf("expected 1 recipient email, got %#v", update.RecipientEmails)
	}
}
