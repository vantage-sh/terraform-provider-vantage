---
page_title: "vantage_custom_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Custom Provider integration.
---

# vantage_custom_provider (Resource)

Manages a Custom Provider integration in Vantage. Custom Providers allow you to upload cost data for services not natively supported by Vantage.

## Example Usage

```terraform
resource "vantage_custom_provider" "example" {
  name        = "My Custom Provider"
  description = "An optional description for this provider"
}
```

## Schema

### Required

- `name` (String) The name of the Custom Provider. Changing this value forces a new resource.

### Optional

- `description` (String) A description of the Custom Provider. Changing this value forces a new resource.

### Read-Only

- `id` (String) Same as `token`.
- `token` (String) Unique token of the Custom Provider integration.
- `status` (String) The status of the integration.
