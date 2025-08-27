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
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The key of the VirtualTagConfig.",
			},
			"backfill_until": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The earliest month VirtualTagConfig should be backfilled to.",
			},
			"overridable": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether the VirtualTagConfig can override a provider-supplied tag on a matching Cost.",
			},
			"created_by_token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The token of the User who created the VirtualTagConfig.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The token of the VirtualTagConfig.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The id of the VirtualTagConfig.",
				// PlanModifiers: []planmodifier.String{
				// 	stringplanmodifier.UseStateForUnknown(),
				// },
			},
			"values": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"business_metric_token": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The token of the associated BusinessMetric.",
							MarkdownDescription: "The token of the associated BusinessMetric.",
							// TODO
							// Validators: []validator.String{
							// 	// Validate only this attribute, cost_metric, or name is configured.
							// 	stringvalidator.ExactlyOneOf(path.Expressions{
							// 		path.MatchRelative().AtParent().AtName("cost_metric"),
							// 		path.MatchRelative().AtParent().AtName("name"),
							// 	}...),
							// },
						},
						"cost_metric": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"aggregation": schema.SingleNestedAttribute{
									Attributes: map[string]schema.Attribute{
										"tag": schema.StringAttribute{
											Optional:            true,
											Description:         "The tag to aggregate on.",
											MarkdownDescription: "The tag to aggregate on.",
										},
									},
									CustomType: resource_virtual_tag_config.AggregationType{
										ObjectType: types.ObjectType{
											AttrTypes: resource_virtual_tag_config.AggregationValue{}.AttributeTypes(ctx),
										},
									},
									Optional: true,
								},
								"filter": schema.StringAttribute{
									Optional:            true,
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
							// Validators: []validator.Object{
							// 	// Validate only this attribute, business_metric_token, or name is configured.
							// 	objectvalidator.ExactlyOneOf(path.Expressions{
							// 		path.MatchRelative().AtParent().AtName("business_metric_token"),
							// 		path.MatchRelative().AtParent().AtName("name"),
							// 	}...),
							// },
						},
						"filter": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The filter VQL for the Value.",
						},
						"name": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "The name of the Value.",
							// Validators: []validator.String{
							// 	// Validate only this attribute, business_metric_token, or cost_metric is configured.
							// 	stringvalidator.ExactlyOneOf(path.Expressions{
							// 		path.MatchRelative().AtParent().AtName("business_metric_token"),
							// 		path.MatchRelative().AtParent().AtName("cost_metric"),
							// 	}...),
							// },
						},
					},
				},
			},
		},
		MarkdownDescription: "Manages a Virtual Tag Config.",
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	model := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.
		NewUpdateVirtualTagConfigParams().
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
