package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccIntegrationsDataSource_basic(t *testing.T) {
	dataSourceName := "data.vantage_integrations.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a custom provider so the list is non-empty, then read
				// all integrations and verify the created one appears.
				Config: testAccIntegrationsDataSourceConfig("DS Integrations Test"),
				Check: resource.ComposeTestCheckFunc(
					// At least one integration is returned.
					resource.TestCheckResourceAttrSet(dataSourceName, "integrations.#"),
					// The created provider's token appears in the list at some index.
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "integrations.*", map[string]string{
						"name": "DS Integrations Test",
					}),
				),
			},
		},
	})
}

func TestAccIntegrationsDataSource_withProviderFilter(t *testing.T) {
	dataSourceName := "data.vantage_integrations.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationsDataSourceWithFilterConfig("DS Integrations Filter Test", "custom_provider"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "integrations.#"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "integrations.*", map[string]string{
						"name": "DS Integrations Filter Test",
					}),
				),
			},
		},
	})
}

func testAccIntegrationsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "vantage_custom_provider" "test" {
  name = %q
}

data "vantage_integrations" "test" {
  depends_on = [vantage_custom_provider.test]
}
`, name)
}

func testAccIntegrationsDataSourceWithFilterConfig(name, providerFilter string) string {
	return fmt.Sprintf(`
resource "vantage_custom_provider" "test" {
  name = %q
}

data "vantage_integrations" "test" {
  provider_filter = %q
  depends_on      = [vantage_custom_provider.test]
}
`, name, providerFilter)
}
