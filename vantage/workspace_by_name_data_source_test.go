package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

// TestAccVantageWorkspaceByName_basic verifies that the vantage_workspace_by_name
// data source can look up a workspace by its display name and returns the correct
// token matching the created resource.
func TestAccVantageWorkspaceByName_basic(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	workspaceName := fmt.Sprintf("tf-test-ws-%s", rName)
	resourceName := "vantage_workspace.test"
	dataSourceName := "data.vantage_workspace_by_name.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceByNameDataSourceConfig(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "token", resourceName, "token"),
					resource.TestCheckResourceAttr(dataSourceName, "name", workspaceName),
				),
			},
		},
	})
}

func testAccWorkspaceByNameDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "vantage_workspace" "test" {
  name = %[1]q
}

data "vantage_workspace_by_name" "test" {
  name       = %[1]q
  depends_on = [vantage_workspace.test]
}
`, name)
}
