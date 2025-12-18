package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogProviderResource_basic(t *testing.T) {
	resourceName := "vantage_datadog_provider.demo"
	config := `
resource "vantage_datadog_provider" "demo" {
  api_key = "ddapikey"
  app_key = "ddappkey"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key", "ddapikey"),
					resource.TestCheckResourceAttr(resourceName, "app_key", "ddappkey"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}