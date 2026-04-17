---
page_title: "vantage_snowflake_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Snowflake Account Integration.
---

# vantage_snowflake_provider (Resource)

Manages a Snowflake Account Integration in Vantage.

~> **Note:** This resource is not yet fully supported. Creating or updating this resource will return an error until SDK support is added.

## Example Usage

```terraform
resource "vantage_snowflake_provider" "example" {
  account_name = "my_account"
  user_name    = "my_user"
  password     = "supersecret"
  role         = "analyst"
}
```

## Schema

### Required

- `account_name` (String) The Snowflake account name.
- `user_name` (String) The Snowflake username.
- `password` (String, Sensitive) The Snowflake password.

### Optional

- `role` (String) The Snowflake role to use.

### Read-Only

- `id` (String) Unique identifier of the Snowflake integration.
