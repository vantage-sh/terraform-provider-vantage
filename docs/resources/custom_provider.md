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
  workspaces  = ["wrkspc_abcd1234"]
}
```

## Schema

### Required

- `name` (String) The display name for the custom provider. Cannot be changed after creation — a warning will be shown and the existing value preserved if a change is attempted.

### Optional

- `description` (String) A description for the custom provider. Cannot be changed after creation — a warning will be shown and the existing value preserved if a change is attempted.
- `workspaces` (Set of String) Workspace tokens to associate with the integration. Can be updated in-place without recreating the resource. **Note:** the Vantage API requires at least one token when updating workspace associations — workspace associations cannot be fully removed once set via Terraform. To disassociate all workspaces, use the Vantage UI or API directly.

### Read-Only

- `id` (String) Same as `token`.
- `token` (String) Unique token of the Custom Provider integration.
- `status` (String) The status of the integration.

## Import

Custom Provider integrations can be imported using their token:

```shell
terraform import vantage_custom_provider.example intgr_custom_provider_abc123
```

Note: `description` is not returned by the API and will not be restored after import.
