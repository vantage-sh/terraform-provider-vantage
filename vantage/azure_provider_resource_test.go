package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAzureProviderResource_basic(t *testing.T) {
	resourceName := "vantage_azure_provider.demo"
	config := `
resource "vantage_azure_provider" "demo" {
  tenant_id = "tenant-123"
  subscription_id = "sub-abc"
  client_id = "client-xyz"
  client_secret = "supersecret"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tenant_id", "tenant-123"),
					resource.TestCheckResourceAttr(resourceName, "subscription_id", "sub-abc"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "client-xyz"),
					resource.TestCheckResourceAttr(resourceName, "client_secret", "supersecret"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}