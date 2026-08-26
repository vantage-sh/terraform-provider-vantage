package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ListUseStateUnlessSiblingsChange returns a planmodifier.List for an
// Optional+Computed list whose value the API derives from sibling attributes.
//
// Plain UseStateForUnknown is wrong for such an attribute: when a sibling
// changes, the API recomputes this list, so keeping the prior value in the plan
// makes the applied result differ from the plan. Leaving it unknown on every
// update is also wrong, because the payload would then send a null that the API
// reads as an instruction to clear the value.
//
// This modifier keeps the prior state only while every listed sibling is
// unchanged, and otherwise leaves the value unknown so the API can recompute it.
func ListUseStateUnlessSiblingsChange(siblings ...path.Path) planmodifier.List {
	return listUseStateUnlessSiblingsChangeModifier{siblings: siblings}
}

type listUseStateUnlessSiblingsChangeModifier struct {
	siblings []path.Path
}

func (m listUseStateUnlessSiblingsChangeModifier) Description(_ context.Context) string {
	return "Preserves prior state while the attributes it is derived from are unchanged."
}

func (m listUseStateUnlessSiblingsChangeModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m listUseStateUnlessSiblingsChangeModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Create and destroy have no prior state worth preserving.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	// Config has an explicit value — let it through unchanged.
	if !req.ConfigValue.IsNull() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if !resp.PlanValue.IsUnknown() {
		return
	}

	for _, sibling := range m.siblings {
		var config, state types.List

		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, sibling, &config)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, sibling, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// A sibling absent from config is itself computed, so it cannot be the
		// source of a change.
		if config.IsNull() {
			continue
		}
		// The sibling is changing, so the API will derive a new value here.
		// Leave the plan unknown rather than promising the stale one.
		if !config.Equal(state) {
			return
		}
	}

	resp.PlanValue = req.StateValue
}
