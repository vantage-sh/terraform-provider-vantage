package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	segmentsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/segments"
)

var (
	_ datasource.DataSource              = &segmentsDataSource{}
	_ datasource.DataSourceWithConfigure = &segmentsDataSource{}
)

func NewSegmentsDataSource() datasource.DataSource {
	return &segmentsDataSource{}
}

type segmentDataSourceModel struct {
	Token              types.String `tfsdk:"token"`
	Title              types.String `tfsdk:"title"`
	Description        types.String `tfsdk:"description"`
	ParentFolder       types.String `tfsdk:"parent_folder"`
	ParentSegmentToken types.String `tfsdk:"parent_segment_token"`
	TrackUnallocated   types.Bool   `tfsdk:"track_unallocated"`
	Priority           types.Int64  `tfsdk:"priority"`
	WorkspaceToken     types.String `tfsdk:"workspace_token"`
	Filter             types.String `tfsdk:"filter"`
}

type segmentsDataSourceModel struct {
	Segments []segmentDataSourceModel `tfsdk:"segments"`
}

type segmentsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *segmentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

// Metadata implements datasource.DataSource.
func (d *segmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_segments"
}

// Read implements datasource.DataSource.
func (d *segmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state segmentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	params := segmentsv2.NewGetSegmentsParams()
	out, err := d.client.V2.Segments.GetSegments(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Segments",
			err.Error(),
		)
		return
	}

	for _, segment := range out.Payload.Segments {

		state.Segments = append(state.Segments, segmentDataSourceModel{
			Token:              types.StringValue(segment.Token),
			Title:              types.StringValue(segment.Title),
			ParentSegmentToken: types.StringValue(segment.ParentSegmentToken),
			Description:        types.StringValue(segment.Description),
			TrackUnallocated:   types.BoolValue(segment.TrackUnallocated),
			Priority:           types.Int64Value(int64(segment.Priority)),
			Filter:             types.StringValue(segment.Filter),
			WorkspaceToken:     types.StringValue(segment.WorkspaceToken),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *segmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}
