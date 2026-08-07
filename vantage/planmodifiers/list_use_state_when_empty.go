package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ListUseStateWhenEmpty returns a planmodifier.List with the behaviour of
// UseStateWhenEmpty, for a list instead of a set.
func ListUseStateWhenEmpty() planmodifier.List {
	return listUseStateWhenEmptyModifier{}
}

type listUseStateWhenEmptyModifier struct{}

func (m listUseStateWhenEmptyModifier) Description(_ context.Context) string {
	return "Preserves empty state when unset in config; allows non-empty state to be cleared when removed from config."
}

func (m listUseStateWhenEmptyModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m listUseStateWhenEmptyModifier) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Config has an explicit value — let it through unchanged.
	if !req.ConfigValue.IsNull() {
		return
	}
	// No prior state to reason about.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	// Prior state is a known, non-empty list. Leave the plan value alone so
	// Terraform detects the change and calls Update, which will clear the field.
	if len(req.StateValue.Elements()) > 0 {
		return
	}
	// Prior state is an empty list and config is null — preserve it to avoid
	// perpetual "known after apply" on resources that never set this field.
	resp.PlanValue = req.StateValue
}
