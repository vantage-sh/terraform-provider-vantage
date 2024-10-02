package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_anomaly_notification"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	anomalynotifsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/anomaly_notifications"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resource.Resource                = (*anomalyNotificationResource)(nil)
	_ resource.ResourceWithConfigure   = (*anomalyNotificationResource)(nil)
	_ resource.ResourceWithImportState = (*anomalyNotificationResource)(nil)
)

func NewAnomalyNotificationResource() resource.Resource {
	return &anomalyNotificationResource{}
}

type anomalyNotificationResource struct {
	client *Client
}

// Configure implements resource.ResourceWithConfigure.
func (r *anomalyNotificationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r *anomalyNotificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_anomaly_notification"
}

func (r *anomalyNotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cost_report_token": schema.StringAttribute{
				Required:            true,
				Description:         "The token of the Cost Report folder that has the notification.",
				MarkdownDescription: "The token of the Cost Report folder that has the notification.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the AnomalyNotification was created. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the AnomalyNotification was created. ISO 8601 Formatted.",
			},
			"recipient_channels": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The Slack/MS Teams channels that receive the notification.",
				MarkdownDescription: "The Slack/MS Teams channels that receive the notification.",
			},
			"threshold": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The threshold amount that must be met for the notification to fire.",
				MarkdownDescription: "The threshold amount that must be met for the notification to fire.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the report alert",
				MarkdownDescription: "The token of the report alert",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the AnomalyNotification was last updated at. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the AnomalyNotification was last updated at. ISO 8601 Formatted.",
			},
			"user_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The tokens of the users that receive the notification.",
				MarkdownDescription: "The tokens of the users that receive the notification.",
			},
		},
	}
}

func (r *anomalyNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_anomaly_notification.AnomalyNotificationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := anomalynotifsv2.NewCreateAnomalyNotificationParams()

	var userTokens []types.String
	if !data.UserTokens.IsNull() && !data.UserTokens.IsUnknown() {
		userTokens = make([]types.String, 0, len(data.UserTokens.Elements()))
		resp.Diagnostics.Append(data.UserTokens.ElementsAs(ctx, &userTokens, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var recipientChannels []types.String
	if !data.RecipientChannels.IsNull() && !data.RecipientChannels.IsUnknown() {
		recipientChannels = make([]types.String, 0, len(data.RecipientChannels.Elements()))
		resp.Diagnostics.Append(data.RecipientChannels.ElementsAs(ctx, &recipientChannels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createAnomalyNotification := &modelsv2.CreateAnomalyNotification{
		CostReportToken:   data.CostReportToken.ValueStringPointer(),
		Threshold:         int32(data.Threshold.ValueInt64()),
		UserTokens:        fromStringsValue(userTokens),
		RecipientChannels: fromStringsValue(recipientChannels),
	}

	params.WithCreateAnomalyNotification(createAnomalyNotification)
	out, err := r.client.V2.AnomalyNotifications.CreateAnomalyNotification(params, r.client.Auth)

	if err != nil {
		if e, ok := err.(*anomalynotifsv2.CreateAnomalyNotificationBadRequest); ok {
			handleBadRequest("Create Anomaly Notification", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create Anomaly Notification", &resp.Diagnostics, err)
		return
	}

	readPayloadIntoResourceModel(out.Payload, &data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *anomalyNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_anomaly_notification.AnomalyNotificationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := anomalynotifsv2.NewGetAnomalyNotificationParams()
	params.SetAnomalyNotificationToken(data.Token.ValueString())
	out, err := r.client.V2.AnomalyNotifications.GetAnomalyNotification(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*anomalynotifsv2.GetAnomalyNotificationNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Anomaly Notification", &resp.Diagnostics, err)
		return
	}

	readPayloadIntoResourceModel(out.Payload, &data)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *anomalyNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *anomalyNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_anomaly_notification.AnomalyNotificationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := anomalynotifsv2.NewUpdateAnomalyNotificationParams()
	params.SetAnomalyNotificationToken(data.Token.ValueString())

	userTokensList, diag := types.ListValueFrom(ctx, types.StringType, data.UserTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var userTokens []string
	userTokensList.ElementsAs(ctx, userTokens, false)

	recipientChannelsList, diag := types.ListValueFrom(ctx, types.StringType, data.RecipientChannels)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var recipientChannels []string
	recipientChannelsList.ElementsAs(ctx, recipientChannels, false)

	updateAnomalyNotification := &modelsv2.UpdateAnomalyNotification{
		Threshold:         int32(data.Threshold.ValueInt64()),
		UserTokens:        userTokens,
		RecipientChannels: recipientChannels,
	}

	params.WithUpdateAnomalyNotification(updateAnomalyNotification)
	out, err := r.client.V2.AnomalyNotifications.UpdateAnomalyNotification(params, r.client.Auth)
	if err != nil {
		handleError("Update Anomaly Notification", &resp.Diagnostics, err)
		return
	}

	readPayloadIntoResourceModel(out.Payload, &data)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *anomalyNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_anomaly_notification.AnomalyNotificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := anomalynotifsv2.NewDeleteAnomalyNotificationParams()
	params.SetAnomalyNotificationToken(data.Token.ValueString())

	_, err := r.client.V2.AnomalyNotifications.DeleteAnomalyNotification(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*anomalynotifsv2.GetAnomalyNotificationNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Anomaly Notification", &resp.Diagnostics, err)
		return
	}
}

func readPayloadIntoResourceModel(payload *modelsv2.AnomalyNotification, data *resource_anomaly_notification.AnomalyNotificationModel) {
	data.Token = types.StringValue(payload.Token)
	data.CostReportToken = types.StringValue(payload.CostReportToken)
	data.CreatedAt = types.StringValue(payload.CreatedAt)
	data.UpdatedAt = types.StringValue(payload.UpdatedAt)
	data.Threshold = types.Int64Value((int64)(payload.Threshold))
	if payload.UserTokens != nil {
		list, diag := stringListFrom(payload.UserTokens)
		if diag.HasError() {
			return
		}
		data.UserTokens = list
	}

	if payload.RecipientChannels != nil {
		list, diag := stringListFrom(payload.RecipientChannels)
		if diag.HasError() {
			return
		}
		data.RecipientChannels = list
	}

}
