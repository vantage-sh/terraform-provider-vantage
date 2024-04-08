package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_report_notifications"
	reportnotifsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/notifications"
)

var _ datasource.DataSource = (*reportNotificationsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*reportNotificationsDataSource)(nil)

func NewReportNotificationsDataSource() datasource.DataSource {
	return &reportNotificationsDataSource{}
}

type reportNotificationsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *reportNotificationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

type reportNotificationsDataSourceModel struct {
	ReportNotifications []reportNotificationDataSourceModel `tfsdk:"report_notifications"`
}

type reportNotificationDataSourceModel struct {
	Change            types.String `tfsdk:"change"`
	CostReportToken   types.String `tfsdk:"cost_report_token"`
	Frequency         types.String `tfsdk:"frequency"`
	Title             types.String `tfsdk:"title"`
	Token             types.String `tfsdk:"token"`
	UserTokens        types.List   `tfsdk:"user_tokens"`
	RecipientChannels types.List   `tfsdk:"recipient_channels"`
}

func (d *reportNotificationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_notifications"
}

func (d *reportNotificationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_report_notifications.ReportNotificationsDataSourceSchema(ctx)
}

func (d *reportNotificationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data reportNotificationsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := reportnotifsv2.NewGetReportNotificationsParams()
	out, err := d.client.V2.Notifications.GetReportNotifications(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Get Vantage Report Notifications", err.Error())
		return
	}

	notifications := []reportNotificationDataSourceModel{}
	for _, notification := range out.Payload.ReportNotifications {

		userTokensVal, diag := types.ListValueFrom(ctx, types.StringType, notification.UserTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		recipientChannelsVal, diag := types.ListValueFrom(ctx, types.StringType, notification.RecipientChannels)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		notifications = append(notifications, reportNotificationDataSourceModel{
			Change:            types.StringValue(notification.Change),
			CostReportToken:   types.StringValue(notification.CostReportToken),
			Frequency:         types.StringValue(notification.Frequency),
			Title:             types.StringValue(notification.Title),
			Token:             types.StringValue(notification.Token),
			UserTokens:        userTokensVal,
			RecipientChannels: recipientChannelsVal,
		})

	}
	data.ReportNotifications = notifications

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
