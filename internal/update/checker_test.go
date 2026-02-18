package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
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
		TagName:     "go-v1.2.0",
		PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Assets: []githubAsset{
			{Name: "moai-adk_go-v1.2.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/moai-adk_go-v1.2.0_darwin_arm64.tar.gz"},
			{Name: "moai-adk_go-v1.2.0_windows_amd64.zip", BrowserDownloadURL: "https://example.com/moai-adk_go-v1.2.0_windows_amd64.zip"},
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
	if info.Version != "go-v1.2.0" {
		t.Errorf("Version = %q, want %q", info.Version, "go-v1.2.0")
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
		TagName:     "go-v1.2.0",
		PublishedAt: time.Now(),
		Assets: []githubAsset{
			{Name: "moai-adk_go-v1.2.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/binary.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
		},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	available, info, err := checker.IsUpdateAvailable("go-v1.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected update to be available")
	}
	if info == nil || info.Version != "go-v1.2.0" {
		t.Errorf("expected version go-v1.2.0, got %v", info)
	}
}

func TestChecker_IsUpdateAvailable_AlreadyCurrent(t *testing.T) {
	t.Parallel()

	release := githubRelease{
		TagName:     "go-v1.2.0",
		PublishedAt: time.Now(),
		Assets:      []githubAsset{},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	available, info, err := checker.IsUpdateAvailable("go-v1.2.0")
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
		TagName:     "go-v1.2.0",
		PublishedAt: time.Now(),
		Assets:      []githubAsset{},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	available, info, err := checker.IsUpdateAvailable("go-v2.0.0")
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
		{"go-v prefix equal", "go-v1.0.0", "go-v1.0.0", 0},
		{"go-v prefix a newer", "go-v2.0.0", "go-v1.0.0", 1},
		{"go-v prefix b newer", "go-v1.0.0", "go-v2.0.0", -1},
		{"go-v vs v prefix", "go-v1.0.0", "v1.0.0", 0},
		{"go-v vs no prefix", "go-v1.0.0", "1.0.0", 0},
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

func TestChecker_CheckLatest_WithChecksums(t *testing.T) {
	t.Parallel()

	// Platform-specific expected values based on runtime.GOOS and runtime.GOARCH
	checksumMap := map[string]string{
		"darwin_amd64":  "392f7de6e7b21c1e4d1681a424ff254ab51b76fa0a8688e7d76c71226b3f520f",
		"darwin_arm64":  "99bf529b15df380c912d2be43b475122000ff9d6e1bb9069699b6199f9ef5d90",
		"linux_amd64":   "a1423b990826e23bfb830a22bb9f4c42254e12a6ddfc2d958abd53e946e789c2",
		"linux_arm64":   "b2534c001927f34cfb941b33cc10e6443355e23d7cfa2d069800c7200a0ef6e3",
		"windows_amd64": "c3645d112038a45dac052c44dd21f7554466f34e8dfb3e170911d8311b1fa7f4",
	}

	// Create a mock checksums.txt server
	checksumsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		checksumsContent := `392f7de6e7b21c1e4d1681a424ff254ab51b76fa0a8688e7d76c71226b3f520f  moai-adk_1.2.0_darwin_amd64.tar.gz
99bf529b15df380c912d2be43b475122000ff9d6e1bb9069699b6199f9ef5d90  moai-adk_1.2.0_darwin_arm64.tar.gz
a1423b990826e23bfb830a22bb9f4c42254e12a6ddfc2d958abd53e946e789c2  moai-adk_1.2.0_linux_amd64.tar.gz
b2534c001927f34cfb941b33cc10e6443355e23d7cfa2d069800c7200a0ef6e3  moai-adk_1.2.0_linux_arm64.tar.gz
c3645d112038a45dac052c44dd21f7554466f34e8dfb3e170911d8311b1fa7f4  moai-adk_1.2.0_windows_amd64.zip`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(checksumsContent))
	}))
	defer checksumsTS.Close()

	// Build platform-specific archive name
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	archiveName := fmt.Sprintf("moai-adk_1.2.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Assets: []githubAsset{
			{Name: archiveName, BrowserDownloadURL: "https://example.com/moai.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: checksumsTS.URL},
		},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	info, err := checker.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify checksum was extracted correctly for this platform
	platformKey := runtime.GOOS + "_" + runtime.GOARCH
	expectedChecksum, ok := checksumMap[platformKey]
	if !ok {
		t.Skipf("no expected checksum for platform %s", platformKey)
	}
	if info.Checksum != expectedChecksum {
		t.Errorf("Checksum = %q, want %q", info.Checksum, expectedChecksum)
	}
}

func TestChecker_CheckLatest_ChecksumDownloadFailed(t *testing.T) {
	t.Parallel()

	// Create a checksums.txt server that returns 404
	checksumsTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer checksumsTS.Close()

	// Build platform-specific archive name so the asset is always found
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	archiveName := fmt.Sprintf("moai-adk_1.2.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	release := githubRelease{
		TagName:     "v1.2.0",
		PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Assets: []githubAsset{
			{Name: archiveName, BrowserDownloadURL: "https://example.com/moai.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: checksumsTS.URL},
		},
	}

	ts := newTestServer(t, release)
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	info, err := checker.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When checksum download fails, it should be empty string (graceful degradation)
	if info.Checksum != "" {
		t.Errorf("Checksum should be empty when download fails, got %q", info.Checksum)
	}

	// But URL should still be set
	if info.URL == "" {
		t.Error("URL should still be set even when checksum download fails")
	}
}

func TestChecker_DownloadChecksum_ParsesCorrectly(t *testing.T) {
	t.Parallel()

	checksumsContent := `392f7de6e7b21c1e4d1681a424ff254ab51b76fa0a8688e7d76c71226b3f520f  moai-adk_1.2.0_darwin_amd64.tar.gz
99bf529b15df380c912d2be43b475122000ff9d6e1bb9069699b6199f9ef5d90  moai-adk_1.2.0_darwin_arm64.tar.gz
a1423b990826e23bfb830a22bb9f4c42254e12a6ddfc2d958abd53e946e789c2  moai-adk_1.2.0_linux_amd64.tar.gz`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(checksumsContent))
	}))
	defer ts.Close()

	c := NewChecker("http://example.com", http.DefaultClient)
	checksum, err := c.(*checker).downloadChecksum(ts.URL, "moai-adk_1.2.0_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedChecksum := "99bf529b15df380c912d2be43b475122000ff9d6e1bb9069699b6199f9ef5d90"
	if checksum != expectedChecksum {
		t.Errorf("checksum = %q, want %q", checksum, expectedChecksum)
	}
}

func TestChecker_DownloadChecksum_FileNotFound(t *testing.T) {
	t.Parallel()

	checksumsContent := `392f7de6e7b21c1e4d1681a424ff254ab51b76fa0a8688e7d76c71226b3f520f  moai-adk_1.2.0_darwin_amd64.tar.gz
99bf529b15df380c912d2be43b475122000ff9d6e1bb9069699b6199f9ef5d90  moai-adk_1.2.0_darwin_arm64.tar.gz`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(checksumsContent))
	}))
	defer ts.Close()

	c := NewChecker("http://example.com", http.DefaultClient)
	_, err := c.(*checker).downloadChecksum(ts.URL, "moai-adk_1.2.0_linux_amd64.tar.gz")
	if err == nil {
		t.Error("expected error when file not found in checksums")
	}
}

func TestChecker_CheckLatest_ReleasesArray(t *testing.T) {
	t.Parallel()

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	archiveName := fmt.Sprintf("moai-adk_2.0.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	releases := []githubRelease{
		{
			TagName:     "go-v2.0.0",
			PublishedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			Assets: []githubAsset{
				{Name: archiveName, BrowserDownloadURL: "https://example.com/v2.tar.gz"},
			},
		},
		{
			TagName:     "go-v1.5.0",
			PublishedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Assets:      []githubAsset{},
		},
		{
			TagName:     "v3.0.0",
			PublishedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Assets:      []githubAsset{},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(releases)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	// URL ends with /releases (not /latest) to trigger array parsing
	checker := NewChecker(ts.URL+"/releases", http.DefaultClient)
	info, err := checker.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("CheckLatest: %v", err)
	}
	if info.Version != "go-v2.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "go-v2.0.0")
	}
}

func TestChecker_CheckLatest_ReleasesArray_NoGoVReleases(t *testing.T) {
	t.Parallel()

	releases := []githubRelease{
		{TagName: "v3.0.0", PublishedAt: time.Now(), Assets: nil},
		{TagName: "v2.0.0", PublishedAt: time.Now(), Assets: nil},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(releases)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	checker := NewChecker(ts.URL+"/releases", http.DefaultClient)
	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("expected error when no go-v releases found")
	}
}

func TestChecker_CheckLatest_404(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestNewChecker_NilClient(t *testing.T) {
	t.Parallel()

	c := NewChecker("https://example.com", nil)
	if c == nil {
		t.Fatal("NewChecker returned nil with nil client")
	}
}

func TestParseSemverParts_PreRelease(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		v    string
		want [3]int
	}{
		{"standard", "1.2.3", [3]int{1, 2, 3}},
		{"pre-release dash", "1.2.3-beta", [3]int{1, 2, 3}},
		{"pre-release plus", "1.2.3+build123", [3]int{1, 2, 3}},
		{"major only", "5", [3]int{5, 0, 0}},
		{"major.minor", "3.7", [3]int{3, 7, 0}},
		{"empty", "", [3]int{0, 0, 0}},
		{"non-numeric", "abc.def.ghi", [3]int{0, 0, 0}},
		{"mixed", "1.abc.3", [3]int{1, 0, 3}},
		{"patch with pre-release", "2.1.0-alpha.1", [3]int{2, 1, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseSemverParts(tt.v)
			if got != tt.want {
				t.Errorf("parseSemverParts(%q) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestChecker_CheckLatest_ReleasesArray_InvalidJSON(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{not an array}"))
	}))
	defer ts.Close()

	checker := NewChecker(ts.URL+"/releases", http.DefaultClient)
	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("expected error for invalid JSON array")
	}
}

func TestChecker_IsUpdateAvailable_Error(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	checker := NewChecker(ts.URL, http.DefaultClient)
	available, info, err := checker.IsUpdateAvailable("v1.0.0")
	if err == nil {
		t.Error("expected error")
	}
	if available {
		t.Error("expected available=false on error")
	}
	if info != nil {
		t.Error("expected nil info on error")
	}
}
