# All integrations
data "vantage_integrations" "all" {}

# Only custom_provider integrations
data "vantage_integrations" "custom" {
  provider_filter = "custom_provider"
}

output "custom_provider_tokens" {
  value = [for i in data.vantage_integrations.custom.integrations : i.token]
}
