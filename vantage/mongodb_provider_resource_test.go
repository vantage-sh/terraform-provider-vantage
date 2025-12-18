package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMongodbProviderResource_basic(t *testing.T) {
	resourceName := "vantage_mongodb_provider.demo"
	config := `
resource "vantage_mongodb_provider" "demo" {
  cluster_uri = "mongodb+srv://cluster0.mongodb.net/test"
  api_key = "supersecretapikey"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cluster_uri", "mongodb+srv://cluster0.mongodb.net/test"),
					resource.TestCheckResourceAttr(resourceName, "api_key", "supersecretapikey"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}