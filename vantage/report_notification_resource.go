package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	notifsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/notifications"
)

type ReportNotificationResource struct {
	client *Client
}

func NewReportNotificationResource() resource.Resource {
	return &ReportNotificationResource{}
}

func (r *ReportNotificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_notification"
}

func (r ReportNotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique report notification identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the report notification",
				Required:            true,
			},
			"cost_report_token": schema.StringAttribute{
				MarkdownDescription: "Token for the cost report to be used in the notification",
				Required:            true,
			},
			"workspace_token": schema.StringAttribute{
				MarkdownDescription: "Token for the workspace the report notification is added toe notification",
				Optional:            true,
			},
			"user_tokens": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Tokens for the users to be notified",
				Required:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"frequency": schema.StringAttribute{
				MarkdownDescription: "The frequency at which the ReportNotification is sent. One of daily/weekly/monthly",
				Required:            true,
			},
			"change": schema.StringAttribute{
				MarkdownDescription: "The kind of change sent ReportNotification. One of percentage/dollars",
				Required:            true,
			},
		},
		MarkdownDescription: "Manages a Report Notification.",
	}
}

func (r *ReportNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *reportNotification
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := notifsv2.NewCreateReportNotificationParams()

	var userTokens []types.String
	if !data.UserTokens.IsNull() && !data.UserTokens.IsUnknown() {
		userTokens = make([]types.String, 0, len(data.UserTokens.Elements()))
		resp.Diagnostics.Append(data.UserTokens.ElementsAs(ctx, &userTokens, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	rp := &modelsv2.PostReportNotifications{
		Title:           data.Title.ValueStringPointer(),
		CostReportToken: data.CostReportToken.ValueStringPointer(),
		WorkspaceToken:  data.WorkspaceToken.ValueString(),
		UserTokens:      fromStringsValue(userTokens),
		Frequency:       data.Frequency.ValueStringPointer(),
		Change:          data.Change.ValueStringPointer(),
	}

	params.WithReportNotifications(rp)
	out, err := r.client.V2.Notifications.CreateReportNotification(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*notifsv2.CreateReportNotificationBadRequest); ok {
			handleBadRequest("Create Report Notification", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create Report Notification", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.CostReportToken = types.StringValue(out.Payload.CostReportToken)
	if out.Payload.UserTokens != nil {
		userTokensValue := make([]types.String, 0, len(out.Payload.UserTokens))
		for _, token := range out.Payload.UserTokens {
			userTokensValue = append(userTokensValue, types.StringValue(token))
		}
		set, diag := types.SetValueFrom(ctx, types.StringType, userTokensValue)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.UserTokens = set
	}
	data.Frequency = types.StringValue(out.Payload.Frequency)
	data.Change = types.StringValue(out.Payload.Change)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ReportNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *reportNotification

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := notifsv2.NewGetReportNotificationParams()
	params.SetReportNotificationToken(state.Token.ValueString())
	out, err := r.client.V2.Notifications.GetReportNotification(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*notifsv2.GetReportNotificationNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Report Notification", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.CostReportToken = types.StringValue(out.Payload.CostReportToken)
	userTokens, diag := types.SetValueFrom(ctx, types.StringType, out.Payload.UserTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	state.UserTokens = userTokens

	state.Frequency = types.StringValue(out.Payload.Frequency)
	state.Change = types.StringValue(out.Payload.Change)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ReportNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *reportNotification
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := notifsv2.NewUpdateReportNotificationParams()
	params.SetReportNotificationToken(data.Token.ValueString())
	userTokensSet, diag := types.SetValueFrom(ctx, types.StringType, data.UserTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	var userTokens []string
	userTokensSet.ElementsAs(ctx, &userTokens, false)
	rp := &modelsv2.PutReportNotifications{
		CostReportToken: data.CostReportToken.ValueString(),
		Change:          data.Change.ValueString(),
		Frequency:       data.Frequency.ValueString(),
		Title:           data.Title.ValueString(),
		UserTokens:      userTokens,
	}

	params.WithReportNotifications(rp)
	out, err := r.client.V2.Notifications.UpdateReportNotification(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*notifsv2.UpdateReportNotificationBadRequest); ok {
			handleBadRequest("Update Report Notification", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update Report Notification", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.CostReportToken = types.StringValue(out.Payload.CostReportToken)
	data.Title = types.StringValue(out.Payload.Title)
	data.Frequency = types.StringValue(out.Payload.Frequency)
	data.Change = types.StringValue(out.Payload.Change)
	if out.Payload.UserTokens != nil {
		userTokens, diag := types.SetValueFrom(ctx, types.StringType, out.Payload.UserTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.UserTokens = userTokens
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ReportNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *reportNotification
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := notifsv2.NewDeleteReportNotificationParams()
	params.SetReportNotificationToken(state.Token.ValueString())
	_, err := r.client.V2.Notifications.DeleteReportNotification(params, r.client.Auth)
	if err != nil {
		handleError("Delete Report Notification", &resp.Diagnostics, err)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ReportNotificationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
