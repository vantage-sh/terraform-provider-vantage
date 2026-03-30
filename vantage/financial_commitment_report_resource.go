package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_financial_commitment_report"
	fcrv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/financial_commitment_reports"
)

var _ resource.Resource = (*financialCommitmentReportResource)(nil)
var _ resource.ResourceWithConfigure = (*financialCommitmentReportResource)(nil)
var _ resource.ResourceWithImportState = (*financialCommitmentReportResource)(nil)

type financialCommitmentReportResource struct {
	client *Client
}

func NewFinancialCommitmentReportResource() resource.Resource {
	return &financialCommitmentReportResource{}
}

func (r *financialCommitmentReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *financialCommitmentReportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *financialCommitmentReportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_financial_commitment_report"
}

func (r *financialCommitmentReportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_financial_commitment_report.FinancialCommitmentReportResourceSchema(ctx)
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

func (r *financialCommitmentReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data financialCommitmentReportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the planned groupings value to preserve empty lists
	plannedGroupings := data.Groupings

	model := data.toCreateModel(ctx)

	params := fcrv2.NewCreateFinancialCommitmentReportParams().WithCreateFinancialCommitmentReport(model)
	out, err := r.client.V2.FinancialCommitmentReports.CreateFinancialCommitmentReport(params, r.client.Auth)

	if err != nil {
		if e, ok := err.(*fcrv2.CreateFinancialCommitmentReportBadRequest); ok {
			handleBadRequest("Create FinancialCommitmentReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Create FinancialCommitmentReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// If the plan had an explicit empty list for groupings, preserve it
	// This prevents inconsistent state when the API returns default groupings
	if !plannedGroupings.IsNull() && !plannedGroupings.IsUnknown() && len(plannedGroupings.Elements()) == 0 {
		data.Groupings = plannedGroupings
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *financialCommitmentReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data financialCommitmentReportModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the current state groupings value to preserve empty lists
	stateGroupings := data.Groupings

	// Read API call logic
	params := fcrv2.NewGetFinancialCommitmentReportParams().WithFinancialCommitmentReportToken(data.Token.ValueString())
	out, err := r.client.V2.FinancialCommitmentReports.GetFinancialCommitmentReport(params, r.client.Auth)
	if err != nil {

		if _, ok := err.(*fcrv2.GetFinancialCommitmentReportNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Read FinancialCommitmentReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// If the previous state had an explicit empty list for groupings, preserve it
	// The API returns default groupings even when empty was specified
	if !stateGroupings.IsNull() && !stateGroupings.IsUnknown() && len(stateGroupings.Elements()) == 0 {
		data.Groupings = stateGroupings
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *financialCommitmentReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data financialCommitmentReportModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the planned groupings value to preserve empty lists
	plannedGroupings := data.Groupings

	model := data.toUpdateModel(ctx)

	params := fcrv2.NewUpdateFinancialCommitmentReportParams().WithUpdateFinancialCommitmentReport(model).WithFinancialCommitmentReportToken(data.Token.ValueString())

	out, err := r.client.V2.FinancialCommitmentReports.UpdateFinancialCommitmentReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*fcrv2.UpdateFinancialCommitmentReportBadRequest); ok {
			handleBadRequest("Update FinancialCommitmentReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update FinancialCommitmentReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// If the plan had an explicit empty list for groupings, preserve it
	// This prevents inconsistent state when the API returns default groupings
	if !plannedGroupings.IsNull() && !plannedGroupings.IsUnknown() && len(plannedGroupings.Elements()) == 0 {
		data.Groupings = plannedGroupings
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *financialCommitmentReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data financialCommitmentReportModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := fcrv2.NewDeleteFinancialCommitmentReportParams().WithFinancialCommitmentReportToken(data.Token.ValueString())

	_, err := r.client.V2.FinancialCommitmentReports.DeleteFinancialCommitmentReport(params, r.client.Auth)
	if err != nil {
		handleError("Delete FinancialCommitmentReport Resource", &resp.Diagnostics, err)
	}
}
