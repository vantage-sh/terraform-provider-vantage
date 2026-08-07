package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// derivedFromSchema mirrors the shape the budget alert uses: a computed value
// the API derives from a list the practitioner sets, plus an unrelated field.
func derivedFromSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"budget_tokens": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"threshold": schema.Int64Attribute{
				Required: true,
			},
			"workspace_token": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func derivedFromObjectType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"budget_tokens":   tftypes.List{ElementType: tftypes.String},
			"threshold":       tftypes.Number,
			"workspace_token": tftypes.String,
		},
	}
}

// derivedFromValue builds one object. A nil budgetTokens slice means unknown.
func derivedFromValue(budgetTokens []string, threshold int64, workspaceToken tftypes.Value) tftypes.Value {
	tokens := tftypes.Value{}
	if budgetTokens == nil {
		tokens = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue)
	} else {
		elements := make([]tftypes.Value, 0, len(budgetTokens))
		for _, token := range budgetTokens {
			elements = append(elements, tftypes.NewValue(tftypes.String, token))
		}
		tokens = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, elements)
	}

	return tftypes.NewValue(derivedFromObjectType(), map[string]tftypes.Value{
		"budget_tokens":   tokens,
		"threshold":       tftypes.NewValue(tftypes.Number, threshold),
		"workspace_token": workspaceToken,
	})
}

func TestStringDerivedFrom(t *testing.T) {
	ctx := context.Background()
	priorWorkspace := tftypes.NewValue(tftypes.String, "wrkspc_1")
	unknownWorkspace := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)

	tests := []struct {
		name        string
		planTokens  []string
		planLevel   int64
		stateTokens []string
		wantPlan    types.String
	}{
		{
			// The reported case: only threshold moves, so the derived value holds.
			name:        "source unchanged keeps the prior value",
			planTokens:  []string{"bdgt_1"},
			planLevel:   90,
			stateTokens: []string{"bdgt_1"},
			wantPlan:    types.StringValue("wrkspc_1"),
		},
		{
			// The budgets moved, so the workspace may move with them.
			name:        "changed source stays unknown",
			planTokens:  []string{"bdgt_2"},
			planLevel:   100,
			stateTokens: []string{"bdgt_1"},
			wantPlan:    types.StringUnknown(),
		},
		{
			name:        "added source element stays unknown",
			planTokens:  []string{"bdgt_1", "bdgt_2"},
			planLevel:   100,
			stateTokens: []string{"bdgt_1"},
			wantPlan:    types.StringUnknown(),
		},
		{
			// The source is itself computed this round, so nothing can be promised.
			name:        "unknown source stays unknown",
			planTokens:  nil,
			planLevel:   100,
			stateTokens: []string{"bdgt_1"},
			wantPlan:    types.StringUnknown(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := derivedFromSchema()
			resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

			StringDerivedFrom(path.Root("budget_tokens")).PlanModifyString(ctx, planmodifier.StringRequest{
				Path:       path.Root("workspace_token"),
				StateValue: types.StringValue("wrkspc_1"),
				PlanValue:  types.StringUnknown(),
				Plan: tfsdk.Plan{
					Schema: s,
					Raw:    derivedFromValue(tt.planTokens, tt.planLevel, unknownWorkspace),
				},
				State: tfsdk.State{
					Schema: s,
					Raw:    derivedFromValue(tt.stateTokens, 100, priorWorkspace),
				},
			}, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}
			if !resp.PlanValue.Equal(tt.wantPlan) {
				t.Fatalf("plan value = %v, want %v", resp.PlanValue, tt.wantPlan)
			}
		})
	}
}

// TestStringDerivedFromOnCreate checks that a create, where there is no prior
// value, leaves the attribute unknown.
func TestStringDerivedFromOnCreate(t *testing.T) {
	ctx := context.Background()
	s := derivedFromSchema()
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	StringDerivedFrom(path.Root("budget_tokens")).PlanModifyString(ctx, planmodifier.StringRequest{
		Path:       path.Root("workspace_token"),
		StateValue: types.StringNull(),
		PlanValue:  types.StringUnknown(),
		Plan: tfsdk.Plan{
			Schema: s,
			Raw:    derivedFromValue([]string{"bdgt_1"}, 100, tftypes.NewValue(tftypes.String, tftypes.UnknownValue)),
		},
		State: tfsdk.State{
			Schema: s,
			Raw:    tftypes.NewValue(derivedFromObjectType(), nil),
		},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.PlanValue.IsUnknown() {
		t.Fatalf("plan value = %v, want unknown", resp.PlanValue)
	}
}
