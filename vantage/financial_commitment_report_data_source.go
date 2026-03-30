package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_financial_commitment_reports"
	fcrv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/financial_commitment_reports"
)

var (
	_ datasource.DataSource              = (*financialCommitmentReportsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = &financialCommitmentReportsDataSource{}
)

func NewFinancialCommitmentReportsDataSource() datasource.DataSource {
	return &financialCommitmentReportsDataSource{}
}

type financialCommitmentReportsDataSource struct {
	client *Client
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *financialCommitmentReportsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}

func (d *financialCommitmentReportsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_financial_commitment_reports"
}

func (d *financialCommitmentReportsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_financial_commitment_reports.FinancialCommitmentReportsDataSourceSchema(ctx)
}

type FinancialCommitmentReportModel struct {
	CreatedAt          types.String `tfsdk:"created_at"`
	DateBucket         types.String `tfsdk:"date_bucket"`
	DateInterval       types.String `tfsdk:"date_interval"`
	Default            types.Bool   `tfsdk:"default"`
	EndDate            types.String `tfsdk:"end_date"`
	Filter             types.String `tfsdk:"filter"`
	Groupings          types.String `tfsdk:"groupings"`
	OnDemandCostsScope types.String `tfsdk:"on_demand_costs_scope"`
	StartDate          types.String `tfsdk:"start_date"`
	Title              types.String `tfsdk:"title"`
	Token              types.String `tfsdk:"token"`
	Id                 types.String `tfsdk:"id"`
	UserToken          types.String `tfsdk:"user_token"`
	WorkspaceToken     types.String `tfsdk:"workspace_token"`
}

type financialCommitmentReportsDataSourceModel struct {
	FinancialCommitmentReports []FinancialCommitmentReportModel `tfsdk:"financial_commitment_reports"`
}

func (d *financialCommitmentReportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data financialCommitmentReportsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := fcrv2.NewGetFinancialCommitmentReportsParams()
	out, err := d.client.V2.FinancialCommitmentReports.GetFinancialCommitmentReports(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Financial Commitment Reports",
			err.Error(),
		)
		return
	}

	reports := []FinancialCommitmentReportModel{}
	for _, fcr := range out.Payload.FinancialCommitmentReports {
		report := FinancialCommitmentReportModel{
			CreatedAt:          types.StringValue(fcr.CreatedAt),
			DateBucket:         types.StringValue(fcr.DateBucket),
			DateInterval:       types.StringPointerValue(fcr.DateInterval),
			Default:            types.BoolValue(fcr.Default),
			EndDate:            types.StringPointerValue(fcr.EndDate),
			Filter:             types.StringPointerValue(fcr.Filter),
			Groupings:          types.StringPointerValue(fcr.Groupings),
			OnDemandCostsScope: types.StringValue(fcr.OnDemandCostsScope),
			StartDate:          types.StringPointerValue(fcr.StartDate),
			Title:              types.StringValue(fcr.Title),
			Token:              types.StringValue(fcr.Token),
			Id:                 types.StringValue(fcr.Token),
			UserToken:          types.StringPointerValue(fcr.UserToken),
			WorkspaceToken:     types.StringValue(fcr.WorkspaceToken),
		}
		reports = append(reports, report)
	}
	data.FinancialCommitmentReports = reports
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
