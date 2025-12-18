---
page_title: "vantage_gcp_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a GCP Account Integration.
---

# vantage_gcp_provider (Resource)

## Example Usage

```terraform
resource "vantage_gcp_provider" "demo" {
  project_id      = "test-project"
  billing_account = "000000-111111-222222"
  service_account = <<EOF
{ "type": "service_account", "project_id": "test-project" }
EOF
}
```

## Schema

### Required
- `project_id` (String)
- `billing_account` (String)
- `service_account` (String, Sensitive)

### Read-Only
- `id` (Integer)