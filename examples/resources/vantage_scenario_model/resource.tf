resource "vantage_scenario_model" "demo" {
  title = "Hiring Plan"

  periods = [
    {
      start_at    = "2026-01-01"
      end_at      = "2026-03-31"
      amount      = 1250.75
      amount_type = "dollar"
    },
    {
      start_at    = "2026-04-01"
      amount      = 10
      amount_type = "percent"
    }
  ]
}
