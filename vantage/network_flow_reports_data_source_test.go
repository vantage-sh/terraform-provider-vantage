package vantage

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccNetworkFlowReportsDataSource_basic(t *testing.T) {
	resourceName := "data.vantage_network_flow_reports.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create report
				Config: testAccNetworkFlowReport("test_title", "network_flow_logs.traffic_category = 'cross_az'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "title", "test_title"),
					resource.TestCheckResourceAttr("vantage_network_flow_report.report", "filter", "network_flow_logs.traffic_category = 'cross_az'"),
				),
			},
			{
				Config: testAccExampleNetworkFlowReportsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVantageCheckNetworkFlowReportsDataSourceExists(resourceName),
				),
			},
		},
	})
}

func testAccVantageCheckNetworkFlowReportsDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reports, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		numReports, err := strconv.Atoi(reports.Primary.Attributes["network_flow_reports.#"])
		if err != nil {
			return err
		}

		if numReports > 0 {
			return nil
		}

		return fmt.Errorf("Reports not found")
	}
}

const testAccExampleNetworkFlowReportsDataSourceConfig = `
data "vantage_network_flow_reports" "test" {}
`
