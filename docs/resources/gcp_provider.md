---
page_title: "vantage_gcp_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a GCP Account Integration.
---

# vantage_gcp_provider (Resource)

Manages a Google Cloud Platform Account Integration in Vantage.

## Example Usage

```terraform
resource "vantage_gcp_provider" "example" {
  project_id      = "my-gcp-project"
  billing_account = "000000-111111-222222"
  dataset_name    = "my_billing_dataset"
}
```

## Schema

### Required

- `project_id` (String) The GCP project ID. Changing this value forces a new resource.
- `billing_account` (String) The GCP billing account ID. Changing this value forces a new resource.
- `dataset_name` (String) The BigQuery dataset name containing the billing export. Changing this value forces a new resource.

### Read-Only

- `id` (String) Same as `token`.
- `token` (String) Unique token of the GCP integration.
- `status` (String) The status of the integration.
