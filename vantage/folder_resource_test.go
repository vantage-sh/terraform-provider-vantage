package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageFolder_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_folder.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageFolderConfig_basic(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
			{
				Config: testAccVantageFolderConfig_basic(rUpdatedTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rUpdatedTitle),
				),
			},
		},
	})
}

func testAccVantageFolderConfig_basic(folderTitle string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_folder" "test" {
  title = %[1]q
	workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
}
`, folderTitle)
}
