---
page_title: "vantage_integrations Data Source - terraform-provider-vantage"
subcategory: "Integrations"
description: |-
  Returns all integrations visible to the API token, optionally filtered by provider type.
---

# vantage_integrations (Data Source)

Returns all integrations visible to the API token, optionally filtered by provider type. Fetches up to 1,000 results from the [Get All Integrations](https://docs.vantage.sh/api/integrations/get-all-integrations) endpoint.

## Example Usage

```terraform
# All integrations
data "vantage_integrations" "all" {}

# Only custom_provider integrations
data "vantage_integrations" "custom" {
  provider_filter = "custom_provider"
}

output "custom_provider_tokens" {
  value = [for i in data.vantage_integrations.custom.integrations : i.token]
}
```

## Schema

### Optional

- `provider_filter` (String) Filter results by provider type (e.g. `custom_provider`). Corresponds to the `provider` query parameter on the Get All Integrations API endpoint.

### Read-Only

- `integrations` (List of Object) The list of integrations returned by the API. (see [below for nested schema](#nestedatt--integrations))

<a id="nestedatt--integrations"></a>
### Nested Schema for `integrations`

Read-Only:

- `token` (String) The unique token of the integration.
- `name` (String) The display name of the integration.
- `status` (String) The status of the integration (e.g. `connected`, `pending`, `importing`, `imported`, `error`, `disconnected`).
- `created_at` (String) The date and time (UTC, ISO 8601) when the integration was created.
- `last_updated` (String) The date and time (UTC, ISO 8601) when the integration was last updated. Null if never updated.
- `workspace_tokens` (Set of String) The tokens of the Workspaces associated with this integration.
- `managed_account_tokens` (Set of String) The tokens of any Managed Accounts associated with this integration.
