package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccAnomalyNotification_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalyNotificationCostReport() + testAccAnomalyNotification(10, "", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vantage_anomaly_notification.test", "token"),
					resource.TestCheckResourceAttr("vantage_anomaly_notification.test", "threshold", "10"),
				),
			},
			{ // update the threshold
				Config: testAccAnomalyNotificationCostReport() + testAccAnomalyNotification(20, "", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_anomaly_notification.test", "threshold", "20"),
				),
			},
			{ // update the channels
				Config: testAccAnomalyNotificationCostReport() + testAccAnomalyNotification(20, "recipient_channels = [\"test\"]", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_anomaly_notification.test", "threshold", "20"),
				),
			},
		},
	})
}

// func testAccAnomalyNotificationUsers() string {
// 	return `
// 	data "vantage_users" "users" {}

// 	`
// }

func testAccAnomalyNotificationCostReport() string {
	return `

data "vantage_workspaces" "workspaces" {}

resource "vantage_cost_report" "test" {
	workspace_token = data.vantage_workspaces.workspaces.workspaces[0].token
	title = "Test Cost Report"
	filter = "costs.provider = 'aws'"
	date_bin = "day"
	chart_type = "line"
}

`
}

func testAccAnomalyNotification(threshold int, channelsStr, userTokensStr string) string {
	return fmt.Sprintf(`
resource "vantage_anomaly_notification" "test" {
	cost_report_token = vantage_cost_report.test.token
	threshold = %d
	%s
	%s
}
	`, threshold, channelsStr, userTokensStr)
}
