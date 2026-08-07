package vantage

import (
	"context"
	"encoding/json"
	"strings"
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
	one := int32(1)
	seven := int32(7)
	thirtyOne := int32(31)

	tests := []struct {
		name     string
		duration *int32
		want     types.Int64
	}{
		{"absent duration is the full month", nil, types.Int64Null()},
		// The API enforces a range of 1 to 31, so those are the only other
		// values a response can carry.
		{"one day", &one, types.Int64Value(1)},
		{"seven days", &seven, types.Int64Value(7)},
		{"thirty one days", &thirtyOne, types.Int64Value(31)},
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

	for name, set := range map[string]types.Set{
		"budget_tokens":      model.BudgetTokens,
		"recipient_channels": model.RecipientChannels,
		"user_tokens":        model.UserTokens,
	} {
		if set.IsNull() {
			t.Fatalf("%s is null, want an empty set", name)
		}
		if got := len(set.Elements()); got != 0 {
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
			BudgetTokens:      testBudgetAlertSet(t, "bdgt_1"),
			Threshold:         types.Int64Value(100),
			DurationInDays:    types.Int64Null(),
			RecipientChannels: types.SetNull(types.StringType),
			UserTokens:        types.SetNull(types.StringType),
			PeriodToTrack:     types.StringNull(),
		}

		payload := model.toCreate(ctx, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}

		// The API requires duration_in_days on create and reads its empty value
		// as "track the full month".
		if payload.DurationInDays != "" {
			t.Fatalf("duration_in_days = %q, want the empty string", payload.DurationInDays)
		}
		if payload.Threshold != 100 {
			t.Fatalf("threshold = %d, want 100", payload.Threshold)
		}
		// An empty recipient list has to be absent. The API answers
		// "user_tokens is empty" to an empty array and to a null alike.
		if payload.RecipientChannels != nil {
			t.Fatalf("recipient_channels = %v, want it left out", payload.RecipientChannels)
		}
		if payload.UserTokens != nil {
			t.Fatalf("user_tokens = %v, want it left out", payload.UserTokens)
		}
		if payload.PeriodToTrack != "" {
			t.Fatalf("period_to_track = %q, want the empty string", payload.PeriodToTrack)
		}
	})

	t.Run("set duration is sent as a string", func(t *testing.T) {
		var diags diag.Diagnostics
		model := &budgetAlertModel{
			BudgetTokens:      testBudgetAlertSet(t, "bdgt_1", "bdgt_2"),
			Threshold:         types.Int64Value(80),
			DurationInDays:    types.Int64Value(7),
			RecipientChannels: testBudgetAlertSet(t, "#costs"),
			UserTokens:        types.SetNull(types.StringType),
			PeriodToTrack:     types.StringValue("end_of_the_month"),
		}

		payload := model.toCreate(ctx, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}

		if payload.DurationInDays != "7" {
			t.Fatalf("duration_in_days = %q, want %q", payload.DurationInDays, "7")
		}
		if got := len(payload.BudgetTokens); got != 2 {
			t.Fatalf("budget_tokens length = %d, want 2", got)
		}
		if got := len(payload.RecipientChannels); got != 1 {
			t.Fatalf("recipient_channels length = %d, want 1", got)
		}
		if payload.PeriodToTrack != "end_of_the_month" {
			t.Fatalf("period_to_track = %q, want %q", payload.PeriodToTrack, "end_of_the_month")
		}
	})
}

// TestBudgetAlertToCreateJSON pins the wire format, which is the thing the API
// actually validates.
func TestBudgetAlertToCreateJSON(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &budgetAlertModel{
		BudgetTokens:      testBudgetAlertSet(t, "bdgt_1"),
		Threshold:         types.Int64Value(100),
		DurationInDays:    types.Int64Null(),
		RecipientChannels: testBudgetAlertSet(t, "#costs"),
		UserTokens:        types.SetNull(types.StringType),
		PeriodToTrack:     types.StringNull(),
	}

	body, err := json.Marshal(model.toCreate(ctx, &diags))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"budget_tokens":["bdgt_1"],"duration_in_days":"","recipient_channels":["#costs"],"threshold":100}`
	if string(body) != want {
		t.Fatalf("body = %s, want %s", body, want)
	}
}

func TestBudgetAlertToUpdate(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &budgetAlertModel{
		BudgetTokens:      testBudgetAlertSet(t, "bdgt_1"),
		Threshold:         types.Int64Value(90),
		DurationInDays:    types.Int64Value(14),
		RecipientChannels: testBudgetAlertSet(t, "#costs", "#ops"),
		UserTokens:        types.SetNull(types.StringType),
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
	if payload.RecipientChannels == nil {
		t.Fatal("recipient_channels is nil, want the two configured channels")
	}
	if got := len(*payload.RecipientChannels); got != 2 {
		t.Fatalf("recipient_channels length = %d, want 2", got)
	}
	// An empty user_tokens has to be absent, not sent as an empty array.
	if payload.UserTokens != nil {
		t.Fatalf("user_tokens = %v, want it left out", payload.UserTokens)
	}
}

// TestBudgetAlertToUpdateJSON pins the update wire format. The reported apply
// failure was an empty user_tokens reaching the API as "user_tokens": [].
func TestBudgetAlertToUpdateJSON(t *testing.T) {
	ctx := context.Background()

	t.Run("empty recipients are left out", func(t *testing.T) {
		var diags diag.Diagnostics
		model := &budgetAlertModel{
			BudgetTokens:      testBudgetAlertSet(t, "bdgt_1"),
			Threshold:         types.Int64Value(90),
			DurationInDays:    types.Int64Null(),
			RecipientChannels: testBudgetAlertSet(t, "#costs"),
			UserTokens:        testBudgetAlertSet(t),
			PeriodToTrack:     types.StringValue("start_of_the_month"),
		}

		body, err := json.Marshal(model.toUpdate(ctx, &diags))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if strings.Contains(string(body), "user_tokens") {
			t.Fatalf("body still carries an empty user_tokens: %s", body)
		}
		want := `{"budget_tokens":["bdgt_1"],"period_to_track":"start_of_the_month","recipient_channels":["#costs"],"threshold":90}`
		if string(body) != want {
			t.Fatalf("body = %s, want %s", body, want)
		}
	})

	t.Run("an emptied channel list is sent so the API clears it", func(t *testing.T) {
		var diags diag.Diagnostics
		model := &budgetAlertModel{
			BudgetTokens:      testBudgetAlertSet(t, "bdgt_1"),
			Threshold:         types.Int64Value(90),
			DurationInDays:    types.Int64Null(),
			RecipientChannels: testBudgetAlertSet(t),
			UserTokens:        testBudgetAlertSet(t, "usr_1"),
			PeriodToTrack:     types.StringNull(),
		}

		body, err := json.Marshal(model.toUpdate(ctx, &diags))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := `{"budget_tokens":["bdgt_1"],"recipient_channels":[],"threshold":90,"user_tokens":["usr_1"]}`
		if string(body) != want {
			t.Fatalf("body = %s, want %s", body, want)
		}
	})
}

func testBudgetAlertSet(t *testing.T, values ...string) types.Set {
	t.Helper()

	set, diags := budgetAlertStringSet(context.Background(), values)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics building a set: %v", diags)
	}
	return set
}
