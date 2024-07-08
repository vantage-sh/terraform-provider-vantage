package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccSegment_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSegment("Demo Segment", nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_segment.test", "title", "Demo Segment"),
					resource.TestCheckNoResourceAttr("vantage_segment.test", "description"),
				),
			},
			{
				Config: testAccSegment("Demo Segment2", nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_segment.test", "title", "Demo Segment2"),
					resource.TestCheckNoResourceAttr("vantage_segment.test", "description"),
				),
			},
			{
				Config: testAccSegment("Demo Segment2", stringPointer("This is still a demo segment")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_segment.test", "title", "Demo Segment2"),
					resource.TestCheckResourceAttr("vantage_segment.test", "description", "This is still a demo segment"),
				),
			},
		},
	})
}

func stringPointer(s string) *string {
	return &s
}

func testAccSegment(title string, description *string) string {
	var descriptionStr string
	if description != nil {
		descriptionStr = fmt.Sprintf(`description = %[1]q`, *description)
	}

	return fmt.Sprintf(`
	
	data "vantage_workspaces" "test" {}

	resource "vantage_segment" "test" {
  title = %[1]q
	%s
	workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
}
`, title, descriptionStr)
}
