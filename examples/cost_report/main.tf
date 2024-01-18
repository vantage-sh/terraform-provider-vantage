terraform {
  required_providers {
    vantage = {
      source = "vantage-sh/vantage"
    }
  }
}

resource "vantage_folder" "demo_folder" {
  title = "Demo Folder"
}

resource "vantage_folder" "demo_folder_child" {
  title               = "Demo Folder First Child"
  parent_folder_token = vantage_folder.demo_folder.token
}

resource "vantage_saved_filter" "demo_filter" {
  title  = "Demo Saved Filter"
  filter = "(costs.provider = 'aws')"
}

resource "vantage_cost_report" "demo_report" {
  folder_token        = vantage_folder.demo_folder.token
  filter              = "costs.provider = 'kubernetes'"
  saved_filter_tokens = [vantage_saved_filter.demo_filter.token]
  title               = "Demo Report"
}

resource "vantage_dashboard" "demo_dashboard" {
  widget_tokens = [vantage_cost_report.demo_report.token]
  title         = "Demo Dashboard"
  date_interval = "last_month"
}

resource "vantage_team" "demo_team" {
  name = "Demo Team"
  description = "Demo Team Description"
  user_emails = ["support@vantage.sh"]
}

resource "vantage_team" "demo_team_2" {
  name = "Another Demo Team"
  description = "Demo Team Description"
  user_tokens = ["usr_36b848747e1683bc", "usr_899b013c355547db"]
  workspace_tokens = ["wrkspc_47c3254c790e9351"]
}
