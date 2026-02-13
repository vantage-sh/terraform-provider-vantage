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
	_ resource.Resource                     = (*CostReportResource)(nil)
	_ resource.ResourceWithConfigure        = (*CostReportResource)(nil)
	_ resource.ResourceWithImportState      = (*CostReportResource)(nil)
	_ resource.ResourceWithConfigValidators = (*CostReportResource)(nil)
)

type CostReportResource struct {
	client *Client
}

func NewCostReportResource() resource.Resource {
	return &CostReportResource{}
}

func (r *CostReportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cost_report"
}

func (r CostReportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_cost_report.CostReportResourceSchema(ctx)
	attrs := s.GetAttributes()

	// Attribute overrides:
	// - token, id, workspace_token, previous_period_start_date: the code generator
	//   does not emit PlanModifiers, so we add UseStateForUnknown to prevent these
	//   Computed fields from showing "(known after apply)" on every plan.
	// - end_date, previous_period_end_date: the swagger spec marks these as required
	//   because of Grape's `given`/`requires` pattern (conditionally required when
	//   start_date or previous_period_start_date is provided), but grape-swagger
	//   loses the conditional context and emits them as unconditionally required.
	//   They are actually optional, so we override them to Optional+Computed.

	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		Description:         attrs["token"].GetDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["id"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["id"].GetMarkdownDescription(),
		Description:         attrs["id"].GetDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["previous_period_start_date"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: attrs["previous_period_start_date"].GetMarkdownDescription(),
		Description:         attrs["previous_period_start_date"].GetDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["previous_period_end_date"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: attrs["previous_period_end_date"].GetMarkdownDescription(),
		Description:         attrs["previous_period_end_date"].GetDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["end_date"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: attrs["end_date"].GetMarkdownDescription(),
		Description:         attrs["end_date"].GetDescription(),
	}

	s.Attributes["workspace_token"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: attrs["workspace_token"].GetMarkdownDescription(),
		Description:         attrs["workspace_token"].GetDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// Override groupings to default to empty string so clearing it works properly
	s.Attributes["groupings"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: attrs["groupings"].GetMarkdownDescription(),
		Description:         attrs["groupings"].GetDescription(),
		Default:             stringdefault.StaticString(""),
	}

	s.MarkdownDescription = "Manages a CostReport."
	resp.Schema = s
}

func (r CostReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data costReportModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toCreateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewCreateCostReportParams().WithCreateCostReport(model)

	out, err := r.client.V2.Costs.CreateCostReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*costsv2.CreateCostReportBadRequest); ok {
			handleBadRequest("Create CostReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create CostReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data costReportModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewGetCostReportParams().WithCostReportToken(data.Token.ValueString())
	out, err := r.client.V2.Costs.GetCostReport(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*costsv2.GetCostReportNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Read CostReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r CostReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data costReportModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := data.toUpdateModel(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewUpdateCostReportParams().WithUpdateCostReport(model).WithCostReportToken(data.Token.ValueString())

	out, err := r.client.V2.Costs.UpdateCostReport(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*costsv2.UpdateCostReportBadRequest); ok {
			handleBadRequest("Update CostReport Resource", &resp.Diagnostics, e.GetPayload())
			return
		}

		handleError("Update CostReport Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r CostReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data costReportModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := costsv2.NewDeleteCostReportParams().WithCostReportToken(data.Token.ValueString())

	_, err := r.client.V2.Costs.DeleteCostReport(params, r.client.Auth)
	if err != nil {
		handleError("Delete CostReport Resource", &resp.Diagnostics, err)
		return
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
