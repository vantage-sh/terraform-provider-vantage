package vantage

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

// testAccCostsCSV is a minimal FOCUS 1.1-compatible CSV for testing.
const testAccCostsCSV = `BilledCost,BillingCurrency,ChargePeriodStart,ChargePeriodEnd,ChargeCategory,ResourceId,ServiceName
10.00,USD,2024-01-01T00:00:00Z,2024-01-31T23:59:59Z,Usage,my-resource,MyService
20.00,USD,2024-01-01T00:00:00Z,2024-01-31T23:59:59Z,Usage,other-resource,MyService
`

// testAccCostsCSVBadHeaders is a CSV missing the required ChargePeriodStart column.
const testAccCostsCSVBadHeaders = `BilledCost,BillingCurrency,ChargeCategory,ResourceId,ServiceName
10.00,USD,Usage,my-resource,MyService
`

func TestAccCustomProviderCostsUploadResource_basic(t *testing.T) {
	resourceName := "vantage_custom_provider_costs_upload.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the upload and verify computed attributes are populated.
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
			// Confirm no drift on a subsequent plan.
			{
				Config:             testAccCostsUploadConfig(testAccCostsCSV),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccCustomProviderCostsUploadResource_badHeaders verifies that a CSV
// missing required FOCUS columns is rejected by the API with a clear error.
func TestAccCustomProviderCostsUploadResource_badHeaders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCostsUploadConfig(testAccCostsCSVBadHeaders),
				ExpectError: regexp.MustCompile(`(?i)required column`),
			},
		},
	})
}

// TestAccCustomProviderCostsUploadResource_autoTransform verifies that the
// auto_transform attribute is stored in state, sent to the API, and produces
// no state drift. It uses a FOCUS-compatible CSV so the test does not depend
// on the OpenAI-backed transformation path (which requires live credentials
// not available in CI and is already covered by Rails unit tests).
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

// TestAccCustomProviderCostsUploadResource_customFilename verifies that a
// user-supplied filename is passed to the API and reflected back in state.
func TestAccCustomProviderCostsUploadResource_customFilename(t *testing.T) {
	resourceName := "vantage_custom_provider_costs_upload.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCostsUploadFilenameConfig(testAccCostsCSV, "january-2024.csv"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					// The API echoes back the filename used in the multipart upload.
					resource.TestCheckResourceAttr(resourceName, "filename", "january-2024.csv"),
				),
			},
		},
	})
}

func testAccCostsUploadFilenameConfig(csvContent, filename string) string {
	return `
resource "vantage_custom_provider" "test" {
  name = "Test Provider for Costs Upload Filename"
}

resource "vantage_custom_provider_costs_upload" "test" {
  integration_token = vantage_custom_provider.test.token
  filename          = "` + filename + `"
  csv_content       = <<-CSV
` + csvContent + `CSV
}
`
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
