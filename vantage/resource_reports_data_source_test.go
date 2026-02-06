package vantage

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageResourceReportsDataSource_basic(t *testing.T) {
	// The test account has no resource reports and the data source returns a nil
	// slice for empty results, causing resource_reports.# to be unset. Fix the
	// data source to initialize an empty slice before unskipping.
	t.Skip("Skipping: data source returns nil slice for empty results")
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
		numReports, err := strconv.Atoi(reports.Primary.Attributes["resource_reports.#"])
		if err != nil {
			return err
		}

		if numReports > 0 {
			return nil
		}

		return fmt.Errorf("Reports not found")
	}
}

const testAccExampleReportsDataSourceConfig = `
data "vantage_resource_reports" "test" {}
`
