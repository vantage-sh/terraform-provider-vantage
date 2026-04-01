resource "vantage_workspace" "example" {
  name                       = "Example Workspace"
  currency                   = "USD"
  enable_currency_conversion = false
  exchange_rate_date         = "daily_rate"
}
