package vantage

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

// budgetAlertModel mirrors the generated resource schema. It is written by hand
// because duration_in_days is declared as an integer, while the generated create
// and update bodies type that field as a string.
type budgetAlertModel struct {
	BudgetTokens        types.List   `tfsdk:"budget_tokens"`
	CreatedAt           types.String `tfsdk:"created_at"`
	DurationInDays      types.Int64  `tfsdk:"duration_in_days"`
	Id                  types.String `tfsdk:"id"`
	IntegrationProvider types.String `tfsdk:"integration_provider"`
	PeriodToTrack       types.String `tfsdk:"period_to_track"`
	RecipientChannels   types.List   `tfsdk:"recipient_channels"`
	Threshold           types.Int64  `tfsdk:"threshold"`
	Token               types.String `tfsdk:"token"`
	UserToken           types.String `tfsdk:"user_token"`
	UserTokens          types.List   `tfsdk:"user_tokens"`
	WorkspaceToken      types.String `tfsdk:"workspace_token"`
}

// budgetAlertDataSourceModel mirrors the generated data source schema, where
// every field the API returns is read-only.
type budgetAlertDataSourceModel struct {
	BudgetTokens        types.List   `tfsdk:"budget_tokens"`
	CreatedAt           types.String `tfsdk:"created_at"`
	DurationInDays      types.Int64  `tfsdk:"duration_in_days"`
	Id                  types.String `tfsdk:"id"`
	IntegrationProvider types.String `tfsdk:"integration_provider"`
	PeriodToTrack       types.String `tfsdk:"period_to_track"`
	RecipientChannels   types.List   `tfsdk:"recipient_channels"`
	Threshold           types.Int64  `tfsdk:"threshold"`
	Token               types.String `tfsdk:"token"`
	UserToken           types.String `tfsdk:"user_token"`
	UserTokens          types.List   `tfsdk:"user_tokens"`
	WorkspaceToken      types.String `tfsdk:"workspace_token"`
}

func (m *budgetAlertModel) applyPayload(ctx context.Context, payload *modelsv2.BudgetAlert) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.IntegrationProvider = types.StringPointerValue(payload.IntegrationProvider)
	m.PeriodToTrack = types.StringPointerValue(payload.PeriodToTrack)
	m.Threshold = types.Int64Value(int64(payload.Threshold))
	m.UserToken = types.StringPointerValue(payload.UserToken)
	m.WorkspaceToken = types.StringPointerValue(payload.WorkspaceToken)
	m.DurationInDays = budgetAlertDurationFromAPI(payload.DurationInDays)

	budgetTokens, diags := stringListFrom(payload.BudgetTokens)
	if diags.HasError() {
		return diags
	}
	m.BudgetTokens = budgetTokens

	recipientChannels, diags := stringListFrom(payload.RecipientChannels)
	if diags.HasError() {
		return diags
	}
	m.RecipientChannels = recipientChannels

	userTokens, diags := stringListFrom(payload.UserTokens)
	if diags.HasError() {
		return diags
	}
	m.UserTokens = userTokens

	return diags
}

func (m *budgetAlertModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.CreateBudgetAlert {
	threshold := int32(m.Threshold.ValueInt64())
	duration := budgetAlertDurationToAPI(m.DurationInDays)

	payload := &modelsv2.CreateBudgetAlert{
		BudgetTokens:      budgetAlertStrings(ctx, m.BudgetTokens, diags),
		DurationInDays:    &duration,
		RecipientChannels: budgetAlertStrings(ctx, m.RecipientChannels, diags),
		Threshold:         &threshold,
		UserTokens:        budgetAlertStrings(ctx, m.UserTokens, diags),
	}

	if !m.PeriodToTrack.IsNull() && !m.PeriodToTrack.IsUnknown() {
		payload.PeriodToTrack = m.PeriodToTrack.ValueString()
	}

	return payload
}

func (m *budgetAlertModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *modelsv2.UpdateBudgetAlert {
	payload := &modelsv2.UpdateBudgetAlert{
		BudgetTokens:      budgetAlertStrings(ctx, m.BudgetTokens, diags),
		DurationInDays:    budgetAlertDurationToAPI(m.DurationInDays),
		RecipientChannels: budgetAlertStrings(ctx, m.RecipientChannels, diags),
		Threshold:         int32(m.Threshold.ValueInt64()),
		UserTokens:        budgetAlertStrings(ctx, m.UserTokens, diags),
	}

	if !m.PeriodToTrack.IsNull() && !m.PeriodToTrack.IsUnknown() {
		payload.PeriodToTrack = m.PeriodToTrack.ValueString()
	}

	return payload
}

// applyPayloadDataSource populates the data source model from the API response.
func (m *budgetAlertDataSourceModel) applyPayloadDataSource(ctx context.Context, payload *modelsv2.BudgetAlert) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Token = types.StringValue(payload.Token)
	m.Id = types.StringValue(payload.Token)
	m.CreatedAt = types.StringValue(payload.CreatedAt)
	m.IntegrationProvider = types.StringPointerValue(payload.IntegrationProvider)
	m.PeriodToTrack = types.StringPointerValue(payload.PeriodToTrack)
	m.Threshold = types.Int64Value(int64(payload.Threshold))
	m.UserToken = types.StringPointerValue(payload.UserToken)
	m.WorkspaceToken = types.StringPointerValue(payload.WorkspaceToken)
	m.DurationInDays = budgetAlertDurationFromAPI(payload.DurationInDays)

	budgetTokens, diags := stringListFrom(payload.BudgetTokens)
	if diags.HasError() {
		return diags
	}
	m.BudgetTokens = budgetTokens

	recipientChannels, diags := stringListFrom(payload.RecipientChannels)
	if diags.HasError() {
		return diags
	}
	m.RecipientChannels = recipientChannels

	userTokens, diags := stringListFrom(payload.UserTokens)
	if diags.HasError() {
		return diags
	}
	m.UserTokens = userTokens

	return diags
}

// budgetAlertDurationToAPI renders duration_in_days for the create and update
// bodies, which both type the field as a string. An unset duration becomes the
// empty string, the value the API reads as "track the full month".
func budgetAlertDurationToAPI(duration types.Int64) string {
	if duration.IsNull() || duration.IsUnknown() {
		return ""
	}
	return strconv.FormatInt(duration.ValueInt64(), 10)
}

// budgetAlertDurationFromAPI reads duration_in_days out of a response, where the
// field is an integer. The API reports the full month as an absent or zero
// duration, and both map to a null attribute so that a config which omits
// duration_in_days does not drift.
func budgetAlertDurationFromAPI(duration *int32) types.Int64 {
	if duration == nil || *duration == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*duration))
}

// budgetAlertStrings converts a Terraform list into the string slice the API
// expects. The API rejects a null array, so an unset list becomes an empty one.
func budgetAlertStrings(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	values := terraformListToStrings(ctx, list, diags)
	if values == nil {
		return []string{}
	}
	return values
}
