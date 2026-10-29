package vantage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAttachmentLabelFiltersCanonicalLifecycle(t *testing.T) {
	for _, kind := range []string{"cost_report", "business_metric"} {
		for _, legacyResponse := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/legacy_response=%t", kind, legacyResponse), func(t *testing.T) {
				api := &attachmentLabelFiltersAPI{t: t, kind: kind, legacyResponse: legacyResponse}
				server := httptest.NewServer(api)
				defer server.Close()
				address := "vantage_" + kind + ".test"
				attachmentField := "business_metric_tokens_with_metadata"
				if kind == "business_metric" {
					attachmentField = "cost_report_tokens_with_metadata"
				}
				legacy := attachmentLabelFiltersConfig(server.URL, kind, `label_filter = ["production", ""]`)
				updated := attachmentLabelFiltersConfig(server.URL, kind, `label_filter = ["development"]`)
				cleared := attachmentLabelFiltersConfig(server.URL, kind, `label_filter = []`)
				canonical := attachmentLabelFiltersConfig(server.URL, kind, `label_filters = { label = ["production", ""], team = ["platform"] }`)
				bothEquivalent := attachmentLabelFiltersConfig(server.URL, kind, `label_filter = ["", "production", "production"]
label_filters = { label = ["production", ""], team = ["platform"] }`)
				emptyAlias := attachmentLabelFiltersConfig(server.URL, kind, `label_filter = []
label_filters = { label = ["production", ""], team = ["platform"] }`)
				canonicalCleared := attachmentLabelFiltersConfig(server.URL, kind, `label_filters = {}`)
				omitted := attachmentLabelFiltersConfig(server.URL, kind, "")
				omittedUpdated := strings.Replace(omitted, "Canonical attachment filters", "Renamed filter attachment", 1)
				topOmitted := attachmentLabelFiltersWithoutAttachmentsConfig(server.URL, kind)
				detached := strings.Replace(topOmitted, `title = "Omitted attachment list"`, `title = "Omitted attachment list"`+"\n"+attachmentField+" = []", 1)
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{Config: legacy, Check: resource.TestCheckResourceAttr(address, attachmentField+".0.label_filter.0", "production")},
						{RefreshState: true, Check: resource.TestCheckResourceAttr(address, attachmentField+".0.label_filter.0", "production")},
						{Config: legacy, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: updated, Check: resource.TestCheckResourceAttr(address, attachmentField+".0.label_filter.0", "development")},
						{Config: updated, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: cleared, Check: resource.TestCheckResourceAttr(address, attachmentField+".0.label_filter.#", "0")},
						{Config: cleared, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: canonical, Check: resource.TestCheckResourceAttr(address, attachmentField+".0.label_filters.team.0", "platform")},
						{Config: canonical, PlanOnly: true, ExpectNonEmptyPlan: false},
						{ResourceName: address, ImportState: true, ImportStateVerify: true},
						{Config: canonical, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: omitted, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: omittedUpdated, Check: resource.TestCheckResourceAttr(address, attachmentField+".0.label_filters.team.0", "platform")},
						{Config: omittedUpdated, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: bothEquivalent},
						{Config: bothEquivalent, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: emptyAlias},
						{Config: emptyAlias, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: canonicalCleared, Check: resource.TestCheckResourceAttr(address, attachmentField+".0.label_filters.%", "0")},
						{Config: canonicalCleared, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: legacy},
						{ResourceName: address, ImportState: true, ImportStateVerify: true},
						{Config: legacy, PlanOnly: true, ExpectNonEmptyPlan: false},
						{Config: detached, Check: resource.TestCheckResourceAttr(address, attachmentField+".#", "0")},
						{Config: detached, PlanOnly: true, ExpectNonEmptyPlan: false},
					},
				})
			})
		}
	}
}

func attachmentLabelFiltersConfig(host, kind, filters string) string {
	attachmentField := "business_metric_tokens_with_metadata"
	tokenField := "business_metric_token"
	resourceFields := `workspace_token = "wrkspc_test"`
	if kind == "business_metric" {
		attachmentField = "cost_report_tokens_with_metadata"
		tokenField = "cost_report_token"
		resourceFields = ""
	}
	return fmt.Sprintf(`
provider "vantage" {
  host = %q
  api_token = "local-test"
}
resource "vantage_%s" "test" {
  title = "Canonical attachment filters"
  %s
  %s = [{
    %s = "attached_test"
    unit_scale = "per_unit"
    %s
  }]
}
`, host, kind, resourceFields, attachmentField, tokenField, filters)
}

func attachmentLabelFiltersWithoutAttachmentsConfig(host, kind string) string {
	workspace := ""
	if kind == "cost_report" {
		workspace = `workspace_token = "wrkspc_test"`
	}
	return fmt.Sprintf(`provider "vantage" {
      host = %q
      api_token = "local-test"
    }
    resource "vantage_%s" "test" {
      title = "Omitted attachment list"
      %s
    }`, host, kind, workspace)
}

type attachmentLabelFiltersAPI struct {
	t              *testing.T
	mu             sync.Mutex
	kind           string
	legacyResponse bool
	resource       map[string]any
}

func (api *attachmentLabelFiltersAPI) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()
	path := "/v2/" + api.kind + "s"
	if !strings.HasPrefix(req.URL.Path, path) {
		api.t.Errorf("unexpected API path: %s", req.URL.Path)
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodDelete {
		api.resource = nil
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			api.t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if api.resource == nil {
			api.resource = map[string]any{
				"token": "resource_test", "created_at": "2026-10-29T00:00:00Z", "workspace_token": "wrkspc_test",
				"filter": "", "groupings": "", "chart_type": "line", "date_bin": "cumulative", "date_interval": "last_month",
				"start_date": "2026-09-01", "end_date": "2026-09-30", "previous_period_start_date": "2026-08-01", "previous_period_end_date": "2026-08-31",
				"saved_filter_tokens": []any{}, "import_type": "csv",
			}
		}
		for key, value := range body {
			api.resource[key] = value
		}
		field := "business_metric_tokens_with_metadata"
		if api.kind == "business_metric" {
			field = "cost_report_tokens_with_metadata"
		}
		if attachments, exists := body[field]; exists {
			if attachments == nil {
				attachments = []any{}
				api.resource[field] = attachments
			}
			for _, raw := range attachments.([]any) {
				attachment := raw.(map[string]any)
				if _, exists := attachment["label_filter"]; exists {
					api.t.Errorf("singular alias leaked into canonical request: %#v", attachment)
				}
				if _, exists := attachment["label_filters"]; !exists {
					attachment["label_filters"] = map[string]any{}
				}
				if _, exists := attachment["calculation_type"]; !exists {
					attachment["calculation_type"] = "unit_cost"
				}
			}
		}
		if req.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
	}
	if api.resource == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	response := api.resource
	if api.legacyResponse {
		encoded, err := json.Marshal(api.resource)
		response = nil
		if err != nil {
			api.t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, &response); err != nil {
			api.t.Fatal(err)
		}
		field := "business_metric_tokens_with_metadata"
		if api.kind == "business_metric" {
			field = "cost_report_tokens_with_metadata"
		}
		if attachments, ok := response[field].([]any); ok {
			for _, raw := range attachments {
				attachment := raw.(map[string]any)
				filters := attachment["label_filters"].(map[string]any)
				if label, exists := filters["label"]; exists {
					attachment["label_filter"] = label
					delete(filters, "label")
				}
				if len(filters) == 0 {
					delete(attachment, "label_filters")
				}
			}
		}
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.t.Error(err)
	}
}
