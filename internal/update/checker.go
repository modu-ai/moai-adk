package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// releaseResponse represents the GitHub Releases API JSON response.
type releaseResponse struct {
	TagName     string          `json:"tag_name"`
	PublishedAt time.Time       `json:"published_at"`
	Assets      []assetResponse `json:"assets"`
}

// assetResponse represents a single release asset.
type assetResponse struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// checker is the concrete implementation of Checker.
type checker struct {
	apiURL string
	client *http.Client
}

// NewChecker creates a Checker that queries the given API URL.
// The apiURL should be the base URL for the releases endpoint
// (e.g., "https://api.github.com/repos/modu-ai/moai-adk-go/releases/latest").
// For testing, pass the httptest.Server URL directly.
func NewChecker(apiURL string, client *http.Client) Checker {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &checker{
		apiURL: apiURL,
		client: client,
	}
}

// CheckLatest fetches the latest release metadata from GitHub.
func (c *checker) CheckLatest(ctx context.Context) (*VersionInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("checker: create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "moai-adk-go-updater")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checker: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checker: unexpected status %d", resp.StatusCode)
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("checker: decode response: %w", err)
	}

	info := &VersionInfo{
		Version: release.TagName,
		Date:    release.PublishedAt,
	}

	// Find the platform-specific binary URL.
	platformName := fmt.Sprintf("moai-%s-%s", runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if asset.Name == platformName {
			info.URL = asset.BrowserDownloadURL
		}
		if asset.Name == "checksums.txt" {
			info.Checksum = asset.BrowserDownloadURL
		}
	}

	return info, nil
}

// IsUpdateAvailable compares the current version against the latest release.
func (c *checker) IsUpdateAvailable(current string) (bool, *VersionInfo, error) {
	info, err := c.CheckLatest(context.Background())
	if err != nil {
		return false, nil, err
	}

	cmp := compareSemver(info.Version, current)
	if cmp <= 0 {
		// Latest is same or older than current.
		return false, nil, nil
	}

	return true, info, nil
}

// compareSemver compares two semantic version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Handles optional "v" prefix.
func compareSemver(a, b string) int {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	aParts := parseSemverParts(a)
	bParts := parseSemverParts(b)

	for i := 0; i < 3; i++ {
		if aParts[i] > bParts[i] {
			return 1
		}
		if aParts[i] < bParts[i] {
			return -1
		}
	}
	return 0
}

// parseSemverParts extracts [major, minor, patch] from a version string.
func parseSemverParts(v string) [3]int {
	var parts [3]int
	segments := strings.SplitN(v, ".", 3)
	for i, seg := range segments {
		if i >= 3 {
			break
		}
		// Strip any pre-release suffix (e.g., "1-beta").
		if idx := strings.IndexAny(seg, "-+"); idx >= 0 {
			seg = seg[:idx]
		}
		n, err := strconv.Atoi(seg)
		if err == nil {
			parts[i] = n
		}
	}
	return parts
}
