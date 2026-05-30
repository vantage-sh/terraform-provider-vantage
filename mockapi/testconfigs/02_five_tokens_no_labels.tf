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

# 5 cost report tokens, no label_filter specified (omitted, uses default)
resource "vantage_business_metric" "five_tokens" {
  title = "Five Token Test"

  cost_report_tokens_with_metadata = [
    {
      cost_report_token = "rprt_token1"
      unit_scale        = "per_unit"
    },
    {
      cost_report_token = "rprt_token2"
      unit_scale        = "per_hundred"
    },
    {
      cost_report_token = "rprt_token3"
      unit_scale        = "per_thousand"
    },
    {
      cost_report_token = "rprt_token4"
      unit_scale        = "per_million"
    },
    {
      cost_report_token = "rprt_token5"
      unit_scale        = "per_billion"
    }
  ]
}
