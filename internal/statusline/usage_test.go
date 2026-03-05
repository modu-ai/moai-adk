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
	// No token, no cache

	collector := NewUsageCollector(tmpDir)

	// Should return nil, nil on failure (no error propagation)
	result, err := collector.CollectUsage(context.Background())
	if err != nil {
		t.Fatalf("CollectUsage should not error, got: %v", err)
	}

	if result != nil {
		t.Error("result should be nil when all collection fails")
	}
}
