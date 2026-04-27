// Package safety — Layer 4: Rate Limiter (REQ-HL-008).
// 슬라이딩 윈도우 기반 자동 업데이트 횟수 제한:
//   - 7일 이내 최대 3회
//   - 업데이트 간 최소 24h cooldown
//
// 상태는 JSON 파일에 영속화되어 프로세스 재시작 후에도 유지된다.
package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// rateLimitMaxUpdates는 슬라이딩 윈도우 내 최대 허용 업데이트 횟수이다.
// REQ-HL-008: max 3 auto-updates per 7-day window.
const rateLimitMaxUpdates = 3

// rateLimitWindow는 슬라이딩 윈도우 기간이다 (7일).
const rateLimitWindow = 7 * 24 * time.Hour

// rateLimitCooldown은 업데이트 간 최소 대기 시간이다 (24h).
const rateLimitCooldown = 24 * time.Hour

// rateLimitState는 rate-limit-state.json의 영속화 스키마이다.
type rateLimitState struct {
	// Updates는 슬라이딩 윈도우 내 업데이트 시각 목록이다 (UTC).
	Updates []time.Time `json:"updates"`
}

// RateLimiter는 슬라이딩 윈도우 기반 rate limiter이다.
//
// @MX:ANCHOR: [AUTO] RateLimiter는 Layer 4의 핵심 컴포넌트이다.
// @MX:REASON: [AUTO] fan_in >= 3: rate_limit_test.go, pipeline.go, Phase 4 coordinator
type RateLimiter struct {
	// statePath는 상태 파일 경로이다.
	statePath string

	// nowFn은 현재 시각 함수 (테스트에서 override 가능).
	nowFn func() time.Time
}

// NewRateLimiter는 지정된 statePath를 사용하는 RateLimiter를 생성한다.
func NewRateLimiter(statePath string) *RateLimiter {
	return &RateLimiter{
		statePath: statePath,
		nowFn:     time.Now,
	}
}

// CheckLimit은 현재 업데이트 가능 여부를 반환한다.
// REQ-HL-008: 7일 이내 3회 초과이거나 24h cooldown 이내이면 차단한다.
//
// 반환값:
//   - allowed: true이면 업데이트 허용
//   - retryAfter: 차단 시 대기해야 할 시간 (allowed=true면 0)
//   - err: 상태 파일 읽기 오류
func (rl *RateLimiter) CheckLimit() (allowed bool, retryAfter time.Duration, err error) {
	state, err := rl.loadState()
	if err != nil {
		return false, 0, err
	}

	now := rl.nowFn()
	windowStart := now.Add(-rateLimitWindow)

	// 슬라이딩 윈도우 내 유효 업데이트만 필터링
	var validUpdates []time.Time
	for _, t := range state.Updates {
		if t.After(windowStart) {
			validUpdates = append(validUpdates, t)
		}
	}

	// 7일 이내 3회 초과 체크
	if len(validUpdates) >= rateLimitMaxUpdates {
		// 가장 오래된 유효 업데이트가 만료되는 시간을 계산
		oldest := validUpdates[0]
		waitUntil := oldest.Add(rateLimitWindow)
		retryAfter = waitUntil.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter, nil
	}

	// 24h cooldown 체크
	if len(validUpdates) > 0 {
		lastUpdate := validUpdates[len(validUpdates)-1]
		cooldownEnd := lastUpdate.Add(rateLimitCooldown)
		if now.Before(cooldownEnd) {
			retryAfter = cooldownEnd.Sub(now)
			return false, retryAfter, nil
		}
	}

	return true, 0, nil
}

// RecordUpdate는 현재 시각을 업데이트 이력에 추가하고 상태 파일에 저장한다.
// REQ-HL-008: 상태는 프로세스 재시작 후에도 유지되어야 한다.
func (rl *RateLimiter) RecordUpdate() error {
	state, err := rl.loadState()
	if err != nil {
		return err
	}

	now := rl.nowFn().UTC()

	// 7일 이내 유효 기록만 유지 + 새 기록 추가
	windowStart := now.Add(-rateLimitWindow)
	var newUpdates []time.Time
	for _, t := range state.Updates {
		if t.After(windowStart) {
			newUpdates = append(newUpdates, t)
		}
	}
	newUpdates = append(newUpdates, now)
	state.Updates = newUpdates

	return rl.saveState(state)
}

// loadState는 상태 파일을 읽어 반환한다.
// 파일이 없으면 빈 상태를 반환한다.
func (rl *RateLimiter) loadState() (rateLimitState, error) {
	var state rateLimitState

	data, err := os.ReadFile(rl.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, fmt.Errorf("safety/rate_limit: 상태 파일 읽기 실패 %s: %w", rl.statePath, err)
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("safety/rate_limit: 상태 파일 파싱 실패 %s: %w", rl.statePath, err)
	}

	return state, nil
}

// saveState는 상태를 파일에 저장한다.
func (rl *RateLimiter) saveState(state rateLimitState) error {
	dir := filepath.Dir(rl.statePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("safety/rate_limit: 디렉토리 생성 실패 %s: %w", dir, err)
		}
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("safety/rate_limit: 상태 직렬화 실패: %w", err)
	}

	if err := os.WriteFile(rl.statePath, data, 0o644); err != nil {
		return fmt.Errorf("safety/rate_limit: 상태 파일 쓰기 실패 %s: %w", rl.statePath, err)
	}

	return nil
}
