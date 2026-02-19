resource "vantage_cost_report" "demo_report" {
  title               = "Demo Report"
  folder_token        = "fldr_3555785cd0409118"
  filter              = "costs.provider = 'aws'"
  saved_filter_tokens = ["svd_fltr_e844a2ccace05933", "svd_fltr_1b4b80a380ef4ba2"]
  workspace_token = "wrkspc_47c3254c790e9351"
  chart_type = "line" # Allowed: area, line, pie, bar, multi-bar
  date_bin = "day"    # Allowed: cumulative, day, week, month, quarter

  settings {
    include_credits      = true
    include_refunds      = true
    include_discounts    = true
    include_tax          = true
    amortize             = false
    unallocated          = false
    aggregate_by         = "cost" # Allowed: cost, usage
    show_previous_period = false
  }

  # optionally, use folder_token instead of workspace_token
  # folder_token = "fldr_47c3254c790e9351"
}
