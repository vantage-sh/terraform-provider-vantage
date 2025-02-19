package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestReportNotification(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create resource report
				Config: testAccReportNotification(rTitle, "\"#cloud-costs\""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_report_notification.report_notification", "title", rTitle),
					resource.TestCheckResourceAttr("vantage_report_notification.report_notification", "recipient_channels.0", "#cloud-costs"),
					resource.TestCheckResourceAttr("vantage_report_notification.report_notification", "frequency", "daily"),
					resource.TestCheckResourceAttr("vantage_report_notification.report_notification", "change", "dollars"),
				),
			},
			{
				// update resource report
				Config: testAccReportNotification(rUpdatedTitle, "\"#cloud-costs\",\"#cloud-costs2\""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_report_notification.report_notification", "title", rUpdatedTitle),
					resource.TestCheckResourceAttr("vantage_report_notification.report_notification", "recipient_channels.0", "#cloud-costs"),
					resource.TestCheckResourceAttr("vantage_report_notification.report_notification", "recipient_channels.1", "#cloud-costs2"),
				),
			},
		},
	})
}

func testAccReportNotification(title string, channels string) string {
	return fmt.Sprintf(`

data "vantage_workspaces" "workspaces" {}
data "vantage_users" "users" {}
data "vantage_cost_reports" "cost_reports" {}

resource "vantage_report_notification" "report_notification" {
	workspace_token = data.vantage_workspaces.workspaces.workspaces[0].token
  title = %[1]q
	user_tokens = data.vantage_users.users.users[*].token
	frequency = "daily"
	change = "dollars"
	cost_report_token = data.vantage_cost_reports.cost_reports.cost_reports[0].token
	recipient_channels = [%s]
}
`, title, channels)
}
