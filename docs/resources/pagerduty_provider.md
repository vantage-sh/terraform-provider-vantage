---
page_title: "vantage_pagerduty_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a PagerDuty Account Integration.
---

# vantage_pagerduty_provider (Resource)

Manages a PagerDuty Account Integration in Vantage.

~> **Note:** This resource is not yet implemented.

## Example Usage

```terraform
resource "vantage_pagerduty_provider" "example" {
  api_key = "pagerduty-api-key"
}
```

## Schema

### Required

- `api_key` (String, Sensitive) The PagerDuty API key.

### Read-Only

- `id` (String) Unique identifier of the PagerDuty integration.
