package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageReportNotification_basic(t *testing.T) {

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
				Config: testAccVantageReportNotification_basicTf(id1, title1, change1, frequency1),
				Check: resource.ComposeTestCheckFunc(
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

func testAccVantageReportNotification_basicTf(id, title, change, frequency string) string {
	return fmt.Sprintf(
		`
		data "vantage_users" "test" {}
		data "vantage_cost_reports" "test" {}
		data "vantage_workspaces" "test" {}

		 resource "vantage_report_notification" %[1]q {
		   title = %[2]q
					change = %[3]q
					frequency = %[4]q
					cost_report_token = data.vantage_cost_reports.test.cost_reports[0].token
					user_tokens = [data.vantage_users.test.users[0].token]
					workspace_token = data.vantage_workspaces.test.workspaces[0].token
		 }
		`, id, title, change, frequency,
	)
}
