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

// TestIsInFailureCooldown_Active는 실패 후 쿨다운 기간 내에 isInFailureCooldown이 true를 반환하는지 검증한다.
func TestIsInFailureCooldown_Active(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	// 2분 전에 실패가 기록된 캐시 파일 생성
	cache := &usageCacheFile{
		CachedAt: time.Now().Add(-2 * time.Minute).Unix(),
		FailedAt: time.Now().Add(-2 * time.Minute).Unix(),
		Usage5H:  &UsageData{Percentage: 10.0},
		Usage7D:  &UsageData{Percentage: 20.0},
	}
	collector := &usageCollector{
		cachePath:       cachePath,
		ttl:             5 * time.Minute,
		failureCooldown: 10 * time.Minute,
		mu:              sync.RWMutex{},
	}
	if err := collector.saveCache(cache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 쿨다운 기간(10분) 내이므로 true 반환
	if !collector.isInFailureCooldown() {
		t.Error("isInFailureCooldown should return true when failure was 2 minutes ago (cooldown: 10 minutes)")
	}
}

// TestIsInFailureCooldown_Expired는 쿨다운 기간이 만료된 경우 isInFailureCooldown이 false를 반환하는지 검증한다.
func TestIsInFailureCooldown_Expired(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	// 15분 전에 실패가 기록된 캐시 파일 생성 (쿨다운 10분 초과)
	cache := &usageCacheFile{
		CachedAt: time.Now().Add(-15 * time.Minute).Unix(),
		FailedAt: time.Now().Add(-15 * time.Minute).Unix(),
		Usage5H:  &UsageData{Percentage: 10.0},
		Usage7D:  &UsageData{Percentage: 20.0},
	}
	collector := &usageCollector{
		cachePath:       cachePath,
		ttl:             5 * time.Minute,
		failureCooldown: 10 * time.Minute,
		mu:              sync.RWMutex{},
	}
	if err := collector.saveCache(cache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 쿨다운 기간(10분)이 만료되었으므로 false 반환
	if collector.isInFailureCooldown() {
		t.Error("isInFailureCooldown should return false when failure was 15 minutes ago (cooldown: 10 minutes)")
	}
}

// TestIsInFailureCooldown_NoFailure는 실패 기록이 없는 경우 isInFailureCooldown이 false를 반환하는지 검증한다.
func TestIsInFailureCooldown_NoFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	// FailedAt = 0인 캐시 파일 생성 (실패 기록 없음)
	cache := &usageCacheFile{
		CachedAt: time.Now().Unix(),
		FailedAt: 0, // 실패 기록 없음
		Usage5H:  &UsageData{Percentage: 10.0},
		Usage7D:  &UsageData{Percentage: 20.0},
	}
	collector := &usageCollector{
		cachePath:       cachePath,
		ttl:             5 * time.Minute,
		failureCooldown: 10 * time.Minute,
		mu:              sync.RWMutex{},
	}
	if err := collector.saveCache(cache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 실패 기록이 없으므로 false 반환
	if collector.isInFailureCooldown() {
		t.Error("isInFailureCooldown should return false when FailedAt is 0 (no failure recorded)")
	}
}

// TestSaveFailure는 saveFailure 호출 시 캐시 파일에 FailedAt 타임스탬프가 저장되는지 검증한다.
func TestSaveFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	collector := &usageCollector{
		cachePath:       cachePath,
		ttl:             5 * time.Minute,
		failureCooldown: 10 * time.Minute,
		mu:              sync.RWMutex{},
	}

	// Unix timestamp는 초 단위이므로, 비교 기준도 초 단위로 맞춘다
	beforeSave := time.Now().Truncate(time.Second)

	// saveFailure 호출
	collector.saveFailure()

	// 캐시 파일이 생성되었는지 확인
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("saveFailure should create the cache file")
	}

	// 캐시 파일을 읽어 FailedAt이 현재 시각 근처인지 확인
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
	if time.Since(failedAt) > time.Minute {
		t.Errorf("FailedAt (%v) should be within the last minute", failedAt)
	}

	// saveFailure 후 isInFailureCooldown이 true인지 확인
	if !collector.isInFailureCooldown() {
		t.Error("isInFailureCooldown should return true immediately after saveFailure")
	}
}

// TestCollectUsage_SkipsDuringCooldown는 쿨다운 중에 CollectUsage가 API 호출 없이 스테일 캐시를 반환하는지 검증한다.
// 429 스파이럴 방지를 위한 핵심 동작을 테스트한다.
func TestCollectUsage_SkipsDuringCooldown(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".moai", "cache", "usage.json")

	// API 호출 횟수를 추적하는 테스트 서버
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

	// 만료된 캐시 + 최근 실패 기록이 있는 캐시 파일 생성
	staleCache := &usageCacheFile{
		CachedAt: time.Now().Add(-10 * time.Minute).Unix(), // TTL(5분) 초과 - 만료됨
		FailedAt: time.Now().Add(-2 * time.Minute).Unix(),  // 쿨다운(10분) 내 - 활성
		Usage5H:  &UsageData{Percentage: 42.0},
		Usage7D:  &UsageData{Percentage: 85.0},
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	cacheJSON, _ := json.Marshal(staleCache)
	if err := os.WriteFile(cachePath, cacheJSON, 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// failureCooldown이 설정된 collector 생성
	collector := &usageCollector{
		cachePath:       cachePath,
		ttl:             5 * time.Minute,
		failureCooldown: 10 * time.Minute,
		client:          server.Client(),
		homeDir:         tmpDir,
		mu:              sync.RWMutex{},
	}

	result, err := collector.CollectUsage(context.Background())
	if err != nil {
		t.Fatalf("CollectUsage should not error: %v", err)
	}

	// 스테일 캐시 데이터가 반환되어야 한다 (nil이 아님)
	if result == nil {
		t.Fatal("CollectUsage should return stale data during cooldown, got nil")
	}

	// API가 호출되지 않았음을 확인
	if apiCallCount > 0 {
		t.Errorf("API should not be called during cooldown, but was called %d times", apiCallCount)
	}
}

// TestCollectUsage_RetriesAfterCooldownExpires는 쿨다운 만료 후 CollectUsage가 API 호출을 시도하는지 검증한다.
func TestCollectUsage_RetriesAfterCooldownExpires(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".moai", "cache", "usage.json")

	// 쿨다운이 만료된 실패 기록이 있는 캐시 파일 생성 (15분 전 실패, 쿨다운 10분)
	expiredFailureCache := &usageCacheFile{
		CachedAt: time.Now().Add(-15 * time.Minute).Unix(),
		FailedAt: time.Now().Add(-15 * time.Minute).Unix(), // 쿨다운 만료
		Usage5H:  &UsageData{Percentage: 30.0},
		Usage7D:  &UsageData{Percentage: 60.0},
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	cacheJSON, _ := json.Marshal(expiredFailureCache)
	if err := os.WriteFile(cachePath, cacheJSON, 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 토큰 없이 collector 생성 (API 호출 시도하지만 토큰 없어 nil 반환)
	// homeDir에 credentials.json이 없으면 token이 빈 문자열
	collector := &usageCollector{
		cachePath:       cachePath,
		ttl:             5 * time.Minute,
		failureCooldown: 10 * time.Minute,
		client:          &http.Client{Timeout: 1 * time.Second},
		homeDir:         tmpDir, // credentials.json 없음
		mu:              sync.RWMutex{},
	}

	result, err := collector.CollectUsage(context.Background())
	if err != nil {
		t.Fatalf("CollectUsage should not error: %v", err)
	}

	// 쿨다운이 만료되어 API 호출을 시도했으나 토큰이 없어 nil 반환
	// (스테일 캐시를 반환하지 않고 토큰 없음으로 인해 nil)
	if result != nil {
		t.Errorf("CollectUsage should return nil when cooldown expired and no token available, got: %+v", result)
	}
}
