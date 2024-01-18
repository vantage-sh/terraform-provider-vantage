package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	segmentsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/segments"
)

type SegmentResource struct {
	client *Client
}

func NewSegmentResource() resource.Resource {
	return &SegmentResource{}
}

type SegmentResourceModel struct {
	Title              types.String `tfsdk:"title"`
	Description        types.String `tfsdk:"description"`
	Priority           types.Int64  `tfsdk:"priority"`
	WorkspaceToken     types.String `tfsdk:"workspace_token"`
	Filter             types.String `tfsdk:"filter"`
	ParentSegmentToken types.String `tfsdk:"parent_segment_token"`
	Token              types.String `tfsdk:"token"`
	TrackUnallocated   types.Bool   `tfsdk:"track_unallocated"`
}

func (r *SegmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_segment"
}

func (r SegmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"title": schema.StringAttribute{
				MarkdownDescription: "The title of the Segment.",
				Required:            true,
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique segment identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the Segment.",
				Optional:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority of the Segment.",
				Optional:            true,
			},
			"workspace_token": schema.StringAttribute{
				MarkdownDescription: "Workspace token to add the segment to.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: "The filter query language to apply to the Segment. Additional documentation available at https://docs.vantage.sh/vql.",
				Optional:            true,
			},
			"track_unallocated": schema.BoolAttribute{
				MarkdownDescription: "Whether or not to track unallocated resources in this Segment.",
				Computed:            true,
				Optional:            true,
			},
			"parent_segment_token": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The token of the parent Segment this new Segment belongs to. Determines the Workspace the segment is assigned to.",
			},
		},
		MarkdownDescription: "Manages a Segment.",
	}
}

func (r SegmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SegmentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := segmentsv2.NewCreateSegmentParams()
	body := &modelsv2.PostSegments{
		Title:              data.Title.ValueStringPointer(),
		Filter:             data.Filter.ValueString(),
		ParentSegmentToken: data.ParentSegmentToken.ValueString(),
		Priority:           int32(data.Priority.ValueInt64()),
		WorkspaceToken:     data.WorkspaceToken.ValueString(),
		TrackUnallocated:   data.TrackUnallocated.ValueBoolPointer(),
	}

	params.WithSegments(body)
	out, err := r.client.V2.Segments.CreateSegment(params, r.client.Auth)
	if err != nil {
		handleError("Create Segment Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	data.ParentSegmentToken = types.StringValue(out.Payload.ParentFolder)
	data.Title = types.StringValue(out.Payload.Title)
	data.Filter = types.StringValue(out.Payload.Filter)
	data.Priority = types.Int64Value(int64(out.Payload.Priority))
	data.TrackUnallocated = types.BoolValue(out.Payload.TrackUnallocated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r SegmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *SegmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := segmentsv2.NewGetSegmentParams()
	params.SetSegmentToken(state.Token.ValueString())
	out, err := r.client.V2.Segments.GetSegment(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*segmentsv2.GetSegmentNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Segment Resource", &resp.Diagnostics, err)
		return
	}


	state.Token = types.StringValue(out.Payload.Token)
	state.Title = types.StringValue(out.Payload.Title)
	state.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	state.ParentSegmentToken = types.StringValue(out.Payload.ParentFolder)
	state.Filter = types.StringValue(out.Payload.Filter)
	state.Priority = types.Int64Value(int64(out.Payload.Priority))
	state.TrackUnallocated = types.BoolValue(out.Payload.TrackUnallocated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r SegmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SegmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}


	params := segmentsv2.NewUpdateSegmentParams()
	params.SetSegmentToken(data.Token.ValueString())

	model := &modelsv2.PutSegments{
		Title:              data.Title.ValueString(),
		Filter:             data.Filter.ValueString(),
		ParentSegmentToken: data.ParentSegmentToken.ValueString(),
		Description:        data.Description.ValueString(),
		Priority:           int32(data.Priority.ValueInt64()),
		TrackUnallocated:   data.TrackUnallocated.ValueBoolPointer(),
	}
	params.WithSegments(model)

	out, err := r.client.V2.Segments.UpdateSegment(params, r.client.Auth)

	if err != nil {
		handleError("Update Segment Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	data.ParentSegmentToken = types.StringValue(out.Payload.ParentFolder) // FIXME(jaxxstorm): is this correct?
	data.Description = types.StringValue(out.Payload.Description)
	data.Filter = types.StringValue(out.Payload.Filter)
	data.Title = types.StringValue(out.Payload.Title)
	data.TrackUnallocated = types.BoolValue(out.Payload.TrackUnallocated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r SegmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *SegmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := segmentsv2.NewDeleteSegmentParams()
	params.SetSegmentToken(state.Token.ValueString())
	_, err := r.client.V2.Segments.DeleteSegment(params, r.client.Auth)
	if err != nil {
		handleError("Delete Segment Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *SegmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
