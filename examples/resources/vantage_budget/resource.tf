resource "vantage_budget" "demo_budget" {
  name = "Demo Budget"
  cost_report_token = vantage_cost_report.demo_report.token
  periods = [
    {
      start_at = "2023-12-01"
      end_at = "2024-01-01"
      amount = 1000
    }
  ]
}
