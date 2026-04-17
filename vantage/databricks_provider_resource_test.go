package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccDatabricksProviderResource_basic(t *testing.T) {
	t.Skip("not yet implemented")
	resourceName := "vantage_databricks_provider.demo"
	config := `
resource "vantage_databricks_provider" "demo" {
  host  = "https://mycompany.cloud.databricks.com"
  token = "databricks-token"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "host", "https://mycompany.cloud.databricks.com"),
					resource.TestCheckResourceAttr(resourceName, "token", "databricks-token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}