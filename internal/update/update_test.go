package update

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in   string
		want [3]int
		ok   bool
	}{
		{"v1.2.3", [3]int{1, 2, 3}, true},
		{"0.0.42", [3]int{0, 0, 42}, true},
		{"  v2.10.0  ", [3]int{2, 10, 0}, true},
		{"v1.2.3-rc1", [3]int{1, 2, 3}, true},
		{"v1.2.3+meta", [3]int{1, 2, 3}, true},
		{"dev", [3]int{}, false},
		{"v1.2", [3]int{}, false},
		{"v1.2.x", [3]int{}, false},
		{"", [3]int{}, false},
	}
	for _, c := range cases {
		got, ok := parseVersion(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("parseVersion(%q) = %v, %v; want %v, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.0.0", "v1.0.1", -1},
		{"v0.0.10", "v0.0.9", 1},
		{"v1.2.3", "v1.2.3", 0},
		{"v0.1.0", "v0.0.99", 1},
		{"v2.0.0", "v1.9.9", 1},
		{"garbage", "v1.0.0", 0},
		{"v1.0.0", "also-garbage", 0},
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestPickNewer(t *testing.T) {
	if got := pickNewer("v0.0.5", "v0.0.4"); got != "v0.0.5" {
		t.Errorf("pickNewer newer = %q, want v0.0.5", got)
	}
	if got := pickNewer("v0.0.4", "v0.0.4"); got != "" {
		t.Errorf("pickNewer equal = %q, want empty", got)
	}
	if got := pickNewer("v0.0.3", "v0.0.4"); got != "" {
		t.Errorf("pickNewer older = %q, want empty", got)
	}
}

func TestShouldCheck(t *testing.T) {
	if shouldCheck("dev") {
		t.Error("shouldCheck(dev) = true, want false")
	}
	if shouldCheck("(devel)") {
		t.Error("shouldCheck((devel)) = true, want false")
	}
	if !shouldCheck("v0.0.1") {
		t.Error("shouldCheck(v0.0.1) = false, want true")
	}

	t.Setenv(disableEnv, "1")
	if shouldCheck("v0.0.1") {
		t.Errorf("shouldCheck with %s set = true, want false", disableEnv)
	}
}

func TestFetchLatest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Version":"v0.0.50","Time":"2026-05-10T00:00:00Z"}`))
		}))
		defer srv.Close()

		got, err := fetchLatest(srv.URL)
		if err != nil {
			t.Fatalf("fetchLatest returned error: %v", err)
		}
		if got != "v0.0.50" {
			t.Errorf("fetchLatest = %q, want v0.0.50", got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", http.StatusNotFound)
		}))
		defer srv.Close()

		if _, err := fetchLatest(srv.URL); err == nil {
			t.Error("fetchLatest on 404 returned nil error, want error")
		}
	})

	t.Run("empty version", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		if _, err := fetchLatest(srv.URL); err == nil {
			t.Error("fetchLatest on empty version returned nil error, want error")
		}
	})
}
