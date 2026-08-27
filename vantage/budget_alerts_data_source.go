package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

type budgetAlertDataSourceModel struct {
	BudgetTokens        types.List   `tfsdk:"budget_tokens"`
	CreatedAt           types.String `tfsdk:"created_at"`
	DurationInDays      types.Int64  `tfsdk:"duration_in_days"`
	Id                  types.String `tfsdk:"id"`
	IntegrationProvider types.String `tfsdk:"integration_provider"`
	PeriodToTrack       types.String `tfsdk:"period_to_track"`
	RecipientChannels   types.List   `tfsdk:"recipient_channels"`
	RecipientEmails     types.List   `tfsdk:"recipient_emails"`
	Threshold           types.Int64  `tfsdk:"threshold"`
	Token               types.String `tfsdk:"token"`
	UserToken           types.String `tfsdk:"user_token"`
	UserTokens          types.List   `tfsdk:"user_tokens"`
	WorkspaceToken      types.String `tfsdk:"workspace_token"`
}

type budgetAlertsDataSourceModel struct {
	BudgetAlerts   []budgetAlertDataSourceModel `tfsdk:"budget_alerts"`
	BudgetToken    types.String                 `tfsdk:"budget_token"`
	WorkspaceToken types.String                 `tfsdk:"workspace_token"`
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
	s := datasource_budget_alerts.BudgetAlertsDataSourceSchema(ctx)

	// Drop the generated CustomType so budgetAlertDataSourceModel can round-trip
	// through State.Set without AttributeTypes drift.
	alertsAttr := s.Attributes["budget_alerts"].(schema.ListNestedAttribute)
	durationAttr := alertsAttr.NestedObject.Attributes["duration_in_days"].(schema.Int64Attribute)
	durationAttr.Description = "The API's nullable integer duration. A null value means the full month; the vantage_budget_alert resource writes the same state as an empty string."
	durationAttr.MarkdownDescription = "The API's nullable integer duration. A null value means the full month; the `vantage_budget_alert` resource writes the same state as `\"\"`."
	alertsAttr.NestedObject.Attributes["duration_in_days"] = durationAttr

	userTokensAttr := alertsAttr.NestedObject.Attributes["user_tokens"].(schema.ListAttribute)
	userTokensAttr.Description = "The organization-user subset of recipients. Freeform verified-domain addresses are excluded; recipient_emails contains the complete address list."
	userTokensAttr.MarkdownDescription = userTokensAttr.Description
	alertsAttr.NestedObject.Attributes["user_tokens"] = userTokensAttr

	recipientEmailsAttr := alertsAttr.NestedObject.Attributes["recipient_emails"].(schema.ListAttribute)
	recipientEmailsAttr.Description = "The complete recipient address list, including organization users and freeform addresses on verified domains."
	recipientEmailsAttr.MarkdownDescription = recipientEmailsAttr.Description
	alertsAttr.NestedObject.Attributes["recipient_emails"] = recipientEmailsAttr

	alertsAttr.NestedObject = schema.NestedAttributeObject{
		Attributes: alertsAttr.NestedObject.Attributes,
	}
	s.Attributes["budget_alerts"] = alertsAttr

	resp.Schema = s
}

func (d *budgetAlertsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data budgetAlertsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payloads, err := fetchAllBudgetAlerts(d.client, data.BudgetToken.ValueStringPointer(), data.WorkspaceToken.ValueStringPointer())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Get Vantage Budget Alerts", err.Error())
		return
	}

	alerts := make([]budgetAlertDataSourceModel, 0, len(payloads))
	for _, payload := range payloads {
		alert := budgetAlertDataSourceModel{
			Token:               types.StringValue(payload.Token),
			Id:                  types.StringValue(payload.Token),
			CreatedAt:           types.StringValue(payload.CreatedAt),
			DurationInDays:      types.Int64PointerValue(int64PointerFromInt32(payload.DurationInDays)),
			IntegrationProvider: types.StringPointerValue(payload.IntegrationProvider),
			PeriodToTrack:       types.StringPointerValue(payload.PeriodToTrack),
			Threshold:           types.Int64Value(int64(payload.Threshold)),
			UserToken:           types.StringPointerValue(payload.UserToken),
			WorkspaceToken:      types.StringPointerValue(payload.WorkspaceToken),
		}

		for _, list := range []struct {
			dst *types.List
			src []string
		}{
			{&alert.BudgetTokens, payload.BudgetTokens},
			{&alert.RecipientChannels, payload.RecipientChannels},
			{&alert.RecipientEmails, payload.RecipientEmails},
			{&alert.UserTokens, payload.UserTokens},
		} {
			value, diag := stringListValueOrEmpty(ctx, list.src)
			resp.Diagnostics.Append(diag...)
			if resp.Diagnostics.HasError() {
				return
			}
			*list.dst = value
		}

		alerts = append(alerts, alert)
	}

	data.BudgetAlerts = alerts
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func int64PointerFromInt32(src *int32) *int64 {
	if src == nil {
		return nil
	}
	value := int64(*src)
	return &value
}

// fetchAllBudgetAlerts pages through the Get All Budget Alerts endpoint until
// links.next is nil, collecting every alert across all pages.
func fetchAllBudgetAlerts(client *Client, budgetToken, workspaceToken *string) ([]*modelsv2.BudgetAlert, error) {
	limit := int32(1000)
	var all []*modelsv2.BudgetAlert
	var page *int32

	for {
		params := budgetalertsv2.NewGetBudgetAlertsParams()
		params.SetLimit(&limit)
		if budgetToken != nil {
			params.SetBudgetToken(budgetToken)
		}
		if workspaceToken != nil {
			params.SetWorkspaceToken(workspaceToken)
		}
		if page != nil {
			params.SetPage(page)
		}

		out, err := client.V2.BudgetAlerts.GetBudgetAlerts(params, client.Auth)
		if err != nil {
			return nil, err
		}

		all = append(all, out.Payload.BudgetAlerts...)

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
