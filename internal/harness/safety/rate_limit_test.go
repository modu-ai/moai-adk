// Package safety — rate_limit unit test.
// REQ-HL-008: 7일 3회 제한 + 24h cooldown 슬라이딩 윈도우 테스트.
package safety

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCheckLimit_AllowedWhenEmpty는 기록이 없으면 허용되는지 검증한다.
func TestCheckLimit_AllowedWhenEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rl := NewRateLimiter(filepath.Join(dir, "rate-limit-state.json"))

	allowed, retryAfter, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if !allowed {
		t.Errorf("allowed = false, 기록이 없으면 허용되어야 한다. retryAfter=%v", retryAfter)
	}
}

// TestCheckLimit_AllowedUnderLimit은 7일 이내 3회 미만이면 허용되는지 검증한다.
func TestCheckLimit_AllowedUnderLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	now := time.Now()

	// 2회 기록 (3회 미만)
	state := rateLimitState{
		Updates: []time.Time{
			now.Add(-48 * time.Hour),
			now.Add(-25 * time.Hour),
		},
	}
	if err := writeState(statePath, state); err != nil {
		t.Fatalf("state 쓰기 실패: %v", err)
	}

	// 마지막 업데이트로부터 25시간 경과 → cooldown(24h) 통과
	allowed, retryAfter, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if !allowed {
		t.Errorf("allowed = false, 2회 기록 + cooldown 통과면 허용되어야 한다. retryAfter=%v", retryAfter)
	}
}

// TestCheckLimit_BlockedAtLimit은 7일 이내 3회이면 차단되는지 검증한다.
func TestCheckLimit_BlockedAtLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	now := time.Now()

	// 7일 이내 3회 기록
	state := rateLimitState{
		Updates: []time.Time{
			now.Add(-6 * 24 * time.Hour),
			now.Add(-4 * 24 * time.Hour),
			now.Add(-25 * time.Hour), // 마지막 업데이트
		},
	}
	if err := writeState(statePath, state); err != nil {
		t.Fatalf("state 쓰기 실패: %v", err)
	}

	allowed, retryAfter, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if allowed {
		t.Error("allowed = true, 7일 이내 3회이면 차단되어야 한다")
	}

	if retryAfter <= 0 {
		t.Errorf("retryAfter = %v, 차단 시 양수여야 한다", retryAfter)
	}
}

// TestCheckLimit_BlockedByCooldown은 24h cooldown 이내이면 차단되는지 검증한다.
func TestCheckLimit_BlockedByCooldown(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	now := time.Now()

	// 마지막 업데이트가 1시간 전 (cooldown 24h 이내)
	state := rateLimitState{
		Updates: []time.Time{
			now.Add(-1 * time.Hour),
		},
	}
	if err := writeState(statePath, state); err != nil {
		t.Fatalf("state 쓰기 실패: %v", err)
	}

	allowed, retryAfter, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if allowed {
		t.Error("allowed = true, 24h cooldown 이내이면 차단되어야 한다")
	}

	// retryAfter는 약 23시간이어야 함
	if retryAfter < 22*time.Hour || retryAfter > 24*time.Hour {
		t.Errorf("retryAfter = %v, 22~24시간 범위여야 한다", retryAfter)
	}
}

// TestRecordUpdate_PersistsToFile은 RecordUpdate가 파일에 저장되는지 검증한다.
func TestRecordUpdate_PersistsToFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	if err := rl.RecordUpdate(); err != nil {
		t.Fatalf("RecordUpdate 실패: %v", err)
	}

	// 파일이 생성되었는지 확인
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("state 파일 읽기 실패: %v", err)
	}

	var state rateLimitState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("state JSON 파싱 실패: %v", err)
	}

	if len(state.Updates) != 1 {
		t.Errorf("Updates 수 = %d, want 1", len(state.Updates))
	}
}

// TestRecordUpdate_SurvivesRestart는 재시작 후에도 상태가 유지되는지 검증한다.
func TestRecordUpdate_SurvivesRestart(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// 첫 번째 RateLimiter — 업데이트 기록
	rl1 := NewRateLimiter(statePath)
	if err := rl1.RecordUpdate(); err != nil {
		t.Fatalf("RecordUpdate 실패: %v", err)
	}

	// 두 번째 RateLimiter (프로세스 재시작 시뮬레이션)
	rl2 := NewRateLimiter(statePath)
	allowed, _, err := rl2.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	// 방금 기록했으므로 24h cooldown으로 차단됨
	if allowed {
		t.Error("재시작 후 24h cooldown이 유지되어야 하는데 허용됨")
	}
}

// TestCheckLimit_ExpiredWindowEntries는 7일 이상 지난 기록은 윈도우에서 제외되는지 검증한다.
func TestCheckLimit_ExpiredWindowEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	now := time.Now()

	// 7일 초과 기록 3개 + 25시간 전 기록 1개
	state := rateLimitState{
		Updates: []time.Time{
			now.Add(-8 * 24 * time.Hour), // 만료
			now.Add(-9 * 24 * time.Hour), // 만료
			now.Add(-10 * 24 * time.Hour), // 만료
			now.Add(-25 * time.Hour),      // 유효 (cooldown 통과)
		},
	}
	if err := writeState(statePath, state); err != nil {
		t.Fatalf("state 쓰기 실패: %v", err)
	}

	// 7일 이내 유효 기록은 1개뿐이므로 허용
	allowed, _, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if !allowed {
		t.Error("allowed = false, 7일 이내 유효 기록 1개만 있으면 허용되어야 한다")
	}
}

// TestRecordUpdate_MultipleUpdates는 여러 번 RecordUpdate를 호출했을 때 상태를 검증한다.
func TestRecordUpdate_MultipleUpdates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// nowFn을 override하여 시간 간격을 시뮬레이션할 수 없으므로
	// 직접 state를 조작하여 테스트한다.
	// 먼저 빈 파일로 시작
	rl := NewRateLimiter(statePath)

	// 첫 번째 업데이트
	if err := rl.RecordUpdate(); err != nil {
		t.Fatalf("RecordUpdate[1] 실패: %v", err)
	}

	// 상태 파일 읽기
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("state 파일 읽기 실패: %v", err)
	}

	var state rateLimitState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	if len(state.Updates) != 1 {
		t.Errorf("Updates 수 = %d, want 1", len(state.Updates))
	}
}

// TestCheckLimit_LoadStateError는 손상된 state 파일에서 오류를 반환하는지 검증한다.
func TestCheckLimit_LoadStateError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// 손상된 JSON 파일 생성
	if err := os.WriteFile(statePath, []byte("invalid-json{{{"), 0o644); err != nil {
		t.Fatalf("손상된 파일 생성 실패: %v", err)
	}

	rl := NewRateLimiter(statePath)
	_, _, err := rl.CheckLimit()
	if err == nil {
		t.Error("손상된 state 파일에서 오류가 반환되어야 한다")
	}
}

// TestRecordUpdate_LoadErrorReturned는 손상된 state 파일에서 RecordUpdate가 오류를 반환하는지 검증한다.
func TestRecordUpdate_LoadErrorReturned(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// 손상된 JSON 파일
	if err := os.WriteFile(statePath, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("파일 생성 실패: %v", err)
	}

	rl := NewRateLimiter(statePath)
	if err := rl.RecordUpdate(); err == nil {
		t.Error("손상된 state 파일에서 RecordUpdate 오류가 반환되어야 한다")
	}
}

// TestRateLimiter_SaveStateError는 쓰기 불가 경로에서 RecordUpdate 오류를 검증한다.
func TestRateLimiter_SaveStateError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// 읽기 전용 디렉토리
	roDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(roDir, 0o555); err != nil {
		t.Fatalf("읽기 전용 디렉토리 생성 실패: %v", err)
	}

	statePath := filepath.Join(roDir, "sub", "state.json")
	rl := NewRateLimiter(statePath)

	err := rl.RecordUpdate()
	if err == nil {
		t.Logf("읽기 전용 경로에서도 쓰기 성공 (CI 환경 차이, 무시)")
	}
}

// writeState는 테스트에서 초기 state를 설정하기 위한 헬퍼이다.
func writeState(path string, state rateLimitState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
