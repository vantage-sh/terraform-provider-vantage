package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/planmodifiers"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_scenario_model"
	scenariomodelsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/scenario_models"
)

var (
	_ resource.Resource                = (*scenarioModelResource)(nil)
	_ resource.ResourceWithConfigure   = (*scenarioModelResource)(nil)
	_ resource.ResourceWithImportState = (*scenarioModelResource)(nil)
)

func NewScenarioModelResource() resource.Resource {
	return &scenarioModelResource{}
}

type scenarioModelResource struct {
	client *Client
}

func (r *scenarioModelResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *scenarioModelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scenario_model"
}

func (r *scenarioModelResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_scenario_model.ScenarioModelResourceSchema(ctx)
	attrs := s.GetAttributes()

	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		Description:         attrs["token"].GetDescription(),
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	// API field is "provider"; Terraform reserves that root attribute name.
	s.Attributes["cloud_provider"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         "The cloud provider filter for the ScenarioModel.",
		MarkdownDescription: "The cloud provider filter for the ScenarioModel.",
		PlanModifiers: []planmodifier.String{
			planmodifiers.NullableString(),
		},
	}
	s.Attributes["service"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         attrs["service"].GetDescription(),
		MarkdownDescription: attrs["service"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			planmodifiers.NullableString(),
		},
	}
	s.Attributes["workspace_token"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         attrs["workspace_token"].GetDescription(),
		MarkdownDescription: attrs["workspace_token"].GetMarkdownDescription(),
	}

	resp.Schema = s
}

func (r *scenarioModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data scenarioModelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedPeriods := data.Periods
	model := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := scenariomodelsv2.NewCreateScenarioModelParams().WithCreateScenarioModel(model)
	out, err := r.client.V2.ScenarioModels.CreateScenarioModel(params, r.client.Auth)
	if handleAPIError("Create Scenario Model", &resp.Diagnostics, err, apiNotFoundError, nil) {
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	if !plannedPeriods.IsNull() && !plannedPeriods.IsUnknown() && len(plannedPeriods.Elements()) == 0 {
		data.Periods = plannedPeriods
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *scenarioModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data scenarioModelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statePeriods := data.Periods
	params := scenariomodelsv2.NewGetScenarioModelParams().WithScenarioModelToken(data.Token.ValueString())
	out, err := r.client.V2.ScenarioModels.GetScenarioModel(params, r.client.Auth)
	if handleAPIError("Read Scenario Model", &resp.Diagnostics, err, apiNotFoundRemove, func() {
		resp.State.RemoveResource(ctx)
	}) {
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	if !statePeriods.IsNull() && !statePeriods.IsUnknown() && len(statePeriods.Elements()) == 0 {
		data.Periods = statePeriods
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *scenarioModelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *scenarioModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data scenarioModelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedPeriods := data.Periods
	model := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := scenariomodelsv2.NewUpdateScenarioModelParams().
		WithScenarioModelToken(data.Token.ValueString()).
		WithUpdateScenarioModel(model)
	out, err := r.client.V2.ScenarioModels.UpdateScenarioModel(params, r.client.Auth)
	if handleAPIError("Update Scenario Model", &resp.Diagnostics, err, apiNotFoundError, nil) {
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	if !plannedPeriods.IsNull() && !plannedPeriods.IsUnknown() && len(plannedPeriods.Elements()) == 0 {
		data.Periods = plannedPeriods
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *scenarioModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data scenarioModelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := scenariomodelsv2.NewDeleteScenarioModelParams().WithScenarioModelToken(data.Token.ValueString())
	_, err := r.client.V2.ScenarioModels.DeleteScenarioModel(params, r.client.Auth)
	if handleAPIError("Delete Scenario Model", &resp.Diagnostics, err, apiNotFoundIgnore, nil) {
		return
	}

}
