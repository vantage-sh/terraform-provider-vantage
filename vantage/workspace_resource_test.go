package vantage

import (
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageWorkspace(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rNameUpdated := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig(rName, "EUR", "true", "daily_rate"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "currency", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "enable_currency_conversion", "true"),
					resource.TestCheckResourceAttr(resourceName, "exchange_rate_date", "daily_rate"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				Config: testAccWorkspaceConfig(rNameUpdated, "GBP", "true", "end_of_billing_period_rate"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "currency", "GBP"),
					resource.TestCheckResourceAttr(resourceName, "enable_currency_conversion", "true"),
					resource.TestCheckResourceAttr(resourceName, "exchange_rate_date", "end_of_billing_period_rate"),
				),
			},
			{
				Config:             testAccWorkspaceConfig(rNameUpdated, "GBP", "true", "end_of_billing_period_rate"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkspaceConfig(name, currency, enableConv, exchangeDate string) string {
	return fmt.Sprintf(`
resource "vantage_workspace" "test" {
  name                         = %[1]q
  currency                     = %[2]q
  enable_currency_conversion   = %[3]s
  exchange_rate_date           = %[4]q
}
`, name, currency, enableConv, exchangeDate)
}

func TestAccVantageWorkspace_invalidCurrencyConversion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWorkspaceConfig("somenam123", "GBP", "false", "daily_rate"),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}
