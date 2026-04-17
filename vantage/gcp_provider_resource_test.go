package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccGcpProviderResource_basic(t *testing.T) {
	t.Skip("Requires real GCP credentials (project_id, billing_account, dataset_name)")

	resourceName := "vantage_gcp_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGcpProviderConfig("my-gcp-project", "000000-111111-222222", "my_billing_dataset"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", "my-gcp-project"),
					resource.TestCheckResourceAttr(resourceName, "billing_account", "000000-111111-222222"),
					resource.TestCheckResourceAttr(resourceName, "dataset_name", "my_billing_dataset"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "token"),
				),
			},
		},
	})
}

func testAccGcpProviderConfig(projectID, billingAccount, datasetName string) string {
	return `
resource "vantage_gcp_provider" "test" {
  project_id      = "` + projectID + `"
  billing_account = "` + billingAccount + `"
  dataset_name    = "` + datasetName + `"
}
`
}
