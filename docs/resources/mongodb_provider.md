---
page_title: "vantage_mongodb_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages a MongoDB Account Integration.
---

# vantage_mongodb_provider (Resource)

Manages a MongoDB Account Integration in Vantage.

~> **Note:** This resource is not yet fully supported. Creating or updating this resource will return an error until SDK support is added.

## Example Usage

```terraform
resource "vantage_mongodb_provider" "example" {
  cluster_uri = "mongodb+srv://cluster0.mongodb.net/test"
  api_key     = "supersecretapikey"
}
```

## Schema

### Required

- `cluster_uri` (String) The MongoDB cluster URI.
- `api_key` (String, Sensitive) The MongoDB Atlas API key.

### Read-Only

- `id` (String) Unique identifier of the MongoDB integration.
