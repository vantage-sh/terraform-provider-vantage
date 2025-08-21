data "vantage_invoices" "all" {}

output "all_invoices" {
  value = data.vantage_invoices.all
}

# Example of accessing specific invoice properties
output "invoice_details" {
  value = {
    count  = length(data.vantage_invoices.all.invoices)
    tokens = data.vantage_invoices.all.invoices[*].token
    totals = data.vantage_invoices.all.invoices[*].total
  }
}
