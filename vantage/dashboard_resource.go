package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

func (r DashboardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_dashboard.DashboardResourceSchema(ctx)
	attrs := s.GetAttributes()

	s.Attributes["date_interval"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
		MarkdownDescription: attrs["end_date"].GetMarkdownDescription(),
	}

	s.Attributes["end_date"] = schema.StringAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: attrs["end_date"].GetMarkdownDescription(),
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
