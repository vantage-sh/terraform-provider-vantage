package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPagerdutyProviderResource_basic(t *testing.T) {
	resourceName := "vantage_pagerduty_provider.demo"
	config := `
resource "vantage_pagerduty_provider" "demo" {
  api_key = "pagerduty-api-key"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key", "pagerduty-api-key"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}