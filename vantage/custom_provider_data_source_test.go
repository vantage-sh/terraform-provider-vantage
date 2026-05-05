package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccCustomProviderDataSource_basic(t *testing.T) {
	resourceName := "vantage_custom_provider.test"
	dataSourceName := "data.vantage_custom_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomProviderDataSourceConfig("DS Test Provider"),
				Check: resource.ComposeTestCheckFunc(
					// Verify data source fields match the resource.
					resource.TestCheckResourceAttrPair(dataSourceName, "token", resourceName, "token"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_at"),
					// Workspace and managed account token sets are present (empty).
					resource.TestCheckResourceAttr(dataSourceName, "workspace_tokens.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "managed_account_tokens.#", "0"),
				),
			},
		},
	})
}

func testAccCustomProviderDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "vantage_custom_provider" "test" {
  name = %q
}

data "vantage_custom_provider" "test" {
  token = vantage_custom_provider.test.token
}
`, name)
}
