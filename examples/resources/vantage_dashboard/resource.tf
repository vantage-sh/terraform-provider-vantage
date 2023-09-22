resource "vantage_dashboard" "demo_dashboard" {
  widget_tokens = ["rprt_a2846903070824f4"]
  title         = "Demo Dashboard"
  date_interval = "last_month"
}
