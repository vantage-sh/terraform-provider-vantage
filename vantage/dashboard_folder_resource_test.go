package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

// Requires the dashboard_folders account feature. Run with TF_ACC=1 and a
// VANTAGE_API_TOKEN for an account with that feature enabled.
func TestAccVantageDashboard_folderToken(t *testing.T) {
	title := "tf-dashboard-folders-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	inParent := testAccVantageDashboardFolderConfig(title, "vantage_folder.parent.token")
	inNested := testAccVantageDashboardFolderConfig(title, "vantage_folder.nested.token")
	atRoot := testAccVantageDashboardFolderConfig(title, "")
	explicitRoot := testAccVantageDashboardFolderConfig(title, `""`)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: inParent, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("vantage_folder.parent", "type", "DashboardFolder"),
				resource.TestCheckResourceAttr("vantage_folder.nested", "type", "DashboardFolder"),
				resource.TestCheckResourceAttr("vantage_folder.parent", "saved_filter_tokens.#", "0"),
				resource.TestCheckResourceAttrPair("vantage_folder.nested", "parent_folder_token", "vantage_folder.parent", "token"),
				resource.TestCheckResourceAttrPair("vantage_dashboard.test", "folder_token", "vantage_folder.parent", "token"),
				resource.TestCheckResourceAttrPair("vantage_dashboard.test", "workspace_token", "vantage_folder.parent", "workspace_token"),
				resource.TestCheckResourceAttrPair("data.vantage_folder.parent", "token", "vantage_folder.parent", "token"),
				resource.TestCheckResourceAttr("data.vantage_folder.parent", "type", "DashboardFolder"),
				resource.TestCheckResourceAttrSet("data.vantage_folders.dashboards", "folders.0.type"),
				resource.TestCheckResourceAttrPair("data.vantage_dashboards.test", "dashboards.0.folder_token", "vantage_folder.parent", "token"),
			)},
			{Config: inParent, PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: "vantage_dashboard.test", ImportState: true, ImportStateIdFunc: importStateToken("vantage_dashboard.test"), ImportStateVerify: true},
			{Config: inNested, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("vantage_dashboard.test", "folder_token", "vantage_folder.nested", "token"),
				resource.TestCheckResourceAttrPair("data.vantage_dashboards.test", "dashboards.0.folder_token", "vantage_folder.nested", "token"),
			)},
			{Config: inNested, PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: atRoot, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr("vantage_dashboard.test", "folder_token"),
				resource.TestCheckNoResourceAttr("data.vantage_dashboards.test", "dashboards.0.folder_token"),
			)},
			{Config: atRoot, PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: "vantage_dashboard.test", ImportState: true, ImportStateIdFunc: importStateToken("vantage_dashboard.test"), ImportStateVerify: true},
			{Config: explicitRoot, Check: resource.TestCheckResourceAttr("vantage_dashboard.test", "folder_token", "")},
			{Config: explicitRoot, PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

func testAccVantageDashboardFolderConfig(title, folder string) string {
	folderAttribute, folderFilter := "", ""
	if folder != "" {
		folderAttribute = "folder_token = " + folder
		folderFilter = folderAttribute
	}
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_folder" "parent" {
 title = %[1]q
 type = "DashboardFolder"
 workspace_token = data.vantage_workspaces.test.workspaces[0].token
}
resource "vantage_folder" "nested" {
 title = "%[1]s-nested"
 type = "DashboardFolder"
 parent_folder_token = vantage_folder.parent.token
 saved_filter_tokens = []
}
resource "vantage_dashboard" "test" {
 title = %[1]q
 date_interval = "last_month"
 saved_filter_tokens = []
 widgets = []
 %[2]s
}
data "vantage_folder" "parent" {
 title = vantage_folder.parent.title
 type = "DashboardFolder"
 workspace_token = vantage_folder.parent.workspace_token
}
data "vantage_folders" "dashboards" {
 type = "DashboardFolder"
 depends_on = [vantage_folder.parent, vantage_folder.nested]
}
data "vantage_dashboards" "test" {
 q = vantage_dashboard.test.title
 %[3]s
 depends_on = [vantage_dashboard.test]
}
`, title, folderAttribute, folderFilter)
}

func TestAccVantageFolder_dashboardFolder(t *testing.T) {
	title := "tf-dashboard-folder-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	nested := testAccVantageFolderDashboardConfig(title, "vantage_folder.parent.token")
	renamed := testAccVantageFolderDashboardConfig(title+"-renamed", "vantage_folder.parent.token")
	root := testAccVantageFolderDashboardConfig(title+"-updated", `""`)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: nested, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("vantage_folder.test", "type", "DashboardFolder"),
				resource.TestCheckResourceAttr("vantage_folder.test", "saved_filter_tokens.#", "0"),
				resource.TestCheckResourceAttrPair("vantage_folder.test", "parent_folder_token", "vantage_folder.parent", "token"),
			)},
			{Config: nested, PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: "vantage_folder.test", ImportState: true, ImportStateIdFunc: importStateToken("vantage_folder.test"), ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "token"},
			{Config: renamed, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("vantage_folder.test", "title", title+"-renamed"),
				resource.TestCheckResourceAttrPair("vantage_folder.test", "parent_folder_token", "vantage_folder.parent", "token"),
			)},
			{Config: renamed, PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: root, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("vantage_folder.test", "title", title+"-updated"),
				resource.TestCheckResourceAttr("vantage_folder.test", "parent_folder_token", ""),
				resource.TestCheckResourceAttr("vantage_folder.test", "saved_filter_tokens.#", "0"),
			)},
			{Config: root, PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: nested, Check: resource.TestCheckResourceAttrPair("vantage_folder.test", "parent_folder_token", "vantage_folder.parent", "token")},
			{Config: nested, PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

func testAccVantageFolderDashboardConfig(title, parent string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}
resource "vantage_folder" "parent" {
 title = "Dashboard Folder Parent"
 type = "DashboardFolder"
 workspace_token = data.vantage_workspaces.test.workspaces[0].token
}
resource "vantage_folder" "test" {
 title = %[1]q
 type = "DashboardFolder"
 parent_folder_token = %[2]s
 workspace_token = vantage_folder.parent.workspace_token
}
`, title, parent)
}

func importStateToken(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok || resourceState.Primary == nil || resourceState.Primary.Attributes["token"] == "" {
			return "", fmt.Errorf("missing token for %s", resourceName)
		}
		return resourceState.Primary.Attributes["token"], nil
	}
}

// Omitting type must keep an existing folder's type instead of planning a
// replacement as CostFolder (e.g. after importing a non-Cost folder).
func TestAccVantageFolder_omittedTypeKeepsExistingType(t *testing.T) {
	for _, folderType := range []string{"ProviderResourceFolder", "DashboardFolder"} {
		t.Run(folderType, func(t *testing.T) {
			title := "tf-folder-type-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: testAccVantageFolderTypeConfig(title, folderType), Check: resource.TestCheckResourceAttr("vantage_folder.test", "type", folderType)},
					{Config: testAccVantageFolderTypeConfig(title, ""), PlanOnly: true, ExpectNonEmptyPlan: false},
					{Config: testAccVantageFolderTypeConfig(title+"-renamed", ""), Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("vantage_folder.test", "title", title+"-renamed"),
						resource.TestCheckResourceAttr("vantage_folder.test", "type", folderType),
					)},
					{ResourceName: "vantage_folder.test", ImportState: true, ImportStateIdFunc: importStateToken("vantage_folder.test"), ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "token"},
					{Config: testAccVantageFolderTypeConfig(title+"-renamed", ""), PlanOnly: true, ExpectNonEmptyPlan: false},
				},
			})
		})
	}
}

func testAccVantageFolderTypeConfig(title, folderType string) string {
	typeAttribute := ""
	if folderType != "" {
		typeAttribute = fmt.Sprintf("type = %q", folderType)
	}
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}
resource "vantage_folder" "test" {
 title = %[1]q
 %[2]s
 workspace_token = data.vantage_workspaces.test.workspaces[0].token
}
`, title, typeAttribute)
}
