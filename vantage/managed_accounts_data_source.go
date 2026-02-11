package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_managed_accounts"
	managedaccountsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/managed_accounts"
)

var _ datasource.DataSource = (*managedAccountsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = &managedAccountsDataSource{}

func NewManagedAccountsDataSource() datasource.DataSource {
	return &managedAccountsDataSource{}
}

type managedAccountsDataSource struct {
	client *Client
}

// managedAccountsDataSourceModelWrapper uses the data source-specific model
// that includes all fields from the generated schema.
type managedAccountsDataSourceModelWrapper struct {
	ManagedAccounts []managedAccountDataSourceModel `tfsdk:"managed_accounts"`
}

func (d *managedAccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *managedAccountsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_accounts"
}

func (d *managedAccountsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_managed_accounts.ManagedAccountsDataSourceSchema(ctx)
}

func (d *managedAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data managedAccountsDataSourceModelWrapper

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	params := managedaccountsv2.NewGetManagedAccountsParams()
	apiRes, err := d.client.V2.ManagedAccounts.GetManagedAccounts(params, d.client.Auth)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get Managed Accounts",
			err.Error(),
		)
		return
	}

	data.ManagedAccounts = []managedAccountDataSourceModel{}
	for _, m := range apiRes.Payload.ManagedAccounts {
		var model managedAccountDataSourceModel
		diag := model.applyPayloadDataSource(ctx, m)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.ManagedAccounts = append(data.ManagedAccounts, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
