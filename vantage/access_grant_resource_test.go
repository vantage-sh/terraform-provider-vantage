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
				Config: testAccVantageAccessGrantConfig_basic("allowed"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_access_grant.test", "access", "allowed"),
					func(s *terraform.State) error {
						resourceAccessGrant := s.RootModule().Resources["vantage_access_grant.test"]
						resource := s.RootModule().Resources["vantage_resource_report.test"]
						team := s.RootModule().Resources["vantage_team.test"]

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
			{ // update access grant for report
				Config: testAccVantageAccessGrantConfig_basic("denied"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_access_grant.test", "access", "denied"),
				),
			},
		},
	})
}

func testAccVantageAccessGrantConfig_basic(access string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_team" "test" {
	name = "test"
}

resource "vantage_resource_report" "test" {
	filter = "resources.provider = 'aws'"
	title = "test"
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_access_grant" "test" {
	resource_token = vantage_resource_report.test.token
	team_token = vantage_team.test.token
	access = %[1]q
}
`, access)
}
