package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/planmodifiers"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_report_forecast"
	reportforecastsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/report_forecasts"
)

var (
	_ resource.Resource                = (*reportForecastResource)(nil)
	_ resource.ResourceWithConfigure   = (*reportForecastResource)(nil)
	_ resource.ResourceWithImportState = (*reportForecastResource)(nil)
)

func NewReportForecastResource() resource.Resource {
	return &reportForecastResource{}
}

type reportForecastResource struct {
	client *Client
}

func (r *reportForecastResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*Client)
}

func (r *reportForecastResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_forecast"
}

func (r *reportForecastResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_report_forecast.ReportForecastResourceSchema(ctx)
	attrs := s.GetAttributes()

	s.Attributes["token"] = schema.StringAttribute{
		Computed:            true,
		Description:         attrs["token"].GetDescription(),
		MarkdownDescription: attrs["token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	s.Attributes["cost_report_token"] = schema.StringAttribute{
		Required:            true,
		Description:         attrs["cost_report_token"].GetDescription(),
		MarkdownDescription: attrs["cost_report_token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
	s.Attributes["business_metric_token"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Description:         attrs["business_metric_token"].GetDescription(),
		MarkdownDescription: attrs["business_metric_token"].GetMarkdownDescription(),
		PlanModifiers: []planmodifier.String{
			planmodifiers.NullableString(),
		},
	}

	resp.Schema = s
}

func (r *reportForecastResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data reportForecastModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedTokens := data.ScenarioModelTokens
	plannedSetAsDefault := data.SetAsDefault
	model := data.toCreate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := reportforecastsv2.NewCreateReportForecastParams().WithCreateReportForecast(model)
	out, err := r.client.V2.ReportForecasts.CreateReportForecast(params, r.client.Auth)
	if handleAPIError("Create Report Forecast", &resp.Diagnostics, err, apiNotFoundError, nil) {
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	preserveReportForecastPlanCollections(&data, plannedTokens, plannedSetAsDefault)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *reportForecastResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data reportForecastModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateTokens := data.ScenarioModelTokens
	stateSetAsDefault := data.SetAsDefault
	params := reportforecastsv2.NewGetReportForecastParams().WithReportForecastToken(data.Token.ValueString())
	out, err := r.client.V2.ReportForecasts.GetReportForecast(params, r.client.Auth)
	if handleAPIError("Read Report Forecast", &resp.Diagnostics, err, apiNotFoundRemove, func() {
		resp.State.RemoveResource(ctx)
	}) {
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	preserveReportForecastPlanCollections(&data, stateTokens, stateSetAsDefault)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *reportForecastResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token"), req, resp)
}

func (r *reportForecastResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data reportForecastModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedTokens := data.ScenarioModelTokens
	plannedSetAsDefault := data.SetAsDefault
	model := data.toUpdate(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	params := reportforecastsv2.NewUpdateReportForecastParams().
		WithReportForecastToken(data.Token.ValueString()).
		WithUpdateReportForecast(model)
	out, err := r.client.V2.ReportForecasts.UpdateReportForecast(params, r.client.Auth)
	if handleAPIError("Update Report Forecast", &resp.Diagnostics, err, apiNotFoundError, nil) {
		return
	}

	diag := data.applyPayload(ctx, out.Payload)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	preserveReportForecastPlanCollections(&data, plannedTokens, plannedSetAsDefault)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *reportForecastResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data reportForecastModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := reportforecastsv2.NewDeleteReportForecastParams().WithReportForecastToken(data.Token.ValueString())
	_, err := r.client.V2.ReportForecasts.DeleteReportForecast(params, r.client.Auth)
	if handleAPIError("Delete Report Forecast", &resp.Diagnostics, err, apiNotFoundIgnore, nil) {
		return
	}

}

func preserveReportForecastPlanCollections(data *reportForecastModel, plannedTokens types.List, plannedSetAsDefault types.Bool) {
	if !plannedTokens.IsNull() && !plannedTokens.IsUnknown() && len(plannedTokens.Elements()) == 0 {
		data.ScenarioModelTokens = plannedTokens
	}
	if !plannedSetAsDefault.IsNull() && !plannedSetAsDefault.IsUnknown() {
		data.SetAsDefault = plannedSetAsDefault
	} else if !data.IsDefault.IsNull() && !data.IsDefault.IsUnknown() {
		data.SetAsDefault = data.IsDefault
	}
}
