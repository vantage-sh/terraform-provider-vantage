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

# Mix of label_filter: some with values, some empty, some omitted
resource "vantage_business_metric" "mixed_filters" {
  title = "Mixed Label Filters Test"

  values = [
    { date = "2026-01-01", amount = 500, label = "service-a" },
    { date = "2026-01-01", amount = 300, label = "service-b" },
    { date = "2026-02-01", amount = 600, label = "service-a" },
    { date = "2026-02-01", amount = 400, label = "service-b" },
    { date = "2026-03-01", amount = 700, label = "service-a" },
  ]

  cost_report_tokens_with_metadata = [
    {
      cost_report_token = "rprt_all"
      unit_scale        = "per_unit"
      label_filter      = []
    },
    {
      cost_report_token = "rprt_svc_a"
      unit_scale        = "per_unit"
      label_filter      = ["service-a"]
    },
    {
      cost_report_token = "rprt_svc_b"
      unit_scale        = "per_unit"
      label_filter      = ["service-b"]
    },
    {
      cost_report_token = "rprt_multi"
      unit_scale        = "per_thousand"
      label_filter      = ["service-a", "service-b"]
    }
  ]
}
