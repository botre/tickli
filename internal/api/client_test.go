package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cliErrors "github.com/botre/tickli/internal/errors"
	"github.com/botre/tickli/internal/types"
	"github.com/go-resty/resty/v2"
)

// newTestClient creates a Client pointing at the given test server.
func newTestClient(serverURL string) *Client {
	rc := resty.New().SetBaseURL(serverURL)
	return &Client{http: rc}
}

func TestGetProject_Inbox(t *testing.T) {
	// GetProject("inbox") should return InboxProject without making any HTTP call.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	p, err := client.GetProject("inbox")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.ID != types.InboxProject.ID {
		t.Errorf("expected project ID %q, got %q", types.InboxProject.ID, p.ID)
	}
	if p.Name != types.InboxProject.Name {
		t.Errorf("expected project name %q, got %q", types.InboxProject.Name, p.Name)
	}
}

func TestGetProject_ByID(t *testing.T) {
	expected := types.Project{
		ID:   "abc123",
		Name: "Work",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/abc123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	p, err := client.GetProject("abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p.ID != expected.ID {
		t.Errorf("expected project ID %q, got %q", expected.ID, p.ID)
	}
	if p.Name != expected.Name {
		t.Errorf("expected project name %q, got %q", expected.Name, p.Name)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return empty JSON object (NullProject)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetProject("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent project, got nil")
	}
	if _, ok := err.(*cliErrors.NotFoundError); !ok {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestGetProject_NameDoesNotResolve(t *testing.T) {
	// Passing a project name (not an ID) should NOT resolve via name lookup.
	// The API will return a 404/error for a name string used as an ID.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The server receives the name as a path segment and returns an error.
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetProject("My Project")
	if err == nil {
		t.Fatal("expected error when passing a project name, got nil")
	}
}

func TestGetProject_EmojiNameDoesNotResolve(t *testing.T) {
	// Emoji-laden names like TickTick uses should not resolve.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetProject("📥Inbox")
	if err == nil {
		t.Fatal("expected error when passing emoji project name, got nil")
	}
}

func TestGetProject_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal server error`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.GetProject("abc123")
	if err == nil {
		t.Fatal("expected error on server error, got nil")
	}
}

func TestListProjects_IncludesInbox(t *testing.T) {
	apiProjects := []types.Project{
		{ID: "proj1", Name: "Work"},
		{ID: "proj2", Name: "Personal"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiProjects)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	projects, err := client.ListProjects()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should include the two API projects plus the synthetic inbox
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(projects))
	}

	hasInbox := false
	for _, p := range projects {
		if p.ID == types.InboxProject.ID {
			hasInbox = true
			break
		}
	}
	if !hasInbox {
		t.Error("expected InboxProject to be included in ListProjects result")
	}
}
