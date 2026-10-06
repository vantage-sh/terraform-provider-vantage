package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_dashboard_notification"
	dashboardnotifsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/dashboard_notifications"
)

var (
	_ resource.Resource                = (*dashboardNotificationResource)(nil)
	_ resource.ResourceWithConfigure   = (*dashboardNotificationResource)(nil)
	_ resource.ResourceWithImportState = (*dashboardNotificationResource)(nil)
)

func NewDashboardNotificationResource() resource.Resource {
	return &dashboardNotificationResource{}
}

type dashboardNotificationResource struct {
	client *Client
}

func (r *dashboardNotificationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r *dashboardNotificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_notification"
}

func (r *dashboardNotificationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_dashboard_notification.DashboardNotificationResourceSchema(ctx)

	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		Description:         "The token of the DashboardNotification",
		MarkdownDescription: "The token of the DashboardNotification",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// workspace_token is create-only and not returned by the API.
	s.Attributes["workspace_token"] = schema.StringAttribute{
		Optional:            true,
		Description:         "The token of the Workspace to add the DashboardNotification to. Required if the API token is associated with multiple Workspaces.",
		MarkdownDescription: "The token of the Workspace to add the DashboardNotification to. Required if the API token is associated with multiple Workspaces.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = s
}

func (r *dashboardNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_dashboard_notification.DashboardNotificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardnotifsv2.NewCreateDashboardNotificationParams()
	createModel, diags := toCreateDashboardNotification(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params.WithCreateDashboardNotification(createModel)
	out, err := r.client.V2.DashboardNotifications.CreateDashboardNotification(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*dashboardnotifsv2.CreateDashboardNotificationBadRequest); ok {
			handleBadRequest("Create Dashboard Notification", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create Dashboard Notification", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(applyDashboardNotificationPayload(ctx, out.Payload, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dashboardNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_dashboard_notification.DashboardNotificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardnotifsv2.NewGetDashboardNotificationParams()
	params.SetDashboardNotificationToken(data.Token.ValueString())
	out, err := r.client.V2.DashboardNotifications.GetDashboardNotification(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*dashboardnotifsv2.GetDashboardNotificationNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Dashboard Notification", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(applyDashboardNotificationPayload(ctx, out.Payload, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dashboardNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *dashboardNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_dashboard_notification.DashboardNotificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardnotifsv2.NewUpdateDashboardNotificationParams()
	params.SetDashboardNotificationToken(data.Token.ValueString())
	updateModel, diags := toUpdateDashboardNotification(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params.WithUpdateDashboardNotification(updateModel)
	out, err := r.client.V2.DashboardNotifications.UpdateDashboardNotification(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*dashboardnotifsv2.UpdateDashboardNotificationBadRequest); ok {
			handleBadRequest("Update Dashboard Notification", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update Dashboard Notification", &resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(applyDashboardNotificationPayload(ctx, out.Payload, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dashboardNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_dashboard_notification.DashboardNotificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardnotifsv2.NewDeleteDashboardNotificationParams()
	params.SetDashboardNotificationToken(data.Token.ValueString())
	_, err := r.client.V2.DashboardNotifications.DeleteDashboardNotification(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*dashboardnotifsv2.DeleteDashboardNotificationNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Delete Dashboard Notification", &resp.Diagnostics, err)
		return
	}
}
