package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestFinancialCommitmentReport(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create resource report
				Config: testAccFinancialCommitmentReport(rTitle, "(financial_commitments.provider = 'aws' AND (financial_commitments.commitment_type IN ('on_demand','savings_plan')))", "all"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "title", rTitle),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "date_bucket", "week"),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "date_interval", "custom"),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "start_date", "2024-01-01"),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "end_date", "2024-01-31"),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "on_demand_costs_scope", "all"),
				),
			},
			{
				// update resource report
				Config: testAccFinancialCommitmentReport(rUpdatedTitle, "financial_commitments.provider = 'aws'", "discountable"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "title", rUpdatedTitle),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "filter", "financial_commitments.provider = 'aws'"),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "on_demand_costs_scope", "discountable"),
				),
			},
			{ // update resource report to date interval
				Config: testAccFinancialCommitmentReportDateInterval(rTitle, "last_7_days"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "title", rTitle),
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "date_interval", "last_7_days"),

					// even though the new terraform does not specify the on_demand_costs_scope, expect that it remains unchanged from the previous step.
					resource.TestCheckResourceAttr("vantage_financial_commitment_report.financial_commitment_report", "on_demand_costs_scope", "discountable"),
				),
			},
		},
	})
}

func testAccFinancialCommitmentReport(title, filter, onDemandCostScope string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_financial_commitment_report" "financial_commitment_report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title = %[1]q
	filter = %[2]q
	date_bucket = "week"
	date_interval = "custom"
	start_date = "2024-01-01"
	end_date = "2024-01-31"
	on_demand_costs_scope = %[3]q
}
	
	`, title, filter, onDemandCostScope)
}

func testAccFinancialCommitmentReportDateInterval(title, dateInterval string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_financial_commitment_report" "financial_commitment_report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title = %[1]q
	date_bucket = "week"
	date_interval = %[2]q
}
	`, title, dateInterval)

}

func TestFinancialCommitmentReport_withEmptyGroupings(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_financial_commitment_report.test_empty_groupings"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFinancialCommitmentReportConfig_withEmptyGroupings(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "groupings.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
			{
				Config: testAccFinancialCommitmentReportConfig_withEmptyGroupings(rUpdatedTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rUpdatedTitle),
					resource.TestCheckResourceAttr(resourceName, "groupings.#", "0"),
				),
			},
		},
	})
}

func testAccFinancialCommitmentReportConfig_withEmptyGroupings(title string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_financial_commitment_report" "test_empty_groupings" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
	title = %[1]q
	date_interval = "last_7_days"
	groupings = []
}
`, title)
}
