package vantage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vantage-sh/terraform-provider-vantage/vantage/acctest"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

// TestAccVantageWorkspace_mockAPI exercises the workspace resource against a
// minimal in-process HTTP server that mimics the Vantage /v2/workspaces API.
// This does not require Ruby or a real Rails app; it is sufficient to validate
// CRUD wiring and provider behavior (see VAN-956).
func TestAccVantageWorkspace_mockAPI(t *testing.T) {
	srv := newWorkspaceMockServer(t)
	t.Setenv("VANTAGE_API_TOKEN", "test-token")
	t.Setenv("VANTAGE_HOST", srv.URL)

	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	rNameUpdated := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	resourceName := "vantage_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceMockConfig(rName, "EUR", "true", "daily_rate"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "currency", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "enable_currency_conversion", "true"),
					resource.TestCheckResourceAttr(resourceName, "exchange_rate_date", "daily_rate"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				Config: testAccWorkspaceMockConfig(rNameUpdated, "GBP", "false", "end_of_billing_period_rate"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "currency", "GBP"),
					resource.TestCheckResourceAttr(resourceName, "enable_currency_conversion", "false"),
					resource.TestCheckResourceAttr(resourceName, "exchange_rate_date", "end_of_billing_period_rate"),
				),
			},
			{
				Config:             testAccWorkspaceMockConfig(rNameUpdated, "GBP", "false", "end_of_billing_period_rate"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWorkspaceMockConfig(name, currency, enableConv, exchangeDate string) string {
	return fmt.Sprintf(`
provider "vantage" {}

resource "vantage_workspace" "test" {
  name                         = %[1]q
  currency                     = %[2]q
  enable_currency_conversion   = %[3]s
  exchange_rate_date           = %[4]q
}
`, name, currency, enableConv, exchangeDate)
}

type workspaceMock struct {
	mu         sync.Mutex
	nextID     int
	workspaces map[string]*modelsv2.Workspace
}

func newWorkspaceMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	m := &workspaceMock{workspaces: make(map[string]*modelsv2.Workspace)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v2/workspaces") {
			http.NotFound(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/workspaces":
			m.handleList(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/v2/workspaces":
			m.handleCreate(w, r)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v2/workspaces/"):
			m.handleGet(w, r)
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/v2/workspaces/"):
			m.handleUpdate(w, r)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/v2/workspaces/"):
			m.handleDelete(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func (m *workspaceMock) tokenFromPath(path string) string {
	const prefix = "/v2/workspaces/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}

func (m *workspaceMock) handleList(w http.ResponseWriter, _ *http.Request) {
	m.mu.Lock()
	list := make([]*modelsv2.Workspace, 0, len(m.workspaces))
	for _, ws := range m.workspaces {
		list = append(list, ws)
	}
	m.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"workspaces": list})
}

func (m *workspaceMock) handleCreate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var in modelsv2.CreateWorkspace
	if err := json.Unmarshal(body, &in); err != nil || in.Name == nil || *in.Name == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	m.mu.Lock()
	m.nextID++
	token := fmt.Sprintf("wrkspc_mock_%d", m.nextID)
	currency := in.Currency
	if currency == "" {
		currency = "USD"
	}
	enable := false
	if in.EnableCurrencyConversion != nil {
		enable = *in.EnableCurrencyConversion
	}
	exchangeDate := "daily_rate"
	if in.ExchangeRateDate != nil && *in.ExchangeRateDate != "" {
		exchangeDate = *in.ExchangeRateDate
	}
	ws := &modelsv2.Workspace{
		Token:                    token,
		Name:                     *in.Name,
		Currency:                 currency,
		EnableCurrencyConversion: enable,
		ExchangeRateDate:         exchangeDate,
		CreatedAt:                "2026-01-01T00:00:00Z",
	}
	m.workspaces[token] = ws
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ws)
}

func (m *workspaceMock) handleGet(w http.ResponseWriter, r *http.Request) {
	token := m.tokenFromPath(r.URL.Path)
	if token == "" {
		http.NotFound(w, r)
		return
	}
	m.mu.Lock()
	ws, ok := m.workspaces[token]
	m.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ws)
}

func (m *workspaceMock) handleUpdate(w http.ResponseWriter, r *http.Request) {
	token := m.tokenFromPath(r.URL.Path)
	if token == "" {
		http.NotFound(w, r)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var in modelsv2.UpdateWorkspace
	if err := json.Unmarshal(body, &in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	m.mu.Lock()
	ws, ok := m.workspaces[token]
	if !ok {
		m.mu.Unlock()
		http.NotFound(w, r)
		return
	}
	if in.Name != "" {
		ws.Name = in.Name
	}
	if in.Currency != "" {
		ws.Currency = in.Currency
	}
	if in.EnableCurrencyConversion != nil {
		ws.EnableCurrencyConversion = *in.EnableCurrencyConversion
	}
	if in.ExchangeRateDate != nil && *in.ExchangeRateDate != "" {
		ws.ExchangeRateDate = *in.ExchangeRateDate
	}
	out := *ws
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&out)
}

func (m *workspaceMock) handleDelete(w http.ResponseWriter, r *http.Request) {
	token := m.tokenFromPath(r.URL.Path)
	if token == "" {
		http.NotFound(w, r)
		return
	}
	m.mu.Lock()
	delete(m.workspaces, token)
	m.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}
