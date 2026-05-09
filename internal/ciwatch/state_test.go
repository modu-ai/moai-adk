package ciwatch_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
)

// TestStateFile_WriteRead verifies a state file can be written and read back.
func TestStateFile_WriteRead(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "ci-watch-active.flag")

	s := ciwatch.WatchState{
		PRNumber:       785,
		StartedAt:      time.Now().UTC().Truncate(time.Second),
		HeartbeatAt:    time.Now().UTC().Truncate(time.Second),
		RequiredChecks: []string{"Lint", "Test (ubuntu-latest)", "CodeQL"},
		AbortRequested: false,
	}

	if err := ciwatch.WriteState(statePath, s); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	got, err := ciwatch.ReadState(statePath)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}

	if got.PRNumber != s.PRNumber {
		t.Errorf("PRNumber = %d, want %d", got.PRNumber, s.PRNumber)
	}
	if len(got.RequiredChecks) != len(s.RequiredChecks) {
		t.Errorf("RequiredChecks len = %d, want %d", len(got.RequiredChecks), len(s.RequiredChecks))
	}
	if got.AbortRequested {
		t.Error("AbortRequested should be false")
	}
}

// TestStateFile_HeartbeatStale verifies that a stale heartbeat (> 90s) is
// detected and allows a new watch to proceed.
func TestStateFile_HeartbeatStale(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "ci-watch-active.flag")

	staleTime := time.Now().UTC().Add(-2 * time.Minute) // 2 min ago — definitely stale
	s := ciwatch.WatchState{
		PRNumber:       100,
		StartedAt:      staleTime,
		HeartbeatAt:    staleTime,
		RequiredChecks: []string{"Lint"},
	}

	if err := ciwatch.WriteState(statePath, s); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	got, err := ciwatch.ReadState(statePath)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}

	// IsStale should return true for a heartbeat > 90s old.
	if !got.IsStale(90 * time.Second) {
		t.Error("expected stale heartbeat to be detected")
	}
}

// TestStateFile_FreshHeartbeat verifies a recent heartbeat is NOT stale.
func TestStateFile_FreshHeartbeat(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "ci-watch-active.flag")

	s := ciwatch.WatchState{
		PRNumber:    200,
		StartedAt:   time.Now().UTC(),
		HeartbeatAt: time.Now().UTC(), // just now
	}

	if err := ciwatch.WriteState(statePath, s); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	got, err := ciwatch.ReadState(statePath)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}

	if got.IsStale(90 * time.Second) {
		t.Error("fresh heartbeat should not be stale")
	}
}

// TestStateFile_AbortFlag verifies that writing abort_requested: true is
// detectable by the reader.
func TestStateFile_AbortFlag(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "ci-watch-active.flag")

	s := ciwatch.WatchState{
		PRNumber:       300,
		StartedAt:      time.Now().UTC(),
		HeartbeatAt:    time.Now().UTC(),
		AbortRequested: false,
	}

	if err := ciwatch.WriteState(statePath, s); err != nil {
		t.Fatalf("WriteState initial: %v", err)
	}

	// Set the abort flag.
	if err := ciwatch.SetAbortFlag(statePath); err != nil {
		t.Fatalf("SetAbortFlag: %v", err)
	}

	got, err := ciwatch.ReadState(statePath)
	if err != nil {
		t.Fatalf("ReadState after abort: %v", err)
	}

	if !got.AbortRequested {
		t.Error("AbortRequested should be true after SetAbortFlag")
	}
}

// TestStateFile_Touch verifies the heartbeat timestamp is updated by Touch.
func TestStateFile_Touch(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "ci-watch-active.flag")

	oldTime := time.Now().UTC().Add(-5 * time.Minute)
	s := ciwatch.WatchState{
		PRNumber:    400,
		StartedAt:   oldTime,
		HeartbeatAt: oldTime,
	}

	if err := ciwatch.WriteState(statePath, s); err != nil {
		t.Fatalf("WriteState: %v", err)
	}

	if err := ciwatch.Touch(statePath); err != nil {
		t.Fatalf("Touch: %v", err)
	}

	got, err := ciwatch.ReadState(statePath)
	if err != nil {
		t.Fatalf("ReadState after Touch: %v", err)
	}

	// HeartbeatAt should be newer than oldTime.
	if !got.HeartbeatAt.After(oldTime) {
		t.Errorf("HeartbeatAt not updated: got %v, old was %v", got.HeartbeatAt, oldTime)
	}
}

// TestStateFile_NotFound verifies graceful error when file is absent.
func TestStateFile_NotFound(t *testing.T) {
	_, err := ciwatch.ReadState("/nonexistent/path/ci-watch-active.flag")
	if err == nil {
		t.Error("expected error for missing state file, got nil")
	}
	if !os.IsNotExist(err) {
		// Allow wrapped errors too.
		if !isNotExistError(err) {
			t.Logf("got non-NotExist error (acceptable): %v", err)
		}
	}
}

func isNotExistError(err error) bool {
	return os.IsNotExist(err) || (err != nil && containsStr(err.Error(), "no such file"))
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && searchStr(s, sub))
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
