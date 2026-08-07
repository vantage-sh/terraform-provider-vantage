package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestBudgetAlertDurationToAPI(t *testing.T) {
	tests := []struct {
		name     string
		duration types.Int64
		want     string
	}{
		{"unset tracks the full month", types.Int64Null(), ""},
		{"unknown tracks the full month", types.Int64Unknown(), ""},
		{"seven days", types.Int64Value(7), "7"},
		{"thirty one days", types.Int64Value(31), "31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := budgetAlertDurationToAPI(tt.duration); got != tt.want {
				t.Fatalf("budgetAlertDurationToAPI(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestBudgetAlertDurationFromAPI(t *testing.T) {
	seven := int32(7)
	zero := int32(0)

	tests := []struct {
		name     string
		duration *int32
		want     types.Int64
	}{
		{"absent duration is the full month", nil, types.Int64Null()},
		{"zero duration is the full month", &zero, types.Int64Null()},
		{"seven days", &seven, types.Int64Value(7)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := budgetAlertDurationFromAPI(tt.duration); !got.Equal(tt.want) {
				t.Fatalf("budgetAlertDurationFromAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBudgetAlertApplyPayload(t *testing.T) {
	ctx := context.Background()
	periodToTrack := "end_of_the_month"
	integrationProvider := "slack"
	userToken := "usr_1"
	workspaceToken := "wrkspc_1"
	duration := int32(7)

	model := &budgetAlertModel{}
	diags := model.applyPayload(ctx, &modelsv2.BudgetAlert{
		BudgetTokens:        []string{"bdgt_1", "bdgt_2"},
		CreatedAt:           "2026-03-19T00:00:00Z",
		DurationInDays:      &duration,
		IntegrationProvider: &integrationProvider,
		PeriodToTrack:       &periodToTrack,
		RecipientChannels:   []string{"#costs"},
		Threshold:           100,
		Token:               "bdgtalrt_1",
		UserToken:           &userToken,
		UserTokens:          []string{"usr_1", "usr_2"},
		WorkspaceToken:      &workspaceToken,
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := model.Token.ValueString(); got != "bdgtalrt_1" {
		t.Fatalf("token = %q, want %q", got, "bdgtalrt_1")
	}
	// The resource aliases id to token.
	if got := model.Id.ValueString(); got != "bdgtalrt_1" {
		t.Fatalf("id = %q, want %q", got, "bdgtalrt_1")
	}
	if got := model.Threshold.ValueInt64(); got != 100 {
		t.Fatalf("threshold = %d, want 100", got)
	}
	if got := model.DurationInDays.ValueInt64(); got != 7 {
		t.Fatalf("duration_in_days = %d, want 7", got)
	}
	if got := model.PeriodToTrack.ValueString(); got != periodToTrack {
		t.Fatalf("period_to_track = %q, want %q", got, periodToTrack)
	}
	if got := model.WorkspaceToken.ValueString(); got != workspaceToken {
		t.Fatalf("workspace_token = %q, want %q", got, workspaceToken)
	}
	if got := len(model.BudgetTokens.Elements()); got != 2 {
		t.Fatalf("budget_tokens length = %d, want 2", got)
	}
	if got := len(model.UserTokens.Elements()); got != 2 {
		t.Fatalf("user_tokens length = %d, want 2", got)
	}
}

// TestBudgetAlertApplyPayloadEmptyLists checks that arrays the API omits become
// empty lists rather than null, so a config that never sets them does not drift.
func TestBudgetAlertApplyPayloadEmptyLists(t *testing.T) {
	ctx := context.Background()

	model := &budgetAlertModel{}
	diags := model.applyPayload(ctx, &modelsv2.BudgetAlert{
		Token:     "bdgtalrt_1",
		CreatedAt: "2026-03-19T00:00:00Z",
		Threshold: 100,
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	for name, list := range map[string]types.List{
		"budget_tokens":      model.BudgetTokens,
		"recipient_channels": model.RecipientChannels,
		"user_tokens":        model.UserTokens,
	} {
		if list.IsNull() {
			t.Fatalf("%s is null, want an empty list", name)
		}
		if got := len(list.Elements()); got != 0 {
			t.Fatalf("%s length = %d, want 0", name, got)
		}
	}

	if !model.DurationInDays.IsNull() {
		t.Fatalf("duration_in_days = %v, want null", model.DurationInDays)
	}
}

func TestBudgetAlertToCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("full month sends an empty duration", func(t *testing.T) {
		var diags diag.Diagnostics
		model := &budgetAlertModel{
			BudgetTokens:      testBudgetAlertList(t, "bdgt_1"),
			Threshold:         types.Int64Value(100),
			DurationInDays:    types.Int64Null(),
			RecipientChannels: types.ListNull(types.StringType),
			UserTokens:        types.ListNull(types.StringType),
			PeriodToTrack:     types.StringNull(),
		}

		payload := model.toCreate(ctx, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}

		if payload.DurationInDays == nil {
			t.Fatal("duration_in_days is nil, want a pointer to the empty string")
		}
		if *payload.DurationInDays != "" {
			t.Fatalf("duration_in_days = %q, want the empty string", *payload.DurationInDays)
		}
		if payload.Threshold == nil || *payload.Threshold != 100 {
			t.Fatalf("threshold = %v, want 100", payload.Threshold)
		}
		// The API rejects a null array, so unset lists become empty ones.
		if payload.RecipientChannels == nil {
			t.Fatal("recipient_channels is nil, want an empty array")
		}
		if payload.UserTokens == nil {
			t.Fatal("user_tokens is nil, want an empty array")
		}
		if payload.PeriodToTrack != "" {
			t.Fatalf("period_to_track = %q, want the empty string", payload.PeriodToTrack)
		}
	})

	t.Run("set duration is sent as a string", func(t *testing.T) {
		var diags diag.Diagnostics
		model := &budgetAlertModel{
			BudgetTokens:      testBudgetAlertList(t, "bdgt_1", "bdgt_2"),
			Threshold:         types.Int64Value(80),
			DurationInDays:    types.Int64Value(7),
			RecipientChannels: testBudgetAlertList(t, "#costs"),
			UserTokens:        types.ListNull(types.StringType),
			PeriodToTrack:     types.StringValue("end_of_the_month"),
		}

		payload := model.toCreate(ctx, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}

		if *payload.DurationInDays != "7" {
			t.Fatalf("duration_in_days = %q, want %q", *payload.DurationInDays, "7")
		}
		if got := len(payload.BudgetTokens); got != 2 {
			t.Fatalf("budget_tokens length = %d, want 2", got)
		}
		if payload.PeriodToTrack != "end_of_the_month" {
			t.Fatalf("period_to_track = %q, want %q", payload.PeriodToTrack, "end_of_the_month")
		}
	})
}

func TestBudgetAlertToUpdate(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &budgetAlertModel{
		BudgetTokens:      testBudgetAlertList(t, "bdgt_1"),
		Threshold:         types.Int64Value(90),
		DurationInDays:    types.Int64Value(14),
		RecipientChannels: testBudgetAlertList(t, "#costs", "#ops"),
		UserTokens:        types.ListNull(types.StringType),
		PeriodToTrack:     types.StringValue("start_of_the_month"),
	}

	payload := model.toUpdate(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if payload.DurationInDays != "14" {
		t.Fatalf("duration_in_days = %q, want %q", payload.DurationInDays, "14")
	}
	if payload.Threshold != 90 {
		t.Fatalf("threshold = %d, want 90", payload.Threshold)
	}
	if got := len(payload.RecipientChannels); got != 2 {
		t.Fatalf("recipient_channels length = %d, want 2", got)
	}
	// An unset list clears the field rather than sending null.
	if payload.UserTokens == nil {
		t.Fatal("user_tokens is nil, want an empty array")
	}
	if got := len(payload.UserTokens); got != 0 {
		t.Fatalf("user_tokens length = %d, want 0", got)
	}
}

func testBudgetAlertList(t *testing.T, values ...string) types.List {
	t.Helper()

	list, diags := stringListFrom(values)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics building a list: %v", diags)
	}
	return list
}
