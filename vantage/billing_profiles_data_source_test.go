package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccBillingProfilesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingProfilesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_billing_profiles.test", "billing_profiles.#"),
				),
			},
		},
	})
}

func TestAccBillingProfilesDataSource_withResource(t *testing.T) {
	nickname := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBillingProfilesDataSourceWithResource(nickname),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_billing_profiles.test", "billing_profiles.#"),
					// Check that the list is not empty (contains at least 1 billing profile)
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["data.vantage_billing_profiles.test"]
						if !ok {
							return fmt.Errorf("Not found: data.vantage_billing_profiles.test")
						}
						if rs.Primary.Attributes["billing_profiles.#"] == "0" {
							return fmt.Errorf("Expected at least 1 billing profile, got 0")
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccBillingProfilesDataSourceConfig() string {
	return `
data "vantage_billing_profiles" "test" {}
`
}

func testAccBillingProfilesDataSourceWithResource(nickname string) string {
	return fmt.Sprintf(`
resource "vantage_billing_profile" "test" {
	nickname = %[1]q
}

data "vantage_billing_profiles" "test" {
	depends_on = [vantage_billing_profile.test]
}
`, nickname)
}
