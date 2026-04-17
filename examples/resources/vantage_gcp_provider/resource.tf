resource "vantage_gcp_provider" "example" {
  project_id      = "my-gcp-project"
  billing_account = "000000-111111-222222"
  dataset_name    = "my_billing_dataset"
}