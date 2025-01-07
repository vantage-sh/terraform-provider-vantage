package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageBudget_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rChildTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget.test"
	childResourceName := "vantage_budget.test_child"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetConfig_basic(rTitle, rChildTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rTitle),
					resource.TestCheckResourceAttr(childResourceName, "name", rChildTitle),
					resource.TestCheckResourceAttr(resourceName, "child_budget_tokens.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
			{
				Config: testAccVantageBudgetConfig_basic(rUpdatedTitle, rChildTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rUpdatedTitle),
				),
			},
		},
	})
}

func testAccVantageBudgetConfig_basic(budgetTitle string, childBudgetTitle string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_budget" "test_child" {
  name = %[2]q
  workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
}

resource "vantage_budget" "test" {
  name = %[1]q
  workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
  child_budget_tokens = [vantage_budget.test_child.token]
}
`, budgetTitle, childBudgetTitle)
}
