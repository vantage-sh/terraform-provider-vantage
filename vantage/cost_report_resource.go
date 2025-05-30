package vantage

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_cost_report"
	costsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/costs"
)

var (
	_ resource.Resource                = (*CostReportResource)(nil)
	_ resource.ResourceWithConfigure   = (*CostReportResource)(nil)
	_ resource.ResourceWithImportState = (*CostReportResource)(nil)
)

type CostReportResource struct {
	client *Client
}

func (r *CostReportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_cost_report.CostReportResourceSchema(ctx)
	attrs := s.GetAttributes()

	// Override the groupings field with a PlanModifier
	s.Attributes["groupings"] = schema.StringAttribute{
		Optional:            attrs["groupings"].IsOptional(),
		Computed:            attrs["groupings"].IsComputed(),
		MarkdownDescription: attrs["groupings"].GetMarkdownDescription(),
		// https://discuss.hashicorp.com/t/framework-migration-test-produces-non-empty-plan/54523/8
		Default: stringdefault.StaticString(""),
	}

	// Override the previous_period_start_date field with a PlanModifier
	s.Attributes["previous_period_start_date"] = schema.StringAttribute{
		Optional:            attrs["previous_period_start_date"].IsOptional(),
		Computed:            attrs["previous_period_start_date"].IsComputed(),
		MarkdownDescription: attrs["previous_period_start_date"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// Override the previous_period_end_date field with a PlanModifier
	s.Attributes["previous_period_end_date"] = schema.StringAttribute{
		Optional:            attrs["previous_period_end_date"].IsOptional(),
		Computed:            attrs["previous_period_end_date"].IsComputed(),
		MarkdownDescription: attrs["previous_period_end_date"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// Override the workspace_token field with a PlanModifier
	s.Attributes["workspace_token"] = schema.StringAttribute{
		Optional:            attrs["workspace_token"].IsOptional(),
		Computed:            attrs["workspace_token"].IsComputed(),
		MarkdownDescription: attrs["workspace_token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// Override the token field with a PlanModifier
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            attrs["token"].IsComputed(),
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	resp.Schema = s
}

func NewCostReportResource() resource.Resource {
	return &CostReportResource{}
}

func (r *CostReportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cost_report"
}

func (r CostReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *costReportModel 
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewCreateCostReportParams().WithCreateCostReport(body)
	out, err := r.client.V2.Costs.CreateCostReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*costsv2.CreateCostReportBadRequest); ok {
			handleBadRequest("Create Cost Report Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	if diag := data.applyPayload(ctx, out.Payload); diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *costReportModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewGetCostReportParams().WithCostReportToken(state.Token.ValueString())
	out, err := r.client.V2.Costs.GetCostReport(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*costsv2.GetCostReportNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	diag := state.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r CostReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r CostReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *costReportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	costsv2.NewUpdateCostReportParams()
	params := costsv2.NewUpdateCostReportParams().
		WithCostReportToken(data.Token.ValueString()).
		WithUpdateCostReport(body)

	out, err := r.client.V2.Costs.UpdateCostReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*costsv2.UpdateCostReportBadRequest); ok {
			handleBadRequest("Update Cost Report Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Update Cost Report Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *costReportModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewDeleteCostReportParams()
	params.SetCostReportToken(state.Token.ValueString())
	_, err := r.client.V2.Costs.DeleteCostReport(params, r.client.Auth)
	if err != nil {
		handleError("Delete Cost Report Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *CostReportResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r *CostReportResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("folder_token"),
			path.MatchRoot("workspace_token"),
		),
	}
}
