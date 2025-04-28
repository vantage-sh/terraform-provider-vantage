package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccManagedAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingRule_adjustment("test", "service", "category", 50) + testAccManagedAccountResource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_managed_account.test", "name", "Test Account"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "contact_email", "test@vantage.sh"),
				),
			},
			// { // update the resource
			// 	Config: testAccManagedAccountResource(),
			// }
		},
	})
}

func testAccManagedAccountResource() string {
	return fmt.Sprintf(`

data "vantage_billing_rules" "rules" {}

resource "vantage_managed_account" "test" {
	name                   = "Test Account"
	contact_email           = "test@vantage.sh"
	billing_rule_tokens = [vantage_billing_rule.test_adjustment.token]
}
	`)
}
