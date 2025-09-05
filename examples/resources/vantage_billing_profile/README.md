# Billing Profile Resource Examples

This directory contains examples for using the `vantage_billing_profile` resource.

## Files

- `resource.tf` - Basic and advanced billing profile examples
- `complete_example.tf` - Complete workflow showing billing profile and invoice creation

## Usage

The billing profile resource manages billing profiles that contain billing address and contact information for managed accounts.

### Required Parameters

- `nickname` - Display name for the billing profile

### Optional Parameters

```hcl
billing_information_attributes = {
  company_name   = "Example Corp"
  address_line_1 = "123 Business Ave"
  city          = "New York"
  state         = "NY"
  postal_code   = "10001"
  country_code  = "US"
  billing_email = ["billing@example.com"]
}
```

### Available Billing Information Fields

- `company_name` - Company name for billing
- `address_line_1` - First line of billing address
- `address_line_2` - Second line of billing address
- `city` - City for billing address
- `state` - State or province for billing address
- `postal_code` - Postal or ZIP code
- `country_code` - ISO country code
- `billing_email` - Array of billing email addresses

### Computed Attributes

- `token` - The unique token of the billing profile
- `managed_accounts_count` - Number of managed accounts using this profile
- `created_at` - When the profile was created
- `updated_at` - When the profile was last updated
