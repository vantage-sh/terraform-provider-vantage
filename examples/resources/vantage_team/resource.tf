resource "vantage_team" "demo_team" {
  name                    = "Demo Team"
  description             = "Demo Team Description"
  default_dashboard_token = "dshbrd_a2846903070824f4"
  user_emails             = ["support@vantage.sh"]
  workspace_tokens        = ["wrkspc_47c3254c790e9351"]
}
