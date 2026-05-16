package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/botre/tickli/internal/types"
	"github.com/go-resty/resty/v2"
)

func TestNormalizeName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"🧼Chores", "chores"},
		{"  Work  ", "work"},
		{"Side Project", "side project"},
		{"📥Inbox", "inbox"},
	}
	for _, c := range cases {
		if got := normalizeName(c.in); got != c.want {
			t.Errorf("normalizeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"kitten", "sitting", 3},
		{"chores", "chores", 0},
		{"chores", "chore", 1},
	}
	for _, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestMatchProjectByName(t *testing.T) {
	projects := []types.Project{
		{ID: "1", Name: "🧼Chores"},
		{ID: "2", Name: "Work"},
		{ID: "3", Name: "Workout"},
		{ID: "4", Name: "Reading List"},
	}
	cases := []struct {
		name    string
		query   string
		wantID  string
		wantErr bool
	}{
		{"exact", "chores", "1", false},
		{"exact ignores emoji", "🧼Chores", "1", false},
		{"exact beats prefix", "work", "2", false},
		{"unique prefix", "read", "4", false},
		{"typo tolerance", "chorse", "1", false},
		{"ambiguous prefix", "wor", "", true},
		{"no match", "nonsense", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := matchProjectByName(projects, c.query)
			if c.wantErr {
				if err == nil {
					t.Fatalf("matchProjectByName(%q) = %q, want error", c.query, got.ID)
				}
				return
			}
			if err != nil {
				t.Fatalf("matchProjectByName(%q) unexpected error: %v", c.query, err)
			}
			if got.ID != c.wantID {
				t.Errorf("matchProjectByName(%q) = %q, want %q", c.query, got.ID, c.wantID)
			}
		})
	}
}

func TestApiErrorMessage(t *testing.T) {
	cases := []struct{ body, want string }{
		{`{"errorCode":"unauthorized","errorMessage":"bad token"}`, "bad token"},
		{`not json at all`, "not json at all"},
		{`{"unrelated":"field"}`, `{"unrelated":"field"}`},
	}
	for _, c := range cases {
		if got := apiErrorMessage(c.body); got != c.want {
			t.Errorf("apiErrorMessage(%q) = %q, want %q", c.body, got, c.want)
		}
	}
}

func TestSortTasksByDueDate(t *testing.T) {
	now := time.Now()
	mk := func(id string, due time.Time) types.Task {
		return types.Task{ID: id, DueDate: types.TickTickTime(due)}
	}
	tasks := []types.Task{
		mk("none", time.Time{}),
		mk("late", now.Add(48*time.Hour)),
		mk("soon", now.Add(time.Hour)),
	}
	sortTasksByDueDate(tasks)

	want := []string{"soon", "late", "none"}
	for i, id := range want {
		if tasks[i].ID != id {
			t.Errorf("position %d = %q, want %q", i, tasks[i].ID, id)
		}
	}
}

// TestGetTask verifies the task lookup resolves a task via a single filter
// call plus a single project fetch, rather than scanning every project.
func TestGetTask(t *testing.T) {
	var filterCalls, getCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/task/filter":
			filterCalls++
			_ = json.NewEncoder(w).Encode([]types.Task{
				{ID: "t1", ProjectID: "p1", Title: "First"},
				{ID: "t2", ProjectID: "p2", Title: "Second"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/project/p2/task/t2":
			getCalls++
			_ = json.NewEncoder(w).Encode(types.Task{ID: "t2", ProjectID: "p2", Title: "Second"})
		default:
			http.Error(w, "unexpected request: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := &Client{http: resty.New().SetBaseURL(srv.URL)}
	task, err := c.GetTask("t2")
	if err != nil {
		t.Fatalf("GetTask returned error: %v", err)
	}
	if task.Title != "Second" {
		t.Errorf("GetTask title = %q, want %q", task.Title, "Second")
	}
	if filterCalls != 1 {
		t.Errorf("filter endpoint hit %d times, want 1", filterCalls)
	}
	if getCalls != 1 {
		t.Errorf("project task endpoint hit %d times, want 1", getCalls)
	}
}

// TestListProjectsCached verifies the project list is fetched once and then
// served from the per-client cache.
func TestListProjectsCached(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]types.Project{{ID: "p1", Name: "Work"}})
	}))
	defer srv.Close()

	c := &Client{http: resty.New().SetBaseURL(srv.URL)}
	for i := 0; i < 3; i++ {
		projects, err := c.ListProjects()
		if err != nil {
			t.Fatalf("ListProjects returned error: %v", err)
		}
		// One project from the API plus the synthetic inbox project.
		if len(projects) != 2 {
			t.Fatalf("ListProjects returned %d projects, want 2", len(projects))
		}
	}
	if calls != 1 {
		t.Errorf("project endpoint hit %d times, want 1 (cached)", calls)
	}
}
