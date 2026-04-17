package vantage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccSnowflakeProviderResource_basic(t *testing.T) {
	t.Skip("Snowflake integration is not yet supported by the vantage-go SDK")

	resourceName := "vantage_snowflake_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSnowflakeProviderConfig("my_account", "my_user", "supersecret", "analyst"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_name", "my_account"),
					resource.TestCheckResourceAttr(resourceName, "user_name", "my_user"),
					resource.TestCheckResourceAttr(resourceName, "role", "analyst"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccSnowflakeProviderConfig(accountName, userName, password, role string) string {
	return `
resource "vantage_snowflake_provider" "test" {
  account_name = "` + accountName + `"
  user_name    = "` + userName + `"
  password     = "` + password + `"
  role         = "` + role + `"
}
`
}
