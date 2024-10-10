package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccResourceReport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create resource report
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
