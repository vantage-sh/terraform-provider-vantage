data "vantage_billing_profiles" "all" {}

output "all_billing_profiles" {
  value = data.vantage_billing_profiles.all
}

# Example of accessing specific billing profile properties
output "billing_profile_details" {
  value = {
    count     = length(data.vantage_billing_profiles.all.billing_profiles)
    nicknames = data.vantage_billing_profiles.all.billing_profiles[*].nickname
    tokens    = data.vantage_billing_profiles.all.billing_profiles[*].token
  }
}
