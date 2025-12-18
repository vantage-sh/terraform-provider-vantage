---
page_title: "vantage_mongodb_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a MongoDB Account Integration.
---

# vantage_mongodb_provider (Resource)

## Example Usage

```terraform
resource "vantage_mongodb_provider" "demo" {
  cluster_uri = "mongodb+srv://cluster0.mongodb.net/test"
  api_key    = "supersecretapikey"
}
```

## Schema

### Required
- `cluster_uri` (String)
- `api_key` (String, Sensitive)

### Read-Only
- `id` (Integer)