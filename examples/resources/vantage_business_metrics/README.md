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
