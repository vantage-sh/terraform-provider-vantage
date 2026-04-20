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
			// Step 1: Create.
			{
				Config: testAccCustomProviderConfig("Test Provider", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test Provider"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "token"),
				),
			},
			// Step 2: No drift after creation.
			{
				Config:             testAccCustomProviderConfig("Test Provider", ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: Attempt to rename — the plan modifier emits a warning and
			// reverts the value, so Terraform sees no effective change.
			{
				Config:             testAccCustomProviderConfig("Renamed Provider", ""),
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
			// Step 1: Create with description.
			{
				Config: testAccCustomProviderConfig("Provider With Desc", "Initial description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Provider With Desc"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
				),
			},
			// Step 2: No drift after creation.
			{
				Config:             testAccCustomProviderConfig("Provider With Desc", "Initial description"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: Attempt to change description — plan modifier reverts it,
			// so no effective change occurs.
			{
				Config:             testAccCustomProviderConfig("Provider With Desc", "Updated description"),
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
