---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vantage_virtual_tag_config Resource - terraform-provider-vantage"
subcategory: ""
description: |-
  Manages a Virtual Tag Config.
---

# vantage_virtual_tag_config (Resource)

Manages a Virtual Tag Config.

## Example Usage

```terraform
data "vantage_virtual_tag_configs" "demo" {}
resource "vantage_virtual_tag_config" "demo_virtual_tag_config" {
  key = "Demo Tag"
  backfill_until = "2024-01-01"
  overridable = true
  values = [
    {
      name = "Demo Value 0"
      filter = "(costs.provider = 'aws' AND costs.region = 'us-east-1') OR (costs.provider = 'gcp' AND costs.region = 'us-central1')"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `backfill_until` (String) The earliest month VirtualTagConfig should be backfilled to.
- `key` (String) The key of the VirtualTagConfig.
- `overridable` (Boolean) Whether the VirtualTagConfig can override a provider-supplied tag on a matching Cost.

### Optional

- `values` (Attributes List) (see [below for nested schema](#nestedatt--values))

### Read-Only

- `created_by_token` (String) The token of the User who created the VirtualTagConfig.
- `token` (String) The token of the VirtualTagConfig.

<a id="nestedatt--values"></a>
### Nested Schema for `values`

Required:

- `name` (String) The name of the Value.

Optional:

- `filter` (String) The filter VQL for the Value.


