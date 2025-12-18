resource "vantage_azure_provider" "example" {
  tenant_id       = "my-tenant-id"
  subscription_id = "my-subscription-id"
  client_id       = "azure-client-id"
  client_secret   = "supersecret"
}