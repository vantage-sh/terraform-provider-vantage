resource "vantage_dashboard" "demo_dashboard" {
  title         = "Demo Dashboard"
  date_interval = "last_month"
  widgets = [
    {
      settings         = { display_type = "chart" }
      widgetable_token = "rprt_a2846903070824f4"
    }
  ]
  workspace_token = "wrkspc_47c3254c790e9351"
}
