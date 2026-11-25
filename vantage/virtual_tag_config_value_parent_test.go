package vantage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVirtualTagConfigStandaloneValues(t *testing.T) {
	var mutex sync.Mutex
	var parent, child map[string]any
	var parentCreates, parentUpdates, childCreates, childUpdates, parentDeletes, childDeletes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		var body map[string]any
		if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Errorf("decode %s %s: %v", req.Method, req.URL.Path, err)
			}
		}
		switch req.URL.Path {
		case "/v2/tags":
			if req.Method != http.MethodPut {
				t.Errorf("unexpected tag method: %s", req.Method)
			}
			fmt.Fprint(w, `{"tags":[]}`)
		case "/v2/virtual_tag_configs", "/v2/virtual_tag_configs/vtag_1":
			switch req.Method {
			case http.MethodPost:
				parentCreates++
				parent = map[string]any{
					"token": "vtag_1", "key": body["key"], "backfill_until": "2026-01-01",
					"overridable": false, "hidden": false, "preferred": false, "collapsed_tag_keys": []any{},
				}
				w.WriteHeader(http.StatusCreated)
			case http.MethodPut:
				parentUpdates++
				if values, present := body["values"]; present {
					t.Errorf("metadata-only parent update must omit values, got %#v", values)
					child = nil
				}
				parent["key"] = body["key"]
			case http.MethodGet:
				if parent == nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}
			case http.MethodDelete:
				parentDeletes++
				if child != nil {
					t.Error("parent deleted before its standalone value")
				}
				parent = nil
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				t.Errorf("unexpected parent method: %s", req.Method)
			}
			parent["values"] = []any{}
			if child != nil {
				parent["values"] = []any{child}
			}
			if err := json.NewEncoder(w).Encode(parent); err != nil {
				t.Errorf("encode parent: %v", err)
			}
		case "/v2/virtual_tag_configs/vtag_1/values", "/v2/virtual_tag_configs/vtag_1/values/vtag_val_1":
			switch req.Method {
			case http.MethodPost:
				childCreates++
				if parent == nil {
					t.Error("value created before its parent")
				}
				child = map[string]any{"token": "vtag_val_1", "percentages": []any{}, "date_ranges": []any{}, "label_transforms": []any{}}
				w.WriteHeader(http.StatusCreated)
			case http.MethodPatch:
				childUpdates++
			case http.MethodGet:
				if child == nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}
			case http.MethodDelete:
				childDeletes++
				child = nil
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				t.Errorf("unexpected child method: %s", req.Method)
			}
			for key, value := range body {
				if value != nil {
					child[key] = value
				}
			}
			if err := json.NewEncoder(w).Encode(child); err != nil {
				t.Errorf("encode child: %v", err)
			}
		default:
			t.Errorf("unexpected endpoint: %s %s", req.Method, req.URL.Path)
			http.NotFound(w, req)
		}
	}))
	defer server.Close()
	config := func(key, name string) string {
		return fmt.Sprintf(`
provider "vantage" {
  host = %q
  api_token = "test-token"
}
resource "vantage_virtual_tag_config" "test" {
  key = %q
	overridable = false
}
resource "vantage_virtual_tag_config_value" "test" {
  virtual_tag_config_token = vantage_virtual_tag_config.test.token
  filter = "costs.provider = 'aws'"
  name = %q
}
data "vantage_virtual_tag_config_values" "test" {
  virtual_tag_config_token = vantage_virtual_tag_config.test.token
	depends_on = [vantage_virtual_tag_config_value.test]
}
`, server.URL, key, name)
	}
	checkChild := func(name string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("vantage_virtual_tag_config_value.test", "token", "vtag_val_1"),
			resource.TestCheckResourceAttr("vantage_virtual_tag_config_value.test", "name", name),
			resource.TestCheckResourceAttr("data.vantage_virtual_tag_config_values.test", "values.0.name", name),
		)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("original-key", "Engineering"), Check: checkChild("Engineering")},
			{Config: config("original-key", "Engineering"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config("updated-key", "Engineering"), Check: checkChild("Engineering")},
			{Config: config("updated-key", "Engineering"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config("updated-key", "Platform"), Check: checkChild("Platform")},
			{Config: config("updated-key", "Platform"), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
	mutex.Lock()
	defer mutex.Unlock()
	if parentCreates != 1 || parentUpdates != 1 || childCreates != 1 || childUpdates != 1 || parentDeletes != 1 || childDeletes != 1 {
		t.Fatalf("API calls: parent create/update/delete=%d/%d/%d child create/update/delete=%d/%d/%d",
			parentCreates, parentUpdates, parentDeletes, childCreates, childUpdates, childDeletes)
	}
}
