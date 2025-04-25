package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestCostAlert(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a new cost alert
				Config: testAccCostAlertConfig(rTitle, "aws.product = 'EC2'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_alert.test", "title", rTitle),
					resource.TestCheckResourceAttr("vantage_cost_alert.test", "threshold", "100"),
					resource.TestCheckResourceAttr("vantage_cost_alert.test", "filter", "aws.product = 'EC2'"),
					resource.TestCheckResourceAttrSet("vantage_cost_alert.test", "token"),
				),
			},
			{
				// Update the title and filter
				Config: testAccCostAlertConfig(rUpdatedTitle, "aws.product = 'S3'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_cost_alert.test", "title", rUpdatedTitle),
					resource.TestCheckResourceAttr("vantage_cost_alert.test", "filter", "aws.product = 'S3'"),
				),
			},
		},
	})
}

func testAccCostAlertConfig(title, filter string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_cost_alert" "test" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
	title           = %[1]q
	threshold       = 100
	interval        = 7
	unit_type       = "day"
}
`, title, filter)
}
