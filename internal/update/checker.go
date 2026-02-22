package update

import (
	"bufio"
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
// (e.g., "https://api.github.com/repos/modu-ai/moai-adk/releases/latest").
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
// If the API URL ends with /releases (not /latest), it returns all releases
// and filters for the appropriate version (e.g., "go-v" prefix for Go versions).
func (c *checker) CheckLatest(ctx context.Context) (*VersionInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("checker: create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "moai-adk-updater")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checker: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("checker: release not found (status 404) - repository may not exist or have no releases")
		}
		return nil, fmt.Errorf("checker: unexpected status %d", resp.StatusCode)
	}

	// Check if the response is an array (releases list) or single object (latest release)
	isArrayResponse := strings.HasSuffix(c.apiURL, "/releases") && !strings.HasSuffix(c.apiURL, "/releases/latest")

	if isArrayResponse {
		// Parse as array and filter for go-v prefix tags
		var releases []releaseResponse
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, fmt.Errorf("checker: decode releases array: %w", err)
		}

		// Filter for go-v prefix tags (e.g., go-v2.0.0)
		var filteredReleases []releaseResponse
		for _, r := range releases {
			if strings.HasPrefix(r.TagName, "go-v") {
				filteredReleases = append(filteredReleases, r)
			}
		}

		if len(filteredReleases) == 0 {
			return nil, fmt.Errorf("checker: no go-v releases found in repository")
		}

		// Get the first (latest) filtered release
		release := filteredReleases[0]
		return c.buildVersionInfo(release), nil
	}

	// Single release response
	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("checker: decode response: %w", err)
	}

	return c.buildVersionInfo(release), nil
}

// buildVersionInfo constructs a VersionInfo from a releaseResponse.
func (c *checker) buildVersionInfo(release releaseResponse) *VersionInfo {
	info := &VersionInfo{
		Version: release.TagName,
		Date:    release.PublishedAt,
	}

	// Find the platform-specific archive URL matching goreleaser format.
	// Archive format: moai-adk_<version>_<os>_<arch>.<ext>
	// Example: moai-adk_2.0.0_darwin_amd64.tar.gz
	// Note: GoReleaser's {{ .Version }} strips "v" prefix, so we must too
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	// Strip "v" and "go-v" prefixes from tag name to match GoReleaser's {{ .Version }}
	version := strings.TrimPrefix(release.TagName, "go-v")
	version = strings.TrimPrefix(version, "v")
	archiveName := fmt.Sprintf("moai-adk_%s_%s_%s.%s", version, runtime.GOOS, runtime.GOARCH, ext)

	var checksumsURL string
	for _, asset := range release.Assets {
		if asset.Name == archiveName {
			info.URL = asset.BrowserDownloadURL
		}
		if asset.Name == "checksums.txt" {
			checksumsURL = asset.BrowserDownloadURL
		}
	}

	// Download and parse checksums.txt to extract the checksum for this platform
	if checksumsURL != "" {
		checksum, err := c.downloadChecksum(checksumsURL, archiveName)
		if err == nil && checksum != "" {
			info.Checksum = checksum
		}
		// If checksum download fails, continue without checksum verification
		// (better to allow update with warning than to block entirely)
	}

	return info
}

// downloadChecksum downloads and parses the checksums.txt file to extract
// the checksum for the specified archive filename.
func (c *checker) downloadChecksum(checksumsURL, archiveName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumsURL, nil)
	if err != nil {
		return "", fmt.Errorf("create checksum request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch checksums: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums status %d", resp.StatusCode)
	}

	// Parse checksums.txt line by line
	// Format: <checksum>  <filename>
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Split by whitespace (checksum and filename)
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		checksum := parts[0]
		filename := parts[1]

		// Check if this line matches our archive name
		if filename == archiveName {
			return checksum, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan checksums: %w", err)
	}

	return "", fmt.Errorf("checksum not found for %s", archiveName)
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
// Handles optional "go-v" and "v" prefixes.
func compareSemver(a, b string) int {
	// Handle go-v prefix for Go edition releases
	a = strings.TrimPrefix(a, "go-v")
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "go-v")
	b = strings.TrimPrefix(b, "v")

	aParts := parseSemverParts(a)
	bParts := parseSemverParts(b)

	for i := range 3 {
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
