package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// ErrNoLanguageDetected는 파일 확장자가 설정된 어떤 언어 서버와도 매핑되지 않을 때
// RouteFor가 반환하는 센티넬 에러입니다.
//
// @MX:ANCHOR: [AUTO] ErrNoLanguageDetected — Manager.RouteFor의 공개 센티넬 에러
// @MX:REASON: fan_in >= 3 — RouteFor, 테스트 어서션, Ralph 엔진, MCP 브리지 등 여러 호출자가 이 센티넬로 분기함
var ErrNoLanguageDetected = errors.New("lsp: no language server configured for this file type")

// Manager는 여러 언어 서버 Client 인스턴스를 조율하는 타입입니다.
// 파일 확장자 + 프로젝트 마커 기반으로 요청을 적절한 Client에 라우팅합니다 (REQ-LC-008).
//
// @MX:ANCHOR: [AUTO] Manager — 멀티 언어 LSP 클라이언트 조율자 (REQ-LC-008, REQ-LC-009, REQ-LC-050)
// @MX:REASON: fan_in >= 3 — Ralph 엔진, Quality Gates, LOOP 커맨드, MCP 브리지가 모두 Manager를 통해 LSP에 접근함
type Manager struct {
	// servers는 언어 이름(예: "go") → ServerConfig 맵입니다.
	servers map[string]config.ServerConfig

	// clients는 언어 이름 → Client 맵입니다. lazy-spawned됩니다 (T-016).
	clients map[string]Client

	// mu는 servers, clients, lastActivity 접근을 보호합니다.
	mu sync.Mutex

	// clientFactory는 테스트에서 DI로 교체 가능한 Client 생성 함수입니다.
	// 기본값: NewClient
	clientFactory func(config.ServerConfig) Client

	// lastActivity는 언어별 마지막 RouteFor 호출 시각을 추적합니다 (REQ-LC-050).
	lastActivity map[string]time.Time

	// idleShutdownSeconds는 클라이언트 유휴 종료 기준 초 수입니다.
	// 0이면 즉시 만료로 처리됩니다.
	idleShutdownSeconds int

	// reaperInterval은 reaper 고루틴의 tick 간격입니다 (테스트에서 WithReaperInterval로 단축 가능).
	reaperInterval time.Duration

	// logger는 상태 전환 및 생명주기 이벤트 로거입니다.
	logger *slog.Logger

	// ctx와 cancel은 Manager 내부 고루틴(reaper)의 생명주기를 제어합니다.
	ctx    context.Context
	cancel context.CancelFunc

	// reaperDone은 reaper 고루틴이 완료되었음을 알리는 채널입니다.
	reaperDone chan struct{}

	// statFn은 파일 존재 여부 확인 함수입니다 (테스트에서 DI 가능).
	// 기본값: os.Stat
	statFn func(string) (os.FileInfo, error)
}

// ManagerOption은 Manager를 설정하는 함수형 옵션입니다.
type ManagerOption func(*Manager)

// WithClientFactory는 테스트에서 DI로 fake Client를 주입하기 위한 옵션입니다.
func WithClientFactory(f func(config.ServerConfig) Client) ManagerOption {
	return func(m *Manager) {
		m.clientFactory = f
	}
}

// WithIdleShutdownSeconds는 Manager의 유휴 종료 타임아웃을 설정합니다.
// 기본값은 REQ-LC-050 기준 600초입니다.
func WithIdleShutdownSeconds(seconds int) ManagerOption {
	return func(m *Manager) {
		m.idleShutdownSeconds = seconds
	}
}

// WithManagerLogger는 Manager의 로거를 설정합니다.
func WithManagerLogger(logger *slog.Logger) ManagerOption {
	return func(m *Manager) {
		m.logger = logger
	}
}

// WithReaperInterval은 reaper 고루틴의 tick 간격을 설정합니다.
// 테스트에서 짧은 간격으로 빠른 idle 감지를 테스트할 때 사용합니다.
// 기본값: 30초.
func WithReaperInterval(d time.Duration) ManagerOption {
	return func(m *Manager) {
		m.reaperInterval = d
	}
}

// defaultReaperInterval은 reaper 기본 tick 간격입니다.
const defaultReaperInterval = 30 * time.Second

// defaultIdleShutdownSeconds는 REQ-LC-050 기준 기본 유휴 종료 타임아웃입니다.
const defaultIdleShutdownSeconds = 600

// NewManager는 ServersConfig를 기반으로 새 Manager를 생성합니다.
// Start 호출 전까지 reaper 고루틴은 시작되지 않습니다.
//
// @MX:ANCHOR: [AUTO] NewManager — Manager 생성 팩토리, Manager 사용의 유일한 진입점
// @MX:REASON: fan_in >= 3 — CLI, 통합 테스트, Ralph 엔진, MCP 브리지가 모두 NewManager를 호출함
func NewManager(cfg *config.ServersConfig, opts ...ManagerOption) *Manager {
	m := &Manager{
		servers:             make(map[string]config.ServerConfig),
		clients:             make(map[string]Client),
		lastActivity:        make(map[string]time.Time),
		idleShutdownSeconds: defaultIdleShutdownSeconds,
		reaperInterval:      defaultReaperInterval,
		logger:              slog.Default(),
	}

	// ServersConfig에서 servers 맵 초기화
	if cfg != nil {
		for lang, sc := range cfg.Servers {
			// Language 필드가 비어 있으면 맵 키로 채움
			if sc.Language == "" {
				sc.Language = lang
			}
			m.servers[lang] = sc
		}
	}

	// 기본 clientFactory: NewClient 사용
	m.clientFactory = func(sc config.ServerConfig) Client {
		return NewClient(sc)
	}
	// 기본 statFn: os.Stat 사용
	m.statFn = os.Stat

	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Start는 Manager의 내부 컨텍스트를 초기화하고 reaper 고루틴을 시작합니다.
// 이 메서드는 Manager를 사용하기 전에 반드시 호출해야 합니다.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		// 이미 Start된 경우 no-op
		return nil
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.reaperDone = make(chan struct{})

	// @MX:WARN: [AUTO] reaper 고루틴 — Manager의 생명주기 동안 유휴 클라이언트를 정리하는 장기 실행 백그라운드 고루틴
	// @MX:REASON: 컨텍스트 취소 없이는 고루틴 누수가 발생할 수 있음. Shutdown에서 반드시 ctx를 취소하고 reaperDone을 기다려야 함.
	go m.reaper(m.ctx, m.reaperDone)
	return nil
}

// Shutdown은 reaper 고루틴을 취소하고 모든 캐시된 Client를 병렬로 종료합니다.
// 클라이언트 종료 에러는 errors.Join으로 집계됩니다 (REQ-LC-050).
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
	}
	reaperDone := m.reaperDone
	m.mu.Unlock()

	// reaper 고루틴이 완료될 때까지 대기 (ctx 타임아웃 존중)
	if reaperDone != nil {
		select {
		case <-reaperDone:
		case <-ctx.Done():
			// ctx 만료 시에도 계속 진행 (best-effort)
		}
	}

	// 모든 클라이언트를 병렬로 종료
	m.mu.Lock()
	snapshot := make(map[string]Client, len(m.clients))
	for lang, c := range m.clients {
		snapshot[lang] = c
	}
	m.clients = make(map[string]Client)
	m.mu.Unlock()

	var shutdownErrs []error
	var wg sync.WaitGroup
	var errsMu sync.Mutex

	for lang, c := range snapshot {
		wg.Add(1)
		go func(language string, cl Client) {
			defer wg.Done()
			if err := cl.Shutdown(ctx); err != nil {
				errsMu.Lock()
				shutdownErrs = append(shutdownErrs, fmt.Errorf("language %s: %w", language, err))
				errsMu.Unlock()
			}
		}(lang, c)
	}
	wg.Wait()

	if len(shutdownErrs) > 0 {
		return errors.Join(shutdownErrs...)
	}
	return nil
}

// detectLanguage는 파일 경로의 확장자를 설정된 언어 서버와 매핑합니다.
//
// 매핑 전략:
//  1. 파일 확장자가 정확히 1개 언어에 매핑되면 즉시 반환
//  2. 여러 언어가 같은 확장자를 가지면 프로젝트 마커로 구분:
//     - 파일 디렉토리부터 루트까지 올라가며 각 언어의 RootMarkers 파일 존재 여부 확인
//     - 정확히 1개 언어만 마커가 있으면 해당 언어 반환
//     - 마커로 구분 불가시 후보 언어 중 결정론적으로 첫 번째 반환
//  3. 확장자 매핑 없으면 (empty, false) 반환
func (m *Manager) detectLanguage(path string) (string, bool) {
	ext := filepath.Ext(path)
	if ext == "" {
		return "", false
	}

	// 확장자로 후보 언어 수집
	// 결정론적 순서를 위해 언어 이름 정렬
	var candidates []string
	for lang, sc := range m.servers {
		for _, fe := range sc.FileExtensions {
			if fe == ext {
				candidates = append(candidates, lang)
				break
			}
		}
	}

	if len(candidates) == 0 {
		return "", false
	}
	if len(candidates) == 1 {
		return candidates[0], true
	}

	// 결정론적 순서 보장을 위해 정렬
	sort.Strings(candidates)

	// 프로젝트 마커로 구분
	dir := filepath.Dir(path)
	matched := findLanguageByMarkers(dir, candidates, m.servers, m.statFn)
	if matched != "" {
		return matched, true
	}

	// 마커로 구분 불가: 첫 번째 후보 반환 (결정론적)
	return candidates[0], true
}

// findLanguageByMarkers는 파일 디렉토리에서 루트까지 올라가며 각 후보 언어의
// RootMarker 파일이 존재하는지 확인합니다.
// 정확히 1개 언어만 마커가 있으면 해당 언어를, 없으면 빈 문자열을 반환합니다.
func findLanguageByMarkers(dir string, candidates []string, servers map[string]config.ServerConfig, stat func(string) (os.FileInfo, error)) string {
	// 경로를 따라 올라가며 마커 탐색
	for {
		var matchedLangs []string
		for _, lang := range candidates {
			sc := servers[lang]
			for _, marker := range sc.RootMarkers {
				markerPath := filepath.Join(dir, marker)
				if _, statErr := stat(markerPath); statErr == nil {
					matchedLangs = append(matchedLangs, lang)
					break
				}
			}
		}

		if len(matchedLangs) == 1 {
			return matchedLangs[0]
		}

		// 부모 디렉토리로 이동
		parent := filepath.Dir(dir)
		if parent == dir {
			// 파일시스템 루트에 도달
			break
		}
		dir = parent
	}
	return ""
}

// RouteFor returns the LSP Client responsible for the given file path.
//
// 동작:
//  1. detectLanguage로 언어 결정
//  2. getOrSpawn으로 Client 획득 (lazy spawn)
//  3. lastActivity 업데이트
//
// @MX:ANCHOR: [AUTO] Manager.RouteFor — 파일 경로 기반 LSP 클라이언트 라우팅 핵심 경로
// @MX:REASON: fan_in >= 3 — Ralph 엔진, Quality Gates, LOOP 커맨드, Aggregator가 모두 RouteFor를 통해 Client를 획득함
func (m *Manager) RouteFor(ctx context.Context, path string) (Client, error) {
	lang, ok := m.detectLanguage(path)
	if !ok {
		return nil, fmt.Errorf("lsp manager: %w (path=%q)", ErrNoLanguageDetected, path)
	}

	c, err := m.getOrSpawn(ctx, lang)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.lastActivity[lang] = time.Now()
	m.mu.Unlock()

	return c, nil
}

// getOrSpawn은 캐시된 Client를 반환하거나, 없으면 새로 생성합니다.
//
// 동시 안전성: mu 락 하에 clients 맵을 확인하고,
// StateShutdown이 아닌 클라이언트가 있으면 재사용합니다.
// 없으면 clientFactory로 새 Client를 생성하고 Start를 호출합니다.
// Start 실패 시 캐시에서 제거합니다.
func (m *Manager) getOrSpawn(ctx context.Context, language string) (Client, error) {
	m.mu.Lock()
	existing, ok := m.clients[language]
	if ok && existing.State() != StateShutdown {
		m.mu.Unlock()
		return existing, nil
	}

	sc, hasCfg := m.servers[language]
	if !hasCfg {
		m.mu.Unlock()
		return nil, fmt.Errorf("lsp manager: no server config for language %q", language)
	}

	// 새 클라이언트 생성 및 캐시 등록 (Start 전)
	c := m.clientFactory(sc)
	m.clients[language] = c
	m.mu.Unlock()

	// Start 호출 (락 외부에서 실행 — 블로킹 가능)
	if err := c.Start(ctx); err != nil {
		// Start 실패: 캐시에서 제거하여 다음 호출 시 재시도 가능하게 함
		m.mu.Lock()
		// 다른 고루틴이 같은 클라이언트를 교체하지 않은 경우만 삭제
		if cur, still := m.clients[language]; still && cur == c {
			delete(m.clients, language)
		}
		m.mu.Unlock()
		return nil, fmt.Errorf("lsp manager: failed to start client for language %q: %w", language, err)
	}

	return c, nil
}

// reaper는 일정 간격으로 유휴 클라이언트를 종료하는 백그라운드 고루틴입니다.
//
// 동작:
//   - reaperInterval마다 각 언어의 lastActivity를 확인
//   - time.Since(lastActivity) > idleShutdownSeconds이면 graceful Shutdown
//   - ctx.Done() 수신 시 reaperDone을 닫고 종료
func (m *Manager) reaper(ctx context.Context, done chan struct{}) {
	defer close(done)

	ticker := time.NewTicker(m.reaperInterval)
	defer ticker.Stop()

	// shutdownTimeout: 개별 클라이언트 graceful shutdown 시 사용
	const shutdownTimeout = 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.reapIdleClients(ctx, shutdownTimeout)
		}
	}
}

// reapIdleClients는 유휴 타임아웃을 초과한 클라이언트를 찾아 종료합니다.
func (m *Manager) reapIdleClients(ctx context.Context, shutdownTimeout time.Duration) {
	now := time.Now()

	m.mu.Lock()
	var toShutdown []struct {
		lang   string
		client Client
	}

	for lang, c := range m.clients {
		activity, hasActivity := m.lastActivity[lang]
		if !hasActivity {
			continue
		}

		elapsed := now.Sub(activity)
		threshold := time.Duration(m.idleShutdownSeconds) * time.Second

		if elapsed > threshold {
			toShutdown = append(toShutdown, struct {
				lang   string
				client Client
			}{lang: lang, client: c})
			delete(m.clients, lang)
			delete(m.lastActivity, lang)
		}
	}
	m.mu.Unlock()

	for _, item := range toShutdown {
		shutCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		if err := item.client.Shutdown(shutCtx); err != nil {
			m.logger.Info("lsp manager: idle client shutdown error",
				slog.String("language", item.lang),
				slog.String("error", err.Error()),
			)
		} else {
			m.logger.Info("lsp manager: idle client shut down",
				slog.String("language", item.lang),
			)
		}
		cancel()
	}
}
