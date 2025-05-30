package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestCostAlert(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create initial cost alert
			{
				Config: testAccWorkspacesDatasource() + testAccCostReport() + testAccCostAlertConfig(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_alert.test", "title", rTitle),
					resource.TestCheckResourceAttr("vantage_cost_alert.test", "threshold", "100"),
					resource.TestCheckResourceAttrSet("vantage_cost_alert.test", "token"),
				),
			},
			// Step 2: Update cost alert
			{
				Config: testAccWorkspacesDatasource() + testAccCostReport() + testAccCostAlertConfig(rUpdatedTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_alert.test", "title", rUpdatedTitle),
				),
			},
			// Step 3: Confirm no changes after apply
			{
				Config:             testAccWorkspacesDatasource() + testAccCostReport() + testAccCostAlertConfig(rUpdatedTitle),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 4: Check data source includes alert
			{
				Config: testAccWorkspacesDatasource() + testAccCostReport() + testAccCostAlertConfig(rUpdatedTitle) + testAccCostAlertsDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_cost_alerts.test", "cost_alerts.0.token"),
					resource.TestCheckResourceAttr("data.vantage_cost_alerts.test", "cost_alerts.0.title", rUpdatedTitle),
				),
			},
			// Step 5: Delete the created test cost alert
			{
				Config: testAccWorkspacesDatasource() + testAccCostReport(),
			},
			// Step 6: Delete alert
			{
				Config: testAccWorkspacesDatasource() + testAccCostReport() + testAccCostAlertsDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vantage_cost_alerts.test", "cost_alerts.#", "0"),
				),
			},
			// Step 7: Delete the created test cost report
			{
				Config: testAccWorkspacesDatasource(),
			},
		},
	})
}

func testAccWorkspacesDatasource() string {
	return `
data "vantage_workspaces" "test_workspace" {}
`
}

func testAccCostReport() string {
	return `
resource "vantage_cost_report" "test_cost_report" {
  workspace_token = data.vantage_workspaces.test_workspace.workspaces[0].token
  title           = "Test Cost Report"
  chart_type      = "line"
  date_bin        = "day"
  date_interval   = "last_month"
}
`
}

func testAccCostAlertConfig(title string) string {
	return fmt.Sprintf(`
resource "vantage_cost_alert" "test" {
  workspace_token = data.vantage_workspaces.test_workspace.workspaces[0].token
  title           = %[1]q
  threshold       = 100
  interval        = "day"
  unit_type       = "percentage"
  report_tokens   = [vantage_cost_report.test_cost_report.token]
}
`, title)
}

func testAccCostAlertsDataSource() string {
	return `
data "vantage_cost_alerts" "test" {}
`
}
