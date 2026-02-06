package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageManagedAccountsDataSource_basic(t *testing.T) {
	resourceName := "data.vantage_managed_accounts.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// This test verifies the data source doesn't throw a
				// "Value Conversion Error" due to schema mismatch.
				// See: https://github.com/vantage-sh/terraform-provider-vantage/issues/154
				Config: testAccManagedAccountsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "managed_accounts.#"),
				),
			},
		},
	})
}

const testAccManagedAccountsDataSourceConfig = `
data "vantage_managed_accounts" "test" {}
`
