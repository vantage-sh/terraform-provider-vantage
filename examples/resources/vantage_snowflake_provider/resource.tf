resource "vantage_snowflake_provider" "example" {
  account_name = "my_account"
  user_name    = "my_user"
  password     = "supersecret"
  role         = "analyst"
}