package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	integrationsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/integrations"
)

var (
	_ datasource.DataSource              = (*customProviderDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*customProviderDataSource)(nil)
)

func NewCustomProviderDataSource() datasource.DataSource {
	return &customProviderDataSource{}
}

type customProviderDataSourceModel struct {
	Token                types.String `tfsdk:"token"`
	Name                 types.String `tfsdk:"name"`
	Status               types.String `tfsdk:"status"`
	CreatedAt            types.String `tfsdk:"created_at"`
	LastUpdated          types.String `tfsdk:"last_updated"`
	WorkspaceTokens      types.Set    `tfsdk:"workspace_tokens"`
	ManagedAccountTokens types.Set    `tfsdk:"managed_account_tokens"`
}

type customProviderDataSource struct {
	client *Client
}

func (d *customProviderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *customProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_provider"
}

func (d *customProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Custom Provider integration by its token.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The token of the Custom Provider integration to look up.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The display name of the Custom Provider integration.",
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

func (d *customProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customProviderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := integrationsv2.NewGetIntegrationParams()
	params.SetIntegrationToken(state.Token.ValueString())

	out, err := d.client.V2.Integrations.GetIntegration(params, d.client.Auth)
	if err != nil {
		handleError("Read Custom Provider", &resp.Diagnostics, err)
		return
	}

	p := out.Payload
	state.Token = types.StringValue(p.Token)
	state.Status = types.StringValue(p.Status)
	state.CreatedAt = types.StringValue(p.CreatedAt)

	if p.AccountIdentifier != nil {
		state.Name = types.StringValue(*p.AccountIdentifier)
	} else {
		state.Name = types.StringNull()
	}

	if p.LastUpdated != nil {
		state.LastUpdated = types.StringValue(*p.LastUpdated)
	} else {
		state.LastUpdated = types.StringNull()
	}

	workspaceTokens, diags := types.SetValueFrom(ctx, types.StringType, p.WorkspaceTokens)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.WorkspaceTokens = workspaceTokens

	managedAccountTokens, diags := types.SetValueFrom(ctx, types.StringType, p.ManagedAccountTokens)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ManagedAccountTokens = managedAccountTokens

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
