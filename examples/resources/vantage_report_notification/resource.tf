resource "vantage_report_notification" "test_notif" {
  cost_report_token = vantage_cost_report.demo_report.token
  title = "Test Notification"
  user_tokens = ["usr_36b848747e1683bc", "usr_899b013c355547db"]
  frequency = "daily"
  change = "dollars"
  workspace_token = "wrkspc_47c3254c790e9351"
}
