package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageCanvas_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rPrompt := "Show me monthly costs by provider"
	resourceName := "vantage_canvas.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageCanvasConfig(rTitle, rPrompt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "prompt", rPrompt),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccVantageCanvasConfig(rUpdatedTitle, rPrompt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rUpdatedTitle),
					resource.TestCheckResourceAttr(resourceName, "prompt", rPrompt),
				),
			},
			{
				Config:             testAccVantageCanvasConfig(rUpdatedTitle, rPrompt),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"data",
				},
			},
		},
	})
}

func TestAccVantageCanvas_updatePrompt(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rPrompt := "Show me monthly costs by provider"
	rUpdatedPrompt := "Show me daily costs by service"
	resourceName := "vantage_canvas.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageCanvasConfig(rTitle, rPrompt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "prompt", rPrompt),
				),
			},
			{
				Config: testAccVantageCanvasConfig(rTitle, rUpdatedPrompt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "prompt", rUpdatedPrompt),
				),
			},
			{
				Config:             testAccVantageCanvasConfig(rTitle, rUpdatedPrompt),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccVantageCanvasConfig(title, prompt string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_canvas" "test" {
  title           = %[1]q
  prompt          = %[2]q
  workspace_token = element(data.vantage_workspaces.test.workspaces, 0).token
}
`, title, prompt)
}
