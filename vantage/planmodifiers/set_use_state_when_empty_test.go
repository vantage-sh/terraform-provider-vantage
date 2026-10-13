package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUseStateWhenEmpty(t *testing.T) {
	t.Parallel()

	stateOne, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"wrkspc_abc"})
	if diags.HasError() {
		t.Fatalf("building state set: %v", diags)
	}
	configOne, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"wrkspc_xyz"})
	if diags.HasError() {
		t.Fatalf("building config set: %v", diags)
	}
	emptySet, diags := types.SetValueFrom(context.Background(), types.StringType, []string{})
	if diags.HasError() {
		t.Fatalf("building empty set: %v", diags)
	}

	tests := map[string]struct {
		config types.Set
		state  types.Set
		plan   types.Set
		want   types.Set
	}{
		"omitted config keeps non-empty API-assigned state": {
			config: types.SetNull(types.StringType),
			state:  stateOne,
			plan:   types.SetUnknown(types.StringType),
			want:   stateOne,
		},
		"explicit config replaces non-empty state": {
			config: configOne,
			state:  stateOne,
			plan:   configOne,
			want:   configOne,
		},
		"omitted config keeps empty state": {
			config: types.SetNull(types.StringType),
			state:  emptySet,
			plan:   types.SetUnknown(types.StringType),
			want:   emptySet,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &planmodifier.SetResponse{PlanValue: test.plan}
			UseStateWhenEmpty().PlanModifySet(context.Background(), planmodifier.SetRequest{
				ConfigValue: test.config,
				StateValue:  test.state,
				PlanValue:   test.plan,
			}, response)

			if !response.PlanValue.Equal(test.want) {
				t.Errorf("plan = %v, want %v", response.PlanValue, test.want)
			}
		})
	}
}
