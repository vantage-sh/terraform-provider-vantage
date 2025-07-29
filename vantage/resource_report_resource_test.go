package vantage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccResourceReport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
			    // create resource report
				Config: testAccResourceReport("test", "resources.provider = 'aws'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_resource_report.resource_report", "title", "test"),
				),
			},
			{
				// update resource report
				Config: testAccResourceReport("test2", "resources.provider = 'aws'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_resource_report.resource_report", "title", "test2"),
				),
			},
			{
				// update resource report with columns
				Config: testAccResourceReportWithColumns("test3", "resources.provider = 'aws' and resources.type = 'aws_instance'", []string{"provider", "label", "region"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_resource_report.resource_report", "title", "test3"),
					resource.TestCheckResourceAttr("vantage_resource_report.resource_report", "columns.#", "3"),
					resource.TestCheckResourceAttr("vantage_resource_report.resource_report", "columns.0", "provider"),
					resource.TestCheckResourceAttr("vantage_resource_report.resource_report", "columns.1", "label"),
					resource.TestCheckResourceAttr("vantage_resource_report.resource_report", "columns.2", "region"),
				),
			},
		},
	})
}

func testAccResourceReport(title, filter string) string {
	return fmt.Sprintf(`

data "vantage_workspaces" "test" {}

resource "vantage_resource_report" "resource_report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title = %[1]q
	filter = %[2]q
}
`, title, filter)
}

func testAccResourceReportWithColumns(title, filter string, columns []string) string {
	columnsStr := `["` + strings.Join(columns, `", "`) + `"]`
	return fmt.Sprintf(`

data "vantage_workspaces" "test" {}

resource "vantage_resource_report" "resource_report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title = %[1]q
	filter = %[2]q
	columns = %[3]s
}
`, title, filter, columnsStr)
}
