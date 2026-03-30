# Basic recommendation view
resource "vantage_recommendation_view" "production" {
  title           = "Production Recommendations"
  workspace_token = data.vantage_workspaces.main.workspaces[0].token
}

# Recommendation view with provider filter
resource "vantage_recommendation_view" "aws_only" {
  title           = "AWS Recommendations"
  workspace_token = data.vantage_workspaces.main.workspaces[0].token
  provider_ids    = ["aws"]
}

# Recommendation view with multiple filters
resource "vantage_recommendation_view" "filtered" {
  title           = "Filtered Recommendations"
  workspace_token = data.vantage_workspaces.main.workspaces[0].token
  provider_ids    = ["aws", "gcp"]
  regions         = ["us-east-1", "us-west-2"]
  start_date      = "2024-01-01"
  end_date        = "2024-12-31"
}

# Recommendation view with tag filter
resource "vantage_recommendation_view" "tagged" {
  title           = "Production Environment Recommendations"
  workspace_token = data.vantage_workspaces.main.workspaces[0].token
  tag_key         = "environment"
  tag_value       = "production"
}
