data "vantage_report_alerts" "all" {} 

output "all_report_alerts" {
  value = data.vantage_report_alerts.all
}