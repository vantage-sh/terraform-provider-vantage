package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*integrationByNameDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*integrationByNameDataSource)(nil)
)

func NewIntegrationByNameDataSource() datasource.DataSource {
	return &integrationByNameDataSource{}
}

// integrationByNameDataSourceModel is the config/state model for the
// vantage_integration_by_name data source.
type integrationByNameDataSourceModel struct {
	// Input fields
	Name           types.String `tfsdk:"name"`
	ProviderFilter types.String `tfsdk:"provider_filter"`

	// Output fields
	Token                types.String `tfsdk:"token"`
	Status               types.String `tfsdk:"status"`
	CreatedAt            types.String `tfsdk:"created_at"`
	LastUpdated          types.String `tfsdk:"last_updated"`
	WorkspaceTokens      types.Set    `tfsdk:"workspace_tokens"`
	ManagedAccountTokens types.Set    `tfsdk:"managed_account_tokens"`
}

type integrationByNameDataSource struct {
	client *Client
}

func (d *integrationByNameDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *integrationByNameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_by_name"
}

func (d *integrationByNameDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up an integration by name. Searches up to 1,000 integrations returned by the Vantage API. Use `provider_filter` to narrow the search to a specific integration type.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the integration to find.",
			},
			"provider_filter": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter integrations by provider type before searching (e.g. `custom_provider`). Corresponds to the `provider` query parameter on the Get All Integrations endpoint.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique token of the matched integration.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the integration (e.g. connected, pending, importing, imported, error, disconnected).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time (UTC, ISO 8601) when the integration was created.",
			},
			"last_updated": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time (UTC, ISO 8601) when the integration was last updated.",
			},
			"workspace_tokens": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tokens of the Workspaces associated with this integration.",
			},
			"managed_account_tokens": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tokens of any Managed Accounts associated with this integration.",
			},
		},
	}
}

func (d *integrationByNameDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state integrationByNameDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var providerFilter *string
	if !state.ProviderFilter.IsNull() && !state.ProviderFilter.IsUnknown() {
		p := state.ProviderFilter.ValueString()
		providerFilter = &p
	}

	allIntegrations, err := fetchAllIntegrations(d.client, providerFilter)
	if err != nil {
		handleError("Read Integration By Name", &resp.Diagnostics, err)
		return
	}

	target := state.Name.ValueString()
	for _, integration := range allIntegrations {
		if integration.AccountIdentifier == nil || *integration.AccountIdentifier != target {
			continue
		}

		state.Token = types.StringValue(integration.Token)
		state.Status = types.StringValue(integration.Status)
		state.CreatedAt = types.StringValue(integration.CreatedAt)

		if integration.LastUpdated != nil {
			state.LastUpdated = types.StringValue(*integration.LastUpdated)
		} else {
			state.LastUpdated = types.StringNull()
		}

		workspaceTokens, diags := types.SetValueFrom(ctx, types.StringType, integration.WorkspaceTokens)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.WorkspaceTokens = workspaceTokens

		managedAccountTokens, diags := types.SetValueFrom(ctx, types.StringType, integration.ManagedAccountTokens)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.ManagedAccountTokens = managedAccountTokens

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	resp.Diagnostics.AddError(
		"Integration Not Found",
		fmt.Sprintf("No integration with name %q was found. If the integration exists, try omitting provider_filter or verify the name is correct.", target),
	)
}
