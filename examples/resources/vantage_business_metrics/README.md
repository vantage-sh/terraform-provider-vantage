# Business Metric Examples

This directory contains examples for the `vantage_business_metric` resource.

## Examples

### Basic Business Metric with Values

See `resource.tf` for a basic example of creating a business metric with manual values.

### Business Metric with Cost Report Token References

See `cost_report_tokens_example.tf` for an example of creating a business metric that references multiple cost reports using the `cost_report_tokens_with_metadata` field.

This pattern is useful when you want to:
- Track business metrics across multiple cost reports
- Use Terraform references to ensure proper dependency management
- Maintain consistent unit scales across different cost views
- Keep empty label filters explicitly defined

The example demonstrates:
1. Creating multiple cost reports with different filters
2. Referencing those cost reports in a business metric
3. Setting unit scales for each cost report token
4. Using empty label_filter arrays (which is the recommended practice)

This configuration ensures that Terraform properly manages the dependencies between resources and that the business metric is only created after all referenced cost reports exist.

### Business Metric from GCP BigQuery

See `gcp_bigquery.tf` for a business metric imported from a BigQuery SQL query.

A change to `gcp_bigquery_metric_fields` replaces the resource. The Vantage API does not update that block in place.

### Business Metric from ClickHouse

See `clickhouse.tf` for a business metric imported from a ClickHouse query endpoint.

A change to `clickhouse_metric_fields` replaces the resource. The Vantage API does not update that block in place.

### Business Metric from Metronome

See `metronome.tf` for revenue and usage metrics using a connected Metronome credential. Revenue defaults to `daily_revenue`. Usage requires a billable metric UUID and display name. Both automatically import customer ID, customer name, first ingest alias, and supported customer custom fields as named label dimensions when available.

Configuration edits update the metric in place. Omitting `cost_report_tokens_with_metadata` retains report attachments; setting it to `[]` clears them. Removing the Metronome block does not disconnect the integration. To switch usage to revenue, set `metric_kind = "revenue"` and `metric = "daily_revenue"`, and remove billable metric fields. Use report attachment `label_filters` to filter by the imported dimensions.

Import an existing metric by token, then use its current configuration in Terraform:

```sh
terraform import vantage_business_metric.metronome_revenue bsnss_mtrc_replace_me
```

Roll out Core support before publishing the generated Go client and releasing the Terraform provider.
