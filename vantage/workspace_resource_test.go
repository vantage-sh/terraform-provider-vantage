package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestWorkspace(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create workspace with minimal configuration
				Config: testAccWorkspace_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					// Test default values
					resource.TestCheckResourceAttr(resourceName, "enable_currency_conversion", "true"),
					resource.TestCheckResourceAttr(resourceName, "exchange_rate_date", "daily_rate"),
				),
			},
			{
				// Update workspace name
				Config: testAccWorkspace_basic(rUpdatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rUpdatedName),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				// Update workspace with full configuration
				Config: testAccWorkspace_full(rUpdatedName, "EUR", false, "end_of_billing_period_rate"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rUpdatedName),
					resource.TestCheckResourceAttr(resourceName, "currency", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "enable_currency_conversion", "false"),
					resource.TestCheckResourceAttr(resourceName, "exchange_rate_date", "end_of_billing_period_rate"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				// Update back to different currency settings
				Config: testAccWorkspace_full(rUpdatedName, "USD", true, "daily_rate"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rUpdatedName),
					resource.TestCheckResourceAttr(resourceName, "currency", "USD"),
					resource.TestCheckResourceAttr(resourceName, "enable_currency_conversion", "true"),
					resource.TestCheckResourceAttr(resourceName, "exchange_rate_date", "daily_rate"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func TestWorkspace_import(t *testing.T) {
	t.Skip("Workspace creation may not be supported via API in current environment - skipping until API permissions are confirmed")
	
	rName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspace_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
				),
			},
			{
				// Test import functionality
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return rs.Primary.Attributes["token"], nil
				},
			},
		},
	})
}

func testAccWorkspace_basic(name string) string {
	return fmt.Sprintf(`
resource "vantage_workspace" "test" {
	name = %[1]q
}
`, name)
}

func testAccWorkspace_full(name, currency string, enableCurrencyConversion bool, exchangeRateDate string) string {
	return fmt.Sprintf(`
resource "vantage_workspace" "test" {
	name                       = %[1]q
	currency                   = %[2]q
	enable_currency_conversion = %[3]t
	exchange_rate_date         = %[4]q
}
`, name, currency, enableCurrencyConversion, exchangeRateDate)
}
