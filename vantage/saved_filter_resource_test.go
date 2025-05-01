package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageSavedFilter_basic(t *testing.T) {

	id := "test-0"
	resourceName := "vantage_saved_filter.test-0"
	filter := "(costs.provider = 'aws')"
	title := "Test SavedFilter"
	titleUpdated := "Test SavedFilter Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create: with filter
			{
				Config: testAccVantageSavedFilter_basicTf(id, title, filter),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "filter", filter),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
			// Update: title
			{
				Config: testAccVantageSavedFilter_titleOnlyTf(id, titleUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "filter", filter),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
				),
			},
		},
	})
}

func testAccVantageSavedFilter_basicTf(id, title, filter string) string {
	return fmt.Sprintf(`
		data "vantage_workspaces" "test" {}
		data "vantage_saved_filters" %[1]q {}

		 resource "vantage_saved_filter" %[1]q {
		   title = %[2]q
			 filter = %[3]q
			 workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
		 }`, id, title, filter,
	)
}

func testAccVantageSavedFilter_titleOnlyTf(id, title string) string {
	return fmt.Sprintf(`
		data "vantage_workspaces" "test" {}
		data "vantage_saved_filters" %[1]q {}

		resource "vantage_saved_filter" %[1]q {
		  title = %[2]q
			 workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
		}`, id, title,
	)
}
