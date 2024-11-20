package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aggregated_by": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The column by which the costs are aggregated.",
				MarkdownDescription: "The column by which the costs are aggregated.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"idle_cost",
						"amount",
						"cost_efficiency",
					),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
			},
			"date_bin": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The date bin of the KubernetesEfficiencyReport.",
				MarkdownDescription: "The date bin of the KubernetesEfficiencyReport.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"cumulative",
						"day",
						"week",
						"month",
					),
				},
				Default: stringdefault.StaticString("day"),
			},
			"date_bucket": schema.StringAttribute{
				Computed:            true,
				Description:         "How costs are grouped and displayed in the KubernetesEfficiencyReport. Possible values: day, week, month.",
				MarkdownDescription: "How costs are grouped and displayed in the KubernetesEfficiencyReport. Possible values: day, week, month.",
				Default:             stringdefault.StaticString("day"),
			},
			"date_interval": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The date interval of the KubernetesEfficiencyReport. Incompatible with 'start_date' and 'end_date' parameters. Defaults to 'this_month' if start_date and end_date are not provided.",
				MarkdownDescription: "The date interval of the KubernetesEfficiencyReport. Incompatible with 'start_date' and 'end_date' parameters. Defaults to 'this_month' if start_date and end_date are not provided.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"this_month",
						"last_7_days",
						"last_30_days",
						"last_month",
						"last_3_months",
						"last_6_months",
						"custom",
						"last_12_months",
						"last_24_months",
						"last_36_months",
						"next_month",
						"next_3_months",
						"next_6_months",
						"next_12_months",
						"year_to_date",
					),
				},
			},
			"default": schema.BoolAttribute{
				Computed:            true,
				Description:         "Indicates whether the KubernetesEfficiencyReport is the default report.",
				MarkdownDescription: "Indicates whether the KubernetesEfficiencyReport is the default report.",
			},
			"end_date": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The end date of the KubernetesEfficiencyReport. ISO 8601 Formatted. Incompatible with 'date_interval' parameter.",
				MarkdownDescription: "The end date of the KubernetesEfficiencyReport. ISO 8601 Formatted. Incompatible with 'date_interval' parameter.",
			},
			"filter": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The filter query language to apply to the KubernetesEfficiencyReport. Additional documentation available at https://docs.vantage.sh/vql.",
				MarkdownDescription: "The filter query language to apply to the KubernetesEfficiencyReport. Additional documentation available at https://docs.vantage.sh/vql.",
			},
			"groupings": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Grouping values for aggregating costs on the KubernetesEfficiencyReport. Valid groupings: cluster_id, namespace, labeled, category, label, label:<label_name>.",
				MarkdownDescription: "Grouping values for aggregating costs on the KubernetesEfficiencyReport. Valid groupings: cluster_id, namespace, labeled, category, label, label:<label_name>.",
			},
			"start_date": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The start date of the KubernetesEfficiencyReport. ISO 8601 Formatted. Incompatible with 'date_interval' parameter.",
				MarkdownDescription: "The start date of the KubernetesEfficiencyReport. ISO 8601 Formatted. Incompatible with 'date_interval' parameter.",
			},
			"title": schema.StringAttribute{
				Required:            true,
				Description:         "The title of the KubernetesEfficiencyReport.",
				MarkdownDescription: "The title of the KubernetesEfficiencyReport.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the report",
				MarkdownDescription: "The token of the report",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the User who created this KubernetesEfficiencyReport.",
				MarkdownDescription: "The token for the User who created this KubernetesEfficiencyReport.",
			},
			"workspace_token": schema.StringAttribute{
				Required:            true,
				Description:         "The Workspace in which the KubernetesEfficiencyReport will be created.",
				MarkdownDescription: "The Workspace in which the KubernetesEfficiencyReport will be created.",
			},
		},
	}
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
