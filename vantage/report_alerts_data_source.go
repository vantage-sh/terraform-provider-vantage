package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_report_alerts"
	reportalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/report_alerts"
)

var _ datasource.DataSource = (*reportAlertsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*reportAlertsDataSource)(nil)

func NewReportAlertsDataSource() datasource.DataSource {
	return &reportAlertsDataSource{}
}

type reportAlertsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *reportAlertsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

type reportAlertsDataSourceModel struct {
	ReportAlerts []reportAlertDataSourceModel `tfsdk:"report_alerts"`
}

type reportAlertDataSourceModel struct {
	CostReportToken   types.String `tfsdk:"cost_report_token"`
	CreatedAt         types.String `tfsdk:"created_at"`
	RecipientChannels types.List   `tfsdk:"recipient_channels"`
	Threshold         types.Int64  `tfsdk:"threshold"`
	Token             types.String `tfsdk:"token"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	UserTokens        types.List   `tfsdk:"user_tokens"`
}

func (d *reportAlertsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_alerts"
}

func (d *reportAlertsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_report_alerts.ReportAlertsDataSourceSchema(ctx)
}

func (d *reportAlertsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data reportAlertsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := reportalertsv2.NewGetReportAlertsParams()
	out, err := d.client.V2.ReportAlerts.GetReportAlerts(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Report Alerts",
			err.Error(),
		)
		return
	}

	reportAlerts := []reportAlertDataSourceModel{}
	for _, reportAlert := range out.Payload.ReportAlerts {
		userTokens, diag := types.ListValueFrom(ctx, types.StringType, reportAlert.UserTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		recipientChannels, diag := types.ListValueFrom(ctx, types.StringType, reportAlert.RecipientChannels)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		reportAlerts = append(reportAlerts, reportAlertDataSourceModel{
			CostReportToken:   types.StringValue(reportAlert.CostReportToken),
			CreatedAt:         types.StringValue(reportAlert.CreatedAt),
			RecipientChannels: recipientChannels,
			Threshold:         types.Int64Value((int64)(reportAlert.Threshold)),
			Token:             types.StringValue(reportAlert.Token),
			UpdatedAt:         types.StringValue(reportAlert.UpdatedAt),
			UserTokens:        userTokens,
		})
		data.ReportAlerts = reportAlerts
	}
	// Save data into Terraform state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
