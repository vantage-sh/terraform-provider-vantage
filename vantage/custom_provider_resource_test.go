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
			// Step 1: Create and verify all fields including system-managed status.
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
			// Step 2: Confirm no drift even though status is system-managed.
			// UseStateForUnknown preserves the API-returned status in the plan
			// so it never appears as (known after apply).
			{
				Config:             testAccCustomProviderConfig("Test Provider", ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: Attempt to rename — the ImmutableAfterCreate plan modifier
			// reverts the value to its state value, so Terraform sees no effective
			// change and the user-defined name field remains stable.
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
			// Step 1: Create with description; verify all user-defined and
			// system-managed fields are populated.
			{
				Config: testAccCustomProviderConfig("Provider With Desc", "Initial description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Provider With Desc"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
				),
			},
			// Step 2: Confirm no drift. Verifies that system-managed status does
			// not cause the plan to appear non-empty on subsequent runs.
			{
				Config:             testAccCustomProviderConfig("Provider With Desc", "Initial description"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: Attempt to change description — plan modifier reverts it
			// to the state value, so user-defined fields remain stable and the
			// system-managed status field does not trigger a replacement.
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
