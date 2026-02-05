terraform {
  required_providers {
    vantage = {
      source = "vantage-sh/vantage"
    }
  }
}

# Get the default workspace
data "vantage_workspaces" "main" {}

locals {
  workspace_token = data.vantage_workspaces.main.workspaces[0].token
}

# -----------------------------------------------------------------------------
# Cost Reports for each child budget
# -----------------------------------------------------------------------------

resource "vantage_cost_report" "aws_compute" {
  workspace_token = local.workspace_token
  title           = "AWS Compute Costs"
  filter          = "costs.provider = 'aws' AND costs.service = 'Amazon Elastic Compute Cloud - Compute'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "aws_storage" {
  workspace_token = local.workspace_token
  title           = "AWS Storage Costs"
  filter          = "costs.provider = 'aws' AND costs.service = 'Amazon Simple Storage Service'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "aws_database" {
  workspace_token = local.workspace_token
  title           = "AWS Database Costs"
  filter          = "costs.provider = 'aws' AND costs.service = 'Amazon Relational Database Service'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "aws_fun" {
  workspace_token = local.workspace_token
  title           = "AWS Fun Costs"
  filter          = "costs.provider = 'aws' AND costs.service = 'Amazon Elastic Compute Cloud - Compute'"
  date_interval   = "last_month"
}
# -----------------------------------------------------------------------------
# Child Budgets - each tracks a specific cost category
# -----------------------------------------------------------------------------

resource "vantage_budget" "compute" {
  name              = "Compute Budget"
  workspace_token   = local.workspace_token
  cost_report_token = vantage_cost_report.aws_compute.token
  periods = [
    {
      start_at = "2025-01-01"
      end_at   = "2025-01-31"
      amount   = 5000
    },
    {
      start_at = "2025-02-01"
      end_at   = "2025-02-28"
      amount   = 5500
    }
  ]
}

resource "vantage_budget" "storage" {
  name              = "Storage Budget"
  workspace_token   = local.workspace_token
  cost_report_token = vantage_cost_report.aws_storage.token
  periods = [
    {
      start_at = "2025-01-01"
      end_at   = "2025-01-31"
      amount   = 2000
    },
    {
      start_at = "2025-02-01"
      end_at   = "2025-02-28"
      amount   = 2200
    }
  ]
}

resource "vantage_budget" "database" {
  name              = "Database Budget"
  workspace_token   = local.workspace_token
  cost_report_token = vantage_cost_report.aws_database.token
  periods = [
    {
      start_at = "2025-01-01"
      end_at   = "2025-01-31"
      amount   = 3000
    },
    {
      start_at = "2025-02-01"
      end_at   = "2025-02-28"
      amount   = 3300
    }
  ]
}

resource "vantage_budget" "fun" {
  name              = "Fun Budget 2"
  workspace_token   = local.workspace_token
  cost_report_token = vantage_cost_report.aws_fun.token
  periods = [
    {
      start_at = "2025-01-01"
      end_at   = "2025-01-31"
      amount   = 1000
    }
  ]
}
# -----------------------------------------------------------------------------
# Compound/Parent Budget - aggregates all child budgets
# -----------------------------------------------------------------------------

resource "vantage_budget" "infrastructure_total" {
  name            = "Total Infrastructure Budget"
  workspace_token = local.workspace_token

  # Reference child budgets - the parent budget will aggregate their totals
  child_budget_tokens = [
    vantage_budget.database.token,
    vantage_budget.fun.token,
    vantage_budget.compute.token,
    vantage_budget.storage.token,
  ]
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------

output "parent_budget_token" {
  description = "Token of the parent compound budget"
  value       = vantage_budget.infrastructure_total.token
}

output "child_budget_tokens" {
  description = "Tokens of all child budgets"
  value = {
    compute  = vantage_budget.compute.token
    storage  = vantage_budget.storage.token
    database = vantage_budget.database.token
    fun      = vantage_budget.fun.token
  }
}
