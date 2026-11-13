package vantage

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccIngestToken_basic(t *testing.T) {
	config := func(name string, days int) string {
		return fmt.Sprintf(`resource "vantage_ingest_token" "test" {
 # Bifrost feed seeded by Core test:access_tokens:create.
 feed_token = "srvc_data_intgrtn_terraform_ingest_test"
 name = %q
 retention_days = %d
}`, name, days)
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) }, ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("terraform-test", 30), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vantage_ingest_token.test", "name", "terraform-test"), resource.TestCheckResourceAttrSet("vantage_ingest_token.test", "plaintext_token"), resource.TestCheckResourceAttr("vantage_ingest_token.test", "telemetry_provider", "bifrost"))},
			{Config: config("terraform-test-updated", 60), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vantage_ingest_token.test", "name", "terraform-test-updated"), resource.TestCheckResourceAttr("vantage_ingest_token.test", "retention_days", "60"))},
			{Config: config("terraform-test-updated", 60), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
