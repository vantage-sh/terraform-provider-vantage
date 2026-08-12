package vantage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

func virtualTagConfigTestSchema(ctx context.Context) resourceschema.Schema {
	var resp frameworkresource.SchemaResponse
	VirtualTagConfigResource{}.Schema(ctx, frameworkresource.SchemaRequest{}, &resp)
	return resp.Schema
}

func virtualTagConfigTestModel(ctx context.Context, key string, preferred bool) *virtualTagConfigModel {
	return &virtualTagConfigModel{
		VirtualTagConfigModel: resource_virtual_tag_config.VirtualTagConfigModel{
			BackfillUntil:    types.StringValue("2026-01-01"),
			CollapsedTagKeys: types.ListNull(resource_virtual_tag_config.CollapsedTagKeysValue{}.Type(ctx)),
			CreatedByToken:   types.StringValue("usr_1"),
			Id:               types.StringValue("vtag_1"),
			Key:              types.StringValue(key),
			Overridable:      types.BoolValue(false),
			Token:            types.StringValue("vtag_1"),
			Values:           types.ListNull(resource_virtual_tag_config.ValuesValue{}.Type(ctx)),
		},
		Preferred: types.BoolValue(preferred),
	}
}

func virtualTagConfigTestState(t *testing.T, ctx context.Context, schema resourceschema.Schema, model *virtualTagConfigModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	if diags := state.Set(ctx, model); diags.HasError() {
		t.Fatalf("setting test state: %v", diags)
	}
	return state
}

func writeVirtualTagConfigResponse(t *testing.T, w http.ResponseWriter, status int, key string) {
	t.Helper()
	createdBy := "usr_1"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(&modelsv2.VirtualTagConfig{
		BackfillUntil:    "2026-01-01",
		CollapsedTagKeys: []*modelsv2.VirtualTagConfigCollapsedTagKey{},
		CreatedByToken:   &createdBy,
		Key:              key,
		Overridable:      false,
		Token:            "vtag_1",
		Values:           []*modelsv2.VirtualTagConfigValue{},
	}); err != nil {
		t.Errorf("encoding virtual tag config response: %v", err)
	}
}

func TestVirtualTagConfigUpdateSyncsNewPreferredWhenPreviousClearFails(t *testing.T) {
	ctx := context.Background()
	var updates []modelsv2.UpdateTag
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodPut && req.URL.Path == "/v2/virtual_tag_configs/vtag_1":
			writeVirtualTagConfigResponse(t, w, http.StatusOK, "new-key")
		case req.Method == http.MethodPut && req.URL.Path == "/v2/tags":
			var update modelsv2.UpdateTag
			if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
				t.Errorf("decoding tag update: %v", err)
			}
			updates = append(updates, update)
			if update.TagKey == "old-key" {
				http.Error(w, "tag update failed", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&modelsv2.Tags{Tags: []*modelsv2.Tag{}})
		default:
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()

	schema := virtualTagConfigTestSchema(ctx)
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, virtualTagConfigTestModel(ctx, "new-key", true)); diags.HasError() {
		t.Fatalf("setting test plan: %v", diags)
	}
	state := virtualTagConfigTestState(t, ctx, schema, virtualTagConfigTestModel(ctx, "old-key", true))
	resp := frameworkresource.UpdateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schema}}

	resource := VirtualTagConfigResource{client: clientForServer(t, srv.URL)}
	resource.Update(ctx, frameworkresource.UpdateRequest{Plan: plan, State: state}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected update diagnostics: %v", resp.Diagnostics)
	}
	if len(updates) != 2 {
		t.Fatalf("got %d tag updates, want 2", len(updates))
	}
	if updates[0].TagKey != "old-key" || updates[0].Preferred == nil || *updates[0].Preferred {
		t.Errorf("previous key update = %#v, want preferred false", updates[0])
	}
	if updates[1].TagKey != "new-key" || updates[1].Preferred == nil || !*updates[1].Preferred {
		t.Errorf("new key update = %#v, want preferred true", updates[1])
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected a warning for the failed previous preferred cleanup")
	}
}

func TestVirtualTagConfigDeleteContinuesWhenPreferredClearFails(t *testing.T) {
	ctx := context.Background()
	deleteCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodPut && req.URL.Path == "/v2/tags":
			http.Error(w, "tag update failed", http.StatusInternalServerError)
		case req.Method == http.MethodDelete && req.URL.Path == "/v2/virtual_tag_configs/vtag_1":
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()

	schema := virtualTagConfigTestSchema(ctx)
	state := virtualTagConfigTestState(t, ctx, schema, virtualTagConfigTestModel(ctx, "key", true))
	resp := frameworkresource.DeleteResponse{State: state}

	resource := VirtualTagConfigResource{client: clientForServer(t, srv.URL)}
	resource.Delete(ctx, frameworkresource.DeleteRequest{State: state}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected delete error: %v", resp.Diagnostics)
	}
	if !deleteCalled {
		t.Fatal("virtual tag config delete was not called")
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected a warning for the failed preferred cleanup")
	}
}

func TestVirtualTagConfigCreatePreservesStateWhenPreferredSyncFails(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/v2/virtual_tag_configs":
			writeVirtualTagConfigResponse(t, w, http.StatusCreated, "key")
		case req.Method == http.MethodPut && req.URL.Path == "/v2/tags":
			http.Error(w, "tag update failed", http.StatusInternalServerError)
		case req.Method == http.MethodGet && req.URL.Path == "/v2/virtual_tag_configs/vtag_1":
			writeVirtualTagConfigResponse(t, w, http.StatusOK, "key")
		case req.Method == http.MethodGet && req.URL.Path == "/v2/tags":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&modelsv2.Tags{Tags: []*modelsv2.Tag{}})
		default:
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()

	schema := virtualTagConfigTestSchema(ctx)
	planModel := virtualTagConfigTestModel(ctx, "key", true)
	planModel.Id = types.StringUnknown()
	planModel.Token = types.StringUnknown()
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, planModel); diags.HasError() {
		t.Fatalf("setting test plan: %v", diags)
	}
	resp := frameworkresource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schema}}

	resource := VirtualTagConfigResource{client: clientForServer(t, srv.URL)}
	resource.Create(ctx, frameworkresource.CreateRequest{Plan: plan}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected preferred sync error")
	}
	var state virtualTagConfigModel
	if diags := resp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("reading response state: %v", diags)
	}
	if state.Token.ValueString() != "vtag_1" {
		t.Fatalf("state token = %q, want vtag_1", state.Token.ValueString())
	}

	readResp := frameworkresource.ReadResponse{State: resp.State}
	resource.Read(ctx, frameworkresource.ReadRequest{State: resp.State}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("unexpected read diagnostics: %v", readResp.Diagnostics)
	}
	if diags := readResp.State.Get(ctx, &state); diags.HasError() {
		t.Fatalf("reading refreshed state: %v", diags)
	}
	if state.Preferred.IsNull() || state.Preferred.IsUnknown() || state.Preferred.ValueBool() {
		t.Fatalf("refreshed preferred = %s, want false", state.Preferred)
	}
}

func TestVirtualTagConfigUpdateTreatsUnsetPreferredAsFalse(t *testing.T) {
	ctx := context.Background()
	var update modelsv2.UpdateTag
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodPut && req.URL.Path == "/v2/virtual_tag_configs/vtag_1":
			writeVirtualTagConfigResponse(t, w, http.StatusOK, "key")
		case req.Method == http.MethodPut && req.URL.Path == "/v2/tags":
			if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
				t.Errorf("decoding tag update: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&modelsv2.Tags{Tags: []*modelsv2.Tag{}})
		default:
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()

	schema := virtualTagConfigTestSchema(ctx)
	planModel := virtualTagConfigTestModel(ctx, "key", true)
	planModel.Preferred = types.BoolNull()
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, planModel); diags.HasError() {
		t.Fatalf("setting test plan: %v", diags)
	}
	state := virtualTagConfigTestState(t, ctx, schema, virtualTagConfigTestModel(ctx, "key", true))
	resp := frameworkresource.UpdateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schema}}

	resource := VirtualTagConfigResource{client: clientForServer(t, srv.URL)}
	resource.Update(ctx, frameworkresource.UpdateRequest{Plan: plan, State: state}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected update diagnostics: %v", resp.Diagnostics)
	}
	if update.TagKey != "key" || update.Preferred == nil || *update.Preferred {
		t.Fatalf("tag update = %#v, want preferred false", update)
	}
}
