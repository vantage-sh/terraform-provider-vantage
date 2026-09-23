package vantage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/datasource_enrichment_sources"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	enrichmentsourcesv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/enrichment_sources"
)

var (
	_ datasource.DataSource              = (*enrichmentSourcesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*enrichmentSourcesDataSource)(nil)
)

func NewEnrichmentSourcesDataSource() datasource.DataSource {
	return &enrichmentSourcesDataSource{}
}

type enrichmentSourcesDataSource struct {
	client *Client
}

type enrichmentSourcesDataSourceModel struct {
	EnrichmentSources types.List `tfsdk:"enrichment_sources"`
}

func (d *enrichmentSourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enrichment_sources"
}

func (d *enrichmentSourcesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_enrichment_sources.EnrichmentSourcesDataSourceSchema(ctx)
}

func (d *enrichmentSourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*Client)
}

func (d *enrichmentSourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data enrichmentSourcesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := enrichmentsourcesv2.NewGetEnrichmentSourcesParams()
	out, err := d.client.V2.EnrichmentSources.GetEnrichmentSources(params, d.client.Auth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Vantage Enrichment Sources",
			err.Error(),
		)
		return
	}

	sources := out.Payload.EnrichmentSources
	if sources == nil {
		sources = []*modelsv2.EnrichmentSource{}
	}

	elements := make([]attr.Value, 0, len(sources))
	for _, source := range sources {
		value, diags := enrichmentSourceDataSourceValueFromAPI(ctx, source)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elements = append(elements, value)
	}

	list, diags := types.ListValue(
		datasource_enrichment_sources.EnrichmentSourcesValue{}.Type(ctx),
		elements,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.EnrichmentSources = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func enrichmentSourceDataSourceValueFromAPI(ctx context.Context, src *modelsv2.EnrichmentSource) (datasource_enrichment_sources.EnrichmentSourcesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if src == nil {
		return datasource_enrichment_sources.NewEnrichmentSourcesValueNull(), diags
	}

	value, d := datasource_enrichment_sources.NewEnrichmentSourcesValue(
		datasource_enrichment_sources.EnrichmentSourcesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"active":            types.BoolValue(src.Active),
			"created_at":        types.StringValue(src.CreatedAt),
			"id":                types.StringValue(src.Token),
			"integration_token": types.StringValue(src.IntegrationToken),
			"title":             types.StringValue(src.Title),
			"token":             types.StringValue(src.Token),
			"type":              types.StringValue(src.Type),
		},
	)
	diags.Append(d...)
	return value, diags
}
