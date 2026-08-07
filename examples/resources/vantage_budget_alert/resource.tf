resource "vantage_budget_alert" "demo_budget_alert" {
  budget_tokens      = [vantage_budget.demo_budget.token]
  threshold          = 100
  recipient_channels = ["#demo-cost-alerts"]
}

# Alerts on the first 7 days of the month instead of the full month. Omit
# duration_in_days to track the full month.
resource "vantage_budget_alert" "demo_first_week_alert" {
  budget_tokens      = [vantage_budget.demo_budget.token]
  threshold          = 80
  period_to_track    = "start_of_the_month"
  duration_in_days   = 7
  recipient_channels = ["#demo-cost-alerts"]
}
