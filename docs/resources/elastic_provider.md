---
page_title: "vantage_elastic_provider Resource - terraform-provider-vantage"
subcategory: "Vendor Integrations"
description: |-
  Manages an Elastic Account Integration.
---

# vantage_elastic_provider (Resource)

Manages an Elastic Account Integration in Vantage.

~> **Note:** This resource is not yet fully supported. Creating or updating this resource will return an error until SDK support is added.

## Example Usage

```terraform
resource "vantage_elastic_provider" "example" {
  api_key = "my-elastic-cloud-api-key"
}
```

## Schema

### Required

- `api_key` (String, Sensitive) The Elastic Cloud API key.

### Read-Only

- `id` (String) Unique identifier of the Elastic integration.
