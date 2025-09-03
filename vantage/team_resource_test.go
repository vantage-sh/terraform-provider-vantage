package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestTeam(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create resource team
				Config: testAccTeam(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_team.team", "name", rName),
					resource.TestCheckResourceAttr("vantage_team.team", "description", ""),
				),
			},
			{
				// update resource team
				Config: testAccTeam(rUpdatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_team.team", "name", rUpdatedName),
					resource.TestCheckResourceAttr("vantage_team.team", "description", ""),
					resource.TestCheckResourceAttr("vantage_team.team", "workspace_tokens.#", "0"),
				),
			},
			{
				// update resource team with description
				Config: testAccTeamWithDescription(rUpdatedName, "test description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_team.team", "name", rUpdatedName),
					resource.TestCheckResourceAttr("vantage_team.team", "description", "test description"),
				),
			},
			{
				// update resource team with workspace tokens
				Config: testAccTeamWithWorkspaceTokens(rUpdatedName, "test description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_team.team", "name", rUpdatedName),
					resource.TestCheckResourceAttr("vantage_team.team", "description", "test description"),
					resource.TestCheckResourceAttr("vantage_team.team", "workspace_tokens.#", "1"),
					resource.TestCheckResourceAttr("vantage_team.team", "user_tokens.#", "0"),
					resource.TestCheckResourceAttr("vantage_team.team", "user_emails.#", "0"),
				),
			},
			{
				// update resource team with user tokens
				Config: testAccTeamWithUserTokens(rUpdatedName, "test description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_team.team", "name", rUpdatedName),
					resource.TestCheckResourceAttr("vantage_team.team", "description", "test description"),
					resource.TestCheckResourceAttr("vantage_team.team", "workspace_tokens.#", "0"),
					resource.TestCheckResourceAttr("vantage_team.team", "user_tokens.#", "1"),
					resource.TestCheckResourceAttr("vantage_team.team", "user_emails.#", "1"),
				),
			},
		},
	})
}

func testAccTeam(title string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}
resource "vantage_team" "team" {
	name = %[1]q
	description = ""
}
`, title)
}

func testAccTeamWithDescription(title, description string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}
resource "vantage_team" "team" {
	name = %[1]q
	description = %[2]q
}
`, title, description)
}

func testAccTeamWithWorkspaceTokens(title, description string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}
resource "vantage_team" "team" {
	workspace_tokens = [data.vantage_workspaces.test.workspaces[0].token]
	name = %[1]q
	description = %[2]q
}
	`, title, description)
}

func testAccTeamWithUserTokens(title, description string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}
data "vantage_users" "test" {}
resource "vantage_team" "team" {
	user_tokens = [data.vantage_users.test.users[0].token]
	name = %[1]q
	description = %[2]q
}
	`, title, description)
}
