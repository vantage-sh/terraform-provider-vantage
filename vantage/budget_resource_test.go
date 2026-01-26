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

func TestAccVantageBudget_withPeriodsUpdate(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget.test_periods"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetConfig_withPeriods(rTitle, 9000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rTitle),
					resource.TestCheckResourceAttr(resourceName, "periods.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.amount", "9000"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
			{
				Config: testAccVantageBudgetConfig_withPeriods(rTitle, 8900),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rTitle),
					resource.TestCheckResourceAttr(resourceName, "periods.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.amount", "8900"),
				),
			},
		},
	})
}

func TestAccVantageBudget_withEmptyPeriods(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget.test_empty_periods"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetConfig_withEmptyPeriods(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rTitle),
					resource.TestCheckResourceAttr(resourceName, "periods.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
			{
				Config: testAccVantageBudgetConfig_withEmptyPeriods(rUpdatedTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rUpdatedTitle),
					resource.TestCheckResourceAttr(resourceName, "periods.#", "0"),
				),
			},
		},
	})
}

func testAccVantageBudgetConfig_basic(budgetTitle string, childBudgetTitle string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test_budget_report" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Budget Test Report"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_budget" "test_child" {
  name = %[2]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.test_budget_report.token
}

resource "vantage_budget" "test" {
  name = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  child_budget_tokens = [vantage_budget.test_child.token]
}
`, budgetTitle, childBudgetTitle)
}

func testAccVantageBudgetConfig_withPeriods(budgetTitle string, amount int) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test_budget_periods_report" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Budget Periods Test Report"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_budget" "test_periods" {
  name = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.test_budget_periods_report.token

  periods = [{
    amount = %[2]d
    start_at = "2024-01-01"
  }]
}
`, budgetTitle, amount)
}

func testAccVantageBudgetConfig_withEmptyPeriods(budgetTitle string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test_budget_empty_periods_report" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Budget Empty Periods Test Report"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_budget" "test_empty_periods" {
  name = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.test_budget_empty_periods_report.token
  periods = []
}
`, budgetTitle)
}
