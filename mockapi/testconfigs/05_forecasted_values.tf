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

# Both values and forecasted_values with 4 tokens
resource "vantage_business_metric" "with_forecast" {
  title = "With Forecast Test"

  values = [
    { date = "2025-01-01", amount = 1000 },
    { date = "2025-02-01", amount = 1100 },
    { date = "2025-03-01", amount = 1200 },
    { date = "2025-04-01", amount = 1300 },
  ]

  forecasted_values = [
    { date = "2026-01-01", amount = 1500 },
    { date = "2026-02-01", amount = 1600 },
    { date = "2026-03-01", amount = 1700 },
  ]

  cost_report_tokens_with_metadata = [
    {
      cost_report_token = "rprt_p1"
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = "rprt_p2"
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = "rprt_p3"
      unit_scale        = "per_thousand"
      label_filter      = []
    },
    {
      cost_report_token = "rprt_p4"
      unit_scale        = "per_million"
      label_filter      = []
    }
  ]
}
