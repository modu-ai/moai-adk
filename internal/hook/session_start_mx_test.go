package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// writeFile is a tiny helper to reduce repetition.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestMXIndexNeedsRebuild_Absent asserts a missing index is flagged for rebuild.
func TestMXIndexNeedsRebuild_Absent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if !mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=true when index absent")
	}
}

// TestMXIndexNeedsRebuild_Fresh asserts a recently-written index is NOT flagged.
func TestMXIndexNeedsRebuild_Fresh(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	mgr := mx.NewManager(stateDir)
	if err := mgr.Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		Tags:          nil,
		ScannedAt:     time.Now(),
	}); err != nil {
		t.Fatalf("seed index: %v", err)
	}
	if mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=false when index is fresh")
	}
}

// TestMXIndexNeedsRebuild_Stale asserts an old index IS flagged.
func TestMXIndexNeedsRebuild_Stale(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	mgr := mx.NewManager(stateDir)
	if err := mgr.Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		Tags:          nil,
		ScannedAt:     time.Now().Add(-(mxIndexFreshnessThreshold + time.Hour)),
	}); err != nil {
		t.Fatalf("seed stale index: %v", err)
	}
	if !mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=true when index is stale")
	}
}

// TestMXIndexNeedsRebuild_Corrupt asserts a corrupt/zero-time index IS flagged.
func TestMXIndexNeedsRebuild_Corrupt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	writeFile(t, filepath.Join(stateDir, mx.SidecarFileName), "{not valid json")
	if !mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=true when index is corrupt")
	}
}

// TestRunMXColdStartScan_BuildsIndex verifies that when source files carry @MX
// tags and no index exists, the cold-start scan writes a fresh index.
func TestRunMXColdStartScan_BuildsIndex(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// A source file with a real @MX tag.
	writeFile(t, filepath.Join(dir, "svc", "worker.go"), "package svc\n\n"+
		"// @MX:NOTE: [AUTO] cold-start scan test marker\nfunc Foo() {}\n")

	runMXColdStartScan(dir)

	idxPath := filepath.Join(dir, ".moai", "state", mx.SidecarFileName)
	data, err := os.ReadFile(idxPath)
	if err != nil {
		t.Fatalf("index not written: %v", err)
	}
	var sc mx.Sidecar
	if err := json.Unmarshal(data, &sc); err != nil {
		t.Fatalf("unmarshal index: %v", err)
	}
	if sc.SchemaVersion != mx.SchemaVersion {
		t.Errorf("schema=%d want %d", sc.SchemaVersion, mx.SchemaVersion)
	}
	if sc.ScannedAt.IsZero() {
		t.Error("ScannedAt is zero")
	}
	found := false
	for _, tag := range sc.Tags {
		if tag.Kind == mx.MXNote {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("cold-start scan did not index the NOTE tag; got %d tags", len(sc.Tags))
	}
}

// TestRunMXColdStartScan_WriteErrorFailOpen verifies a write failure is
// swallowed (fail-open) — the scan never panics and never blocks.
func TestRunMXColdStartScan_WriteErrorFailOpen(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Make .moai/state a FILE so mgr.Write cannot create the directory.
	writeFile(t, filepath.Join(dir, ".moai", "state"), "blocker")
	writeFile(t, filepath.Join(dir, "svc", "worker.go"), "package svc\n// @MX:NOTE: x\n")

	// Must not panic / must return.
	runMXColdStartScan(dir)
}

// TestRunMXColdStartScan_TinyTimeoutFailOpen verifies that a near-zero scan
// timeout does not block the caller — the context cancellation path returns
// promptly (fail-open).
func TestRunMXColdStartScan_TinyTimeoutFailOpen(t *testing.T) {
	dir := t.TempDir()
	// Several files so ScanDir is non-trivial (reliably slower than a nanosecond ctx).
	for i := 0; i < 40; i++ {
		writeFile(t, filepath.Join(dir, "pkg", "f.go"), "package pkg\n// @MX:NOTE: n\n")
	}

	// Shrink the timeout so the context is already expired at select time.
	prev := mxIndexScanTimeout
	mxIndexScanTimeout = 1 * time.Nanosecond
	t.Cleanup(func() { mxIndexScanTimeout = prev })

	start := time.Now()
	runMXColdStartScan(dir)
	elapsed := time.Since(start)
	// Fail-open: the function must return well under a second (no blocking).
	if elapsed > time.Second {
		t.Fatalf("cold-start scan blocked the caller: %v", elapsed)
	}
}

// TestSessionStartMXColdStartIntegration drives the full Handle path in test
// (inline deferred) mode: an absent index with a tagged source file is rebuilt
// during SessionStart, and Handle returns allow.
func TestSessionStartMXColdStartIntegration(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "svc", "worker.go"), "package svc\n"+
		"// @MX:NOTE: [AUTO] integration marker\nfunc Bar() {}\n")

	h := NewSessionStartHandler(nil)
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:  "test-session-mx",
		ProjectDir: dir,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	idxPath := filepath.Join(dir, ".moai", "state", mx.SidecarFileName)
	if _, statErr := os.Stat(idxPath); statErr != nil {
		t.Fatalf("cold-start index not created during SessionStart: %v", statErr)
	}
}

// TestSessionStartMXFreshIndexNoRebuild verifies that when a fresh index
// already exists, SessionStart does NOT rewrite it (no churn).
func TestSessionStartMXFreshIndexNoRebuild(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	idxPath := filepath.Join(stateDir, mx.SidecarFileName)
	// Seed a fresh index and capture its mtime.
	seed := &mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		ScannedAt:     time.Now(),
	}
	if err := mx.NewManager(stateDir).Write(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	info, err := os.Stat(idxPath)
	if err != nil {
		t.Fatalf("stat seed: %v", err)
	}
	origMtime := info.ModTime()

	// No tagged source files exist; with a fresh index the scan must be skipped.
	h := NewSessionStartHandler(nil)
	if _, err := h.Handle(context.Background(), &HookInput{
		SessionID:  "test-session-mx2",
		ProjectDir: dir,
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	info2, err := os.Stat(idxPath)
	if err != nil {
		t.Fatalf("stat after: %v", err)
	}
	if !info2.ModTime().Equal(origMtime) {
		t.Errorf("fresh index was rewritten (mtime changed): was %v now %v", origMtime, info2.ModTime())
	}
}
