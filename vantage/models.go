package vantage

import "github.com/hashicorp/terraform-plugin-framework/types"

// accessGrant is a struct that represents the access grant data model. Used in both resources and data sources.
type accessGrant struct {
	Token         types.String `tfsdk:"token"`
	TeamToken     types.String `tfsdk:"team_token"`
	ResourceToken types.String `tfsdk:"resource_token"`
	Access        types.String `tfsdk:"access"`
}

type accessGrants struct {
	AccessGrants []accessGrant `tfsdk:"access_grants"`
}

// costReport is a struct that represents the cost report data model. Used in both resources and data sources.
type costReport struct {
	Token             types.String `tfsdk:"token"`
	Title             types.String `tfsdk:"title"`
	FolderToken       types.String `tfsdk:"folder_token"`
	Filter            types.String `tfsdk:"filter"`
	SavedFilterTokens types.List   `tfsdk:"saved_filter_tokens"`
	WorkspaceToken    types.String `tfsdk:"workspace_token"`
	Groupings         types.String `tfsdk:"groupings"`
}

type costReports struct {
	CostReports []costReport `tfsdk:"cost_reports"`
}
