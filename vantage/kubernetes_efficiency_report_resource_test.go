package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestKubernetesReport(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create resource report
				Config: testAccKubernetesReport(rTitle, "kubernetes.cluster_id = 'foo'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_kubernetes_efficiency_report.kubernetes_efficiency_report", "title", rTitle),
					resource.TestCheckResourceAttr("vantage_kubernetes_efficiency_report.kubernetes_efficiency_report", "date_bucket", "week"),
				),
			},
			{
				// update resource report
				Config: testAccKubernetesReport(rUpdatedTitle, "kubernetes.cluster_id = 'bar'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_kubernetes_efficiency_report.kubernetes_efficiency_report", "title", rUpdatedTitle),
					resource.TestCheckResourceAttr("vantage_kubernetes_efficiency_report.kubernetes_efficiency_report", "filter", "kubernetes.cluster_id = 'bar'"),
				),
			},
			{
				// update resource report to use date interval
				Config: testAccKubernetesReportDateInterval(rUpdatedTitle, "kubernetes.cluster_id = 'bar'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vantage_kubernetes_efficiency_report.kubernetes_efficiency_report", "title", rUpdatedTitle),
					resource.TestCheckResourceAttr("vantage_kubernetes_efficiency_report.kubernetes_efficiency_report", "filter", "kubernetes.cluster_id = 'bar'"),
				),
			},
		},
	})
}

func testAccKubernetesReport(title, filter string) string {
	return fmt.Sprintf(`

data "vantage_workspaces" "test" {}

resource "vantage_kubernetes_efficiency_report" "kubernetes_efficiency_report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title = %[1]q
	filter = %[2]q
	aggregated_by = "amount"
	date_bucket = "week"
	date_interval = "custom"
	start_date = "2024-01-01"
	end_date = "2024-01-31"
	groupings = ["namespace","label:app"]
}
`, title, filter)
}

func testAccKubernetesReportDateInterval(title, filter string) string {
	return fmt.Sprintf(`

data "vantage_workspaces" "test" {}

resource "vantage_kubernetes_efficiency_report" "kubernetes_efficiency_report" {
	workspace_token = data.vantage_workspaces.test.workspaces[0].token
  title = %[1]q
	filter = %[2]q
	aggregated_by = "amount"
	date_bucket = "week"
	date_interval = "last_7_days"
	groupings = ["namespace","label:app"]
}
`, title, filter)
}
