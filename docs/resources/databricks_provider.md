---
page_title: "vantage_databricks_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a Databricks Account Integration.
---

# vantage_databricks_provider (Resource)

## Example Usage

```terraform
resource "vantage_databricks_provider" "demo" {
  host = "https://mycompany.cloud.databricks.com"
  token = "databricks-token"
}
```

## Schema

### Required
- `host` (String)
- `token` (String, Sensitive)

### Read-Only
- `id` (Integer)