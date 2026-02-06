package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	integrationsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/integrations"
)

var (
	_ datasource.DataSource              = &integrationsDataSource{}
	_ datasource.DataSourceWithConfigure = &integrationsDataSource{}
)

func NewIntegrationsDataSource() datasource.DataSource {
	return &integrationsDataSource{}
}

type integrationsDataSource struct {
	client *Client
}

type integrationDataSourceModel struct {
	Token             types.String `tfsdk:"token"`
	Provider          types.String `tfsdk:"provider"`
	AccountIdentifier types.String `tfsdk:"account_identifier"`
	Status            types.String `tfsdk:"status"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

type integrationsDataSourceModel struct {
	Integrations []integrationDataSourceModel `tfsdk:"integrations"`
}

func (d *integrationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integrations"
}

func (d *integrationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"integrations": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"token": schema.StringAttribute{
							Computed: true,
						},
						"provider": schema.StringAttribute{
							Computed: true,
						},
						"account_identifier": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"created_at": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *integrationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state integrationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	params := integrationsv2.NewGetIntegrationsParams()
	out, err := d.client.V2.Integrations.GetIntegrations(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Integrations",
			err.Error(),
		)
		return
	}

	state.Integrations = []integrationDataSourceModel{}
	for _, i := range out.Payload.Integrations {
		state.Integrations = append(state.Integrations, integrationDataSourceModel{
			Token:             types.StringValue(i.Token),
			Provider:          types.StringValue(i.Provider),
			AccountIdentifier: types.StringPointerValue(i.AccountIdentifier),
			Status:            types.StringValue(i.Status),
			CreatedAt:         types.StringValue(i.CreatedAt),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *integrationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}
