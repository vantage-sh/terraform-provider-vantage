package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccAzureProviderResource_basic(t *testing.T) {
	t.Skip("Requires real Azure credentials (tenant, app_id, password)")

	resourceName := "vantage_azure_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureProviderConfig("tenant-123", "app-abc", "supersecret"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tenant", "tenant-123"),
					resource.TestCheckResourceAttr(resourceName, "app_id", "app-abc"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "token"),
				),
			},
		},
	})
}

func testAccAzureProviderConfig(tenant, appID, password string) string {
	return `
resource "vantage_azure_provider" "test" {
  tenant   = "` + tenant + `"
  app_id   = "` + appID + `"
  password = "` + password + `"
}
`
}
