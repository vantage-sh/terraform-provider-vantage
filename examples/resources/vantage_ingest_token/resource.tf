resource "vantage_ingest_token" "example" {
  feed_token     = "srvc_data_intgrtn_example"
  name           = "Production"
  retention_days = 30
}
