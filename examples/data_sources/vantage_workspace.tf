data "vantage_workspace" "example" {
  name = "Production"
}

output "workspace_token" {
  value = data.vantage_workspace.example.token
}
