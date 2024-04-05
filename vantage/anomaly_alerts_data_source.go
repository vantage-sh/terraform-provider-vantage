package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_anomaly_alerts"
	anomalyalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/anomaly_alerts"
)

var _ datasource.DataSource = (*anomalyAlertsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*anomalyAlertsDataSource)(nil)

func NewAnomalyAlertsDataSource() datasource.DataSource {
	return &anomalyAlertsDataSource{}
}

type anomalyAlertsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *anomalyAlertsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

type anomalyAlertsDataSourceModel struct {
	AnomalyAlerts []anomalyAlertDataSourceModel `tfsdk:"anomaly_alerts"`
}

type anomalyAlertDataSourceModel struct {
	AlertedAt       types.String `tfsdk:"alerted_at"`
	Amount          types.String `tfsdk:"amount"`
	Category        types.String `tfsdk:"category"`
	CostReportToken types.String `tfsdk:"cost_report_token"`
	CreatedAt       types.String `tfsdk:"created_at"`
	Feedback        types.String `tfsdk:"feedback"`
	PreviousAmount  types.String `tfsdk:"previous_amount"`
	Provider        types.String `tfsdk:"provider"`
	Service         types.String `tfsdk:"service"`
	SevenDayAverage types.String `tfsdk:"seven_day_average"`
	Status          types.String `tfsdk:"status"`
	Token           types.String `tfsdk:"token"`
}

func (d *anomalyAlertsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_anomaly_alerts"
}

func (d *anomalyAlertsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_anomaly_alerts.AnomalyAlertsDataSourceSchema(ctx)
}

func (d *anomalyAlertsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data anomalyAlertsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	params := anomalyalertsv2.NewGetAnomalyAlertsParams()
	out, err := d.client.V2.AnomalyAlerts.GetAnomalyAlerts(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Anomaly Alerts",
			err.Error(),
		)
		return
	}

	alerts := []anomalyAlertDataSourceModel{}
	for _, a := range out.Payload.AnomalyAlerts {
		alerts = append(alerts, anomalyAlertDataSourceModel{
			AlertedAt:       types.StringValue(a.AlertedAt),
			Amount:          types.StringValue(a.Amount),
			Category:        types.StringValue(a.Category),
			CostReportToken: types.StringValue(a.CostReportToken),
			CreatedAt:       types.StringValue(a.CreatedAt),
			Feedback:        types.StringValue(a.Feedback),
			PreviousAmount:  types.StringValue(a.PreviousAmount),
			Provider:        types.StringValue(a.Provider),
			Service:         types.StringValue(a.Service),
			SevenDayAverage: types.StringValue(a.SevenDayAverage),
			Status:          types.StringValue(a.Status),
			Token:           types.StringValue(a.Token),
		})
	}
	data.AnomalyAlerts = alerts
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}
