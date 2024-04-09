data "vantage_anomaly_notifications" "all" {} 

output "all_anomaly_notifications" {
  value = data.vantage_anomaly_notifications.all
}