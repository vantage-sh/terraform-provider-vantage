package vantage

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestAccVantageRecommendationView_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_recommendation_view.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccRecommendationViewConfig_basic(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_token"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			// Update title
			{
				Config: testAccRecommendationViewConfig_basic(rUpdatedTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rUpdatedTitle),
				),
			},
			// Verify no drift
			{
				Config:             testAccRecommendationViewConfig_basic(rUpdatedTitle),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageRecommendationView_withFilters(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_recommendation_view.test_filters"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with provider filter
			{
				Config: testAccRecommendationViewConfig_withProviders(rTitle, []string{"aws"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "provider_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provider_ids.0", "aws"),
				),
			},
			// Update providers filter
			{
				Config: testAccRecommendationViewConfig_withProviders(rTitle, []string{"aws", "gcp"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "provider_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "provider_ids.0", "aws"),
					resource.TestCheckResourceAttr(resourceName, "provider_ids.1", "gcp"),
				),
			},
			// Verify no drift
			{
				Config:             testAccRecommendationViewConfig_withProviders(rTitle, []string{"aws", "gcp"}),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageRecommendationView_withRegions(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_recommendation_view.test_regions"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with regions
			{
				Config: testAccRecommendationViewConfig_withRegions(rTitle, []string{"us-east-1"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", "us-east-1"),
				),
			},
			// Update regions
			{
				Config: testAccRecommendationViewConfig_withRegions(rTitle, []string{"us-east-1", "us-west-2"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "regions.1", "us-west-2"),
				),
			},
		},
	})
}

func TestAccVantageRecommendationView_withDateFilters(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_recommendation_view.test_dates"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with date filters
			{
				Config: testAccRecommendationViewConfig_withDates(rTitle, "2024-01-01", "2024-12-31"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "start_date", "2024-01-01"),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2024-12-31"),
				),
			},
			// Update date filters
			{
				Config: testAccRecommendationViewConfig_withDates(rTitle, "2024-06-01", "2024-12-31"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "start_date", "2024-06-01"),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2024-12-31"),
				),
			},
		},
	})
}

func TestAccVantageRecommendationView_withTags(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_recommendation_view.test_tags"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with tag filters
			{
				Config: testAccRecommendationViewConfig_withTags(rTitle, "environment", "production"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "tag_key", "environment"),
					resource.TestCheckResourceAttr(resourceName, "tag_value", "production"),
				),
			},
			// Update tag filters
			{
				Config: testAccRecommendationViewConfig_withTags(rTitle, "environment", "staging"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tag_key", "environment"),
					resource.TestCheckResourceAttr(resourceName, "tag_value", "staging"),
				),
			},
		},
	})
}

func TestAccVantageRecommendationView_allFilters(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_recommendation_view.test_all"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all filters
			{
				Config: testAccRecommendationViewConfig_allFilters(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "provider_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "start_date", "2024-01-01"),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2024-12-31"),
					resource.TestCheckResourceAttr(resourceName, "tag_key", "environment"),
					resource.TestCheckResourceAttr(resourceName, "tag_value", "production"),
				),
			},
		},
	})
}

func testAccRecommendationViewConfig_basic(title string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_recommendation_view" "test" {
  title           = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
}
`, title)
}

func testAccRecommendationViewConfig_withProviders(title string, providers []string) string {
	providersStr := ""
	for i, p := range providers {
		if i > 0 {
			providersStr += ", "
		}
		providersStr += fmt.Sprintf("%q", p)
	}

	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_recommendation_view" "test_filters" {
  title           = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  provider_ids    = [%[2]s]
}
`, title, providersStr)
}

func testAccRecommendationViewConfig_withRegions(title string, regions []string) string {
	regionsStr := ""
	for i, r := range regions {
		if i > 0 {
			regionsStr += ", "
		}
		regionsStr += fmt.Sprintf("%q", r)
	}

	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_recommendation_view" "test_regions" {
  title           = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  regions         = [%[2]s]
}
`, title, regionsStr)
}

func testAccRecommendationViewConfig_withDates(title, startDate, endDate string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_recommendation_view" "test_dates" {
  title           = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  start_date      = %[2]q
  end_date        = %[3]q
}
`, title, startDate, endDate)
}

func testAccRecommendationViewConfig_withTags(title, tagKey, tagValue string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_recommendation_view" "test_tags" {
  title           = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  tag_key         = %[2]q
  tag_value       = %[3]q
}
`, title, tagKey, tagValue)
}

func testAccRecommendationViewConfig_allFilters(title string) string {
	return fmt.Sprintf(`
data "vantage_workspaces" "test" {}

resource "vantage_recommendation_view" "test_all" {
  title           = %[1]q
  workspace_token = data.vantage_workspaces.test.workspaces[0].token
  provider_ids    = ["aws", "gcp"]
  regions         = ["us-east-1", "us-west-2"]
  start_date      = "2024-01-01"
  end_date        = "2024-12-31"
  tag_key         = "environment"
  tag_value       = "production"
}
`, title)
}
