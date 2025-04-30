package vantage

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccManagedAccount_basic(t *testing.T) {
	t.Logf("Managed Account Domain name: %s", os.Getenv("MANAGED_ACCOUNT_DOMAIN"))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedAccountBillingRules() + testAccManagedAccountResource("br-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_managed_account.test", "name", "Test Account"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "contact_email", "test@vantage.sh"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "billing_rule_tokens.#", "1"),
				),
			},
			{
				Config: testAccManagedAccountBillingRules() + testAccManagedAccountResource("br-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_managed_account.test", "name", "Test Account"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "contact_email", "test@vantage.sh"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "billing_rule_tokens.#", "1"),
				),
			},
		},
	})
}
func testAccManagedAccountBillingRules() string {
	return `
resource "vantage_billing_rule" "br-1" {
	title = "br1"
	type = "adjustment"
	service = "service"
	category = "category"
	percentage = 50.0
}

resource "vantage_billing_rule" "br-2" {
	title = "br1"
	type = "adjustment"
	service = "service"
	category = "category"
	percentage = 50.0
}

`
}
func testAccManagedAccountResource(billing_rule_id string) string {
	domain := os.Getenv("MANAGED_ACCOUNT_DOMAIN")
	if domain == "" {
		domain = "vantage.sh"
	}

	return fmt.Sprintf(`

resource "vantage_managed_account" "test" {
	name                   = "Test Account"
	contact_email           = "test@%[1]s"
	billing_rule_tokens = [vantage_billing_rule.%[2]s.token]
}
	`, domain, billing_rule_id)
}
