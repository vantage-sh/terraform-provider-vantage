package vantage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

// ---------------------------------------------------------------------------
// pageFromURL unit tests
// ---------------------------------------------------------------------------

func TestPageFromURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    int32
		wantErr bool
	}{
		{
			name:   "standard next-page URL",
			rawURL: "https://api.vantage.sh/v2/integrations?limit=1000&page=2",
			want:   2,
		},
		{
			name:   "page 10",
			rawURL: "https://api.vantage.sh/v2/integrations?page=10&limit=1000&provider=custom_provider",
			want:   10,
		},
		{
			name:    "missing page parameter",
			rawURL:  "https://api.vantage.sh/v2/integrations?limit=1000",
			wantErr: true,
		},
		{
			name:    "non-integer page value",
			rawURL:  "https://api.vantage.sh/v2/integrations?page=two",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			rawURL:  "://bad-url",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := pageFromURL(tc.rawURL)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%d)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got page %d, want %d", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// fetchAllIntegrations unit tests — driven by a mock HTTP server
// ---------------------------------------------------------------------------

// mockIntegration builds a minimal *modelsv2.Integration for use in test payloads.
func mockIntegration(token, name string) *modelsv2.Integration {
	acct := name
	return &modelsv2.Integration{
		Token:                token,
		AccountIdentifier:    &acct,
		Provider:             "custom_provider",
		Status:               modelsv2.IntegrationStatusConnected,
		CreatedAt:            "2024-01-01T00:00:00Z",
		ManagedAccountTokens: []string{},
		WorkspaceTokens:      []string{},
	}
}

// integrationsResponse mirrors the JSON shape the Vantage API returns for
// GET /v2/integrations so that the SDK can deserialise it correctly.
type integrationsResponse struct {
	Integrations []*modelsv2.Integration `json:"integrations"`
	Links        *modelsv2.Links         `json:"links,omitempty"`
}

// newMockIntegrationsServer returns an httptest.Server that serves paginated
// integration results. Each element of pages is served on the corresponding
// page number (1-indexed). links.next is set automatically for all but the
// last page.
func newMockIntegrationsServer(t *testing.T, pages [][]*modelsv2.Integration) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/integrations" {
			http.NotFound(w, r)
			return
		}

		pageStr := r.URL.Query().Get("page")
		pageNum := 1
		if pageStr != "" {
			fmt.Sscanf(pageStr, "%d", &pageNum)
		}

		idx := pageNum - 1
		if idx < 0 || idx >= len(pages) {
			http.Error(w, "page out of range", http.StatusBadRequest)
			return
		}

		var links *modelsv2.Links
		if pageNum < len(pages) {
			next := fmt.Sprintf("%s/v2/integrations?limit=1000&page=%d", r.Host, pageNum+1)
			// Include the scheme so url.Parse works in pageFromURL.
			next = "http://" + next
			links = &modelsv2.Links{Next: &next}
		}

		resp := integrationsResponse{
			Integrations: pages[idx],
			Links:        links,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	return srv
}

// clientForServer creates a *Client pointing at the given test server URL.
func clientForServer(t *testing.T, serverURL string) *Client {
	t.Helper()
	c, err := NewClient(serverURL, "test-token", false, 10*time.Second)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestFetchAllIntegrations_singlePage(t *testing.T) {
	page1 := []*modelsv2.Integration{
		mockIntegration("intgr_a", "Provider A"),
		mockIntegration("intgr_b", "Provider B"),
	}
	srv := newMockIntegrationsServer(t, [][]*modelsv2.Integration{page1})
	defer srv.Close()

	got, err := fetchAllIntegrations(clientForServer(t, srv.URL), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d integrations, want 2", len(got))
	}
	if got[0].Token != "intgr_a" || got[1].Token != "intgr_b" {
		t.Errorf("unexpected tokens: %v %v", got[0].Token, got[1].Token)
	}
}

func TestFetchAllIntegrations_multiplePages(t *testing.T) {
	page1 := []*modelsv2.Integration{mockIntegration("intgr_1", "P1")}
	page2 := []*modelsv2.Integration{mockIntegration("intgr_2", "P2")}
	page3 := []*modelsv2.Integration{mockIntegration("intgr_3", "P3"), mockIntegration("intgr_4", "P4")}

	srv := newMockIntegrationsServer(t, [][]*modelsv2.Integration{page1, page2, page3})
	defer srv.Close()

	got, err := fetchAllIntegrations(clientForServer(t, srv.URL), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 4 {
		t.Errorf("got %d integrations, want 4", len(got))
	}
	tokens := []string{got[0].Token, got[1].Token, got[2].Token, got[3].Token}
	want := []string{"intgr_1", "intgr_2", "intgr_3", "intgr_4"}
	for i, tok := range tokens {
		if tok != want[i] {
			t.Errorf("index %d: got token %q, want %q", i, tok, want[i])
		}
	}
}

func TestFetchAllIntegrations_emptyResult(t *testing.T) {
	srv := newMockIntegrationsServer(t, [][]*modelsv2.Integration{{}})
	defer srv.Close()

	got, err := fetchAllIntegrations(clientForServer(t, srv.URL), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d integrations, want 0", len(got))
	}
}

func TestFetchAllIntegrations_withProviderFilter(t *testing.T) {
	var capturedProvider string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedProvider = r.URL.Query().Get("provider")
		resp := integrationsResponse{
			Integrations: []*modelsv2.Integration{mockIntegration("intgr_x", "X")},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	filter := "custom_provider"
	_, err := fetchAllIntegrations(clientForServer(t, srv.URL), &filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedProvider != "custom_provider" {
		t.Errorf("provider filter not forwarded: got %q, want %q", capturedProvider, "custom_provider")
	}
}

func TestFetchAllIntegrations_stopsAtNilNext(t *testing.T) {
	// Verify the loop does not make extra requests once links.next is nil.
	requestCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// Always return links.next = nil, regardless of page param.
		resp := integrationsResponse{
			Integrations: []*modelsv2.Integration{mockIntegration("intgr_only", "Only")},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	got, err := fetchAllIntegrations(clientForServer(t, srv.URL), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d integrations, want 1", len(got))
	}
	if requestCount != 1 {
		t.Errorf("made %d requests, want exactly 1", requestCount)
	}
}

func TestFetchAllIntegrations_apiError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := fetchAllIntegrations(clientForServer(t, srv.URL), nil)
	if err == nil {
		t.Fatal("expected error from API, got nil")
	}
}
