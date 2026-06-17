data "vantage_workspaces" "demo" {}

resource "vantage_canvas" "demo" {
  title           = "Monthly Cost Overview"
  prompt          = "Show me monthly costs by provider"
  workspace_token = element(data.vantage_workspaces.demo.workspaces, 0).token
}
