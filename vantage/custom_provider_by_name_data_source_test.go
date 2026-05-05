package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccCustomProviderByNameDataSource_basic(t *testing.T) {
	resourceName := "vantage_custom_provider.test"
	dataSourceName := "data.vantage_custom_provider_by_name.test"
	providerName := "DS ByName Test Provider"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomProviderByNameDataSourceConfig(providerName),
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

func TestAccCustomProviderByNameDataSource_withProviderFilter(t *testing.T) {
	resourceName := "vantage_custom_provider.test"
	dataSourceName := "data.vantage_custom_provider_by_name.test"
	providerName := "DS ByName Filter Test Provider"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomProviderByNameWithFilterConfig(providerName, "custom_provider"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "token", resourceName, "token"),
					resource.TestCheckResourceAttr(dataSourceName, "name", providerName),
				),
			},
		},
	})
}

func testAccCustomProviderByNameDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "vantage_custom_provider" "test" {
  name = %q
}

data "vantage_custom_provider_by_name" "test" {
  name       = %q
  depends_on = [vantage_custom_provider.test]
}
`, name, name)
}

func testAccCustomProviderByNameWithFilterConfig(name, providerFilter string) string {
	return fmt.Sprintf(`
resource "vantage_custom_provider" "test" {
  name = %q
}

data "vantage_custom_provider_by_name" "test" {
  name            = %q
  provider_filter = %q
  depends_on      = [vantage_custom_provider.test]
}
`, name, name, providerFilter)
}
