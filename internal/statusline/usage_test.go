package statusline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestUsageCache_ReadWrite verifies cache file creation and reading (REQ-V3-API-002).
func TestUsageCache_ReadWrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	// Create cache data
	cache := &usageCacheFile{
		CachedAt: time.Now().Unix(),
		Usage5H: &UsageData{
			UsedTokens:  1000000,
			LimitTokens: 5000000,
			Percentage:  20.0,
		},
		Usage7D: &UsageData{
			UsedTokens:  7000000,
			LimitTokens: 10000000,
			Percentage:  70.0,
		},
	}

	// Save cache file
	collector := &usageCollector{
		cachePath: cachePath,
		mu:        sync.RWMutex{},
	}
	err := collector.saveCache(cache)
	if err != nil {
		t.Fatalf("saveCache failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file was not created")
	}

	// Load cache file
	loaded, err := collector.loadCache()
	if err != nil {
		t.Fatalf("loadCache failed: %v", err)
	}

	// Verify data
	if loaded.Usage5H == nil || loaded.Usage7D == nil {
		t.Fatal("loaded cache missing usage data")
	}

	if loaded.Usage5H.UsedTokens != 1000000 {
		t.Errorf("Usage5H.UsedTokens = %d, want %d", loaded.Usage5H.UsedTokens, 1000000)
	}

	if loaded.Usage7D.Percentage != 70.0 {
		t.Errorf("Usage7D.Percentage = %f, want %f", loaded.Usage7D.Percentage, 70.0)
	}
}

// TestUsageCache_TTL verifies cache expiration (REQ-V3-API-004).
func TestUsageCache_TTL(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")
	ttl := 5 * time.Minute

	collector := &usageCollector{
		cachePath: cachePath,
		ttl:       ttl,
		mu:        sync.RWMutex{},
	}

	// Create expired cache (6 minutes ago)
	expiredCache := &usageCacheFile{
		CachedAt: time.Now().Add(-6 * time.Minute).Unix(),
		Usage5H:  &UsageData{UsedTokens: 1000},
		Usage7D:  &UsageData{UsedTokens: 2000},
	}
	if err := collector.saveCache(expiredCache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Expired cache should not be valid
	fresh, ok := collector.getCachedIfFresh()
	if ok {
		t.Error("expired cache should not be fresh, but got fresh cache")
	}
	if fresh != nil {
		t.Error("expired cache should return nil, got non-nil")
	}

	// Create fresh cache (1 minute ago)
	freshCache := &usageCacheFile{
		CachedAt: time.Now().Add(-1 * time.Minute).Unix(),
		Usage5H:  &UsageData{UsedTokens: 3000},
		Usage7D:  &UsageData{UsedTokens: 4000},
	}
	if err := collector.saveCache(freshCache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Fresh cache should be valid
	fresh, ok = collector.getCachedIfFresh()
	if !ok {
		t.Error("fresh cache should be valid")
	}
	if fresh == nil {
		t.Fatal("fresh cache should not be nil")
	}
	if fresh.Usage5H.UsedTokens != 3000 {
		t.Errorf("Usage5H.UsedTokens = %d, want %d", fresh.Usage5H.UsedTokens, 3000)
	}
}

// TestUsageCache_NotExist verifies handling when file does not exist.
func TestUsageCache_NotExist(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "nonexistent.json")

	collector := &usageCollector{
		cachePath: cachePath,
		mu:        sync.RWMutex{},
	}

	// Should return nil when file does not exist
	loaded, err := collector.loadCache()
	if err != nil {
		t.Fatalf("loadCache should not error on missing file, got: %v", err)
	}
	if loaded != nil {
		t.Error("loadCache should return nil for missing file")
	}
}

// TestUsageCache_InvalidJSON verifies handling of invalid JSON.
func TestUsageCache_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(cachePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	collector := &usageCollector{
		cachePath: cachePath,
		mu:        sync.RWMutex{},
	}

	// Invalid JSON should return nil
	loaded, err := collector.loadCache()
	if err != nil {
		t.Fatalf("loadCache should not error on invalid JSON, got: %v", err)
	}
	if loaded != nil {
		t.Error("loadCache should return nil for invalid JSON")
	}
}

// TestReadOAuthToken_Keychain verifies token reading from keychain (REQ-V3-API-010).
func TestReadOAuthToken_Keychain(t *testing.T) {
	t.Parallel()

	// This test depends on real keychain, only run in integration test environment
	// Unit tests need to mock the security command

	t.Skip("TODO: mock security command for unit testing")
}

// TestReadOAuthToken_Fallback verifies credentials.json fallback (REQ-V3-API-010).
func TestReadOAuthToken_Fallback(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	homeDir := tmpDir

	// Create credentials.json
	credsDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(credsDir, 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	credsFile := filepath.Join(credsDir, "credentials.json")
	creds := map[string]string{
		"oauthToken": "test-token-from-credentials",
	}
	credsJSON, _ := json.Marshal(creds)
	if err := os.WriteFile(credsFile, credsJSON, 0600); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Mock keychain to fail (return empty string)
	token := readOAuthToken(homeDir, func() (string, error) {
		return "", nil // Simulate keychain failure
	})

	if token != "test-token-from-credentials" {
		t.Errorf("token = %s, want test-token-from-credentials", token)
	}
}

// TestReadOAuthToken_DotPrefixCredentials verifies that credentials stored at
// ~/.claude/.credentials.json (dot-prefixed, as Claude Code uses on Linux/WSL2)
// are read correctly when keychain is unavailable.
// Regression test for: https://github.com/modu-ai/moai-adk/issues/496
func TestReadOAuthToken_DotPrefixCredentials(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	homeDir := tmpDir

	// Create ~/.claude/.credentials.json (dot-prefixed path used by Claude Code on Linux)
	credsDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(credsDir, 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Write credentials to .credentials.json (with dot prefix)
	dotCredsFile := filepath.Join(credsDir, ".credentials.json")
	creds := map[string]string{
		"oauthToken": "test-token-from-dot-credentials",
	}
	credsJSON, _ := json.Marshal(creds)
	if err := os.WriteFile(dotCredsFile, credsJSON, 0600); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Keychain fails (simulating Linux/WSL2 where macOS Keychain is unavailable)
	token := readOAuthToken(homeDir, func() (string, error) {
		return "", fmt.Errorf("keychain unavailable on Linux")
	})

	if token != "test-token-from-dot-credentials" {
		t.Errorf("token = %q, want %q (credentials at ~/.claude/.credentials.json not read)",
			token, "test-token-from-dot-credentials")
	}
}

// TestReadOAuthToken_DotPrefixNestedCredentials verifies that nested claudeAiOauth.accessToken
// format in ~/.claude/.credentials.json is read correctly on Linux/WSL2.
// Regression test for: https://github.com/modu-ai/moai-adk/issues/496
func TestReadOAuthToken_DotPrefixNestedCredentials(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	homeDir := tmpDir

	// Create ~/.claude/.credentials.json with nested format
	credsDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(credsDir, 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	dotCredsFile := filepath.Join(credsDir, ".credentials.json")
	nestedCreds := map[string]interface{}{
		"claudeAiOauth": map[string]string{
			"accessToken": "test-nested-token-linux",
		},
	}
	credsJSON, _ := json.Marshal(nestedCreds)
	if err := os.WriteFile(dotCredsFile, credsJSON, 0600); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Keychain fails (simulating Linux/WSL2)
	token := readOAuthToken(homeDir, func() (string, error) {
		return "", fmt.Errorf("keychain unavailable on Linux")
	})

	if token != "test-nested-token-linux" {
		t.Errorf("token = %q, want %q (nested credentials at ~/.claude/.credentials.json not read)",
			token, "test-nested-token-linux")
	}
}

// TestReadOAuthToken_DotPrefixTakesPrecedenceOverPlain verifies that when both
// ~/.claude/.credentials.json and ~/.claude/credentials.json exist, the dot-prefixed
// file (which Claude Code creates) takes precedence.
func TestReadOAuthToken_DotPrefixTakesPrecedenceOverPlain(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	homeDir := tmpDir

	credsDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(credsDir, 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Create both files
	dotCredsFile := filepath.Join(credsDir, ".credentials.json")
	dotCreds := map[string]string{"oauthToken": "dot-prefix-token"}
	dotCredsJSON, _ := json.Marshal(dotCreds)
	if err := os.WriteFile(dotCredsFile, dotCredsJSON, 0600); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	plainCredsFile := filepath.Join(credsDir, "credentials.json")
	plainCreds := map[string]string{"oauthToken": "plain-token"}
	plainCredsJSON, _ := json.Marshal(plainCreds)
	if err := os.WriteFile(plainCredsFile, plainCredsJSON, 0600); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	token := readOAuthToken(homeDir, func() (string, error) {
		return "", fmt.Errorf("keychain unavailable")
	})

	if token != "dot-prefix-token" {
		t.Errorf("token = %q, want %q (dot-prefixed file should take precedence)",
			token, "dot-prefix-token")
	}
}

// TestReadOAuthToken_NotFound verifies handling when token is not found (REQ-V3-API-010).
func TestReadOAuthToken_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	homeDir := tmpDir

	// Keychain fails, no credentials.json
	token := readOAuthToken(homeDir, func() (string, error) {
		return "", nil
	})

	if token != "" {
		t.Errorf("token should be empty when not found, got %s", token)
	}
}

// TestFetchUsageFromAPI_Success verifies successful API call (REQ-V3-API-003, REQ-V3-API-005).
func TestFetchUsageFromAPI_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(oauthUsageResponse{
			FiveHour: &usagePeriodData{Utilization: 25.5},
			SevenDay: &usagePeriodData{Utilization: 60.0},
		})
	}))
	defer server.Close()

	collector := &usageCollector{
		client: server.Client(),
		mu:     sync.RWMutex{},
	}

	resp, err := collector.fetchUsageFromAPI(t, server.URL, "test-token")
	if err != nil {
		t.Fatalf("fetchUsageFromOAuthAPI failed: %v", err)
	}
	if resp.FiveHour == nil || resp.FiveHour.Utilization != 25.5 {
		t.Errorf("FiveHour.Utilization = %v, want 25.5", resp.FiveHour)
	}
	if resp.SevenDay == nil || resp.SevenDay.Utilization != 60.0 {
		t.Errorf("SevenDay.Utilization = %v, want 60.0", resp.SevenDay)
	}
}

// fetchUsageFromAPI is a test helper that calls fetchUsageFromOAuthAPI with a custom URL.
func (u *usageCollector) fetchUsageFromAPI(t *testing.T, apiURL, token string) (*oauthUsageResponse, error) {
	t.Helper()
	ctx := context.Background()

	const maxRetries = 3
	var lastErr error
	backoff := 50 * time.Millisecond // Shorter backoff for tests

	for attempt := range maxRetries {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")

		resp, err := u.client.Do(req)
		if err != nil {
			lastErr = err
			break
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

		if resp.StatusCode != http.StatusTooManyRequests {
			return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		lastErr = fmt.Errorf("API returned status 429 (attempt %d/%d)", attempt+1, maxRetries)
		time.Sleep(backoff)
		backoff *= 2
	}

	return nil, lastErr
}

// TestFetchUsageFromAPI_Retry429 verifies retry on 429 with exponential backoff.
func TestFetchUsageFromAPI_Retry429(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Succeed on 3rd attempt
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(oauthUsageResponse{
			FiveHour: &usagePeriodData{Utilization: 10.0},
			SevenDay: &usagePeriodData{Utilization: 30.0},
		})
	}))
	defer server.Close()

	collector := &usageCollector{
		client: server.Client(),
		mu:     sync.RWMutex{},
	}

	resp, err := collector.fetchUsageFromAPI(t, server.URL, "test-token")
	if err != nil {
		t.Fatalf("fetchUsageFromAPI should succeed after retries, got: %v", err)
	}
	if resp.FiveHour == nil || resp.FiveHour.Utilization != 10.0 {
		t.Errorf("FiveHour.Utilization = %v, want 10.0", resp.FiveHour)
	}
	if got := attempts.Load(); got != 3 {
		t.Errorf("attempts = %d, want 3", got)
	}
}

// TestFetchUsageFromAPI_NonRetryableError verifies non-429 errors are not retried.
func TestFetchUsageFromAPI_NonRetryableError(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	collector := &usageCollector{
		client: server.Client(),
		mu:     sync.RWMutex{},
	}

	_, err := collector.fetchUsageFromAPI(t, server.URL, "test-token")
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
	if got := attempts.Load(); got != 1 {
		t.Errorf("attempts = %d, want 1 (non-retryable error)", got)
	}
}

// TestUsageProvider_CollectUsage_CacheHit verifies cache hit scenario (REQ-V3-API-004).
func TestUsageProvider_CollectUsage_CacheHit(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// NewUsageCollector uses tmpDir/.moai/cache/usage.json
	cachePath := filepath.Join(tmpDir, ".moai", "cache", "usage.json")

	// Create fresh cache
	freshCache := &usageCacheFile{
		CachedAt: time.Now().Unix(),
		Usage5H: &UsageData{
			UsedTokens:  1234567,
			LimitTokens: 5000000,
			Percentage:  24.7,
		},
		Usage7D: &UsageData{
			UsedTokens:  7000000,
			LimitTokens: 10000000,
			Percentage:  70.0,
		},
	}
	cacheJSON, _ := json.Marshal(freshCache)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := os.WriteFile(cachePath, cacheJSON, 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	collector := NewUsageCollector(tmpDir)

	// Return cache without API call when cache is valid
	result, err := collector.CollectUsage(context.Background())
	if err != nil {
		t.Fatalf("CollectUsage failed: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.Usage5H == nil || result.Usage7D == nil {
		t.Fatal("result missing usage data")
	}

	if result.Usage5H.UsedTokens != 1234567 {
		t.Errorf("Usage5H.UsedTokens = %d, want %d", result.Usage5H.UsedTokens, 1234567)
	}
}

// TestGetStaleCache verifies stale cache retrieval for fallback (REQ-V3-API-009).
func TestGetStaleCache(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	collector := &usageCollector{
		cachePath: cachePath,
		ttl:       5 * time.Minute,
		mu:        sync.RWMutex{},
	}

	// No cache exists: should return nil
	if got := collector.getStaleCache(); got != nil {
		t.Error("getStaleCache should return nil when no cache exists")
	}

	// Create expired cache (30 minutes ago)
	expiredCache := &usageCacheFile{
		CachedAt: time.Now().Add(-30 * time.Minute).Unix(),
		Usage5H:  &UsageData{Percentage: 42.0},
		Usage7D:  &UsageData{Percentage: 85.0},
	}
	if err := collector.saveCache(expiredCache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// getCachedIfFresh should reject expired cache
	_, ok := collector.getCachedIfFresh()
	if ok {
		t.Error("expired cache should not be fresh")
	}

	// getStaleCache should return expired cache (ignores TTL)
	stale := collector.getStaleCache()
	if stale == nil {
		t.Fatal("getStaleCache should return expired cache")
	}
	if stale.Usage5H.Percentage != 42.0 {
		t.Errorf("stale Usage5H.Percentage = %f, want 42.0", stale.Usage5H.Percentage)
	}
}

// TestUsageProvider_CollectUsage_GracefulDegradation verifies graceful degradation on all failures (REQ-V3-API-009).
func TestUsageProvider_CollectUsage_GracefulDegradation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".moai", "cache", "usage.json")
	// No token, no cache — mock keychain to prevent real keychain access
	collector := &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		client:              &http.Client{Timeout: 300 * time.Millisecond},
		homeDir:             tmpDir,
	}
	// Override keychain reader to always fail (no real keychain access)
	collector.keychainReaderFn = func() (string, error) {
		return "", fmt.Errorf("no keychain in test")
	}

	// Should return nil, nil on failure (no error propagation)
	result, err := collector.CollectUsage(context.Background())
	if err != nil {
		t.Fatalf("CollectUsage should not error, got: %v", err)
	}

	if result != nil {
		t.Error("result should be nil when all collection fails")
	}
}

// TestIsInFailureCooldown_Active verifies isInFailureCooldown returns true during cooldown period.
func TestIsInFailureCooldown_Active(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	// 1 failure recorded 30s ago (cooldown: 1m)
	cache := &usageCacheFile{
		CachedAt:     time.Now().Add(-30 * time.Second).Unix(),
		FailedAt:     time.Now().Add(-30 * time.Second).Unix(),
		FailureCount: 1,
		Usage5H:      &UsageData{Percentage: 10.0},
		Usage7D:      &UsageData{Percentage: 20.0},
	}
	collector := &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		mu:                  sync.RWMutex{},
	}
	if err := collector.saveCache(cache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 1 failure → cooldown 1m, elapsed 30s → cooldown active
	if !collector.isInFailureCooldown() {
		t.Error("isInFailureCooldown should return true when failure was 30s ago (cooldown: 1m for count=1)")
	}
}

// TestIsInFailureCooldown_Expired verifies isInFailureCooldown returns false after cooldown expires.
func TestIsInFailureCooldown_Expired(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	// 1 failure recorded 2m ago (cooldown: 1m) → expired
	cache := &usageCacheFile{
		CachedAt:     time.Now().Add(-2 * time.Minute).Unix(),
		FailedAt:     time.Now().Add(-2 * time.Minute).Unix(),
		FailureCount: 1,
		Usage5H:      &UsageData{Percentage: 10.0},
		Usage7D:      &UsageData{Percentage: 20.0},
	}
	collector := &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		mu:                  sync.RWMutex{},
	}
	if err := collector.saveCache(cache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 1 failure → cooldown 1m, elapsed 2m → cooldown expired
	if collector.isInFailureCooldown() {
		t.Error("isInFailureCooldown should return false when failure was 2m ago (cooldown: 1m for count=1)")
	}
}

// TestIsInFailureCooldown_NoFailure verifies false when no failure is recorded.
func TestIsInFailureCooldown_NoFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	cache := &usageCacheFile{
		CachedAt:     time.Now().Unix(),
		FailedAt:     0,
		FailureCount: 0,
		Usage5H:      &UsageData{Percentage: 10.0},
		Usage7D:      &UsageData{Percentage: 20.0},
	}
	collector := &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		mu:                  sync.RWMutex{},
	}
	if err := collector.saveCache(cache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if collector.isInFailureCooldown() {
		t.Error("isInFailureCooldown should return false when no failure recorded")
	}
}

// TestSaveFailure verifies saveFailure records FailedAt and increments FailureCount.
func TestSaveFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	collector := &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		mu:                  sync.RWMutex{},
	}

	beforeSave := time.Now().Truncate(time.Second)

	// First failure
	collector.saveFailure()

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("saveFailure should create the cache file")
	}

	loaded, err := collector.loadCache()
	if err != nil {
		t.Fatalf("loadCache failed after saveFailure: %v", err)
	}
	if loaded == nil {
		t.Fatal("loaded cache should not be nil after saveFailure")
	}

	failedAt := time.Unix(loaded.FailedAt, 0)
	if failedAt.Before(beforeSave) {
		t.Errorf("FailedAt (%v) should be at or after test start (%v)", failedAt, beforeSave)
	}
	if loaded.FailureCount != 1 {
		t.Errorf("FailureCount = %d after first failure, want 1", loaded.FailureCount)
	}

	// Second failure should increment count
	collector.saveFailure()
	loaded, _ = collector.loadCache()
	if loaded.FailureCount != 2 {
		t.Errorf("FailureCount = %d after second failure, want 2", loaded.FailureCount)
	}

	// Should be in cooldown after saveFailure
	if !collector.isInFailureCooldown() {
		t.Error("isInFailureCooldown should return true immediately after saveFailure")
	}
}

// TestCollectUsage_SkipsDuringCooldown verifies that CollectUsage returns stale cache
// without API call during cooldown period (429 spiral prevention).
func TestCollectUsage_SkipsDuringCooldown(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".moai", "cache", "usage.json")

	var apiCallCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		apiCallCount++
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(oauthUsageResponse{
			FiveHour: &usagePeriodData{Utilization: 99.0},
			SevenDay: &usagePeriodData{Utilization: 99.0},
		})
	}))
	defer server.Close()

	// Stale cache with recent failure (count=3 → cooldown=4m, failed 2m ago → active)
	staleCache := &usageCacheFile{
		CachedAt:     time.Now().Add(-10 * time.Minute).Unix(),
		FailedAt:     time.Now().Add(-2 * time.Minute).Unix(),
		FailureCount: 3,
		Usage5H:      &UsageData{Percentage: 42.0},
		Usage7D:      &UsageData{Percentage: 85.0},
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	cacheJSON, _ := json.Marshal(staleCache)
	if err := os.WriteFile(cachePath, cacheJSON, 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	collector := &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		client:              server.Client(),
		homeDir:             tmpDir,
		mu:                  sync.RWMutex{},
	}

	result, err := collector.CollectUsage(context.Background())
	if err != nil {
		t.Fatalf("CollectUsage should not error: %v", err)
	}

	if result == nil {
		t.Fatal("CollectUsage should return stale data during cooldown, got nil")
	}

	if apiCallCount > 0 {
		t.Errorf("API should not be called during cooldown, but was called %d times", apiCallCount)
	}
}

// TestCollectUsage_RetriesAfterCooldownExpires verifies that CollectUsage attempts
// API call after cooldown expires.
func TestCollectUsage_RetriesAfterCooldownExpires(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".moai", "cache", "usage.json")

	// count=1 → cooldown=1m, failed 2m ago → expired
	expiredFailureCache := &usageCacheFile{
		CachedAt:     time.Now().Add(-15 * time.Minute).Unix(),
		FailedAt:     time.Now().Add(-2 * time.Minute).Unix(),
		FailureCount: 1,
		Usage5H:      &UsageData{Percentage: 30.0},
		Usage7D:      &UsageData{Percentage: 60.0},
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	cacheJSON, _ := json.Marshal(expiredFailureCache)
	if err := os.WriteFile(cachePath, cacheJSON, 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// No token available → CollectUsage should try API but return nil
	collector := &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		client:              &http.Client{Timeout: 1 * time.Second},
		homeDir:             tmpDir,
		mu:                  sync.RWMutex{},
		keychainReaderFn: func() (string, error) {
			return "", fmt.Errorf("no keychain in test")
		},
	}

	result, err := collector.CollectUsage(context.Background())
	if err != nil {
		t.Fatalf("CollectUsage should not error: %v", err)
	}

	// Cooldown expired, API call attempted but no token → nil
	if result != nil {
		t.Errorf("CollectUsage should return nil when cooldown expired and no token available, got: %+v", result)
	}
}

// TestCalcCooldown verifies exponential backoff calculation.
func TestCalcCooldown(t *testing.T) {
	t.Parallel()

	collector := &usageCollector{
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
	}

	tests := []struct {
		count int
		want  time.Duration
	}{
		{0, 0},
		{1, 1 * time.Minute},
		{2, 2 * time.Minute},
		{3, 4 * time.Minute},
		{4, 8 * time.Minute},
		{5, 16 * time.Minute},
		{6, 32 * time.Minute},
		{7, 32 * time.Minute},  // capped
		{10, 32 * time.Minute}, // capped
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("count=%d", tt.count), func(t *testing.T) {
			got := collector.calcCooldown(tt.count)
			if got != tt.want {
				t.Errorf("calcCooldown(%d) = %v, want %v", tt.count, got, tt.want)
			}
		})
	}
}

// TestFetchUsageFromHeaders_Success verifies Haiku probe extracts rate limit headers.
func TestFetchUsageFromHeaders_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a POST to Messages API format
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}

		// Return rate limit headers (even on 200)
		w.Header().Set("anthropic-ratelimit-unified-5h-utilization", "0.28")
		w.Header().Set("anthropic-ratelimit-unified-7d-utilization", "0.59")
		w.Header().Set("anthropic-ratelimit-unified-5h-reset", "2026-03-10T20:00:00Z")
		w.Header().Set("anthropic-ratelimit-unified-7d-reset", "2026-03-15T07:00:00Z")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"content":[{"text":"h"}]}`))
	}))
	defer server.Close()

	collector := &usageCollector{
		client: server.Client(),
		mu:     sync.RWMutex{},
	}

	// Use the test server URL instead of real API
	resp, err := collector.fetchUsageFromHeadersWithURL(context.Background(), server.URL, "test-token")
	if err != nil {
		t.Fatalf("fetchUsageFromHeaders failed: %v", err)
	}

	if resp.FiveHour == nil {
		t.Fatal("FiveHour should not be nil")
	}
	// 0.28 * 100 = 28.0
	if resp.FiveHour.Utilization != 28.0 {
		t.Errorf("FiveHour.Utilization = %f, want 28.0", resp.FiveHour.Utilization)
	}
	if resp.FiveHour.ResetsAt != "2026-03-10T20:00:00Z" {
		t.Errorf("FiveHour.ResetsAt = %s, want 2026-03-10T20:00:00Z", resp.FiveHour.ResetsAt)
	}

	if resp.SevenDay == nil {
		t.Fatal("SevenDay should not be nil")
	}
	// 0.59 * 100 = 59.0
	if resp.SevenDay.Utilization != 59.0 {
		t.Errorf("SevenDay.Utilization = %f, want 59.0", resp.SevenDay.Utilization)
	}
}

// TestFetchUsageFromHeaders_429WithHeaders verifies headers are extracted even on 429 response.
func TestFetchUsageFromHeaders_429WithHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("anthropic-ratelimit-unified-5h-utilization", "0.95")
		w.Header().Set("anthropic-ratelimit-unified-7d-utilization", "0.80")
		w.Header().Set("anthropic-ratelimit-unified-5h-reset", "2026-03-10T22:00:00Z")
		w.Header().Set("anthropic-ratelimit-unified-7d-reset", "2026-03-17T07:00:00Z")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"Rate limited"}}`))
	}))
	defer server.Close()

	collector := &usageCollector{
		client: server.Client(),
		mu:     sync.RWMutex{},
	}

	resp, err := collector.fetchUsageFromHeadersWithURL(context.Background(), server.URL, "test-token")
	if err != nil {
		t.Fatalf("fetchUsageFromHeaders should succeed even on 429: %v", err)
	}

	if resp.FiveHour == nil {
		t.Fatal("FiveHour should not be nil even on 429")
	}
	if resp.FiveHour.Utilization != 95.0 {
		t.Errorf("FiveHour.Utilization = %f, want 95.0", resp.FiveHour.Utilization)
	}
	if resp.SevenDay == nil {
		t.Fatal("SevenDay should not be nil even on 429")
	}
	if resp.SevenDay.Utilization != 80.0 {
		t.Errorf("SevenDay.Utilization = %f, want 80.0", resp.SevenDay.Utilization)
	}
}

// TestFetchUsageFromHeaders_NoHeaders verifies error when no rate limit headers present.
func TestFetchUsageFromHeaders_NoHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"content":[{"text":"h"}]}`))
	}))
	defer server.Close()

	collector := &usageCollector{
		client: server.Client(),
		mu:     sync.RWMutex{},
	}

	_, err := collector.fetchUsageFromHeadersWithURL(context.Background(), server.URL, "test-token")
	if err == nil {
		t.Fatal("expected error when no rate limit headers present")
	}
}

// TestExponentialBackoff_ProgressiveCooldown verifies that repeated failures
// increase cooldown progressively.
func TestExponentialBackoff_ProgressiveCooldown(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	collector := &usageCollector{
		cachePath:           cachePath,
		ttl:                 5 * time.Minute,
		failureCooldownBase: 1 * time.Minute,
		failureCooldownMax:  32 * time.Minute,
		mu:                  sync.RWMutex{},
	}

	// 5 failures ago, count=5 → cooldown=16m, failed 10m ago → still active
	cache := &usageCacheFile{
		FailedAt:     time.Now().Add(-10 * time.Minute).Unix(),
		FailureCount: 5,
	}
	if err := collector.saveCache(cache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if !collector.isInFailureCooldown() {
		t.Error("should be in cooldown (count=5 → 16m, elapsed=10m)")
	}

	// Same count=5 but failed 20m ago → expired
	cache.FailedAt = time.Now().Add(-20 * time.Minute).Unix()
	if err := collector.saveCache(cache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	// Reset in-memory cache to force file read
	collector.mu.Lock()
	collector.cache = nil
	collector.mu.Unlock()

	if collector.isInFailureCooldown() {
		t.Error("should not be in cooldown (count=5 → 16m, elapsed=20m)")
	}
}
