package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	accounttagsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/tags"
	tagsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/virtual_tags"
)

var (
	_ resource.Resource                = (*VirtualTagConfigResource)(nil)
	_ resource.ResourceWithConfigure   = (*VirtualTagConfigResource)(nil)
	_ resource.ResourceWithImportState = (*VirtualTagConfigResource)(nil)
)

type VirtualTagConfigResource struct {
	client *Client
}

func NewVirtualTagConfigResource() resource.Resource {
	return &VirtualTagConfigResource{}
}

func (r *VirtualTagConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_tag_config"
}

func (r VirtualTagConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Because we generate our schema from a Swagger/OpenAPI v2 spec, we're unable to express some of the constraints we want to enforce.
	// A major one is that name, business_metric_token, cost_metric, and percentages are all mutually exclusive,
	// and one must be provided.
	//
	// Because our swagger spec is translated without that, we run into problems when we have nested attributes marked as Required.
	//
	// Here we modify the generated schema to make the nested attributes Optional instead of Required.
	resp.Schema = resource_virtual_tag_config.VirtualTagConfigResourceSchema(ctx)

	resp.Schema.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		Description:         "The token of the VirtualTagConfig.",
		MarkdownDescription: "The token of the VirtualTagConfig.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	generatedValues := resp.Schema.Attributes["values"].(schema.ListNestedAttribute)
	generatedValuesAttrs := generatedValues.NestedObject.Attributes

	resp.Schema.Attributes["values"] = schema.ListNestedAttribute{
		Optional:            generatedValues.Optional,
		Computed:            generatedValues.Computed,
		Description:         generatedValues.Description,
		MarkdownDescription: generatedValues.MarkdownDescription,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				// Reuse generated attributes unchanged
				"business_metric_token": generatedValuesAttrs["business_metric_token"],
				"date_ranges":           generatedValuesAttrs["date_ranges"],
				"display_name":          generatedValuesAttrs["display_name"],
				"filter":                generatedValuesAttrs["filter"],
				"label_key":             generatedValuesAttrs["label_key"],
				"label_transforms":      generatedValuesAttrs["label_transforms"],
				"label_values":          generatedValuesAttrs["label_values"],
				"name":                  generatedValuesAttrs["name"],
				"percentages":           generatedValuesAttrs["percentages"],
				"token": schema.StringAttribute{
					// Optional+Computed so GET→config round-trips (e.g. copying
					// values from the data source) remain valid. Create/update
					// payloads intentionally omit token.
					Optional:            true,
					Computed:            true,
					Description:         "The token of the Value.",
					MarkdownDescription: "The token of the Value.",
				},
				// Override cost_metric: make aggregation, aggregation.tag, and filter Optional
				"cost_metric": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"aggregation": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"tag": schema.StringAttribute{
									Optional:            true, // Generated has Required
									Description:         "The tag to aggregate on.",
									MarkdownDescription: "The tag to aggregate on.",
								},
							},
							CustomType: resource_virtual_tag_config.AggregationType{
								ObjectType: types.ObjectType{
									AttrTypes: resource_virtual_tag_config.AggregationValue{}.AttributeTypes(ctx),
								},
							},
							Optional: true, // Generated has Required
						},
						"filter": schema.StringAttribute{
							Optional:            true, // Generated has Required
							Description:         "The filter VQL for the cost metric.",
							MarkdownDescription: "The filter VQL for the cost metric.",
						},
					},
					CustomType: resource_virtual_tag_config.CostMetricType{
						ObjectType: types.ObjectType{
							AttrTypes: resource_virtual_tag_config.CostMetricValue{}.AttributeTypes(ctx),
						},
					},
					Optional: true,
					Computed: true,
				},
			},
		},
	}

	resp.Schema.Attributes["preferred"] = schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "Whether this virtual tag key is marked as preferred in the Vantage UI.",
	}
}

func (r VirtualTagConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *virtualTagConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	preferredToSync := data.Preferred
	model := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.NewCreateVirtualTagConfigParams().WithCreateVirtualTagConfig(model)
	out, err := r.client.V2.VirtualTags.CreateVirtualTagConfig(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*tagsv2.CreateVirtualTagConfigBadRequest); ok {
			handleBadRequest("Create Virtual Tag Config Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create Virtual Tag Config Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	if !preferredToSync.IsNull() && !preferredToSync.IsUnknown() {
		data.Preferred = preferredToSync
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.syncPreferred(ctx, data.Key.ValueString(), preferredToSync); err != nil {
		handleError("Set Preferred Virtual Tag Config", &resp.Diagnostics, err)
		return
	}
}

func (r VirtualTagConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *virtualTagConfigModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.NewGetVirtualTagConfigParams().WithToken(state.Token.ValueString())
	out, err := r.client.V2.VirtualTags.GetVirtualTagConfig(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*tagsv2.GetVirtualTagConfigNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Virtual Tag Config Resource", &resp.Diagnostics, err)
		return
	}

	diag := state.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r VirtualTagConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r VirtualTagConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *virtualTagConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state *virtualTagConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyChanged := !plannedValueEqual(data.Key, state.Key)
	if (data.Preferred.IsNull() || data.Preferred.IsUnknown()) &&
		!state.Preferred.IsNull() &&
		!state.Preferred.IsUnknown() &&
		state.Preferred.ValueBool() {
		data.Preferred = types.BoolValue(false)
	}
	preferredChanged := !plannedValueEqual(data.Preferred, state.Preferred)
	preferredToSync := data.Preferred

	changes := virtualTagConfigValueChanges{requiresParentUpdate: data.Values.IsNull() || data.Values.IsUnknown()}
	if !changes.requiresParentUpdate {
		planValues := data.valuesFromTf(ctx, &resp.Diagnostics)
		stateValues := state.valuesFromTf(ctx, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		changes = diffVirtualTagConfigValues(planValues, stateValues)
	}

	if !data.parentFieldsEqual(state) || changes.requiresParentUpdate {
		model := data.toUpdate(ctx, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		params := tagsv2.
			NewUpdateVirtualTagConfigParams().
			WithToken(data.Token.ValueString()).
			WithUpdateVirtualTagConfig(model)
		out, _, err := r.client.V2.VirtualTags.UpdateVirtualTagConfig(params, r.client.Auth)
		if err != nil {
			handleError("Update Virtual Tag Config Resource", &resp.Diagnostics, err)
			return
		}
		resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !preferredToSync.IsNull() && !preferredToSync.IsUnknown() {
			data.Preferred = preferredToSync
		}
		if keyChanged && !state.Preferred.IsNull() && !state.Preferred.IsUnknown() && state.Preferred.ValueBool() {
			if err := r.syncPreferred(ctx, state.Key.ValueString(), types.BoolValue(false)); err != nil {
				resp.Diagnostics.AddWarning(
					"Unable to Clear Previous Preferred Virtual Tag Config",
					"The virtual tag config key was updated, but its previous preferred tag setting could not be cleared. The new key will still be synchronized.\n\nConnection Error: "+err.Error(),
				)
			}
		}
		if err := r.syncPreferred(ctx, data.Key.ValueString(), preferredToSync); err != nil {
			handleError("Update Preferred Virtual Tag Config", &resp.Diagnostics, err)
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if len(changes.creates) == 0 && len(changes.updates) == 0 && len(changes.deletes) == 0 {
		if preferredChanged {
			if err := r.syncPreferred(ctx, data.Key.ValueString(), preferredToSync); err != nil {
				handleError("Update Preferred Virtual Tag Config", &resp.Diagnostics, err)
				return
			}
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	updateModels := make([]*modelsv2.UpdateVirtualTagConfigValue, len(changes.updates))
	for i, value := range changes.updates {
		updateModels[i] = value.toUpdateValue(ctx, &resp.Diagnostics)
	}
	createModels := make([]*modelsv2.CreateVirtualTagConfigValue, len(changes.creates))
	for i, value := range changes.creates {
		createModels[i] = value.toCreateValue(ctx, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	refreshState := func() {
		params := tagsv2.NewGetVirtualTagConfigParams().WithToken(data.Token.ValueString())
		out, err := r.client.V2.VirtualTags.GetVirtualTagConfig(params, r.client.Auth)
		if err != nil {
			handleError("Refresh Virtual Tag Config Resource", &resp.Diagnostics, err)
			return
		}
		resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		}
	}
	fail := func(title string, err error) {
		data.Preferred = state.Preferred
		refreshState()
		handleError(title, &resp.Diagnostics, err)
	}

	for i, value := range changes.updates {
		params := tagsv2.
			NewUpdateVirtualTagConfigValueParams().
			WithVirtualTagConfigToken(data.Token.ValueString()).
			WithVirtualTagConfigValueToken(value.Token.ValueString()).
			WithUpdateVirtualTagConfigValue(updateModels[i])
		if _, err := r.client.V2.VirtualTags.UpdateVirtualTagConfigValue(params, r.client.Auth); err != nil {
			fail("Update Virtual Tag Config Value", err)
			return
		}
	}
	for _, value := range changes.deletes {
		params := tagsv2.
			NewDeleteVirtualTagConfigValueParams().
			WithVirtualTagConfigToken(data.Token.ValueString()).
			WithVirtualTagConfigValueToken(value.Token.ValueString())
		if _, err := r.client.V2.VirtualTags.DeleteVirtualTagConfigValue(params, r.client.Auth); err != nil {
			fail("Delete Virtual Tag Config Value", err)
			return
		}
	}
	for i := range changes.creates {
		params := tagsv2.
			NewCreateVirtualTagConfigValueParams().
			WithVirtualTagConfigToken(data.Token.ValueString()).
			WithCreateVirtualTagConfigValue(createModels[i])
		if _, err := r.client.V2.VirtualTags.CreateVirtualTagConfigValue(params, r.client.Auth); err != nil {
			fail("Create Virtual Tag Config Value", err)
			return
		}
	}

	if preferredChanged {
		if err := r.syncPreferred(ctx, data.Key.ValueString(), preferredToSync); err != nil {
			fail("Update Preferred Virtual Tag Config", err)
			return
		}
	}
	refreshState()
}

func (r VirtualTagConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *virtualTagConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.Preferred.IsNull() && !state.Preferred.IsUnknown() && state.Preferred.ValueBool() {
		if err := r.syncPreferred(ctx, state.Key.ValueString(), types.BoolValue(false)); err != nil {
			resp.Diagnostics.AddWarning(
				"Unable to Clear Preferred Virtual Tag Config",
				"The virtual tag config will still be deleted, but its preferred tag setting could not be cleared.\n\nConnection Error: "+err.Error(),
			)
		}
	}

	params := tagsv2.NewDeleteVirtualTagConfigParams()
	params.SetToken(state.Token.ValueString())
	_, err := r.client.V2.VirtualTags.DeleteVirtualTagConfig(params, r.client.Auth)
	if err != nil {
		handleError("Delete Virtual Tag Config Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *VirtualTagConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r VirtualTagConfigResource) syncPreferred(ctx context.Context, key string, preferred types.Bool) error {
	if preferred.IsNull() || preferred.IsUnknown() {
		return nil
	}

	// Use tag_keys rather than tag_key: UpdateTag's TagKeys field has no
	// omitempty, so TagKey-only payloads serialize as "tag_keys": null and
	// the API rejects them with exactly_one_of.
	params := accounttagsv2.NewUpdateTagParams().WithContext(ctx).WithUpdateTag(&modelsv2.UpdateTag{
		TagKeys:   []string{key},
		Preferred: preferred.ValueBoolPointer(),
	})
	_, err := r.client.V2.Tags.UpdateTag(params, r.client.Auth)
	return err
}
