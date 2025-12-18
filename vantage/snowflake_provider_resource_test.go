package vantage

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSnowflakeProviderResource_basic(t *testing.T) {
	resourceName := "vantage_snowflake_provider.demo"
	config := `
resource "vantage_snowflake_provider" "demo" {
  account_name = "my_account"
  user_name = "my_user"
  password = "supersecret"
  role = "analyst"
}
`
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_name", "my_account"),
					resource.TestCheckResourceAttr(resourceName, "user_name", "my_user"),
					resource.TestCheckResourceAttr(resourceName, "password", "supersecret"),
					resource.TestCheckResourceAttr(resourceName, "role", "analyst"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}