---
page_title: "vantage_azure_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages an Azure Account Integration.
---

# vantage_azure_provider (Resource)

## Example Usage

```terraform
resource "vantage_azure_provider" "demo" {
  tenant_id       = "tenant-123"
  subscription_id = "sub-abc"
  client_id       = "client-xyz"
  client_secret   = "supersecret"
}
```

## Schema

### Required
- `tenant_id` (String)
- `subscription_id` (String)
- `client_id` (String)
- `client_secret` (String, Sensitive)

### Read-Only
- `id` (Integer)