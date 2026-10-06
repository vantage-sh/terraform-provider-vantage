data "vantage_dashboard_notifications" "all" {}

output "all_dashboard_notifications" {
  value = data.vantage_dashboard_notifications.all
}
