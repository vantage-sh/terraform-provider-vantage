package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateWhenEmpty returns a planmodifier.Set that, when the attribute is
// absent from config (null), preserves the prior state value — including a
// non-empty set the API may have assigned. An explicit config value is planned
// as-is.
//
// This gives Optional+Computed set fields the following behaviour:
//   - User never sets the field → prior state (empty or API-assigned) is kept,
//     so refresh/import of an auto-assigned workspace does not cause perpetual
//     drift.
//   - User sets the field to an explicit value → config value is used as-is.
func UseStateWhenEmpty() planmodifier.Set {
	return useStateWhenEmptyModifier{}
}

type useStateWhenEmptyModifier struct{}

func (m useStateWhenEmptyModifier) Description(_ context.Context) string {
	return "Preserves prior state when unset in config, including non-empty API-assigned sets."
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
	// Config omitted — preserve prior state (empty or non-empty) so an
	// API-assigned workspace does not show up as perpetual drift.
	resp.PlanValue = req.StateValue
}
