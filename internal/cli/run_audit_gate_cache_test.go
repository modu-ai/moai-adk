//go:build integration

// Package cli provides integration tests for audit gate 24h cache behavior.
//
// AC-WAG-09: cache hit within 24h reuses verdict; cache miss on artifact change.
// Run with: go test -tags=integration -race ./internal/cli/... -run TestAuditCache
package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// TestAuditCacheHitWithin24Hours verifies AC-WAG-09:
// A PASS verdict cached within 24h is reused without re-invoking plan-auditor.
//
// Given: SPEC-CACHE-001 with prior PASS recorded T0-1h
// When: /moai run SPEC-CACHE-001 is called at T0 (1h later, within 24h)
// Then: plan-auditor is NOT called; cache hit is recorded.
//
// AC: AC-WAG-09
// REQ: REQ-WAG-003 (non-functional)
func TestAuditCacheHitWithin24Hours(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-PASS-001")

	t0 := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	cacheTime := t0.Add(-1 * time.Hour) // 1h ago

	cache := runtime.NewInMemoryCache()

	// Compute hash and seed the cache with a PASS result from 1h ago.
	hash, err := cache.ComputeHash(specDir)
	if err != nil {
		t.Fatalf("ComputeHash: %v", err)
	}
	cache.Store("SPEC-DUMMY-PASS-001", hash, &runtime.AuditResult{
		Verdict:        runtime.VerdictPass,
		AuditAt:        cacheTime,
		AuditorVersion: "plan-auditor/v1-cached",
		SpecID:         "SPEC-DUMMY-PASS-001",
	})

	auditor := &deterministicAuditor{verdict: runtime.VerdictPass}

	gate := &runtime.GateConfig{
		SpecID:     "SPEC-DUMMY-PASS-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    auditor,
		Cache:      cache,
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: t0},
		UserName:   "test-user",
	}

	result, err := gate.Invoke(context.Background())
	if err != nil {
		t.Fatalf("Invoke(): %v", err)
	}

	// Verify cache hit.
	if !result.CacheHit {
		t.Error("CacheHit = false, want true (AC-WAG-09)")
	}
	if result.Verdict != runtime.VerdictPass {
		t.Errorf("Verdict = %q, want PASS on cache hit (AC-WAG-09)", result.Verdict)
	}

	// Verify auditor was NOT called.
	if auditor.callCount > 0 {
		t.Errorf("auditor called %d times on cache hit, want 0 (AC-WAG-09)", auditor.callCount)
	}

	if !result.CachedAuditAt.Equal(cacheTime) {
		t.Errorf("CachedAuditAt = %v, want %v", result.CachedAuditAt, cacheTime)
	}
}

// TestAuditCacheInvalidatedOnPlanArtifactChange verifies AC-WAG-09:
// Cache is invalidated when plan artifacts change (hash differs).
//
// Given: cache has PASS for hash H0
// When: spec.md is modified (hash becomes H1)
// Then: plan-auditor is re-invoked (cache miss)
//
// AC: AC-WAG-09
// REQ: REQ-WAG-003 (non-functional)
func TestAuditCacheInvalidatedOnPlanArtifactChange(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	// Create a mutable spec dir for this test.
	specDir := t.TempDir()
	specContent := "# Test\n\n### REQ-001 (Ubiquitous)\nThe system shall do X.\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile spec.md: %v", err)
	}

	t0 := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	cache := runtime.NewInMemoryCache()

	// Seed cache with the original hash.
	hash0, err := cache.ComputeHash(specDir)
	if err != nil {
		t.Fatalf("ComputeHash initial: %v", err)
	}
	cache.Store("SPEC-CHANGE-001", hash0, &runtime.AuditResult{
		Verdict:        runtime.VerdictPass,
		AuditAt:        t0.Add(-30 * time.Minute),
		AuditorVersion: "plan-auditor/v1",
	})

	// Modify spec.md to change the hash.
	modifiedContent := specContent + "\n### REQ-002 (Ubiquitous)\nThe system shall also do Y.\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(modifiedContent), 0o644); err != nil {
		t.Fatalf("WriteFile modified spec.md: %v", err)
	}

	auditor := &deterministicAuditor{verdict: runtime.VerdictPass}

	gate := &runtime.GateConfig{
		SpecID:     "SPEC-CHANGE-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    auditor,
		Cache:      cache,
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: t0},
		UserName:   "test-user",
	}

	result, _ := gate.Invoke(context.Background())

	// Cache should have been invalidated — auditor is called.
	if auditor.callCount == 0 {
		t.Error("auditor was not called after artifact change — cache was not invalidated (AC-WAG-09)")
	}
	if result.CacheHit {
		t.Error("CacheHit = true after artifact change, want false (AC-WAG-09)")
	}
}
