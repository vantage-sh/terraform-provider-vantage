package vantage

import "github.com/hashicorp/terraform-plugin-framework/types"

// Always send a non-nil pointer: omitting this field on PUT would leave the
// dashboard in its current folder when the attribute is removed from config.
func folderTokenRequest(value types.String) *string {
	token := ""
	if !value.IsNull() && !value.IsUnknown() {
		token = value.ValueString()
	}
	return &token
}

func folderTokenState(previous types.String, token *string) types.String {
	if token != nil && *token != "" {
		return types.StringValue(*token)
	}
	if !previous.IsNull() && !previous.IsUnknown() && previous.ValueString() == "" {
		return types.StringValue("")
	}
	return types.StringNull()
}
