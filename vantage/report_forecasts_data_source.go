package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_report_forecasts"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	reportforecastsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/report_forecasts"
)

var (
	_ datasource.DataSource              = (*reportForecastsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*reportForecastsDataSource)(nil)
)

func NewReportForecastsDataSource() datasource.DataSource {
	return &reportForecastsDataSource{}
}

type reportForecastsDataSource struct {
	client *Client
}

type reportForecastsDataSourceModel struct {
	CostReportToken types.String `tfsdk:"cost_report_token"`
	ReportForecasts types.List   `tfsdk:"report_forecasts"`
}

func (d *reportForecastsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *reportForecastsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_forecasts"
}

func (d *reportForecastsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_report_forecasts.ReportForecastsDataSourceSchema(ctx)
}

func (d *reportForecastsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data reportForecastsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := reportforecastsv2.NewGetReportForecastsParams().
		WithCostReportToken(data.CostReportToken.ValueString())
	out, err := d.client.V2.ReportForecasts.GetReportForecasts(params, d.client.Auth)
	if err != nil {
		if e, ok := err.(*reportforecastsv2.GetReportForecastsNotFound); ok {
			handleBadRequest("Get Report Forecasts", &resp.Diagnostics, e.GetPayload())
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Report Forecasts",
			err.Error(),
		)
		return
	}

	forecasts := out.Payload.ReportForecasts
	if forecasts == nil {
		forecasts = []*modelsv2.ReportForecast{}
	}

	elements := make([]attr.Value, 0, len(forecasts))
	for _, forecast := range forecasts {
		value, diags := reportForecastDataSourceValueFromAPI(ctx, forecast)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elements = append(elements, value)
	}

	list, diags := types.ListValue(
		datasource_report_forecasts.ReportForecastsValue{}.Type(ctx),
		elements,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ReportForecasts = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func reportForecastDataSourceValueFromAPI(ctx context.Context, src *modelsv2.ReportForecast) (datasource_report_forecasts.ReportForecastsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if src == nil {
		return datasource_report_forecasts.NewReportForecastsValueNull(), diags
	}

	tokens := src.ScenarioModelTokens
	if tokens == nil {
		tokens = []string{}
	}
	tokenList, d := types.ListValueFrom(ctx, types.StringType, tokens)
	diags.Append(d...)
	if diags.HasError() {
		return datasource_report_forecasts.NewReportForecastsValueNull(), diags
	}

	value, d := datasource_report_forecasts.NewReportForecastsValue(
		datasource_report_forecasts.ReportForecastsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"business_metric_token": types.StringPointerValue(src.BusinessMetricToken),
			"cost_report_token":     types.StringValue(src.CostReportToken),
			"created_at":            types.StringValue(src.CreatedAt),
			"created_by_token":      types.StringPointerValue(src.CreatedByToken),
			"id":                    types.StringValue(src.Token),
			"is_default":            types.BoolValue(src.IsDefault),
			"scenario_model_tokens": tokenList,
			"title":                 types.StringValue(src.Title),
			"token":                 types.StringValue(src.Token),
			"updated_at":            types.StringValue(src.UpdatedAt),
		},
	)
	diags.Append(d...)
	return value, diags
}
