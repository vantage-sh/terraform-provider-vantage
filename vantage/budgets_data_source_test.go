package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBudgetsDataSourceParamsIncludeFilters(t *testing.T) {
	t.Parallel()

	params := getBudgetsParams(budgetsDataSourceModel{
		Q:              types.StringValue("production"),
		WorkspaceToken: types.StringValue("wrkspc_test"),
	})

	if params.Q == nil || *params.Q != "production" {
		t.Fatalf("q = %v, want production", params.Q)
	}
	if params.WorkspaceToken == nil || *params.WorkspaceToken != "wrkspc_test" {
		t.Fatalf("workspace_token = %v, want wrkspc_test", params.WorkspaceToken)
	}
}
