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

var testListSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"derived": schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true},
		"sibling": schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true},
	},
}

func listOf(values ...string) tftypes.Value {
	if values == nil {
		return tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil)
	}

	elements := make([]tftypes.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, tftypes.NewValue(tftypes.String, v))
	}
	return tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, elements)
}

func objectOf(derived, sibling tftypes.Value) tftypes.Value {
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"derived": tftypes.List{ElementType: tftypes.String},
		"sibling": tftypes.List{ElementType: tftypes.String},
	}}
	return tftypes.NewValue(objType, map[string]tftypes.Value{
		"derived": derived,
		"sibling": sibling,
	})
}

func runModifier(t *testing.T, configSibling, stateSibling tftypes.Value, stateDerived []string) types.List {
	t.Helper()

	ctx := context.Background()
	state, diags := types.ListValueFrom(ctx, types.StringType, stateDerived)
	if diags.HasError() {
		t.Fatalf("building state list: %v", diags)
	}

	req := planmodifier.ListRequest{
		Path:        path.Root("derived"),
		ConfigValue: types.ListNull(types.StringType),
		StateValue:  state,
		Config:      tfsdk.Config{Schema: testListSchema, Raw: objectOf(listOf(nil...), configSibling)},
		State:       tfsdk.State{Schema: testListSchema, Raw: objectOf(listOf(stateDerived...), stateSibling)},
		Plan:        tfsdk.Plan{Schema: testListSchema, Raw: objectOf(listOf(nil...), configSibling)},
	}
	resp := &planmodifier.ListResponse{PlanValue: types.ListUnknown(types.StringType)}

	ListUseStateUnlessSiblingsChange(path.Root("sibling")).PlanModifyList(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("modifier reported errors: %v", resp.Diagnostics)
	}
	return resp.PlanValue
}

// An unrelated update must not lose the derived value, since the API would read
// the resulting null as an instruction to clear it.
func TestListUseStateUnlessSiblingsChange_keepsStateWhenSiblingUnchanged(t *testing.T) {
	t.Parallel()

	plan := runModifier(t, listOf("a"), listOf("a"), []string{"derived-value"})

	if plan.IsUnknown() {
		t.Fatal("plan is unknown, want the prior state value")
	}
	if len(plan.Elements()) != 1 {
		t.Errorf("plan has %d elements, want 1", len(plan.Elements()))
	}
}

// When a sibling changes the API recomputes this list, so promising the stale
// value would produce "inconsistent result after apply".
func TestListUseStateUnlessSiblingsChange_yieldsWhenSiblingChanges(t *testing.T) {
	t.Parallel()

	plan := runModifier(t, listOf("a", "b"), listOf("a"), []string{"derived-value"})

	if !plan.IsUnknown() {
		t.Errorf("plan is %v, want unknown so the API can recompute it", plan)
	}
}

// A sibling that is absent from config is itself computed, so it cannot be the
// source of a change.
func TestListUseStateUnlessSiblingsChange_ignoresSiblingAbsentFromConfig(t *testing.T) {
	t.Parallel()

	plan := runModifier(t, listOf(nil...), listOf("a"), []string{"derived-value"})

	if plan.IsUnknown() {
		t.Error("plan is unknown, want the prior state value")
	}
}
