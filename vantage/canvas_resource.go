package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_canvases"
	canvasesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/canvases"
)

var (
	_ resource.Resource                = (*canvasResource)(nil)
	_ resource.ResourceWithConfigure   = (*canvasResource)(nil)
	_ resource.ResourceWithImportState = (*canvasResource)(nil)
)

type canvasResource struct {
	client *Client
}

func NewCanvasResource() resource.Resource {
	return &canvasResource{}
}

func (r *canvasResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r *canvasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_canvas"
}

func (r *canvasResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_canvases.CanvasesResourceSchema(ctx)
	attrs := s.GetAttributes()

	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["workspace_token"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: attrs["workspace_token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["data"] = schema.SingleNestedAttribute{
		Computed:            true,
		MarkdownDescription: "The structured table data of the Canvas.",
		Attributes: map[string]schema.Attribute{
			"error": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Error message if the refresh workflow failed.",
			},
			"table": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON-encoded tabular data produced by the Canvas refresh workflow.",
			},
		},
	}

	s.MarkdownDescription = "Manages a Canvas."

	resp.Schema = s
}

func (r *canvasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *canvasModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := canvasesv2.NewCreateCanvasParams().WithCreateCanvas(data.toCreate())
	out, err := r.client.V2.Canvases.CreateCanvas(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*canvasesv2.CreateCanvasBadRequest); ok {
			handleBadRequest("Create Canvas", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Canvas", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *canvasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *canvasModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := canvasesv2.NewGetCanvasParams().WithCanvasToken(data.Token.ValueString())
	out, err := r.client.V2.Canvases.GetCanvas(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*canvasesv2.GetCanvasNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Read Canvas", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *canvasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *canvasModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := canvasesv2.NewUpdateCanvasParams().
		WithCanvasToken(data.Token.ValueString()).
		WithUpdateCanvas(data.toUpdate())

	out, err := r.client.V2.Canvases.UpdateCanvas(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*canvasesv2.UpdateCanvasBadRequest); ok {
			handleBadRequest("Update Canvas", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update Canvas", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(data.applyPayload(ctx, out.Payload)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *canvasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *canvasModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := canvasesv2.NewDeleteCanvasParams().WithCanvasToken(data.Token.ValueString())
	_, err := r.client.V2.Canvases.DeleteCanvas(params, r.client.Auth)
	if err != nil {
		handleError("Delete Canvas", &resp.Diagnostics, err)
	}
}

func (r *canvasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}
