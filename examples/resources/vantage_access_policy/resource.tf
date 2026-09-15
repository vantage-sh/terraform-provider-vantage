resource "vantage_access_policy" "demo_access_policy" {
  title       = "Engineering costs"
  description = "Limits cost visibility to the Engineering team."
  team_tokens = [vantage_team.demo_team.token]

  policy = {
    api_version = "v1"
    filter      = "(vantage.provider = 'aws')"
  }
}
