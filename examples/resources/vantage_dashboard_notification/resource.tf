resource "vantage_dashboard_notification" "demo_notification" {
  title           = "Weekly Executive Dashboard"
  dashboard_token = vantage_dashboard.demo_dashboard.token
  frequency       = "weekly"
  user_tokens     = ["usr_36b848747e1683bc"]
  workspace_token = "wrkspc_47c3254c790e9351"
}
