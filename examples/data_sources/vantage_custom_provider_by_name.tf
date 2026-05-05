data "vantage_custom_provider_by_name" "example" {
  name            = "My Custom Provider"
  provider_filter = "custom_provider"
}

output "token" {
  value = data.vantage_custom_provider_by_name.example.token
}
