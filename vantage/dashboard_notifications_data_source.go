package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_dashboard_notifications"
	dashboardnotifsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/dashboard_notifications"
)

var (
	_ datasource.DataSource              = (*dashboardNotificationsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dashboardNotificationsDataSource)(nil)
)

func NewDashboardNotificationsDataSource() datasource.DataSource {
	return &dashboardNotificationsDataSource{}
}

type dashboardNotificationsDataSource struct {
	client *Client
}

func (d *dashboardNotificationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

type dashboardNotificationsDataSourceModel struct {
	DashboardNotifications []dashboardNotificationDataSourceModel `tfsdk:"dashboard_notifications"`
}

type dashboardNotificationDataSourceModel struct {
	DashboardToken  types.String `tfsdk:"dashboard_token"`
	Frequency       types.String `tfsdk:"frequency"`
	Id              types.String `tfsdk:"id"`
	RecipientEmails types.List   `tfsdk:"recipient_emails"`
	Title           types.String `tfsdk:"title"`
	Token           types.String `tfsdk:"token"`
	UserTokens      types.List   `tfsdk:"user_tokens"`
}

func (d *dashboardNotificationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_notifications"
}

func (d *dashboardNotificationsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_dashboard_notifications.DashboardNotificationsDataSourceSchema(ctx)
}

func (d *dashboardNotificationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dashboardNotificationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardnotifsv2.NewGetDashboardNotificationsParams()
	out, err := d.client.V2.DashboardNotifications.GetDashboardNotifications(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Get Vantage Dashboard Notifications", err.Error())
		return
	}

	notifications := make([]dashboardNotificationDataSourceModel, 0, len(out.Payload.DashboardNotifications))
	for _, notification := range out.Payload.DashboardNotifications {
		userTokens, diag := types.ListValueFrom(ctx, types.StringType, notification.UserTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		recipientEmails, diag := types.ListValueFrom(ctx, types.StringType, notification.RecipientEmails)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		notifications = append(notifications, dashboardNotificationDataSourceModel{
			DashboardToken:  types.StringValue(notification.DashboardToken),
			Frequency:       types.StringValue(notification.Frequency),
			Id:              types.StringValue(notification.Token),
			RecipientEmails: recipientEmails,
			Title:           types.StringValue(notification.Title),
			Token:           types.StringValue(notification.Token),
			UserTokens:      userTokens,
		})
	}

	data.DashboardNotifications = notifications
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
