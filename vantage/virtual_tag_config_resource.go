package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	tagsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/virtual_tags"
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
			"values": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"filter": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "The filter VQL for the Value.",
						},
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The name of the Value.",
						},
					},
				},
			},
		},
		MarkdownDescription: "Manages a Virtual Tag Config.",
	}
}

func (r VirtualTagConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *VirtualTagConfigResourceModel
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
	var state *VirtualTagConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.NewGetVirtualTagConfigParams().WithVirtualTagConfigToken(state.Token.ValueString())
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

func (r VirtualTagConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *VirtualTagConfigResourceModel
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
		WithVirtualTagConfigToken(data.Token.ValueString()).
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
	var state *VirtualTagConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := tagsv2.NewDeleteVirtualTagConfigParams()
	params.SetVirtualTagConfigToken(state.Token.ValueString())
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
