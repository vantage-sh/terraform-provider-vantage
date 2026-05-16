package vantage

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	integrationsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/integrations"
)

var (
	_ datasource.DataSource              = (*integrationsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*integrationsDataSource)(nil)
)

func NewIntegrationsDataSource() datasource.DataSource {
	return &integrationsDataSource{}
}

type integrationItemModel struct {
	Token                types.String `tfsdk:"token"`
	Name                 types.String `tfsdk:"name"`
	Status               types.String `tfsdk:"status"`
	CreatedAt            types.String `tfsdk:"created_at"`
	LastUpdated          types.String `tfsdk:"last_updated"`
	WorkspaceTokens      types.Set    `tfsdk:"workspace_tokens"`
	ManagedAccountTokens types.Set    `tfsdk:"managed_account_tokens"`
}

type integrationsDataSourceModel struct {
	ProviderFilter  types.String              `tfsdk:"provider_filter"`
	Integrations []integrationItemModel `tfsdk:"integrations"`
}

type integrationsDataSource struct {
	client *Client
}

func (d *integrationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *integrationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integrations"
}

func (d *integrationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"token": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique token of the integration.",
		},
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The display name of the integration.",
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
			MarkdownDescription: "The date and time (UTC, ISO 8601) when the integration was last updated. Null if never updated.",
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
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns all integrations visible to the API token, optionally filtered by provider type. Fetches up to 1,000 results.",
		Attributes: map[string]schema.Attribute{
			"provider_filter": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter results by provider type (e.g. `custom_provider`). Corresponds to the `provider` query parameter on the Get All Integrations API endpoint.",
			},
			"integrations": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The list of integrations returned by the API.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

func (d *integrationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state integrationsDataSourceModel
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
		handleError("Read Integrations", &resp.Diagnostics, err)
		return
	}

	state.Integrations = make([]integrationItemModel, 0, len(allIntegrations))
	for _, integration := range allIntegrations {
		item := integrationItemModel{
			Token:     types.StringValue(integration.Token),
			Status:    types.StringValue(integration.Status),
			CreatedAt: types.StringValue(integration.CreatedAt),
		}

		if integration.AccountIdentifier != nil {
			item.Name = types.StringValue(*integration.AccountIdentifier)
		} else {
			item.Name = types.StringNull()
		}

		if integration.LastUpdated != nil {
			item.LastUpdated = types.StringValue(*integration.LastUpdated)
		} else {
			item.LastUpdated = types.StringNull()
		}

		workspaceTokens, diags := types.SetValueFrom(ctx, types.StringType, integration.WorkspaceTokens)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		item.WorkspaceTokens = workspaceTokens

		managedAccountTokens, diags := types.SetValueFrom(ctx, types.StringType, integration.ManagedAccountTokens)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		item.ManagedAccountTokens = managedAccountTokens

		state.Integrations = append(state.Integrations, item)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// fetchAllIntegrations pages through the Get All Integrations endpoint until
// links.next is nil, collecting every integration across all pages.
func fetchAllIntegrations(client *Client, providerFilter *string) ([]*modelsv2.Integration, error) {
	limit := int32(1000)
	var all []*modelsv2.Integration
	var page *int32

	for {
		params := integrationsv2.NewGetIntegrationsParams()
		params.SetLimit(&limit)
		if providerFilter != nil {
			params.SetProvider(providerFilter)
		}
		if page != nil {
			params.SetPage(page)
		}

		out, err := client.V2.Integrations.GetIntegrations(params, client.Auth)
		if err != nil {
			return nil, err
		}

		all = append(all, out.Payload.Integrations...)

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

// pageFromURL extracts the "page" query parameter from a pagination link URL.
func pageFromURL(rawURL string) (int32, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0, err
	}
	pageStr := u.Query().Get("page")
	if pageStr == "" {
		return 0, fmt.Errorf("no page parameter in URL")
	}
	n, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("page parameter %q is not an integer: %w", pageStr, err)
	}
	return int32(n), nil
}
