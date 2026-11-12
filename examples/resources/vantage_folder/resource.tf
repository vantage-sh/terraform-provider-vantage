resource "vantage_folder" "demo_folder" {
  title = "Demo Folder"

  # Optional: Create a folder for provider resources instead of cost reports
  # type = "ProviderResourceFolder"

  # Include either the parent_folder_token or workspace_token
  # If both are included, the API will use the parent_folder_token

  # Uncomment one of the following:
  # parent_folder_token = "fldr_47c3254c790e9351"
  workspace_token = "wrkspc_47c3254c790e9351"

  # Optional: Apply saved filters to all Cost Reports in this folder
  # saved_filter_tokens = ["svd_fltr_1234567890abcdef"]
}


# Dashboard Folders require the account feature to be enabled.
resource "vantage_folder" "dashboards" {
  title           = "Dashboards"
  type            = "DashboardFolder"
  workspace_token = "wrkspc_47c3254c790e9351"
}

resource "vantage_folder" "team_dashboards" {
  title               = "Team Dashboards"
  type                = "DashboardFolder"
  parent_folder_token = vantage_folder.dashboards.token
  # Set parent_folder_token = "" to move this folder to root.
  # saved_filter_tokens only applies to CostFolder; omit it here.
}

resource "vantage_dashboard" "team" {
  title        = "Team Dashboard"
  folder_token = vantage_folder.team_dashboards.token
  # The workspace is inferred from the folder.
}
