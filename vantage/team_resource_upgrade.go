package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_team"
)

// teamResourceModelV0 matches schema version 0, where member collections were lists.
type teamResourceModelV0 struct {
	DefaultDashboardToken types.String `tfsdk:"default_dashboard_token"`
	Description           types.String `tfsdk:"description"`
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Role                  types.String `tfsdk:"role"`
	Token                 types.String `tfsdk:"token"`
	UserEmails            types.List   `tfsdk:"user_emails"`
	UserTokens            types.List   `tfsdk:"user_tokens"`
	WorkspaceTokens       types.List   `tfsdk:"workspace_tokens"`
}

// teamResourcePriorSchemaV0 is the schema used when state was written as version 0
// (lists for user_emails, user_tokens, and workspace_tokens).
func teamResourcePriorSchemaV0(ctx context.Context) *schema.Schema {
	s := resource_team.TeamResourceSchema(ctx)
	return &s
}

func upgradeTeamStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var prior teamResourceModelV0

	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	upgraded := teamResourceModel{
		DefaultDashboardToken: prior.DefaultDashboardToken,
		Description:           prior.Description,
		Id:                    prior.Id,
		Name:                  prior.Name,
		Role:                  prior.Role,
		Token:                 prior.Token,
		UserEmails:            stringListToSet(ctx, prior.UserEmails, &resp.Diagnostics),
		UserTokens:            stringListToSet(ctx, prior.UserTokens, &resp.Diagnostics),
		WorkspaceTokens:       stringListToSet(ctx, prior.WorkspaceTokens, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
}

// stringListToSet converts a prior-schema string list into a set, preserving null
// and deduplicating elements so list→set upgrades never fail on duplicates.
func stringListToSet(ctx context.Context, list types.List, diags *diag.Diagnostics) types.Set {
	if list.IsNull() {
		return types.SetNull(types.StringType)
	}
	if list.IsUnknown() {
		return types.SetUnknown(types.StringType)
	}

	var elems []string
	diags.Append(list.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return types.SetNull(types.StringType)
	}

	seen := make(map[string]struct{}, len(elems))
	unique := make([]string, 0, len(elems))
	for _, e := range elems {
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		unique = append(unique, e)
	}

	set, d := types.SetValueFrom(ctx, types.StringType, unique)
	diags.Append(d...)
	return set
}
