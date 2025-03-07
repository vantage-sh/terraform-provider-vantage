package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccCostReport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create cost report
				Config: costReportTF("test", "costs.provider = 'aws'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test", "title", "test"),
				),
			},
			{
				Config: costReportWithoutDatesTF("test", "costs.provider = 'aws'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test", "title", "test"),
					resource.TestCheckResourceAttrSet("vantage_cost_report.test", "start_date"),
					resource.TestCheckResourceAttrSet("vantage_cost_report.test", "end_date"),
				),
			},
		},
	})
}

func costReportTF(resourceTitle, filter string) string {
	return fmt.Sprintf(`
  data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "test" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "%s"
		filter = "%s"
		chart_type = "line"
		date_bin = "day"
		date_interval = "custom"
		start_date = "2025-01-01"
		end_date = "2025-01-31"
}`, resourceTitle, filter)
}

func costReportWithoutDatesTF(resourceTitle, filter string) string {
	return fmt.Sprintf(`
  data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "test" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "%s"
		filter = "%s"
		chart_type = "line"
		date_bin = "day"
		date_interval = "last_month"
}`, resourceTitle, filter)
}
