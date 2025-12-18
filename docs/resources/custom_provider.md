---
page_title: "vantage_custom_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Custom Provider integration.
---

# vantage_custom_provider (Resource)

## Example Usage

```terraform
resource "vantage_custom_provider" "demo" {
  name       = "Test Provider"
  identifier = "unique_identifier"
}
```

## Schema

### Required
- `name` (String)
- `identifier` (String)

### Read-Only
- `id` (Integer)