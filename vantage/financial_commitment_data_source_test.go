package vantage

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccFinancialCommitmentReportsDataSource_basic(t *testing.T) {
	resourceName := "data.vantage_financial_commitment_reports.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create report
				Config: testAccFinancialCommitmentReport("test_title", "(financial_commitments.provider = 'aws' AND (financial_commitments.commitment_type IN ('on_demand','savings_plan')))", "all"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "title", "test_title"),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "filter", "(financial_commitments.provider = 'aws' AND (financial_commitments.commitment_type IN ('on_demand','savings_plan')))"),
				),
			},
			{
				Config: testAccExampleFinancialCommitmentReportsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVantageCheckFinancialCommitmentReportsDataSourceExists(resourceName),
				),
			},
		},
	})
}

func testAccVantageCheckFinancialCommitmentReportsDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reports, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		numReports, err := strconv.Atoi(reports.Primary.Attributes["financial_commitment_reports.#"])
		if err != nil {
			return err
		}

		if numReports > 0 {
			return nil
		}

		return fmt.Errorf("Reports not found")
	}
}

const testAccExampleFinancialCommitmentReportsDataSourceConfig = `
data "vantage_financial_commitment_reports" "test" {}
`
