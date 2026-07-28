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
	startDate := beginningOfCurrentMonth.AddDate(0, -1, 0).Format("2006-01-02")
	endDate := beginningOfCurrentMonth.AddDate(0, 0, -1).Format("2006-01-02")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create: without widgets
			{
				Config: testAccDashboard_basicTfDatasourceWorkspaces() +
					testAccDashboard_basicTf("test-no-widgets", startDate, endDate, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "end_date", endDate),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-widgets", "start_date", startDate),
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
						startDate,
						endDate,
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
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "end_date", endDate),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "start_date", startDate),
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
				Config: testAccDashboard_basicTfDatasourceWorkspaces() +
					testAccDashboard_basicTf("test-with-widgets", startDate, endDate, `widgets = []`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "end_date", endDate),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "start_date", startDate),
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
						startDate,
						endDate,
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
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "end_date", endDate),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "start_date", startDate),
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

func TestAccDashboard_withCostReportWidget(t *testing.T) {
	now := time.Now()
	beginningOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startDate := beginningOfCurrentMonth.AddDate(0, -1, 0).Format("2006-01-02")
	endDate := beginningOfCurrentMonth.AddDate(0, 0, -1).Format("2006-01-02")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{ // Create: with widgets
			{
				Config: testAccDashboard_basicTfDatasourceWorkspaces() +
					testAccDashboard_basicTfCostReport("test-report") +
					testAccDashboard_basicTf(
						"test-with-widgets",
						startDate,
						endDate,
						`widgets = [
							{
								settings = { display_type = "table" }
								title = "Custom Widget Title",
								widgetable_token = vantage_cost_report.test-report.token
							}
						]`,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-report", "title", "test-report"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "end_date", endDate),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "saved_filters.#", "0"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "start_date", startDate),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "title", "test-with-widgets"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.#", "1"),

					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.0.title", "Custom Widget Title"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-with-widgets", "widgets.0.settings.display_type", "table"),
					resource.TestCheckResourceAttrSet("vantage_dashboard.test-with-widgets", "widgets.0.widgetable_token"),

					resource.TestCheckResourceAttrSet("vantage_dashboard.test-with-widgets", "workspace_token"),
				),
			},
		},
	},
	)

}

func TestAccDashboard_hasDateInterval(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboard_withDateInterval(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "date_interval", "this_month"),
				),
			},
			{
				Config: testAccDashboard_nullDateInterval(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("vantage_dashboard.test-date-interval", "date_interval"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "start_date", ""),
				),
			},
			{
				Config: testAccDashboard_withDates(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "start_date", "2023-01-01"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "end_date", "2023-01-31"),
				),
			},
		}})
}

func TestAccDashboard_dateInterval(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboard_nullDateInterval(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("vantage_dashboard.test-date-interval", "date_interval"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "start_date", ""),
				),
			},
			{
				Config: testAccDashboard_withDateInterval(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "date_interval", "this_month"),
				),
			},
			{
				Config: testAccDashboard_nullDateInterval(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("vantage_dashboard.test-date-interval", "date_interval"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "start_date", ""),
				),
			},
			{
				Config: testAccDashboard_withDates(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "start_date", "2023-01-01"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "end_date", "2023-01-31"),
				),
			},
			{
				Config: testAccDashboard_nullDateInterval(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-date-interval", "start_date", ""),
				),
			},
		},
	})
}

// A null date_interval is only marked "known after apply" when the plan already
// differs from prior state, and it applies back to null, so it never sustains a
// diff of its own. Adding dates to a null interval must also survive apply even
// though toUpdate rewrites the interval to "custom".
func TestAccDashboard_nullDateIntervalNoDrift(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with no interval and no dates.
			{
				Config: testAccDashboard_noDrift("no-drift", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("vantage_dashboard.test-no-drift", "date_interval"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-drift", "start_date", ""),
				),
			},
			{
				Config:             testAccDashboard_noDrift("no-drift", ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},

			// Change an unrelated attribute: this is the only case where the
			// framework marks the null interval unknown.
			{
				Config: testAccDashboard_noDrift("no-drift-renamed", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-drift", "title", "no-drift-renamed"),
					resource.TestCheckNoResourceAttr("vantage_dashboard.test-no-drift", "date_interval"),
				),
			},
			{
				Config:             testAccDashboard_noDrift("no-drift-renamed", ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},

			// Add dates while the interval is null.
			{
				Config: testAccDashboard_noDrift("no-drift-renamed", `
					start_date = "2023-01-01"
					end_date   = "2023-01-31"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-drift", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-drift", "start_date", "2023-01-01"),
					resource.TestCheckResourceAttr("vantage_dashboard.test-no-drift", "end_date", "2023-01-31"),
				),
			},
			{
				Config: testAccDashboard_noDrift("no-drift-renamed", `
					start_date = "2023-01-01"
					end_date   = "2023-01-31"`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccDashboard_noDrift(title, dates string) string {
	return fmt.Sprintf(`
		data "vantage_workspaces" "test" {}

		resource "vantage_dashboard" "test-no-drift" {
			workspace_token = data.vantage_workspaces.test.workspaces[0].token
			title = %q
			%s
		}
	`, title, dates)
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

func testAccDashboard_basicTfCostReport(id string) string {
	return fmt.Sprintf(`
		resource "vantage_cost_report" %[1]q {
			workspace_token = data.vantage_workspaces.test.workspaces[0].token
			title = %[1]q
			filter = "costs.provider = 'aws'"
			date_bin = "day"
		  chart_type = "line"
		}`, id)
}

func testAccDashboard_basicTf(id, startDate, endDate, widgetsStr string) string {
	return fmt.Sprintf(`
		resource "vantage_dashboard" %[1]q {
			workspace_token = data.vantage_workspaces.test.workspaces[0].token
		 	title = %[1]q
			start_date = "%[2]s"
			end_date = "%[3]s"
			%[4]s

		}`, id, startDate, endDate, widgetsStr)
}

func testAccDashboard_nullDateInterval() string {
	return `
		data "vantage_workspaces" "test" {}

		resource "vantage_dashboard" "test-date-interval" {
			workspace_token = data.vantage_workspaces.test.workspaces[0].token
			title = "test"
		}
	`
}

func testAccDashboard_withDateInterval() string {
	return `
	data "vantage_workspaces" "test" {}

	resource "vantage_dashboard" "test-date-interval" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "test"
		date_interval = "this_month"
	}
`
}

func testAccDashboard_withDates() string {
	return `
	data "vantage_workspaces" "test" {}

	resource "vantage_dashboard" "test-date-interval" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "test"
		start_date = "2023-01-01"
		end_date = "2023-01-31"
	}
`
}
