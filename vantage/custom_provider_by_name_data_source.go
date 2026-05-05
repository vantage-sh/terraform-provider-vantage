package vantage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	integrationsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/integrations"
)

var (
	_ datasource.DataSource              = (*customProviderByNameDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*customProviderByNameDataSource)(nil)
)

func NewCustomProviderByNameDataSource() datasource.DataSource {
	return &customProviderByNameDataSource{}
}

// customProviderByNameDataSourceModel is the config/state model for this data source.
// It embeds the same read-only fields returned by vantage_custom_provider so
// callers get a consistent shape regardless of which data source they use.
type customProviderByNameDataSourceModel struct {
	// Input fields
	Name           types.String `tfsdk:"name"`
	ProviderFilter types.String `tfsdk:"provider_filter"`

	// Output fields (same as customProviderDataSourceModel)
	Token                types.String `tfsdk:"token"`
	Status               types.String `tfsdk:"status"`
	CreatedAt            types.String `tfsdk:"created_at"`
	LastUpdated          types.String `tfsdk:"last_updated"`
	WorkspaceTokens      types.Set    `tfsdk:"workspace_tokens"`
	ManagedAccountTokens types.Set    `tfsdk:"managed_account_tokens"`
}

type customProviderByNameDataSource struct {
	client *Client
}

func (d *customProviderByNameDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *customProviderByNameDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_provider_by_name"
}

func (d *customProviderByNameDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a Custom Provider integration by name. Searches up to 1000 integrations returned by the Vantage API. Use `provider_filter` to narrow the search to a specific integration type.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The display name of the Custom Provider integration to find.",
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

func (d *customProviderByNameDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customProviderByNameDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limit := int32(1000)
	params := integrationsv2.NewGetIntegrationsParams()
	params.SetLimit(&limit)

	if !state.ProviderFilter.IsNull() && !state.ProviderFilter.IsUnknown() {
		p := state.ProviderFilter.ValueString()
		params.SetProvider(&p)
	}

	out, err := d.client.V2.Integrations.GetIntegrations(params, d.client.Auth)
	if err != nil {
		handleError("Read Custom Provider By Name", &resp.Diagnostics, err)
		return
	}

	target := state.Name.ValueString()
	for _, integration := range out.Payload.Integrations {
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
		"Custom Provider Not Found",
		fmt.Sprintf("No integration with name %q was found. If the integration exists, try omitting provider_filter or verify the name is correct.", target),
	)
}
