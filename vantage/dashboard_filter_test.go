package vantage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestDashboardFilterRequests(t *testing.T) {
	for _, value := range []types.String{types.StringNull(), types.StringValue(""), types.StringValue(`vantage.provider = 'aws'`)} {
		t.Run(value.String(), func(t *testing.T) {
			model := &dashboardModel{Filter: value}
			var diags diag.Diagnostics
			create := model.toCreate(context.Background(), &diags)
			update := model.toUpdate(context.Background(), &diags)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if value.IsNull() && create.Filter != nil {
				t.Fatal("create must omit unset filter")
			}
			if !value.IsNull() && (create.Filter == nil || *create.Filter != value.ValueString()) {
				t.Fatal("create lost filter")
			}
			encoded, err := json.Marshal(update)
			if err != nil {
				t.Fatal(err)
			}
			var body map[string]interface{}
			if err := json.Unmarshal(encoded, &body); err != nil {
				t.Fatal(err)
			}
			if body["filter"] != value.ValueString() {
				t.Fatalf("update filter = %#v, want %q", body["filter"], value.ValueString())
			}
		})
	}
}

func TestDashboardFilterState(t *testing.T) {
	normalized := `vantage.provider = 'aws'`
	changed := `vantage.provider = 'gcp'`
	baseline, _ := json.Marshal(normalized)
	configured := types.StringValue(`(vantage.provider = 'aws')`)
	for _, tc := range []struct {
		name     string
		remote   *string
		previous []byte
		want     types.String
	}{
		{"normalized", &normalized, baseline, configured},
		{"remote edit", &changed, baseline, types.StringValue(changed)},
		{"remote clear", nil, baseline, types.StringNull()},
		{"import", &normalized, nil, types.StringValue(normalized)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := dashboardFilterState(configured, tc.remote, tc.previous); !got.Equal(tc.want) {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
	model := &dashboardModel{Filter: configured}
	if diags := model.applyPayload(context.Background(), &modelsv2.Dashboard{}); diags.HasError() {
		t.Fatal(diags)
	}
	if !model.Filter.IsNull() {
		t.Fatal("absent API filter must clear stale state")
	}
	value, diags := dashboardDataSourceValueFromPayload(context.Background(), &modelsv2.Dashboard{Filter: &normalized})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if value.Filter.ValueString() != normalized {
		t.Fatal("data source lost filter")
	}
}

func TestDashboardFilterReadTracksSuccessiveRemoteChanges(t *testing.T) {
	ctx := context.Background()
	original := "vantage.provider = 'aws'"
	edited := "vantage.provider = 'gcp'"
	remote := &original
	payload := &modelsv2.Dashboard{Token: "dshbrd_test", Title: "Test", WorkspaceToken: "wrkspc_test", SavedFilterTokens: []string{}, Widgets: []*modelsv2.DashboardWidget{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/v2/dashboards/dshbrd_test" {
			t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		response := *payload
		response.Filter = remote
		if err := json.NewEncoder(w).Encode(&response); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()
	server, err := providerserver.NewProtocol6WithError(New())()
	if err != nil {
		t.Fatal(err)
	}
	config, err := json.Marshal(map[string]interface{}{"host": srv.URL, "api_token": "test", "timeout": nil})
	if err != nil {
		t.Fatal(err)
	}
	configured, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{Config: &tfprotov6.DynamicValue{JSON: config}})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range configured.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("configure: %s %s", d.Summary, d.Detail)
		}
	}
	var schemaResp resource.SchemaResponse
	DashboardResource{}.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	model := &dashboardModel{}
	payload.Filter = &original
	if d := model.applyPayload(ctx, payload); d.HasError() {
		t.Fatal(d)
	}
	model.Filter = types.StringValue("(" + original + ")")
	state := tfsdk.State{Schema: schemaResp.Schema}
	if d := state.Set(ctx, model); d.HasError() {
		t.Fatal(d)
	}
	typ := schemaResp.Schema.Type().TerraformType(ctx)
	current, err := tfprotov6.NewDynamicValue(typ, state.Raw)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, _ := json.Marshal(original)
	private, _ := json.Marshal(map[string][]byte{"dashboard_filter": snapshot})
	for _, step := range []struct {
		name   string
		filter *string
		want   types.String
	}{
		{"normalization", &original, model.Filter},
		{"external edit", &edited, types.StringValue(edited)},
		{"external revert", &original, types.StringValue(original)},
		{"external clear", nil, types.StringNull()},
		{"external restore", &original, types.StringValue(original)},
	} {
		t.Run(step.name, func(t *testing.T) {
			remote = step.filter
			resp, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: "vantage_dashboard", CurrentState: &current, Private: private})
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range resp.Diagnostics {
				if d.Severity == tfprotov6.DiagnosticSeverityError {
					t.Fatalf("read: %s %s", d.Summary, d.Detail)
				}
			}
			raw, err := resp.NewState.Unmarshal(typ)
			if err != nil {
				t.Fatal(err)
			}
			state.Raw = raw
			if d := state.Get(ctx, model); d.HasError() {
				t.Fatal(d)
			}
			if !model.Filter.Equal(step.want) {
				t.Errorf("filter = %s, want %s", model.Filter, step.want)
			}
			current, private = *resp.NewState, resp.Private
		})
	}
}
