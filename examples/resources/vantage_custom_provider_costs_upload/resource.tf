resource "vantage_custom_provider_costs_upload" "example" {
  provider_id = vantage_custom_provider.example.id
  period      = "2023-12"
  content     = file("${path.module}/costs.csv")
}