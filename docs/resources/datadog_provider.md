---
page_title: "vantage_datadog_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Datadog Account Integration.
---

# vantage_datadog_provider (Resource)

Manages a Datadog Account Integration in Vantage.

~> **Note:** This resource is not yet fully supported. Creating or updating this resource will return an error until SDK support is added.

## Example Usage

```terraform
resource "vantage_datadog_provider" "example" {
  api_key = "ddapikey"
  app_key = "ddappkey"
}
```

## Schema

### Required

- `api_key` (String, Sensitive) The Datadog API key.
- `app_key` (String, Sensitive) The Datadog application key.

### Read-Only

- `id` (String) Unique identifier of the Datadog integration.
