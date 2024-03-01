package vantage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageWorkspacesDataSource_basic(t *testing.T) {
	resourceName := "data.vantage_workspaces.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVantageCheckDefaultWorkspaceExists(resourceName),
				),
			},
		},
	})
}

func testAccVantageCheckDefaultWorkspaceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		workspaces, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		for k, v := range workspaces.Primary.Attributes {
			if strings.HasSuffix(k, "name") && v == "Default" {
				// Default workspace found
				return nil
			}
		}

		return fmt.Errorf("Default workspace not found")
	}
}

const testAccExampleDataSourceConfig = `
data "vantage_workspaces" "test" {}
`
