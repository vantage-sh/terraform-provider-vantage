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

type dashboard struct {
	Token          types.String `tfsdk:"token"`
	Title          types.String `tfsdk:"title"`
	WidgetTokens   types.List   `tfsdk:"widget_tokens"`
	DateBin        types.String `tfsdk:"date_bin"`
	DateInterval   types.String `tfsdk:"date_interval"`
	StartDate      types.String `tfsdk:"start_date"`
	EndDate        types.String `tfsdk:"end_date"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}

type dashboards struct {
	Dashboards []dashboard `tfsdk:"dashboards"`
}

type folder struct {
	Title             types.String `tfsdk:"title"`
	ParentFolderToken types.String `tfsdk:"parent_folder_token"`
	Token             types.String `tfsdk:"token"`
	WorkspaceToken    types.String `tfsdk:"workspace_token"`
	SavedFilterTokens types.List   `tfsdk:"saved_filter_tokens"`
}

type folders struct {
	Folders []folder `tfsdk:"folders"`
}

type reportNotification struct {
	Title           types.String `tfsdk:"title"`
	Token           types.String `tfsdk:"token"`
	CostReportToken types.String `tfsdk:"cost_report_token"`
	WorkspaceToken  types.String `tfsdk:"workspace_token"`
	UserTokens      types.Set    `tfsdk:"user_tokens"`
	Frequency       types.String `tfsdk:"frequency"`
	Change          types.String `tfsdk:"change"`
}

type savedFilter struct {
	Token            types.String `tfsdk:"token"`
	Title            types.String `tfsdk:"title"`
	Filter           types.String `tfsdk:"filter"`
	WorkspaceToken   types.String `tfsdk:"workspace_token"`
	CostReportTokens types.List   `tfsdk:"cost_report_tokens"`
}

type savedFilters struct {
	Filters []savedFilter `tfsdk:"filters"`
}

type segment struct {
	Title              types.String `tfsdk:"title"`
	Description        types.String `tfsdk:"description"`
	Priority           types.Int64  `tfsdk:"priority"`
	WorkspaceToken     types.String `tfsdk:"workspace_token"`
	Filter             types.String `tfsdk:"filter"`
	ParentSegmentToken types.String `tfsdk:"parent_segment_token"`
	Token              types.String `tfsdk:"token"`
	TrackUnallocated   types.Bool   `tfsdk:"track_unallocated"`
}

type segments struct {
	Segments []segment `tfsdk:"segments"`
}
