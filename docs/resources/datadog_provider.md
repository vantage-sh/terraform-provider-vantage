---
page_title: "vantage_datadog_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Datadog Account Integration.
---

# vantage_datadog_provider (Resource)

## Example Usage

```terraform
resource "vantage_datadog_provider" "demo" {
  api_key = "ddapikey"
  app_key = "ddappkey"
}
```

## Schema

### Required
- `api_key` (String, Sensitive)
- `app_key` (String, Sensitive)

### Read-Only
- `id` (Integer)