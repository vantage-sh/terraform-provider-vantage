package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageAccessGrant_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create access grant for report
				Config: testAccVantageAccessGrantConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						resourceAccessGrant := s.RootModule().Resources["vantage_access_grant.access_grant"]
						resource := s.RootModule().Resources["vantage_resource_report.resource_report"]
						team := s.RootModule().Resources["vantage_team.team"]

						if resourceAccessGrant.Primary.Attributes["resource_token"] != resource.Primary.Attributes["token"] {
							return fmt.Errorf(
								"expected resource_token to be %s, got %s",
								resource.Primary.Attributes["token"],
								resourceAccessGrant.Primary.Attributes["resource_token"],
							)
						}

						if resourceAccessGrant.Primary.Attributes["team_token"] != team.Primary.Attributes["token"] {
							return fmt.Errorf(
								"expected team token to be %s, got %s",
								team.Primary.Attributes["token"],
								resourceAccessGrant.Primary.Attributes["team_token"])
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccVantageAccessGrantConfig_basic() string {
	return `
data "vantage_workspaces" "test" {}

resource "vantage_team" "team" {
	name = "test"
}
resource "vantage_resource_report" "resource_report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
	title = "test"
	filter = "resources.provider = 'aws'"
}

resource "vantage_access_grant" "access_grant" {
	team_token = vantage_team.team.token
	resource_token = vantage_resource_report.resource_report.token
}
`
}
