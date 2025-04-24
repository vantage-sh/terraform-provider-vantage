package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_cost_alert"
	costalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/cost_alerts"
)

var (
	_ resource.Resource                = (*costAlertResource)(nil)
	_ resource.ResourceWithConfigure   = (*costAlertResource)(nil)
	_ resource.ResourceWithImportState = (*costAlertResource)(nil)
)

type costAlertResource struct {
	client *Client
}

func NewCostAlertResource() resource.Resource {
	return &costAlertResource{}
}

func (r *costAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *costAlertResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cost_alert"
}

func (r *costAlertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_cost_alert.CostAlertResourceSchema(ctx)
}

func (r *costAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *costAlertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costalertsv2.NewCreateCostAlertParams().WithCreateCostAlert(input)
	out, err := r.client.V2.CostAlerts.CreateCostAlert(params, r.client.Auth)
	if err != nil {
		handleError("Create Cost Alert", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	resp.Diagnostics.Append(diag...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *costAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *costAlertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costalertsv2.NewGetCostAlertParams().WithCostAlertToken(data.Token.ValueString())
	out, err := r.client.V2.CostAlerts.GetCostAlert(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*costalertsv2.GetCostAlertNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Read Cost Alert", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	resp.Diagnostics.Append(diag...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *costAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *costAlertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costalertsv2.NewUpdateCostAlertParams().
		WithCostAlertToken(data.Token.ValueString()).
		WithUpdateCostAlert(input)

	out, err := r.client.V2.CostAlerts.UpdateCostAlert(params, r.client.Auth)
	if err != nil {
		handleError("Update Cost Alert", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	resp.Diagnostics.Append(diag...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *costAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *costAlertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costalertsv2.NewDeleteCostAlertParams().
		WithCostAlertToken(data.Token.ValueString())

	_, err := r.client.V2.CostAlerts.DeleteCostAlert(params, r.client.Auth)
	if err != nil {
		handleError("Delete Cost Alert", &resp.Diagnostics, err)
	}
}

func (r *costAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}
