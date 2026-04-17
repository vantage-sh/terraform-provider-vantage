resource "vantage_custom_provider" "example" {
  name = "My Custom Provider"
}

resource "vantage_custom_provider_costs_upload" "example" {
  integration_token = vantage_custom_provider.example.token
  csv_content       = file("${path.module}/costs.csv")
  auto_transform    = true
}