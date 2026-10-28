package vantage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func TestFolderResource_dashboardFolderLifecycle(t *testing.T) {
	ctx := context.Background()
	for _, folderType := range []string{"DashboardFolder", "CostFolder"} {
		t.Run(folderType, func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				calls++
				if (req.Method != http.MethodPut && req.Method != http.MethodGet) || req.URL.Path != "/v2/folders/fldr_test" {
					t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
				}
				if req.Method == http.MethodPut {
					var payload map[string]interface{}
					if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					if parent, ok := payload["parent_folder_token"]; !ok || parent != "" {
						t.Errorf("parent_folder_token = %#v, present %v; want explicit empty string", parent, ok)
					}
					filters, ok := payload["saved_filter_tokens"].([]interface{})
					if !ok || len(filters) != 0 {
						t.Errorf("saved_filter_tokens = %#v, want empty JSON array", payload["saved_filter_tokens"])
					}
				}
				title := "Updated folder"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(&modelsv2.Folder{Token: "fldr_test", Title: &title, Type: folderType, SavedFilterTokens: []string{}, WorkspaceToken: "wrkspc_test"})
			}))
			defer srv.Close()
			folder := FolderResource{client: clientForServer(t, srv.URL)}
			var schemaResp resource.SchemaResponse
			folder.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			model := FolderResourceModel{Token: types.StringValue("fldr_test"), Title: types.StringValue("Updated folder"), Type: types.StringValue(folderType), ParentFolderToken: types.StringValue(""), WorkspaceToken: types.StringValue("wrkspc_test"), SavedFilterTokens: types.ListValueMust(types.StringType, nil)}
			state := tfsdk.State{Schema: schemaResp.Schema}
			if d := state.Set(ctx, &model); d.HasError() {
				t.Fatal(d)
			}
			plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: state.Raw}
			resp := resource.UpdateResponse{State: state}
			folder.Update(ctx, resource.UpdateRequest{Plan: plan}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatal(resp.Diagnostics)
			}
			var updated FolderResourceModel
			if d := resp.State.Get(ctx, &updated); d.HasError() {
				t.Fatal(d)
			}
			if !updated.ParentFolderToken.Equal(types.StringValue("")) {
				t.Errorf("parent_folder_token = %v, want explicit empty string", updated.ParentFolderToken)
			}
			if updated.SavedFilterTokens.IsNull() || len(updated.SavedFilterTokens.Elements()) != 0 {
				t.Errorf("saved_filter_tokens = %v, want empty list", updated.SavedFilterTokens)
			}
			if updated.Type.ValueString() != folderType {
				t.Errorf("type = %v", updated.Type)
			}
			readResp := resource.ReadResponse{State: resp.State}
			folder.Read(ctx, resource.ReadRequest{State: resp.State}, &readResp)
			if readResp.Diagnostics.HasError() {
				t.Fatal(readResp.Diagnostics)
			}
			if d := readResp.State.Get(ctx, &updated); d.HasError() {
				t.Fatal(d)
			}
			if !updated.ParentFolderToken.Equal(types.StringValue("")) {
				t.Errorf("refreshed parent_folder_token = %v", updated.ParentFolderToken)
			}
			if calls != 2 {
				t.Errorf("requests = %d, want 2", calls)
			}
		})
	}
}

func TestFolderResource_typeStateAndReplacement(t *testing.T) {
	ctx := context.Background()
	var resp resource.SchemaResponse
	FolderResource{}.Schema(ctx, resource.SchemaRequest{}, &resp)
	attribute := resp.Schema.Attributes["type"].(schema.StringAttribute)
	// No schema default: a default would be planned for an omitted type and
	// force replacement of existing non-Cost folders.
	if !attribute.Optional || !attribute.Computed || attribute.Default != nil {
		t.Fatalf("type must be optional+computed without a default: %#v", attribute)
	}
	request := planmodifier.StringRequest{ConfigValue: types.StringValue("DashboardFolder"), PlanValue: types.StringValue("DashboardFolder"), StateValue: types.StringValue("CostFolder")}
	// RequiresReplace only runs for an existing resource and a non-null plan.
	objectType := resp.Schema.Type().TerraformType(ctx)
	request.State = tfsdk.State{Schema: resp.Schema, Raw: tftypes.NewValue(objectType, tftypes.UnknownValue)}
	request.Plan = tfsdk.Plan{Schema: resp.Schema, Raw: tftypes.NewValue(objectType, tftypes.UnknownValue)}
	var modifierResp planmodifier.StringResponse
	for _, modifier := range attribute.PlanModifiers {
		modifier.PlanModifyString(ctx, request, &modifierResp)
	}
	if !modifierResp.RequiresReplace {
		t.Fatal("changing folder type must require replacement")
	}
}

func TestFolderResource_savedFiltersValidation(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		folderType types.String
		filters    []attr.Value
		wantError  bool
	}{
		{types.StringValue("DashboardFolder"), []attr.Value{types.StringValue("svd_fltr_a"), types.StringValue("svd_fltr_b")}, true},
		{types.StringValue("ProviderResourceFolder"), []attr.Value{types.StringValue("svd_fltr_a")}, true},
		{types.StringValue("CostFolder"), []attr.Value{types.StringValue("svd_fltr_a")}, false},
		{types.StringValue("DashboardFolder"), []attr.Value{}, false},
		{types.StringUnknown(), []attr.Value{types.StringValue("svd_fltr_a")}, false},
	} {
		var schemaResp resource.SchemaResponse
		FolderResource{}.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		model := FolderResourceModel{Type: tc.folderType, SavedFilterTokens: types.ListValueMust(types.StringType, tc.filters)}
		if d := state.Set(ctx, &model); d.HasError() {
			t.Fatal(d)
		}
		var resp resource.ValidateConfigResponse
		FolderResource{}.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: state.Raw}}, &resp)
		if resp.Diagnostics.HasError() != tc.wantError {
			t.Errorf("type=%v filters=%v: diagnostics=%v, wantError=%v", tc.folderType, tc.filters, resp.Diagnostics, tc.wantError)
		}
	}
}

func TestFolderResource_omittedParentPreservesState(t *testing.T) {
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	FolderResource{}.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	attribute := schemaResp.Schema.Attributes["parent_folder_token"].(schema.StringAttribute)
	request := planmodifier.StringRequest{ConfigValue: types.StringNull(), PlanValue: types.StringUnknown(), StateValue: types.StringValue("fldr_parent")}
	objectType := schemaResp.Schema.Type().TerraformType(ctx)
	request.State = tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objectType, tftypes.UnknownValue)}
	var resp planmodifier.StringResponse
	for _, modifier := range attribute.PlanModifiers {
		modifier.PlanModifyString(ctx, request, &resp)
	}
	if !resp.PlanValue.Equal(request.StateValue) {
		t.Fatalf("omitted parent plan = %v, want existing parent %v", resp.PlanValue, request.StateValue)
	}
}
