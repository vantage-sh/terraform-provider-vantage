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

func TestAccVantageFolder_withSavedFilterTokens(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_folder.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageFolderConfig_withSavedFilterTokens(rTitle, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "1"),
				),
			},
			{
				Config: testAccVantageFolderConfig_withSavedFilterTokens(rTitle, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "2"),
				),
			},
			{
				Config:             testAccVantageFolderConfig_withSavedFilterTokens(rTitle, 2),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageFolder_preservesSavedFilterTokensWhenOmitted(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_folder.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageFolderConfig_withSavedFilterTokens(rTitle, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "2"),
				),
			},
			{
				Config: testAccVantageFolderConfig_basic(rUpdatedTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rUpdatedTitle),
					resource.TestCheckResourceAttr(resourceName, "saved_filter_tokens.#", "2"),
				),
			},
			{
				Config:             testAccVantageFolderConfig_basic(rUpdatedTitle),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccVantageFolderConfig_withSavedFilterTokens(folderTitle string, filterCount int) string {
	config := `
data "vantage_workspaces" "test" {}

resource "vantage_saved_filter" "test1" {
  title           = "Test Filter 1"
  filter          = "(costs.provider = 'aws')"
  workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
}

resource "vantage_saved_filter" "test2" {
  title           = "Test Filter 2"
  filter          = "(costs.provider = 'gcp')"
  workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
}
`

	if filterCount == 1 {
		config += fmt.Sprintf(`
resource "vantage_folder" "test" {
  title               = %[1]q
  workspace_token     = element(data.vantage_workspaces.test.workspaces, 0).token
  saved_filter_tokens = [vantage_saved_filter.test1.token]
}
`, folderTitle)
	} else {
		config += fmt.Sprintf(`
resource "vantage_folder" "test" {
  title               = %[1]q
  workspace_token     = element(data.vantage_workspaces.test.workspaces, 0).token
  saved_filter_tokens = [vantage_saved_filter.test1.token, vantage_saved_filter.test2.token]
}
`, folderTitle)
	}

	return config
}
