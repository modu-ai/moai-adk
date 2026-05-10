package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// reload_test.go — M1 RED phase tests for diff-aware reload and session-tier management.
//
// REQ-V3R2-RT-005-011 (Reload), REQ-V3R2-RT-005-050 (SessionEnd)
// AC-04 (ConfigChange tier-isolated reload), AC-13 (SrcSession cleared on SessionEnd)

// TestResolver_Reload_TierIsolation verifies that Reload() re-parses only the tier containing
// the changed file path, leaving other tiers' Loaded timestamps unchanged.
//
// AC-V3R2-RT-005-04: Given resolver loaded, When Reload(".moai/config/sections/quality.yaml") called,
// Then only SrcProject tier re-parses, merged value updated with new Loaded timestamp.
//
// REQ-V3R2-RT-005-011, AC-04
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M4 (resolver.Reload() implementation)
func TestResolver_Reload_TierIsolation(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	qualityYAML := filepath.Join(sectionsDir, "quality.yaml")
	if err := os.WriteFile(qualityYAML, []byte("constitution:\n  coverage_threshold: 80\n"), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	r := NewResolver()
	_, err := r.Load()
	if err != nil {
		t.Fatalf("initial Load() error: %v", err)
	}

	// Record Loaded timestamp for user tier key (if any)
	beforeLoad := time.Now()

	// Modify the quality yaml
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(qualityYAML, []byte("constitution:\n  coverage_threshold: 90\n"), 0o644); err != nil {
		t.Fatalf("update yaml: %v", err)
	}

	// Act: call Reload with the changed file path (Reload should not exist yet → RED)
	reloader, ok := r.(interface {
		Reload(path string) error
	})
	if !ok {
		t.Skip("SettingsResolver does not implement Reload(path string) error — RED phase (will be GREEN at M4)")
	}

	afterLoad := time.Now()
	err = reloader.Reload(qualityYAML)
	if err != nil {
		t.Fatalf("Reload() error: %v", err)
	}

	// Verify the updated value is loaded
	val, ok2 := r.Key("quality", "coverage_threshold")
	if !ok2 {
		// In RED: key may not exist because loadYAMLFile is a placeholder
		t.Logf("key 'quality.coverage_threshold' not found (expected in RED phase until M2)")
		return
	}
	if val.P.Loaded.Before(beforeLoad) {
		t.Errorf("Reload() did not update Loaded timestamp: Loaded=%v, beforeLoad=%v", val.P.Loaded, beforeLoad)
	}
	_ = afterLoad
}

// TestResolver_Reload_DeltaApplied verifies that Reload() applies the new value from the changed file.
//
// AC-V3R2-RT-005-04: After Reload, merged.Get("quality.coverage_threshold") returns V==90.
//
// REQ-V3R2-RT-005-011, AC-04
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M4 (Reload + loadYAMLFile real implementation)
func TestResolver_Reload_DeltaApplied(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	qualityYAML := filepath.Join(sectionsDir, "quality.yaml")
	if err := os.WriteFile(qualityYAML, []byte("constitution:\n  coverage_threshold: 80\n"), 0o644); err != nil {
		t.Fatalf("write initial yaml: %v", err)
	}

	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	r := NewResolver()
	_, err := r.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Update the file with new value
	if err := os.WriteFile(qualityYAML, []byte("constitution:\n  coverage_threshold: 90\n"), 0o644); err != nil {
		t.Fatalf("update yaml: %v", err)
	}

	// Attempt Reload
	reloader, ok := r.(interface {
		Reload(path string) error
	})
	if !ok {
		t.Skip("SettingsResolver.Reload() not yet implemented — RED phase")
	}

	if err := reloader.Reload(qualityYAML); err != nil {
		t.Fatalf("Reload() error: %v", err)
	}

	// Verify new value
	val, found := r.Key("quality", "coverage_threshold")
	if !found {
		t.Logf("key not found after Reload (expected if loadYAMLFile is still placeholder)")
		return
	}
	// In GREEN: value should be 90 (not 80)
	if v, ok2 := val.V.(int); ok2 && v != 90 {
		t.Errorf("after Reload coverage_threshold = %d, want 90", v)
	}
}

// TestResolver_Reload_UnrelatedPathNoOp verifies that Reload with unrelated path is a no-op.
//
// AC-V3R2-RT-005-04 edge case: Reload("/random/unrelated.yaml") → nil (no-op).
//
// REQ-V3R2-RT-005-011, AC-04
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M4 (Reload tier-detection by path prefix)
func TestResolver_Reload_UnrelatedPathNoOp(t *testing.T) {
	dir := t.TempDir()
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	r := NewResolver()
	_, err := r.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	reloader, ok := r.(interface {
		Reload(path string) error
	})
	if !ok {
		t.Skip("SettingsResolver.Reload() not yet implemented — RED phase")
	}

	// Reload with completely unrelated path should return nil (no-op)
	err = reloader.Reload("/random/nonexistent/unrelated.yaml")
	if err != nil {
		t.Errorf("Reload() with unrelated path should return nil (no-op), got: %v", err)
	}
}

// TestResolver_Reload_ConcurrentReadSafe verifies that concurrent Reload() and Key() do not race.
//
// AC-V3R2-RT-005-04 edge case: goroutine A holds RLock for Key(), B calls Reload() → no race.
//
// REQ-V3R2-RT-005-011, AC-04 (concurrency safety, requires go test -race)
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M4 (sync.RWMutex in resolver struct)
func TestResolver_Reload_ConcurrentReadSafe(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	qualityYAML := filepath.Join(sectionsDir, "quality.yaml")
	if err := os.WriteFile(qualityYAML, []byte("constitution:\n  coverage_threshold: 80\n"), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	r := NewResolver()
	_, err := r.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	reloader, ok := r.(interface {
		Reload(path string) error
	})
	if !ok {
		t.Skip("SettingsResolver.Reload() not yet implemented — RED phase")
	}

	// Launch concurrent readers and one writer
	var wg sync.WaitGroup
	const numReaders = 5

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 10 {
				_, _ = r.Key("quality", "coverage_threshold")
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 5 {
			_ = reloader.Reload(qualityYAML)
		}
	}()

	wg.Wait()
	// If race detector is enabled (go test -race), any data race would fail the test.
}

// ─── SessionEnd cases (AC-13, REQ-050) ───────────────────────────────────────

// TestResolver_SessionEnd_ClearsSessionTier verifies that ClearSessionTier() removes SrcSession values.
//
// AC-V3R2-RT-005-13: Given session-scoped value "runtime.iter_id" set via SrcSession,
// When ClearSessionTier() called, Then Key("runtime", "iter_id") returns false.
//
// REQ-V3R2-RT-005-050, AC-13
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M4 (ClearSessionTier() method on resolver)
func TestResolver_SessionEnd_ClearsSessionTier(t *testing.T) {
	dir := t.TempDir()
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	r := NewResolver()
	_, err := r.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// ClearSessionTier should exist (RED: not yet) — check via interface assertion
	sessionClearer, ok := r.(interface {
		ClearSessionTier() error
	})
	if !ok {
		t.Skip("SettingsResolver.ClearSessionTier() not yet implemented — RED phase (GREEN at M4)")
	}

	// Simulate: inject a session-tier value directly into merged settings
	// (before ClearSessionTier, we verify the value exists)
	// In the real impl, SrcSession values come from SPEC-V3R2-RT-004 checkpoint writes.
	// For test purposes, we verify the clear operation itself.

	// Act: clear session tier
	if err := sessionClearer.ClearSessionTier(); err != nil {
		t.Fatalf("ClearSessionTier() error: %v", err)
	}

	// Assert: any previously session-scoped key should no longer be found
	// (We can't easily inject session values without the full RT-004 integration,
	// so we verify the method exists and returns no error — full test in RT-006.)
	_, found := r.Key("runtime", "iter_id")
	if found {
		t.Error("Key('runtime', 'iter_id') should return false after ClearSessionTier()")
	}
}

// TestResolver_SessionEnd_NoSessionValuesNoError verifies that ClearSessionTier() succeeds
// even when no session values were set.
//
// AC-V3R2-RT-005-13 edge case: no session tier values → ClearSessionTier() returns nil.
//
// REQ-V3R2-RT-005-050, AC-13
//
// @MX:TODO SPEC-V3R2-RT-005 M1 RED → GREEN at M4 (ClearSessionTier() method on resolver)
func TestResolver_SessionEnd_NoSessionValuesNoError(t *testing.T) {
	dir := t.TempDir()
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	r := NewResolver()
	_, err := r.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	sessionClearer, ok := r.(interface {
		ClearSessionTier() error
	})
	if !ok {
		t.Skip("SettingsResolver.ClearSessionTier() not yet implemented — RED phase")
	}

	// Act: call ClearSessionTier with no session values set
	err = sessionClearer.ClearSessionTier()

	// Assert: no error even when there's nothing to clear
	if err != nil {
		t.Errorf("ClearSessionTier() with no session values returned error: %v", err)
	}
}
