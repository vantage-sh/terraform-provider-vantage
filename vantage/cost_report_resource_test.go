package vantage

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestCostReport_requiresFolderOrWorkspace(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "vantage" {
  api_token = "test"
}

resource "vantage_cost_report" "test" {
  title = "test"
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Missing Attribute Configuration"),
			},
		},
	})
}

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

func TestAccCostReport_settings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create cost report with settings
				Config: costReportWithSettings("test-settings", "test-settings", true, true, true, true, false, false, "cost", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "title", "test-settings"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.include_credits", "true"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.include_refunds", "true"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.include_discounts", "true"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.include_tax", "true"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.amortize", "false"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.unallocated", "false"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.aggregate_by", "cost"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.show_previous_period", "false"),
				),
			},
			{ // update settings
				Config: costReportWithSettings("test-settings", "test-settings", false, false, false, false, true, true, "count", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.include_credits", "false"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.include_refunds", "false"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.include_discounts", "false"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.include_tax", "false"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.amortize", "true"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.unallocated", "true"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.aggregate_by", "count"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-settings", "settings.show_previous_period", "true"),
				),
			},
			{ // confirm no drift
				Config:             costReportWithSettings("test-settings", "test-settings", false, false, false, false, true, true, "count", true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func costReportWithSettings(resourceName, resourceTitle string, includeCredits, includeRefunds, includeDiscounts, includeTax, amortize, unallocated bool, aggregateBy string, showPreviousPeriod bool) string {
	return fmt.Sprintf(`
	data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "%s" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "%s"
		filter = "costs.provider = 'aws'"
		date_interval = "last_month"

		settings = {
			include_credits      = %t
			include_refunds      = %t
			include_discounts    = %t
			include_tax          = %t
			amortize             = %t
			unallocated          = %t
			aggregate_by         = "%s"
			show_previous_period = %t
		}
	}`, resourceName, resourceTitle, includeCredits, includeRefunds, includeDiscounts, includeTax, amortize, unallocated, aggregateBy, showPreviousPeriod)
}

func TestAccCostReport_chartSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create with chart_settings
				Config: costReportWithChartSettings("test-cs", "test-chart-settings", "date", "cost"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-cs", "title", "test-chart-settings"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-cs", "chart_settings.x_axis_dimension.#", "1"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-cs", "chart_settings.x_axis_dimension.0", "date"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-cs", "chart_settings.y_axis_dimension", "cost"),
				),
			},
			{ // update chart_settings
				Config: costReportWithChartSettings("test-cs", "test-chart-settings", "service", "count"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_report.test-cs", "chart_settings.x_axis_dimension.#", "1"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-cs", "chart_settings.x_axis_dimension.0", "service"),
					resource.TestCheckResourceAttr("vantage_cost_report.test-cs", "chart_settings.y_axis_dimension", "count"),
				),
			},
			{ // confirm no drift
				Config:             costReportWithChartSettings("test-cs", "test-chart-settings", "service", "count"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func costReportWithChartSettings(resourceName, resourceTitle, xAxisDimension, yAxisDimension string) string {
	return fmt.Sprintf(`
	data "vantage_workspaces" "test" {}

	resource "vantage_cost_report" "%s" {
		workspace_token = data.vantage_workspaces.test.workspaces[0].token
		title = "%s"
		filter = "costs.provider = 'aws'"
		date_interval = "last_7_days"
		chart_settings = {
			x_axis_dimension = ["%s"]
			y_axis_dimension = "%s"
		}
	}`, resourceName, resourceTitle, xAxisDimension, yAxisDimension)
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

// ---------------------------------------------------------------------------
// New comprehensive tests
// ---------------------------------------------------------------------------

func TestAccVantageCostReport_createUpdateNoDrift(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with all common fields
			{
				Config: testAccCostReportConfig_full(rTitle, "costs.provider = 'aws'", "last_month", "line", "cumulative", "service"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "filter", "costs.provider = 'aws'"),
					resource.TestCheckResourceAttr(resourceName, "date_interval", "last_month"),
					resource.TestCheckResourceAttr(resourceName, "chart_type", "line"),
					resource.TestCheckResourceAttr(resourceName, "date_bin", "cumulative"),
					resource.TestCheckResourceAttr(resourceName, "groupings", "service"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "start_date"),
					resource.TestCheckResourceAttrSet(resourceName, "end_date"),
				),
			},
			// Step 2: Update all mutable fields
			{
				Config: testAccCostReportConfig_full(rUpdatedTitle, "costs.provider = 'gcp'", "last_7_days", "bar", "day", "provider"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rUpdatedTitle),
					resource.TestCheckResourceAttr(resourceName, "filter", "costs.provider = 'gcp'"),
					resource.TestCheckResourceAttr(resourceName, "date_interval", "last_7_days"),
					resource.TestCheckResourceAttr(resourceName, "chart_type", "bar"),
					resource.TestCheckResourceAttr(resourceName, "date_bin", "day"),
					resource.TestCheckResourceAttr(resourceName, "groupings", "provider"),
				),
			},
			// Step 3: Verify no drift
			{
				Config:             testAccCostReportConfig_full(rUpdatedTitle, "costs.provider = 'gcp'", "last_7_days", "bar", "day", "provider"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageCostReport_customDates(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with custom dates
			{
				Config: testAccCostReportConfig_customDates(rTitle, "2025-01-01", "2025-01-31"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "date_interval", "custom"),
					resource.TestCheckResourceAttr(resourceName, "start_date", "2025-01-01"),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2025-01-31"),
				),
			},
			// Step 2: Update custom dates
			{
				Config: testAccCostReportConfig_customDates(rTitle, "2025-02-01", "2025-02-28"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "start_date", "2025-02-01"),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2025-02-28"),
				),
			},
			// Step 3: Switch from custom to preset date_interval
			{
				Config: testAccCostReportConfig_dateInterval(rTitle, "last_month"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "date_interval", "last_month"),
					resource.TestCheckResourceAttrSet(resourceName, "start_date"),
					resource.TestCheckResourceAttrSet(resourceName, "end_date"),
				),
			},
			// Step 4: Verify no drift
			{
				Config:             testAccCostReportConfig_dateInterval(rTitle, "last_month"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageCostReport_groupingsMultiple(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with single grouping
			{
				Config: testAccCostReportConfig_groupings(rTitle, "service"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "groupings", "service"),
				),
			},
			// Step 2: Update to multiple groupings (comma-separated)
			{
				Config: testAccCostReportConfig_groupings(rTitle, "provider,service,region"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "groupings", "provider,service,region"),
				),
			},
			// Step 3: Verify no drift with multiple groupings
			{
				Config:             testAccCostReportConfig_groupings(rTitle, "provider,service,region"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 4: Clear groupings
			{
				Config: testAccCostReportConfig_noGroupings(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "groupings", ""),
				),
			},
			// Step 5: Verify no drift after clearing
			{
				Config:             testAccCostReportConfig_noGroupings(rTitle),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageCostReport_previousPeriod(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with previous period dates
			{
				Config: testAccCostReportConfig_previousPeriod(rTitle, "2024-12-01", "2024-12-31"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "previous_period_start_date", "2024-12-01"),
					resource.TestCheckResourceAttr(resourceName, "previous_period_end_date", "2024-12-31"),
				),
			},
			// Step 2: Update previous period dates
			{
				Config: testAccCostReportConfig_previousPeriod(rTitle, "2024-11-01", "2024-11-30"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "previous_period_start_date", "2024-11-01"),
					resource.TestCheckResourceAttr(resourceName, "previous_period_end_date", "2024-11-30"),
				),
			},
			// Step 3: Verify no drift
			{
				Config:             testAccCostReportConfig_previousPeriod(rTitle, "2024-11-01", "2024-11-30"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageCostReport_savedFilterTokens(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	sfTitle1 := "tf-test-sf-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	sfTitle2 := "tf-test-sf-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with one saved filter
			{
				Config: testAccCostReportConfig_savedFilterTokens(rTitle, sfTitle1, sfTitle2, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "1"),
				),
			},
			// Step 2: Update to two saved filters
			{
				Config: testAccCostReportConfig_savedFilterTokens(rTitle, sfTitle1, sfTitle2, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "2"),
				),
			},
			// Step 3: Verify no drift
			{
				Config:             testAccCostReportConfig_savedFilterTokens(rTitle, sfTitle1, sfTitle2, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageCostReport_chartSettings(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with chart_settings
			{
				Config: testAccCostReportConfig_chartSettings(rTitle, "date", "cost"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "chart_settings.x_axis_dimension.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "chart_settings.x_axis_dimension.0", "date"),
					resource.TestCheckResourceAttr(resourceName, "chart_settings.y_axis_dimension", "cost"),
				),
			},
			// Step 2: Update chart_settings
			{
				Config: testAccCostReportConfig_chartSettings(rTitle, "service", "usage"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "chart_settings.x_axis_dimension.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "chart_settings.x_axis_dimension.0", "service"),
					resource.TestCheckResourceAttr(resourceName, "chart_settings.y_axis_dimension", "usage"),
				),
			},
			// Step 3: Verify no drift
			{
				Config:             testAccCostReportConfig_chartSettings(rTitle, "service", "usage"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helpers for new tests
// ---------------------------------------------------------------------------

func testAccCostReportConfig_full(title, filter, dateInterval, chartType, dateBin, groupings string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = %[2]q
  date_interval   = %[3]q
  chart_type      = %[4]q
  date_bin        = %[5]q
  groupings       = %[6]q
}
`, title, filter, dateInterval, chartType, dateBin, groupings)
}

func testAccCostReportConfig_customDates(title, startDate, endDate string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "custom"
  start_date      = %[2]q
  end_date        = %[3]q
}
`, title, startDate, endDate)
}

func testAccCostReportConfig_dateInterval(title, dateInterval string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = %[2]q
}
`, title, dateInterval)
}

func testAccCostReportConfig_groupings(title, groupings string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"
  groupings       = %[2]q
}
`, title, groupings)
}

func testAccCostReportConfig_noGroupings(title string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"
}
`, title)
}

func testAccCostReportConfig_previousPeriod(title, prevStart, prevEnd string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token            = data.vantage_workspaces.test.workspaces[0].token
  title                      = %[1]q
  filter                     = "costs.provider = 'aws'"
  date_interval              = "custom"
  start_date                 = "2025-01-01"
  end_date                   = "2025-01-31"
  previous_period_start_date = %[2]q
  previous_period_end_date   = %[3]q
}
`, title, prevStart, prevEnd)
}

func testAccCostReportConfig_chartSettings(title, xAxisDimension, yAxisDimension string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"
  chart_settings = {
    x_axis_dimension = [%[2]q]
    y_axis_dimension = %[3]q
  }
}
`, title, xAxisDimension, yAxisDimension)
}

func testAccCostReportConfig_savedFilterTokens(title, sfTitle1, sfTitle2 string, useBoth bool) string {
	tokenRefs := `[vantage_saved_filter.sf1.token]`
	if useBoth {
		tokenRefs = `[vantage_saved_filter.sf1.token, vantage_saved_filter.sf2.token]`
	}

	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_saved_filter" "sf1" {
  title           = %[2]q
  filter          = "(costs.provider = 'aws')"
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_saved_filter" "sf2" {
  title           = %[3]q
  filter          = "(costs.provider = 'gcp')"
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_cost_report" "test" {
  workspace_token     = data.vantage_workspaces.test.workspaces[0].token
  title               = %[1]q
  filter              = "costs.provider = 'aws'"
  date_interval       = "last_7_days"
  saved_filter_tokens = %[4]s
}
`, title, sfTitle1, sfTitle2, tokenRefs)
}

// ---------------------------------------------------------------------------
// Settings tests
// ---------------------------------------------------------------------------

func TestAccVantageCostReport_settings(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	// NOTE: UpdateCostReportSettings in vantage-go uses `bool` with `omitempty`,
	// so false values are omitted from the JSON payload on update. This means
	// settings can only be toggled from false→true via update, not true→false.
	// The test is structured accordingly: create with false values, update to true.

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with explicit settings (false values work on create via *bool pointers)
			{
				Config: testAccCostReportConfig_settings(rTitle, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "settings.amortize", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.include_credits", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.include_tax", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.aggregate_by", "cost"),
					resource.TestCheckResourceAttr(resourceName, "settings.include_discounts", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings.include_refunds", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.show_previous_period", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings.unallocated", "false"),
				),
			},
			// Step 2: Update false→true (works because true is non-zero for omitempty)
			{
				Config: testAccCostReportConfig_settings(rTitle, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "settings.amortize", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings.include_credits", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings.include_tax", "true"),
				),
			},
			// Step 3: Verify no drift
			{
				Config:             testAccCostReportConfig_settings(rTitle, true, true, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCostReportConfig_settings(title string, amortize, includeCredits, includeTax bool) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"
  settings = {
    aggregate_by         = "cost"
    amortize             = %[2]t
    complete_period      = false
    include_credits      = %[3]t
    include_discounts    = true
    include_refunds      = false
    include_tax          = %[4]t
    show_previous_period = true
    unallocated          = false
  }
}
`, title, amortize, includeCredits, includeTax)
}

// TestAccVantageCostReport_settingsCompletePeriod exercises the new
// `complete_period` bool inside the settings nested block. It follows the
// same false→true pattern as the other settings test because
// UpdateCostReportSettings.CompletePeriod is a plain bool with omitempty in
// vantage-go, so an explicit `false` is dropped from the update payload.
func TestAccVantageCostReport_settingsCompletePeriod(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with complete_period=false (works via *bool on create).
			{
				Config: testAccCostReportConfig_completePeriod(rTitle, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "settings.complete_period", "false"),
				),
			},
			// Step 2: Update false→true.
			{
				Config: testAccCostReportConfig_completePeriod(rTitle, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "settings.complete_period", "true"),
				),
			},
			// Step 3: Verify no drift.
			{
				Config:             testAccCostReportConfig_completePeriod(rTitle, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCostReportConfig_completePeriod(title string, completePeriod bool) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"
  settings = {
    aggregate_by         = "cost"
    amortize             = true
    complete_period      = %[2]t
    include_credits      = false
    include_discounts    = true
    include_refunds      = false
    include_tax          = true
    show_previous_period = true
    unallocated          = false
  }
}
`, title, completePeriod)
}

// TestAccVantageCostReport_dateBinHour verifies that the `hour` value on
// date_bin is accepted after the schema regeneration exposed it.
func TestAccVantageCostReport_dateBinHour(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test-hour"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: costReportWithDateBin("test-hour", rTitle, "hour"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "date_bin", "hour"),
				),
			},
			{
				Config:             costReportWithDateBin("test-hour", rTitle, "hour"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageCostReport_defaultForecast(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCostReportConfig_defaultForecast(rTitle, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_forecast.kind", "baseline"),
					resource.TestCheckNoResourceAttr(resourceName, "default_forecast.report_forecast_token"),
				),
			},
			{
				Config: testAccCostReportConfig_defaultForecast(rTitle, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "default_forecast.kind"),
				),
			},
			{
				Config: testAccCostReportConfig_defaultForecast(rTitle, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_forecast.kind", "baseline"),
				),
			},
			{
				Config:             testAccCostReportConfig_defaultForecast(rTitle, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCostReportConfig_defaultForecast(title string, includeDefaultForecast bool) string {
	defaultForecast := ""
	if includeDefaultForecast {
		defaultForecast = `
  default_forecast = {
    kind = "baseline"
  }`
	}

	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"
%[2]s
}
`, title, defaultForecast)
}

// ---------------------------------------------------------------------------
// Business metric tokens with metadata tests
// ---------------------------------------------------------------------------

func TestAccVantageCostReport_businessMetricMetadata(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	bmTitle := "tf-test-bm-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCostReportConfig_businessMetricMetadata(
					rTitle,
					bmTitle,
					"unit_cost",
					"Platform Unit Cost",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.calculation_type", "unit_cost"),
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.label", "Platform Unit Cost"),
				),
			},
			{
				Config: testAccCostReportConfig_businessMetricMetadata(
					rTitle,
					bmTitle,
					"gross_margin",
					"Platform Gross Margin",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.calculation_type", "gross_margin"),
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.label", "Platform Gross Margin"),
				),
			},
			{
				Config: testAccCostReportConfig_businessMetricMetadata(
					rTitle,
					bmTitle,
					"gross_margin",
					"Platform Gross Margin",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCostReportConfig_businessMetricMetadata(
	title string,
	businessMetricTitle string,
	calculationType string,
	label string,
) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_business_metric" "test" {
  title = %[2]q

  lifecycle {
    ignore_changes = [cost_report_tokens_with_metadata]
  }
}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"

  business_metric_tokens_with_metadata = [{
    business_metric_token = vantage_business_metric.test.token
    calculation_type      = %[3]q
    label                 = %[4]q
    unit_scale             = "per_unit"
  }]
}
`, title, businessMetricTitle, calculationType, label)
}

func TestAccVantageCostReport_businessMetricLabelFilters(t *testing.T) {
	businessMetricToken := os.Getenv("VANTAGE_CLICKHOUSE_BUSINESS_METRIC_TOKEN")
	if businessMetricToken == "" {
		t.Skip("VANTAGE_CLICKHOUSE_BUSINESS_METRIC_TOKEN must reference a ClickHouse BusinessMetric")
	}

	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCostReportConfig_businessMetricLabelFilters(rTitle, businessMetricToken, "platform"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.label_filters.team.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.label_filters.team.0", "platform"),
				),
			},
			{
				Config: testAccCostReportConfig_businessMetricLabelFilters(rTitle, businessMetricToken, "finops"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.label_filters.team.0", "finops"),
				),
			},
			{
				Config:             testAccCostReportConfig_businessMetricLabelFilters(rTitle, businessMetricToken, "finops"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCostReportConfig_businessMetricLabelFilters(title, businessMetricToken, team string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"

  business_metric_tokens_with_metadata = [{
    business_metric_token = %[2]q
    calculation_type      = "unit_cost"
    label_filters = {
      team = [%[3]q]
    }
    unit_scale = "per_unit"
  }]
}
`, title, businessMetricToken, team)
}

func TestAccVantageCostReport_businessMetricTokensWithMetadata(t *testing.T) {
	// This test requires existing business metric tokens. Skip if not available.
	t.Skip("Requires pre-existing business metric tokens")

	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"
	bmToken := "bsnss_mtrc_REPLACE_ME" // replace with a valid token to run

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with one business metric
			{
				Config: testAccCostReportConfig_businessMetrics(rTitle, bmToken, "per_unit"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.business_metric_token", bmToken),
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.unit_scale", "per_unit"),
				),
			},
			// Step 2: Update unit_scale
			{
				Config: testAccCostReportConfig_businessMetrics(rTitle, bmToken, "per_thousand"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "business_metric_tokens_with_metadata.0.unit_scale", "per_thousand"),
				),
			},
			// Step 3: Verify no drift
			{
				Config:             testAccCostReportConfig_businessMetrics(rTitle, bmToken, "per_thousand"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCostReportConfig_businessMetrics(title, bmToken, unitScale string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = %[1]q
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_7_days"
  business_metric_tokens_with_metadata {
    business_metric_token = %[2]q
    unit_scale            = %[3]q
  }
}
`, title, bmToken, unitScale)
}

// ---------------------------------------------------------------------------
// Array ordering verification tests
// ---------------------------------------------------------------------------

// TestAccVantageCostReport_savedFilterTokensOrdering verifies that the API
// returns saved_filter_tokens in the same order they were submitted,
// preventing perpetual drift. Regression test following the pattern from
// TestAccVantageBudget_multipleChildBudgets (sha 1398103).
func TestAccVantageCostReport_savedFilterTokensOrdering(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	sfTitle1 := "tf-test-sf-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	sfTitle2 := "tf-test-sf-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	sfTitle3 := "tf-test-sf-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with 3 saved filters
			{
				Config: testAccCostReportConfig_savedFilterTokensOrdering(rTitle, sfTitle1, sfTitle2, sfTitle3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "3"),
				),
			},
			// Update title only — saved_filter_tokens should not cause drift
			{
				Config: testAccCostReportConfig_savedFilterTokensOrdering(rTitle+"-updated", sfTitle1, sfTitle2, sfTitle3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "3"),
				),
			},
			// Verify no drift
			{
				Config:             testAccCostReportConfig_savedFilterTokensOrdering(rTitle+"-updated", sfTitle1, sfTitle2, sfTitle3),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCostReportConfig_savedFilterTokensOrdering(title, sfTitle1, sfTitle2, sfTitle3 string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_saved_filter" "sf1" {
  title           = %[2]q
  filter          = "(costs.provider = 'aws')"
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_saved_filter" "sf2" {
  title           = %[3]q
  filter          = "(costs.provider = 'gcp')"
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_saved_filter" "sf3" {
  title           = %[4]q
  filter          = "(costs.provider = 'azure')"
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_cost_report" "test" {
  workspace_token     = data.vantage_workspaces.test.workspaces[0].token
  title               = %[1]q
  filter              = "costs.provider = 'aws'"
  date_interval       = "last_7_days"
  saved_filter_tokens = [
    vantage_saved_filter.sf1.token,
    vantage_saved_filter.sf2.token,
    vantage_saved_filter.sf3.token,
  ]
}
`, title, sfTitle1, sfTitle2, sfTitle3)
}

// TestAccVantageCostReport_savedFilterTokensReorder verifies that the API
// preserves the new order when saved_filter_tokens are reordered on update.
// If this test fails, the API is not respecting the updated ordering.
func TestAccVantageCostReport_savedFilterTokensReorder(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	sfTitle1 := "tf-test-sf-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	sfTitle2 := "tf-test-sf-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	sfTitle3 := "tf-test-sf-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_cost_report.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with order [sf1, sf2, sf3]
			{
				Config: testAccCostReportConfig_savedFilterTokensReorder(rTitle, sfTitle1, sfTitle2, sfTitle3, "sf1", "sf2", "sf3"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "saved_filter_tokens.0", "vantage_saved_filter.sf1", "token"),
					resource.TestCheckResourceAttrPair(resourceName, "saved_filter_tokens.1", "vantage_saved_filter.sf2", "token"),
					resource.TestCheckResourceAttrPair(resourceName, "saved_filter_tokens.2", "vantage_saved_filter.sf3", "token"),
				),
			},
			// Step 2: Reorder to [sf3, sf1, sf2]
			{
				Config: testAccCostReportConfig_savedFilterTokensReorder(rTitle, sfTitle1, sfTitle2, sfTitle3, "sf3", "sf1", "sf2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "saved_filter_tokens.0", "vantage_saved_filter.sf3", "token"),
					resource.TestCheckResourceAttrPair(resourceName, "saved_filter_tokens.1", "vantage_saved_filter.sf1", "token"),
					resource.TestCheckResourceAttrPair(resourceName, "saved_filter_tokens.2", "vantage_saved_filter.sf2", "token"),
				),
			},
			// Step 3: Verify no drift after reorder
			{
				Config:             testAccCostReportConfig_savedFilterTokensReorder(rTitle, sfTitle1, sfTitle2, sfTitle3, "sf3", "sf1", "sf2"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCostReportConfig_savedFilterTokensReorder(title, sfTitle1, sfTitle2, sfTitle3, first, second, third string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_saved_filter" "sf1" {
  title           = %[2]q
  filter          = "(costs.provider = 'aws')"
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_saved_filter" "sf2" {
  title           = %[3]q
  filter          = "(costs.provider = 'gcp')"
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_saved_filter" "sf3" {
  title           = %[4]q
  filter          = "(costs.provider = 'azure')"
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}

resource "vantage_cost_report" "test" {
  workspace_token     = data.vantage_workspaces.test.workspaces[0].token
  title               = %[1]q
  filter              = "costs.provider = 'aws'"
  date_interval       = "last_7_days"
  saved_filter_tokens = [
    vantage_saved_filter.%[5]s.token,
    vantage_saved_filter.%[6]s.token,
    vantage_saved_filter.%[7]s.token,
  ]
}
`, title, sfTitle1, sfTitle2, sfTitle3, first, second, third)
}
