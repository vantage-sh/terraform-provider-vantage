resource "vantage_access_policy" "demo_access_policy" {
  title       = "AWS East Only Costs"
  description = "Limits cost visibility to just east coast AWS."
  team_tokens = [vantage_team.demo_team.token]

  policy = {
    api_version = "v1"
    filter      = "(vantage.provider = 'aws' AND vantage.region = 'us-east-1')"
  }
}
