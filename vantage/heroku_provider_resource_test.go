package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHerokuProviderResource_basic(t *testing.T) {
	resourceName := "vantage_heroku_provider.demo"
	config := `
resource "vantage_heroku_provider" "demo" {
  api_key = "heroku-api-key"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key", "heroku-api-key"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}