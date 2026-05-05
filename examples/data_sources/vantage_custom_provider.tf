data "vantage_custom_provider" "example" {
  token = "intgr_custom_provider_abc123"
}

output "custom_provider_status" {
  value = data.vantage_custom_provider.example.status
}
