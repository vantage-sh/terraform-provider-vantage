package vantage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestVirtualTagConfigsDataSourceReadsTagSettings(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/v2/virtual_tag_configs" {
			http.NotFound(w, req)
			return
		}

		createdBy := "usr_1"
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&modelsv2.VirtualTagConfigs{
			VirtualTagConfigs: []*modelsv2.VirtualTagConfig{
				{
					BackfillUntil:    "2026-01-01",
					CollapsedTagKeys: []*modelsv2.VirtualTagConfigCollapsedTagKey{},
					CreatedByToken:   &createdBy,
					Hidden:           true,
					Key:              "key",
					Overridable:      false,
					Preferred:        true,
					Token:            "vtag_1",
					Values:           []*modelsv2.VirtualTagConfigValue{},
				},
			},
		}); err != nil {
			t.Errorf("encoding virtual tag configs response: %v", err)
		}
	}))
	defer srv.Close()

	dataSource := virtualTagConfigsDataSource{client: clientForServer(t, srv.URL)}
	var schemaResp datasource.SchemaResponse
	dataSource.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)

	virtualTagConfigsType, diags := schemaResp.Schema.TypeAtPath(ctx, path.Root("virtual_tag_configs"))
	if diags.HasError() {
		t.Fatalf("getting virtual_tag_configs type: %v", diags)
	}
	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			schemaResp.Schema.Type().TerraformType(ctx),
			map[string]tftypes.Value{
				"virtual_tag_configs": tftypes.NewValue(virtualTagConfigsType.TerraformType(ctx), nil),
			},
		),
	}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}

	dataSource.Read(ctx, datasource.ReadRequest{Config: config}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected read diagnostics: %v", resp.Diagnostics)
	}
	var state virtualTagConfigsDataSourceModel
	if diags := resp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("reading data source state: %v", diags)
	}
	if len(state.VirtualTagConfigs) != 1 {
		t.Fatalf("virtual tag config count = %d, want 1", len(state.VirtualTagConfigs))
	}
	if !state.VirtualTagConfigs[0].Hidden.ValueBool() {
		t.Error("hidden = false, want true")
	}
	if !state.VirtualTagConfigs[0].Preferred.ValueBool() {
		t.Error("preferred = false, want true")
	}
}
