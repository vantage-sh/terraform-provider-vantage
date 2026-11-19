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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestAccVantageCustomCalculation_Lifecycle(t *testing.T) {
	records := map[string]map[string]any{}
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		path := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
		if len(path) < 2 {
			t.Errorf("Unexpected request %s %s", req.Method, req.URL.Path)
			http.NotFound(w, req)
			return
		}
		kind := path[1]
		token := map[string]string{"custom_calculations": "calc_native", "cost_reports": "report_native", "dashboards": "dash_native"}[kind]
		if token == "" {
			t.Errorf("Unexpected request %s %s", req.Method, req.URL.Path)
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		prior := records[kind]
		switch req.Method {
		case http.MethodGet:
			if prior == nil {
				http.NotFound(w, req)
				return
			}
		case http.MethodPost, http.MethodPut:
			var body map[string]any
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Error(err)
				http.Error(w, "bad request", 400)
				return
			}
			if req.Method == http.MethodPost {
				prior = map[string]any{"token": token}
			}
			if kind == "cost_reports" && req.Method == http.MethodPut && body["title"] == "Renamed report" {
				// Match Core's contract: null preserves attachments; an empty array clears them.
				if tokens, exists := body["custom_calculation_tokens"]; !exists || tokens != nil {
					t.Errorf("unmanaged calculation tokens must encode as null, got %#v (present: %t)", tokens, exists)
				}
			}
			for key, value := range body {
				if value != nil {
					prior[key] = value
				}
			}
			switch kind {
			case "custom_calculations":
				if prior["description"] == "" {
					prior["description"] = nil
				}
				if prior["cost_report_tokens"] == nil {
					prior["cost_report_tokens"] = []string{}
				}
				filter, ok := prior["filter"].(map[string]any)
				if !ok {
					t.Error("Missing calculation filter")
					http.Error(w, "missing filter", 400)
					return
				}
				if units, ok := filter["usage_unit"].([]any); ok && len(units) == 1 {
					filter["usage_unit"] = units[0]
				}
			case "cost_reports":
				if prior["custom_calculation_tokens"] == nil {
					prior["custom_calculation_tokens"] = []string{}
				}
			}
			records[kind] = prior
			if req.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			}
		case http.MethodDelete:
			delete(records, kind)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		json.NewEncoder(w).Encode(prior)
	}))
	defer srv.Close()

	config := func(units, description, attachments, widgets string) string {
		return fmt.Sprintf(`
provider "vantage" {
  host = %q
  api_token = "native-test"
}
resource "vantage_custom_calculation" "test" {
  title = "Test calculation"
  %s
  filter = {
    type = "sum"
    filter = "costs.provider = 'aws'"
    %s
  }
}
resource "vantage_cost_report" "test" {
  title = "Test report"
  workspace_token = "wrkspc_test"
  date_interval = "last_month"
  %s
}
`, srv.URL, description, units, attachments) + widgets
	}
	dashboard := `
resource "vantage_dashboard" "test" {
  title = "Test dashboard"
  workspace_token = "wrkspc_test"
  date_interval = "last_month"
  widgets = [{
    title = "Test widget"
    widgetable_token = vantage_cost_report.test.token
    settings = {
      display_type = "kpi"
      kpi_type = "custom_calculation"
      kpi_custom_calculation_token = vantage_custom_calculation.test.token
    }
  }]
}
`
	attached := "custom_calculation_tokens = [vantage_custom_calculation.test.token]"
	first := config(`usage_unit = ["Requests"]`, `description = "Test description"`, attached, dashboard)
	second := config(`usage_unit = ["Requests", "requests"]`, "", attached, dashboard)
	omitted := strings.Replace(config(`usage_unit = ["Requests", "requests"]`, "", "", dashboard), "Test report", "Renamed report", 1)
	cleared := config("", "", "custom_calculation_tokens = []", "")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: first, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("vantage_custom_calculation.test", "filter.usage_unit.0", "Requests"),
				resource.TestCheckResourceAttr("vantage_cost_report.test", "custom_calculation_tokens.#", "1"),
				resource.TestCheckResourceAttr("vantage_dashboard.test", "widgets.0.settings.kpi_custom_calculation_token", "calc_native"),
			)},
			{Config: second, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("vantage_custom_calculation.test", "filter.usage_unit.#", "2"),
				resource.TestCheckNoResourceAttr("vantage_custom_calculation.test", "description"),
			)},
			{Config: omitted, Check: resource.TestCheckResourceAttr("vantage_cost_report.test", "custom_calculation_tokens.#", "1")},
			{Config: cleared, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr("vantage_custom_calculation.test", "filter.usage_unit.#"),
				resource.TestCheckResourceAttr("vantage_cost_report.test", "custom_calculation_tokens.#", "0"),
			)},
			{Config: cleared, PlanOnly: true},
			{ResourceName: "vantage_custom_calculation.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestCustomCalculationDivisionRoundTrip(t *testing.T) {
	ctx := context.Background()
	var existing modelsv2.CustomCalculation
	if err := json.Unmarshal([]byte(`{
		"token":"calc_test","title":"Cost per request","operation":"division","display_as":"ratio","unit_scale":1,
		"cost_report_tokens":["report_test"],
		"filter":{"type":"division","filter_numerator":"costs.provider = 'aws'","filter_denominator":"costs.provider = 'aws'","usage_unit_numerator":null,"usage_unit_denominator":"Requests"}
	}`), &existing); err != nil {
		t.Fatal(err)
	}
	var prior customCalculationResourceModel
	if diags := prior.applyPayload(ctx, &existing); diags.HasError() {
		t.Fatal(diags)
	}
	plan := prior
	plan.Title = types.StringValue("Renamed")
	var diags diag.Diagnostics
	rename := plan.toUpdateModel(ctx, prior, types.SetNull(types.StringType), &diags)
	if rename.Title != "Renamed" || rename.Filter != nil || rename.CostReportTokens != nil || diags.HasError() {
		t.Fatalf("Rename must preserve the definition and unmanaged reports: %#v, %v", rename, diags)
	}

	renameWire, err := json.Marshal(rename)
	if err != nil {
		t.Fatal(err)
	}
	var renameBody map[string]json.RawMessage
	if err := json.Unmarshal(renameWire, &renameBody); err != nil {
		t.Fatal(err)
	}
	if string(renameBody["cost_report_tokens"]) != "null" {
		t.Fatalf("unmanaged report tokens must encode as null: %s", renameWire)
	}

	plan.Filter.UsageUnitDenominator, _ = types.ListValueFrom(ctx, types.StringType, []string{"Requests", "requests"})
	clear := plan.toUpdateModel(ctx, prior, types.SetValueMust(types.StringType, nil), &diags)
	wire, err := json.Marshal(clear)
	if err != nil || diags.HasError() {
		t.Fatalf("Serialize calculation update: %v, %v", err, diags)
	}
	var body struct {
		Filter map[string]json.RawMessage
		Tokens json.RawMessage `json:"cost_report_tokens"`
	}
	if err := json.Unmarshal(wire, &body); err != nil {
		t.Fatal(err)
	}
	if string(body.Tokens) != "[]" || string(body.Filter["usage_unit_denominator"]) != `["Requests","requests"]` || body.Filter["usage_unit_numerator"] != nil {
		t.Fatalf("Unexpected filter or clear request: %s", wire)
	}
}
