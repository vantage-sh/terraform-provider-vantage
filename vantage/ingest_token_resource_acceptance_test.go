package vantage

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccIngestToken_basic(t *testing.T) {
	feed := os.Getenv("VANTAGE_TEST_TELEMETRY_FEED_TOKEN")
	if feed == "" {
		t.Skip("VANTAGE_TEST_TELEMETRY_FEED_TOKEN is required")
	}
	config := func(name string, days int) string {
		return fmt.Sprintf(`resource "vantage_ingest_token" "test" {
 feed_token = %q
 name = %q
 retention_days = %d
}`, feed, name, days)
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) }, ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("terraform-test", 30), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vantage_ingest_token.test", "name", "terraform-test"), resource.TestCheckResourceAttrSet("vantage_ingest_token.test", "plaintext_token"))},
			{Config: config("terraform-test-updated", 60), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vantage_ingest_token.test", "name", "terraform-test-updated"), resource.TestCheckResourceAttr("vantage_ingest_token.test", "retention_days", "60"))},
			{Config: config("terraform-test-updated", 60), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
