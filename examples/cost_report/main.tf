terraform {
  required_providers {
    vantage = {
      source = "vantage-sh/vantage"
    }
  }
}

resource "vantage_folder" "demo_folder" {
  title           = "Demo Folder"
  workspace_token = "wrkspc_47c3254c790e9351"
}

resource "vantage_folder" "demo_folder_child" {
  title               = "Demo Folder First Child"
  parent_folder_token = vantage_folder.demo_folder.token
}

resource "vantage_saved_filter" "demo_filter" {
  title           = "Demo Saved Filter"
  filter          = "(costs.provider = 'aws')"
  workspace_token = "wrkspc_47c3254c790e9351"
}

resource "vantage_cost_report" "demo_report" {
  folder_token        = vantage_folder.demo_folder.token
  filter              = "costs.provider = 'kubernetes'"
  saved_filter_tokens = [vantage_saved_filter.demo_filter.token]
  title               = "Demo Report"
 }
resource "vantage_dashboard" "demo_dashboard" {
  widget_tokens   = [vantage_cost_report.demo_report.token]
  title           = "Demo Dashboard"
  date_interval   = "last_month"
  workspace_token = "wrkspc_47c3254c790e9351"
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
  workspace_tokens = ["wrkspc_47c3254c790e9351"]
}
resource "vantage_access_grant" "demo_access_grant" {
  team_token     = vantage_team.demo_team.token
  resource_token = vantage_dashboard.demo_dashboard.token
}

locals {
  metrics_csv = csvdecode(file("${path.module}/metrics.csv"))
  sorted_dates = distinct(reverse(sort(local.metrics_csv[*].date)))
  sorted_metrics = flatten(
        [for value in local.sorted_dates:
            [ for elem in local.metrics_csv: 
                 elem if value == elem.date
            ]     
        ]) 
}

resource "vantage_business_metric" "demo_metric2" {
  title           = "Demo Metric"
  cost_report_tokens_with_metadata = [
    {
      cost_report_token = vantage_cost_report.demo_report.token
      unit_scale = "per_hundred"
    }
  ]

  values = [for row in local.sorted_metrics : {
    date = row.date
    amount = row.amount
  }]
}

data "vantage_business_metrics" "demo2" {}

output "business_metrics" {
  value = data.vantage_business_metrics.demo2
}
 


