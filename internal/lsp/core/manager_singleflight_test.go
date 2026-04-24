package core

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// ─── AC-UTIL-003-003: singleflight.Group 필드 존재 확인 ────────────────────

// TestManager_SingleflightField_Present는 Manager.sf 필드가 존재하고
// 즉시 Do 호출이 가능한 zero-value 상태임을 확인한다 (AC-UTIL-003-003).
func TestManager_SingleflightField_Present(t *testing.T) {
	t.Parallel()

	m := NewManager(makeTestServersConfig())

	// 같은 패키지에서 직접 필드 접근 — 필드가 없으면 컴파일 실패
	// zero-value 준비 상태 검증: Do 호출이 정상적으로 동작해야 한다
	called := false
	val, err, shared := m.sf.Do("probe-key", func() (any, error) {
		called = true
		return "result", nil
	})

	if !called {
		t.Error("sf.Do did not invoke the function")
	}
	if err != nil {
		t.Errorf("sf.Do error = %v, want nil", err)
	}
	if val != "result" {
		t.Errorf("sf.Do value = %v, want 'result'", val)
	}
	_ = shared // 공유 여부는 이 테스트에서 검증하지 않음
}

// ─── AC-UTIL-003-004: singleflight 장벽 ──────────────────────────────────

// TestGetOrSpawn_SingleflightBarrier_SecondBlocksUntilFirst는
// 두 번째 동시 getOrSpawn 호출이 첫 번째 factory 반환까지 차단됨을 확인한다 (AC-UTIL-003-004).
func TestGetOrSpawn_SingleflightBarrier_SecondBlocksUntilFirst(t *testing.T) {
	t.Parallel()

	var factoryCount atomic.Int32
	// factory가 차단 해제 신호를 받을 때까지 지연됨
	firstStarted := make(chan struct{})
	unblock := make(chan struct{})

	m := NewManager(
		makeTestServersConfig(),
		WithClientFactory(func(cfg config.ServerConfig) Client {
			factoryCount.Add(1)
			// 첫 번째 factory 진입 신호
			select {
			case firstStarted <- struct{}{}:
			default:
			}
			// 차단 해제 신호 대기
			<-unblock
			return &fakeClient{state: StateSpawning}
		}),
	)

	ctx := context.Background()

	// 첫 번째 고루틴 시작
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = m.getOrSpawn(ctx, "go")
	}()

	// factory가 시작될 때까지 대기
	select {
	case <-firstStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("first factory did not start within 2s")
	}

	// 두 번째 고루틴 시작: factory가 차단 중인 상태에서 getOrSpawn 호출
	done2 := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(done2)
		_, _ = m.getOrSpawn(ctx, "go")
	}()

	// 두 번째 고루틴이 즉시 반환되지 않아야 한다 (sf.Do 차단 중)
	select {
	case <-done2:
		// 두 번째가 즉시 반환되었다 — 빠른 경로(캐시 히트)를 탔을 수 있음.
		// factory 호출 횟수가 1이면 sf.Do를 통한 공유를 의미하므로 허용한다.
		t.Logf("second caller returned early (cache hit path) — factory calls: %d", factoryCount.Load())
	case <-time.After(50 * time.Millisecond):
		// 예상된 동작: 두 번째 고루틴이 sf.Do 내부에서 차단 중
	}

	// factory 차단 해제
	close(unblock)
	wg.Wait()

	// factory는 정확히 1번 호출되어야 한다
	if got := factoryCount.Load(); got != 1 {
		t.Errorf("factory called %d times, want 1 (singleflight should deduplicate)", got)
	}
}

// ─── AC-UTIL-003-005: 16 goroutine 동시 RouteFor, factory 정확히 1회 ───────

// TestRouteFor_ExactlyOnce_16ConcurrentCallers는 16개의 동시 RouteFor 호출에서
// clientFactory가 정확히 1회 호출됨을 확인한다 (AC-UTIL-003-005).
func TestRouteFor_ExactlyOnce_16ConcurrentCallers(t *testing.T) {
	t.Parallel()

	var factoryCount atomic.Int32

	m := NewManager(
		makeTestServersConfig(),
		WithClientFactory(func(cfg config.ServerConfig) Client {
			factoryCount.Add(1)
			// 동시성 시나리오를 더 두드러지게 만들기 위해 짧은 지연
			time.Sleep(5 * time.Millisecond)
			return &fakeClient{state: StateSpawning}
		}),
	)

	const N = 16
	clients := make([]Client, N)
	errs := make([]error, N)

	ctx := context.Background()
	var wg sync.WaitGroup

	// 최대한 동시에 시작하기 위한 시작 게이트
	startGate := make(chan struct{})

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-startGate
			clients[i], errs[i] = m.RouteFor(ctx, "/workspace/main.go")
		}(i)
	}

	// 모든 고루틴이 준비된 후 동시 시작
	close(startGate)
	wg.Wait()

	// factory는 정확히 1회 호출되어야 한다
	if got := factoryCount.Load(); got != 1 {
		t.Errorf("clientFactory called %d times, want exactly 1", got)
	}

	// 모든 호출이 에러 없이 완료되어야 한다
	for i := 0; i < N; i++ {
		if errs[i] != nil {
			t.Errorf("goroutine %d: RouteFor error = %v", i, errs[i])
		}
		if clients[i] == nil {
			t.Errorf("goroutine %d: client is nil", i)
		}
	}

	// 모든 클라이언트가 동일한 포인터여야 한다
	for i := 1; i < N; i++ {
		if clients[i] != clients[0] {
			t.Errorf("goroutine %d: got different client pointer (want same as goroutine 0)", i)
		}
	}
}

// ─── AC-UTIL-003-006: Start 실패 시 캐시 부재 및 재시도 확인 ──────────────

// TestGetOrSpawn_StartError_CacheAbsentAndRetry는 c.Start 실패 시
// 캐시에 클라이언트가 남지 않고 다음 호출에서 factory가 다시 호출됨을 확인한다 (AC-UTIL-003-006).
func TestGetOrSpawn_StartError_CacheAbsentAndRetry(t *testing.T) {
	t.Parallel()

	startErr := errors.New("start: simulated failure")
	var callCount atomic.Int32

	m := NewManager(
		makeTestServersConfig(),
		WithClientFactory(func(cfg config.ServerConfig) Client {
			callCount.Add(1)
			return &fakeClient{state: StateSpawning, startErr: startErr}
		}),
	)

	ctx := context.Background()

	// 첫 번째 호출: Start 실패 → 에러 반환
	c1, err1 := m.getOrSpawn(ctx, "go")
	if err1 == nil {
		t.Fatal("first getOrSpawn: expected error, got nil")
	}
	if c1 != nil {
		t.Error("first getOrSpawn: expected nil client on Start error")
	}
	if !errors.Is(err1, startErr) {
		t.Errorf("first getOrSpawn: error = %v, want to wrap %v", err1, startErr)
	}

	// 캐시에 클라이언트가 없어야 한다
	m.mu.Lock()
	_, cacheHit := m.clients["go"]
	m.mu.Unlock()
	if cacheHit {
		t.Error("cache should be absent after Start error (REQ-UTIL-003-006)")
	}

	// 두 번째 호출: factory가 다시 호출되어야 한다 (sf.Do 키가 해제됨)
	_, _ = m.getOrSpawn(ctx, "go")
	if got := callCount.Load(); got != 2 {
		t.Errorf("factory called %d times total, want 2 (retry on second call)", got)
	}
}

// ─── AC-UTIL-003-011: 경쟁 감지기 — 16 goroutine getOrSpawn ──────────────

// TestGetOrSpawnConcurrent_RaceDetector는 16개 동시 getOrSpawn 호출에서
// 데이터 경쟁이 발생하지 않음을 go test -race로 검증한다 (AC-UTIL-003-011).
//
// 실행 방법: go test -race -run TestGetOrSpawnConcurrent ./internal/lsp/core/
func TestGetOrSpawnConcurrent_RaceDetector(t *testing.T) {
	// t.Parallel()은 -race 단독 실행 시 간섭을 줄이기 위해 의도적으로 생략

	var spawnCount atomic.Int32

	m := NewManager(
		makeTestServersConfig(),
		WithClientFactory(func(cfg config.ServerConfig) Client {
			spawnCount.Add(1)
			// 실제 경쟁 조건을 유발할 수 있도록 약간의 지연
			time.Sleep(2 * time.Millisecond)
			return &fakeClient{state: StateSpawning}
		}),
	)

	const N = 16
	ctx := context.Background()
	var wg sync.WaitGroup

	startGate := make(chan struct{})

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startGate
			// RouteFor는 lastActivity 업데이트도 수행하므로 더 넓은 경쟁 표면을 커버함
			_, _ = m.RouteFor(ctx, "/workspace/main.go")
		}()
	}

	close(startGate)
	wg.Wait()

	// 경쟁 감지기가 이 테스트를 통과하면 데이터 경쟁 없음
	// factory 호출 횟수는 1이어야 함 (singleflight 보장)
	if got := spawnCount.Load(); got != 1 {
		t.Errorf("spawnCount = %d, want 1 (singleflight should prevent duplicates)", got)
	}
}
