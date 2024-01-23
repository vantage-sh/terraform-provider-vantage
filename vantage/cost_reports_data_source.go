package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	costsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/costs"
)

var (
	_ datasource.DataSource              = &costReportsDataSource{}
	_ datasource.DataSourceWithConfigure = &costReportsDataSource{}
)

func NewCostReportsDataSource() datasource.DataSource {
	return &costReportsDataSource{}
}

type costReportsDataSource struct {
	client *Client
}

type costReportDataSourceModel struct {
	Token             types.String `tfsdk:"token"`
	Title             types.String `tfsdk:"title"`
	Filter            types.String `tfsdk:"filter"`
	FolderToken       types.String `tfsdk:"folder_token"`
	WorkspaceToken    types.String `tfsdk:"workspace_token"`
	SavedFilterTokens types.List   `tfsdk:"saved_filter_tokens"`
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
		costReports = append(costReports, costReportDataSourceModel{
			Title:             types.StringValue(r.Title),
			Token:             types.StringValue(r.Token),
			Filter:            types.StringValue(r.Filter),
			FolderToken:       types.StringValue(r.FolderToken),
			WorkspaceToken:    types.StringValue(r.WorkspaceToken),
			SavedFilterTokens: savedFilterTokens,
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
