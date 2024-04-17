package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	filtersv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/saved_filters"
)

var (
	_ datasource.DataSource              = &savedFiltersDataSource{}
	_ datasource.DataSourceWithConfigure = &savedFiltersDataSource{}
)

func NewSavedFiltersDataSource() datasource.DataSource {
	return &savedFiltersDataSource{}
}

type savedFiltersDataSource struct {
	client *Client
}

type savedFilterDataSourceModel struct {
	Title            types.String `tfsdk:"title"`
	CostReportTokens types.List   `tfsdk:"cost_report_tokens"`
	Token            types.String `tfsdk:"token"`
	WorkspaceToken   types.String `tfsdk:"workspace_token"`
}

type savedFiltersDataSourceModel struct {
	Filters []savedFilterDataSourceModel `tfsdk:"filters"`
}

func (d *savedFiltersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saved_filters"
}

func (d *savedFiltersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"title": schema.StringAttribute{
							Computed: true,
						},
						"cost_report_tokens": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"token": schema.StringAttribute{
							Computed: true,
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

func (d *savedFiltersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state savedFiltersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	params := filtersv2.NewGetSavedFiltersParams()
	out, err := d.client.V2.SavedFilters.GetSavedFilters(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage SavedFilters",
			err.Error(),
		)
		return
	}

	filters := []savedFilterDataSourceModel{}
	for _, f := range out.Payload.SavedFilters {
		costReportTokens, diag := types.ListValueFrom(ctx, types.StringType, f.CostReportTokens)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		filter := savedFilterDataSourceModel{
			Title:            types.StringValue(f.Title),
			Token:            types.StringValue(f.Token),
			WorkspaceToken:   types.StringValue(f.WorkspaceToken),
			CostReportTokens: costReportTokens,
		}
		filters = append(filters, filter)
	}
	state.Filters = filters

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *savedFiltersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}
