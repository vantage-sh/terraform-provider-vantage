package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccInvoice_basic(t *testing.T) {
	t.Skip("Invoice tests require a managed account token - should be enabled when managed accounts are available")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInvoiceResource("test-managed-account-token", "2024-01-01", "2024-01-31"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_invoice.test", "account_token", "test-managed-account-token"),
					resource.TestCheckResourceAttr("vantage_invoice.test", "billing_period_start", "2024-01-01"),
					resource.TestCheckResourceAttr("vantage_invoice.test", "billing_period_end", "2024-01-31"),
					resource.TestCheckResourceAttrSet("vantage_invoice.test", "token"),
					resource.TestCheckResourceAttrSet("vantage_invoice.test", "created_at"),
					resource.TestCheckResourceAttrSet("vantage_invoice.test", "status"),
				),
			},
			{
				// Update the billing period
				Config: testAccInvoiceResource("test-managed-account-token", "2024-02-01", "2024-02-29"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_invoice.test", "billing_period_start", "2024-02-01"),
					resource.TestCheckResourceAttr("vantage_invoice.test", "billing_period_end", "2024-02-29"),
				),
			},
		},
	})
}

func testAccInvoiceResource(accountToken, startDate, endDate string) string {
	return fmt.Sprintf(`
resource "vantage_invoice" "test" {
	account_token          = %[1]q
	billing_period_start   = %[2]q
	billing_period_end     = %[3]q
}
`, accountToken, startDate, endDate)
}
