data "vantage_virtual_tag_config_values" "allocation" {
  virtual_tag_config_token = "vtag_1234"
}

output "engineering_value_tokens" {
  value = [
    for value in data.vantage_virtual_tag_config_values.allocation.values : value.token
    if value.name == "Engineering"
  ]
}

output "values_by_token" {
  value = {
    for value in data.vantage_virtual_tag_config_values.allocation.values : value.token => value
  }
}