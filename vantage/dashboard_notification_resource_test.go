package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageDashboardNotification_basic(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_dashboard_notification.test"
	title := fmt.Sprintf("tf-dash-notif-%s", rName)
	updatedTitle := fmt.Sprintf("tf-dash-notif-updated-%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageDashboardNotificationConfig(title, "weekly"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "frequency", "weekly"),
					resource.TestCheckResourceAttr(resourceName, "user_tokens.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "dashboard_token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
					resource.TestCheckResourceAttrSet(resourceName, "recipient_emails.#"),
				),
			},
			{
				Config: testAccVantageDashboardNotificationConfig(updatedTitle, "monthly"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", updatedTitle),
					resource.TestCheckResourceAttr(resourceName, "frequency", "monthly"),
					resource.TestCheckResourceAttr(resourceName, "user_tokens.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
			{
				Config:             testAccVantageDashboardNotificationConfig(updatedTitle, "monthly"),
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

func TestAccVantageDashboardNotificationsDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	title := fmt.Sprintf("tf-dash-notif-ds-%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageDashboardNotificationConfig(title, "daily") + `
data "vantage_dashboard_notifications" "all" {}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_dashboard_notifications.all", "dashboard_notifications.#"),
				),
			},
		},
	})
}

func testAccVantageDashboardNotificationConfig(title, frequency string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}
data "vantage_users" "test" {}

resource "vantage_dashboard" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "tf-dash-for-notif-%[1]s"
  date_interval   = "this_month"
}

resource "vantage_dashboard_notification" "test" {
  title            = %[1]q
  frequency        = %[2]q
  dashboard_token  = vantage_dashboard.test.token
  workspace_token  = data.vantage_workspaces.test.workspaces[0].token
  user_tokens      = [data.vantage_users.test.users[0].token]
}
`, title, frequency)
}
