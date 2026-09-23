package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccEnrichmentSourcesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrichmentSourcesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_enrichment_sources.test", "enrichment_sources.#"),
				),
			},
		},
	})
}

func TestAccEnrichmentSourceDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// List all enrichment sources, then look up the first entry by token
				// via the singular data source. This avoids depending on a
				// separately-managed fixture token.
				Config: testAccEnrichmentSourceDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_enrichment_source.test", "token"),
					resource.TestCheckResourceAttrPair(
						"data.vantage_enrichment_source.test", "token",
						"data.vantage_enrichment_sources.test", "enrichment_sources.0.token",
					),
					resource.TestCheckResourceAttrSet("data.vantage_enrichment_source.test", "title"),
					resource.TestCheckResourceAttrSet("data.vantage_enrichment_source.test", "type"),
					resource.TestCheckResourceAttrSet("data.vantage_enrichment_source.test", "integration_token"),
					resource.TestCheckResourceAttrSet("data.vantage_enrichment_source.test", "active"),
					resource.TestCheckResourceAttrSet("data.vantage_enrichment_source.test", "created_at"),
				),
			},
		},
	})
}

func testAccEnrichmentSourcesDataSourceConfig() string {
	return `
data "vantage_enrichment_sources" "test" {}
`
}

func testAccEnrichmentSourceDataSourceConfig() string {
	return `
data "vantage_enrichment_sources" "test" {}

data "vantage_enrichment_source" "test" {
  token = data.vantage_enrichment_sources.test.enrichment_sources[0].token
}
`
}
