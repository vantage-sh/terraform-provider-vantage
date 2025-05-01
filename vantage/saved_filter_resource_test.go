package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageSavedFilter_basic(t *testing.T) {
	hasFilterId := "test-has-filter"
	hasFilterResourceName := "vantage_saved_filter.test-has-filter"
	hasFilterTitle := "Test Saved Filter with Filter"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageSavedFilter_basicTf(hasFilterId, hasFilterTitle, "(costs.provider = 'aws')"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(hasFilterResourceName, "title", hasFilterTitle),
					resource.TestCheckResourceAttr(hasFilterResourceName, "filter", "(costs.provider = 'aws')"),
					resource.TestCheckResourceAttrSet(hasFilterResourceName, "token"),
					resource.TestCheckResourceAttrSet(hasFilterResourceName, "workspace_token"),
				),
			},
		},
	})
}

func testAccVantageSavedFilter_basicTf(id, title, filter string) string {
	return fmt.Sprintf(
		`
		data "vantage_workspaces" "test" {}
		data "vantage_saved_filters" %[1]q {}

		 resource "vantage_saved_filter" %[1]q {
		   title = %[2]q
			 filter = %[3]q
			 workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
		 }
		`, id, title, filter,
	)
}
