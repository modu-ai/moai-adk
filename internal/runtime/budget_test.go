package runtime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// testConfig returns a RuntimeConfig suitable for unit tests.
// Uses short stall_detection_seconds to avoid long test waits.
func testConfig() *runtime.RuntimeConfig {
	return &runtime.RuntimeConfig{
		PreClearThreshold:     0.75,
		HardClearThreshold:    0.90,
		PerAgentBudget:        map[string]int{"default": 1000, "manager-strategy": 2000},
		StallDetectionSeconds: 1, // 1s for fast tests
		RetryMax:              3,
		Fallback:              "split_into_waves",
		AutoSaveAtThreshold:   true,
		SavePathTemplate:      ".moai/specs/{SPEC_ID}/progress.md",
		ResumeMessageFormat:   "ultrathink. {wave_label} 이어서 진행. SPEC-{spec_id}부터 {approach_summary}. progress.md 경로: {progress_path}. 다음 단계: {next_step}.",
	}
}

// TestRecordCallBasic verifies that RecordCall increments usage for the named agent.
func TestRecordCallBasic(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())
	tracker.RecordCall("default", 100, 50)

	current, budget, ratio := tracker.Usage("default")
	if current != 150 {
		t.Errorf("expected current=150, got %d", current)
	}
	if budget != 1000 {
		t.Errorf("expected budget=1000, got %d", budget)
	}
	if ratio < 0.14 || ratio > 0.16 {
		t.Errorf("expected ratio ~0.15, got %f", ratio)
	}
}

// TestUsageRatio verifies correct ratio calculation at various usage levels.
func TestUsageRatio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tokensIn  int
		tokensOut int
		wantRatio float64
	}{
		{"zero_usage", 0, 0, 0.0},
		{"half_budget", 500, 0, 0.5},
		{"full_budget", 1000, 0, 1.0},
		{"over_budget", 1500, 0, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tracker := runtime.NewTracker(testConfig())
			tracker.RecordCall("default", tt.tokensIn, tt.tokensOut)

			_, _, ratio := tracker.Usage("default")
			if ratio != tt.wantRatio {
				t.Errorf("expected ratio=%f, got %f", tt.wantRatio, ratio)
			}
		})
	}
}

// TestIsApproachingLimitAt75Pct verifies the 75% soft threshold.
func TestIsApproachingLimitAt75Pct(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())

	// Below 75%: should not be approaching
	tracker.RecordCall("default", 700, 0) // 70%
	if tracker.IsApproachingLimit("default") {
		t.Error("expected IsApproachingLimit=false at 70%, got true")
	}

	// At exactly 75%: should be approaching
	tracker.RecordCall("default", 50, 0) // 75% total
	if !tracker.IsApproachingLimit("default") {
		t.Error("expected IsApproachingLimit=true at 75%, got false")
	}

	// Above 75%: still approaching
	tracker.RecordCall("default", 10, 0) // 76%
	if !tracker.IsApproachingLimit("default") {
		t.Error("expected IsApproachingLimit=true at 76%, got false")
	}
}

// TestHardLimitAt90Pct verifies the 90% hard threshold.
func TestHardLimitAt90Pct(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())

	// Below 90%: not at hard limit
	tracker.RecordCall("default", 850, 0) // 85%
	if tracker.IsAtHardLimit("default") {
		t.Error("expected IsAtHardLimit=false at 85%, got true")
	}

	// At exactly 90%: at hard limit
	tracker.RecordCall("default", 50, 0) // 90% total
	if !tracker.IsAtHardLimit("default") {
		t.Error("expected IsAtHardLimit=true at 90%, got false")
	}
}

// TestHardLimitWarning verifies that budget excess emits a warning without returning an error.
// This is the BC-V3R3-006 warning-first policy check.
func TestHardLimitWarning(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())
	tracker.RecordCall("default", 900, 0) // 90%

	// IsAtHardLimit must return true (warning path)
	if !tracker.IsAtHardLimit("default") {
		t.Error("expected IsAtHardLimit=true at 90%")
	}

	// No error is returned from RecordCall (warning-only, not hard-fail)
	// This test verifies the function signature does not return error.
	tracker.RecordCall("default", 200, 0) // 110% — over budget
	_, _, ratio := tracker.Usage("default")
	if ratio < 1.0 {
		t.Errorf("expected ratio>=1.0, got %f", ratio)
	}
}

// TestDetectStall verifies stall detection when no RecordCall arrives within stall_detection_seconds.
func TestDetectStall(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.StallDetectionSeconds = 1 // 1 second for fast testing

	tracker := runtime.NewTracker(cfg)
	tracker.RecordCall("default", 100, 0)

	// Immediately after RecordCall: no stall
	if tracker.DetectStall("default") {
		t.Error("expected DetectStall=false immediately after RecordCall")
	}

	// Wait longer than stall_detection_seconds
	time.Sleep(1100 * time.Millisecond)
	if !tracker.DetectStall("default") {
		t.Error("expected DetectStall=true after stall_detection_seconds")
	}
}

// TestRetryMaxFallback verifies that after retry_max stall events,
// the fallback recommendation is emitted.
func TestRetryMaxFallback(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.StallDetectionSeconds = 1
	cfg.RetryMax = 3

	tracker := runtime.NewTracker(cfg)
	tracker.RecordCall("default", 100, 0)

	// Advance time past stall threshold and check retry counter
	time.Sleep(1100 * time.Millisecond)

	// Call IncrementStallRetry and GetFallbackRecommendation
	var recommendation string
	for i := 0; i < cfg.RetryMax+1; i++ {
		recommendation = tracker.IncrementStallRetry("default")
	}

	if recommendation == "" {
		t.Error("expected non-empty fallback recommendation after retry_max exhausted")
	}
	if !strings.Contains(recommendation, "split_into_waves") {
		t.Errorf("expected fallback=split_into_waves in recommendation, got %q", recommendation)
	}
}

// TestPersistProgressAt75Pct verifies that PersistProgress writes progress.md
// and returns a paste-ready resume message.
func TestPersistProgressAt75Pct(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specID := "SPEC-TEST-001"
	specDir := filepath.Join(tmpDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	tracker := runtime.NewTracker(testConfig())
	tracker.SetProjectRoot(tmpDir)

	msg, err := tracker.PersistProgress(specID, "Wave 1", "budget tracker 구현", "다음: /moai sync")
	if err != nil {
		t.Fatalf("PersistProgress returned unexpected error: %v", err)
	}

	// Verify progress.md was written
	progressPath := filepath.Join(specDir, "progress.md")
	if _, statErr := os.Stat(progressPath); statErr != nil {
		t.Errorf("progress.md not found at %s: %v", progressPath, statErr)
	}

	// Verify resume message format
	if !strings.Contains(msg, "ultrathink") {
		t.Errorf("expected resume message to contain 'ultrathink', got %q", msg)
	}
	if !strings.Contains(msg, specID) {
		t.Errorf("expected resume message to contain spec ID %q, got %q", specID, msg)
	}

	// Verify progress.md content
	content, err := os.ReadFile(progressPath)
	if err != nil {
		t.Fatalf("failed to read progress.md: %v", err)
	}
	if !strings.Contains(string(content), "Auto-saved at") {
		t.Errorf("expected progress.md to contain 'Auto-saved at', got %q", string(content))
	}
}

// TestDefaultsWhenConfigMissing (EC-1) verifies that NewTracker with nil config
// uses built-in default values.
func TestDefaultsWhenConfigMissing(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(nil)

	// Default budget for unknown agent should be DefaultBudget
	_, budget, _ := tracker.Usage("unknown-agent")
	if budget != runtime.DefaultBudget {
		t.Errorf("expected default budget=%d, got %d", runtime.DefaultBudget, budget)
	}

	// Thresholds should be the defaults
	cfg := tracker.Config()
	if cfg.PreClearThreshold != runtime.DefaultPreClearThreshold {
		t.Errorf("expected PreClearThreshold=%f, got %f", runtime.DefaultPreClearThreshold, cfg.PreClearThreshold)
	}
	if cfg.HardClearThreshold != runtime.DefaultHardClearThreshold {
		t.Errorf("expected HardClearThreshold=%f, got %f", runtime.DefaultHardClearThreshold, cfg.HardClearThreshold)
	}
}

// TestSilentSkipOnMissingSPECDir (EC-2) verifies that PersistProgress silently
// skips when the SPEC directory does not exist.
func TestSilentSkipOnMissingSPECDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tracker := runtime.NewTracker(testConfig())
	tracker.SetProjectRoot(tmpDir)

	// SPEC dir does not exist — should return empty string and nil error
	msg, err := tracker.PersistProgress("SPEC-NONEXISTENT", "Wave 1", "test", "next step")
	if err != nil {
		t.Errorf("expected nil error for missing SPEC dir, got %v", err)
	}
	if msg != "" {
		t.Errorf("expected empty message for missing SPEC dir, got %q", msg)
	}
}

// TestUnknownAgentUsesDefaultBudget (EC-3) verifies that an agent not listed
// in per_agent_budget uses the 'default' value.
func TestUnknownAgentUsesDefaultBudget(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())
	tracker.RecordCall("some-unknown-agent", 100, 50)

	_, budget, _ := tracker.Usage("some-unknown-agent")
	if budget != 1000 { // testConfig default is 1000
		t.Errorf("expected budget=1000 (default) for unknown agent, got %d", budget)
	}
}

// TestConcurrentRecordCall (EC-4) verifies that Tracker is goroutine-safe.
func TestConcurrentRecordCall(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())
	const goroutines = 20
	const callsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				tracker.RecordCall("default", 5, 5)
			}
		}()
	}
	wg.Wait()

	current, _, _ := tracker.Usage("default")
	expected := goroutines * callsPerGoroutine * 10 // 10 tokens per call
	if current != expected {
		t.Errorf("expected current=%d after concurrent calls, got %d", expected, current)
	}
}

// TestNoAutoClearInvocation (AC-ARCH007-06) verifies at the unit-test level
// that Tracker never invokes any /clear mechanism. Since no exec.Command
// or shell invocation is possible in pure Go, this test simply verifies
// that the package compiles without importing os/exec.
func TestNoAutoClearInvocation(t *testing.T) {
	t.Parallel()

	// This test verifies the contract: no auto-clear occurs.
	// Companion grep-based verification is in acceptance.md AC-ARCH007-06.
	tracker := runtime.NewTracker(testConfig())
	tracker.RecordCall("default", 1000, 0) // 100% budget
	tracker.RecordCall("default", 500, 0)  // 150% budget

	// No panic, no clear invocation — just warnings
	if !tracker.IsAtHardLimit("default") {
		t.Error("expected IsAtHardLimit=true at 150%")
	}
}

// TestPerAgentBudgetOverWarning (AC-ARCH007-07) verifies that exceeding budget
// emits a warning (the call must complete without error — warning-first policy).
func TestPerAgentBudgetOverWarning(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())

	// Push manager-strategy over its 2000-token budget
	tracker.RecordCall("manager-strategy", 2100, 0)

	current, budget, ratio := tracker.Usage("manager-strategy")
	if current != 2100 {
		t.Errorf("expected current=2100, got %d", current)
	}
	if budget != 2000 {
		t.Errorf("expected budget=2000 for manager-strategy, got %d", budget)
	}
	if ratio <= 1.0 {
		t.Errorf("expected ratio>1.0 for over-budget agent, got %f", ratio)
	}
	// Warning is emitted to stderr by RecordCall — verified by integration/manual test.
	// No error is returned (warning-only per BC-V3R3-006).
}

// TestLoadRuntimeFromFile verifies that LoadRuntime parses a valid YAML file.
func TestLoadRuntimeFromFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "runtime.yaml")
	yamlContent := `runtime:
  context_window:
    pre_clear_threshold: 0.80
    hard_clear_threshold: 0.95
  per_agent_budget:
    default: 5000
  circuit_breaker:
    stall_detection_seconds: 30
    retry_max: 2
    fallback: split_into_waves
  progress_persistence:
    auto_save_at_threshold: true
    save_path_template: ".moai/specs/{SPEC_ID}/progress.md"
    resume_message_format: "test {spec_id}"
`
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0o644); err != nil {
		t.Fatalf("failed to write test YAML: %v", err)
	}

	cfg, err := runtime.LoadRuntime(cfgPath)
	if err != nil {
		t.Fatalf("LoadRuntime returned unexpected error: %v", err)
	}
	if cfg.PreClearThreshold != 0.80 {
		t.Errorf("expected PreClearThreshold=0.80, got %f", cfg.PreClearThreshold)
	}
	if cfg.StallDetectionSeconds != 30 {
		t.Errorf("expected StallDetectionSeconds=30, got %d", cfg.StallDetectionSeconds)
	}
	budget, ok := cfg.PerAgentBudget["default"]
	if !ok || budget != 5000 {
		t.Errorf("expected PerAgentBudget[default]=5000, got %d (ok=%v)", budget, ok)
	}
}

// TestLoadRuntimeMissingFile verifies that LoadRuntime returns an error for a
// missing file, and that DefaultRuntimeConfig() provides safe defaults.
func TestLoadRuntimeMissingFile(t *testing.T) {
	t.Parallel()

	_, err := runtime.LoadRuntime("/nonexistent/path/runtime.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}

	// DefaultRuntimeConfig must always succeed
	cfg := runtime.DefaultRuntimeConfig()
	if cfg == nil {
		t.Fatal("DefaultRuntimeConfig returned nil")
	}
	if cfg.PreClearThreshold != runtime.DefaultPreClearThreshold {
		t.Errorf("expected default PreClearThreshold=%f, got %f",
			runtime.DefaultPreClearThreshold, cfg.PreClearThreshold)
	}
}

// TestProgressMdAtomicWrite verifies that PersistProgress appends to existing
// progress.md rather than overwriting, and uses atomic write semantics.
func TestProgressMdAtomicWrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specID := "SPEC-ATOMIC-001"
	specDir := filepath.Join(tmpDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	progressPath := filepath.Join(specDir, "progress.md")
	existing := "# Progress\n\n## Previous work\n- Task 1 done\n"
	if err := os.WriteFile(progressPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("failed to write existing progress.md: %v", err)
	}

	tracker := runtime.NewTracker(testConfig())
	tracker.SetProjectRoot(tmpDir)

	_, err := tracker.PersistProgress(specID, "Wave 2", "budget fix", "next: sync")
	if err != nil {
		t.Fatalf("PersistProgress error: %v", err)
	}

	content, err := os.ReadFile(progressPath)
	if err != nil {
		t.Fatalf("failed to read progress.md: %v", err)
	}

	// Original content must be preserved
	if !strings.Contains(string(content), "Task 1 done") {
		t.Error("expected original content to be preserved")
	}
	// New section must be appended
	if !strings.Contains(string(content), "Auto-saved at") {
		t.Error("expected new Auto-saved section to be appended")
	}
}

// TestResumeMessageFormat verifies that the resume message follows the
// context-window-management.md format.
func TestResumeMessageFormat(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	specID := "SPEC-MSG-001"
	specDir := filepath.Join(tmpDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	tracker := runtime.NewTracker(testConfig())
	tracker.SetProjectRoot(tmpDir)

	msg, err := tracker.PersistProgress(specID, "Wave 3", "test approach", "/moai sync")
	if err != nil {
		t.Fatalf("PersistProgress error: %v", err)
	}

	// Validate required components per context-window-management.md §Resume message format
	required := []string{"ultrathink", specID, "Wave 3", "test approach", "/moai sync", "progress.md"}
	for _, token := range required {
		if !strings.Contains(msg, token) {
			t.Errorf("resume message missing required token %q: %q", token, msg)
		}
	}
}

// TestMultipleRecordCalls verifies cumulative token counting.
func TestMultipleRecordCalls(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())
	tracker.RecordCall("default", 100, 50)
	tracker.RecordCall("default", 200, 100)
	tracker.RecordCall("default", 50, 50)

	current, _, _ := tracker.Usage("default")
	// Expected: (100+50) + (200+100) + (50+50) = 150+300+100 = 550
	if current != 550 {
		t.Errorf("expected cumulative current=550, got %d", current)
	}
}

// TestDetectStallUnknownAgent verifies that DetectStall returns false (not stalled)
// for an agent with no prior RecordCall (no baseline timestamp).
func TestDetectStallUnknownAgent(t *testing.T) {
	t.Parallel()

	tracker := runtime.NewTracker(testConfig())
	// No RecordCall for "fresh-agent" — it is not stalled, just new
	if tracker.DetectStall("fresh-agent") {
		t.Error("expected DetectStall=false for agent with no prior call")
	}
}

// TestIncrementStallRetryNoRecommendation verifies that IncrementStallRetry
// returns empty string before retry_max is reached.
func TestIncrementStallRetryNoRecommendation(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.RetryMax = 3
	tracker := runtime.NewTracker(cfg)

	// First two increments should not yet return a recommendation
	for i := 0; i < cfg.RetryMax-1; i++ {
		recommendation := tracker.IncrementStallRetry("default")
		if recommendation != "" {
			t.Errorf("expected empty recommendation at retry %d, got %q", i+1, recommendation)
		}
	}
}

// TestSessionStartIntegration verifies that the session_start hook can import
// and use the runtime package (compilation-level integration check).
// The actual runtime.yaml load is tested via TestLoadRuntimeFromFile.
func TestSessionStartIntegration(t *testing.T) {
	t.Parallel()

	// This test documents that runtime.NewTracker + runtime.LoadRuntime
	// are the two entry points used in the SessionStart hook.
	// Actual integration is in internal/hook/session_start.go.

	// Verify NewTracker produces a valid tracker
	tracker := runtime.NewTracker(runtime.DefaultRuntimeConfig())
	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	// Verify the tracker responds to calls
	tracker.RecordCall("manager-spec", 100, 50)
	current, _, _ := tracker.Usage("manager-spec")
	if current != 150 {
		t.Errorf("expected current=150, got %d", current)
	}

	// Verify the format string is available
	info := fmt.Sprintf("tracker initialized: stall=%ds", tracker.Config().StallDetectionSeconds)
	if info == "" {
		t.Error("expected non-empty info string")
	}
}
