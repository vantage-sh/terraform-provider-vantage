package vantage

import (
	"fmt"
	"os"
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
	token := os.Getenv("ENRICHMENT_SOURCE_TOKEN")
	if token == "" {
		t.Skip("Skipping test: ENRICHMENT_SOURCE_TOKEN not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrichmentSourceDataSourceConfig(token),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vantage_enrichment_source.test", "token", token),
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

func testAccEnrichmentSourceDataSourceConfig(token string) string {
	return fmt.Sprintf(`
data "vantage_enrichment_source" "test" {
  token = %q
}
`, token)
}
