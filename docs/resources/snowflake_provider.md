---
page_title: "vantage_snowflake_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Snowflake Account Integration.
---

# vantage_snowflake_provider (Resource)

## Example Usage

```terraform
resource "vantage_snowflake_provider" "demo" {
  account_name = "my_account"
  user_name    = "my_user"
  password     = "supersecret"
  role         = "analyst"
}
```

## Schema

### Required
- `account_name` (String)
- `user_name` (String)
- `password` (String, Sensitive)

### Optional
- `role` (String)

### Read-Only
- `id` (Integer)