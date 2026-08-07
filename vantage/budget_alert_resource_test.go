package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageBudgetAlert_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget_alert.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with a threshold of 100 percent.
			{
				Config: testAccVantageBudgetAlertConfig_basic(rTitle, 100),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold", "100"),
					resource.TestCheckResourceAttr(resourceName, "budget_tokens.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
					resource.TestCheckResourceAttr(resourceName, "user_tokens.#", "1"),
					// The API leaves period_to_track unset when the configuration
					// does not choose one, despite what its description says.
					resource.TestCheckNoResourceAttr(resourceName, "period_to_track"),
				),
			},
			// Update the threshold. Everything the API fills in has to survive the
			// update, otherwise a one-attribute change plans as a rewrite of the
			// whole alert.
			{
				Config: testAccVantageBudgetAlertConfig_basic(rTitle, 80),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("threshold"), knownvalue.Int64Exact(80)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("token"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("created_at"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("user_token"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("user_tokens"), knownvalue.NotNull()),
						// Derived values hold because the attributes behind them
						// do not move when only the threshold changes.
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("workspace_token"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("integration_provider"), knownvalue.NotNull()),
						// The API leaves period_to_track unset here. Null is still a
						// known value, which is the point: the plan states it instead
						// of deferring it to apply.
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("period_to_track"), knownvalue.Null()),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold", "80"),
				),
			},
			// Confirm no drift.
			{
				Config:             testAccVantageBudgetAlertConfig_basic(rTitle, 80),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Import the alert by token.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// The data source lists the alert.
			{
				Config: testAccVantageBudgetAlertConfig_basic(rTitle, 80) + testAccVantageBudgetAlertsDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_budget_alerts.test", "budget_alerts.0.token"),
				),
			},
		},
	})
}

// TestAccVantageBudgetAlert_fullMonth covers the duration_in_days design. An
// omitted duration tracks the full month, and the API cannot move an alert back
// to the full month, so clearing the field replaces the alert.
func TestAccVantageBudgetAlert_fullMonth(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget_alert.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without duration_in_days: the alert tracks the full month.
			{
				Config: testAccVantageBudgetAlertConfig_fullMonth(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "duration_in_days"),
					resource.TestCheckResourceAttr(resourceName, "threshold", "100"),
				),
			},
			// Confirm the omitted duration does not drift.
			{
				Config:             testAccVantageBudgetAlertConfig_fullMonth(rTitle),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Set a duration.
			{
				Config: testAccVantageBudgetAlertConfig_withDuration(rTitle, 7),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "duration_in_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "period_to_track", "start_of_the_month"),
				),
			},
			// Update the duration in place.
			{
				Config: testAccVantageBudgetAlertConfig_withDuration(rTitle, 14),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "duration_in_days", "14"),
				),
			},
			// Confirm no drift on a set duration.
			{
				Config:             testAccVantageBudgetAlertConfig_withDuration(rTitle, 14),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Removing the duration replaces the alert and tracks the full month.
			{
				Config: testAccVantageBudgetAlertConfig_fullMonth(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "duration_in_days"),
				),
			},
		},
	})
}

// TestAccVantageBudgetAlert_multipleBudgets covers the list fields with more than
// one element, which catches a schema that generated a list as a single object.
func TestAccVantageBudgetAlert_multipleBudgets(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget_alert.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Two budgets watched by one alert.
			{
				Config: testAccVantageBudgetAlertConfig_multipleBudgets(rTitle, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "budget_tokens.#", "2"),
				),
			},
			// Confirm no drift on a two-element list.
			{
				Config:             testAccVantageBudgetAlertConfig_multipleBudgets(rTitle, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Drop back to one budget.
			{
				Config: testAccVantageBudgetAlertConfig_multipleBudgets(rTitle, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "budget_tokens.#", "1"),
				),
			},
		},
	})
}

// testAccVantageBudgetAlertBudgets declares the cost report and the budgets that
// the alerts below watch.
func testAccVantageBudgetAlertBudgets(title string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

# The API refuses an alert that reaches nobody, so every alert below names a
# user. Users avoid depending on a connected Slack or Teams integration.
data "vantage_users" "test" {}

resource "vantage_cost_report" "test_budget_alert_report" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Budget Alert Test Report %[1]s"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_budget" "test_budget_alert" {
  name              = %[1]q
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.test_budget_alert_report.token
}
`, title)
}

func testAccVantageBudgetAlertConfig_basic(title string, threshold int) string {
	return testAccVantageBudgetAlertBudgets(title) + fmt.Sprintf(`
resource "vantage_budget_alert" "test" {
  budget_tokens = [vantage_budget.test_budget_alert.token]
  threshold     = %[1]d
  user_tokens   = [data.vantage_users.test.users[0].token]
}
`, threshold)
}

func testAccVantageBudgetAlertConfig_fullMonth(title string) string {
	return testAccVantageBudgetAlertBudgets(title) + `
resource "vantage_budget_alert" "test" {
  budget_tokens = [vantage_budget.test_budget_alert.token]
  threshold     = 100
  user_tokens   = [data.vantage_users.test.users[0].token]
}
`
}

func testAccVantageBudgetAlertConfig_withDuration(title string, durationInDays int) string {
	return testAccVantageBudgetAlertBudgets(title) + fmt.Sprintf(`
resource "vantage_budget_alert" "test" {
  budget_tokens    = [vantage_budget.test_budget_alert.token]
  threshold        = 100
  period_to_track  = "start_of_the_month"
  duration_in_days = %[1]d
  user_tokens      = [data.vantage_users.test.users[0].token]
}
`, durationInDays)
}

func testAccVantageBudgetAlertConfig_multipleBudgets(title string, bothBudgets bool) string {
	budgetTokens := "[vantage_budget.test_budget_alert.token]"
	if bothBudgets {
		budgetTokens = "[vantage_budget.test_budget_alert.token, vantage_budget.test_budget_alert_second.token]"
	}

	return testAccVantageBudgetAlertBudgets(title) + fmt.Sprintf(`
resource "vantage_budget" "test_budget_alert_second" {
  name              = "%[1]s-second"
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.test_budget_alert_report.token
}

resource "vantage_budget_alert" "test" {
  budget_tokens = %[2]s
  threshold     = 100
  user_tokens   = [data.vantage_users.test.users[0].token]
}
`, title, budgetTokens)
}

func testAccVantageBudgetAlertsDataSource() string {
	return `
data "vantage_budget_alerts" "test" {}
`
}
