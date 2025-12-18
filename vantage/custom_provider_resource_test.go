package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomProviderResource_basic(t *testing.T) {
	resourceName := "vantage_custom_provider.demo"
	config := `
resource "vantage_custom_provider" "demo" {
  name = "Test Provider"
  identifier = "unique_identifier"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test Provider"),
					resource.TestCheckResourceAttr(resourceName, "identifier", "unique_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}