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
				"filter":                generatedValuesAttrs["filter"],
				"name":                  generatedValuesAttrs["name"],
				"percentages":           generatedValuesAttrs["percentages"],
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
}

func (r VirtualTagConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *virtualTagConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.NewCreateVirtualTagConfigParams().
		WithContext(ctx).
		WithCreateVirtualTagConfig(model)
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r VirtualTagConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *virtualTagConfigModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.NewGetVirtualTagConfigParams().
		WithContext(ctx).
		WithToken(state.Token.ValueString())
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

	model := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.
		NewUpdateVirtualTagConfigParams().
		WithContext(ctx).
		WithToken(data.Token.ValueString()).
		WithUpdateVirtualTagConfig(model)

	out, err := r.client.V2.VirtualTags.UpdateVirtualTagConfig(params, r.client.Auth)
	if err != nil {
		handleError("Update Virtual Tag Config Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r VirtualTagConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *virtualTagConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.NewDeleteVirtualTagConfigParams().
		WithContext(ctx)
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
