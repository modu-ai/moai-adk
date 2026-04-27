//go:build integration
// +build integration

// Package harness_integration — T-P5-05: Rate Limiter 4번째 업데이트 거부 검증 (REQ-HL-008).
package harness_integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/safety"
)

// rateLimitTestState는 rate-limit-state.json의 직접 조작을 위한 내부 구조체이다.
// 테스트에서 시계를 직접 제어할 수 없으므로 상태 파일을 직접 설정한다.
// (rate_limit_test.go의 writeState 헬퍼와 동일한 방식)
type rateLimitTestState struct {
	Updates []time.Time `json:"updates"`
}

// writeRateLimitState는 지정된 경로에 rate limit 상태를 직접 쓴다.
func writeRateLimitState(t *testing.T, path string, state rateLimitTestState) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("상태 파일 디렉토리 생성 실패: %v", err)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("rate limit state JSON 직렬화 실패: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("rate limit state 파일 쓰기 실패: %v", err)
	}
}

// TestIT05_RateLimiterFourthUpdateRejected는 7일 슬라이딩 윈도우 내
// 3회 업데이트 이후 4번째 업데이트 시도가 거부되는지 검증한다.
// REQ-HL-008: max 3 auto-updates per 7-day window.
//
// 결정론적 보장: 시계 함수 주입 대신 상태 파일 직접 조작을 사용한다.
// (clock injection이 RateLimiter에 package-private이므로 상태 파일 조작이 유일한 방법)
func TestIT05_RateLimiterFourthUpdateRejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// ── 기준 시각: 현재보다 충분히 최근 (7일 이내) ──────────────────────
	// 7일 이내에 3회 업데이트가 이미 발생한 상태를 설정한다.
	// 각 업데이트 간격을 25시간으로 설정하여 24h cooldown 조건도 충족한다.
	//
	// 예시 타임라인 (현재 시각 기준):
	//   T-6d: 업데이트 1 (7일 이내)
	//   T-5d: 업데이트 2 (7일 이내, cooldown OK)
	//   T-4d: 업데이트 3 (7일 이내, cooldown OK)
	//   T-now: 업데이트 4 → 거부 (7일 내 3회 이미 소진)
	now := time.Now()
	state := rateLimitTestState{
		Updates: []time.Time{
			now.Add(-6 * 24 * time.Hour), // 6일 전 (7일 이내)
			now.Add(-5 * 24 * time.Hour), // 5일 전 (7일 이내)
			now.Add(-4 * 24 * time.Hour), // 4일 전 (7일 이내, 마지막 업데이트)
		},
	}
	writeRateLimitState(t, statePath, state)

	rl := safety.NewRateLimiter(statePath)

	// ── 4번째 시도 → 차단 검증 ────────────────────────────────────────────
	allowed, retryAfter, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if allowed {
		t.Error("4번째 업데이트 시도가 허용됨 — 7일 이내 3회 소진 후 차단 기대")
	}

	if retryAfter <= 0 {
		t.Errorf("retryAfter = %v, 차단 시 양수여야 한다", retryAfter)
	}

	// retryAfter는 첫 번째 업데이트(6일 전)가 7일 창을 벗어날 때까지의 대기 시간
	// 약 1일(24h) 정도가 기대됨
	if retryAfter > 2*24*time.Hour {
		t.Errorf("retryAfter = %v, 너무 크다 (2일 이하 기대)", retryAfter)
	}

	// ── 7일 이전 업데이트만 있는 경우: 허용 검증 ─────────────────────────
	t.Run("expired_window_allows", func(t *testing.T) {
		expiredDir := t.TempDir()
		expiredPath := filepath.Join(expiredDir, "rate-limit-state.json")

		// 7일 이상 전 업데이트 3개 → 윈도우 만료, 허용
		expiredState := rateLimitTestState{
			Updates: []time.Time{
				now.Add(-10 * 24 * time.Hour), // 10일 전 (만료)
				now.Add(-9 * 24 * time.Hour),  // 9일 전 (만료)
				now.Add(-8 * 24 * time.Hour),  // 8일 전 (만료)
			},
		}
		writeRateLimitState(t, expiredPath, expiredState)

		expiredRL := safety.NewRateLimiter(expiredPath)
		expAllowed, _, expErr := expiredRL.CheckLimit()
		if expErr != nil {
			t.Fatalf("CheckLimit 실패: %v", expErr)
		}
		if !expAllowed {
			t.Error("만료된 업데이트만 있을 때 허용되어야 한다")
		}
	})

	// ── 빈 상태: 허용 검증 ────────────────────────────────────────────────
	t.Run("empty_state_allows", func(t *testing.T) {
		emptyDir := t.TempDir()
		emptyPath := filepath.Join(emptyDir, "rate-limit-state.json")

		emptyRL := safety.NewRateLimiter(emptyPath)
		emptyAllowed, _, emptyErr := emptyRL.CheckLimit()
		if emptyErr != nil {
			t.Fatalf("CheckLimit 실패: %v", emptyErr)
		}
		if !emptyAllowed {
			t.Error("빈 상태에서 허용되어야 한다")
		}
	})
}
