resource "vantage_scenario_model" "demo" {
  title = "Hiring Plan"

  periods = [{
    start_at    = "2026-01-01"
    end_at      = "2026-03-31"
    amount      = 1250.75
    amount_type = "dollar"
  }]
}

resource "vantage_report_forecast" "demo" {
  title                 = "Board Plan"
  cost_report_token     = vantage_cost_report.demo_report.token
  scenario_model_tokens = [vantage_scenario_model.demo.token]
  set_as_default        = true
}
