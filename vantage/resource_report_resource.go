package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resourcereportsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/resource_reports"
)

var (
	_ resource.Resource                = (*resourceReportResource)(nil)
	_ resource.ResourceWithConfigure   = (*resourceReportResource)(nil)
	_ resource.ResourceWithImportState = (*resourceReportResource)(nil)
)

func NewResourceReportResource() resource.Resource {
	return &resourceReportResource{}
}

type resourceReportResource struct {
	client *Client
}

func (r *resourceReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)

}

func (r *resourceReportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *resourceReportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_report"
}

func (r *resourceReportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"columns": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Array of column names to display in the table. Column names should match those returned by the /resource_reports/columns endpoint. The order determines the display order. Only available for reports with a single resource type filter.",
				MarkdownDescription: "Array of column names to display in the table. Column names should match those returned by the /resource_reports/columns endpoint. The order determines the display order. Only available for reports with a single resource type filter.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
				MarkdownDescription: "The date and time, in UTC, the report was created. ISO 8601 Formatted.",
			},
			"created_by_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the User or Team who created this ResourceReport.",
				MarkdownDescription: "The token for the User or Team who created this ResourceReport.",
			},
			"filter": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The VQL filter for the ResourceReport.",
				MarkdownDescription: "The VQL filter for the ResourceReport.",
			},
			"title": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The title of the ResourceReport.",
				MarkdownDescription: "The title of the ResourceReport.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the report",
				MarkdownDescription: "The token of the report",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The token of the report",
				MarkdownDescription: "The token of the report",
				// PlanModifiers: []planmodifier.String{
				// 	stringplanmodifier.UseStateForUnknown(),
				// },
			},
			"user_token": schema.StringAttribute{
				Computed:            true,
				Description:         "The token for the User who created this ResourceReport.",
				MarkdownDescription: "The token for the User who created this ResourceReport.",
			},
			"workspace_token": schema.StringAttribute{
				Required:            true,
				Description:         "The token of the Workspace to add the ResourceReport to.",
				MarkdownDescription: "The token of the Workspace to add the ResourceReport to.",
			},
		},
	}
}

func (r *resourceReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceReportModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toCreateModel()

	params := resourcereportsv2.NewCreateResourceReportParams().WithCreateResourceReport(model)
	out, err := r.client.V2.ResourceReports.CreateResourceReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*resourcereportsv2.CreateResourceReportBadRequest); ok {
			handleBadRequest("Create ResourceReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create ResourceReport Resource", &resp.Diagnostics, err)
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

func (r *resourceReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceReportModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := resourcereportsv2.NewGetResourceReportParams().WithResourceReportToken(data.Token.ValueString())
	out, err := r.client.V2.ResourceReports.GetResourceReport(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*resourcereportsv2.GetResourceReportNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Read ResourceReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resourceReportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toUpdateModel()

	params := resourcereportsv2.NewUpdateResourceReportParams().WithUpdateResourceReport(model).WithResourceReportToken(data.Token.ValueString())

	out, err := r.client.V2.ResourceReports.UpdateResourceReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*resourcereportsv2.UpdateResourceReportBadRequest); ok {
			handleBadRequest("Update ResourceReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update ResourceReport Resource", &resp.Diagnostics, err)
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

func (r *resourceReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *resourceReportModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := resourcereportsv2.NewDeleteResourceReportParams().WithResourceReportToken(data.Token.ValueString())

	_, err := r.client.V2.ResourceReports.DeleteResourceReport(params, r.client.Auth)
	if err != nil {
		handleError("Delete ResourceReport Resource", &resp.Diagnostics, err)
		return
	}
}
