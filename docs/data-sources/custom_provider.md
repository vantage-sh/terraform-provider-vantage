---
page_title: "vantage_custom_provider Data Source - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Retrieves a Custom Provider integration by its token.
---

# vantage_custom_provider (Data Source)

Retrieves a Custom Provider integration by its token. Use this data source to read metadata for an existing integration — for example, to reference its workspace associations or status in other resources.

## Example Usage

```terraform
data "vantage_custom_provider" "example" {
  token = "intgr_custom_provider_abc123"
}

output "custom_provider_status" {
  value = data.vantage_custom_provider.example.status
}
```

## Schema

### Required

- `token` (String) The token of the Custom Provider integration to look up.

### Read-Only

- `name` (String) The display name of the Custom Provider integration.
- `status` (String) The status of the integration (e.g. `connected`, `pending`, `importing`, `imported`, `error`, `disconnected`).
- `created_at` (String) The date and time (UTC, ISO 8601) when the integration was created.
- `last_updated` (String) The date and time (UTC, ISO 8601) when the integration was last updated. Null if never updated.
- `workspace_tokens` (Set of String) The tokens of the Workspaces associated with this integration.
- `managed_account_tokens` (Set of String) The tokens of any Managed Accounts associated with this integration.
