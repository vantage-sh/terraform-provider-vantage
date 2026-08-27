resource "vantage_budget" "demo_budget" {
  name              = "Demo Budget"
  cost_report_token = vantage_cost_report.demo_report.token

  period_cadence = {
    starts_at      = "2024-01-22"
    interval_count = 2
    interval_unit  = "week"
  }

  periods = [
    {
      start_at = "2024-01-22"
      end_at   = "2024-02-04"
      amount   = 100
    }
  ]
}
