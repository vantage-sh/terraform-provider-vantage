resource "vantage_custom_calculation" "tokens" {
  title       = "Token usage"
  description = "Token usage across both unit spellings."
  operation   = "sum"

  filter = {
    type       = "sum"
    filter     = "costs.provider = 'openai'"
    usage_unit = ["Tokens", "tokens"]
  }
}

resource "vantage_cost_report" "ai" {
  title                     = "AI usage"
  workspace_token           = var.workspace_token
  date_interval             = "last_month"
  custom_calculation_tokens = [vantage_custom_calculation.tokens.token]
}

variable "workspace_token" {
  type = string
}
