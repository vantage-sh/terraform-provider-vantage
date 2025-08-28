package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_network_flow_reports"
	nfrv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/network_flow_reports"
)

var (
	_ datasource.DataSource              = &networkFlowReportDataSource{}
	_ datasource.DataSourceWithConfigure = &networkFlowReportDataSource{}
)

func NewNetworkFlowReportDataSource() datasource.DataSource {
	return &networkFlowReportDataSource{}
}

type networkFlowReportDataSource struct {
	client *Client
}

func (d *networkFlowReportDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

func (d *networkFlowReportDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_flow_reports"
}

func (d *networkFlowReportDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_network_flow_reports.NetworkFlowReportsDataSourceSchema(ctx)
}

type NetworkFlowReportModel struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByToken types.String `tfsdk:"created_by_token"`

	DateInterval   types.String `tfsdk:"date_interval"`
	Default        types.Bool   `tfsdk:"default"`
	EndDate        types.String `tfsdk:"end_date"`
	Filter         types.String `tfsdk:"filter"`
	FlowDirection  types.String `tfsdk:"flow_direction"`
	FlowWeight     types.String `tfsdk:"flow_weight"`
	Groupings      types.String `tfsdk:"groupings"`
	StartDate      types.String `tfsdk:"start_date"`
	Title          types.String `tfsdk:"title"`
	Token          types.String `tfsdk:"token"`
	Id             types.String `tfsdk:"id"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}

type networkFlowReportDataSourceModel struct {
	NetworkFlowReports []NetworkFlowReportModel `tfsdk:"network_flow_reports"`
}

func (d *networkFlowReportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data networkFlowReportDataSourceModel
	// resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := nfrv2.NewGetNetworkFlowReportsParams()
	out, err := d.client.V2.NetworkFlowReports.GetNetworkFlowReports(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Network Flow Reports",
			err.Error(),
		)
		return
	}

	reports := []NetworkFlowReportModel{}
	for _, nfr := range out.Payload.NetworkFlowReports {
		reports = append(reports, NetworkFlowReportModel{
			CreatedAt:      types.StringValue(nfr.CreatedAt),
			CreatedByToken: types.StringValue(nfr.CreatedByToken),
			DateInterval:   types.StringValue(nfr.DateInterval),
			Default:        types.BoolValue(nfr.Default),
			EndDate:        types.StringValue(nfr.EndDate),
			Filter:         types.StringValue(nfr.Filter),
			FlowWeight:     types.StringValue(nfr.FlowWeight),
			FlowDirection:  types.StringValue(nfr.FlowDirection),
			Groupings:      types.StringValue(nfr.Groupings),
			StartDate:      types.StringValue(nfr.StartDate),
			Title:          types.StringValue(nfr.Title),
			Token:          types.StringValue(nfr.Token),
			Id:             types.StringValue(nfr.Token),
			WorkspaceToken: types.StringValue(nfr.WorkspaceToken),
		})
	}
	data.NetworkFlowReports = reports
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
