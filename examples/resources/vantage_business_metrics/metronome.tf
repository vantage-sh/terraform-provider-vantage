resource "vantage_business_metric" "metronome_revenue" {
  title = "Metronome revenue"
  metronome_metric_fields = {
    integration_token = "accss_crdntl_replace_me"
    metric_kind       = "revenue"
    metric            = "daily_revenue"
  }
}

resource "vantage_business_metric" "metronome_usage" {
  title = "Metronome API requests"
  metronome_metric_fields = {
    integration_token    = "accss_crdntl_replace_me"
    metric_kind          = "usage"
    metric               = "usage"
    billable_metric_id   = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
    billable_metric_name = "API requests"
  }
}
