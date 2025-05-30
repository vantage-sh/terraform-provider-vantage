---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vantage_network_flow_report Resource - terraform-provider-vantage"
subcategory: ""
description: |-
  
---

# vantage_network_flow_report (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `title` (String) The title of the NetworkFlowReport.
- `workspace_token` (String) The Workspace in which the NetworkFlowReport will be created.

### Optional

- `date_interval` (String) The date interval of the NetworkFlowReport. Unless 'custom' is used, this is incompatible with 'start_date' and 'end_date' parameters. Defaults to 'last_7_days'.
- `end_date` (String) The end date of the NetworkFlowReport. YYYY-MM-DD formatted. Incompatible with 'date_interval' parameter.
- `filter` (String) The filter query language to apply to the NetworkFlowReport. Additional documentation available at https://docs.vantage.sh/vql.
- `flow_direction` (String) The flow direction of the NetworkFlowReport.
- `flow_weight` (String) The dimension by which the logs in the report are sorted. Defaults to costs.
- `groupings` (List of String) Grouping values for aggregating data on the NetworkFlowReport. Valid groupings: account_id, az_id, dstaddr, dsthostname, flow_direction, interface_id, instance_id, peer_resource_uuid, peer_account_id, peer_vpc_id, peer_region, peer_az_id, peer_subnet_id, peer_interface_id, peer_instance_id, region, resource_uuid, srcaddr, srchostname, subnet_id, traffic_category, traffic_path, vpc_id.
- `start_date` (String) The start date of the NetworkFlowReport. YYYY-MM-DD formatted. Incompatible with 'date_interval' parameter.

### Read-Only

- `created_at` (String) The date and time, in UTC, the report was created. ISO 8601 Formatted.
- `created_by_token` (String) The token for the User or Team that created this NetworkFlowReport.
- `default` (Boolean) Indicates whether the NetworkFlowReport is the default report.
- `token` (String) The token of the report


