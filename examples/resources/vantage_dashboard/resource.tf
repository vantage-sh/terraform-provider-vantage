# Requires Dashboard Folders to be enabled for the account.
resource "vantage_folder" "dashboards" {
  title           = "Dashboards"
  type            = "DashboardFolder"
  workspace_token = "wrkspc_47c3254c790e9351"
}

resource "vantage_dashboard" "demo_dashboard" {
  title         = "Demo Dashboard"
  date_interval = "last_month"
  # Remove folder_token to move the dashboard back to the workspace root.
  folder_token = vantage_folder.dashboards.token
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
    },
    {
      title           = "Notes"
      widgetable_type = "free_text"
      content = jsonencode({
        type = "doc"
        content = [
          {
            type = "paragraph"
            content = [
              {
                type = "text"
                text = "Use free text widgets for dashboard context."
              }
            ]
          }
        ]
      })
    }
  ]
  workspace_token = "wrkspc_47c3254c790e9351"
}
