data "vantage_report_notifications" "all" {}

output "all_report_notifications" {
  value = data.vantage_report_notifications.all
}

