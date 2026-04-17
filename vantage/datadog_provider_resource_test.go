package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccDatadogProviderResource_basic(t *testing.T) {
	t.Skip("Datadog integration is not yet supported by the vantage-go SDK")

	resourceName := "vantage_datadog_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogProviderConfig("ddapikey", "ddappkey"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key", "ddapikey"),
					resource.TestCheckResourceAttr(resourceName, "app_key", "ddappkey"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccDatadogProviderConfig(apiKey, appKey string) string {
	return `
resource "vantage_datadog_provider" "test" {
  api_key = "` + apiKey + `"
  app_key = "` + appKey + `"
}
`
}
