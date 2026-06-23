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
// fetchAllFolders unit tests — driven by a mock HTTP server
// ---------------------------------------------------------------------------

func mockFolder(token, title, workspaceToken string, parentFolderToken *string) *modelsv2.Folder {
	return &modelsv2.Folder{
		Token:             token,
		Title:             &title,
		WorkspaceToken:    workspaceToken,
		ParentFolderToken: parentFolderToken,
		SavedFilterTokens: []string{},
	}
}

type foldersResponse struct {
	Folders []*modelsv2.Folder `json:"folders"`
	Links   *modelsv2.Links    `json:"links,omitempty"`
}

func newMockFoldersServer(t *testing.T, pages [][]*modelsv2.Folder) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/folders" {
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
			next := fmt.Sprintf("http://%s/v2/folders?limit=1000&page=%d", r.Host, pageNum+1)
			links = &modelsv2.Links{Next: &next}
		}

		resp := foldersResponse{
			Folders: pages[idx],
			Links:   links,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	return srv
}

func TestFetchAllFolders_singlePage(t *testing.T) {
	page1 := []*modelsv2.Folder{
		mockFolder("fldr_a", "Folder A", "wrkspc_1", nil),
		mockFolder("fldr_b", "Folder B", "wrkspc_1", nil),
	}
	srv := newMockFoldersServer(t, [][]*modelsv2.Folder{page1})
	defer srv.Close()

	got, err := fetchAllFolders(clientForServer(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d folders, want 2", len(got))
	}
	if got[0].Token != "fldr_a" || got[1].Token != "fldr_b" {
		t.Errorf("unexpected tokens: %v %v", got[0].Token, got[1].Token)
	}
}

func TestFetchAllFolders_multiplePages(t *testing.T) {
	parent := "fldr_root"
	page1 := []*modelsv2.Folder{mockFolder("fldr_1", "F1", "wrkspc_1", nil)}
	page2 := []*modelsv2.Folder{mockFolder("fldr_2", "F2", "wrkspc_1", &parent)}
	page3 := []*modelsv2.Folder{
		mockFolder("fldr_3", "F3", "wrkspc_2", nil),
		mockFolder("fldr_4", "F4", "wrkspc_2", nil),
	}

	srv := newMockFoldersServer(t, [][]*modelsv2.Folder{page1, page2, page3})
	defer srv.Close()

	got, err := fetchAllFolders(clientForServer(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 4 {
		t.Errorf("got %d folders, want 4", len(got))
	}
	want := []string{"fldr_1", "fldr_2", "fldr_3", "fldr_4"}
	for i, f := range got {
		if f.Token != want[i] {
			t.Errorf("index %d: got token %q, want %q", i, f.Token, want[i])
		}
	}
}

func TestFetchAllFolders_emptyResult(t *testing.T) {
	srv := newMockFoldersServer(t, [][]*modelsv2.Folder{{}})
	defer srv.Close()

	got, err := fetchAllFolders(clientForServer(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d folders, want 0", len(got))
	}
}

func TestFetchAllFolders_stopsAtNilNext(t *testing.T) {
	requestCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := foldersResponse{
			Folders: []*modelsv2.Folder{mockFolder("fldr_only", "Only", "wrkspc_1", nil)},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	got, err := fetchAllFolders(clientForServer(t, srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d folders, want 1", len(got))
	}
	if requestCount != 1 {
		t.Errorf("made %d requests, want exactly 1", requestCount)
	}
}

func TestFetchAllFolders_apiError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := fetchAllFolders(clientForServer(t, srv.URL))
	if err == nil {
		t.Fatal("expected error from API, got nil")
	}
}
