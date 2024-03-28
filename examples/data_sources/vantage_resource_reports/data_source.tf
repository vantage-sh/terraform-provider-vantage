data "vantage_resource_reports" "reports" {
}

output "reports" {
  value = data.vantage_resource_reports.reports
}
