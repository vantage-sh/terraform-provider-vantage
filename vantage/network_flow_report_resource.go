package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_network_flow_report"
	nfrv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/network_flow_reports"
)

var (
	_ resource.Resource                = (*networkFlowReportResource)(nil)
	_ resource.ResourceWithConfigure   = (*networkFlowReportResource)(nil)
	_ resource.ResourceWithImportState = (*networkFlowReportResource)(nil)
)

type networkFlowReportResource struct {
	client *Client
}

func NewNetworkFlowReportResource() resource.Resource {
	return &networkFlowReportResource{}
}

func (r *networkFlowReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *networkFlowReportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *networkFlowReportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_flow_report"
}

func (r *networkFlowReportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_network_flow_report.NetworkFlowReportResourceSchema(ctx)
	attrs := s.GetAttributes()
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	resp.Schema = s
}

func (r *networkFlowReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkFlowReportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toCreateModel(ctx)
	params := nfrv2.NewCreateNetworkFlowReportParams().WithCreateNetworkFlowReport(model)
	out, err := r.client.V2.NetworkFlowReports.CreateNetworkFlowReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*nfrv2.CreateNetworkFlowReportBadRequest); ok {
			handleBadRequest("Create NetworkFlowReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create NetworkFlowReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkFlowReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkFlowReportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := nfrv2.NewGetNetworkFlowReportParams().WithNetworkFlowReportToken(data.Token.ValueString())
	out, err := r.client.V2.NetworkFlowReports.GetNetworkFlowReport(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*nfrv2.GetNetworkFlowReportNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		handleError("Get NetworkFlowReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkFlowReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data networkFlowReportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toUpdateModel(ctx)
	params := nfrv2.NewUpdateNetworkFlowReportParams().WithNetworkFlowReportToken(data.Token.ValueString()).WithUpdateNetworkFlowReport(model)
	out, err := r.client.V2.NetworkFlowReports.UpdateNetworkFlowReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*nfrv2.UpdateNetworkFlowReportBadRequest); ok {
			handleBadRequest("Update NetworkFlowReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update NetworkFlowReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkFlowReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkFlowReportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := nfrv2.NewDeleteNetworkFlowReportParams().WithNetworkFlowReportToken(data.Token.ValueString())
	_, err := r.client.V2.NetworkFlowReports.DeleteNetworkFlowReport(params, r.client.Auth)
	if err != nil {
		handleError("Delete NetworkFlowReport Resource", &resp.Diagnostics, err)
	}
}
