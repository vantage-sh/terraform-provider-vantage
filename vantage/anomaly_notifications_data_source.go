package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_anomaly_notifications"
	anomalynotifsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/anomaly_notifications"
)

var _ datasource.DataSource = (*anomalyNotificationsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*anomalyNotificationsDataSource)(nil)

func NewAnomalyNotificationsDataSource() datasource.DataSource {
	return &anomalyNotificationsDataSource{}
}

type anomalyNotificationsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *anomalyNotificationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

type anomalyNotificationsDataSourceModel struct {
	AnomalyNotifications []anomalyNotificationDataSourceModel `tfsdk:"anomaly_notifications"`
}

type anomalyNotificationDataSourceModel struct {
	CostReportToken   types.String `tfsdk:"cost_report_token"`
	CreatedAt         types.String `tfsdk:"created_at"`
	RecipientChannels types.List   `tfsdk:"recipient_channels"`
	Threshold         types.Int64  `tfsdk:"threshold"`
	Token             types.String `tfsdk:"token"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	UserTokens        types.List   `tfsdk:"user_tokens"`
}

func (d *anomalyNotificationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_anomaly_notifications"
}

func (d *anomalyNotificationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_anomaly_notifications.AnomalyNotificationsDataSourceSchema(ctx)
}

func (d *anomalyNotificationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data anomalyNotificationsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := anomalynotifsv2.NewGetAnomalyNotificationsParams()
	out, err := d.client.V2.AnomalyNotifications.GetAnomalyNotifications(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Report Alerts",
			err.Error(),
		)
		return
	}

	anomalyNotifications := []anomalyNotificationDataSourceModel{}
	for _, anomalyNotification := range out.Payload.AnomalyNotifications {
		userTokens, diag := types.ListValueFrom(ctx, types.StringType, anomalyNotification.UserTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		recipientChannels, diag := types.ListValueFrom(ctx, types.StringType, anomalyNotification.RecipientChannels)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		anomalyNotifications = append(anomalyNotifications, anomalyNotificationDataSourceModel{
			CostReportToken:   types.StringValue(anomalyNotification.CostReportToken),
			CreatedAt:         types.StringValue(anomalyNotification.CreatedAt),
			RecipientChannels: recipientChannels,
			Threshold:         types.Int64Value((int64)(anomalyNotification.Threshold)),
			Token:             types.StringValue(anomalyNotification.Token),
			UpdatedAt:         types.StringValue(anomalyNotification.UpdatedAt),
			UserTokens:        userTokens,
		})
		data.AnomalyNotifications = anomalyNotifications
	}
	// Save data into Terraform state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
