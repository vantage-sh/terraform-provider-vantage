package vantage

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config_value"
	tagsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/virtual_tags"
)

var (
	_ resource.Resource                = (*VirtualTagConfigValueResource)(nil)
	_ resource.ResourceWithConfigure   = (*VirtualTagConfigValueResource)(nil)
	_ resource.ResourceWithImportState = (*VirtualTagConfigValueResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*VirtualTagConfigValueResource)(nil)
)

type VirtualTagConfigValueResource struct {
	client *Client
}

type virtualTagConfigValueResourceModel struct {
	virtualTagConfigValueModel
	Id                    types.String `tfsdk:"id"`
	VirtualTagConfigToken types.String `tfsdk:"virtual_tag_config_token"`
}

func NewVirtualTagConfigValueResource() resource.Resource {
	return &VirtualTagConfigValueResource{}
}

func (r *VirtualTagConfigValueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_tag_config_value"
}

func (r *VirtualTagConfigValueResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_virtual_tag_config_value.VirtualTagConfigValueResourceSchema(ctx)
	resp.Schema.MarkdownDescription = "Manages an individual virtual tag config value.\n\n" +
		"Provide one allocation mode: `name`, `business_metric_token`, `cost_metric`, or a non-empty `percentages` list.\n\n" +
		"~> **Value ownership:** Manage values either with separate `vantage_virtual_tag_config_value` resources OR inline with the parent `vantage_virtual_tag_config.values` attribute, never both for the same config. When using this resource, omit `values` from the parent; parent updates will not modify or remove these values. No `ignore_changes` setting is needed.\n\n" +
		"Omitted optional attributes retain their API values. To clear a stored list, set it to `[]`; to clear label filters, set `label_filters = {}`. These changes replace only this value because the value update endpoint cannot clear them. Changing the parent config token also requires replacement."
	labelFilters := resp.Schema.Attributes["label_filters"].(schema.MapAttribute)
	labelFilters.MarkdownDescription = "Business metric row filters. Each key must match, and values within a key are alternatives. Cannot include the fan-out `label_key`. Setting an existing non-empty map to `{}` replaces this value to clear the filters."
	resp.Schema.Attributes["label_filters"] = labelFilters
	resp.Schema.Attributes["virtual_tag_config_token"] = schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "The token of the parent virtual tag config. Changing this forces replacement.",
		PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	for _, name := range []string{"token", "id"} {
		attribute := resp.Schema.Attributes[name].(schema.StringAttribute)
		attribute.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
		resp.Schema.Attributes[name] = attribute
	}
	metric := resp.Schema.Attributes["cost_metric"].(schema.SingleNestedAttribute)
	metric.CustomType = resource_virtual_tag_config.CostMetricType{ObjectType: types.ObjectType{
		AttrTypes: resource_virtual_tag_config.CostMetricValue{}.AttributeTypes(ctx),
	}}
	aggregation := metric.Attributes["aggregation"].(schema.SingleNestedAttribute)
	aggregation.CustomType = resource_virtual_tag_config.AggregationType{ObjectType: types.ObjectType{
		AttrTypes: resource_virtual_tag_config.AggregationValue{}.AttributeTypes(ctx),
	}}
	metric.Attributes["aggregation"] = aggregation
	resp.Schema.Attributes["cost_metric"] = metric
	for name, customType := range map[string]basetypes.ObjectTypable{
		"date_ranges":      resource_virtual_tag_config.DateRangesValue{}.Type(ctx).(basetypes.ObjectTypable),
		"label_transforms": resource_virtual_tag_config.LabelTransformsValue{}.Type(ctx).(basetypes.ObjectTypable),
		"percentages":      resource_virtual_tag_config.PercentagesValue{}.Type(ctx).(basetypes.ObjectTypable),
	} {
		attribute := resp.Schema.Attributes[name].(schema.ListNestedAttribute)
		attribute.NestedObject.CustomType = customType
		resp.Schema.Attributes[name] = attribute
	}
}

func (r *VirtualTagConfigValueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *Client, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *VirtualTagConfigValueResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	var plan, state virtualTagConfigValueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for name, values := range map[string]struct{ plan, state types.List }{
		"date_ranges":      {plan.DateRanges, state.DateRanges},
		"label_transforms": {plan.LabelTransforms, state.LabelTransforms},
		"label_values":     {plan.LabelValues, state.LabelValues},
		"percentages":      {plan.Percentages, state.Percentages},
	} {
		if !values.plan.IsUnknown() && clearsList(values.plan, values.state) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root(name))
		}
	}
	if !plan.LabelFilters.IsUnknown() && clearsMap(plan.LabelFilters, state.LabelFilters) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("label_filters"))
	}
	for name, values := range map[string]struct{ plan, state types.String }{
		"name":                  {plan.Name, state.Name},
		"business_metric_token": {plan.BusinessMetricToken, state.BusinessMetricToken},
		"label_key":             {plan.LabelKey, state.LabelKey},
		"filter":                {plan.Filter, state.Filter},
	} {
		if !values.plan.IsUnknown() && values.plan.ValueString() == "" && values.state.ValueString() != "" {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root(name))
		}
	}
	if !plan.CostMetric.IsUnknown() && plan.CostMetric.IsNull() && !state.CostMetric.IsNull() {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("cost_metric"))
	}
}

func (r *VirtualTagConfigValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data virtualTagConfigValueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload := data.toCreateValue(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	params := tagsv2.NewCreateVirtualTagConfigValueParams().
		WithVirtualTagConfigToken(data.VirtualTagConfigToken.ValueString()).
		WithCreateVirtualTagConfigValue(payload)
	out, err := r.client.V2.VirtualTags.CreateVirtualTagConfigValue(params, r.client.Auth)
	if err != nil {
		handleError("Create Virtual Tag Config Value", &resp.Diagnostics, err)
		return
	}
	resp.Diagnostics.Append(data.applyValuePayload(ctx, out.Payload)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *VirtualTagConfigValueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data virtualTagConfigValueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := tagsv2.NewGetVirtualTagConfigValueParams().
		WithVirtualTagConfigToken(data.VirtualTagConfigToken.ValueString()).
		WithVirtualTagConfigValueToken(data.Token.ValueString())
	out, err := r.client.V2.VirtualTags.GetVirtualTagConfigValue(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*tagsv2.GetVirtualTagConfigValueNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Read Virtual Tag Config Value", &resp.Diagnostics, err)
		return
	}
	resp.Diagnostics.Append(data.applyValuePayload(ctx, out.Payload)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *VirtualTagConfigValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state virtualTagConfigValueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.fillUnknownsFrom(&state.virtualTagConfigValueModel)
	payload := data.toUpdateValue(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	params := tagsv2.NewUpdateVirtualTagConfigValueParams().
		WithVirtualTagConfigToken(data.VirtualTagConfigToken.ValueString()).
		WithVirtualTagConfigValueToken(state.Token.ValueString()).
		WithUpdateVirtualTagConfigValue(payload)
	out, err := r.client.V2.VirtualTags.UpdateVirtualTagConfigValue(params, r.client.Auth)
	if err != nil {
		handleError("Update Virtual Tag Config Value", &resp.Diagnostics, err)
		return
	}
	resp.Diagnostics.Append(data.applyValuePayload(ctx, out.Payload)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *VirtualTagConfigValueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data virtualTagConfigValueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := tagsv2.NewDeleteVirtualTagConfigValueParams().
		WithVirtualTagConfigToken(data.VirtualTagConfigToken.ValueString()).
		WithVirtualTagConfigValueToken(data.Token.ValueString())
	_, err := r.client.V2.VirtualTags.DeleteVirtualTagConfigValue(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*tagsv2.DeleteVirtualTagConfigValueNotFound); ok {
			return
		}
		handleError("Delete Virtual Tag Config Value", &resp.Diagnostics, err)
	}
}

func (r *VirtualTagConfigValueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parent, token, ok := strings.Cut(req.ID, "/")
	if !ok || parent == "" || token == "" || strings.Contains(token, "/") {
		resp.Diagnostics.AddError("Invalid Import Identifier", "Expected virtual_tag_config_token/virtual_tag_config_value_token.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("virtual_tag_config_token"), parent)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("token"), token)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), token)...)
}
