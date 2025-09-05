# Fetch all billing profiles
data "vantage_billing_profiles" "all_profiles" {}

# Basic billing profile with minimal configuration
resource "vantage_billing_profile" "basic_profile" {
  nickname = "Basic Company Profile Test"
}

# Billing profile with nested billing information attributes
resource "vantage_billing_profile" "complete_profile_nested" {
  nickname = "Complete Company Profile Testng"

  billing_information_attributes = {
    company_name   = "Example Corp"
    address_line_1 = "123 Business Ave"
    address_line_2 = "Suite 100"
    city           = "New York"
    state          = "NY"
    postal_code    = "10001"
    country_code   = "US"
    billing_email  = ["billing@example.com"]
  }

  banking_information_attributes = {
    beneficiary_name = "John Doe"
    bank_name        = "Example Bank"
  }

  business_information_attributes = {
    metadata = {
      custom_fields = [
        {
          name  = "VAT"
          value = "123456789"
        }
      ]
    }
  }
}
