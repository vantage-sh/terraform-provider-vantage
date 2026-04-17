package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

// testAccCostsCSV is a minimal FOCUS-compatible CSV for testing.
const testAccCostsCSV = `BilledCost,BillingCurrency,BillingPeriodStart,BillingPeriodEnd,ChargeCategory,ResourceId,ServiceName
10.00,USD,2024-01-01,2024-01-31,Usage,my-resource,MyService
20.00,USD,2024-01-01,2024-01-31,Usage,other-resource,MyService
`

func TestAccCustomProviderCostsUploadResource_basic(t *testing.T) {
	resourceName := "vantage_custom_provider_costs_upload.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCostsUploadConfig(testAccCostsCSV),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "integration_token"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "import_status"),
					// id and token must be identical
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "token"),
				),
			},
		},
	})
}

func testAccCostsUploadConfig(csvContent string) string {
	return `
resource "vantage_custom_provider" "test" {
  name = "Test Provider for Costs Upload"
}

resource "vantage_custom_provider_costs_upload" "test" {
  integration_token = vantage_custom_provider.test.token
  csv_content       = <<-CSV
` + csvContent + `CSV
}
`
}
