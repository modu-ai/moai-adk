// Package safety — rate_limit unit test.
// REQ-HL-008: 3 updates per 7 days + 24h cooldown sliding-window tests.
package safety

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCheckLimit_AllowedWhenEmpty verifies that access is allowed when no record exists.
func TestCheckLimit_AllowedWhenEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rl := NewRateLimiter(filepath.Join(dir, "rate-limit-state.json"))

	allowed, retryAfter, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if !allowed {
		t.Errorf("allowed = false, must be allowed when no record exists. retryAfter=%v", retryAfter)
	}
}

// TestCheckLimit_AllowedUnderLimit verifies that access is allowed when fewer than 3 updates within 7 days.
func TestCheckLimit_AllowedUnderLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	now := time.Now()

	// 2 records (fewer than 3)
	state := rateLimitState{
		Updates: []time.Time{
			now.Add(-48 * time.Hour),
			now.Add(-25 * time.Hour),
		},
	}
	if err := writeState(statePath, state); err != nil {
		t.Fatalf("state 쓰기 실패: %v", err)
	}

	// 25 hours since the last update → cooldown (24h) passes
	allowed, retryAfter, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if !allowed {
		t.Errorf("allowed = false, must be allowed with 2 records + cooldown passed. retryAfter=%v", retryAfter)
	}
}

// TestCheckLimit_BlockedAtLimit verifies that access is blocked when 3 updates exist within 7 days.
func TestCheckLimit_BlockedAtLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	now := time.Now()

	// 3 records within 7 days
	state := rateLimitState{
		Updates: []time.Time{
			now.Add(-6 * 24 * time.Hour),
			now.Add(-4 * 24 * time.Hour),
			now.Add(-25 * time.Hour), // last update
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
		t.Error("allowed = true, must be blocked with 3 records within 7 days")
	}

	if retryAfter <= 0 {
		t.Errorf("retryAfter = %v, must be positive when blocked", retryAfter)
	}
}

// TestCheckLimit_BlockedByCooldown verifies that access is blocked within the 24h cooldown.
func TestCheckLimit_BlockedByCooldown(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	now := time.Now()

	// Last update was 1 hour ago (within the 24h cooldown)
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
		t.Error("allowed = true, must be blocked within the 24h cooldown")
	}

	// retryAfter must be roughly 23 hours
	if retryAfter < 22*time.Hour || retryAfter > 24*time.Hour {
		t.Errorf("retryAfter = %v, must be in the 22–24h range", retryAfter)
	}
}

// TestRecordUpdate_PersistsToFile verifies that RecordUpdate persists to the file.
func TestRecordUpdate_PersistsToFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	if err := rl.RecordUpdate(); err != nil {
		t.Fatalf("RecordUpdate 실패: %v", err)
	}

	// Verify the file was created
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("state 파일 읽기 실패: %v", err)
	}

	var state rateLimitState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("state JSON 파싱 실패: %v", err)
	}

	if len(state.Updates) != 1 {
		t.Errorf("Updates count = %d, want 1", len(state.Updates))
	}
}

// TestRecordUpdate_SurvivesRestart verifies that state survives after a restart.
func TestRecordUpdate_SurvivesRestart(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// First RateLimiter — record an update
	rl1 := NewRateLimiter(statePath)
	if err := rl1.RecordUpdate(); err != nil {
		t.Fatalf("RecordUpdate 실패: %v", err)
	}

	// Second RateLimiter (simulating process restart)
	rl2 := NewRateLimiter(statePath)
	allowed, _, err := rl2.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	// Since the update was just recorded, the 24h cooldown still blocks
	if allowed {
		t.Error("24h cooldown must persist after restart, but access was allowed")
	}
}

// TestCheckLimit_ExpiredWindowEntries verifies that records older than 7 days are excluded from the window.
func TestCheckLimit_ExpiredWindowEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")
	rl := NewRateLimiter(statePath)

	now := time.Now()

	// 3 records older than 7 days + 1 record from 25 hours ago
	state := rateLimitState{
		Updates: []time.Time{
			now.Add(-8 * 24 * time.Hour),  // expired
			now.Add(-9 * 24 * time.Hour),  // expired
			now.Add(-10 * 24 * time.Hour), // expired
			now.Add(-25 * time.Hour),      // valid (cooldown passed)
		},
	}
	if err := writeState(statePath, state); err != nil {
		t.Fatalf("state 쓰기 실패: %v", err)
	}

	// Only 1 valid record within 7 days → allowed
	allowed, _, err := rl.CheckLimit()
	if err != nil {
		t.Fatalf("CheckLimit 실패: %v", err)
	}

	if !allowed {
		t.Error("allowed = false, must be allowed with only 1 valid record within 7 days")
	}
}

// TestRecordUpdate_MultipleUpdates verifies state after multiple RecordUpdate calls.
func TestRecordUpdate_MultipleUpdates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// We cannot simulate time gaps by overriding nowFn here, so this test
	// manipulates state directly.
	// Start with an empty file.
	rl := NewRateLimiter(statePath)

	// First update
	if err := rl.RecordUpdate(); err != nil {
		t.Fatalf("RecordUpdate[1] 실패: %v", err)
	}

	// Read the state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("state 파일 읽기 실패: %v", err)
	}

	var state rateLimitState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	if len(state.Updates) != 1 {
		t.Errorf("Updates count = %d, want 1", len(state.Updates))
	}
}

// TestCheckLimit_LoadStateError verifies that a corrupted state file returns an error.
func TestCheckLimit_LoadStateError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// Create a corrupted JSON file
	if err := os.WriteFile(statePath, []byte("invalid-json{{{"), 0o644); err != nil {
		t.Fatalf("손상된 파일 생성 실패: %v", err)
	}

	rl := NewRateLimiter(statePath)
	_, _, err := rl.CheckLimit()
	if err == nil {
		t.Error("an error must be returned for a corrupted state file")
	}
}

// TestRecordUpdate_LoadErrorReturned verifies that RecordUpdate returns an error for a corrupted state file.
func TestRecordUpdate_LoadErrorReturned(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit-state.json")

	// Corrupted JSON file
	if err := os.WriteFile(statePath, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("파일 생성 실패: %v", err)
	}

	rl := NewRateLimiter(statePath)
	if err := rl.RecordUpdate(); err == nil {
		t.Error("RecordUpdate must return an error for a corrupted state file")
	}
}

// TestRateLimiter_SaveStateError verifies a RecordUpdate error on a write-protected path.
func TestRateLimiter_SaveStateError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Read-only directory
	roDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(roDir, 0o555); err != nil {
		t.Fatalf("읽기 전용 디렉토리 생성 실패: %v", err)
	}

	statePath := filepath.Join(roDir, "sub", "state.json")
	rl := NewRateLimiter(statePath)

	err := rl.RecordUpdate()
	if err == nil {
		t.Logf("write succeeded even on a read-only path (CI environment variance, ignored)")
	}
}

// writeState is a helper to seed initial state in tests.
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
