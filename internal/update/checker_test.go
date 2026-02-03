package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// githubRelease mimics the GitHub Releases API response structure.
type githubRelease struct {
	TagName     string        `json:"tag_name"`
	PublishedAt time.Time     `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func newTestServer(t *testing.T, release githubRelease) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := json.Marshal(release)
		if err != nil {
			t.Fatalf("marshal release: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
}

func TestChecker_CheckLatest_Success(t *testing.T) {
	t.Parallel()

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Assets: []githubAsset{
			{Name: "moai-darwin-arm64", BrowserDownloadURL: "https://example.com/moai-darwin-arm64"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
		},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	info, err := checker.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil VersionInfo")
	}
	if info.Version != "v1.2.0" {
		t.Errorf("Version = %q, want %q", info.Version, "v1.2.0")
	}
	if info.Date.IsZero() {
		t.Error("expected non-zero Date")
	}
}

func TestChecker_CheckLatest_NetworkError(t *testing.T) {
	t.Parallel()

	// Use a server that's immediately closed to simulate network error.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	info, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("expected error for closed server")
	}
	if info != nil {
		t.Error("expected nil VersionInfo on error")
	}
}

func TestChecker_CheckLatest_ContextCancelled(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	checker := NewChecker(ts.URL, http.DefaultClient)
	_, err := checker.CheckLatest(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestChecker_CheckLatest_InvalidJSON(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestChecker_CheckLatest_ServerError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestChecker_IsUpdateAvailable_NewerVersion(t *testing.T) {
	t.Parallel()

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Now(),
		Assets: []githubAsset{
			{Name: "moai-darwin-arm64", BrowserDownloadURL: "https://example.com/binary"},
		},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	available, info, err := checker.IsUpdateAvailable("v1.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected update to be available")
	}
	if info == nil || info.Version != "v1.2.0" {
		t.Errorf("expected version v1.2.0, got %v", info)
	}
}

func TestChecker_IsUpdateAvailable_AlreadyCurrent(t *testing.T) {
	t.Parallel()

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Now(),
		Assets:      []githubAsset{},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	available, info, err := checker.IsUpdateAvailable("v1.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected no update available")
	}
	if info != nil {
		t.Error("expected nil VersionInfo when already current")
	}
}

func TestChecker_IsUpdateAvailable_NewerCurrentVersion(t *testing.T) {
	t.Parallel()

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Now(),
		Assets:      []githubAsset{},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	available, info, err := checker.IsUpdateAvailable("v2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected no update when current is newer")
	}
	if info != nil {
		t.Error("expected nil VersionInfo when current is newer")
	}
}

func TestCompareSemver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"equal", "v1.0.0", "v1.0.0", 0},
		{"a newer major", "v2.0.0", "v1.0.0", 1},
		{"b newer major", "v1.0.0", "v2.0.0", -1},
		{"a newer minor", "v1.2.0", "v1.1.0", 1},
		{"b newer minor", "v1.1.0", "v1.2.0", -1},
		{"a newer patch", "v1.0.2", "v1.0.1", 1},
		{"b newer patch", "v1.0.1", "v1.0.2", -1},
		{"no v prefix", "1.0.0", "1.0.0", 0},
		{"mixed prefix", "v1.0.0", "1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := compareSemver(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
