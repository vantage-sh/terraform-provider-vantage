resource "vantage_resource_report" "demo_resource_report" {
  workspace_token = "wrkspc_47c3254c790e9351"
  title           = "Demo Resource Report"
  filter          = "resources.provider = 'aws' AND resources.type = 'aws_cloudtrail'"
}
