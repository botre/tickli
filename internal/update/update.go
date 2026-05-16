// Package update checks whether a newer release of tickli is available and
// upgrades the binary in place by re-running `go install`.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

const (
	// modulePath is the Go module path used to install tickli.
	modulePath = "github.com/botre/tickli"
	// proxyURL returns the latest published version as JSON ({"Version":...}).
	// It is the same source of truth `go install ...@latest` uses.
	proxyURL = "https://proxy.golang.org/github.com/botre/tickli/@latest"
	// disableEnv, when set, turns off the background update check.
	disableEnv = "TICKLI_NO_UPDATE_CHECK"

	checkTTL     = 24 * time.Hour
	fetchTimeout = 3 * time.Second
	// printGrace bounds how long the notice printer waits for a slow check.
	printGrace = 1500 * time.Millisecond
)

// cache is the on-disk record of the most recent update check.
type cache struct {
	CheckedAt     time.Time `json:"checkedAt"`
	LatestVersion string    `json:"latestVersion"`
}

// StartCheck launches an update check in the background and returns a function
// that prints a one-line notice to w when a newer version is available.
//
// The check runs concurrently with the caller's work (which is typically
// blocked on network I/O of its own), so it adds no perceptible latency. The
// returned function waits at most printGrace for the check to finish.
func StartCheck(current string) func(w io.Writer) {
	if !shouldCheck(current) {
		return func(io.Writer) {}
	}

	result := make(chan string, 1)
	go func() {
		result <- newerVersion(current)
	}()

	return func(w io.Writer) {
		select {
		case latest := <-result:
			if latest != "" {
				fmt.Fprintf(w, "\nA new version of tickli is available: %s (you have %s)\n", latest, current)
				fmt.Fprintln(w, "Run 'tickli update' to upgrade.")
			}
		case <-time.After(printGrace):
			// The check is too slow this run; the cache it primes will be
			// used on the next invocation.
		}
	}
}

// Run upgrades tickli in place by re-running `go install`.
func Run(stdout, stderr io.Writer) error {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("the Go toolchain is required to run 'tickli update' but 'go' is not on your PATH.\n" +
			"Install Go from https://go.dev/dl/, or upgrade tickli with whatever package manager you installed it with")
	}

	target := modulePath + "@latest"
	fmt.Fprintf(stderr, "Running: go install %s\n", target)

	c := exec.Command(goBin, "install", target)
	c.Stdout = stdout
	c.Stderr = stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}

	fmt.Fprintln(stdout, "tickli updated. Run 'tickli version' to confirm.")
	return nil
}

// shouldCheck reports whether an update check is worth running.
func shouldCheck(current string) bool {
	if _, disabled := os.LookupEnv(disableEnv); disabled {
		return false
	}
	// A dev build (no release tag) has nothing meaningful to compare against.
	if _, ok := parseVersion(current); !ok {
		return false
	}
	return true
}

// newerVersion returns the latest version if it is newer than current,
// otherwise an empty string. It serves a fresh result from the on-disk cache
// and only hits the network when the cache is missing or stale.
func newerVersion(current string) string {
	c, _ := loadCache()
	if c != nil && time.Since(c.CheckedAt) < checkTTL {
		return pickNewer(c.LatestVersion, current)
	}

	latest, err := fetchLatest(proxyURL)
	if err != nil {
		// Network failed: fall back to a stale cache rather than nothing.
		if c != nil {
			return pickNewer(c.LatestVersion, current)
		}
		return ""
	}

	_ = saveCache(&cache{CheckedAt: time.Now(), LatestVersion: latest})
	return pickNewer(latest, current)
}

// pickNewer returns latest when it is strictly newer than current.
func pickNewer(latest, current string) string {
	if latest != "" && compareVersions(latest, current) > 0 {
		return latest
	}
	return ""
}

// fetchLatest queries the Go module proxy for the latest published version.
func fetchLatest(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("module proxy returned status %d", resp.StatusCode)
	}

	var body struct {
		Version string `json:"Version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Version == "" {
		return "", fmt.Errorf("module proxy returned an empty version")
	}
	return body.Version, nil
}

// cachePath returns the path to the update-check cache file, creating its
// parent directory if needed.
func cachePath() (string, error) {
	return xdg.CacheFile("tickli/update-check.json")
}

func loadCache() (*cache, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// saveCache writes the cache atomically so a process exiting mid-write cannot
// leave a corrupt file behind.
func saveCache(c *cache) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// compareVersions compares two "vX.Y.Z" version strings, returning -1 if a is
// older than b, 1 if newer, and 0 if equal or either is unparseable.
func compareVersions(a, b string) int {
	pa, oka := parseVersion(a)
	pb, okb := parseVersion(b)
	if !oka || !okb {
		return 0
	}
	for i := 0; i < 3; i++ {
		switch {
		case pa[i] < pb[i]:
			return -1
		case pa[i] > pb[i]:
			return 1
		}
	}
	return 0
}

// parseVersion parses a "vMAJOR.MINOR.PATCH" string into its numeric parts.
// Any pre-release or build metadata suffix is ignored.
func parseVersion(v string) ([3]int, bool) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}
