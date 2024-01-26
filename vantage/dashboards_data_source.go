package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	dashboardsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/dashboards"
)

var (
	_ datasource.DataSource              = &dashboardsDataSource{}
	_ datasource.DataSourceWithConfigure = &dashboardsDataSource{}
)

func NewDashboardsDataSource() datasource.DataSource {
	return &dashboardsDataSource{}
}

type dashboardDataSourceModel struct {
	Token          types.String `tfsdk:"token"`
	Title          types.String `tfsdk:"title"`
	WidgetTokens   types.List   `tfsdk:"widget_tokens"`
	DateBin        types.String `tfsdk:"date_bin"`
	DateInterval   types.String `tfsdk:"date_interval"`
	StartDate      types.String `tfsdk:"start_date"`
	EndDate        types.String `tfsdk:"end_date"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}

type dashboardsDataSourceModel struct {
	Dashboards []dashboardDataSourceModel `tfsdk:"dashboards"`
}

type dashboardsDataSource struct {
	client *Client
}

func (d *dashboardsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

func (d *dashboardsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboards"
}

func (d *dashboardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dashboardsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	params := dashboardsv2.NewGetDashboardsParams()
	out, err := d.client.V2.Dashboards.GetDashboards(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Dashboards",
			err.Error(),
		)
		return
	}

	for _, dashboard := range out.Payload.Dashboards {

		var dashboardModel dashboardDataSourceModel
		dashboardModel.Token = types.StringValue(dashboard.Token)
		dashboardModel.Title = types.StringValue(dashboard.Title)
		dashboardModel.DateBin = types.StringValue(dashboard.DateBin)
		dashboardModel.DateInterval = types.StringValue(dashboard.DateInterval)
		dashboardModel.StartDate = types.StringValue(dashboard.StartDate)
		dashboardModel.EndDate = types.StringValue(dashboard.EndDate)
		dashboardModel.WorkspaceToken = types.StringValue(dashboard.WorkspaceToken)

		widgetTokens, diag := types.ListValueFrom(ctx, types.StringType, dashboard.WidgetTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		dashboardModel.WidgetTokens = widgetTokens
		state.Dashboards = append(state.Dashboards, dashboardModel)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *dashboardsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dashboards": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed: true,
						},
						"title": schema.StringAttribute{
							Computed: true,
						},
						"widget_tokens": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"date_bin": schema.StringAttribute{
							Computed: true,
						},
						"date_interval": schema.StringAttribute{
							Computed: true,
						},
						"start_date": schema.StringAttribute{
							Computed: true,
						},
						"end_date": schema.StringAttribute{
							Computed: true,
						},
						"workspace_token": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}

}
