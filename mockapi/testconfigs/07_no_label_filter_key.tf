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

# Tests the case where label_filter is NOT specified (omitted from block)
# This triggers the "null vs empty list" edge case
resource "vantage_business_metric" "omitted_label_filter" {
  title = "Omitted Label Filter Test"

  cost_report_tokens_with_metadata = [
    {
      cost_report_token = "rprt_first"
      unit_scale        = "per_unit"
      # label_filter omitted - should default to []
    },
    {
      cost_report_token = "rprt_second"
      unit_scale        = "per_unit"
      # label_filter omitted - should default to []
    },
    {
      cost_report_token = "rprt_third"
      unit_scale        = "per_unit"
      # label_filter omitted - should default to []
    }
  ]
}
