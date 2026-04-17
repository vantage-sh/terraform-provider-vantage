---
page_title: "vantage_azure_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages an Azure Account Integration.
---

# vantage_azure_provider (Resource)

Manages an Azure Account Integration in Vantage.

## Example Usage

```terraform
resource "vantage_azure_provider" "example" {
  tenant   = "my-tenant-id"
  app_id   = "azure-app-client-id"
  password = "supersecret"
}
```

## Schema

### Required

- `tenant` (String) The Azure Active Directory tenant ID. Changing this value forces a new resource.
- `app_id` (String) The Azure application (client) ID. Changing this value forces a new resource.
- `password` (String, Sensitive) The Azure application client secret. Changing this value forces a new resource.

### Read-Only

- `id` (String) Same as `token`.
- `token` (String) Unique token of the Azure integration.
- `status` (String) The status of the integration.
