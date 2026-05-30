terraform {
  required_providers {
    vantage = {
      source = "registry.terraform.io/vantage-sh/vantage"
    }
  }
}

provider "vantage" {
  host = "http://localhost:9090"
}

# Values with labels + label_filter in cost report tokens
resource "vantage_business_metric" "labeled_values" {
  title = "Labeled Values Test"

  values = [
    { date = "2026-01-01", amount = 100, label = "team-a" },
    { date = "2026-01-01", amount = 200, label = "team-b" },
    { date = "2026-02-01", amount = 150, label = "team-a" },
    { date = "2026-02-01", amount = 250, label = "team-b" },
  ]

  cost_report_tokens_with_metadata = [
    {
      cost_report_token = "rprt_report_alpha"
      unit_scale        = "per_unit"
      label_filter      = ["team-a"]
    },
    {
      cost_report_token = "rprt_report_beta"
      unit_scale        = "per_unit"
      label_filter      = ["team-b"]
    },
    {
      cost_report_token = "rprt_report_all"
      unit_scale        = "per_unit"
      label_filter      = []
    }
  ]
}
