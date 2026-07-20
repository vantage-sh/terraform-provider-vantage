---
page_title: "vantage_workspace Data Source - terraform-provider-vantage"
subcategory: "Workspaces"
description: |-
  Looks up a workspace by name.
---

# vantage_workspace (Data Source)

Looks up a workspace by its display name. Searches all workspaces returned by the [Get All Workspaces](https://docs.vantage.sh/api/workspaces/get-all-workspaces) endpoint and returns the first match.

## Example Usage

```terraform
data "vantage_workspace" "example" {
  name = "Production"
}

output "workspace_token" {
  value = data.vantage_workspace.example.token
}
```

## Schema

### Required

- `name` (String) The name of the workspace to find.

### Read-Only

- `token` (String) The unique token of the matched workspace.
