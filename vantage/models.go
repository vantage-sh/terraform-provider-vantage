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
