package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_business_metrics"
	businessmetricsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/business_metrics"
)

var (
	_ datasource.DataSource              = (*businessMetricsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*businessMetricsDataSource)(nil)
)

func NewBusinessMetricsDataSource() datasource.DataSource {
	return &businessMetricsDataSource{}
}

type businessMetricsDataSource struct {
	client *Client
}

func (r *businessMetricsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

type businessMetricsDataSourceModel struct {
	BusinessMetrics []businessMetricResourceModel `tfsdk:"business_metrics"`
}

func (d *businessMetricsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_business_metrics"
}

func (d *businessMetricsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_business_metrics.BusinessMetricsDataSourceSchema(ctx)
}

func (d *businessMetricsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data businessMetricsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := businessmetricsv2.NewGetBusinessMetricsParams()

	out, err := d.client.V2.BusinessMetrics.GetBusinessMetrics(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Business Metrics",
			err.Error(),
		)
		return
	}

	metrics := []businessMetricResourceModel{}
	for _, metric := range out.Payload.BusinessMetrics {
		model := businessMetricResourceModel{}
		diag := model.applyPayload(ctx, metric, true)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		metrics = append(metrics, model)
	}

	data.BusinessMetrics = metrics
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
