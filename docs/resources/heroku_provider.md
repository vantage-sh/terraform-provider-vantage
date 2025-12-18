---
page_title: "vantage_heroku_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Heroku Account Integration.
---

# vantage_heroku_provider (Resource)

## Example Usage

```terraform
resource "vantage_heroku_provider" "demo" {
  api_key = "heroku-api-key"
}
```

## Schema

### Required
- `api_key` (String, Sensitive)

### Read-Only
- `id` (Integer)