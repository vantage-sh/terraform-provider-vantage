package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomProviderCostsUploadResource_basic(t *testing.T) {
	resourceName := "vantage_custom_provider_costs_upload.demo"
	config := `
resource "vantage_custom_provider_costs_upload" "demo" {
  provider_id = 1
  period      = "2023-12"
  content     = "date,amount\n2023-12-01,100\n2023-12-02,200"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "provider_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "period", "2023-12"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
				),
			},
		},
	})
}