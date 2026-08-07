package planmodifiers

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// StringDerivedFrom returns a planmodifier.String for a computed attribute that
// the API derives from other attributes of the same resource.
//
// The prior value stands while every source attribute holds still, so a change
// elsewhere in the resource no longer plans this attribute as "known after
// apply". When a source attribute does change, the value is left unknown and
// the API decides it.
//
// This is safer than UseStateForUnknown for a derived value. UseStateForUnknown
// promises the value never changes, and Terraform fails the apply with
// "Provider produced inconsistent result after apply" when that promise breaks.
func StringDerivedFrom(sources ...path.Path) planmodifier.String {
	return stringDerivedFrom{sources: sources}
}

type stringDerivedFrom struct{ sources []path.Path }

func (m stringDerivedFrom) Description(_ context.Context) string {
	names := make([]string, 0, len(m.sources))
	for _, source := range m.sources {
		names = append(names, source.String())
	}
	return "Keeps the prior value while " + strings.Join(names, " and ") + " stays unchanged."
}

func (m stringDerivedFrom) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m stringDerivedFrom) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// No prior value to reuse. This is a create.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	for _, source := range m.sources {
		var planValue, stateValue attr.Value

		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, source, &planValue)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, source, &stateValue)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// A source is still being decided, so this value cannot be predicted.
		if planValue.IsUnknown() {
			return
		}
		// A source changed, so the API may return a different value.
		if !planValue.Equal(stateValue) {
			return
		}
	}

	resp.PlanValue = req.StateValue
}
