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

# Values + cost report tokens: simulates the customer's scenario with csvdecode()
locals {
  order_volume_data = [
    { date = "2026-01-01", amount = 1000.50 },
    { date = "2026-02-01", amount = 1500.75 },
    { date = "2026-03-01", amount = 2000.25 },
    { date = "2026-04-01", amount = 1800.00 },
    { date = "2026-05-01", amount = 2200.50 },
  ]
}

resource "vantage_business_metric" "values_and_tokens" {
  title  = "Values and Tokens Test"
  values = local.order_volume_data

  cost_report_tokens_with_metadata = [
    {
      cost_report_token = "rprt_zzz"
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = "rprt_yyy"
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = "rprt_xxx"
      unit_scale        = "per_unit"
      label_filter      = []
    }
  ]
}

output "values_order" {
  value = [
    vantage_business_metric.values_and_tokens.values[0].date,
    vantage_business_metric.values_and_tokens.values[1].date,
    vantage_business_metric.values_and_tokens.values[2].date,
  ]
}

output "tokens_order" {
  value = [
    vantage_business_metric.values_and_tokens.cost_report_tokens_with_metadata[0].cost_report_token,
    vantage_business_metric.values_and_tokens.cost_report_tokens_with_metadata[1].cost_report_token,
    vantage_business_metric.values_and_tokens.cost_report_tokens_with_metadata[2].cost_report_token,
  ]
}
