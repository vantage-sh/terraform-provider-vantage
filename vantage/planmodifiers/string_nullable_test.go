package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNullableString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config types.String
		state  types.String
		plan   types.String
		want   types.String
	}{
		"removed configured value clears": {
			config: types.StringNull(),
			state:  types.StringValue("value"),
			plan:   types.StringUnknown(),
			want:   types.StringValue(""),
		},
		"explicit value remains": {
			config: types.StringValue("new"),
			state:  types.StringValue("old"),
			plan:   types.StringValue("new"),
			want:   types.StringValue("new"),
		},
		"already empty remains empty": {
			config: types.StringNull(),
			state:  types.StringValue(""),
			plan:   types.StringValue(""),
			want:   types.StringValue(""),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &planmodifier.StringResponse{PlanValue: test.plan}
			NullableString().PlanModifyString(context.Background(), planmodifier.StringRequest{
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
