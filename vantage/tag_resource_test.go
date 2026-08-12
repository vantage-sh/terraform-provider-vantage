package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageTag_preferred(t *testing.T) {
	key := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageTagPreferred(key, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "preferred", "true"),
					resource.TestCheckResourceAttr(resourceName, "id", key),
				),
			},
			{
				Config: testAccVantageTagPreferred(key, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "preferred", "false"),
				),
			},
			{
				Config:             testAccVantageTagPreferred(key, false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageTag_preferredAndHidden(t *testing.T) {
	key := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageTagPreferredAndHidden(key, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "preferred", "true"),
					resource.TestCheckResourceAttr(resourceName, "hidden", "true"),
				),
			},
			{
				Config: testAccVantageTagPreferredAndHidden(key, true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "preferred", "true"),
					resource.TestCheckResourceAttr(resourceName, "hidden", "false"),
				),
			},
			{
				Config:             testAccVantageTagPreferredAndHidden(key, true, false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccVantageTagPreferred(key string, preferred bool) string {
	return fmt.Sprintf(`
resource "vantage_virtual_tag_config" "tag_key" {
  key         = %[1]q
  overridable = false
}

resource "vantage_tag" "test" {
  key       = vantage_virtual_tag_config.tag_key.key
  preferred = %[2]t
}
`, key, preferred)
}

func testAccVantageTagPreferredAndHidden(key string, preferred, hidden bool) string {
	return fmt.Sprintf(`
resource "vantage_virtual_tag_config" "tag_key" {
  key         = %[1]q
  overridable = false
}

resource "vantage_tag" "test" {
  key       = vantage_virtual_tag_config.tag_key.key
  preferred = %[2]t
  hidden    = %[3]t
}
`, key, preferred, hidden)
}
