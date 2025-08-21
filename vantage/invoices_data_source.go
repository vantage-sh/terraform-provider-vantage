package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_invoices"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	invoicesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/invoices"
)

var _ datasource.DataSource = (*invoicesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*invoicesDataSource)(nil)

func NewInvoicesDataSource() datasource.DataSource {
	return &invoicesDataSource{}
}

type invoicesDataSource struct {
	client *Client
}

func (d *invoicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invoices"
}

func (d *invoicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_invoices.InvoicesDataSourceSchema(ctx)
}

func (d *invoicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *Client, got something else.",
		)
		return
	}

	d.client = client
}

func (d *invoicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_invoices.InvoicesModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to get invoices
	params := invoicesv2.NewGetInvoicesParams()
	result, err := d.client.V2.Invoices.GetInvoices(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Invoices",
			err.Error(),
		)
		return
	}

	// Convert API response to Terraform model
	var invoicesList []attr.Value
	for _, invoice := range result.Payload.Invoices {
		// Create an invoice value using the generated type
		invoiceValue, diag := datasource_invoices.NewInvoicesValue(
			datasource_invoices.InvoicesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"token":                 types.StringPointerValue(&invoice.Token),
				"total":                types.StringPointerValue(&invoice.Total),
				"status":               types.StringPointerValue(&invoice.Status),
				"billing_period_start": types.StringPointerValue(&invoice.BillingPeriodStart),
				"billing_period_end":   types.StringPointerValue(&invoice.BillingPeriodEnd),
				"created_at":           types.StringPointerValue(&invoice.CreatedAt),
				"updated_at":           types.StringPointerValue(&invoice.UpdatedAt),
				"account_name":         types.StringPointerValue(&invoice.AccountName),
				"account_token":        types.StringPointerValue(&invoice.AccountToken),
				"invoice_number":       types.StringPointerValue(&invoice.InvoiceNumber),
				"msp_account_token":    types.StringPointerValue(&invoice.MspAccountToken),
			},
		)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		invoicesList = append(invoicesList, invoiceValue)
	}
	
	// Create the list of invoices
	invoicesListValue, diag := types.ListValue(
		datasource_invoices.InvoicesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_invoices.InvoicesValue{}.AttributeTypes(ctx),
			},
		},
		invoicesList,
	)
	
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.Invoices = invoicesListValue

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type invoicesDataSourceModel struct {
	Invoices []*modelsv2.Invoice
}
