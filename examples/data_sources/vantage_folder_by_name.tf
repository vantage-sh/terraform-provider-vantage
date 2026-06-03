data "vantage_folder_by_name" "example" {
  title           = "Engineering"
  workspace_token = "wrkspc_1a2b3c4d5e6f"
}

output "folder_token" {
  value = data.vantage_folder_by_name.example.token
}
