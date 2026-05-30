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

# This config will be used for TITLE UPDATE test:
# Step 1: create with title "Title Update Test v1"
# Step 2: update to "Title Update Test v2" and verify no token reordering
resource "vantage_business_metric" "title_update" {
  title = "Title Update Test v2"

  cost_report_tokens_with_metadata = [
    { cost_report_token = "rprt_x1", unit_scale = "per_unit", label_filter = [] },
    { cost_report_token = "rprt_x2", unit_scale = "per_unit", label_filter = [] },
    { cost_report_token = "rprt_x3", unit_scale = "per_thousand", label_filter = [] },
    { cost_report_token = "rprt_x4", unit_scale = "per_million", label_filter = [] },
  ]
}
