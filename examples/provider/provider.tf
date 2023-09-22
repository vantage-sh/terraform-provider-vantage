terraform {
  required_providers {
    vantage = {
      source = "vantage-sh/vantage"
    }
  }
}

provider "vantage" {
  # This can also be configured with the `VANTAGE_API_TOKEN` environment variable
  # and this block removed entirely:
  # export VANTAGE_API_TOKEN=an-api-token
  # terraform plan
  api_token = var.api_token
}

resource "vantage_folder" "aws" {
  title = "AWS Costs"
}

resource "vantage_cost_report" "aws" {
  folder_token = vantage_folder.aws.token
  filter       = "costs.provider = 'aws'"
  title        = "AWS Costs"
}
