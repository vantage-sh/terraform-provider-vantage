---
page_title: "vantage_databricks_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Databricks Account Integration.
---

# vantage_databricks_provider (Resource)

Manages a Databricks Account Integration in Vantage.

~> **Note:** This resource is not yet implemented.

## Example Usage

```terraform
resource "vantage_databricks_provider" "example" {
  host  = "https://mycompany.cloud.databricks.com"
  token = "databricks-token"
}
```

## Schema

### Required

- `host` (String) The Databricks workspace URL.
- `token` (String, Sensitive) The Databricks personal access token.

### Read-Only

- `id` (String) Unique identifier of the Databricks integration.
