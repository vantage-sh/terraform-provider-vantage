package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccInvoice_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInvoiceResourceWithDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_invoice.test", "account_token"),
					resource.TestCheckResourceAttr("vantage_invoice.test", "billing_period_start", "2024-01-01"),
					resource.TestCheckResourceAttr("vantage_invoice.test", "billing_period_end", "2024-01-31"),
					resource.TestCheckResourceAttrSet("vantage_invoice.test", "token"),
					resource.TestCheckResourceAttrSet("vantage_invoice.test", "id"),
					resource.TestCheckResourceAttrSet("vantage_invoice.test", "created_at"),
					resource.TestCheckResourceAttrSet("vantage_invoice.test", "status"),
				),
			},
		},
	})
}

const testAccInvoiceResourceWithDataSource = `
data "vantage_managed_accounts" "test" {}

resource "vantage_invoice" "test" {
	account_token          = data.vantage_managed_accounts.test.managed_accounts[0].token
	billing_period_start   = "2024-01-01"
	billing_period_end     = "2024-01-31"
}
`
