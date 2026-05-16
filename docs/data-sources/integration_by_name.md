---
page_title: "vantage_integration_by_name Data Source - terraform-provider-vantage"
subcategory: "Integrations"
description: |-
  Looks up an integration by name.
---

# vantage_integration_by_name (Data Source)

Looks up an integration by its display name. Searches up to 1,000 integrations returned by the [Get All Integrations](https://docs.vantage.sh/api/integrations/get-all-integrations) endpoint and returns the first match.

Use `provider_filter` to restrict the search to a specific integration type, which can improve performance when you have many integrations.

## Example Usage

```terraform
# Look up by name only
data "vantage_integration_by_name" "example" {
  name = "My Custom Provider"
}

# Look up by name, restricted to custom_provider integrations
data "vantage_integration_by_name" "filtered" {
  name            = "My Custom Provider"
  provider_filter = "custom_provider"
}

output "token" {
  value = data.vantage_integration_by_name.example.token
}
```

## Schema

### Required

- `name` (String) The display name of the integration to find.

### Optional

- `provider_filter` (String) Filter integrations by provider type before searching (e.g. `custom_provider`). Corresponds to the `provider` query parameter on the Get All Integrations API endpoint.

### Read-Only

- `token` (String) The unique token of the matched integration.
- `status` (String) The status of the integration (e.g. `connected`, `pending`, `importing`, `imported`, `error`, `disconnected`).
- `created_at` (String) The date and time (UTC, ISO 8601) when the integration was created.
- `last_updated` (String) The date and time (UTC, ISO 8601) when the integration was last updated. Null if never updated.
- `workspace_tokens` (Set of String) The tokens of the Workspaces associated with this integration.
- `managed_account_tokens` (Set of String) The tokens of any Managed Accounts associated with this integration.
