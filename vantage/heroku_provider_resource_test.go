package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccHerokuProviderResource_basic(t *testing.T) {
	t.Skip("not yet implemented")
	resourceName := "vantage_heroku_provider.demo"
	config := `
resource "vantage_heroku_provider" "demo" {
  api_key = "heroku-api-key"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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