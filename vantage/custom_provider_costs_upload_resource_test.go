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
			// Step 1: create the upload and verify computed attributes are populated.
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
			// Step 2: verify the resource can be imported by its token.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				// csv_content is write-only and never returned by the API.
				ImportStateVerifyIgnore: []string{"csv_content", "auto_transform"},
			},
		},
	})
}

// TestAccCustomProviderCostsUploadResource_autoTransform verifies that setting
// auto_transform = true is accepted and does not cause perpetual drift.
func TestAccCustomProviderCostsUploadResource_autoTransform(t *testing.T) {
	resourceName := "vantage_custom_provider_costs_upload.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCostsUploadAutoTransformConfig(testAccCostsCSV),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttr(resourceName, "auto_transform", "true"),
				),
			},
			// Confirm no drift on a subsequent plan.
			{
				Config:             testAccCostsUploadAutoTransformConfig(testAccCostsCSV),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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

func testAccCostsUploadAutoTransformConfig(csvContent string) string {
	return `
resource "vantage_custom_provider" "test" {
  name = "Test Provider for Costs Upload Auto Transform"
}

resource "vantage_custom_provider_costs_upload" "test" {
  integration_token = vantage_custom_provider.test.token
  auto_transform    = true
  csv_content       = <<-CSV
` + csvContent + `CSV
}
`
}
