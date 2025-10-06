resource "vantage_cost_report" "demo_report" {
  title               = "Demo Report"
  folder_token        = "fldr_3555785cd0409118"
  filter              = "costs.provider = 'aws'"
  saved_filter_tokens = ["svd_fltr_e844a2ccace05933", "svd_fltr_1b4b80a380ef4ba2"]
  workspace_token = "wrkspc_47c3254c790e9351"
  chart_type = "line" # Allowed: area, line, pie, bar, multi-bar
  date_bin = "day"    # Allowed: cumulative, day, week, month, quarter

  # optionally, use folder_token instead of workspace_token
  # folder_token = "fldr_47c3254c790e9351"
}
