package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateWhenEmpty returns a planmodifier.Set that, when the attribute is
// absent from config (null), preserves the prior state value only if that
// state value is an empty set. If the state value is non-empty it is left
// alone so the plan reflects a change, allowing the Update method to clear
// the field by calling the API with an empty collection.
//
// This gives Optional+Computed set fields the following behaviour:
//   - User never sets the field → state stays empty, no perpetual drift.
//   - User sets the field and later removes it from config → plan detects a
//     change and the provider clears the value via the API.
//   - User sets the field to an explicit value → config value is used as-is.
func UseStateWhenEmpty() planmodifier.Set {
	return useStateWhenEmptyModifier{}
}

type useStateWhenEmptyModifier struct{}

func (m useStateWhenEmptyModifier) Description(_ context.Context) string {
	return "Preserves empty state when unset in config; allows non-empty state to be cleared when removed from config."
}

func (m useStateWhenEmptyModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateWhenEmptyModifier) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Config has an explicit value — let it through unchanged.
	if !req.ConfigValue.IsNull() {
		return
	}
	// No prior state to reason about.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	// Prior state is a known, non-empty set. Leave the plan value alone so
	// Terraform detects the change and calls Update, which will clear the field.
	if len(req.StateValue.Elements()) > 0 {
		return
	}
	// Prior state is an empty set and config is null — preserve it to avoid
	// perpetual "known after apply" on resources that never set this field.
	resp.PlanValue = req.StateValue
}
