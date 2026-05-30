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

# Basic test: 3 cost report tokens in specific order
resource "vantage_business_metric" "three_tokens" {
  title = "Three Token Test"

  cost_report_tokens_with_metadata = [
    {
      cost_report_token = "rprt_aaa"
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = "rprt_bbb"
      unit_scale        = "per_thousand"
      label_filter      = []
    },
    {
      cost_report_token = "rprt_ccc"
      unit_scale        = "per_million"
      label_filter      = []
    }
  ]
}

output "token_order" {
  value = [
    vantage_business_metric.three_tokens.cost_report_tokens_with_metadata[0].cost_report_token,
    vantage_business_metric.three_tokens.cost_report_tokens_with_metadata[1].cost_report_token,
    vantage_business_metric.three_tokens.cost_report_tokens_with_metadata[2].cost_report_token,
  ]
}
