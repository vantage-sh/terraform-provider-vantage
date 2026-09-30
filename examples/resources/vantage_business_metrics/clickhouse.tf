resource "vantage_business_metric" "clickhouse" {
  title = "ClickHouse Revenue"

  clickhouse_metric_fields = {
    integration_token = "accss_crdntl_example"
    query_endpoint_id = "7c6a3a87-12fd-41f5-afdf-caa4697a2886"
  }
}
