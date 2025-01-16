package vantage

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccKubernetesEfficiencyReportsDataSource_basic(t *testing.T) {
	resourceName := "data.vantage_kubernetes_efficiency_reports.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create resource report
				Config: testAccKubernetesReport("test_title", "kubernetes.cluster_id = 'foo'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_kubernetes_efficiency_report.kubernetes_efficiency_report", "title", "test_title"),
					resource.TestCheckResourceAttr("vantage_kubernetes_efficiency_report.kubernetes_efficiency_report", "date_bucket", "week"),
				),
			},
			{
				Config: testAccExampleKubernetesReportsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVantageCheckKubernetesEfficiencyReportsDataSourceExists(resourceName),
					testAccCheckGroupings(resourceName),
				),
			},
		},
	})
}

func testAccVantageCheckKubernetesEfficiencyReportsDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reports, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		numReports, err := strconv.Atoi(reports.Primary.Attributes["kubernetes_efficiency_reports.#"])
		if err != nil {
			return err
		}

		if numReports > 0 {
			return nil
		}

		return fmt.Errorf("Reports not found")
	}
}

func testAccCheckGroupings(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reports, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		numReports, err := strconv.Atoi(reports.Primary.Attributes["kubernetes_efficiency_reports.#"])
		if err != nil {
			return err
		}

		// assume the last report is the one we just created
		groupingsKey := fmt.Sprintf("kubernetes_efficiency_reports.%d.groupings", numReports-1)
		if reports.Primary.Attributes[groupingsKey] != "namespace,label:app" {
			return fmt.Errorf("groupings should be 'namespace,label:app', got %s", reports.Primary.Attributes[groupingsKey])
		}
		return nil
	}

}

const testAccExampleKubernetesReportsDataSourceConfig = `
data "vantage_kubernetes_efficiency_reports" "test" {}
`
