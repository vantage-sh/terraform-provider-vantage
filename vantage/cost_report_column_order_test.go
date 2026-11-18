package vantage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_cost_report"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestCostReportColumnOrderLifecycle(t *testing.T) {
	ctx := context.Background()
	for _, order := range [][]string{{"service", "provider", "costAmount"}, {"costAmount", "provider", "service"}, {}, nil} {
		t.Run(stringMustJSON(t, order), func(t *testing.T) {
			var settings struct {
				TableColumnOrder json.RawMessage `json:"table_column_order"`
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var body struct {
					Settings json.RawMessage `json:"settings"`
				}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if err := json.Unmarshal(body.Settings, &settings); err != nil {
					t.Error(err)
				}
				if order == nil {
					if settings.TableColumnOrder != nil {
						t.Errorf("omitted order sent as %s", settings.TableColumnOrder)
					}
				} else if string(settings.TableColumnOrder) != stringMustJSON(t, order) {
					t.Errorf("order = %s; want %s", settings.TableColumnOrder, stringMustJSON(t, order))
				}
				w.Header().Set("Content-Type", "application/json")
				responseSettings := map[string]interface{}{}
				if len(order) > 0 {
					responseSettings["table_column_order"] = order
				}
				json.NewEncoder(w).Encode(map[string]interface{}{"token": "rprt_test", "settings": responseSettings})
			}))
			defer srv.Close()
			client := clientForServer(t, srv.URL)
			model := costReportModel{Token: types.StringValue("rprt_test"), Title: types.StringValue("Column order"), SavedFilterTokens: types.ListNull(types.StringType), BusinessMetricTokensWithMetadata: types.ListNull(resource_cost_report.BusinessMetricTokensWithMetadataType{ObjectType: types.ObjectType{AttrTypes: resource_cost_report.BusinessMetricTokensWithMetadataValue{}.AttributeTypes(ctx)}}), Settings: resource_cost_report.SettingsValue{}}
			var schemaResp resource.SchemaResponse
			r := CostReportResource{client: client}
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			attrs := resource_cost_report.SettingsValue{}.AttributeTypes(ctx)
			settingsAttrs := map[string]attr.Value{
				"aggregate_by": types.StringNull(),
				"amortize":     types.BoolNull(), "complete_period": types.BoolNull(),
				"include_credits": types.BoolNull(), "include_discounts": types.BoolNull(),
				"include_refunds": types.BoolNull(), "include_tax": types.BoolNull(),
				"show_previous_period": types.BoolNull(), "unallocated": types.BoolNull(),
			}
			orderValue, d := types.ListValueFrom(ctx, types.StringType, order)
			if d.HasError() {
				t.Fatal(d)
			}
			settingsAttrs["table_column_order"] = orderValue
			model.Settings, d = resource_cost_report.NewSettingsValue(attrs, settingsAttrs)
			if d.HasError() {
				t.Fatal(d)
			}
			var diags diag.Diagnostics
			create := model.toCreateModel(ctx, &diags)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if !reflect.DeepEqual(create.Settings.TableColumnOrder, order) {
				t.Fatalf("create order = %#v; want %#v", create.Settings.TableColumnOrder, order)
			}
			state := tfsdk.State{Schema: schemaResp.Schema}
			if d := state.Set(ctx, &model); d.HasError() {
				t.Fatal(d)
			}
			resp := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: tfsdk.Plan{Schema: schemaResp.Schema, Raw: state.Raw}}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatal(resp.Diagnostics)
			}
			var updated costReportModel
			if d := resp.State.Get(ctx, &updated); d.HasError() {
				t.Fatal(d)
			}
			if !updated.Settings.TableColumnOrder.Equal(orderValue) {
				t.Errorf("state order = %v; want %v", updated.Settings.TableColumnOrder, orderValue)
			}
		})
	}
}

func stringMustJSON(t *testing.T, value interface{}) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestCostReportsDataSourceColumnOrder(t *testing.T) {
	ctx := context.Background()
	order := []string{"service", "provider", "costAmount"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"cost_reports": []*modelsv2.CostReport{
			{Token: "rprt_ordered", Settings: &modelsv2.CostReportSettings{TableColumnOrder: order}},
			{Token: "rprt_default"},
		}})
	}))
	defer srv.Close()
	d := costReportsDataSource{client: clientForServer(t, srv.URL)}
	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	objectType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
		"cost_reports": tftypes.NewValue(objectType.AttributeTypes["cost_reports"], nil),
	})}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, datasource.ReadRequest{Config: config}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	var state costReportsDataSourceModel
	if diags := resp.State.Get(ctx, &state); diags.HasError() {
		t.Fatal(diags)
	}
	if len(state.CostReports) != 2 {
		t.Fatalf("reports = %d; want 2", len(state.CostReports))
	}
	for i, expected := range [][]string{order, {}} {
		got := state.CostReports[i].Settings.Attributes()["table_column_order"]
		want, diags := types.ListValueFrom(ctx, types.StringType, expected)
		if diags.HasError() {
			t.Fatal(diags)
		}
		if !got.Equal(want) {
			t.Errorf("report %d order = %v; want %v", i, got, want)
		}
	}
}
