package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGcpProviderResource_basic(t *testing.T) {
	resourceName := "vantage_gcp_provider.demo"
	config := `
resource "vantage_gcp_provider" "demo" {
  project_id = "test-project"
  billing_account = "000000-111111-222222"
  service_account = <<EOF
{ "type": "service_account", "project_id": "test-project" }
EOF
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project_id", "test-project"),
					resource.TestCheckResourceAttr(resourceName, "billing_account", "000000-111111-222222"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}