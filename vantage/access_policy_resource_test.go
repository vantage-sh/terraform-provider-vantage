package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageAccessPolicy_basic(t *testing.T) {
	resourceName := "vantage_access_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageAccessPolicyConfig_basic(
					"tf-acc-access-policy",
					"Initial access policy",
					"vantage.provider = 'aws'",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", "tf-acc-access-policy"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial access policy"),
					resource.TestCheckResourceAttr(resourceName, "policy.api_version", "v1"),
					resource.TestCheckResourceAttr(resourceName, "policy.filter", "vantage.provider = 'aws'"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					func(s *terraform.State) error {
						policy := s.RootModule().Resources[resourceName]
						team := s.RootModule().Resources["vantage_team.test"]
						if policy.Primary.Attributes["team_tokens.0"] != team.Primary.Attributes["token"] {
							return fmt.Errorf(
								"expected team_tokens.0 to be %s, got %s",
								team.Primary.Attributes["token"],
								policy.Primary.Attributes["team_tokens.0"],
							)
						}
						return nil
					},
				),
			},
			{
				Config: testAccVantageAccessPolicyConfig_basic(
					"tf-acc-access-policy-updated",
					"Updated access policy",
					"vantage.provider = 'gcp'",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", "tf-acc-access-policy-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated access policy"),
					resource.TestCheckResourceAttr(resourceName, "policy.filter", "vantage.provider = 'gcp'"),
				),
			},
			{
				Config: testAccVantageAccessPolicyConfig_basic(
					"tf-acc-access-policy-updated",
					"Updated access policy",
					"vantage.provider = 'gcp'",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageAccessPolicy_parenthesizedFilterNoDrift(t *testing.T) {
	resourceName := "vantage_access_policy.test"
	filter := "(vantage.provider = 'aws')"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageAccessPolicyConfig_basic(
					"tf-acc-access-policy-parens",
					"Parenthesized filter",
					filter,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "policy.filter", filter),
				),
			},
			{
				// Refresh + plan must keep the configured filter even if the API
				// stores a de-parenthesized canonical form.
				Config: testAccVantageAccessPolicyConfig_basic(
					"tf-acc-access-policy-parens",
					"Parenthesized filter",
					filter,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageAccessPolicy_clearDescription(t *testing.T) {
	resourceName := "vantage_access_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageAccessPolicyConfig_basic(
					"tf-acc-access-policy-desc",
					"Has a description",
					"vantage.provider = 'aws'",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Has a description"),
				),
			},
			{
				Config: testAccVantageAccessPolicyConfig_noDescription(
					"tf-acc-access-policy-desc",
					"vantage.provider = 'aws'",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "description"),
				),
			},
			{
				Config: testAccVantageAccessPolicyConfig_noDescription(
					"tf-acc-access-policy-desc",
					"vantage.provider = 'aws'",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageAccessPolicy_clearTeamTokens(t *testing.T) {
	resourceName := "vantage_access_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageAccessPolicyConfig_basic(
					"tf-acc-access-policy-teams",
					"Has teams",
					"vantage.provider = 'aws'",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "team_tokens.#", "1"),
				),
			},
			{
				Config: testAccVantageAccessPolicyConfig_emptyTeamTokens(
					"tf-acc-access-policy-teams",
					"Has teams",
					"vantage.provider = 'aws'",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "team_tokens.#", "0"),
				),
			},
			{
				Config: testAccVantageAccessPolicyConfig_emptyTeamTokens(
					"tf-acc-access-policy-teams",
					"Has teams",
					"vantage.provider = 'aws'",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccVantageAccessPolicyConfig_basic(title, description, filter string) string {
	return fmt.Sprintf(`
resource "vantage_team" "test" {
  name = "tf-acc-access-policy-team"
}

resource "vantage_access_policy" "test" {
  title       = %[1]q
  description = %[2]q
  team_tokens = [vantage_team.test.token]

  policy = {
    api_version = "v1"
    filter      = %[3]q
  }
}
`, title, description, filter)
}

func testAccVantageAccessPolicyConfig_noDescription(title, filter string) string {
	return fmt.Sprintf(`
resource "vantage_team" "test" {
  name = "tf-acc-access-policy-team"
}

resource "vantage_access_policy" "test" {
  title       = %[1]q
  team_tokens = [vantage_team.test.token]

  policy = {
    api_version = "v1"
    filter      = %[2]q
  }
}
`, title, filter)
}

func testAccVantageAccessPolicyConfig_emptyTeamTokens(title, description, filter string) string {
	return fmt.Sprintf(`
resource "vantage_team" "test" {
  name = "tf-acc-access-policy-team"
}

resource "vantage_access_policy" "test" {
  title       = %[1]q
  description = %[2]q
  team_tokens = []

  policy = {
    api_version = "v1"
    filter      = %[3]q
  }
}
`, title, description, filter)
}
