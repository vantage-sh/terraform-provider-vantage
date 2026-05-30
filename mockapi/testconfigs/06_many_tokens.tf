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

# 8 cost report tokens - stress test for ordering
resource "vantage_business_metric" "many_tokens" {
  title = "Many Tokens Test"

  cost_report_tokens_with_metadata = [
    { cost_report_token = "rprt_a1", unit_scale = "per_unit", label_filter = [] },
    { cost_report_token = "rprt_b2", unit_scale = "per_unit", label_filter = [] },
    { cost_report_token = "rprt_c3", unit_scale = "per_unit", label_filter = [] },
    { cost_report_token = "rprt_d4", unit_scale = "per_unit", label_filter = [] },
    { cost_report_token = "rprt_e5", unit_scale = "per_unit", label_filter = [] },
    { cost_report_token = "rprt_f6", unit_scale = "per_thousand", label_filter = [] },
    { cost_report_token = "rprt_g7", unit_scale = "per_million", label_filter = [] },
    { cost_report_token = "rprt_h8", unit_scale = "per_billion", label_filter = [] },
  ]
}

output "all_tokens" {
  value = [for t in vantage_business_metric.many_tokens.cost_report_tokens_with_metadata : t.cost_report_token]
}
