---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vantage_virtual_tag_configs Data Source - terraform-provider-vantage"
subcategory: ""
description: |-
  
---

# vantage_virtual_tag_configs (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `virtual_tag_configs` (Attributes List) (see [below for nested schema](#nestedatt--virtual_tag_configs))

<a id="nestedatt--virtual_tag_configs"></a>
### Nested Schema for `virtual_tag_configs`

Read-Only:

- `backfill_until` (String) The earliest month VirtualTagConfig should be backfilled to.
- `created_by_token` (String) The token of the Creator of the VirtualTagConfig.
- `key` (String) The key of the VirtualTagConfig.
- `overridable` (Boolean) Whether the VirtualTagConfig can override a provider-supplied tag on a matching Cost.
- `token` (String) The token of the VirtualTagConfig.
- `values` (Attributes List) Values for the VirtualTagConfig, with match precedence determined by their relative order in the list. (see [below for nested schema](#nestedatt--virtual_tag_configs--values))

<a id="nestedatt--virtual_tag_configs--values"></a>
### Nested Schema for `virtual_tag_configs.values`

Read-Only:

- `business_metric_token` (String) The token of the associated BusinessMetric.
- `cost_metric` (Attributes) (see [below for nested schema](#nestedatt--virtual_tag_configs--values--cost_metric))
- `filter` (String) The filter VQL for the Value.
- `name` (String) The name of the Value.

<a id="nestedatt--virtual_tag_configs--values--cost_metric"></a>
### Nested Schema for `virtual_tag_configs.values.cost_metric`

Read-Only:

- `aggregation` (Attributes) (see [below for nested schema](#nestedatt--virtual_tag_configs--values--cost_metric--aggregation))
- `filter` (String) The filter VQL for the cost metric.

<a id="nestedatt--virtual_tag_configs--values--cost_metric--aggregation"></a>
### Nested Schema for `virtual_tag_configs.values.cost_metric.filter`

Read-Only:

- `tag` (String) The tag to aggregate on.


