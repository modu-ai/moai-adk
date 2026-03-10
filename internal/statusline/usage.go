package statusline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UsageProvider collects API usage information (REQ-V3-API-001).
type UsageProvider interface {
	CollectUsage(ctx context.Context) (*UsageResult, error)
}

// usageCacheFile represents the JSON structure of the cache file (REQ-V3-API-002).
type usageCacheFile struct {
	CachedAt     int64      `json:"cached_at"`                // Unix timestamp
	FailedAt     int64      `json:"failed_at,omitempty"`      // Last API failure time (Unix timestamp)
	FailureCount int        `json:"failure_count,omitempty"`  // Consecutive failure count (for exponential backoff)
	Usage5H      *UsageData `json:"usage_5h"`
	Usage7D      *UsageData `json:"usage_7d"`
}

// usageCollector collects and caches Anthropic API usage.
// 5-minute TTL, 300ms timeout, graceful degradation (REQ-V3-API-002~005, REQ-V3-API-009).
type usageCollector struct {
	mu                 sync.RWMutex
	cache              *usageCacheFile
	cachePath          string
	ttl                time.Duration
	failureCooldownBase time.Duration // Exponential backoff base cooldown (default: 1m)
	failureCooldownMax  time.Duration // Exponential backoff max cooldown (default: 32m)
	client             *http.Client
	homeDir            string
}

// NewUsageCollector creates a new UsageProvider.
// Cache is stored at ~/.moai/cache/usage.json (REQ-V3-API-002).
func NewUsageCollector(homeDir string) UsageProvider {
	cacheDir := filepath.Join(homeDir, ".moai", "cache")
	cachePath := filepath.Join(cacheDir, "usage.json")

	return &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		homeDir: homeDir,
	}
}

// CollectUsage returns 5-hour/7-day usage data.
// Returns cached data if fresh, otherwise fetches from API (REQ-V3-API-004~005).
// Returns nil, nil on any failure for graceful degradation (REQ-V3-API-009).
func (u *usageCollector) CollectUsage(ctx context.Context) (*UsageResult, error) {
	// Cache hit: return fresh cache (REQ-V3-API-004)
	if cached, ok := u.getCachedIfFresh(); ok {
		slog.Debug("usage cache hit")
		return u.toUsageResult(cached), nil
	}

	// Skip API call during failure cooldown (prevents 429 spiral)
	if u.isInFailureCooldown() {
		slog.Debug("usage API in failure cooldown, skipping")
		if stale := u.getStaleCache(); stale != nil {
			return u.toUsageResult(stale), nil
		}
		return nil, nil
	}

	// Cache miss: retrieve OAuth token (REQ-V3-API-010)
	token := readOAuthToken(u.homeDir, u.readTokenFromKeychain)
	if token == "" {
		slog.Debug("oauth token not found, skipping usage collection")
		return nil, nil
	}

	// Fetch 5H/7D usage from OAuth API in a single call (REQ-V3-API-005)
	apiResp, err := u.fetchUsageFromOAuthAPI(ctx, token)
	if err != nil {
		slog.Debug("oauth usage fetch failed", "error", err)
		u.saveFailure() // Record failure to prevent 429 spiral
		// Return stale cache if available (stale-while-revalidate pattern)
		if stale := u.getStaleCache(); stale != nil {
			slog.Debug("returning stale usage cache")
			return u.toUsageResult(stale), nil
		}
		return nil, nil
	}

	// Convert API response to internal format
	var usage5H, usage7D *UsageData
	if apiResp.FiveHour != nil {
		usage5H = &UsageData{
			Percentage: apiResp.FiveHour.Utilization,
			ResetsAt:   apiResp.FiveHour.ResetsAt,
		}
	}
	if apiResp.SevenDay != nil {
		usage7D = &UsageData{
			Percentage: apiResp.SevenDay.Utilization,
			ResetsAt:   apiResp.SevenDay.ResetsAt,
		}
	}

	if usage5H == nil && usage7D == nil {
		return nil, nil
	}

	// Update cache (async, continue on failure)
	cache := &usageCacheFile{
		CachedAt:     time.Now().Unix(),
		FailedAt:     0, // Reset failure state on success
		FailureCount: 0, // Reset consecutive failure count on success
		Usage5H:      usage5H,
		Usage7D:      usage7D,
	}
	go func() {
		u.mu.Lock()
		u.cache = cache
		u.mu.Unlock()
		_ = u.saveCache(cache)
	}()

	return u.toUsageResult(cache), nil
}

// calcCooldown calculates exponential backoff cooldown based on consecutive failure count.
// cooldown = min(base * 2^(count-1), max)
// count=1: 1m, count=2: 2m, count=3: 4m, count=4: 8m, count=5: 16m, count>=6: 32m
func (u *usageCollector) calcCooldown(failureCount int) time.Duration {
	if failureCount <= 0 {
		return 0
	}
	cooldown := u.failureCooldownBase
	for i := 1; i < failureCount; i++ {
		cooldown *= 2
		if cooldown >= u.failureCooldownMax {
			return u.failureCooldownMax
		}
	}
	return cooldown
}

// isInFailureCooldown checks if currently within exponential backoff cooldown period.
// Checks in-memory cache first, then falls back to file cache.
func (u *usageCollector) isInFailureCooldown() bool {
	u.mu.RLock()
	if u.cache != nil && u.cache.FailedAt > 0 && u.cache.FailureCount > 0 {
		failedAt := time.Unix(u.cache.FailedAt, 0)
		cooldown := u.calcCooldown(u.cache.FailureCount)
		if time.Since(failedAt) < cooldown {
			u.mu.RUnlock()
			return true
		}
	}
	u.mu.RUnlock()

	// Check file cache
	loaded, err := u.loadCache()
	if err != nil || loaded == nil {
		return false
	}
	if loaded.FailedAt > 0 && loaded.FailureCount > 0 {
		failedAt := time.Unix(loaded.FailedAt, 0)
		cooldown := u.calcCooldown(loaded.FailureCount)
		return time.Since(failedAt) < cooldown
	}
	return false
}

// saveFailure records failure timestamp and increments consecutive failure count.
// Preserves existing usage data (stale data can still be displayed).
func (u *usageCollector) saveFailure() {
	u.mu.Lock()
	if u.cache == nil {
		u.cache = &usageCacheFile{}
	}
	u.cache.FailedAt = time.Now().Unix()
	u.cache.FailureCount++
	cache := u.cache
	u.mu.Unlock()
	_ = u.saveCache(cache)
}

// getCachedIfFresh returns cached data if it exists and is within TTL (REQ-V3-API-004).
func (u *usageCollector) getCachedIfFresh() (*usageCacheFile, bool) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	// Check in-memory cache
	if u.cache != nil {
		cachedAt := time.Unix(u.cache.CachedAt, 0)
		if time.Since(cachedAt) < u.ttl {
			return u.cache, true
		}
	}

	// Load file cache
	loaded, err := u.loadCache()
	if err != nil {
		return nil, false
	}

	if loaded == nil {
		return nil, false
	}

	// Check TTL
	cachedAt := time.Unix(loaded.CachedAt, 0)
	if time.Since(cachedAt) >= u.ttl {
		return nil, false
	}

	// Update in-memory cache
	u.cache = loaded
	return loaded, true
}

// getStaleCache returns expired cache data for stale-while-revalidate pattern.
// Unlike getCachedIfFresh, this ignores TTL and returns any available cache.
func (u *usageCollector) getStaleCache() *usageCacheFile {
	u.mu.RLock()
	if u.cache != nil {
		defer u.mu.RUnlock()
		return u.cache
	}
	u.mu.RUnlock()

	loaded, err := u.loadCache()
	if err != nil || loaded == nil {
		return nil
	}
	return loaded
}

// loadCache loads cache from file. Returns nil, nil if file doesn't exist or JSON is invalid.
func (u *usageCollector) loadCache() (*usageCacheFile, error) {
	data, err := os.ReadFile(u.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cache usageCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		slog.Debug("cache file parse failed", "error", err)
		return nil, nil
	}

	return &cache, nil
}

// saveCache writes cache to file using atomic write (temp file + rename).
func (u *usageCollector) saveCache(cache *usageCacheFile) error {
	if err := os.MkdirAll(filepath.Dir(u.cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	tmpPath := u.cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tmpPath, u.cachePath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// toUsageResult converts cache structure to UsageResult.
func (u *usageCollector) toUsageResult(cache *usageCacheFile) *UsageResult {
	if cache == nil {
		return nil
	}

	if cache.Usage5H == nil && cache.Usage7D == nil {
		return nil
	}

	return &UsageResult{
		Usage5H: cache.Usage5H,
		Usage7D: cache.Usage7D,
	}
}

// readOAuthToken retrieves the OAuth token (REQ-V3-API-010).
// Priority: 1) macOS Keychain ("Claude Code-credentials"), 2) ~/.claude/credentials.json.
func readOAuthToken(homeDir string, keychainReader func() (string, error)) string {
	// Try keychain first (macOS only)
	if token, err := keychainReader(); err == nil && token != "" {
		return token
	}

	// Fallback: credentials.json (REQ-V3-API-010)
	credsPath := filepath.Join(homeDir, ".claude", "credentials.json")
	data, err := os.ReadFile(credsPath)
	if err != nil {
		return ""
	}

	// Try top-level oauthToken
	var simpleCreds struct {
		OAuthToken string `json:"oauthToken"`
	}
	if err := json.Unmarshal(data, &simpleCreds); err == nil && simpleCreds.OAuthToken != "" {
		return simpleCreds.OAuthToken
	}

	// Try nested claudeAiOauth.accessToken
	var nestedCreds struct {
		ClaudeAiOauth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(data, &nestedCreds); err == nil && nestedCreds.ClaudeAiOauth.AccessToken != "" {
		return nestedCreds.ClaudeAiOauth.AccessToken
	}

	return ""
}

// readTokenFromKeychain reads the OAuth token from macOS Keychain (REQ-V3-API-010).
// Reads JSON from "Claude Code-credentials" service and extracts claudeAiOauth.accessToken.
func (u *usageCollector) readTokenFromKeychain() (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Keychain value is JSON: {"claudeAiOauth":{"accessToken":"...",...}}
	var keychainData struct {
		ClaudeAiOauth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(output, &keychainData); err != nil {
		// Try plain text token if not JSON
		token := string(output)
		if len(token) > 0 && token[len(token)-1] == '\n' {
			token = token[:len(token)-1]
		}
		if token != "" {
			slog.Debug("oauth token read from keychain as plain text (redacted)")
			return token, nil
		}
		return "", err
	}

	if keychainData.ClaudeAiOauth.AccessToken != "" {
		slog.Debug("oauth token read from keychain (redacted)")
		return keychainData.ClaudeAiOauth.AccessToken, nil
	}

	return "", fmt.Errorf("no accessToken found in keychain data")
}

// oauthUsageResponse represents the Anthropic OAuth usage API response.
// Endpoint: GET https://api.anthropic.com/api/oauth/usage
// Response: {"five_hour":{"utilization":6.0,"resets_at":"..."},"seven_day":{"utilization":35.0,"resets_at":"..."}}
type oauthUsageResponse struct {
	FiveHour *usagePeriodData `json:"five_hour"`
	SevenDay *usagePeriodData `json:"seven_day"`
}

// usagePeriodData represents a single usage period (5H or 7D).
type usagePeriodData struct {
	Utilization float64 `json:"utilization"` // Usage percentage (0-100)
	ResetsAt    string  `json:"resets_at"`   // ISO 8601 reset timestamp
}

// fetchUsageFromHeadersWithURL makes a minimal POST request to apiURL and extracts
// usage from Anthropic rate limit response headers (REQ-V3-API-011).
// Works for both 200 and 429 responses; returns error only if no headers are present.
func (u *usageCollector) fetchUsageFromHeadersWithURL(ctx context.Context, apiURL string, token string) (*oauthUsageResponse, error) {
	// Minimal Haiku probe body to trigger rate limit headers
	probeBody := `{"model":"claude-haiku-4-5-20251001","max_tokens":1,"messages":[{"role":"user","content":"h"}]}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(probeBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Extract rate limit headers (available on both 200 and 429 responses)
	var result oauthUsageResponse

	if util5h := resp.Header.Get("anthropic-ratelimit-unified-5h-utilization"); util5h != "" {
		val, parseErr := strconv.ParseFloat(util5h, 64)
		if parseErr == nil {
			result.FiveHour = &usagePeriodData{
				Utilization: val * 100, // Convert 0-1 to 0-100
				ResetsAt:    resp.Header.Get("anthropic-ratelimit-unified-5h-reset"),
			}
		}
	}

	if util7d := resp.Header.Get("anthropic-ratelimit-unified-7d-utilization"); util7d != "" {
		val, parseErr := strconv.ParseFloat(util7d, 64)
		if parseErr == nil {
			result.SevenDay = &usagePeriodData{
				Utilization: val * 100, // Convert 0-1 to 0-100
				ResetsAt:    resp.Header.Get("anthropic-ratelimit-unified-7d-reset"),
			}
		}
	}

	if result.FiveHour == nil && result.SevenDay == nil {
		return nil, fmt.Errorf("no rate limit headers found in response")
	}

	return &result, nil
}

// fetchUsageFromOAuthAPI fetches 5H/7D usage from the Anthropic OAuth API with retry (REQ-V3-API-005).
// Endpoint: GET https://api.anthropic.com/api/oauth/usage
// Retries up to 3 times with exponential backoff on 429 (rate limit).
// Timeout is handled by http.Client.Timeout (REQ-V3-API-003).
func (u *usageCollector) fetchUsageFromOAuthAPI(ctx context.Context, token string) (*oauthUsageResponse, error) {
	const apiURL = "https://api.anthropic.com/api/oauth/usage"
	const maxRetries = 3

	var lastErr error
	backoff := 2 * time.Second

	for attempt := range maxRetries {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			return nil, err
		}

		// Token must not be logged (REQ-V3-API-008)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("anthropic-beta", "oauth-2025-04-20")

		resp, err := u.client.Do(req)
		if err != nil {
			lastErr = err
			break // network error, no retry
		}

		if resp.StatusCode == http.StatusOK {
			var apiResp oauthUsageResponse
			err = json.NewDecoder(resp.Body).Decode(&apiResp)
			_ = resp.Body.Close()
			if err != nil {
				return nil, err
			}
			return &apiResp, nil
		}

		_ = resp.Body.Close()

		// Only retry on 429 (rate limit)
		if resp.StatusCode != http.StatusTooManyRequests {
			return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		lastErr = fmt.Errorf("API returned status 429 (attempt %d/%d)", attempt+1, maxRetries)
		slog.Debug("usage API rate limited, retrying", "attempt", attempt+1, "backoff", backoff)

		// Respect Retry-After header if present
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if secs, parseErr := time.ParseDuration(retryAfter + "s"); parseErr == nil && secs > 0 {
				backoff = secs
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}

		backoff *= 2 // exponential backoff
	}

	return nil, lastErr
}
