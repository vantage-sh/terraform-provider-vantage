package vantage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
)

// ---------------------------------------------------------------------------
// fetchAllWorkspaces unit tests — driven by a mock HTTP server
// ---------------------------------------------------------------------------

func mockWorkspace(token, name string) *modelsv2.Workspace {
	return &modelsv2.Workspace{
		Token: token,
		Name:  name,
	}
}

type workspacesResponse struct {
	Workspaces []*modelsv2.Workspace `json:"workspaces"`
	Links      *modelsv2.Links       `json:"links,omitempty"`
}

func newMockWorkspacesServer(t *testing.T, pages [][]*modelsv2.Workspace) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/workspaces" {
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
			next := fmt.Sprintf("http://%s/v2/workspaces?limit=1000&page=%d", r.Host, pageNum+1)
			links = &modelsv2.Links{Next: &next}
		}

		resp := workspacesResponse{
			Workspaces: pages[idx],
			Links:      links,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	return srv
}

func TestFetchAllWorkspaces_singlePage(t *testing.T) {
	page1 := []*modelsv2.Workspace{
		mockWorkspace("wrkspc_a", "Workspace A"),
		mockWorkspace("wrkspc_b", "Workspace B"),
	}
	srv := newMockWorkspacesServer(t, [][]*modelsv2.Workspace{page1})
	defer srv.Close()

	got, err := fetchAllWorkspaces(clientForServer(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d workspaces, want 2", len(got))
	}
	if got[0].Token != "wrkspc_a" || got[1].Token != "wrkspc_b" {
		t.Errorf("unexpected tokens: %v %v", got[0].Token, got[1].Token)
	}
}

func TestFetchAllWorkspaces_multiplePages(t *testing.T) {
	page1 := []*modelsv2.Workspace{mockWorkspace("wrkspc_1", "W1")}
	page2 := []*modelsv2.Workspace{mockWorkspace("wrkspc_2", "W2")}
	page3 := []*modelsv2.Workspace{mockWorkspace("wrkspc_3", "W3"), mockWorkspace("wrkspc_4", "W4")}

	srv := newMockWorkspacesServer(t, [][]*modelsv2.Workspace{page1, page2, page3})
	defer srv.Close()

	got, err := fetchAllWorkspaces(clientForServer(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 4 {
		t.Errorf("got %d workspaces, want 4", len(got))
	}
	want := []string{"wrkspc_1", "wrkspc_2", "wrkspc_3", "wrkspc_4"}
	for i, ws := range got {
		if ws.Token != want[i] {
			t.Errorf("index %d: got token %q, want %q", i, ws.Token, want[i])
		}
	}
}

func TestFetchAllWorkspaces_emptyResult(t *testing.T) {
	srv := newMockWorkspacesServer(t, [][]*modelsv2.Workspace{{}})
	defer srv.Close()

	got, err := fetchAllWorkspaces(clientForServer(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d workspaces, want 0", len(got))
	}
}

func TestFetchAllWorkspaces_stopsAtNilNext(t *testing.T) {
	requestCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := workspacesResponse{
			Workspaces: []*modelsv2.Workspace{mockWorkspace("wrkspc_only", "Only")},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	got, err := fetchAllWorkspaces(clientForServer(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d workspaces, want 1", len(got))
	}
	if requestCount != 1 {
		t.Errorf("made %d requests, want exactly 1", requestCount)
	}
}

func TestFetchAllWorkspaces_apiError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := fetchAllWorkspaces(clientForServer(t, srv.URL))
	if err == nil {
		t.Fatal("expected error from API, got nil")
	}
}
