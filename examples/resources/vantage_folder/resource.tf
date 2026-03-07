resource "vantage_folder" "demo_folder" {
  title = "Demo Folder"

  # Include either the parent_folder_token or workspace_token
  # If both are included, the API will use the parent_folder_token

  # Uncomment one of the following:
  # parent_folder_token = "fldr_47c3254c790e9351"
  workspace_token = "wrkspc_47c3254c790e9351"

  # Optional: Apply saved filters to all Cost Reports in this folder
  # saved_filter_tokens = ["svd_fltr_1234567890abcdef"]
}

