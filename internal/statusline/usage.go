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
	"sync"
	"time"
)

// UsageProvider는 API 사용량 정보를 수집하는 인터페이스다 (REQ-V3-API-001).
type UsageProvider interface {
	CollectUsage(ctx context.Context) (*UsageResult, error)
}

// usageCacheFile은 캐시 파일의 JSON 구조를 나타낸다 (REQ-V3-API-002).
type usageCacheFile struct {
	CachedAt int64       `json:"cached_at"` // Unix timestamp
	Usage5H  *UsageData  `json:"usage_5h"`
	Usage7D  *UsageData  `json:"usage_7d"`
}

// usageCollector는 Anthropic API 사용량을 수집하고 캐시한다.
// 5분 TTL, 300ms 타임아웃, graceful degradation (REQ-V3-API-002~005, REQ-V3-API-009).
type usageCollector struct {
	mu        sync.RWMutex
	cache     *usageCacheFile
	cachePath string
	ttl       time.Duration
	client    *http.Client
	homeDir   string
}

// NewUsageCollector는 새로운 UsageProvider를 생성한다.
// 캐시는 ~/.moai/cache/usage.json에 저장된다 (REQ-V3-API-002).
func NewUsageCollector(homeDir string) UsageProvider {
	cacheDir := filepath.Join(homeDir, ".moai", "cache")
	cachePath := filepath.Join(cacheDir, "usage.json")

	return &usageCollector{
		cachePath: cachePath,
		ttl:       5 * time.Minute, // REQ-V3-API-002: 5분 TTL
		client: &http.Client{
			Timeout: 300 * time.Millisecond, // REQ-V3-API-003: 300ms 타임아웃
		},
		homeDir: homeDir,
	}
}

// CollectUsage는 5시간/7일 사용량을 반환한다.
// 캐시가 유효하면 캐시를 반환하고, 만료되면 API에서 가져온다 (REQ-V3-API-004~005).
// 모든 실패에서 nil, nil을 반환하여 graceful degradation (REQ-V3-API-009).
func (u *usageCollector) CollectUsage(ctx context.Context) (*UsageResult, error) {
	// 캐시 히트: 신선한 캐시 반환 (REQ-V3-API-004)
	if cached, ok := u.getCachedIfFresh(); ok {
		slog.Debug("usage cache hit")
		return u.toUsageResult(cached), nil
	}

	// 캐시 미스: OAuth 토큰 조회 (REQ-V3-API-010)
	token := readOAuthToken(u.homeDir, u.readTokenFromKeychain)
	if token == "" {
		slog.Debug("oauth token not found, skipping usage collection")
		return nil, nil // Graceful degradation (REQ-V3-API-009)
	}

	// API에서 사용량 가져오기 (REQ-V3-API-005)
	usage5H, err := u.fetchUsageFromAPI(ctx, token, "5h")
	if err != nil {
		slog.Debug("5h usage fetch failed", "error", err)
		// 오류가 있어도 계속 시도 (graceful degradation)
	}

	usage7D, err := u.fetchUsageFromAPI(ctx, token, "7d")
	if err != nil {
		slog.Debug("7d usage fetch failed", "error", err)
		// 오류가 있어도 계속 시도 (graceful degradation)
	}

	// 둘 다 실패하면 nil 반환
	if usage5H == nil && usage7D == nil {
		return nil, nil
	}

	// 캐시 업데이트 (비동기로, 실패해도 계속 진행)
	cache := &usageCacheFile{
		CachedAt: time.Now().Unix(),
		Usage5H:  usage5H,
		Usage7D:  usage7D,
	}
	go func() {
		u.mu.Lock()
		u.cache = cache
		u.mu.Unlock()
		_ = u.saveCache(cache)
	}()

	return u.toUsageResult(cache), nil
}

// getCachedIfFresh는 캐시가 존재하고 TTL 내에 있으면 반환한다 (REQ-V3-API-004).
func (u *usageCollector) getCachedIfFresh() (*usageCacheFile, bool) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	// 메모리 캐시 확인
	if u.cache != nil {
		cachedAt := time.Unix(u.cache.CachedAt, 0)
		if time.Since(cachedAt) < u.ttl {
			return u.cache, true
		}
	}

	// 파일 캐시 로드
	loaded, err := u.loadCache()
	if err != nil {
		return nil, false
	}

	if loaded == nil {
		return nil, false
	}

	// TTL 확인
	cachedAt := time.Unix(loaded.CachedAt, 0)
	if time.Since(cachedAt) >= u.ttl {
		return nil, false
	}

	// 메모리 캐시 업데이트
	u.cache = loaded
	return loaded, true
}

// loadCache는 파일에서 캐시를 로드한다. 파일이 없거나 JSON이 잘못되면 nil, nil을 반환.
func (u *usageCollector) loadCache() (*usageCacheFile, error) {
	data, err := os.ReadFile(u.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 파일 없음은 정상 상황
		}
		return nil, err
	}

	var cache usageCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		slog.Debug("cache file parse failed", "error", err)
		return nil, nil // 잘못된 JSON은 무시
	}

	return &cache, nil
}

// saveCache는 캐시를 파일에 저장한다. atomic write를 사용한다.
func (u *usageCollector) saveCache(cache *usageCacheFile) error {
	// 캐시 디렉토리 생성
	if err := os.MkdirAll(filepath.Dir(u.cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Atomic write: 임시 파일에 쓰고 rename
	tmpPath := u.cachePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tmpPath, u.cachePath); err != nil {
		_ = os.Remove(tmpPath) // cleanup
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// toUsageResult는 캐시 구조를 UsageResult로 변환한다.
func (u *usageCollector) toUsageResult(cache *usageCacheFile) *UsageResult {
	if cache == nil {
		return nil
	}

	// 둘 다 비어있으면 nil 반환
	if cache.Usage5H == nil && cache.Usage7D == nil {
		return nil
	}

	return &UsageResult{
		Usage5H: cache.Usage5H,
		Usage7D: cache.Usage7D,
	}
}

// readOAuthToken은 OAuth 토큰을 조회한다 (REQ-V3-API-010).
// 우선순위: 1) macOS Keychain, 2) ~/.claude/credentials.json.
func readOAuthToken(homeDir string, keychainReader func() (string, error)) string {
	// 1. 키체인에서 시도 (macOS 전용)
	if token, err := keychainReader(); err == nil && token != "" {
		return token
	}

	// 2. credentials.json fallback (REQ-V3-API-010)
	credsPath := filepath.Join(homeDir, ".claude", "credentials.json")
	data, err := os.ReadFile(credsPath)
	if err != nil {
		return ""
	}

	var creds struct {
		OAuthToken string `json:"oauthToken"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return ""
	}

	return creds.OAuthToken
}

// readTokenFromKeychain은 macOS Keychain에서 토큰을 읽는다 (REQ-V3-API-010).
func (u *usageCollector) readTokenFromKeychain() (string, error) {
	// security find-generic-password -s "claude.ai" -w
	cmd := exec.Command("security", "find-generic-password", "-s", "claude.ai", "-w")
	output, err := cmd.Output()
	if err != nil {
		return "", err // 키체인 조회 실패는 정상 상황 (fallback으로 처리)
	}

	token := string(output)
	// 토큰 로그 금지 (REQ-V3-API-008)
	slog.Debug("oauth token read from keychain (redacted)")
	return token, nil
}

// fetchUsageFromAPI는 Anthropic API에서 사용량을 가져온다 (REQ-V3-API-005).
// 타임아웃은 http.Client.Timeout으로 처리 (REQ-V3-API-003).
func (u *usageCollector) fetchUsageFromAPI(ctx context.Context, token, period string) (*UsageData, error) {
	// API endpoint (실제 endpoint는 문서 확인 필요)
	// TODO: 실제 API endpoint 확인 후 업데이트
	url := fmt.Sprintf("https://api.anthropic.com/v1/tokens/usage?period=%s", period)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// 토큰 로그 금지 (REQ-V3-API-008)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp struct {
		Usage []struct {
			TokensUsed int64   `json:"tokens_used"`
			Limit      int64   `json:"limit"`
			Percentage float64 `json:"percentage"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Usage) == 0 {
		return nil, fmt.Errorf("no usage data in response")
	}

	return &UsageData{
		UsedTokens:  apiResp.Usage[0].TokensUsed,
		LimitTokens: apiResp.Usage[0].Limit,
		Percentage:  apiResp.Usage[0].Percentage,
	}, nil
}
