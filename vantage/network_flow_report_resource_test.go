package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestNetworkFlowReport(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create resource report
				Config: testAccNetworkFlowReport(rTitle, "network_flow_logs.traffic_category = 'cross_az'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "title", rTitle),
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "start_date", "2024-01-01"),
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "end_date", "2024-01-31"),
				),
			},
			{
				// update resource report
				Config: testAccNetworkFlowReport(rUpdatedTitle, "network_flow_logs.traffic_category = 'foo'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "title", rUpdatedTitle),
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "filter", "network_flow_logs.traffic_category = 'foo'"),
				),
			},
			{ // update resource report to date interval
				Config: testAccNetworkFlowReportDateInterval(rTitle, "last_7_days"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "title", rTitle),
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "date_interval", "last_7_days"),
				),
			},
		},
	})
}

func testAccNetworkFlowReport(title, filter string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_network_flow_report" "report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title = %[1]q
	filter = %[2]q
	date_interval = "custom"
	start_date = "2024-01-01"
	end_date = "2024-01-31"
}
	
	`, title, filter)
}

func testAccNetworkFlowReportDateInterval(title, dateInterval string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_network_flow_report" "report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title = %[1]q
	date_interval = %[2]q
}
	`, title, dateInterval)

}
