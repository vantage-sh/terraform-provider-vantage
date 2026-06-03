data "vantage_workspace_by_name" "example" {
  name = "Production"
}

output "workspace_token" {
  value = data.vantage_workspace_by_name.example.token
}
