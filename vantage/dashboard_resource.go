package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_dashboard"
	dashboardsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/dashboards"
)

var (
	_ resource.Resource                = (*DashboardResource)(nil)
	_ resource.ResourceWithConfigure   = (*DashboardResource)(nil)
	_ resource.ResourceWithImportState = (*DashboardResource)(nil)
)

type DashboardResource struct {
	client *Client
}

func NewDashboardResource() resource.Resource {
	return &DashboardResource{}
}

func (r *DashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

type NullableModifier struct{}

func (m *NullableModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {

	if req.ConfigValue.IsNull() && !req.StateValue.IsNull() {
		resp.PlanValue = types.StringUnknown()
	}

	// handle that the API sets the value of date_interval to "custom"
	if req.ConfigValue.IsNull() && req.StateValue.Equal(types.StringValue("custom")) {
		resp.PlanValue = types.StringValue("custom")
	}
}

func (NullableModifier) Description(_ context.Context) string {
	return "Custom plan modifier for handling nullable values"
}

func (NullableModifier) MarkdownDescription(_ context.Context) string {
	return "Custom plan modifier for handling nullable values"
}

func (r DashboardResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	var plan, state *dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state == nil || plan == nil {
		return
	}

	var configModel dashboardModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configHasDates := !configModel.StartDate.IsNull() && configModel.StartDate.ValueString() != "" &&
		!configModel.EndDate.IsNull() && configModel.EndDate.ValueString() != ""

	// If date_interval was preserved as "custom" by the NullableModifier but the
	// user removed dates from config, clear it so the API can reset the dashboard.
	if configModel.DateInterval.IsNull() && plan.DateInterval.ValueString() == "custom" && !configHasDates {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("date_interval"), types.StringNull())...)
		return
	}

	// If date_interval is removed from config (null in plan) but exists in state,
	// mark that an update is required
	if plan.DateInterval.IsNull() && !state.DateInterval.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("date_interval"), types.StringNull())...)
	}

}

func (r DashboardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_dashboard.DashboardResourceSchema(ctx)
	attrs := s.GetAttributes()

	s.Attributes["date_interval"] = schema.StringAttribute{
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
			&NullableModifier{},
		},
		MarkdownDescription: attrs["date_interval"].GetMarkdownDescription(),
	}

	s.Attributes["start_date"] = schema.StringAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: attrs["start_date"].GetMarkdownDescription(),
		Default:             stringdefault.StaticString(""),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	s.Attributes["end_date"] = schema.StringAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: attrs["end_date"].GetMarkdownDescription(),
		Default:             stringdefault.StaticString(""),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	s.Attributes["title"] = schema.StringAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: attrs["title"].GetMarkdownDescription(),
	}
	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	resp.Schema = s
}

func (r DashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardsv2.NewCreateDashboardParams().WithCreateDashboard(body)
	out, err := r.client.V2.Dashboards.CreateDashboard(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*dashboardsv2.CreateDashboardBadRequest); ok {
			handleBadRequest("Create Dashboard Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Dashboard Resource", &resp.Diagnostics, err)
		return
	}

	if diag := data.applyPayload(ctx, out.Payload); diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r DashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *dashboardModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardsv2.NewGetDashboardParams().WithDashboardToken(state.Token.ValueString())
	out, err := r.client.V2.Dashboards.GetDashboard(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*dashboardsv2.GetDashboardNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Dashboard Resource", &resp.Diagnostics, err)
		return
	}

	diag := state.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r DashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r DashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	params := dashboardsv2.NewUpdateDashboardParams().
		WithDashboardToken(data.Token.ValueString()).
		WithUpdateDashboard(body)

	out, err := r.client.V2.Dashboards.UpdateDashboard(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*dashboardsv2.UpdateDashboardBadRequest); ok {
			handleBadRequest("Update Dashboard Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Dashboard Resource", &resp.Diagnostics, err)
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r DashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *dashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardsv2.NewDeleteDashboardParams()
	params.SetDashboardToken(state.Token.ValueString())
	_, err := r.client.V2.Dashboards.DeleteDashboard(params, r.client.Auth)
	if err != nil {
		handleError("Delete Dashboard Resource", &resp.Diagnostics, err)
	}
}

// Configure adds the provider configured client to the data source.
func (r *DashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
