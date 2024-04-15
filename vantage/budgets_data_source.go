package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_budgets"
	budgetsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/budgets"
)

var _ datasource.DataSource = (*budgetsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*budgetsDataSource)(nil)

func NewBudgetsDataSource() datasource.DataSource {
	return &budgetsDataSource{}
}

type budgetsDataSource struct {
	client *Client
}

type budgetsDataSourceModel struct {
	Budgets []budgetModel `tfsdk:"budgets"`
}

func (d *budgetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}
func (d *budgetsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_budgets"
}

func (d *budgetsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_budgets.BudgetsDataSourceSchema(ctx)
}

func (d *budgetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data budgetsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := budgetsv2.NewGetBudgetsParams()
	out, err := d.client.V2.Budgets.GetBudgets(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Budgets",
			err.Error(),
		)
		return
	}
	budgets := []budgetModel{}
	for _, budget := range out.Payload.Budgets {
		model := budgetModel{}
		diag := applyBudgetPayload(ctx, budget, &model)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		budgets = append(budgets, model)
	}

	data.Budgets = budgets

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
