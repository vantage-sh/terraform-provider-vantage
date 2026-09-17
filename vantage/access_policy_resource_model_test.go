package vantage

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestAccessPolicyModel_toCreateAndApplyPayload(t *testing.T) {
	ctx := context.Background()

	policy, d := types.ObjectValue(accessPolicyPolicyAttrTypes, map[string]attr.Value{
		"api_version": types.StringValue("v1"),
		"filter":      types.StringValue("(vantage.provider = 'aws')"),
	})
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}

	teamTokens, d := types.ListValueFrom(ctx, types.StringType, []string{"team_abc"})
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}

	m := &accessPolicyModel{
		Title:       types.StringValue("Engineering costs"),
		Description: types.StringValue("Limits cost visibility"),
		Policy:      policy,
		TeamTokens:  teamTokens,
	}

	var diags diag.Diagnostics
	create := m.toCreate(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if create.Title == nil || *create.Title != "Engineering costs" {
		t.Fatalf("unexpected title: %#v", create.Title)
	}
	if create.Description == nil || *create.Description != "Limits cost visibility" {
		t.Fatalf("unexpected description: %#v", create.Description)
	}
	if create.Policy == nil || create.Policy.APIVersion == nil || *create.Policy.APIVersion != "v1" {
		t.Fatalf("unexpected policy api_version: %#v", create.Policy)
	}
	if create.Policy.Policy == nil || create.Policy.Policy.Filter == nil || *create.Policy.Policy.Filter != "(vantage.provider = 'aws')" {
		t.Fatalf("unexpected policy filter: %#v", create.Policy.Policy)
	}
	if len(create.TeamTokens) != 1 || create.TeamTokens[0] != "team_abc" {
		t.Fatalf("unexpected team tokens: %#v", create.TeamTokens)
	}

	payload := &modelsv2.AccessPolicy{
		Token:       "accss_plcy_test",
		Title:       "Engineering costs",
		Description: create.Description,
		Policy: &modelsv2.AccessPolicyDocument{
			APIVersion: modelsv2.AccessPolicyDocumentAPIVersionV1,
			Policy: &modelsv2.AccessPolicyRules{
				Filter: "(vantage.provider = 'aws')",
			},
		},
		TeamTokens: []string{"team_abc"},
	}

	diags = m.applyPayload(ctx, payload)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if m.Token.ValueString() != "accss_plcy_test" || m.Id.ValueString() != "accss_plcy_test" {
		t.Fatalf("unexpected token/id: %s / %s", m.Token.ValueString(), m.Id.ValueString())
	}

	var applied accessPolicyPolicyModel
	if d := m.Policy.As(ctx, &applied, basetypes.ObjectAsOptions{}); d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}
	if applied.APIVersion.ValueString() != "v1" || applied.Filter.ValueString() != "(vantage.provider = 'aws')" {
		t.Fatalf("unexpected applied policy: %#v", applied)
	}
}

func TestAccessPolicyModel_teamTokensDefaultEmpty(t *testing.T) {
	ctx := context.Background()
	policy, d := types.ObjectValue(accessPolicyPolicyAttrTypes, map[string]attr.Value{
		"api_version": types.StringValue("v1"),
		"filter":      types.StringValue("(vantage.provider = 'aws')"),
	})
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}

	m := &accessPolicyModel{
		Title:      types.StringValue("No teams"),
		Policy:     policy,
		TeamTokens: types.ListNull(types.StringType),
	}

	var diags diag.Diagnostics
	create := m.toCreate(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if create.TeamTokens == nil || len(create.TeamTokens) != 0 {
		t.Fatalf("expected empty team tokens, got %#v", create.TeamTokens)
	}
}

func TestAccessPolicyModel_toUpdateClearsNullDescription(t *testing.T) {
	ctx := context.Background()
	policy, d := types.ObjectValue(accessPolicyPolicyAttrTypes, map[string]attr.Value{
		"api_version": types.StringValue("v1"),
		"filter":      types.StringValue("vantage.provider = 'aws'"),
	})
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}

	m := &accessPolicyModel{
		Title:       types.StringValue("Clear description"),
		Description: types.StringNull(),
		Policy:      policy,
		TeamTokens:  types.ListNull(types.StringType),
	}

	var diags diag.Diagnostics
	update := m.toUpdate(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if update.Description == nil {
		t.Fatal("expected description pointer so omitempty still sends an empty string")
	}
	if *update.Description != "" {
		t.Fatalf("expected empty description to clear API value, got %#v", *update.Description)
	}
}

func TestAccessPolicyModel_toUpdateSerializesEmptyTeamTokens(t *testing.T) {
	ctx := context.Background()
	policy, d := types.ObjectValue(accessPolicyPolicyAttrTypes, map[string]attr.Value{
		"api_version": types.StringValue("v1"),
		"filter":      types.StringValue("vantage.provider = 'aws'"),
	})
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}

	emptyTeams, d := types.ListValueFrom(ctx, types.StringType, []string{})
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}

	m := &accessPolicyModel{
		Title:      types.StringValue("Clear teams"),
		Policy:     policy,
		TeamTokens: emptyTeams,
	}

	var diags diag.Diagnostics
	update := m.toUpdate(ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if update.TeamTokens == nil {
		t.Fatal("expected non-nil empty team_tokens slice")
	}
	if len(update.TeamTokens) != 0 {
		t.Fatalf("expected empty team_tokens, got %#v", update.TeamTokens)
	}

	encoded, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	if got := string(encoded); !strings.Contains(got, `"team_tokens":[]`) {
		t.Fatalf("update JSON = %s, want explicit empty team_tokens array", got)
	}
}

func TestAccessPolicyModel_applyPayloadNullsBlankDescription(t *testing.T) {
	ctx := context.Background()
	m := &accessPolicyModel{}
	blank := ""
	diags := m.applyPayload(ctx, &modelsv2.AccessPolicy{
		Token:       "accss_plcy_blank",
		Title:       "Blank description",
		Description: &blank,
		Policy: &modelsv2.AccessPolicyDocument{
			APIVersion: modelsv2.AccessPolicyDocumentAPIVersionV1,
			Policy: &modelsv2.AccessPolicyRules{
				Filter: "vantage.provider = 'aws'",
			},
		},
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !m.Description.IsNull() {
		t.Fatalf("expected blank API description to map to null, got %#v", m.Description)
	}
}
