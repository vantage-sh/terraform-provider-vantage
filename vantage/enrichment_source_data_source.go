package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_enrichment_source"
	enrichmentsourcesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/enrichment_sources"
)

var (
	_ datasource.DataSource              = (*enrichmentSourceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*enrichmentSourceDataSource)(nil)
)

func NewEnrichmentSourceDataSource() datasource.DataSource {
	return &enrichmentSourceDataSource{}
}

type enrichmentSourceDataSource struct {
	client *Client
}

func (d *enrichmentSourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enrichment_source"
}

func (d *enrichmentSourceDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_enrichment_source.EnrichmentSourceDataSourceSchema(ctx)
}

func (d *enrichmentSourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *enrichmentSourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_enrichment_source.EnrichmentSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := enrichmentsourcesv2.NewGetEnrichmentSourceParams().
		WithEnrichmentSourceToken(data.Token.ValueString())
	out, err := d.client.V2.EnrichmentSources.GetEnrichmentSource(params, d.client.Auth)
	if err != nil {
		if e, ok := err.(*enrichmentsourcesv2.GetEnrichmentSourceNotFound); ok {
			handleBadRequest("Get Enrichment Source", &resp.Diagnostics, e.GetPayload())
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Enrichment Source",
			err.Error(),
		)
		return
	}

	src := out.Payload
	data.Active = types.BoolValue(src.Active)
	data.CreatedAt = types.StringValue(src.CreatedAt)
	data.IntegrationToken = types.StringValue(src.IntegrationToken)
	data.Title = types.StringValue(src.Title)
	data.Token = types.StringValue(src.Token)
	data.Type = types.StringValue(src.Type)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
