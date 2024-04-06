package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_report_alert"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	reportalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/report_alerts"
)

var _ resource.Resource = (*reportAlertResource)(nil)
var _ resource.ResourceWithConfigure = (*reportAlertResource)(nil)

func NewReportAlertResource() resource.Resource {
	return &reportAlertResource{}
}

type reportAlertResource struct {
	client *Client
}

// Configure implements resource.ResourceWithConfigure.
func (r *reportAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r *reportAlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_alert"
}

func (r *reportAlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_report_alert.ReportAlertResourceSchema(ctx)
}

func (r *reportAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_report_alert.ReportAlertModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := reportalertsv2.NewCreateReportAlertParams()

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

	createReportAlert := &modelsv2.CreateReportAlert{
		CostReportToken:   data.CostReportToken.ValueStringPointer(),
		Threshold:         int32(data.Threshold.ValueInt64()),
		UserTokens:        fromStringsValue(userTokens),
		RecipientChannels: fromStringsValue(recipientChannels),
	}

	params.WithCreateReportAlert(createReportAlert)
	out, err := r.client.V2.ReportAlerts.CreateReportAlert(params, r.client.Auth)

	if err != nil {
		if e, ok := err.(*reportalertsv2.CreateReportAlertBadRequest); ok {
			handleBadRequest("Create Report Alert", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create Report Alert", &resp.Diagnostics, err)
		return
	}

	readPayloadIntoResourceModel(out.Payload, &data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *reportAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_report_alert.ReportAlertModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := reportalertsv2.NewGetReportAlertParams()
	params.SetReportAlertToken(data.Token.ValueString())
	out, err := r.client.V2.ReportAlerts.GetReportAlert(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*reportalertsv2.GetReportAlertNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Report Alert", &resp.Diagnostics, err)
		return
	}

	readPayloadIntoResourceModel(out.Payload, &data)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *reportAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_report_alert.ReportAlertModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := reportalertsv2.NewUpdateReportAlertParams()
	params.SetReportAlertToken(data.Token.ValueString())

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

	updateReportAlert := &modelsv2.UpdateReportAlert{
		Threshold:         int32(data.Threshold.ValueInt64()),
		UserTokens:        userTokens,
		RecipientChannels: recipientChannels,
	}

	params.WithUpdateReportAlert(updateReportAlert)
	out, err := r.client.V2.ReportAlerts.UpdateReportAlert(params, r.client.Auth)
	if err != nil {
		handleError("Update Report Alert", &resp.Diagnostics, err)
		return
	}

	readPayloadIntoResourceModel(out.Payload, &data)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *reportAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_report_alert.ReportAlertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := reportalertsv2.NewDeleteReportAlertParams()
	params.SetReportAlertToken(data.Token.ValueString())
	_, err := r.client.V2.ReportAlerts.DeleteReportAlert(params, r.client.Auth)
	if err != nil {
		handleError("Delete Report Alert", &resp.Diagnostics, err)
	}
	// Delete API call logic
}

func readPayloadIntoResourceModel(payload *modelsv2.ReportAlert, data *resource_report_alert.ReportAlertModel) {
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
