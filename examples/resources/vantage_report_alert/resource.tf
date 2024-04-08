resource "vantage_report_alert" "demo_report_alert" {
  cost_report_token = "rpt_47c3254c790e9351"
  threshold = 10
  recipient_channels = ["#alerts"]
}