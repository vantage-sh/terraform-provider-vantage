package vantage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUpgradeTeamStateV0toV1(t *testing.T) {
	ctx := context.Background()

	priorSchema := teamResourcePriorSchemaV0(ctx)
	priorState := tfsdk.State{Schema: *priorSchema}

	userEmails, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("b@example.com"),
		types.StringValue("a@example.com"),
		types.StringValue("b@example.com"), // duplicate should be dropped
	})
	if diags.HasError() {
		t.Fatalf("list value: %v", diags)
	}
	userTokens, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("usr_2"),
		types.StringValue("usr_1"),
	})
	if diags.HasError() {
		t.Fatalf("list value: %v", diags)
	}
	workspaceTokens := types.ListNull(types.StringType)

	prior := teamResourceModelV0{
		DefaultDashboardToken: types.StringValue(""),
		Description:           types.StringValue("desc"),
		Id:                    types.StringValue("team_abc"),
		Name:                  types.StringValue("Engineering"),
		Role:                  types.StringValue("editor"),
		Token:                 types.StringValue("team_abc"),
		UserEmails:            userEmails,
		UserTokens:            userTokens,
		WorkspaceTokens:       workspaceTokens,
	}
	diags = priorState.Set(ctx, prior)
	if diags.HasError() {
		t.Fatalf("set prior state: %v", diags)
	}

	resp := resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: teamResourceSchemaV1(ctx)},
	}
	upgradeTeamStateV0toV1(ctx, resource.UpgradeStateRequest{State: &priorState}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade diagnostics: %v", resp.Diagnostics)
	}

	var upgraded teamResourceModel
	diags = resp.State.Get(ctx, &upgraded)
	if diags.HasError() {
		t.Fatalf("get upgraded state: %v", diags)
	}

	if upgraded.Id.ValueString() != "team_abc" || upgraded.Token.ValueString() != "team_abc" {
		t.Fatalf("unexpected id/token: id=%q token=%q", upgraded.Id.ValueString(), upgraded.Token.ValueString())
	}
	if upgraded.Name.ValueString() != "Engineering" || upgraded.Description.ValueString() != "desc" {
		t.Fatalf("unexpected name/description: name=%q description=%q", upgraded.Name.ValueString(), upgraded.Description.ValueString())
	}
	if upgraded.WorkspaceTokens.IsNull() != true {
		t.Fatalf("expected null workspace_tokens, got %#v", upgraded.WorkspaceTokens)
	}

	assertSetContains(t, ctx, upgraded.UserEmails, "a@example.com", "b@example.com")
	assertSetLen(t, upgraded.UserEmails, 2)
	assertSetContains(t, ctx, upgraded.UserTokens, "usr_1", "usr_2")
	assertSetLen(t, upgraded.UserTokens, 2)
}

func TestStringListToSet_NullAndEmpty(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	nullSet := stringListToSet(ctx, types.ListNull(types.StringType), &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if !nullSet.IsNull() {
		t.Fatalf("expected null set, got %#v", nullSet)
	}

	emptyList, d := types.ListValue(types.StringType, []attr.Value{})
	if d.HasError() {
		t.Fatalf("list value: %v", d)
	}
	emptySet := stringListToSet(ctx, emptyList, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if emptySet.IsNull() || emptySet.IsUnknown() || len(emptySet.Elements()) != 0 {
		t.Fatalf("expected empty set, got %#v", emptySet)
	}
}

func assertSetLen(t *testing.T, set types.Set, want int) {
	t.Helper()
	if len(set.Elements()) != want {
		t.Fatalf("set len = %d, want %d", len(set.Elements()), want)
	}
}

func assertSetContains(t *testing.T, ctx context.Context, set types.Set, want ...string) {
	t.Helper()
	var got []string
	diags := set.ElementsAs(ctx, &got, false)
	if diags.HasError() {
		t.Fatalf("elements as: %v", diags)
	}
	have := make(map[string]struct{}, len(got))
	for _, g := range got {
		have[g] = struct{}{}
	}
	for _, w := range want {
		if _, ok := have[w]; !ok {
			t.Fatalf("set missing %q; got %v", w, got)
		}
	}
}
