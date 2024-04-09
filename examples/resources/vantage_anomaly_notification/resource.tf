resource "vantage_anomaly_notification" "demo_anomaly_notification" {
  cost_report_token = "rpt_47c3254c790e9351"
  threshold = 10
  recipient_channels = ["#alerts"]
}