resource "vantage_budget_alert" "demo_budget_alert" {
  budget_tokens = [vantage_budget.demo_budget.token]
  threshold     = 90

  # Days from the start or end of the month to evaluate. Use "" for the full month.
  duration_in_days = "7"
  period_to_track  = "start_of_the_month"

  # Organization users or addresses on a verified domain.
  recipient_emails = ["finops@example.com"]
}
