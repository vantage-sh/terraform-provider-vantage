package vantage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func testAccScenarioModelsPreCheck(t *testing.T) {
	t.Helper()
	acctest.PreCheck(t)

	req, err := http.NewRequest(http.MethodGet, "https://api.vantage.sh/v2/scenario_models", nil)
	if err != nil {
		t.Fatalf("unable to build scenario models pre-check request: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("VANTAGE_API_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unable to probe scenario models API: %s", err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusForbidden {
		t.Skipf("scenario model acceptance tests require a User API token: %s", strings.TrimSpace(string(payload)))
	}
}

func TestAccVantageScenarioModel_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	rUpdatedTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_scenario_model.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccScenarioModelsPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageScenarioModelConfig_basic(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rTitle),
					resource.TestCheckResourceAttr(resourceName, "periods.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.amount", "1250.75"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.amount_type", "dollar"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.start_at", "2026-01-01"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.end_at", "2026-03-31"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVantageScenarioModelConfig_updated(rUpdatedTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", rUpdatedTitle),
					resource.TestCheckResourceAttr(resourceName, "periods.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.amount", "100"),
					resource.TestCheckResourceAttr(resourceName, "periods.0.amount_type", "percent"),
					resource.TestCheckResourceAttr(resourceName, "periods.1.amount", "2500"),
					resource.TestCheckResourceAttr(resourceName, "periods.1.amount_type", "dollar"),
				),
			},
			{
				Config:             testAccVantageScenarioModelConfig_updated(rUpdatedTitle),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccVantageScenarioModelsDataSource_basic(t *testing.T) {
	rTitle := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccScenarioModelsPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVantageScenarioModelsDataSourceConfig(rTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vantage_scenario_models.test", "scenario_models.#"),
				),
			},
		},
	})
}

func testAccVantageScenarioModelConfig_basic(title string) string {
	return fmt.Sprintf(`
resource "vantage_scenario_model" "test" {
  title = %q

  periods = [{
    start_at    = "2026-01-01"
    end_at      = "2026-03-31"
    amount      = 1250.75
    amount_type = "dollar"
  }]
}
`, title)
}

func testAccVantageScenarioModelConfig_updated(title string) string {
	return fmt.Sprintf(`
resource "vantage_scenario_model" "test" {
  title = %q

  periods = [
    {
      start_at    = "2026-01-01"
      end_at      = "2026-03-31"
      amount      = 100
      amount_type = "percent"
    },
    {
      start_at    = "2026-04-01"
      amount      = 2500
      amount_type = "dollar"
    }
  ]
}
`, title)
}

func testAccVantageScenarioModelsDataSourceConfig(title string) string {
	return fmt.Sprintf(`
%s

data "vantage_scenario_models" "test" {
  depends_on = [vantage_scenario_model.test]
}
`, testAccVantageScenarioModelConfig_basic(title))
}
