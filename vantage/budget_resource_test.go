package vantage

import (
	"fmt"
	"os"
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

func TestAccVantageBudget_multipleChildBudgets(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget.parent"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetConfig_multipleChildren(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rTitle),
					resource.TestCheckResourceAttr(resourceName, "child_budget_tokens.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "child_budget_tokens.0", "vantage_budget.child_1", "token"),
					resource.TestCheckResourceAttrPair(resourceName, "child_budget_tokens.1", "vantage_budget.child_2", "token"),
					resource.TestCheckResourceAttrPair(resourceName, "child_budget_tokens.2", "vantage_budget.child_3", "token"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
				),
			},
			{
				Config: testAccVantageBudgetConfig_multipleChildren(rUpdatedTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rUpdatedTitle),
					resource.TestCheckResourceAttr(resourceName, "child_budget_tokens.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "child_budget_tokens.0", "vantage_budget.child_1", "token"),
					resource.TestCheckResourceAttrPair(resourceName, "child_budget_tokens.1", "vantage_budget.child_2", "token"),
					resource.TestCheckResourceAttrPair(resourceName, "child_budget_tokens.2", "vantage_budget.child_3", "token"),
				),
			},
		},
	})
}

func TestAccVantageBudget_withPeriodCadence(t *testing.T) {
	// period_cadence requires API support (FIN-4289) and flexible_budget_periods on the account.
	if os.Getenv("VANTAGE_BUDGET_PERIOD_CADENCE_ACC") == "" {
		t.Skip("set VANTAGE_BUDGET_PERIOD_CADENCE_ACC=1 once staging/local API exposes period_cadence and flexible_budget_periods is enabled")
	}

	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget.test_cadence"
	var originalToken string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetConfig_withPeriodCadence(rTitle, "2024-01-22", 2, "week"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rTitle),
					resource.TestCheckResourceAttr(resourceName, "period_cadence.starts_at", "2024-01-22"),
					resource.TestCheckResourceAttr(resourceName, "period_cadence.interval_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "period_cadence.interval_unit", "week"),
					resource.TestCheckResourceAttr(resourceName, "periods.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.amount", "100"),
					resource.TestCheckResourceAttrWith(resourceName, "token", func(value string) error {
						originalToken = value
						if value == "" {
							return fmt.Errorf("token is empty")
						}
						return nil
					}),
				),
			},
			{
				// Core treats cadence as immutable once periods exist, so a
				// configured cadence change must replace the Budget.
				Config: testAccVantageBudgetConfig_withPeriodCadence(rTitle, "2024-02-01", 1, "month"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rTitle),
					resource.TestCheckResourceAttr(resourceName, "period_cadence.starts_at", "2024-02-01"),
					resource.TestCheckResourceAttr(resourceName, "period_cadence.interval_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "period_cadence.interval_unit", "month"),
					resource.TestCheckResourceAttrWith(resourceName, "token", func(value string) error {
						if value == originalToken {
							return fmt.Errorf("token did not change after cadence replacement")
						}
						return nil
					}),
				),
			},
			{
				Config:             testAccVantageBudgetConfig_withPeriodCadence(rTitle, "2024-02-01", 1, "month"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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

func testAccVantageBudgetConfig_multipleChildren(parentTitle string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "child_report_1" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Child Report 1"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "child_report_2" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Child Report 2"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_cost_report" "child_report_3" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Child Report 3"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_budget" "child_1" {
  name              = "Child Budget 1"
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.child_report_1.token
}

resource "vantage_budget" "child_2" {
  name              = "Child Budget 2"
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.child_report_2.token
}

resource "vantage_budget" "child_3" {
  name              = "Child Budget 3"
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.child_report_3.token
}

resource "vantage_budget" "parent" {
  name              = %[1]q
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  child_budget_tokens = [
    vantage_budget.child_1.token,
    vantage_budget.child_2.token,
    vantage_budget.child_3.token,
  ]
}
`, parentTitle)
}

func testAccVantageBudgetConfig_withPeriodCadence(budgetTitle, startsAt string, intervalCount int, intervalUnit string) string {
	startsAtLine := ""
	if startsAt != "" {
		startsAtLine = fmt.Sprintf("    starts_at      = %q\n", startsAt)
	}

	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test_budget_cadence_report" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Budget Cadence Test Report"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_budget" "test_cadence" {
  name              = %[1]q
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.test_budget_cadence_report.token

  period_cadence = {
%[2]s    interval_count = %[3]d
    interval_unit  = %[4]q
  }

  periods = [{
    start_at = "2024-01-22"
    end_at   = "2024-02-04"
    amount   = 100
  }]
}
`, budgetTitle, startsAtLine, intervalCount, intervalUnit)
}
