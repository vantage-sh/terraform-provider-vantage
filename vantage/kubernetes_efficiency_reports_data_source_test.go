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
			{
				Config: testAccExampleKubernetesReportsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVantageCheckKReportsExist(resourceName),
				),
			},
		},
	})
}

func testAccVantageCheckKReportsExist(resourceName string) resource.TestCheckFunc {
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

const testAccExampleKubernetesReportsDataSourceConfig = `
data "vantage_kubernetes_efficiency_reports" "test" {}
`
