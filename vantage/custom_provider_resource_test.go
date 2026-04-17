package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccCustomProviderResource_basic(t *testing.T) {
	resourceName := "vantage_custom_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomProviderConfig("Test Provider", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test Provider"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					// id and token must be identical
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "token"),
				),
			},
			// Verify no drift after creation — guards against the AccountIdentifier/Provider mix-up bug.
			{
				Config:             testAccCustomProviderConfig("Test Provider", ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccCustomProviderResource_withDescription(t *testing.T) {
	resourceName := "vantage_custom_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomProviderConfig("Described Provider", "A test description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Described Provider"),
					resource.TestCheckResourceAttr(resourceName, "description", "A test description"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
				),
			},
			// Verify no drift with description set.
			{
				Config:             testAccCustomProviderConfig("Described Provider", "A test description"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCustomProviderConfig(name, description string) string {
	if description == "" {
		return `
resource "vantage_custom_provider" "test" {
  name = "` + name + `"
}
`
	}
	return `
resource "vantage_custom_provider" "test" {
  name        = "` + name + `"
  description = "` + description + `"
}
`
}
