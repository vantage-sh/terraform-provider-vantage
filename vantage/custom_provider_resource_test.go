package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
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
				// Step 3: Import by token and verify all importable fields match.
			// description is write-only (not returned by the API) so it is excluded.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description"},
			},
			// Step 4: Attempt to rename — the ImmutableAfterCreate plan modifier
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

func TestAccCustomProviderResource_withWorkspaces(t *testing.T) {
	resourceName := "vantage_custom_provider.test"
	workspaceName := "tf-acc-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create without workspaces; confirm workspaces is empty set.
			{
				Config: testAccCustomProviderConfig("Provider With Workspaces", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttr(resourceName, "workspaces.#", "0"),
				),
			},
			// Step 2: Add the test workspace; confirm it appears in state without
			// replacing the custom provider resource.
			{
				Config: testAccCustomProviderWorkspacesConfig("Provider With Workspaces", workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "workspaces.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						resourceName, "workspaces.*",
						"vantage_workspace.test", "token",
					),
				),
			},
			// Step 3: Confirm no drift after workspace update.
			{
				Config:             testAccCustomProviderWorkspacesConfig("Provider With Workspaces", workspaceName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Note: removing workspaces is not tested here. The Vantage API
			// requires workspace_tokens to be non-empty, so associations cannot
			// be fully cleared once set via this endpoint.
		},
	})
}

func testAccCustomProviderWorkspacesConfig(providerName, workspaceName string) string {
	return fmt.Sprintf(`
resource "vantage_workspace" "test" {
  name = %q
}

resource "vantage_custom_provider" "test" {
  name       = %q
  workspaces = [vantage_workspace.test.token]
}
`, workspaceName, providerName)
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
