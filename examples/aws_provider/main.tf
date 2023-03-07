terraform {
  required_providers {
    vantage = {
      source = "registry.terraform.io/vantage-sh/vantage"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4"
    }
  }
  required_version = ">= 1.0.0"
}

provider "aws" {
  region = "us-east-1"
  assume_role {
    role_arn = var.aws_assume_role
  }
}


data "vantage_aws_provider_info" "default" {
}

data "aws_iam_policy_document" "vantage_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = [data.vantage_aws_provider_info.default.iam_role_arn]
    }
    condition {
      variable = "sts:ExternalId"
      test     = "StringEquals"
      values   = [data.vantage_aws_provider_info.default.external_id]
    }
  }
}

resource "aws_iam_role" "vantage_cross_account_connection" {
  name               = "vantage_cross_account_connection"
  assume_role_policy = data.aws_iam_policy_document.vantage_assume_role.json

  inline_policy {
    name   = "root"
    policy = data.vantage_aws_provider_info.default.root_policy
  }

  inline_policy {
    name   = "VantageAutoPilot"
    policy = data.vantage_aws_provider_info.default.autopilot_policy
  }

  inline_policy {
    name   = "VantageCloudWatchMetricsReadOnly"
    policy = data.vantage_aws_provider_info.default.cloudwatch_metrics_policy
  }

  inline_policy {
    name   = "VantageAdditionalResourceReadOnly"
    policy = data.vantage_aws_provider_info.default.additional_resources_policy
  }
}

resource "aws_iam_policy_attachment" "vantage_cross_account_connection" {
  name       = "vantage_cross_account_connection-view-only"
  roles      = [aws_iam_role.vantage_cross_account_connection.name]
  policy_arn = "arn:aws:iam::aws:policy/job-function/ViewOnlyAccess"
}

resource "vantage_aws_provider" "main" {
  cross_account_arn = aws_iam_role.vantage_cross_account_connection.arn
}
