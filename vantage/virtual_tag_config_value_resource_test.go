package vantage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
)

func TestVirtualTagConfigValueLifecycle(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		initial string
		updated string
		checks  map[string]string
		replace bool
	}{
		{
			name: "percentages",
			initial: `
  display_name = "Initial allocation"
  percentages = [{ value = "Engineering", pct = 60 }, { value = "Support", pct = 40 }]
  date_ranges = [{ start_date = "2026-01-01", end_date = null }, { start_date = null, end_date = "2025-12-31" }]
`,
			updated: `
  display_name = "Updated allocation"
  percentages = [{ value = "Engineering", pct = 75 }, { value = "Support", pct = 25 }]
  date_ranges = [{ start_date = "2026-02-01", end_date = null }, { start_date = null, end_date = "2026-01-31" }]
`,
			checks: map[string]string{"display_name": "Updated allocation", "percentages.#": "2", "percentages.0.pct": "75", "percentages.1.pct": "25", "date_ranges.#": "2", "date_ranges.0.start_date": "2026-02-01", "date_ranges.1.end_date": "2026-01-31"},
		},
		{
			name: "business_metric",
			initial: `
  business_metric_token = "bsnss_mtrc_1"
  label_key = "team"
  label_values = ["payments", "support"]
  label_filters = { app = ["consumer", "internal"] }
	label_transforms = [{ type = "split", delimiter = "/", index = 0 }, { type = "format", template = "team-{0}" }]
`,
			updated: `
  business_metric_token = "bsnss_mtrc_2"
  label_key = "department"
  label_values = ["engineering", "operations"]
  label_filters = { app = ["frontend", "backend"] }
	label_transforms = [{ type = "split", delimiter = ":", index = 1 }, { type = "format", template = "department-{0}" }]
`,
			checks: map[string]string{"business_metric_token": "bsnss_mtrc_2", "label_key": "department", "label_values.#": "2", "label_values.0": "engineering", "label_values.1": "operations", "label_filters.app.0": "frontend", "label_filters.app.1": "backend", "label_transforms.#": "2", "label_transforms.0.delimiter": ":", "label_transforms.0.index": "1", "label_transforms.1.template": "department-{0}"},
		},
		{
			name: "cost_metric",
			initial: `
  display_name = "Initial metric"
  cost_metric = { filter = "costs.provider = 'aws'", aggregation = { tag = "environment" } }
`,
			updated: `
  display_name = "Updated metric"
  cost_metric = { filter = "costs.provider = 'gcp'", aggregation = { tag = "team" } }
`,
			checks: map[string]string{"display_name": "Updated metric", "cost_metric.filter": "costs.provider = 'gcp'", "cost_metric.aggregation.tag": "team"},
		},
		{
			name:    "name",
			initial: `name = "Initial"`,
			updated: `name = "Updated"`,
			checks:  map[string]string{"name": "Updated"},
		},
		{
			name: "clear_date_ranges",
			initial: `
  name = "Unbounded"
  date_ranges = [{ start_date = "2026-01-01", end_date = null }]
`,
			updated: `
  name = "Unbounded"
  date_ranges = []
`,
			checks:  map[string]string{"name": "Unbounded", "date_ranges.#": "0"},
			replace: true,
		},
		{
			name: "clear_label_filters",
			initial: `
  business_metric_token = "bsnss_mtrc_1"
  label_key = "team"
  label_filters = { app = ["consumer", "internal"] }
`,
			updated: `
  business_metric_token = "bsnss_mtrc_1"
  label_key = "team"
  label_filters = {}
`,
			checks:  map[string]string{"label_filters.%": "0"},
			replace: true,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var mutex sync.Mutex
			var stored map[string]any
			var creates, updates, deletes, reads int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				mutex.Lock()
				defer mutex.Unlock()
				w.Header().Set("Content-Type", "application/json")
				if req.URL.Path != "/v2/virtual_tag_configs/vtag_1/values" && req.URL.Path != "/v2/virtual_tag_configs/vtag_1/values/vtag_val_1" {
					t.Errorf("unexpected endpoint: %s %s", req.Method, req.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				switch req.Method {
				case http.MethodPost, http.MethodPatch:
					var payload map[string]any
					if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
						t.Errorf("decode request: %v", err)
					}
					if req.Method == http.MethodPost {
						creates++
						stored = map[string]any{"token": "vtag_val_1", "percentages": []any{}, "date_ranges": []any{}, "label_transforms": []any{}}
						w.WriteHeader(http.StatusCreated)
					} else {
						updates++
					}
					for key, value := range payload {
						if value != nil {
							stored[key] = value
						}
					}
				case http.MethodGet:
					reads++
					if stored == nil {
						w.WriteHeader(http.StatusNotFound)
						return
					}
				case http.MethodDelete:
					deletes++
					stored = nil
					w.WriteHeader(http.StatusNoContent)
					return
				default:
					t.Errorf("unexpected method: %s", req.Method)
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				if err := json.NewEncoder(w).Encode(stored); err != nil {
					t.Errorf("encode response: %v", err)
				}
			}))
			defer server.Close()
			config := func(fields, filter string) string {
				return fmt.Sprintf(`
provider "vantage" {
  host = %q
  api_token = "test-token"
}
resource "vantage_virtual_tag_config_value" "test" {
  virtual_tag_config_token = "vtag_1"
  filter = %q
  %s
}
`, server.URL, filter, fields)
			}
			resourceName := "vantage_virtual_tag_config_value.test"
			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resourceName, "token", "vtag_val_1"),
				resource.TestCheckResourceAttr(resourceName, "id", "vtag_val_1"),
				resource.TestCheckResourceAttr(resourceName, "virtual_tag_config_token", "vtag_1"),
				resource.TestCheckResourceAttr(resourceName, "filter", "costs.provider = 'gcp'"),
			}
			for field, value := range scenario.checks {
				checks = append(checks, resource.TestCheckResourceAttr(resourceName, field, value))
			}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config(scenario.initial, "costs.provider = 'aws'")},
					{Config: config(scenario.initial, "costs.provider = 'aws'"), PlanOnly: true, ExpectNonEmptyPlan: false},
					{Config: config(scenario.updated, "costs.provider = 'gcp'"), Check: resource.ComposeTestCheckFunc(checks...)},
					{Config: config(scenario.updated, "costs.provider = 'gcp'"), PlanOnly: true, ExpectNonEmptyPlan: false},
					{ResourceName: resourceName, ImportState: true, ImportStateId: "vtag_1/vtag_val_1", ImportStateVerify: true},
				},
			})
			mutex.Lock()
			defer mutex.Unlock()
			wantCreates, wantUpdates, wantDeletes := 1, 1, 1
			if scenario.replace {
				wantCreates, wantUpdates, wantDeletes = 2, 0, 2
			}
			if creates != wantCreates || updates != wantUpdates || deletes != wantDeletes || reads == 0 {
				t.Fatalf("API calls: create=%d update=%d delete=%d read=%d", creates, updates, deletes, reads)
			}
		})
	}
}

func TestVirtualTagConfigValueNotFound(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/virtual_tag_configs/vtag_1/values/vtag_val_1" {
			t.Errorf("unexpected path: %s", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"errors":["not found"]}`)
	}))
	defer server.Close()
	valueResource := VirtualTagConfigValueResource{client: clientForServer(t, server.URL)}
	var schemaResp frameworkresource.SchemaResponse
	valueResource.Schema(ctx, frameworkresource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
	importResp := frameworkresource.ImportStateResponse{State: state}
	valueResource.ImportState(ctx, frameworkresource.ImportStateRequest{ID: "vtag_1/vtag_val_1"}, &importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("import: %v", importResp.Diagnostics)
	}
	readResp := frameworkresource.ReadResponse{State: importResp.State}
	valueResource.Read(ctx, frameworkresource.ReadRequest{State: importResp.State}, &readResp)
	if readResp.Diagnostics.HasError() || !readResp.State.Raw.IsNull() {
		t.Fatalf("read missing value: %v, state: %v", readResp.Diagnostics, readResp.State.Raw)
	}
	deleteResp := frameworkresource.DeleteResponse{}
	valueResource.Delete(ctx, frameworkresource.DeleteRequest{State: importResp.State}, &deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("delete missing value: %v", deleteResp.Diagnostics)
	}
	for _, id := range []string{"", "vtag_1", "/vtag_val_1", "vtag_1/", "vtag_1/vtag_val_1/extra"} {
		resp := frameworkresource.ImportStateResponse{State: state}
		valueResource.ImportState(ctx, frameworkresource.ImportStateRequest{ID: id}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("import %q should fail", id)
		}
	}
}

func TestAccVantageVirtualTagConfigValue_basic(t *testing.T) {
	key := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_virtual_tag_config_value.test"
	config := func(name string, percentage int) string {
		return fmt.Sprintf(`
resource "vantage_virtual_tag_config" "test" {
  key = %q
	overridable = false
}
resource "vantage_virtual_tag_config_value" "test" {
  virtual_tag_config_token = vantage_virtual_tag_config.test.token
  filter = "costs.provider = 'aws'"
  display_name = %q
  percentages = [
    { value = "Engineering", pct = %d },
    { value = "Support", pct = %d },
  ]
  date_ranges = [
    { start_date = "2026-01-01", end_date = null },
    { start_date = null, end_date = "2025-12-31" },
  ]
}
`, key, name, percentage, 100-percentage)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("Initial", 60), Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "display_name", "Initial"),
				resource.TestCheckResourceAttr(resourceName, "percentages.0.pct", "60"),
				resource.TestCheckResourceAttr(resourceName, "percentages.1.pct", "40"),
				resource.TestCheckResourceAttr(resourceName, "date_ranges.#", "2"),
			)},
			{Config: config("Updated", 75), Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "display_name", "Updated"),
				resource.TestCheckResourceAttr(resourceName, "percentages.0.pct", "75"),
				resource.TestCheckResourceAttr(resourceName, "percentages.1.pct", "25"),
			)},
			{Config: config("Updated", 75), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateIdFunc: func(state *terraform.State) (string, error) {
				attributes := state.RootModule().Resources[resourceName].Primary.Attributes
				return strings.Join([]string{attributes["virtual_tag_config_token"], attributes["token"]}, "/"), nil
			}},
		},
	})
}
