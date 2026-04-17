package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccMongodbProviderResource_basic(t *testing.T) {
	t.Skip("MongoDB integration is not yet supported by the vantage-go SDK")

	resourceName := "vantage_mongodb_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMongodbProviderConfig("mongodb+srv://cluster0.mongodb.net/test", "supersecretapikey"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cluster_uri", "mongodb+srv://cluster0.mongodb.net/test"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccMongodbProviderConfig(clusterURI, apiKey string) string {
	return `
resource "vantage_mongodb_provider" "test" {
  cluster_uri = "` + clusterURI + `"
  api_key     = "` + apiKey + `"
}
`
}
