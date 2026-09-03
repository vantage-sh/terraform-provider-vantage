package vantage

import (
	"fmt"
	"testing"
	"time"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVantageReportForecast_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	modelTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_report_forecast.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccScenarioModelsPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageReportForecastConfig_basic(rTitle, modelTitle, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "scenario_model_tokens.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "set_as_default", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_default", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "cost_report_token"),
					resource.TestCheckResourceAttrPair(resourceName, "scenario_model_tokens.0", "vantage_scenario_model.test", "token"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"set_as_default"},
			},
			{
				Config: testAccVantageReportForecastConfig_basic(rUpdatedTitle, modelTitle, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rUpdatedTitle),
					resource.TestCheckResourceAttr(resourceName, "set_as_default", "false"),
					resource.TestCheckResourceAttr(resourceName, "is_default", "false"),
					resource.TestCheckResourceAttr(resourceName, "scenario_model_tokens.#", "1"),
				),
			},
			{
				Config:             testAccVantageReportForecastConfig_basic(rUpdatedTitle, modelTitle, false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageReportForecast_clearBusinessMetric(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	modelTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	metricTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_report_forecast.test_clear"
	now := time.Now().UTC()
	historicalDates := []string{
		now.AddDate(0, 0, -7).Format(time.DateOnly),
		now.AddDate(0, 0, -6).Format(time.DateOnly),
	}
	forecastDates := []string{
		now.Format(time.DateOnly),
		now.AddDate(0, 0, 1).Format(time.DateOnly),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccScenarioModelsPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageReportForecastConfig_withBusinessMetric(rTitle, modelTitle, metricTitle, historicalDates, forecastDates),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttrSet(resourceName, "business_metric_token"),
				),
			},
			{
				Config: testAccVantageReportForecastConfig_basicNamed(rTitle, modelTitle, "test_clear", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "business_metric_token", ""),
				),
			},
			{
				Config:             testAccVantageReportForecastConfig_basicNamed(rTitle, modelTitle, "test_clear", false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageReportForecastsDataSource_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	modelTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccScenarioModelsPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageReportForecastsDataSourceConfig(rTitle, modelTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_report_forecasts.test", "report_forecasts.#"),
				),
			},
		},
	})
}

func testAccVantageReportForecastConfig_basic(title, modelTitle string, setAsDefault bool) string {
	return testAccVantageReportForecastConfig_basicNamed(title, modelTitle, "test", setAsDefault)
}

func testAccVantageReportForecastConfig_basicNamed(title, modelTitle, resourceLabel string, setAsDefault bool) string {
	setAsDefaultLine := fmt.Sprintf("\n  set_as_default        = %t", setAsDefault)

	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Report Forecast Test Report"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_scenario_model" "test" {
  title = %q

  periods = [{
    start_at    = "2026-01-01"
    end_at      = "2026-03-31"
    amount      = 1000
    amount_type = "dollar"
  }]
}

resource "vantage_report_forecast" %q {
  title                 = %q
  cost_report_token     = vantage_cost_report.test.token
  scenario_model_tokens = [vantage_scenario_model.test.token]%s
}
`, modelTitle, resourceLabel, title, setAsDefaultLine)
}

func testAccVantageReportForecastConfig_withBusinessMetric(title, modelTitle, metricTitle string, historicalDates, forecastDates []string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_report" "test" {
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title           = "Report Forecast Test Report"
  filter          = "costs.provider = 'aws'"
  date_interval   = "last_month"
}

resource "vantage_scenario_model" "test" {
  title = %[1]q

  periods = [{
    start_at    = "2026-01-01"
    end_at      = "2026-03-31"
    amount      = 1000
    amount_type = "dollar"
  }]
}

resource "vantage_business_metric" "test" {
  title = %[2]q

  values = [
    { date = %[4]q, amount = "100.50" },
    { date = %[5]q, amount = "200.75" },
  ]

  forecasted_values = [
    { date = %[6]q, amount = "300.25" },
    { date = %[7]q, amount = "400.50" },
  ]
}

resource "vantage_report_forecast" "test_clear" {
  title                 = %[3]q
  cost_report_token     = vantage_cost_report.test.token
  scenario_model_tokens = [vantage_scenario_model.test.token]
  business_metric_token = vantage_business_metric.test.token
}
`, modelTitle, metricTitle, title, historicalDates[0], historicalDates[1], forecastDates[0], forecastDates[1])
}

func testAccVantageReportForecastsDataSourceConfig(title, modelTitle string) string {
	return fmt.Sprintf(`
%s

data "vantage_report_forecasts" "test" {
  cost_report_token = vantage_cost_report.test.token
  depends_on        = [vantage_report_forecast.test]
}
`, testAccVantageReportForecastConfig_basic(title, modelTitle, false))
}
