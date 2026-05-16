package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccIntegrationByNameDataSource_basic(t *testing.T) {
	resourceName := "vantage_custom_provider.test"
	dataSourceName := "data.vantage_integration_by_name.test"
	providerName := "DS ByName Test Provider"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationByNameDataSourceConfig(providerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "token", resourceName, "token"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttr(dataSourceName, "name", providerName),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_at"),
					resource.TestCheckResourceAttr(dataSourceName, "workspace_tokens.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "managed_account_tokens.#", "0"),
				),
			},
		},
	})
}

func TestAccIntegrationByNameDataSource_withProviderFilter(t *testing.T) {
	resourceName := "vantage_custom_provider.test"
	dataSourceName := "data.vantage_integration_by_name.test"
	providerName := "DS ByName Filter Test Provider"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationByNameWithFilterConfig(providerName, "custom_provider"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "token", resourceName, "token"),
					resource.TestCheckResourceAttr(dataSourceName, "name", providerName),
				),
			},
		},
	})
}

func testAccIntegrationByNameDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "vantage_custom_provider" "test" {
  name = %q
}

data "vantage_integration_by_name" "test" {
  name       = %q
  depends_on = [vantage_custom_provider.test]
}
`, name, name)
}

func testAccIntegrationByNameWithFilterConfig(name, providerFilter string) string {
	return fmt.Sprintf(`
resource "vantage_custom_provider" "test" {
  name = %q
}

data "vantage_integration_by_name" "test" {
  name            = %q
  provider_filter = %q
  depends_on      = [vantage_custom_provider.test]
}
`, name, name, providerFilter)
}
