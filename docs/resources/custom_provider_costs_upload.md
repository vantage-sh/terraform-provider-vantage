---
page_title: "vantage_custom_provider_costs_upload Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Uploads cost data for a Custom Provider integration.
---

# vantage_custom_provider_costs_upload (Resource)

Uploads a CSV file of cost data for a Custom Provider integration. Each upload is immutable — any change to `integration_token` or `csv_content` will destroy and recreate the resource.

## Example Usage

```terraform
resource "vantage_custom_provider" "example" {
  name = "My Custom Provider"
}

resource "vantage_custom_provider_costs_upload" "example" {
  integration_token = vantage_custom_provider.example.token
  csv_content       = file("${path.module}/costs.csv")
  auto_transform    = true
}
```

## Schema

### Required

- `integration_token` (String) The token of the Custom Provider integration to upload costs for. Changing this value forces a new resource.
- `csv_content` (String, Sensitive) CSV content to upload as costs data. Changing this value forces a new resource.

### Optional

- `auto_transform` (Boolean) When true, attempts to automatically transform the CSV to match the FOCUS format.

### Read-Only

- `id` (String) Same as `token`.
- `token` (String) Unique token of the costs upload.
- `import_status` (String) The import status of the upload (e.g. `processing`, `complete`, `error`).
- `start_date` (String) The start date of the costs in the upload.
- `end_date` (String) The end date of the costs in the upload.
- `amount` (String) The total amount of costs in the upload.
- `filename` (String) The filename of the uploaded costs file.
