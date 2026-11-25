package vantage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVirtualTagConfigValuesDataSource(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		payload string
		checks  map[string]string
		empty   bool
	}{
		{
			name: "populated",
			payload: `{
  "token":"vtag_val_1", "filter":"costs.provider = 'aws'", "name":"Engineering",
  "business_metric_token":"bsnss_mtrc_1", "label_key":"team", "label_values":["payments","support"],
  "label_filters":{"app":["consumer","internal"]}, "display_name":"Allocation",
  "cost_metric":{"filter":"costs.provider = 'gcp'", "aggregation":{"tag":"environment"}},
  "percentages":[{"value":"Engineering","pct":62.5},{"value":"Support","pct":37.5}],
  "date_ranges":[{"start_date":"2026-01-01","end_date":null},{"start_date":null,"end_date":"2025-12-31"}],
  "label_transforms":[{"type":"split","delimiter":"/","index":0},{"type":"format","template":"team-{0}"}]
}`,
			checks: map[string]string{
				"token": "vtag_val_1", "virtual_tag_config_token": "vtag_1", "filter": "costs.provider = 'aws'", "name": "Engineering",
				"business_metric_token": "bsnss_mtrc_1", "label_key": "team", "label_values.#": "2", "label_values.0": "payments", "label_values.1": "support",
				"label_filters.app.0": "consumer", "label_filters.app.1": "internal", "display_name": "Allocation",
				"cost_metric.filter": "costs.provider = 'gcp'", "cost_metric.aggregation.tag": "environment",
				"percentages.#": "2", "percentages.0.pct": "62.5", "percentages.1.pct": "37.5", "percentages.0.value": "Engineering", "percentages.1.value": "Support",
				"date_ranges.#": "2", "date_ranges.0.start_date": "2026-01-01", "date_ranges.1.end_date": "2025-12-31",
				"label_transforms.#": "2", "label_transforms.0.index": "0", "label_transforms.0.delimiter": "/", "label_transforms.1.template": "team-{0}",
			},
		},
		{
			name:    "nullable_value",
			payload: `{"token":"vtag_val_1","filter":null,"name":null,"date_ranges":[],"label_transforms":[],"percentages":[]}`,
			checks:  map[string]string{"token": "vtag_val_1", "date_ranges.#": "0", "label_transforms.#": "0", "percentages.#": "0"},
		},
		{name: "empty", empty: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet || req.URL.Path != "/v2/virtual_tag_configs/vtag_1" {
					t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
					http.NotFound(w, req)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				values := scenario.payload
				if scenario.name == "populated" {
					values += `,{"token":"vtag_val_2","name":"Support","filter":"costs.provider = 'gcp'","date_ranges":[],"label_transforms":[],"percentages":[]}`
				}
				fmt.Fprintf(w, `{"token":"vtag_1","values":[%s]}`, values)
			}))
			defer server.Close()
			config := fmt.Sprintf(`
provider "vantage" {
  host = %q
  api_token = "test-token"
}
data "vantage_virtual_tag_config_values" "test" {
  virtual_tag_config_token = "vtag_1"
}
output "engineering_token" {
	value = try(one([for value in data.vantage_virtual_tag_config_values.test.values : value if value.name == "Engineering"]).token, null)
}
`, server.URL)
			dataSourceName := "data.vantage_virtual_tag_config_values.test"
			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(dataSourceName, "token", "vtag_1"),
				resource.TestCheckResourceAttr(dataSourceName, "virtual_tag_config_token", "vtag_1"),
			}
			for field, value := range scenario.checks {
				if field != "virtual_tag_config_token" {
					checks = append(checks, resource.TestCheckResourceAttr(dataSourceName, "values.0."+field, value))
				}
			}
			count := "1"
			if scenario.empty {
				count = "0"
			} else if scenario.name == "populated" {
				count = "2"
				checks = append(checks,
					resource.TestCheckResourceAttr(dataSourceName, "values.1.token", "vtag_val_2"),
					resource.TestCheckResourceAttr(dataSourceName, "values.1.name", "Support"),
					resource.TestCheckOutput("engineering_token", "vtag_val_1"),
				)
			}
			checks = append(checks, resource.TestCheckResourceAttr(dataSourceName, "values.#", count))
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config, Check: resource.ComposeTestCheckFunc(checks...)},
					{Config: config, PlanOnly: true, ExpectNonEmptyPlan: false},
				},
			})
		})
	}
}

func TestVirtualTagConfigValuesDataSourceErrors(t *testing.T) {
	ctx := context.Background()
	for _, code := range []int{http.StatusNotFound, http.StatusForbidden} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(code)
				fmt.Fprint(w, `{"errors":["unavailable"]}`)
			}))
			defer server.Close()
			dataSource := virtualTagConfigValuesDataSource{client: clientForServer(t, server.URL)}
			var schemaResp datasource.SchemaResponse
			dataSource.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
			state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
			if diags := state.SetAttribute(ctx, path.Root("virtual_tag_config_token"), "vtag_1"); diags.HasError() {
				t.Fatal(diags)
			}
			resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			dataSource.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: state.Raw}}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected API error diagnostics")
			}
		})
	}
}
