package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	s := datasource_budgets.BudgetsDataSourceSchema(ctx)
	budgetsAttr := s.Attributes["budgets"].(schema.ListNestedAttribute)
	attrs := make(map[string]schema.Attribute, len(budgetsAttr.NestedObject.Attributes)+1)
	for k, v := range budgetsAttr.NestedObject.Attributes {
		attrs[k] = v
	}
	attrs["period_cadence"] = schema.SingleNestedAttribute{
		Computed:            true,
		Description:         "The interval cadence for budget periods.",
		MarkdownDescription: "The interval cadence for budget periods.",
		Attributes: map[string]schema.Attribute{
			"starts_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The anchor date for budget period intervals. ISO 8601 date (YYYY-MM-DD).",
				MarkdownDescription: "The anchor date for budget period intervals. ISO 8601 date (YYYY-MM-DD).",
			},
			"interval_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of interval units per budget period.",
				MarkdownDescription: "The number of interval units per budget period.",
			},
			"interval_unit": schema.StringAttribute{
				Computed:            true,
				Description:         "The unit for budget period intervals. One of: day, week, month, year.",
				MarkdownDescription: "The unit for budget period intervals. One of: day, week, month, year.",
			},
		},
	}
	// Drop the generated CustomType so the shared budgetModel (with period_cadence)
	// can round-trip through State.Set without AttributeTypes drift.
	budgetsAttr.NestedObject = schema.NestedAttributeObject{
		Attributes: attrs,
	}
	s.Attributes["budgets"] = budgetsAttr
	resp.Schema = s
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
		diag := applyBudgetPayload(ctx, true, budget, &model)
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
