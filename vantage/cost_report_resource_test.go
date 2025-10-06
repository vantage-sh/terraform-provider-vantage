package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccCostReport_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create cost report
				Config: costReportTF("test-1", "test", "costs.provider = 'aws'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-1", "title", "test"),
				),
			},
			{ // update test-1
				Config: costReportWithoutDatesTF("test-1", "test-updated", "costs.provider = 'aws'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-1", "title", "test-updated"),
					resource.TestCheckResourceAttrSet("vantage_cost_report.test-1", "start_date"),
					resource.TestCheckResourceAttrSet("vantage_cost_report.test-1", "end_date"),
				),
			},
			{ // create cost report without dates (should default to last month)
				Config: costReportWithoutDatesTF("test-2", "test", "costs.provider = 'aws'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-2", "title", "test"),
					resource.TestCheckResourceAttrSet("vantage_cost_report.test-2", "start_date"),
					resource.TestCheckResourceAttrSet("vantage_cost_report.test-2", "end_date"),
				),
			},
			{ // create cost report with different chart types
				Config: costReportWithChartType("test-3", "test", "line"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-3", "title", "test"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-3", "chart_type", "line"),
				),
			},
			{ // update cost report with different chart types
				Config: costReportWithChartType("test-3", "test", "bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-3", "title", "test"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-3", "chart_type", "bar"),
				),
			},
			{ // create cost report with different date bins
				Config: costReportWithDateBin("test-4", "test", "day"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-4", "title", "test"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-4", "date_bin", "day"),
				),
			},
			{ // update cost report with different date bins
				Config: costReportWithDateBin("test-4", "test", "month"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-4", "title", "test"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-4", "date_bin", "month"),
				),
			},
		},
	})
}

func TestAccCostReport_grouping(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create cost report
				Config: costReportWithGrouping(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-grouping", "groupings", "service"),
				),
			},
			{
				Config: costReportWithoutGrouping(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-grouping", "groupings", ""),
				),
			},
		},
	},
	)
}

func costReportTF(resourceName, resourceTitle, filter string) string {
	return fmt.Sprintf(`
  data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "%s" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "%s"
		filter = "%s"
		date_interval = "custom"
		start_date = "2025-01-01"
		end_date = "2025-01-31"
}`, resourceName, resourceTitle, filter)
}

func costReportWithoutDatesTF(resourceName, resourceTitle, filter string) string {
	return fmt.Sprintf(`
  data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "%s" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "%s"
		filter = "%s"
		date_interval = "last_month"
}`, resourceName, resourceTitle, filter)
}

func costReportWithChartType(resourceName, resourceTitle, chartType string) string {
	return fmt.Sprintf(`
	data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "%s" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "%s"
		filter = "costs.provider = 'aws'"
		chart_type = "%s"
		date_interval = "last_7_days"
	}`, resourceName, resourceTitle, chartType)
}

func costReportWithDateBin(resourceName, resourceTitle, dateBin string) string {
	return fmt.Sprintf(`
	data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "%s" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "%s"
		filter = "costs.provider = 'aws'"
		date_bin = "%s"
		date_interval = "last_7_days"
	}`, resourceName, resourceTitle, dateBin)
}

func costReportWithGrouping() string {
	return `
  data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "test-grouping" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "test"
		filter = "costs.provider = 'aws'"
		chart_type = "line"
		date_bin = "day"
		groupings = "service"
}`
}

func costReportWithoutGrouping() string {
	return `
  data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "test-grouping" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "test"
		filter = "costs.provider = 'aws'"
		chart_type = "line"
		date_bin = "day"
}`
}
