package statusline

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestUsageCache_ReadWrite 캐시 파일을 생성하고 읽기를 검증한다 (REQ-V3-API-002).
func TestUsageCache_ReadWrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "usage.json")

	// 캐시 데이터 생성
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

	// 캜시 파일 저장
	collector := &usageCollector{
		cachePath: cachePath,
		mu:        sync.RWMutex{},
	}
	err := collector.saveCache(cache)
	if err != nil {
		t.Fatalf("saveCache failed: %v", err)
	}

	// 파일 존재 확인
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file was not created")
	}

	// 캐시 파일 로드
	loaded, err := collector.loadCache()
	if err != nil {
		t.Fatalf("loadCache failed: %v", err)
	}

	// 데이터 검증
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

// TestUsageCache_TTL 캐시 만료 검증 (REQ-V3-API-004).
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

	// 만료된 캐시 생성 (6분 전)
	expiredCache := &usageCacheFile{
		CachedAt: time.Now().Add(-6 * time.Minute).Unix(),
		Usage5H:  &UsageData{UsedTokens: 1000},
		Usage7D:  &UsageData{UsedTokens: 2000},
	}
	if err := collector.saveCache(expiredCache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 만료된 캐시는 유효하지 않아야 함
	fresh, ok := collector.getCachedIfFresh()
	if ok {
		t.Error("expired cache should not be fresh, but got fresh cache")
	}
	if fresh != nil {
		t.Error("expired cache should return nil, got non-nil")
	}

	// 신선한 캐시 생성 (1분 전)
	freshCache := &usageCacheFile{
		CachedAt: time.Now().Add(-1 * time.Minute).Unix(),
		Usage5H:  &UsageData{UsedTokens: 3000},
		Usage7D:  &UsageData{UsedTokens: 4000},
	}
	if err := collector.saveCache(freshCache); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 신선한 캐시는 유효해야 함
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

// TestUsageCache_NotExist 파일이 없을 경우 처리 검증.
func TestUsageCache_NotExist(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "nonexistent.json")

	collector := &usageCollector{
		cachePath: cachePath,
		mu:        sync.RWMutex{},
	}

	// 파일이 없으면 nil을 반환해야 함
	loaded, err := collector.loadCache()
	if err != nil {
		t.Fatalf("loadCache should not error on missing file, got: %v", err)
	}
	if loaded != nil {
		t.Error("loadCache should return nil for missing file")
	}
}

// TestUsageCache_InvalidJSON 잘못된 JSON 처리 검증.
func TestUsageCache_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "invalid.json")

	// 잘못된 JSON 작성
	if err := os.WriteFile(cachePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	collector := &usageCollector{
		cachePath: cachePath,
		mu:        sync.RWMutex{},
	}

	// 잘못된 JSON은 nil을 반환해야 함
	loaded, err := collector.loadCache()
	if err != nil {
		t.Fatalf("loadCache should not error on invalid JSON, got: %v", err)
	}
	if loaded != nil {
		t.Error("loadCache should return nil for invalid JSON")
	}
}

// TestReadOAuthToken_Keychain 키체인에서 토큰 읽기 검증 (REQ-V3-API-010).
func TestReadOAuthToken_Keychain(t *testing.T) {
	t.Parallel()

	// 이 테스트는 실제 키체인에 의존하므로 통합 테스트 환경에서만 실행
	// 단위 테스트에서는 security 명령어를 mock 해야 함

	t.Skip("TODO: mock security command for unit testing")
}

// TestReadOAuthToken_Fallback credentials.json fallback 검증 (REQ-V3-API-010).
func TestReadOAuthToken_Fallback(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	homeDir := tmpDir

	// credentials.json 생성
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

	// 키체인은 실패하도록 mock (빈 문자열 반환)
	token := readOAuthToken(homeDir, func() (string, error) {
		return "", nil // 키체인 실패 시뮬레이션
	})

	if token != "test-token-from-credentials" {
		t.Errorf("token = %s, want test-token-from-credentials", token)
	}
}

// TestReadOAuthToken_NotFound 토큰을 찾을 수 없는 경우 검증 (REQ-V3-API-010).
func TestReadOAuthToken_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	homeDir := tmpDir

	// 키체인 실패, credentials.json 없음
	token := readOAuthToken(homeDir, func() (string, error) {
		return "", nil
	})

	if token != "" {
		t.Errorf("token should be empty when not found, got %s", token)
	}
}

// TestFetchUsageFromAPI_Success 성공적인 API 호출 검증 (REQ-V3-API-003, REQ-V3-API-005).
func TestFetchUsageFromAPI_Success(t *testing.T) {
	t.Parallel()

	// 실제 API 호출은 integration test로 분리
	// 단위 테스트에서는 mock HTTP server 사용

	t.Skip("TODO: add mock HTTP server for unit testing")
}

// TestFetchUsageFromAPI_Timeout 300ms 타임아웃 검증 (REQ-V3-API-003).
func TestFetchUsageFromAPI_Timeout(t *testing.T) {
	t.Parallel()

	// 느린 응답을 시뮬레이션하는 mock server 필요

	t.Skip("TODO: add mock HTTP server with delay for unit testing")
}

// TestFetchUsageFromAPI_Error API 오류 처리 검증 (REQ-V3-API-009).
func TestFetchUsageFromAPI_Error(t *testing.T) {
	t.Parallel()

	// 오류 응답을 시뮬레이션하는 mock server 필요

	t.Skip("TODO: add mock HTTP server with error response for unit testing")
}

// TestUsageProvider_CollectUsage_CacheHit 캐시 히트 시나리오 검증 (REQ-V3-API-004).
func TestUsageProvider_CollectUsage_CacheHit(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// NewUsageCollector는 tmpDir/.moai/cache/usage.json를 사용한다
	cachePath := filepath.Join(tmpDir, ".moai", "cache", "usage.json")

	// 신선한 캐시 생성
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

	// 캐시가 유효하면 API 호출 없이 캐시 반환
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

// TestUsageProvider_CollectUsage_APIFallback 캐시 만료 시 API 호출 검증 (REQ-V3-API-005).
func TestUsageProvider_CollectUsage_APIFallback(t *testing.T) {
	t.Parallel()

	// 실제 API 호출이 필요하므로 integration test로 분리

	t.Skip("TODO: add integration test with real API or mock server")
}

// TestUsageProvider_CollectUsage_GracefulDegradation 모든 실패 시 graceful degradation 검증 (REQ-V3-API-009).
func TestUsageProvider_CollectUsage_GracefulDegradation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// 토큰 없음, 캐시 없음

	collector := NewUsageCollector(tmpDir)

	// 실패해도 nil, nil을 반환 (오류 전파 안 함)
	result, err := collector.CollectUsage(context.Background())
	if err != nil {
		t.Fatalf("CollectUsage should not error, got: %v", err)
	}

	if result != nil {
		t.Error("result should be nil when all collection fails")
	}
}
