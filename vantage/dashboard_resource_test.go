package vantage

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccDashboard_basic(t *testing.T) {
	now := time.Now()
	beginningOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startDatePrevMonth := beginningOfCurrentMonth.AddDate(0, -1, 0).Format("2006-01-02")
	endDatePrevMonth := beginningOfCurrentMonth.AddDate(0, 0, -1).Format("2006-01-02")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create: without widgets
			{
				Config: testAccDashboard_basicTfDatasourceWorkspaces() +
					testAccDashboard_basicTf("test-no-widgets", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "date_interval", "last_month"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "end_date", endDatePrevMonth),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "start_date", startDatePrevMonth),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "title", "test-no-widgets"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "widgets.#", "0"),
					resource.TestCheckResourceAttrSet("vantage_dashboard.test-no-widgets", "workspace_token"),
				),
			},

			// Create: with widgets
			{
				Config: testAccDashboard_basicTfDatasourceWorkspaces() +
					testAccDashboard_basicTfReports("test-report") +
					testAccDashboard_basicTf(
						"test-with-widgets",
						`widgets = [
							{
								settings = { display_type = "table" }
								title = "Custom Widget Title",
								widgetable_token = vantage_resource_report.test-report.token
							}
						]`,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_resource_report.test-report", "title", "test-report"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "date_interval", "last_month"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "end_date", endDatePrevMonth),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "start_date", startDatePrevMonth),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "title", "test-with-widgets"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.#", "1"),

					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.0.title", "Custom Widget Title"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.0.settings.display_type", "table"),
					resource.TestCheckResourceAttrSet("vantage_dashboard.test-with-widgets", "widgets.0.widgetable_token"),

					resource.TestCheckResourceAttrSet("vantage_dashboard.test-with-widgets", "workspace_token"),
				),
			},

			// Update: remove widget
			{
				Config: testAccDashboard_basicTf("test-with-widgets", `widgets = []`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "date_interval", "last_month"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "end_date", endDatePrevMonth),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "start_date", startDatePrevMonth),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "title", "test-with-widgets"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.#", "0"),
					resource.TestCheckResourceAttrSet("vantage_dashboard.test-with-widgets", "workspace_token"),
				),
			},

			// Update: add widgets
			{
				Config: testAccDashboard_basicTfDatasourceWorkspaces() +
					testAccDashboard_basicTfReports("test-report-2") +
					testAccDashboard_basicTfReports("test-report-3") +
					testAccDashboard_basicTf(
						"test-with-widgets",
						`widgets = [
							{
								settings = { display_type = "table" }
								title = "Custom Widget Title (2)",
								widgetable_token = vantage_resource_report.test-report-2.token
							},
							{
								settings = { display_type = "chart" }
								title = "Custom Widget Title (3)",
								widgetable_token = vantage_resource_report.test-report-3.token
							}
						]`,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_resource_report.test-report-2", "title", "test-report-2"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "date_interval", "last_month"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "end_date", endDatePrevMonth),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "start_date", startDatePrevMonth),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "title", "test-with-widgets"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.#", "2"),

					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.0.title", "Custom Widget Title (2)"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.0.settings.display_type", "table"),
					resource.TestCheckResourceAttrSet("vantage_dashboard.test-with-widgets", "widgets.0.widgetable_token"),

					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.1.title", "Custom Widget Title (3)"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.1.settings.display_type", "chart"),
					resource.TestCheckResourceAttrSet("vantage_dashboard.test-with-widgets", "widgets.1.widgetable_token"),

					resource.TestCheckResourceAttrSet("vantage_dashboard.test-with-widgets", "workspace_token"),
				),
			},
		},
	})
}

func testAccDashboard_basicTfDatasourceWorkspaces() string {
	return `
		data "vantage_workspaces" "test" {}
		`
}

func testAccDashboard_basicTfReports(id string) string {
	return fmt.Sprintf(`
		resource "vantage_resource_report" %[1]q {
			workspace_token = data.vantage_workspaces.test.workspaces[0].token
			title = %[1]q
			filter = "resources.provider = 'aws'"
		}`, id)
}

func testAccDashboard_basicTf(id string, widgetsStr string) string {
	return fmt.Sprintf(`
		resource "vantage_dashboard" %[1]q {
		 	title = %[1]q
			date_interval = "last_month"
			%[2]s

		}`, id, widgetsStr)
}
