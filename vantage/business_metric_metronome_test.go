package vantage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func metronomeTestValue(ctx context.Context, s schema.Schema, fields map[string]string) tftypes.Value {
	objectType := s.Type().TerraformType(ctx).(tftypes.Object)
	attrs := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, typ := range objectType.AttributeTypes {
		attrs[name] = tftypes.NewValue(typ, nil)
	}
	attrs["title"] = tftypes.NewValue(tftypes.String, "Metronome")
	if fields != nil {
		fieldType := objectType.AttributeTypes["metronome_metric_fields"].(tftypes.Object)
		values := make(map[string]tftypes.Value, len(fieldType.AttributeTypes))
		for name, typ := range fieldType.AttributeTypes {
			values[name] = tftypes.NewValue(typ, nil)
			if value, ok := fields[name]; ok {
				values[name] = tftypes.NewValue(typ, value)
			}
		}
		attrs["metronome_metric_fields"] = tftypes.NewValue(fieldType, values)
	}
	return tftypes.NewValue(objectType, attrs)
}

func metronomeTestWithReports(t *testing.T, value tftypes.Value, filters map[string][]string) tftypes.Value {
	t.Helper()
	objectType := value.Type().(tftypes.Object)
	listType := objectType.AttributeTypes["cost_report_tokens_with_metadata"].(tftypes.List)
	reportType := listType.ElementType.(tftypes.Object)
	report := make(map[string]tftypes.Value, len(reportType.AttributeTypes))
	for name, typ := range reportType.AttributeTypes {
		report[name] = tftypes.NewValue(typ, nil)
	}
	report["cost_report_token"] = tftypes.NewValue(tftypes.String, "rprt_metronome")
	report["unit_scale"] = tftypes.NewValue(tftypes.String, "per_unit")
	report["calculation_type"] = tftypes.NewValue(tftypes.String, "unit_cost")
	filterType := reportType.AttributeTypes["label_filters"].(tftypes.Map)
	filterValues := make(map[string]tftypes.Value, len(filters))
	for key, values := range filters {
		items := make([]tftypes.Value, len(values))
		for i, item := range values {
			items[i] = tftypes.NewValue(tftypes.String, item)
		}
		filterValues[key] = tftypes.NewValue(filterType.ElementType, items)
	}
	report["label_filters"] = tftypes.NewValue(filterType, filterValues)
	var attrs map[string]tftypes.Value
	if err := value.As(&attrs); err != nil {
		t.Fatal(err)
	}
	attrs["cost_report_tokens_with_metadata"] = tftypes.NewValue(listType, []tftypes.Value{tftypes.NewValue(reportType, report)})
	return tftypes.NewValue(objectType, attrs)
}

func TestBusinessMetricMetronomeLifecycle(t *testing.T) {
	for _, kind := range []string{"revenue", "usage"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			fields := map[string]string{"integration_token": "accss_crdntl_metronome", "metric_kind": kind}
			resolved := map[string]interface{}{"metric_kind": kind, "metric": "daily_revenue", "billable_metric_id": nil, "billable_metric_name": nil}
			if kind == "usage" {
				fields["billable_metric_id"] = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
				fields["billable_metric_name"] = "API requests"
				fields["metric"] = "usage"
				resolved["metric"] = "usage"
				resolved["billable_metric_id"] = fields["billable_metric_id"]
				resolved["billable_metric_name"] = fields["billable_metric_name"]
			}
			filters := map[string][]string{"customer_id": {"customer-one", "customer-two"}, "region": {"us-east-1"}}
			requests := []string{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				requests = append(requests, req.Method)
				if req.Method == http.MethodPost || req.Method == http.MethodPut {
					var body map[string]json.RawMessage
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Fatalf("decode request: %v", err)
					}
					var got map[string]string
					if err := json.Unmarshal(body["metronome_metric_fields"], &got); err != nil {
						t.Fatalf("decode Metronome configuration: %v", err)
					}
					if !reflect.DeepEqual(got, fields) {
						t.Errorf("configuration = %#v, want %#v", got, fields)
					}
					var attachments []struct {
						LabelFilters map[string][]string `json:"label_filters"`
					}
					if err := json.Unmarshal(body["cost_report_tokens_with_metadata"], &attachments); err != nil {
						t.Fatalf("decode report filters: %v", err)
					}
					if len(attachments) != 1 || !reflect.DeepEqual(attachments[0].LabelFilters, filters) {
						t.Errorf("report filters = %#v, want %#v", attachments, filters)
					}
					if req.Method == http.MethodPut && req.URL.Path != "/v2/business_metrics/bsnss_mtrc_metronome" {
						t.Errorf("update path = %s", req.URL.Path)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if req.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"token": "bsnss_mtrc_metronome", "title": "Metronome", "import_type": "metronome_metrics",
					"integration_token": fields["integration_token"], "metronome_metric_fields": resolved,
					"cost_report_tokens_with_metadata": []interface{}{map[string]interface{}{
						"cost_report_token": "rprt_metronome", "unit_scale": "per_unit", "calculation_type": "unit_cost", "label_filters": filters,
					}},
				})
			}))
			defer server.Close()
			r := &businessMetricResource{client: clientForServer(t, server.URL)}
			var schemaResp fwresource.SchemaResponse
			r.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
			s := schemaResp.Schema
			create := fwresource.CreateResponse{State: tfsdk.State{Schema: s}}
			r.Create(ctx, fwresource.CreateRequest{Plan: tfsdk.Plan{Raw: metronomeTestWithReports(t, metronomeTestValue(ctx, s, fields), filters), Schema: s}}, &create)
			if create.Diagnostics.HasError() {
				t.Fatalf("create: %v", create.Diagnostics)
			}
			assertMetronomeState(t, ctx, create.State, fields, filters)

			imported := fwresource.ImportStateResponse{State: tfsdk.State{Raw: metronomeTestValue(ctx, s, nil), Schema: s}}
			r.ImportState(ctx, fwresource.ImportStateRequest{ID: "bsnss_mtrc_metronome"}, &imported)
			if imported.Diagnostics.HasError() {
				t.Fatalf("import: %v", imported.Diagnostics)
			}
			read := fwresource.ReadResponse{State: imported.State}
			r.Read(ctx, fwresource.ReadRequest{State: imported.State}, &read)
			if read.Diagnostics.HasError() {
				t.Fatalf("read imported configuration: %v", read.Diagnostics)
			}
			assertMetronomeState(t, ctx, read.State, fields, filters)

			fields["integration_token"] = "accss_crdntl_updated"
			if kind == "usage" {
				fields["billable_metric_id"] = "bbbbbbbb-cccc-dddd-eeee-ffffffffffff"
				fields["billable_metric_name"] = "Output tokens"
				resolved["billable_metric_id"] = fields["billable_metric_id"]
				resolved["billable_metric_name"] = fields["billable_metric_name"]
			}
			plan := tfsdk.Plan{Raw: metronomeTestWithReports(t, metronomeTestValue(ctx, s, fields), filters), Schema: s}
			plan.Raw = metronomeTestWithIDs(t, plan.Raw)
			update := fwresource.UpdateResponse{State: read.State}
			r.Update(ctx, fwresource.UpdateRequest{Plan: plan, State: read.State}, &update)
			if update.Diagnostics.HasError() {
				t.Fatalf("update: %v", update.Diagnostics)
			}
			assertMetronomeState(t, ctx, update.State, fields, filters)
			if !reflect.DeepEqual(requests, []string{http.MethodPost, http.MethodGet, http.MethodPut}) {
				t.Fatalf("requests = %v", requests)
			}
		})
	}
}

func assertMetronomeState(t *testing.T, ctx context.Context, state tfsdk.State, expected map[string]string, filters map[string][]string) {
	t.Helper()
	var model businessMetricResourceModel
	if d := state.Get(ctx, &model); d.HasError() {
		t.Fatalf("read state: %v", d)
	}
	if model.Token.ValueString() != "bsnss_mtrc_metronome" || model.Id != model.Token {
		t.Fatalf("metric identity changed: %s %s", model.Token, model.Id)
	}
	var importedFilters map[string][]string
	if d := state.GetAttribute(ctx, path.Root("cost_report_tokens_with_metadata").AtListIndex(0).AtName("label_filters"), &importedFilters); d.HasError() {
		t.Fatalf("read report filters: %v", d)
	}
	if !reflect.DeepEqual(importedFilters, filters) {
		t.Fatalf("report filters changed: %#v, want %#v", importedFilters, filters)
	}
	fields := model.MetronomeMetricFields
	if fields.IntegrationToken.ValueString() != expected["integration_token"] || fields.MetricKind.ValueString() != expected["metric_kind"] {
		t.Fatalf("unexpected configuration: %#v", fields)
	}
	if expected["metric_kind"] == "usage" && (fields.BillableMetricId.ValueString() != expected["billable_metric_id"] || fields.BillableMetricName.ValueString() != expected["billable_metric_name"]) {
		t.Fatalf("usage selection lost: %#v", fields)
	}
}

func metronomeTestWithIDs(t *testing.T, value tftypes.Value) tftypes.Value {
	t.Helper()
	var attrs map[string]tftypes.Value
	if err := value.As(&attrs); err != nil {
		t.Fatal(err)
	}
	attrs["token"] = tftypes.NewValue(tftypes.String, "bsnss_mtrc_metronome")
	attrs["id"] = attrs["token"]
	return tftypes.NewValue(value.Type(), attrs)
}

func TestBusinessMetricMetronomePlan(t *testing.T) {
	ctx := context.Background()
	var schemaResp fwresource.SchemaResponse
	NewBusinessMetricResource().Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema
	objectType := s.Type().TerraformType(ctx).(tftypes.Object)
	reportListType := objectType.AttributeTypes["cost_report_tokens_with_metadata"].(tftypes.List)
	reportType := reportListType.ElementType.(tftypes.Object)
	reportAttrs := map[string]tftypes.Value{}
	for name, typ := range reportType.AttributeTypes {
		reportAttrs[name] = tftypes.NewValue(typ, nil)
	}
	reportAttrs["cost_report_token"] = tftypes.NewValue(tftypes.String, "rprt_metronome")
	reports := tftypes.NewValue(reportListType, []tftypes.Value{tftypes.NewValue(reportType, reportAttrs)})
	emptyReports := tftypes.NewValue(reportListType, []tftypes.Value{})
	for _, kind := range []string{"revenue", "usage"} {
		for _, clearing := range []bool{false, true} {
			name := kind + "/retain_reports"
			if clearing {
				name = kind + "/clear_reports"
			}
			t.Run(name, func(t *testing.T) {
				fields := map[string]string{"integration_token": "accss_crdntl_metronome", "metric_kind": kind, "metric": "daily_revenue"}
				if kind == "usage" {
					fields["metric"] = "usage"
					fields["billable_metric_id"] = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
					fields["billable_metric_name"] = "API requests"
				}
				prior := metronomeTestWithIDs(t, metronomeTestValue(ctx, s, fields))
				var priorAttrs map[string]tftypes.Value
				_ = prior.As(&priorAttrs)
				priorAttrs["cost_report_tokens_with_metadata"] = reports
				prior = tftypes.NewValue(objectType, priorAttrs)
				fields["integration_token"] = "accss_crdntl_updated"
				config := metronomeTestValue(ctx, s, fields)
				if clearing {
					var attrs map[string]tftypes.Value
					_ = config.As(&attrs)
					attrs["cost_report_tokens_with_metadata"] = emptyReports
					config = tftypes.NewValue(objectType, attrs)
				}
				priorDV, _ := tfprotov6.NewDynamicValue(objectType, prior)
				configDV, _ := tfprotov6.NewDynamicValue(objectType, config)
				server, err := providerserver.NewProtocol6WithError(New())()
				if err != nil {
					t.Fatal(err)
				}
				resp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: "vantage_business_metric", PriorState: &priorDV, ProposedNewState: &configDV, Config: &configDV})
				if err != nil {
					t.Fatal(err)
				}
				for _, d := range resp.Diagnostics {
					if d.Severity == tfprotov6.DiagnosticSeverityError {
						t.Fatalf("plan: %s %s", d.Summary, d.Detail)
					}
				}
				if len(resp.RequiresReplace) != 0 {
					t.Fatalf("configuration edit requires replacement: %v", resp.RequiresReplace)
				}
				planned, err := resp.PlannedState.Unmarshal(objectType)
				if err != nil {
					t.Fatal(err)
				}
				var attrs map[string]tftypes.Value
				if err := planned.As(&attrs); err != nil {
					t.Fatal(err)
				}
				want := reports
				if clearing {
					want = emptyReports
				}
				if !attrs["cost_report_tokens_with_metadata"].Equal(want) {
					t.Fatalf("planned report links = %v, want %v", attrs["cost_report_tokens_with_metadata"], want)
				}
			})
		}
	}
}

func TestBusinessMetricMetronomeRejectedUpdatePreservesState(t *testing.T) {
	ctx := context.Background()
	var schemaResp fwresource.SchemaResponse
	NewBusinessMetricResource().Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema
	prior := tfsdk.State{Raw: metronomeTestWithIDs(t, metronomeTestValue(ctx, s, map[string]string{"integration_token": "accss_crdntl_metronome", "metric_kind": "revenue"})), Schema: s}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests++
		if req.Method != http.MethodPut {
			t.Errorf("unexpected mutation: %s", req.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"errors":["invalid billable metric"]}`))
	}))
	defer server.Close()
	plan := tfsdk.Plan{Raw: metronomeTestWithIDs(t, metronomeTestValue(ctx, s, map[string]string{"metric_kind": "usage", "billable_metric_id": "invalid"})), Schema: s}
	resp := fwresource.UpdateResponse{State: prior}
	(&businessMetricResource{client: clientForServer(t, server.URL)}).Update(ctx, fwresource.UpdateRequest{Plan: plan, State: prior}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected update diagnostic")
	}
	if requests != 1 || !resp.State.Raw.Equal(prior.Raw) {
		t.Fatalf("rejected update changed state or made extra requests: %d", requests)
	}
}

func TestBusinessMetricMetronomeTerraform(t *testing.T) {
	for _, kind := range []string{"revenue", "usage"} {
		t.Run(kind, func(t *testing.T) {
			var mu sync.Mutex
			var stored map[string]interface{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				if req.Method == http.MethodDelete {
					stored = nil
					w.WriteHeader(http.StatusNoContent)
					return
				}
				if req.Method == http.MethodPost || req.Method == http.MethodPut {
					var body map[string]interface{}
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Errorf("decode body: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					fields := body["metronome_metric_fields"].(map[string]interface{})
					if req.Method == http.MethodPost {
						stored = map[string]interface{}{
							"token": "bsnss_mtrc_metronome", "title": "Metronome", "import_type": "metronome_metrics", "integration_token": "accss_crdntl_metronome",
							"metronome_metric_fields":          map[string]interface{}{"metric_kind": kind, "metric": "daily_revenue", "billable_metric_id": nil, "billable_metric_name": nil},
							"cost_report_tokens_with_metadata": []interface{}{},
						}
					}
					config := stored["metronome_metric_fields"].(map[string]interface{})
					for k, v := range fields {
						if k == "integration_token" {
							stored["integration_token"] = v
						} else {
							config[k] = v
						}
					}
					if reports, ok := body["cost_report_tokens_with_metadata"]; ok && reports != nil {
						stored["cost_report_tokens_with_metadata"] = reports
					}
					if req.Method == http.MethodPost {
						w.WriteHeader(http.StatusCreated)
					}
				}
				if stored == nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				_ = json.NewEncoder(w).Encode(stored)
			}))
			defer server.Close()
			config := func(credential string) string {
				selection := ""
				if kind == "usage" {
					selection = `metric = "usage"
billable_metric_id = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
billable_metric_name = "API requests"`
				}
				return fmt.Sprintf(`provider "vantage" {
 host = %q
 api_token = "test"
}
resource "vantage_business_metric" "test" {
 title = "Metronome"
 metronome_metric_fields = {
 integration_token = %q
 metric_kind = %q
 %s
 }
}`, server.URL, credential, kind, selection)
			}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config("accss_crdntl_metronome"), Check: resource.TestCheckResourceAttr("vantage_business_metric.test", "metronome_metric_fields.metric_kind", kind)},
					{Config: config("accss_crdntl_metronome"), PlanOnly: true, ExpectNonEmptyPlan: false},
					{ResourceName: "vantage_business_metric.test", ImportState: true, ImportStateVerify: true},
					{Config: config("accss_crdntl_updated"), Check: resource.TestCheckResourceAttr("vantage_business_metric.test", "metronome_metric_fields.integration_token", "accss_crdntl_updated")},
					{Config: config("accss_crdntl_updated"), PlanOnly: true, ExpectNonEmptyPlan: false},
				},
			})
		})
	}
}
