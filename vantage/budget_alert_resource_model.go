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
	BudgetTokens        types.Set    `tfsdk:"budget_tokens"`
	CreatedAt           types.String `tfsdk:"created_at"`
	DurationInDays      types.Int64  `tfsdk:"duration_in_days"`
	Id                  types.String `tfsdk:"id"`
	IntegrationProvider types.String `tfsdk:"integration_provider"`
	PeriodToTrack       types.String `tfsdk:"period_to_track"`
	RecipientChannels   types.Set    `tfsdk:"recipient_channels"`
	Threshold           types.Int64  `tfsdk:"threshold"`
	Token               types.String `tfsdk:"token"`
	UserToken           types.String `tfsdk:"user_token"`
	UserTokens          types.Set    `tfsdk:"user_tokens"`
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

	budgetTokens, diags := budgetAlertStringSet(ctx, payload.BudgetTokens)
	if diags.HasError() {
		return diags
	}
	m.BudgetTokens = budgetTokens

	recipientChannels, diags := budgetAlertStringSet(ctx, payload.RecipientChannels)
	if diags.HasError() {
		return diags
	}
	m.RecipientChannels = recipientChannels

	userTokens, diags := budgetAlertStringSet(ctx, payload.UserTokens)
	if diags.HasError() {
		return diags
	}
	m.UserTokens = userTokens

	return diags
}

// budgetAlertStringSet builds the set for a token field. The API returns these
// tokens in its own order, which is why they are sets. An array the API omits
// becomes an empty set rather than a null one, so a configuration that never
// names recipients does not drift.
func budgetAlertStringSet(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	if values == nil {
		values = []string{}
	}
	return types.SetValueFrom(ctx, types.StringType, values)
}

// createBudgetAlertBody is the create request body.
//
// The generated model tags the recipient arrays without omitempty, and the API
// rejects both "user_tokens": [] and "user_tokens": null with "user_tokens is
// empty". The key has to be absent, which the generated model cannot express.
// duration_in_days carries no omitempty on purpose: the API requires the field
// on create, and reads its empty value as "track the full month".
type createBudgetAlertBody struct {
	BudgetTokens      []string `json:"budget_tokens"`
	DurationInDays    string   `json:"duration_in_days"`
	PeriodToTrack     string   `json:"period_to_track,omitempty"`
	RecipientChannels []string `json:"recipient_channels,omitempty"`
	Threshold         int32    `json:"threshold"`
	UserTokens        []string `json:"user_tokens,omitempty"`
}

// updateBudgetAlertBody is the update request body.
//
// duration_in_days is left out when unset. The API keeps the stored value, and
// it refuses to clear one, so the resource is replaced instead.
//
// recipient_channels is a pointer because the API accepts an empty array there
// and clears the channels. An absent list is left out; an empty one is sent.
type updateBudgetAlertBody struct {
	BudgetTokens      []string  `json:"budget_tokens,omitempty"`
	DurationInDays    string    `json:"duration_in_days,omitempty"`
	PeriodToTrack     string    `json:"period_to_track,omitempty"`
	RecipientChannels *[]string `json:"recipient_channels,omitempty"`
	Threshold         int32     `json:"threshold"`
	UserTokens        []string  `json:"user_tokens,omitempty"`
}

func (m *budgetAlertModel) toCreate(ctx context.Context, diags *diag.Diagnostics) *createBudgetAlertBody {
	return &createBudgetAlertBody{
		BudgetTokens:      budgetAlertStrings(ctx, m.BudgetTokens, diags),
		DurationInDays:    budgetAlertDurationToAPI(m.DurationInDays),
		PeriodToTrack:     budgetAlertOptionalString(m.PeriodToTrack),
		RecipientChannels: budgetAlertStrings(ctx, m.RecipientChannels, diags),
		Threshold:         int32(m.Threshold.ValueInt64()),
		UserTokens:        budgetAlertStrings(ctx, m.UserTokens, diags),
	}
}

func (m *budgetAlertModel) toUpdate(ctx context.Context, diags *diag.Diagnostics) *updateBudgetAlertBody {
	return &updateBudgetAlertBody{
		BudgetTokens:      budgetAlertStrings(ctx, m.BudgetTokens, diags),
		DurationInDays:    budgetAlertDurationToAPI(m.DurationInDays),
		PeriodToTrack:     budgetAlertOptionalString(m.PeriodToTrack),
		RecipientChannels: budgetAlertClearableStrings(ctx, m.RecipientChannels, diags),
		Threshold:         int32(m.Threshold.ValueInt64()),
		UserTokens:        budgetAlertStrings(ctx, m.UserTokens, diags),
	}
}

// budgetAlertOptionalString returns the value to send for a field that is left
// out when the configuration does not set it.
func budgetAlertOptionalString(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
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
// field is an integer. The API reports a full month as an absent duration, which
// maps to a null attribute so that a configuration omitting duration_in_days does
// not drift. Any other value is between 1 and 31, the range the API enforces.
func budgetAlertDurationFromAPI(duration *int32) types.Int64 {
	if duration == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*duration))
}

// budgetAlertStrings returns the values to send, or nil to leave the field out
// of the request. The API answers "user_tokens is empty" to an empty array and
// to a null, so an empty list has to be absent rather than sent.
func budgetAlertStrings(ctx context.Context, set types.Set, diags *diag.Diagnostics) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	values := []string{}
	diags.Append(set.ElementsAs(ctx, &values, false)...)
	if len(values) == 0 {
		return nil
	}
	return values
}

// budgetAlertClearableStrings returns the values to send for a field that an
// empty array clears. A list the configuration does not set is left out, and an
// empty list is sent so that the API clears the stored value.
func budgetAlertClearableStrings(ctx context.Context, set types.Set, diags *diag.Diagnostics) *[]string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	values := []string{}
	diags.Append(set.ElementsAs(ctx, &values, false)...)
	return &values
}
