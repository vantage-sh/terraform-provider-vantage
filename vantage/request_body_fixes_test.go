package vantage

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/resource_virtual_tag_config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	dashboardsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/dashboards"
	foldersv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/folders"
	ingestv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/ingest_tokens"
	tagsv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/virtual_tags"
)

func TestRequestBodyFixes(t *testing.T) {
	var body map[string]json.RawMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	c := clientForServer(t, server.URL)
	feed := "feed"
	exp := strfmt.DateTime(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
	for _, expiry := range []strfmt.DateTime{{}, exp} {
		body = nil
		_, _ = c.V2.IngestTokens.CreateIngestToken(ingestv2.NewCreateIngestTokenParams().WithCreateIngestToken(&modelsv2.CreateIngestToken{FeedToken: &feed, ExpiresAt: expiry}), c.Auth)
		_, present := body["expires_at"]
		if present != !time.Time(expiry).IsZero() {
			t.Fatalf("expiry %v: %s", expiry, body)
		}
	}
	for _, keys := range [][]*modelsv2.UpdateVirtualTagConfigCollapsedTagKeysItems0{nil, {}} {
		body = nil
		_, _, _ = c.V2.VirtualTags.UpdateVirtualTagConfig(tagsv2.NewUpdateVirtualTagConfigParams().WithToken("tok").WithUpdateVirtualTagConfig(&modelsv2.UpdateVirtualTagConfig{CollapsedTagKeys: keys, Key: "key"}), c.Auth)
		value := body["collapsed_tag_keys"]
		want := "null"
		if keys != nil {
			want = "[]"
		}
		if string(value) != want || string(body["key"]) != `"key"` {
			t.Fatalf("virtual tag body: %s", body)
		}
	}
	for _, values := range [][]*modelsv2.UpdateVirtualTagConfigValuesItems0{nil, {}} {
		body = nil
		_, _, _ = c.V2.VirtualTags.UpdateVirtualTagConfig(tagsv2.NewUpdateVirtualTagConfigParams().WithToken("tok").WithUpdateVirtualTagConfig(&modelsv2.UpdateVirtualTagConfig{Values: values, Key: "key"}), c.Auth)
		want := "null"
		if values != nil {
			want = "[]"
		}
		if string(body["values"]) != want || string(body["key"]) != `"key"` {
			t.Fatalf("virtual tag values: %s", body)
		}
	}
	docType := "doc"
	for _, content := range [][]interface{}{nil, {}} {
		body = nil
		dashboard := &modelsv2.UpdateDashboard{Title: "title", Widgets: []*modelsv2.UpdateDashboardWidgetsItems0{{WidgetableType: "free_text", Content: &modelsv2.UpdateDashboardWidgetsItems0Content{Type: &docType, Content: content}}}}
		_, _ = c.V2.Dashboards.UpdateDashboard(dashboardsv2.NewUpdateDashboardParams().WithDashboardToken("tok").WithUpdateDashboard(dashboard), c.Auth)
		var widgets []struct {
			Content map[string]json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal(body["widgets"], &widgets); err != nil || len(widgets) != 1 {
			t.Fatalf("dashboard widgets %s: %v", body, err)
		}
		want := "null"
		if content != nil {
			want = "[]"
		}
		if string(widgets[0].Content["content"]) != want || string(widgets[0].Content["type"]) != `"doc"` || string(body["title"]) != `"title"` {
			t.Fatalf("dashboard content: %s", body)
		}
	}
	body = nil
	_, _ = c.V2.Folders.UpdateFolder(foldersv2.NewUpdateFolderParams().WithFolderToken("tok").WithUpdateFolder(&modelsv2.UpdateFolder{SavedFilterTokens: []string{}, Title: "title"}), c.Auth)
	if string(body["saved_filter_tokens"]) != "[]" || string(body["title"]) != `"title"` {
		t.Fatalf("folder body: %s", body)
	}
}

func TestFolderUpdateParentSerialization(t *testing.T) {
	ctx := context.Background()
	var body map[string]json.RawMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	r := FolderResource{client: clientForServer(t, server.URL)}
	var sr resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sr)
	for _, tc := range []struct {
		parent  types.String
		want    string
		present bool
	}{
		{types.StringUnknown(), "", false}, {types.StringNull(), `""`, true}, {types.StringValue("parent"), `"parent"`, true},
	} {
		model := FolderResourceModel{Token: types.StringValue("tok"), Title: types.StringValue("new"), ParentFolderToken: tc.parent, Type: types.StringValue("CostFolder"), WorkspaceToken: types.StringValue("ws"), SavedFilterTokens: types.ListNull(types.StringType)}
		plan := tfsdk.Plan{Schema: sr.Schema}
		if d := plan.Set(ctx, &model); d.HasError() {
			t.Fatal(d)
		}
		body = nil
		resp := resource.UpdateResponse{State: tfsdk.State{Schema: sr.Schema}}
		r.Update(ctx, resource.UpdateRequest{Plan: plan}, &resp)
		value, present := body["parent_folder_token"]
		if present != tc.present || (present && string(value) != tc.want) {
			t.Fatalf("parent %v: %s", tc.parent, body)
		}
	}
}

func TestVirtualTagParentUpdateClearsValues(t *testing.T) {
	ctx := context.Background()
	var body map[string]json.RawMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	schema := virtualTagConfigTestSchema(ctx)
	stateModel := virtualTagConfigTestModel(ctx, "old", false)
	planModel := virtualTagConfigTestModel(ctx, "new", false)
	empty, d := types.ListValue(resource_virtual_tag_config.ValuesValue{}.Type(ctx), nil)
	if d.HasError() {
		t.Fatal(d)
	}
	planModel.Values = empty
	state := virtualTagConfigTestState(t, ctx, schema, stateModel)
	plan := tfsdk.Plan{Schema: schema}
	if d := plan.Set(ctx, planModel); d.HasError() {
		t.Fatal(d)
	}
	resp := resource.UpdateResponse{State: state}
	r := VirtualTagConfigResource{client: clientForServer(t, server.URL)}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, &resp)
	if string(body["values"]) != "[]" || string(body["key"]) != `"new"` {
		t.Fatalf("parent update body: %s; diagnostics: %v", body, resp.Diagnostics)
	}
}
