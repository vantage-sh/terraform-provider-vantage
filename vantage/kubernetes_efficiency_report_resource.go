package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_kubernetes_efficiency_report"
	k8seffreportsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/kubernetes_efficiency_reports"
)

var _ resource.Resource = (*kubernetesEfficiencyReportResource)(nil)
var _ resource.ResourceWithConfigure = (*kubernetesEfficiencyReportResource)(nil)
var _ resource.ResourceWithImportState = (*kubernetesEfficiencyReportResource)(nil)

func NewKubernetesEfficiencyReportResource() resource.Resource {
	return &kubernetesEfficiencyReportResource{}
}

type kubernetesEfficiencyReportResource struct {
	client *Client
}

func (r *kubernetesEfficiencyReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *kubernetesEfficiencyReportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)

}
func (r *kubernetesEfficiencyReportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_efficiency_report"
}

func (r *kubernetesEfficiencyReportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_kubernetes_efficiency_report.KubernetesEfficiencyReportResourceSchema(ctx)
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

func (r *kubernetesEfficiencyReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data kubernetesEfficiencyReportModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toCreateModel(ctx)

	params := k8seffreportsv2.NewCreateKubernetesEfficiencyReportParams().WithCreateKubernetesEfficiencyReport(model)
	out, err := r.client.V2.KubernetesEfficiencyReports.CreateKubernetesEfficiencyReport(params, r.client.Auth)

	if err != nil {
		if e, ok := err.(*k8seffreportsv2.CreateKubernetesEfficiencyReportBadRequest); ok {
			handleBadRequest("Create KubernetesEfficiencyReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create KubernetesEfficiencyReport Resource", &resp.Diagnostics, err)
		return

	}
	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kubernetesEfficiencyReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data kubernetesEfficiencyReportModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	params := k8seffreportsv2.NewGetKubernetesEfficiencyReportParams().WithKubernetesEfficiencyReportToken(data.Token.ValueString())
	out, err := r.client.V2.KubernetesEfficiencyReports.GetKubernetesEfficiencyReport(params, r.client.Auth)
	if err != nil {

		if _, ok := err.(*k8seffreportsv2.GetKubernetesEfficiencyReportNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Read KubernetesEfficiency Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kubernetesEfficiencyReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data kubernetesEfficiencyReportModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toUpdateModel(ctx)

	params := k8seffreportsv2.NewUpdateKubernetesEfficiencyReportParams().WithUpdateKubernetesEfficiencyReport(model).WithKubernetesEfficiencyReportToken(data.Token.ValueString())

	out, err := r.client.V2.KubernetesEfficiencyReports.UpdateKubernetesEfficiencyReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*k8seffreportsv2.UpdateKubernetesEfficiencyReportBadRequest); ok {
			handleBadRequest("Update KubernetesEfficiencyReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update KubernetesEfficiencyReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kubernetesEfficiencyReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data kubernetesEfficiencyReportModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := k8seffreportsv2.NewDeleteKubernetesEfficiencyReportParams().WithKubernetesEfficiencyReportToken(data.Token.ValueString())

	_, err := r.client.V2.KubernetesEfficiencyReports.DeleteKubernetesEfficiencyReport(params, r.client.Auth)
	if err != nil {
		handleError("Delete KubernetesEfficiencyReport Resource", &resp.Diagnostics, err)
	}
}
