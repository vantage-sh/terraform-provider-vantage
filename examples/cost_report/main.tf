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
