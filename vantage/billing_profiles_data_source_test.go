package vantage

import (
	"fmt"
	"os"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccBillingProfilesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingProfilesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_billing_profiles.test", "billing_profiles.#"),
				),
			},
		},
	})
}

func TestAccBillingProfilesDataSource_withResource(t *testing.T) {
	nickname := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingProfilesDataSourceWithResource(nickname),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_billing_profiles.test", "billing_profiles.#"),
					// Check that the list is not empty (contains at least 1 billing profile)
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["data.vantage_billing_profiles.test"]
						if !ok {
							return fmt.Errorf("Not found: data.vantage_billing_profiles.test")
						}
						if rs.Primary.Attributes["billing_profiles.#"] == "0" {
							return fmt.Errorf("Expected at least 1 billing profile, got 0")
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccBillingProfilesDataSourceConfig() string {
	return `
data "vantage_billing_profiles" "test" {}
`
}

func testAccBillingProfilesDataSourceWithResource(nickname string) string {
	return fmt.Sprintf(`
resource "vantage_billing_profile" "test" {
	nickname = %[1]q
}

data "vantage_billing_profiles" "test" {
	depends_on = [vantage_billing_profile.test]
}
`, nickname)
}

// TestAccBillingProfilesDataSource_multipleAdjustmentItems verifies that the data source
// returns ALL adjustment items, not just the first one. This test was added to catch
// a bug where the schema incorrectly defined adjustment_items as a single object
// instead of a list, causing only the first item to be returned.
//
// Note: This test requires MSP invoicing to be enabled on the account.
// It will be skipped in environments without MSP access.
func TestAccBillingProfilesDataSource_multipleAdjustmentItems(t *testing.T) {
	// Skip if MSP billing is not available (indicated by MANAGED_ACCOUNT_DOMAIN not being set)
	if os.Getenv("MANAGED_ACCOUNT_DOMAIN") == "" {
		t.Skip("Skipping test: MSP invoicing required (MANAGED_ACCOUNT_DOMAIN not set)")
	}

	nickname := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingProfileWithMultipleAdjustmentItems(nickname),
				Check: resource.ComposeTestCheckFunc(
					// Verify the resource has multiple adjustment items
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "invoice_adjustment_attributes.adjustment_items.#", "2"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "invoice_adjustment_attributes.adjustment_items.0.name", "State Tax"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "invoice_adjustment_attributes.adjustment_items.1.name", "Processing Fee"),

					// Verify the data source returns ALL adjustment items (not just the first)
					testAccCheckBillingProfileDataSourceHasAllAdjustmentItems("data.vantage_billing_profiles.test", nickname, 2),
				),
			},
		},
	})
}

func testAccBillingProfileWithMultipleAdjustmentItems(nickname string) string {
	return fmt.Sprintf(`
resource "vantage_billing_profile" "test" {
	nickname = %[1]q

	invoice_adjustment_attributes = {
		adjustment_items = [
			{
				name             = "State Tax"
				adjustment_type  = "charge"
				calculation_type = "percentage"
				amount           = 8.25
			},
			{
				name             = "Processing Fee"
				adjustment_type  = "charge"
				calculation_type = "fixed"
				amount           = 25.00
			}
		]
	}
}

data "vantage_billing_profiles" "test" {
	depends_on = [vantage_billing_profile.test]
}
`, nickname)
}

// testAccCheckBillingProfileDataSourceHasAllAdjustmentItems verifies that a billing profile
// in the data source has the expected number of adjustment items. This catches the bug
// where only the first adjustment item was returned.
func testAccCheckBillingProfileDataSourceHasAllAdjustmentItems(dataSourceName, nickname string, expectedCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}

		// Find the billing profile with the matching nickname
		profileCount := rs.Primary.Attributes["billing_profiles.#"]
		for i := 0; ; i++ {
			prefix := fmt.Sprintf("billing_profiles.%d.", i)
			if _, exists := rs.Primary.Attributes[prefix+"nickname"]; !exists {
				break
			}

			if rs.Primary.Attributes[prefix+"nickname"] == nickname {
				// Found the profile, check adjustment_items count
				adjustmentItemsKey := prefix + "invoice_adjustment_attributes.adjustment_items.#"
				actualCount := rs.Primary.Attributes[adjustmentItemsKey]

				if actualCount != fmt.Sprintf("%d", expectedCount) {
					return fmt.Errorf(
						"billing profile %q has %s adjustment_items, expected %d. "+
							"This may indicate the data source is only returning the first item.",
						nickname, actualCount, expectedCount,
					)
				}

				// Verify each item has a name (basic sanity check)
				for j := 0; j < expectedCount; j++ {
					itemNameKey := fmt.Sprintf("%sinvoice_adjustment_attributes.adjustment_items.%d.name", prefix, j)
					if name, exists := rs.Primary.Attributes[itemNameKey]; !exists || name == "" {
						return fmt.Errorf("adjustment_items[%d].name is missing or empty", j)
					}
				}

				return nil
			}
		}

		return fmt.Errorf("billing profile with nickname %q not found in data source (searched %s profiles)", nickname, profileCount)
	}
}
