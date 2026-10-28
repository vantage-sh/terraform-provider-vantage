resource "vantage_virtual_tag_config" "allocation" {
  key         = "Cost Center"
  overridable = false
}

resource "vantage_virtual_tag_config_value" "engineering" {
  virtual_tag_config_token = vantage_virtual_tag_config.allocation.token
  filter                   = "costs.provider = 'aws'"
  name                     = "Engineering"
}

resource "vantage_virtual_tag_config_value" "shared" {
  virtual_tag_config_token = vantage_virtual_tag_config.allocation.token
  filter                   = "costs.provider = 'gcp'"
  display_name             = "Shared costs"
  percentages = [
    { value = "Engineering", pct = 75 },
    { value = "Support", pct = 25 },
  ]
  date_ranges = [
    { start_date = "2026-01-01", end_date = null },
  ]
}