data "vantage_invoices" "all_invoices" {}

resource "vantage_invoice" "demo_invoice" {
  account_token        = "acct_f87c7c90365bcdcd"
  billing_period_start = "2024-01-01"
  billing_period_end   = "2024-01-31"
}
