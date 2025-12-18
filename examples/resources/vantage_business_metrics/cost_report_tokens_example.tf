# Example: Business Metric with Cost Report Token References
# This example demonstrates using Terraform references to cost reports
# in the cost_report_tokens_with_metadata field

data "vantage_workspaces" "default" {}

# Create multiple cost reports
resource "vantage_cost_report" "all_resources" {
  workspace_token = data.vantage_workspaces.default.workspaces[0].token
  title           = "All Resources"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "domains" {
  workspace_token = data.vantage_workspaces.default.workspaces[0].token
  title           = "Domains"
  filter          = "costs.service = 'Route 53'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "main_view" {
  workspace_token = data.vantage_workspaces.default.workspaces[0].token
  title           = "Main View"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "providers" {
  workspace_token = data.vantage_workspaces.default.workspaces[0].token
  title           = "Providers"
  filter          = "(costs.provider = 'aws' OR costs.provider = 'gcp')"
  date_interval   = "last_month"
}

# Create a business metric that references the cost reports
resource "vantage_business_metric" "fills_trades" {
  title = "Fills (Trades)"
  
  cost_report_tokens_with_metadata = [
    {
      cost_report_token = vantage_cost_report.all_resources.token
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = vantage_cost_report.domains.token
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = vantage_cost_report.main_view.token
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = vantage_cost_report.providers.token
      unit_scale        = "per_unit"
      label_filter      = []
    }
  ]
}
