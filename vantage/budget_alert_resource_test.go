package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageBudgetAlert_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget_alert.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetAlertConfig(rTitle, 80, "7", "start_of_the_month"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "duration_in_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "period_to_track", "start_of_the_month"),
					resource.TestCheckResourceAttr(resourceName, "budget_tokens.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
			{
				Config: testAccVantageBudgetAlertConfig(rTitle, 95, "14", "end_of_the_month"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold", "95"),
					resource.TestCheckResourceAttr(resourceName, "duration_in_days", "14"),
					resource.TestCheckResourceAttr(resourceName, "period_to_track", "end_of_the_month"),
				),
			},
			{
				Config:             testAccVantageBudgetAlertConfig(rTitle, 95, "14", "end_of_the_month"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccVantageBudgetAlertConfig(rTitle, 95, "", "end_of_the_month"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "duration_in_days", ""),
				),
			},
			{
				Config:             testAccVantageBudgetAlertConfig(rTitle, 95, "", "end_of_the_month"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVantageBudgetAlert_withRecipientEmails(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget_alert.test_emails"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetAlertConfig_recipientEmails(rTitle, `["finops@vantage.sh"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "recipient_emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recipient_emails.0", "finops@vantage.sh"),
				),
			},
			{
				Config: testAccVantageBudgetAlertConfig_recipientEmails(rTitle, `["finops@vantage.sh", "eng@vantage.sh"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "recipient_emails.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recipient_emails.1", "eng@vantage.sh"),
				),
			},
			{
				Config:             testAccVantageBudgetAlertConfig_recipientEmails(rTitle, `["finops@vantage.sh", "eng@vantage.sh"]`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageBudgetAlertsDataSource_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetAlertsDataSourceConfig(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_budget_alerts.test", "budget_alerts.#"),
				),
			},
		},
	})
}

func testAccVantageBudgetAlertBudget(budgetTitle string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test_budget_alert_report" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Budget Alert Test Report"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_budget" "test_budget_alert" {
  name              = %[1]q
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  cost_report_token = vantage_cost_report.test_budget_alert_report.token
}
`, budgetTitle)
}

func testAccVantageBudgetAlertConfig(budgetTitle string, threshold int, durationInDays, periodToTrack string) string {
	return testAccVantageBudgetAlertBudget(budgetTitle) + fmt.Sprintf(`
resource "vantage_budget_alert" "test" {
  budget_tokens    = [vantage_budget.test_budget_alert.token]
  threshold        = %[1]d
  duration_in_days = %[2]q
  period_to_track  = %[3]q
}
`, threshold, durationInDays, periodToTrack)
}

func testAccVantageBudgetAlertConfig_recipientEmails(budgetTitle, recipientEmails string) string {
	return testAccVantageBudgetAlertBudget(budgetTitle) + fmt.Sprintf(`
resource "vantage_budget_alert" "test_emails" {
  budget_tokens    = [vantage_budget.test_budget_alert.token]
  threshold        = 100
  duration_in_days = "7"
  recipient_emails = %[1]s
}
`, recipientEmails)
}

func testAccVantageBudgetAlertsDataSourceConfig(budgetTitle string) string {
	return testAccVantageBudgetAlertConfig(budgetTitle, 80, "7", "start_of_the_month") + `
data "vantage_budget_alerts" "test" {
  budget_token = vantage_budget.test_budget_alert.token

  depends_on = [vantage_budget_alert.test]
}
`
}
