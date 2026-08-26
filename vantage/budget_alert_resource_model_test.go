package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func stringList(t *testing.T, values ...string) types.List {
	t.Helper()

	list, diagnostics := types.ListValueFrom(context.Background(), types.StringType, values)
	if diagnostics.HasError() {
		t.Fatalf("building string list: %v", diagnostics)
	}
	return list
}

func TestBudgetAlertCreateMapping(t *testing.T) {
	t.Parallel()

	model := budgetAlertModel{
		BudgetTokens:    stringList(t, "bdgt_123"),
		DurationInDays:  types.StringValue("7"),
		Threshold:       types.Int64Value(80),
		RecipientEmails: stringList(t, "finops@example.com"),
		PeriodToTrack:   types.StringValue("end_of_the_month"),
		UserTokens:      types.ListNull(types.StringType),
		WorkspaceToken:  types.StringValue("wrkspc_123"),
	}

	payload := model.toCreate(context.Background(), &diag.Diagnostics{})

	if got := payload.RecipientEmails; len(got) != 1 || got[0] != "finops@example.com" {
		t.Errorf("recipient_emails = %v, want [finops@example.com]", got)
	}
	if payload.DurationInDays == nil || *payload.DurationInDays != "7" {
		t.Errorf("duration_in_days = %v, want 7", payload.DurationInDays)
	}
	if payload.Threshold == nil || *payload.Threshold != 80 {
		t.Errorf("threshold = %v, want 80", payload.Threshold)
	}
	if got := payload.PeriodToTrack; got != "end_of_the_month" {
		t.Errorf("period_to_track = %q, want end_of_the_month", got)
	}
	if got := payload.BudgetTokens; len(got) != 1 || got[0] != "bdgt_123" {
		t.Errorf("budget_tokens = %v, want [bdgt_123]", got)
	}
	if got := payload.WorkspaceToken; got != "wrkspc_123" {
		t.Errorf("workspace_token = %q, want wrkspc_123", got)
	}
}

// Unset recipient lists are omitted so create/update can target a single
// recipient field without sending empty arrays for the others.
func TestBudgetAlertOmitsUnsetRecipientLists(t *testing.T) {
	t.Parallel()

	model := budgetAlertModel{
		BudgetTokens:      stringList(t, "bdgt_123"),
		DurationInDays:    types.StringValue("7"),
		Threshold:         types.Int64Value(80),
		RecipientEmails:   stringList(t, "finops@example.com"),
		UserTokens:        types.ListNull(types.StringType),
		RecipientChannels: stringList(t),
	}

	created := model.toCreate(context.Background(), &diag.Diagnostics{})
	if created.UserTokens != nil {
		t.Errorf("create user_tokens = %v, want nil", created.UserTokens)
	}
	if created.RecipientChannels != nil {
		t.Errorf("create recipient_channels = %v, want nil for an empty list", created.RecipientChannels)
	}

	updated := model.toUpdate(context.Background(), &diag.Diagnostics{})
	if updated.UserTokens != nil {
		t.Errorf("update user_tokens = %v, want nil", updated.UserTokens)
	}
	if updated.RecipientChannels != nil {
		t.Errorf("update recipient_channels = %v, want nil for an empty list", updated.RecipientChannels)
	}
}

// The API treats a present-but-null recipient field as "drop the existing
// recipients", and the generated client cannot omit the key. Recipient lists
// must therefore carry a plan modifier rather than planning as unknown.
func TestBudgetAlertRecipientListsPreservePriorValues(t *testing.T) {
	t.Parallel()

	s := (&budgetAlertResource{}).schema(context.Background())

	for _, name := range []string{"user_tokens", "recipient_emails", "recipient_channels"} {
		attr, ok := s.Attributes[name].(schema.ListAttribute)
		if !ok {
			t.Fatalf("%s is not a ListAttribute", name)
		}
		if len(attr.PlanModifiers) == 0 {
			t.Errorf("%s has no plan modifiers, so an omitted list would plan as unknown and clear the recipients", name)
		}
	}
}

// Recipients are required on create, but an update may omit them because they
// resolve from prior state, so the check must not be a config validator.
func TestBudgetAlertRecipientValidationIsCreateOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	s := (&budgetAlertResource{}).schema(ctx)

	emptyConfig := tfsdk.Config{
		Schema: s,
		Raw:    tftypes.NewValue(s.Type().TerraformType(ctx), nil),
	}

	var diags diag.Diagnostics
	validateBudgetAlertRecipients(ctx, emptyConfig, &diags)
	if !diags.HasError() {
		t.Error("a config with no recipients should fail on create")
	}

	if _, ok := interface{}((&budgetAlertResource{})).(resource.ResourceWithConfigValidators); ok {
		t.Error("recipient validation must not run as a config validator, since updates may omit recipients")
	}
}

// budgetAlertConfigWithUserTokens builds a config whose only populated
// recipient field is user_tokens.
func budgetAlertConfigWithUserTokens(t *testing.T, userTokens tftypes.Value) tfsdk.Config {
	t.Helper()

	ctx := context.Background()
	s := (&budgetAlertResource{}).schema(ctx)
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	values := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["user_tokens"] = userTokens

	return tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, values)}
}

func TestBudgetAlertRecipientValidationRejectsEmptyLists(t *testing.T) {
	t.Parallel()

	listType := tftypes.List{ElementType: tftypes.String}
	ctx := context.Background()

	cases := map[string]struct {
		userTokens tftypes.Value
		wantError  bool
	}{
		// An empty list is sent as null, which the API rejects, so it must be
		// caught at plan time rather than surfacing as a server error.
		"empty list":     {tftypes.NewValue(listType, []tftypes.Value{}), true},
		"populated list": {tftypes.NewValue(listType, []tftypes.Value{tftypes.NewValue(tftypes.String, "usr_1")}), false},
		// Resolved from elsewhere in the config, so it cannot be judged yet.
		"unknown list": {tftypes.NewValue(listType, tftypes.UnknownValue), false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var diags diag.Diagnostics
			validateBudgetAlertRecipients(ctx, budgetAlertConfigWithUserTokens(t, tc.userTokens), &diags)

			if got := diags.HasError(); got != tc.wantError {
				t.Errorf("validation error = %v, want %v (%v)", got, tc.wantError, diags)
			}
		})
	}
}

func TestBudgetAlertCreateOmitsUnknownPeriodToTrack(t *testing.T) {
	t.Parallel()

	model := budgetAlertModel{
		BudgetTokens:   stringList(t, "bdgt_123"),
		DurationInDays: types.StringValue(""),
		Threshold:      types.Int64Value(100),
		PeriodToTrack:  types.StringUnknown(),
	}

	payload := model.toCreate(context.Background(), &diag.Diagnostics{})

	if payload.PeriodToTrack != "" {
		t.Errorf("period_to_track = %q, want empty so the API applies its default", payload.PeriodToTrack)
	}
}

func TestBudgetAlertUpdateSendsRecipientEmails(t *testing.T) {
	t.Parallel()

	model := budgetAlertModel{
		BudgetTokens:      stringList(t, "bdgt_123"),
		DurationInDays:    types.StringValue("14"),
		Threshold:         types.Int64Value(90),
		RecipientEmails:   stringList(t, "a@example.com", "b@example.com"),
		RecipientChannels: types.ListNull(types.StringType),
		UserTokens:        stringList(t, "usr_123"),
	}

	payload := model.toUpdate(context.Background(), &diag.Diagnostics{})

	if got := payload.RecipientEmails; len(got) != 2 {
		t.Errorf("recipient_emails = %v, want 2 addresses", got)
	}
	if got := payload.UserTokens; len(got) != 1 || got[0] != "usr_123" {
		t.Errorf("user_tokens = %v, want [usr_123]", got)
	}
	if got := payload.Threshold; got != 90 {
		t.Errorf("threshold = %d, want 90", got)
	}
}

func TestBudgetAlertResponseMapping(t *testing.T) {
	t.Parallel()

	durationInDays := int32(7)
	periodToTrack := "start_of_the_month"
	workspaceToken := "wrkspc_123"

	model := budgetAlertModel{}
	diagnostics := model.applyPayload(context.Background(), &modelsv2.BudgetAlert{
		Token:           "bdgtalrt_123",
		CreatedAt:       "2024-03-19T00:00:00Z",
		DurationInDays:  &durationInDays,
		PeriodToTrack:   &periodToTrack,
		Threshold:       80,
		WorkspaceToken:  &workspaceToken,
		BudgetTokens:    []string{"bdgt_123"},
		RecipientEmails: []string{"finops@example.com"},
		UserTokens:      []string{},
	})
	if diagnostics.HasError() {
		t.Fatalf("applying payload: %v", diagnostics)
	}

	if got := model.Token.ValueString(); got != "bdgtalrt_123" {
		t.Errorf("token = %q, want bdgtalrt_123", got)
	}
	if got := model.Id.ValueString(); got != "bdgtalrt_123" {
		t.Errorf("id = %q, want bdgtalrt_123", got)
	}
	if got := model.DurationInDays.ValueString(); got != "7" {
		t.Errorf("duration_in_days = %q, want 7", got)
	}
	if got := model.Threshold.ValueInt64(); got != 80 {
		t.Errorf("threshold = %d, want 80", got)
	}
	if got := len(model.RecipientEmails.Elements()); got != 1 {
		t.Errorf("recipient_emails length = %d, want 1", got)
	}
}

// A full-month alert is written as an empty duration_in_days and read back as
// null, so the response has to map onto the same empty string to avoid drift.
func TestBudgetAlertFullMonthDurationRoundTrips(t *testing.T) {
	t.Parallel()

	model := budgetAlertModel{}
	diagnostics := model.applyPayload(context.Background(), &modelsv2.BudgetAlert{
		Token:           "bdgtalrt_123",
		DurationInDays:  nil,
		Threshold:       100,
		BudgetTokens:    []string{"bdgt_123"},
		RecipientEmails: []string{},
		UserTokens:      []string{},
	})
	if diagnostics.HasError() {
		t.Fatalf("applying payload: %v", diagnostics)
	}

	if model.DurationInDays.IsNull() {
		t.Fatal("duration_in_days is null, want empty string")
	}
	if got := model.DurationInDays.ValueString(); got != "" {
		t.Errorf("duration_in_days = %q, want empty string", got)
	}
}
