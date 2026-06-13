package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// SPEC-SEC-HARDEN-001 §M5 — Circuit breaker half-open single-permit invariant +
// recovered OnStateChange goroutine.
//
// reproduction-first 계약:
//   - AC-SEC-M5-001 (RED): half-open 상태에서 N개 동시 Call이 모두 fn()을 실행함
//     (단일 permit invariant 부재 → 픽스 전 concurrent fn() > 1).
//   - AC-SEC-M5-002 (GREEN): 픽스 후 정확히 1개만 fn() 실행, 나머지는 ErrCircuitOpen.
//   - AC-SEC-M5-003 (GREEN): trial 해소 후 permit 해제 → 다음 Call이 새 state로 governed.
//   - AC-SEC-M5-004 (OBSERVATIONAL RED): panic하는 OnStateChange가 unrecovered goroutine에서
//     테스트 프로세스를 crash시킴 → 자동 테스트로 commit 불가 (suite abort). 수동 1회 관측,
//     progress.md 기록. 본 파일에는 commit하지 않는다.
//   - AC-SEC-M5-005 (GREEN, AUTOMATED): 픽스 후 panic하는 OnStateChange가 goroutine 내부에서
//     recover되어 프로세스 생존 + post-transition state 정상.
//   - AC-SEC-M5-006 (NO-REG): closed/open threshold, timeout 승격, half-open success/failure
//     전이, metrics, Reset() 동기 callback 경로 불변.

// promoteToHalfOpen opens the breaker (threshold failures) then waits out the
// timeout so the next State()/Call() promotes it to half-open.
func promoteToHalfOpen(t *testing.T, cb *CircuitBreaker, threshold int, timeout time.Duration) {
	t.Helper()
	ctx := context.Background()
	opErr := errors.New("op failed")
	for range threshold {
		_ = cb.Call(ctx, func() error { return opErr })
	}
	if cb.State() != StateOpen {
		t.Fatalf("setup: expected open after %d failures, got %v", threshold, cb.State())
	}
	time.Sleep(timeout + 20*time.Millisecond)
}

// TestCircuitBreaker_HalfOpenSinglePermit 는 AC-SEC-M5-001 (RED) + M5-002 (GREEN) 다.
// half-open 상태에서 N개 goroutine이 blocking fn()으로 동시 Call하면:
//   - 픽스 전: 1개 초과가 fn()에 동시 진입 (concurrentMax > 1) → FAIL.
//   - 픽스 후: 정확히 1개만 fn() 실행, 나머지 N-1은 ErrCircuitOpen.
func TestCircuitBreaker_HalfOpenSinglePermit(t *testing.T) {
	threshold := 2
	timeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{Threshold: threshold, Timeout: timeout})

	// 첫 Call(State() 호출 포함)이 half-open으로 승격시킨다.
	promoteToHalfOpen(t, cb, threshold, timeout)

	const n = 8
	var (
		inFlight      atomic.Int32 // 현재 fn() 내부에 있는 goroutine 수
		concurrentMax atomic.Int32 // 관측된 최대 동시 fn() 수
		executed      atomic.Int32 // fn()을 실제 실행한 goroutine 수
		rejected      atomic.Int32 // ErrCircuitOpen을 받은 goroutine 수
		release       = make(chan struct{})
	)

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			err := cb.Call(ctx, func() error {
				executed.Add(1)
				cur := inFlight.Add(1)
				// 관측된 최대 동시 진입 수 갱신.
				for {
					m := concurrentMax.Load()
					if cur <= m || concurrentMax.CompareAndSwap(m, cur) {
						break
					}
				}
				<-release // barrier: 진입자가 동시에 머무르게 한다 (다중 진입 관측 창).
				inFlight.Add(-1)
				return nil
			})
			if errors.Is(err, ErrCircuitOpen) {
				rejected.Add(1)
			}
		}()
	}

	// 적어도 1개가 fn()에 진입할 때까지 대기 (단일 permit이면 정확히 1개).
	waitUntil(t, func() bool { return executed.Load() >= 1 }, 2*time.Second)
	// 픽스 전 다중 진입(concurrentMax > 1) 관측 여유를 준다.
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	if concurrentMax.Load() > 1 {
		t.Errorf("half-open admitted %d concurrent fn() executions, want exactly 1 (single-permit invariant) (AC-SEC-M5-001/002)", concurrentMax.Load())
	}
	if executed.Load() != 1 {
		t.Errorf("half-open executed fn() %d times, want exactly 1 (AC-SEC-M5-002)", executed.Load())
	}
	if rejected.Load() != n-1 {
		t.Errorf("half-open rejected %d callers with ErrCircuitOpen, want %d (AC-SEC-M5-002)", rejected.Load(), n-1)
	}
}

// TestCircuitBreaker_HalfOpenPermitReleased 는 AC-SEC-M5-003 (GREEN) 다.
// 단일 half-open trial이 해소되면 permit이 해제되어 다음 Call이 새 state로 governed된다.
func TestCircuitBreaker_HalfOpenPermitReleased(t *testing.T) {
	t.Parallel()

	threshold := 2
	timeout := 50 * time.Millisecond
	ctx := context.Background()

	t.Run("success trial closes then admits", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{Threshold: threshold, Timeout: timeout})
		promoteToHalfOpen(t, cb, threshold, timeout)
		// 성공 trial → closed.
		if err := cb.Call(ctx, func() error { return nil }); err != nil {
			t.Fatalf("half-open success trial error = %v", err)
		}
		if cb.State() != StateClosed {
			t.Fatalf("after success trial state = %v, want closed", cb.State())
		}
		// permit 해제 확인: 후속 Call이 admit된다 (closed 상태).
		if err := cb.Call(ctx, func() error { return nil }); err != nil {
			t.Errorf("subsequent call after permit release error = %v, want nil (permit released) (AC-SEC-M5-003)", err)
		}
	})

	t.Run("failure trial reopens then rejects", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{Threshold: threshold, Timeout: timeout})
		promoteToHalfOpen(t, cb, threshold, timeout)
		opErr := errors.New("trial failed")
		// 실패 trial → open.
		if err := cb.Call(ctx, func() error { return opErr }); !errors.Is(err, opErr) {
			t.Fatalf("half-open failure trial error = %v, want opErr", err)
		}
		if cb.State() != StateOpen {
			t.Fatalf("after failure trial state = %v, want open", cb.State())
		}
		// permit이 해제되었으되 state가 open이므로 후속 Call은 ErrCircuitOpen.
		if err := cb.Call(ctx, func() error { return nil }); !errors.Is(err, ErrCircuitOpen) {
			t.Errorf("subsequent call after failure trial error = %v, want ErrCircuitOpen (AC-SEC-M5-003)", err)
		}
	})
}

// TestCircuitBreaker_PanickingCallbackRecovered 는 AC-SEC-M5-005 (GREEN, AUTOMATED) 다.
// 픽스 후 panic하는 OnStateChange는 goroutine 내부에서 recover되어 프로세스가 생존하고,
// breaker의 post-transition state가 정상이어야 한다.
//
// (AC-SEC-M5-004 OBSERVATIONAL RED는 픽스 전 panic이 프로세스를 crash시키므로 자동 테스트로
//  commit하지 않는다 — 수동 1회 관측 후 progress.md에 기록.)
func TestCircuitBreaker_PanickingCallbackRecovered(t *testing.T) {
	t.Parallel()

	threshold := 2
	var transitions atomic.Int32
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   30 * time.Second,
		OnStateChange: func(_, _ CircuitState) {
			transitions.Add(1)
			panic("intentional callback panic")
		},
	})

	ctx := context.Background()
	opErr := errors.New("op failed")
	// threshold 실패 → closed→open 전이 → goroutine에서 panic하는 OnStateChange 발사.
	for range threshold {
		_ = cb.Call(ctx, func() error { return opErr })
	}

	// 프로세스가 생존했으면 여기 도달한다 (recover 미존재 시 suite crash).
	// 비동기 goroutine이 callback을 실행할 시간을 준다.
	waitUntil(t, func() bool { return transitions.Load() >= 1 }, 2*time.Second)

	if cb.State() != StateOpen {
		t.Errorf("post-transition state = %v, want open (state correct despite panicking callback) (AC-SEC-M5-005)", cb.State())
	}
}

// TestCircuitBreaker_MetricsAndTransitionsUnchanged 는 AC-SEC-M5-006 (NO-REG) 다.
// closed→open threshold, timeout 승격, half-open success→closed, metrics 회계가 불변임을 확인한다.
func TestCircuitBreaker_MetricsAndTransitionsUnchanged(t *testing.T) {
	t.Parallel()

	threshold := 3
	timeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{Threshold: threshold, Timeout: timeout})
	ctx := context.Background()
	opErr := errors.New("op failed")

	// 1 success (closed).
	_ = cb.Call(ctx, func() error { return nil })
	// threshold failures → open.
	for range threshold {
		_ = cb.Call(ctx, func() error { return opErr })
	}
	if cb.State() != StateOpen {
		t.Fatalf("state after threshold failures = %v, want open", cb.State())
	}
	// open 상태 즉시 Call → rejected.
	if err := cb.Call(ctx, func() error { return nil }); !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("open-state call error = %v, want ErrCircuitOpen", err)
	}
	// timeout 후 success trial → closed.
	time.Sleep(timeout + 20*time.Millisecond)
	if err := cb.Call(ctx, func() error { return nil }); err != nil {
		t.Errorf("half-open success trial error = %v, want nil", err)
	}
	if cb.State() != StateClosed {
		t.Errorf("state after half-open success = %v, want closed", cb.State())
	}

	m := cb.Metrics()
	// TotalCalls: 1 success + threshold failures + 1 rejected + 1 half-open success.
	wantTotal := int64(1 + threshold + 1 + 1)
	if m.TotalCalls != wantTotal {
		t.Errorf("TotalCalls = %d, want %d", m.TotalCalls, wantTotal)
	}
	// SuccessCount: 1 initial + 1 half-open trial.
	if m.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", m.SuccessCount)
	}
	if m.FailureCount != int64(threshold) {
		t.Errorf("FailureCount = %d, want %d", m.FailureCount, threshold)
	}
	if m.RejectedCount != 1 {
		t.Errorf("RejectedCount = %d, want 1", m.RejectedCount)
	}
}

// TestCircuitBreaker_ResetSynchronousCallbackUnchanged 는 AC-SEC-M5-006 의 일부로,
// Reset()의 동기 OnStateChange 경로가 M5 변경에 영향받지 않음을 확인한다.
func TestCircuitBreaker_ResetSynchronousCallbackUnchanged(t *testing.T) {
	t.Parallel()

	threshold := 2
	var resetCb atomic.Int32
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold: threshold,
		Timeout:   30 * time.Second,
		OnStateChange: func(_, to CircuitState) {
			if to == StateClosed {
				resetCb.Add(1)
			}
		},
	})
	ctx := context.Background()
	opErr := errors.New("op failed")
	for range threshold {
		_ = cb.Call(ctx, func() error { return opErr })
	}
	if cb.State() != StateOpen {
		t.Fatalf("setup: want open, got %v", cb.State())
	}

	// Reset()은 동기적으로 OnStateChange(open→closed)를 호출한다.
	cb.Reset()
	if cb.State() != StateClosed {
		t.Errorf("state after Reset = %v, want closed", cb.State())
	}
	// 동기 호출이므로 Reset 반환 시점에 callback이 이미 1회 실행됨.
	if resetCb.Load() < 1 {
		t.Errorf("Reset synchronous OnStateChange not invoked (want >=1), got %d", resetCb.Load())
	}
	// Reset 후 정상 Call admit.
	if err := cb.Call(ctx, func() error { return nil }); err != nil {
		t.Errorf("call after Reset error = %v, want nil", err)
	}
}

// waitUntil polls cond up to timeout, failing the test if cond never holds.
func waitUntil(t *testing.T, cond func() bool, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	if !cond() {
		t.Fatalf("waitUntil: condition not met within %v", timeout)
	}
}
