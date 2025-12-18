---
page_title: "vantage_pagerduty_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a PagerDuty Account Integration.
---

# vantage_pagerduty_provider (Resource)

## Example Usage

```terraform
resource "vantage_pagerduty_provider" "demo" {
  api_key = "pagerduty-api-key"
}
```

## Schema

### Required
- `api_key` (String, Sensitive)

### Read-Only
- `id` (Integer)