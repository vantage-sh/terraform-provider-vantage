resource "vantage_dashboard" "demo_dashboard" {
  title         = "Demo Dashboard"
  date_interval = "last_month"
  widgets = [
    {
      settings = {
        display_type = "chart"
        grid = {
          x = 0
          y = 0
          w = 6
          h = 4
        }
      }
      widgetable_token = "rprt_a2846903070824f4"
    },
    {
      settings = {
        display_type    = "kpi"
        kpi_calculation = "sum"
        kpi_type        = "cost"
        grid = {
          x = 6
          y = 0
          w = 6
          h = 4
        }
      }
      title            = "Total Spend"
      widgetable_token = "rprt_a2846903070824f4"
    }
  ]
  workspace_token = "wrkspc_47c3254c790e9351"
}
