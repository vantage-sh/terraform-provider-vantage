package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageReportNotification_basic(t *testing.T) {
	costReportTitle := fmt.Sprint("cost report for notification ", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	costReportFilter := "costs.provider = 'aws'"

	id1 := "test-notification-1"
	resourceName1 := "vantage_report_notification.test-notification-1"
	title1 := "Test Notification 1"
	change1 := "dollars"
	frequency1 := "daily"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageReportNotification_CreateTeamTf("team-1") +
					costReportTF("test", costReportTitle, costReportFilter) +
					testAccVantageReportNotification_basicTf(id1, title1, change1, frequency1, costReportTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_team.team", "name", "team-1"),

					resource.TestCheckResourceAttr("vantage_cost_report.test", "title", costReportTitle),
					resource.TestCheckResourceAttrSet("vantage_cost_report.test", "token"),

					resource.TestCheckResourceAttr(resourceName1, "title", title1),
					resource.TestCheckResourceAttr(resourceName1, "change", change1),
					resource.TestCheckResourceAttr(resourceName1, "frequency", frequency1),
					resource.TestCheckResourceAttr(resourceName1, "user_tokens.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName1, "token"),
					resource.TestCheckResourceAttrSet(resourceName1, "cost_report_token"),
				),
			},
		},
	})
}

func testAccVantageReportNotification_basicTf(id, title, change, frequency, costReportTitle string) string {
	return fmt.Sprintf(
		`
		 resource "vantage_report_notification" %[1]q {
		   title = %[2]q
			 change = %[3]q
			 frequency = %[4]q
		   cost_report_token = resource.vantage_cost_report.test.token
			 user_tokens = [data.vantage_users.test.users[0].token]
			 workspace_token = data.vantage_workspaces.test.workspaces[0].token
		 }
		`, id, title, change, frequency, costReportTitle,
	)
}

func testAccVantageReportNotification_CreateTeamTf(name string) string {
	return fmt.Sprintf(`
data "vantage_users" "test" {}
resource "vantage_team" "team" {
	workspace_tokens = [data.vantage_workspaces.test.workspaces[0].token]
	user_tokens = [data.vantage_users.test.users[0].token]
	name = %[1]q
	description = ""
}
	`, name)
}
