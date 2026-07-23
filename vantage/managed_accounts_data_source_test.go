package vantage

import (
	"fmt"
	"os"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageManagedAccountsDataSource_basic(t *testing.T) {
	resourceName := "data.vantage_managed_accounts.test"

	domain := os.Getenv("MANAGED_ACCOUNT_DOMAIN")
	if domain == "" {
		domain = "vantage.sh"
	}
	address := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	contactEmail := fmt.Sprintf("%s@%s", address, domain)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// This test verifies the data source doesn't throw a
				// "Value Conversion Error" due to schema mismatch.
				// See: https://github.com/vantage-sh/terraform-provider-vantage/issues/154
				Config: testAccManagedAccountsDataSourceConfig(contactEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "managed_accounts.#"),
				),
			},
		},
	})
}

func testAccManagedAccountsDataSourceConfig(contactEmail string) string {
	return fmt.Sprintf(`
resource "vantage_managed_account" "test" {
	name          = "Test Account"
	contact_email = %[1]q
}

data "vantage_managed_accounts" "test" {
	depends_on = [vantage_managed_account.test]
}
`, contactEmail)
}
