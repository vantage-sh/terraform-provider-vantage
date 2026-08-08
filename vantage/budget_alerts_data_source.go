package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_budget_alerts"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	budgetalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/budget_alerts"
)

var (
	_ datasource.DataSource              = (*budgetAlertsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*budgetAlertsDataSource)(nil)
)

func NewBudgetAlertsDataSource() datasource.DataSource {
	return &budgetAlertsDataSource{}
}

type budgetAlertsDataSource struct {
	client *Client
}

type budgetAlertsDataSourceModel struct {
	BudgetAlerts []budgetAlertDataSourceModel `tfsdk:"budget_alerts"`
}

func (d *budgetAlertsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *budgetAlertsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_budget_alerts"
}

func (d *budgetAlertsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_budget_alerts.BudgetAlertsDataSourceSchema(ctx)
}

func (d *budgetAlertsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data budgetAlertsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alerts, err := fetchAllBudgetAlerts(d.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Get Vantage Budget Alerts", err.Error())
		return
	}

	for _, alert := range alerts {
		if alert == nil {
			continue
		}

		var model budgetAlertDataSourceModel
		diags := model.applyPayloadDataSource(ctx, alert)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.BudgetAlerts = append(data.BudgetAlerts, model)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// fetchAllBudgetAlerts pages through the Get All Budget Alerts endpoint until
// links.next is nil, collecting every budget alert across all pages.
func fetchAllBudgetAlerts(client *Client) ([]*modelsv2.BudgetAlert, error) {
	limit := int32(1000)
	var all []*modelsv2.BudgetAlert
	var page *int32

	for {
		params := budgetalertsv2.NewGetBudgetAlertsParams()
		params.SetLimit(&limit)
		if page != nil {
			params.SetPage(page)
		}

		out, err := client.V2.BudgetAlerts.GetBudgetAlerts(params, client.Auth)
		if err != nil {
			return nil, err
		}

		all = append(all, out.Payload.BudgetAlerts...)

		// Stop when there is no next page.
		if out.Payload.Links == nil || out.Payload.Links.Next == nil {
			break
		}

		nextPage, err := pageFromURL(*out.Payload.Links.Next)
		if err != nil {
			return nil, fmt.Errorf("parsing next page from links.next %q: %w", *out.Payload.Links.Next, err)
		}
		page = &nextPage
	}

	return all, nil
}
