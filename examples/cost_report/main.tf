terraform {
  required_providers {
    vantage = {
      source = "vantage-sh/vantage"
    }
  }
}

data "vantage_workspaces" "demo" {}
resource "vantage_folder" "demo_folder" {
  title           = "Demo Folder"
  workspace_token = data.vantage_workspaces.demo.workspaces[0].token
}

resource "vantage_folder" "demo_folder_child" {
  title               = "Demo Folder First Child"
  parent_folder_token = vantage_folder.demo_folder.token
}

resource "vantage_saved_filter" "demo_filter" {
  title           = "Demo Saved Filter"
  filter          = "(costs.provider = 'aws')"
  workspace_token = data.vantage_workspaces.demo.workspaces[0].token
}

resource "vantage_cost_report" "demo_report" {
  folder_token        = vantage_folder.demo_folder.token
  filter              = "costs.provider = 'kubernetes'"
  saved_filter_tokens = [vantage_saved_filter.demo_filter.token]
  title               = "Demo Report"
  date_bin = "day"
  chart_type = "line"
}
resource "vantage_dashboard" "demo_dashboard" {
  title         = "Demo Dashboard"
  date_interval = "last_month"
  widgets = [
    {
      settings         = { display_type = "chart" }
      widgetable_token = vantage_cost_report.demo_report.token
    }
  ]
  workspace_token = data.vantage_workspaces.demo.workspaces[0].token
  # saved_filter_tokens = [vantage_saved_filter.demo_filter.token]
}

resource "vantage_team" "demo_team" {
  name        = "Demo Team"
  description = "Demo Team Description"
  user_emails = ["support@vantage.sh"]
}

resource "vantage_team" "demo_team_2" {
  name             = "Another Demo Team"
  description      = "Demo Team Description"
  user_tokens      = ["usr_36b848747e1683bc", "usr_899b013c355547db"]
  workspace_tokens = [ data.vantage_workspaces.demo.workspaces[0].token ]
}
resource "vantage_access_grant" "demo_access_grant" {
  team_token     = vantage_team.demo_team.token
  resource_token = vantage_dashboard.demo_dashboard.token
}

locals {
  metrics_csv  = csvdecode(file("${path.module}/metrics.csv"))
  sorted_dates = distinct(reverse(sort(local.metrics_csv[*].date)))
  sorted_metrics = flatten(
    [for value in local.sorted_dates :
      [for elem in local.metrics_csv :
        elem if value == elem.date
      ]
  ])
}

resource "vantage_business_metric" "demo_metric2" {
  title = "Demo Metric"
  cost_report_tokens_with_metadata = [
    {
      cost_report_token = vantage_cost_report.demo_report.token
      unit_scale        = "per_hundred"
    }
  ]

  values = [for row in local.sorted_metrics : {
    date   = row.date
    amount = row.amount
  }]
}

data "vantage_business_metrics" "demo2" {}

output "business_metrics" {
  value = data.vantage_business_metrics.demo2
}

resource "vantage_budget" "demo_budget" {
  name              = "Demo Budget"
  cost_report_token = vantage_cost_report.demo_report.token
  periods = [
    {
      start_at = "2023-12-01"
      end_at   = "2024-01-01"
      amount   = 1000
    }
  ]
}

data "vantage_budgets" "demo" {}

output "budgets" {
  value = data.vantage_budgets.demo
}
