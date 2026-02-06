package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_dashboards"
	dashboardsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/dashboards"
)

var (
	_ datasource.DataSource              = &dashboardsDataSource{}
	_ datasource.DataSourceWithConfigure = &dashboardsDataSource{}
)

type dashboardsDataSourceModel struct {
	Dashboards []dashboardModel `tfsdk:"dashboards"`
}

type dashboardsDataSource struct {
	client *Client
}

func NewDashboardsDataSource() datasource.DataSource {
	return &dashboardsDataSource{}
}

func (d *dashboardsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_dashboards.DashboardsDataSourceSchema(ctx)
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

	state.Dashboards = []dashboardModel{}
	for _, dashboard := range out.Payload.Dashboards {
		d := dashboardModel{}
		if diag := d.applyPayload(ctx, dashboard); diag.HasError() {
			resp.Diagnostics.Append(diag...)
		}
		state.Dashboards = append(state.Dashboards, d)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
