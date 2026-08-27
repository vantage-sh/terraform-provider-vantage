package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NullableString returns a modifier for Optional+Computed strings whose APIs
// use an empty Terraform string to request a JSON null clear.
func NullableString() planmodifier.String {
	return nullableStringModifier{}
}

type nullableStringModifier struct{}

func (m nullableStringModifier) Description(_ context.Context) string {
	return "Sets the value to empty when removed from configuration, allowing the API to clear the field."
}

func (m nullableStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m nullableStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() && !req.StateValue.IsNull() && req.StateValue.ValueString() != "" {
		resp.PlanValue = types.StringValue("")
	}
}
