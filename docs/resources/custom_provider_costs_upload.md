---
page_title: "vantage_custom_provider_costs_upload Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Uploads cost data for a Custom Provider.
---

# vantage_custom_provider_costs_upload (Resource)

## Example Usage

```terraform
resource "vantage_custom_provider_costs_upload" "demo" {
  provider_id = 1
  period      = "2023-12"
  content     = "date,amount\n2023-12-01,100\n2023-12-02,200"
}
```

## Schema

### Required
- `provider_id` (Integer)
- `period` (String)
- `content` (String, Sensitive)

### Read-Only
- `status` (String)
- `id` (Integer)