package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	costsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/costs"
)

var (
	_ datasource.DataSource              = &costReportsDataSource{}
	_ datasource.DataSourceWithConfigure = &costReportsDataSource{}
)

var costReportChartSettingsAttrTypes = map[string]attr.Type{
	"x_axis_dimension": types.ListType{ElemType: types.StringType},
	"y_axis_dimension": types.StringType,
}

func chartSettingsFromPayload(ctx context.Context, cs *modelsv2.ChartSettings) (basetypes.ObjectValue, error) {
	if cs == nil {
		return types.ObjectNull(costReportChartSettingsAttrTypes), nil
	}

	xAxisDimension, diags := types.ListValueFrom(ctx, types.StringType, cs.XAxisDimension)
	if diags.HasError() {
		return types.ObjectNull(costReportChartSettingsAttrTypes), fmt.Errorf("error converting x_axis_dimension")
	}

	attrValues := map[string]attr.Value{
		"x_axis_dimension": xAxisDimension,
		"y_axis_dimension": types.StringValue(cs.YAxisDimension),
	}

	obj, diags := types.ObjectValue(costReportChartSettingsAttrTypes, attrValues)
	if diags.HasError() {
		return types.ObjectNull(costReportChartSettingsAttrTypes), fmt.Errorf("error building chart_settings object")
	}
	return obj, nil
}

func NewCostReportsDataSource() datasource.DataSource {
	return &costReportsDataSource{}
}

type costReportsDataSource struct {
	client *Client
}

type costReportDataSourceModel struct {
	Token                   types.String `tfsdk:"token"`
	Title                   types.String `tfsdk:"title"`
	Filter                  types.String `tfsdk:"filter"`
	FolderToken             types.String `tfsdk:"folder_token"`
	WorkspaceToken          types.String `tfsdk:"workspace_token"`
	SavedFilterTokens       types.List   `tfsdk:"saved_filter_tokens"`
	Groupings               types.String `tfsdk:"groupings"`
	StartDate               types.String `tfsdk:"start_date"`
	EndDate                 types.String `tfsdk:"end_date"`
	PreviousPeriodStartDate types.String `tfsdk:"previous_period_start_date"`
	PreviousPeriodEndDate   types.String `tfsdk:"previous_period_end_date"`
	DateInterval            types.String `tfsdk:"date_interval"`
	ChartType               types.String `tfsdk:"chart_type"`
	DateBin                 types.String `tfsdk:"date_bin"`
	ChartSettings           types.Object `tfsdk:"chart_settings"`
}

type costReportsDataSourceModel struct {
	CostReports []costReportDataSourceModel `tfsdk:"cost_reports"`
}

func (d *costReportsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cost_reports"
}

func (d *costReportsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cost_reports": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed: true,
						},
						"title": schema.StringAttribute{
							Computed: true,
						},
						"filter": schema.StringAttribute{
							Computed: true,
						},
						"folder_token": schema.StringAttribute{
							Computed: true,
						},
						"saved_filter_tokens": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"workspace_token": schema.StringAttribute{
							Computed: true,
						},
						"groupings": schema.StringAttribute{
							Computed: true,
						},
						"start_date": schema.StringAttribute{
							Computed: true,
						},
						"end_date": schema.StringAttribute{
							Computed: true,
						},
						"previous_period_start_date": schema.StringAttribute{
							Computed: true,
						},
						"previous_period_end_date": schema.StringAttribute{
							Computed: true,
						},
						"date_interval": schema.StringAttribute{
							Computed: true,
						},
						"chart_type": schema.StringAttribute{
							Computed: true,
						},
						"date_bin": schema.StringAttribute{
							Computed: true,
						},
						"chart_settings": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"x_axis_dimension": schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
								"y_axis_dimension": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *costReportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state costReportsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	params := costsv2.NewGetCostReportsParams()
	out, err := d.client.V2.Costs.GetCostReports(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Costs",
			err.Error(),
		)
		return
	}

	costReports := []costReportDataSourceModel{}

	for _, r := range out.Payload.CostReports {
		savedFilterTokens, diag := types.ListValueFrom(ctx, types.StringType, r.SavedFilterTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		chartSettingsObj, csErr := chartSettingsFromPayload(ctx, r.ChartSettings)
		if csErr != nil {
			resp.Diagnostics.AddError("Error reading chart_settings", csErr.Error())
			return
		}
		costReports = append(costReports, costReportDataSourceModel{
			Title:                   types.StringValue(r.Title),
			Token:                   types.StringValue(r.Token),
			Filter:                  types.StringPointerValue(r.Filter),
			FolderToken:             types.StringPointerValue(r.FolderToken),
			WorkspaceToken:          types.StringValue(r.WorkspaceToken),
			SavedFilterTokens:       savedFilterTokens,
			Groupings:               ptrStringOrEmpty(r.Groupings),
			StartDate:               types.StringPointerValue(r.StartDate),
			EndDate:                 types.StringPointerValue(r.EndDate),
			PreviousPeriodStartDate: types.StringPointerValue(r.PreviousPeriodStartDate),
			PreviousPeriodEndDate:   types.StringPointerValue(r.PreviousPeriodEndDate),
			DateInterval:            types.StringValue(r.DateInterval),
			ChartType:               types.StringValue(r.ChartType),
			DateBin:                 types.StringValue(r.DateBin),
			ChartSettings:           chartSettingsObj,
		})
	}
	state.CostReports = costReports
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *costReportsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}
