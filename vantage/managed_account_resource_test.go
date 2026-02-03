package vantage

import (
	"fmt"
	"os"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccManagedAccount_basic(t *testing.T) {
	domain := os.Getenv("MANAGED_ACCOUNT_DOMAIN")
	if domain == "" {
		domain = "vantage.sh"
	}
	address := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	contactEmail := fmt.Sprintf("%s@%s", address, domain)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedAccountBillingRules() + testAccManagedAccountResource("br-1", contactEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_managed_account.test", "name", "Test Account"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "contact_email", contactEmail),
					resource.TestCheckNoResourceAttr("vantage_managed_account.test", "access_credential_tokens"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "billing_rule_tokens.#", "1"),
				),
			},
			{
				Config: testAccManagedAccountBillingRules() + testAccManagedAccountResource("br-2", contactEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_managed_account.test", "name", "Test Account"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "contact_email", contactEmail),
					resource.TestCheckNoResourceAttr("vantage_managed_account.test", "access_credential_tokens"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "billing_rule_tokens.#", "1"),
				),
			},
			{
				Config: testAccManagedAccountBillingRules() + testAccManagedAccountWithDelegationsResource("br-2", "ac-1", contactEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_managed_account.test", "name", "Test Account"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "contact_email", contactEmail),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "access_credential_tokens.#", "1"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "billing_rule_tokens.#", "1"),
				),
			},
		},
	})
}

func TestAccManagedAccount_withEmailDomain(t *testing.T) {
	domain := os.Getenv("MANAGED_ACCOUNT_DOMAIN")
	if domain == "" {
		domain = "vantage.sh"
	}
	address := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	contactEmail := fmt.Sprintf("%s@%s", address, domain)
	emailDomain := fmt.Sprintf("%s.example.com", sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum))
	updatedEmailDomain := fmt.Sprintf("%s.example.com", sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create managed account with email_domain
			{
				Config: testAccManagedAccountWithEmailDomain(contactEmail, emailDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_managed_account.test", "name", "Test Account"),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "contact_email", contactEmail),
					resource.TestCheckResourceAttr("vantage_managed_account.test", "email_domain", emailDomain),
					resource.TestCheckResourceAttrSet("vantage_managed_account.test", "token"),
				),
			},
			// Step 2: Update email_domain
			{
				Config: testAccManagedAccountWithEmailDomain(contactEmail, updatedEmailDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_managed_account.test", "email_domain", updatedEmailDomain),
				),
			},
			// Step 3: Confirm no changes after apply
			{
				Config:             testAccManagedAccountWithEmailDomain(contactEmail, updatedEmailDomain),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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
func testAccManagedAccountResource(billingRuleId, contactEmail string) string {
	return fmt.Sprintf(`

resource "vantage_managed_account" "test" {
	name                   = "Test Account"
	contact_email           = "%[1]s"
	billing_rule_tokens = [vantage_billing_rule.%[2]s.token]
}
	`, contactEmail, billingRuleId)
}
func testAccManagedAccountWithDelegationsResource(billingRuleId, accessCredentialToken, contactEmail string) string {
	return fmt.Sprintf(`

resource "vantage_managed_account" "test" {
	name                   = "Test Account"
	contact_email           = "%[1]s"
	access_credential_tokens = ["%[3]s"]
	billing_rule_tokens = [vantage_billing_rule.%[2]s.token]
}
	`, contactEmail, billingRuleId, accessCredentialToken)
}

func testAccManagedAccountWithEmailDomain(contactEmail, emailDomain string) string {
	return fmt.Sprintf(`
resource "vantage_managed_account" "test" {
	name          = "Test Account"
	contact_email = %[1]q
	email_domain  = %[2]q
}
`, contactEmail, emailDomain)
}
