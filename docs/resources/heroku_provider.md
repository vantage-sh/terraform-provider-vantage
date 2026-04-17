---
page_title: "vantage_heroku_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Heroku Account Integration.
---

# vantage_heroku_provider (Resource)

Manages a Heroku Account Integration in Vantage.

~> **Note:** This resource is not yet implemented.

## Example Usage

```terraform
resource "vantage_heroku_provider" "example" {
  api_key = "heroku-api-key"
}
```

## Schema

### Required

- `api_key` (String, Sensitive) The Heroku API key.

### Read-Only

- `id` (String) Unique identifier of the Heroku integration.
