---
page_title: "vantage_folder_by_name Data Source - terraform-provider-vantage"
subcategory: ""
description: |-
  Looks up a folder by title.
---

# vantage_folder_by_name (Data Source)

Looks up a folder by its title. Searches all folders returned by the [Get All Folders](https://docs.vantage.sh/api/folders/get-all-folders) endpoint and returns the first match.

Use `workspace_token` or `parent_folder_token` to narrow the search when multiple folders share the same title.

## Example Usage

```terraform
# Look up by title only
data "vantage_folder_by_name" "example" {
  title = "Engineering"
}

# Look up by title within a specific workspace
data "vantage_folder_by_name" "filtered" {
  title           = "Engineering"
  workspace_token = "wrkspc_1a2b3c4d5e6f"
}

output "folder_token" {
  value = data.vantage_folder_by_name.example.token
}
```

## Schema

### Required

- `title` (String) The title of the folder to find.

### Optional

- `workspace_token` (String) Filter folders by workspace token. If not specified, the first folder matching the title is returned. Also populated as an output with the workspace token of the matched folder.
- `parent_folder_token` (String) Filter folders by parent folder token. Also populated as an output with the parent folder token of the matched folder.

### Read-Only

- `token` (String) The unique token of the matched folder.
