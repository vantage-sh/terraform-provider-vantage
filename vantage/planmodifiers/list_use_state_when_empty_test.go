package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListUseStateWhenEmpty(t *testing.T) {
	ctx := context.Background()

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	filledList := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("#costs")})
	unknownPlan := types.ListUnknown(types.StringType)

	tests := []struct {
		name        string
		configValue types.List
		stateValue  types.List
		wantPlan    types.List
	}{
		{
			// The configuration wins whenever it sets a value.
			name:        "config value passes through",
			configValue: filledList,
			stateValue:  emptyList,
			wantPlan:    unknownPlan,
		},
		{
			// Nothing was ever set, so the empty list stays and the plan is quiet.
			name:        "empty state is preserved",
			configValue: types.ListNull(types.StringType),
			stateValue:  emptyList,
			wantPlan:    emptyList,
		},
		{
			// The value left the configuration, so Update must clear it.
			name:        "non-empty state stays unknown so update clears it",
			configValue: types.ListNull(types.StringType),
			stateValue:  filledList,
			wantPlan:    unknownPlan,
		},
		{
			// There is no prior value to fall back on.
			name:        "null state is left alone",
			configValue: types.ListNull(types.StringType),
			stateValue:  types.ListNull(types.StringType),
			wantPlan:    unknownPlan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &planmodifier.ListResponse{PlanValue: unknownPlan}
			ListUseStateWhenEmpty().PlanModifyList(ctx, planmodifier.ListRequest{
				ConfigValue: tt.configValue,
				StateValue:  tt.stateValue,
				PlanValue:   unknownPlan,
			}, resp)

			if !resp.PlanValue.Equal(tt.wantPlan) {
				t.Fatalf("plan value = %v, want %v", resp.PlanValue, tt.wantPlan)
			}
		})
	}
}
