package vantage

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccResourceReportColumnsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceReportColumnsDataSource("aws_instance"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vantage_resource_report_columns.test", "resource_type", "aws_instance"),
					resource.TestCheckResourceAttrSet("data.vantage_resource_report_columns.test", "columns.#"),
					resource.TestCheckTypeSetElemAttr("data.vantage_resource_report_columns.test", "columns.*", "provider"),
					resource.TestCheckTypeSetElemAttr("data.vantage_resource_report_columns.test", "columns.*", "label"),
					resource.TestCheckTypeSetElemAttr("data.vantage_resource_report_columns.test", "columns.*", "accruedCosts"),
				),
			},
		},
	})
}

func TestAccResourceReportColumnsDataSourceInvalidResourceType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceReportColumnsDataSource("invalid_resource_type"),
				ExpectError: regexp.MustCompile("Invalid resource type: invalid_resource_type"),
			},
		},
	})
}

func testAccResourceReportColumnsDataSource(resourceType string) string {
	return `
data "vantage_resource_report_columns" "test" {
  resource_type = "` + resourceType + `"
}
`
}