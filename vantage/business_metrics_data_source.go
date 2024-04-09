package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	BusinessMetrics []businessMetricDataSourceModel `tfsdk:"business_metrics"`
}

type businessMetricDataSourceModel struct {
	CostReportTokensWithMetadata types.List   `tfsdk:"cost_report_tokens_with_metadata"`
	CreatedByToken               types.String `tfsdk:"created_by_token"`
	Title                        types.String `tfsdk:"title"`
	Token                        types.String `tfsdk:"token"`
	Values                       types.List   `tfsdk:"values"`
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

	metrics := []businessMetricDataSourceModel{}
	for _, metric := range out.Payload.BusinessMetrics {
		costReportTokens, diag := types.ListValueFrom(ctx, types.StringType, metric.CostReportTokensWithMetadata)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		values, diag := types.ListValueFrom(ctx, types.StringType, metric.Values)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		metrics = append(metrics, businessMetricDataSourceModel{
			CostReportTokensWithMetadata: costReportTokens,
			CreatedByToken:               types.StringValue(metric.CreatedByToken),
			Title:                        types.StringValue(metric.Title),
			Token:                        types.StringValue(metric.Token),
			Values:                       values,
		})
	}

	data.BusinessMetrics = metrics
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
