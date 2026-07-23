// Package planmodifiers provides custom Terraform plan modifiers.
package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ImmutableAfterCreate returns a plan modifier that prevents a string
// attribute from being changed after the resource has been created.
// On create (no prior state) any value is accepted. On update, if the
// configured value differs from the current state value a warning is
// emitted and the plan is reverted to the state value, leaving the
// resource unchanged.
func ImmutableAfterCreate(fieldName string) planmodifier.String {
	return immutableAfterCreate{fieldName: fieldName}
}

type immutableAfterCreate struct{ fieldName string }

func (m immutableAfterCreate) Description(_ context.Context) string {
	return fmt.Sprintf("The %q field cannot be changed after the resource has been created.", m.fieldName)
}

func (m immutableAfterCreate) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutableAfterCreate) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Allow any value on create (state is null/unknown before first apply).
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// No change — nothing to do.
	if req.ConfigValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddAttributeWarning(
		req.Path,
		fmt.Sprintf("Cannot update %q", m.fieldName),
		fmt.Sprintf(
			"The %q field cannot be changed after the resource has been created. "+
				"The existing value %q will be preserved. "+
				"To use a different value, destroy and recreate the resource.",
			m.fieldName,
			req.StateValue.ValueString(),
		),
	)

	// Revert the planned value back to the current state value so no update
	// is attempted and no perpetual drift appears in subsequent plans.
	resp.PlanValue = req.StateValue
}
