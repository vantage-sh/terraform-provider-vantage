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
			{
				// create apply to all rule
				Config: testAccBillingRule_applyToAll(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_apply_to_all", "title", "test_apply_to_all"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_apply_to_all", "type", "exclusion"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_apply_to_all", "charge_type", "RIFee"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_apply_to_all", "apply_to_all", "true"),
				),
			},
			{
				// update apply to all rule
				Config: testAccBillingRule_applyToAll(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_apply_to_all", "title", "test_apply_to_all"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_apply_to_all", "type", "exclusion"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_apply_to_all", "charge_type", "RIFee"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_apply_to_all", "apply_to_all", "false"),
				),
			},
			{
				// create custom rule
				Config: testAccBillingRule_custom("UPDATE aws SET aws.product/ProductFamily = 'Support'\nWHERE aws.lineItem/LineItemType = 'Fee' AND aws.product/ProductName = 'AWS Support'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_custom", "title", "test_custom"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_custom", "type", "custom"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_custom", "sql_query", "UPDATE aws SET aws.product/ProductFamily = 'Support'\nWHERE aws.lineItem/LineItemType = 'Fee' AND aws.product/ProductName = 'AWS Support'"),
				),
			},
			{
				// update custom rule
				Config: testAccBillingRule_custom("UPDATE aws SET aws.product/ProductFamily = 'Support'\nWHERE aws.lineItem/LineItemType = 'Fee'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_rule.test_custom", "title", "test_custom"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_custom", "type", "custom"),
					resource.TestCheckResourceAttr("vantage_billing_rule.test_custom", "sql_query", "UPDATE aws SET aws.product/ProductFamily = 'Support'\nWHERE aws.lineItem/LineItemType = 'Fee'"),
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
	start_period = ""
	amount = 0.0
	percentage = 0.0
	service = ""
	category = ""
	sub_category = ""
	sql_query = ""
}
`, title, chargeType)
}

func testAccBillingRule_applyToAll(applyToAll bool) string {
	return fmt.Sprintf(`
resource "vantage_billing_rule" "test_apply_to_all" {
	title = "test_apply_to_all"
	type = "exclusion"
	charge_type = "RIFee"
	apply_to_all = %[1]t
	start_period = ""
	amount = 0.0
	percentage = 0.0
	service = ""
	category = ""
	sub_category = ""
	sql_query = ""
}
	`, applyToAll)
}
func testAccBillingRule_adjustment(title, service, category string, percentage float32) string {
	return fmt.Sprintf(`
resource "vantage_billing_rule" "test_adjustment" {
	title = %[1]q
	type = "adjustment"
	service = %[2]q
	category = %[3]q
	percentage = %[4]f
	start_period = ""
	amount = 0.0
	charge_type = ""
	sub_category = ""
	sql_query = ""
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
		charge_type = ""
		percentage = 0.0
		sql_query = ""
	}
	`, title, service, category, subCategory, startPeriod, amount)
}

func testAccBillingRule_custom(query string) string {
	return fmt.Sprintf(`
resource "vantage_billing_rule" "test_custom" {
	title = "test_custom"
	type = "custom"
	sql_query = %[1]q
	charge_type = ""
	start_period = ""
	amount = 0.0
	percentage = 0.0
	service = ""
	category = ""
	sub_category = ""
}
	`, query)
}
