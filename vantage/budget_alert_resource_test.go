package vantage

import (
	"fmt"
	"os"
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
					resource.TestCheckResourceAttr(resourceName, "user_tokens.#", "1"),
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
				// Clearing duration_in_days covers the full month.
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
				Config: testAccVantageBudgetAlertConfig_recipientEmails(rTitle, `[data.vantage_users.test.users[0].email]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "recipient_emails.#", "1"),
					resource.TestCheckResourceAttrPair(
						resourceName, "recipient_emails.0",
						"data.vantage_users.test", "users.0.email",
					),
				),
			},
			{
				Config: testAccVantageBudgetAlertConfig_recipientEmails(rTitle, `[data.vantage_users.test.users[0].email, data.vantage_users.test.users[1].email]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "recipient_emails.#", "2"),
					resource.TestCheckResourceAttrPair(
						resourceName, "recipient_emails.1",
						"data.vantage_users.test", "users.1.email",
					),
				),
			},
			{
				Config: testAccVantageBudgetAlertConfig_recipientEmails(rTitle, `[data.vantage_users.test.users[1].email, data.vantage_users.test.users[0].email]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "recipient_emails.0",
						"data.vantage_users.test", "users.1.email",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "recipient_emails.1",
						"data.vantage_users.test", "users.0.email",
					),
				),
			},
			{
				Config:             testAccVantageBudgetAlertConfig_recipientEmails(rTitle, `[data.vantage_users.test.users[1].email, data.vantage_users.test.users[0].email]`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageBudgetAlert_switchesRecipientSource(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget_alert.test_switch"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetAlertConfig_switchRecipient(rTitle, "user_tokens", 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "user_tokens.0",
						"data.vantage_users.test", "users.0.token",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "recipient_emails.0",
						"data.vantage_users.test", "users.0.email",
					),
				),
			},
			{
				Config: testAccVantageBudgetAlertConfig_switchRecipient(rTitle, "recipient_emails", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "recipient_emails.0",
						"data.vantage_users.test", "users.1.email",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "user_tokens.0",
						"data.vantage_users.test", "users.1.token",
					),
				),
			},
			{
				Config:             testAccVantageBudgetAlertConfig_switchRecipient(rTitle, "recipient_emails", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageBudgetAlert_clearsRecipientChannels(t *testing.T) {
	if os.Getenv("VANTAGE_BUDGET_ALERT_CHANNEL_ACC") == "" {
		t.Skip("set VANTAGE_BUDGET_ALERT_CHANNEL_ACC=1 when the acceptance workspace has a notification integration")
	}

	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget_alert.test_channels"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetAlertConfig_channels(rTitle, `["#terraform-provider-test"]`),
				Check:  resource.TestCheckResourceAttr(resourceName, "recipient_channels.#", "1"),
			},
			{
				Config: testAccVantageBudgetAlertConfig_channels(rTitle, `[]`),
				Check:  resource.TestCheckResourceAttr(resourceName, "recipient_channels.#", "0"),
			},
			{
				Config:             testAccVantageBudgetAlertConfig_channels(rTitle, `[]`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// An update that touches only threshold must not drop the recipients that the
// config no longer mentions.
func TestAccVantageBudgetAlert_preservesRecipientsOnUnrelatedUpdate(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_budget_alert.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageBudgetAlertConfig(rTitle, 80, "7", "start_of_the_month"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_tokens.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recipient_emails.#", "1"),
				),
			},
			{
				Config: testAccVantageBudgetAlertConfig_thresholdOnly(rTitle, 90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "threshold", "90"),
					resource.TestCheckResourceAttr(resourceName, "user_tokens.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "recipient_emails.#", "1"),
				),
			},
			{
				Config:             testAccVantageBudgetAlertConfig_thresholdOnly(rTitle, 90),
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
					resource.TestCheckResourceAttr("data.vantage_budget_alerts.test", "budget_alerts.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.vantage_budget_alerts.test", "budget_alerts.0.token",
						"vantage_budget_alert.test", "token",
					),
					resource.TestCheckResourceAttr("data.vantage_budget_alerts.test", "budget_alerts.0.threshold", "80"),
					resource.TestCheckResourceAttr("data.vantage_budget_alerts.test", "budget_alerts.0.duration_in_days", "7"),
				),
			},
		},
	})
}

func testAccVantageBudgetAlertBudget(budgetTitle string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

data "vantage_users" "test" {}

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

// The API merges the emails of user_tokens into recipient_emails, so these
// configs set only one of the two to keep the applied value equal to config.
func testAccVantageBudgetAlertConfig(budgetTitle string, threshold int, durationInDays, periodToTrack string) string {
	return testAccVantageBudgetAlertBudget(budgetTitle) + fmt.Sprintf(`
resource "vantage_budget_alert" "test" {
  workspace_token  = data.vantage_workspaces.test.workspaces[0].token
  budget_tokens    = [vantage_budget.test_budget_alert.token]
  user_tokens      = [data.vantage_users.test.users[0].token]
  threshold        = %[1]d
  duration_in_days = %[2]q
  period_to_track  = %[3]q
}
`, threshold, durationInDays, periodToTrack)
}

// Same alert as testAccVantageBudgetAlertConfig, with every recipient list
// dropped from the config so they resolve from prior state.
func testAccVantageBudgetAlertConfig_thresholdOnly(budgetTitle string, threshold int) string {
	return testAccVantageBudgetAlertBudget(budgetTitle) + fmt.Sprintf(`
resource "vantage_budget_alert" "test" {
  workspace_token  = data.vantage_workspaces.test.workspaces[0].token
  budget_tokens    = [vantage_budget.test_budget_alert.token]
  threshold        = %[1]d
  duration_in_days = "7"
  period_to_track  = "start_of_the_month"
}
`, threshold)
}

func testAccVantageBudgetAlertConfig_recipientEmails(budgetTitle, recipientEmails string) string {
	return testAccVantageBudgetAlertBudget(budgetTitle) + fmt.Sprintf(`
resource "vantage_budget_alert" "test_emails" {
  workspace_token  = data.vantage_workspaces.test.workspaces[0].token
  budget_tokens    = [vantage_budget.test_budget_alert.token]
  threshold        = 100
  duration_in_days = "7"
  recipient_emails = %[1]s
}
`, recipientEmails)
}

func testAccVantageBudgetAlertConfig_switchRecipient(budgetTitle, recipientType string, userIndex int) string {
	return testAccVantageBudgetAlertBudget(budgetTitle) + fmt.Sprintf(`
resource "vantage_budget_alert" "test_switch" {
  workspace_token  = data.vantage_workspaces.test.workspaces[0].token
  budget_tokens    = [vantage_budget.test_budget_alert.token]
  threshold        = 100
  duration_in_days = "7"
  %[1]s            = [data.vantage_users.test.users[%[2]d].%[3]s]
}
`, recipientType, userIndex, map[string]string{
		"user_tokens":      "token",
		"recipient_emails": "email",
	}[recipientType])
}

func testAccVantageBudgetAlertConfig_channels(budgetTitle, recipientChannels string) string {
	return testAccVantageBudgetAlertBudget(budgetTitle) + fmt.Sprintf(`
resource "vantage_budget_alert" "test_channels" {
  workspace_token   = data.vantage_workspaces.test.workspaces[0].token
  budget_tokens     = [vantage_budget.test_budget_alert.token]
  threshold         = 100
  duration_in_days  = "7"
  recipient_emails  = [data.vantage_users.test.users[0].email]
  recipient_channels = %[1]s
}
`, recipientChannels)
}

func testAccVantageBudgetAlertsDataSourceConfig(budgetTitle string) string {
	return testAccVantageBudgetAlertConfig(budgetTitle, 80, "7", "start_of_the_month") + `
data "vantage_budget_alerts" "test" {
  budget_token = vantage_budget.test_budget_alert.token

  depends_on = [vantage_budget_alert.test]
}
`
}
