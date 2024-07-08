package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_billing_rules"
	billingrulesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/billing_rules"
)

var _ datasource.DataSource = (*billingRulesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*billingRulesDataSource)(nil)

type billingRulesDataSourceModel struct {
	BillingRules []datasourceBillingRuleModel `tfsdk:"billing_rules"`
}

func NewBillingRulesDataSource() datasource.DataSource {
	return &billingRulesDataSource{}
}

type billingRulesDataSource struct {
	client *Client
}

func (d *billingRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *billingRulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_rules"
}

func (d *billingRulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_billing_rules.BillingRulesDataSourceSchema(ctx)
}

func (d *billingRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data billingRulesDataSourceModel
	// var data datasource_billing_rules.BillingRulesModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	params := billingrulesv2.NewGetBillingRulesParams()
	apiRes, err := d.client.V2.BillingRules.GetBillingRules(params, d.client.Auth)

	// Example data value setting
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Billing Rules",
			err.Error(),
		)
		return
	}

	for _, billingRule := range apiRes.Payload.BillingRules {
		var model billingRuleModel
		diag := model.applyPayload(ctx, billingRule)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		data.BillingRules = append(data.BillingRules, model.toDatasourceModel())
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
