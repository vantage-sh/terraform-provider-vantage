resource "vantage_business_metric" "bigquery" {
  title = "BigQuery Revenue"

  gcp_bigquery_metric_fields = {
    integration_token = "accss_crdntl_example"
    query_project_id  = "my-query-project"
    sql_query         = "SELECT date, value, label FROM `project.dataset.metrics`"
  }
}
