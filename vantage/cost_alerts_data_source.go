package vantage

import (
	"context"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_cost_alerts"
	costalertsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/cost_alerts"
)

var (
	_ datasource.DataSource              = (*costAlertsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*costAlertsDataSource)(nil)
)

func NewCostAlertsDataSource() datasource.DataSource {
	return &costAlertsDataSource{}
}

type costAlertsDataSource struct {
	client *Client
}

func (d *costAlertsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cost_alerts"
}

func (d *costAlertsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cost_alerts.CostAlertsDataSourceSchema(ctx)
}

func (d *costAlertsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if client, ok := req.ProviderData.(*Client); ok {
		d.client = client
	}
}

type costAlertsDataSourceModel struct {
	CostAlerts []costAlertDataSourceValue `tfsdk:"cost_alerts"`
}

func (d *costAlertsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	params := costalertsv2.NewGetCostAlertsParams()
	out, err := d.client.V2.CostAlerts.GetCostAlerts(params, d.client.Auth)

	if err != nil {
		resp.Diagnostics.AddError("Unable to Get Vantage Cost Alerts", err.Error())
        return
	}

	var alerts []costAlertDataSourceValue
	for _, alert := range out.Payload.CostAlerts {
		emailRecipients, diag := types.ListValueFrom(ctx, types.StringType, alert.EmailRecipients)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		slackChannels, diag := types.ListValueFrom(ctx, types.StringType, alert.SlackChannels)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		teamsChannels, diag := types.ListValueFrom(ctx, types.StringType, alert.TeamsChannels)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		reportTokens, diag := types.ListValueFrom(ctx, types.StringType, alert.ReportTokens)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		var minimumThreshold basetypes.NumberValue
		if alert.MinimumThreshold != nil {
			minimumThreshold = basetypes.NewNumberValue(big.NewFloat(float64(*alert.MinimumThreshold)))
		} else {
			minimumThreshold = basetypes.NewNumberNull()
		}

		alerts = append(alerts, costAlertDataSourceValue{
			Token:            types.StringValue(alert.Token),
			Title:            types.StringValue(alert.Title),
			Interval:         types.StringValue(alert.Interval),
			Threshold:        basetypes.NewNumberValue(big.NewFloat(alert.Threshold)),
			MinimumThreshold: minimumThreshold,
			UnitType:         types.StringValue(alert.UnitType),
			EmailRecipients:  emailRecipients,
			SlackChannels:    slackChannels,
			TeamsChannels:    teamsChannels,
			ReportTokens:     reportTokens,
		})
	}

	var state costAlertsDataSourceModel
	state.CostAlerts = alerts

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
