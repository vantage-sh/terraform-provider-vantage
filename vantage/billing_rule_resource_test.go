package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccBillingRule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create exclusion
				Config: testAccBillingRule_exclusion("test", "RIFee"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_exclusion", "title", "test"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_exclusion", "type", "exclusion"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_exclusion", "charge_type", "RIFee"),
					resource.TestCheckResourceAttrSet("vantage_billing_rule.test_exclusion", "token"),
				),
			},
			{ // update exclusion
				Config: testAccBillingRule_exclusion("test2", "RIFee2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_exclusion", "title", "test2"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_exclusion", "charge_type", "RIFee2"),
				),
			},
			{ // create adjustment
				Config: testAccBillingRule_adjustment("test3", "service", "category", 50),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "title", "test3"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "type", "adjustment"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "service", "service"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "category", "category"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "percentage", "50"),
					resource.TestCheckResourceAttrSet("vantage_billing_rule.test_adjustment", "token"),
				),
			},
			{ // update existing adjustment rule
				Config: testAccBillingRule_adjustment("test4", "service2", "category2", 60),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "title", "test4"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "service", "service2"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "category", "category2"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_adjustment", "percentage", "60"),
				),
			},
			{
				Config: testAccBillingRule_charge("test5", "service", "category", "subCategory", "2023-01-01", 0.7),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "title", "test5"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "service", "service"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "category", "category"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "sub_category", "subCategory"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "start_period", "2023-01-01"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "amount", "0.7"),
					resource.TestCheckResourceAttrSet("vantage_billing_rule.test_charge", "token"),
				),
			},
			{ // update charge rule
				Config: testAccBillingRule_charge("test6", "service2", "category2", "subCategory2", "2023-01-02", 0.8),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "title", "test6"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "service", "service2"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "category", "category2"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "sub_category", "subCategory2"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "start_period", "2023-01-02"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_charge", "amount", "0.8"),
				),
			},
		},
	})
}

func testAccBillingRule_exclusion(title, chargeType string) string {
	return fmt.Sprintf(`
resource "vantage_billing_rule" "test_exclusion" {
  title = %[1]q
	type = "exclusion"
	charge_type = %[2]q
}
`, title, chargeType)
}

func testAccBillingRule_adjustment(title, service, category string, percentage float32) string {
	return fmt.Sprintf(`
resource "vantage_billing_rule" "test_adjustment" {
	title = %[1]q
	type = "adjustment"
	service = %[2]q
	category = %[3]q
	percentage = %[4]f
}
	`, title, service, category, percentage)
}

func testAccBillingRule_charge(title, service, category, subCategory, startPeriod string, amount float32) string {
	return fmt.Sprintf(`
	resource "vantage_billing_rule" "test_charge" {
		title = %[1]q
		type = "charge"
		service = %[2]q
		category = %[3]q
		sub_category = %[4]q
		start_period = %[5]q
		amount = %[6]f
	}
	`, title, service, category, subCategory, startPeriod, amount)
}
