package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resourcereportsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/resource_reports"
)

var (
	_ datasource.DataSource              = &resourceReportsDataSource{}
	_ datasource.DataSourceWithConfigure = &resourceReportsDataSource{}
)

func NewResourceReportsDataSource() datasource.DataSource {
	return &resourceReportsDataSource{}
}

type resourceReportDataSourceModel struct {
	Token          types.String `tfsdk:"token"`
	Title          types.String `tfsdk:"title"`
	UserToken      types.String `tfsdk:"user_token"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

type resourceReportsDataSourceModel struct {
	ResourceReports []resourceReportDataSourceModel `tfsdk:"resource_reports"`
}

type resourceReportsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (r *resourceReportsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

// Metadata implements datasource.DataSource.
func (r *resourceReportsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_reports"
}

// Read implements datasource.DataSource.
func (r *resourceReportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state resourceReportsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	params := resourcereportsv2.NewGetResourceReportsParams()
	out, err := r.client.V2.ResourceReports.GetResourceReports(params, r.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Resource Reports",
			err.Error(),
		)
		return
	}

	for _, report := range out.Payload.ResourceReports {
		state.ResourceReports = append(state.ResourceReports, resourceReportDataSourceModel{
			Token:          types.StringValue(report.Token),
			Title:          types.StringValue(report.Title),
			UserToken:      types.StringPointerValue(report.UserToken),
			WorkspaceToken: types.StringValue(report.WorkspaceToken),
			CreatedAt:      types.StringValue(report.CreatedAt),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements datasource.DataSource.
func (r *resourceReportsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_reports": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed: true,
						},
						"title": schema.StringAttribute{
							Computed: true,
						},
						"user_token": schema.StringAttribute{
							Computed: true,
						},
						"workspace_token": schema.StringAttribute{
							Computed: true,
						},
						"created_at": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}
