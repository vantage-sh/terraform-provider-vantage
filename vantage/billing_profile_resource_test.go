package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccBillingProfile_basic(t *testing.T) {
	nickname := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create billing profile with minimal required fields
				Config: testAccBillingProfileResource(nickname),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "nickname", nickname),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "token"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "id"),
				),
			},
			{
				// Update the nickname
				Config: testAccBillingProfileResource(nickname + "-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "nickname", nickname+"-updated"),
				),
			},
		},
	})
}

func TestAccBillingProfile_withNestedAttributes(t *testing.T) {
	nickname := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	companyName := "Test Company " + sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create billing profile with nested billing information
				Config: testAccBillingProfileWithBillingInfo(nickname, companyName, "123 Main St", "New York", "NY", "10001", "US"),
				Check: resource.ComposeTestCheckFunc(
					// Basic attributes
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "nickname", nickname),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "token"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "id"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "created_at"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "updated_at"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "managed_accounts_count", "0"),

				// Billing information attributes
				resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.company_name", companyName),
				resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.address_line_1", "123 Main St"),
				resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.city", "New York"),
				resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.state", "NY"),
				resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.postal_code", "10001"),
				resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.country_code", "US"),
				resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "billing_information_attributes.token"),

				// Verify billing_email is empty array
				resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.billing_email.#", "0"),
				),
			},
			{
				// Update billing information
				Config: testAccBillingProfileWithBillingInfo(nickname, companyName+" Updated", "456 Oak St", "Boston", "MA", "02101", "US"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.company_name", companyName+" Updated"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.address_line_1", "456 Oak St"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.city", "Boston"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.state", "MA"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.postal_code", "02101"),

					// Verify token persists through updates
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "billing_information_attributes.token"),
				),
			},
		},
	})
}

// Test comprehensive nested attributes including the problematic fields
func TestAccBillingProfile_withCompleteNestedAttributes(t *testing.T) {
	nickname := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	companyName := "Complete Test " + sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlphaNum)
	email := "test-" + sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlphaNum) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create with comprehensive billing info including address_line_2 and billing_email
				Config: testAccBillingProfileWithCompleteInfo(nickname, companyName, "123 Main St", "Suite 100", "New York", "NY", "10001", "US", email),
				Check: resource.ComposeTestCheckFunc(
					// Verify all billing information attributes including problematic ones
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "nickname", nickname),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.company_name", companyName),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.address_line_1", "123 Main St"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.address_line_2", "Suite 100"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.city", "New York"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.state", "NY"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.postal_code", "10001"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.country_code", "US"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "billing_information_attributes.token"),

					// Verify billing_email array
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.billing_email.#", "1"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.billing_email.0", email),
				),
			},
			{
				// Update to different address but keep address_line_2 to test field persistence
				Config: testAccBillingProfileWithCompleteInfo(nickname, companyName+" Updated", "456 Oak Ave", "Floor 2", "Boston", "MA", "02101", "US", email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.address_line_1", "456 Oak Ave"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.address_line_2", "Floor 2"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.billing_email.#", "1"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.city", "Boston"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.company_name", companyName+" Updated"),
				),
			},
		},
	})
}

// Test banking information attributes (newly fixed functionality)
func TestAccBillingProfile_withBankingAttributes(t *testing.T) {
	nickname := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	bankName := "Test Bank " + sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlphaNum)
	beneficiaryName := "Test Beneficiary " + sdkacctest.RandStringFromCharSet(5, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create with banking information attributes
				Config: testAccBillingProfileWithBankingInfo(nickname, bankName, beneficiaryName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "nickname", nickname),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "banking_information_attributes.bank_name", bankName),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "banking_information_attributes.beneficiary_name", beneficiaryName),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "banking_information_attributes.token"),
				),
			},
			{
				// Update banking information
				Config: testAccBillingProfileWithBankingInfo(nickname, bankName+" Updated", beneficiaryName+" Updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "banking_information_attributes.bank_name", bankName+" Updated"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "banking_information_attributes.beneficiary_name", beneficiaryName+" Updated"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "banking_information_attributes.token"),
				),
			},
		},
	})
}

// Test combined banking and billing attributes
func TestAccBillingProfile_withBothNestedAttributes(t *testing.T) {
	nickname := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create with both banking and billing information
				Config: testAccBillingProfileWithBothAttributes(nickname),
				Check: resource.ComposeTestCheckFunc(
					// Verify banking attributes
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "banking_information_attributes.bank_name", "Example Bank"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "banking_information_attributes.beneficiary_name", "John Doe"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "banking_information_attributes.token"),

					// Verify billing attributes
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.company_name", "Example Corp"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.address_line_1", "123 Business Ave"),
					resource.TestCheckResourceAttr("vantage_billing_profile.test", "billing_information_attributes.address_line_2", "Suite 100"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "billing_information_attributes.token"),

					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "banking_information_attributes.token"),
					resource.TestCheckResourceAttrSet("vantage_billing_profile.test", "billing_information_attributes.token"),
				),
			},
		},
	})
}

func testAccBillingProfileResource(nickname string) string {
	return fmt.Sprintf(`
resource "vantage_billing_profile" "test" {
	nickname = %[1]q
}
`, nickname)
}

func testAccBillingProfileWithBillingInfo(nickname, companyName, address, city, state, postalCode, countryCode string) string {
	return fmt.Sprintf(`
resource "vantage_billing_profile" "test" {
	nickname = %[1]q
	billing_information_attributes = {
		company_name   = %[2]q
		address_line_1 = %[3]q
		city          = %[4]q
		state         = %[5]q
		postal_code   = %[6]q
		country_code  = %[7]q
	}
}
`, nickname, companyName, address, city, state, postalCode, countryCode)
}

func testAccBillingProfileWithCompleteInfo(nickname, companyName, address1, address2, city, state, postalCode, countryCode, email string) string {
	return fmt.Sprintf(`
resource "vantage_billing_profile" "test" {
	nickname = %[1]q
	billing_information_attributes = {
		company_name   = %[2]q
		address_line_1 = %[3]q
		address_line_2 = %[4]q
		city          = %[5]q
		state         = %[6]q
		postal_code   = %[7]q
		country_code  = %[8]q
		billing_email = [%[9]q]
	}
}
`, nickname, companyName, address1, address2, city, state, postalCode, countryCode, email)
}

func testAccBillingProfileWithBankingInfo(nickname, bankName, beneficiaryName string) string {
	return fmt.Sprintf(`
resource "vantage_billing_profile" "test" {
	nickname = %[1]q
	banking_information_attributes = {
		bank_name        = %[2]q
		beneficiary_name = %[3]q
	}
}
`, nickname, bankName, beneficiaryName)
}

func testAccBillingProfileWithBothAttributes(nickname string) string {
	return fmt.Sprintf(`
resource "vantage_billing_profile" "test" {
	nickname = %[1]q

	banking_information_attributes = {
		bank_name        = "Example Bank"
		beneficiary_name = "John Doe"
	}

	billing_information_attributes = {
		company_name   = "Example Corp"
		address_line_1 = "123 Business Ave"
		address_line_2 = "Suite 100"
		city          = "New York"
		state         = "NY"
		postal_code   = "10001"
		country_code  = "US"
		billing_email = ["test@example.com"]
	}
}
`, nickname)
}
