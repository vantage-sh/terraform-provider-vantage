package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageResourceReportsDataSource_basic(t *testing.T) {
	resourceName := "data.vantage_resource_reports.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleReportsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVantageCheckReportsExist(resourceName),
				),
			},
		},
	})
}

func testAccVantageCheckReportsExist(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reports, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if len(reports.Primary.Attributes) > 0 {
			return nil
		}

		return fmt.Errorf("Reports not found")
	}
}

const testAccExampleReportsDataSourceConfig = `
data "vantage_resource_reports" "test" {}
`
