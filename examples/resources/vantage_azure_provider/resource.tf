resource "vantage_azure_provider" "example" {
  tenant   = "my-tenant-id"
  app_id   = "azure-app-client-id"
  password = "supersecret"
}