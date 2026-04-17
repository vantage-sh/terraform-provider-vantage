package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccElasticProviderResource_basic(t *testing.T) {
	t.Skip("Elastic integration is not yet supported by the vantage-go SDK")

	resourceName := "vantage_elastic_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticProviderConfig("test-elastic-api-key"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccElasticProviderConfig(apiKey string) string {
	return `
resource "vantage_elastic_provider" "test" {
  api_key = "` + apiKey + `"
}
`
}
