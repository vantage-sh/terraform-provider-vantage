package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	dashboardsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/dashboards"
)

type DashboardResource struct {
	client *Client
}

func NewDashboardResource() resource.Resource {
	return &DashboardResource{}
}

type DashboardResourceModel struct {
	Token          types.String `tfsdk:"token"`
	Title          types.String `tfsdk:"title"`
	WidgetTokens   types.List   `tfsdk:"widget_tokens"`
	DateBin        types.String `tfsdk:"date_bin"`
	DateInterval   types.String `tfsdk:"date_interval"`
	StartDate      types.String `tfsdk:"start_date"`
	EndDate        types.String `tfsdk:"end_date"`
	WorkspaceToken types.String `tfsdk:"workspace_token"`
}

func (r *DashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r DashboardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique dashboard identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the dashboard",
				Required:            true,
			},
			"widget_tokens": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Tokens for widgets present in the dashboard. Currently only cost report tokens are supported.",
				Required:            true,
			},
			"date_bin": schema.StringAttribute{
				MarkdownDescription: "Determines how to group costs in the Dashboard.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"date_interval": schema.StringAttribute{
				MarkdownDescription: "Determines the date range in the Dashboard. Guaranteed to be set to 'custom' if 'start_date' and 'end_date' are set.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The start date for the date range for CostReports in the Dashboard. ISO 8601 Formatted. Overwrites 'date_interval' if set.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"end_date": schema.StringAttribute{
				MarkdownDescription: "The end date for the date range for CostReports in the Dashboard. ISO 8601 Formatted. Overwrites 'date_interval' if set.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_token": schema.StringAttribute{
				MarkdownDescription: "The token for the Workspace the Dashboard is a part of.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		MarkdownDescription: "Manages a Dashboard.",
	}
}

func (r DashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	widgetTokens := []types.String{}
	if !data.WidgetTokens.IsNull() && !data.WidgetTokens.IsUnknown() {
		widgetTokens = make([]types.String, 0, len(data.WidgetTokens.Elements()))
		resp.Diagnostics.Append(data.WidgetTokens.ElementsAs(ctx, &widgetTokens, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	params := dashboardsv2.NewCreateDashboardParams()
	body := &modelsv2.PostDashboards{
		Title:          data.Title.ValueStringPointer(),
		WidgetTokens:   fromStringsValue(widgetTokens),
		DateBin:        data.DateBin.ValueString(),
		DateInterval:   data.DateInterval.ValueString(),
		StartDate:      data.StartDate.ValueString(),
		EndDate:        data.EndDate.ValueStringPointer(),
		WorkspaceToken: data.WorkspaceToken.ValueString(),
	}
	params.WithDashboards(body)
	out, err := r.client.V2.Dashboards.CreateDashboard(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*dashboardsv2.CreateDashboardBadRequest); ok {
			handleBadRequest("Create Dashboard Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Dashboard Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Title = types.StringValue(out.Payload.Title)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	data.StartDate = types.StringValue(out.Payload.StartDate)
	data.EndDate = types.StringValue(out.Payload.EndDate)
	data.DateBin = types.StringValue(out.Payload.DateBin)
	data.DateInterval = types.StringValue(out.Payload.DateInterval)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r DashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *DashboardResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := dashboardsv2.NewGetDashboardParams()
	params.SetDashboardToken(state.Token.ValueString())
	out, err := r.client.V2.Dashboards.GetDashboard(params, r.client.Auth)
	if err != nil {
		if _, ok := err.(*dashboardsv2.GetDashboardNotFound); ok {
			resp.State.RemoveResource(ctx)
			return
		}

		handleError("Get Dashboard Resource", &resp.Diagnostics, err)
		return
	}

	state.Token = types.StringValue(out.Payload.Token)
	state.Title = types.StringValue(out.Payload.Title)
	state.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	state.StartDate = types.StringValue(out.Payload.StartDate)
	state.EndDate = types.StringValue(out.Payload.EndDate)
	state.DateBin = types.StringValue(out.Payload.DateBin)
	state.DateInterval = types.StringValue(out.Payload.DateInterval)
	widgets, diag := types.ListValueFrom(ctx, types.StringType, out.Payload.WidgetTokens)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	state.WidgetTokens = widgets

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r DashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	widgetTokens := []types.String{}
	if !data.WidgetTokens.IsNull() && !data.WidgetTokens.IsUnknown() {
		widgetTokens = make([]types.String, 0, len(data.WidgetTokens.Elements()))
		resp.Diagnostics.Append(data.WidgetTokens.ElementsAs(ctx, &widgetTokens, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	params := dashboardsv2.NewUpdateDashboardParams()
	body := &modelsv2.PutDashboards{
		Title:        data.Title.ValueString(),
		WidgetTokens: fromStringsValue(widgetTokens),
		DateBin:      data.DateBin.ValueString(),
	}
	if data.DateInterval.ValueString() == "" || data.DateInterval.ValueString() == "custom" {
		body.StartDate = data.StartDate.ValueString()
		body.EndDate = data.EndDate.ValueStringPointer()
	} else {
		body.DateInterval = data.DateInterval.ValueString()
	}
	params.WithDashboardToken(data.Token.ValueString())
	params.WithDashboards(body)
	out, err := r.client.V2.Dashboards.UpdateDashboard(params, r.client.Auth)
	if err != nil {
		if e, ok := err.(*dashboardsv2.UpdateDashboardBadRequest); ok {
			handleBadRequest("Update Dashboard Resource", &resp.Diagnostics, e.GetPayload())
			return
		}
		handleError("Create Dashboard Resource", &resp.Diagnostics, err)
		return
	}

	data.Token = types.StringValue(out.Payload.Token)
	data.Title = types.StringValue(out.Payload.Title)
	data.WorkspaceToken = types.StringValue(out.Payload.WorkspaceToken)
	data.StartDate = types.StringValue(out.Payload.StartDate)
	data.EndDate = types.StringValue(out.Payload.EndDate)
	data.DateBin = types.StringValue(out.Payload.DateBin)
	data.DateInterval = types.StringValue(out.Payload.DateInterval)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r DashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *DashboardResourceModel
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
