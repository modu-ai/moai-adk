package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// The four freshness cases moved here with the freshness decision itself: they
// are exactly the inputs 'moai mx query' must rebuild on. Absent and stale were
// the only cases the hook-side gate was ever exercised against in production;
// corrupt and empty are the two that silently served an empty result set.

func TestMXIndexNeedsRebuild_Absent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if !mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=true when index absent")
	}
}

func TestMXIndexNeedsRebuild_Fresh(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	if err := mx.NewManager(stateDir).Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		ScannedAt:     time.Now(),
	}); err != nil {
		t.Fatalf("seed index: %v", err)
	}
	if mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=false when index is fresh")
	}
}

func TestMXIndexNeedsRebuild_Stale(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	if err := mx.NewManager(stateDir).Write(&mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		ScannedAt:     time.Now().Add(-(mxIndexFreshnessThreshold + time.Hour)),
	}); err != nil {
		t.Fatalf("seed stale index: %v", err)
	}
	if !mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=true when index is stale")
	}
}

func TestMXIndexNeedsRebuild_Corrupt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".moai", "state", mx.SidecarFileName), "{not valid json")
	if !mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=true when index is corrupt")
	}
}

func TestMXIndexNeedsRebuild_Empty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".moai", "state", mx.SidecarFileName), "")
	if !mxIndexNeedsRebuild(dir) {
		t.Fatal("expected mxIndexNeedsRebuild=true when index is empty")
	}
}

// TestBuildMXIndex_WritesTaggedIndex verifies the build writes a well-formed
// index carrying the tags found in the tree.
func TestBuildMXIndex_WritesTaggedIndex(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "svc", "worker.go"), "package svc\n\n"+
		"// @MX:NOTE: [AUTO] index build test marker\nfunc Foo() {}\n")

	if err := buildMXIndex(dir); err != nil {
		t.Fatalf("buildMXIndex: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".moai", "state", mx.SidecarFileName))
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
		t.Errorf("build did not index the NOTE tag; got %d tags", len(sc.Tags))
	}
}

// TestBuildMXIndex_WriteErrorIsLoud is the consumer-side inversion of the
// hook's fail-open contract. A background best-effort scan could swallow a
// write failure because nothing consumed its output in that process; here the
// caller serves query results from this index, so a swallowed failure would be
// served as an empty result set — a wrong answer rather than a degraded one.
func TestBuildMXIndex_WriteErrorIsLoud(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Make .moai/state a FILE so the sidecar directory cannot be created.
	writeFile(t, filepath.Join(dir, ".moai", "state"), "blocker")
	writeFile(t, filepath.Join(dir, "svc", "worker.go"), "package svc\n// @MX:NOTE: x\n")

	if err := buildMXIndex(dir); err == nil {
		t.Fatal("expected a loud error when the sidecar cannot be written, got nil")
	}
}

func TestBuildMXIndex_EmptyProjectDir(t *testing.T) {
	t.Parallel()
	if err := buildMXIndex(""); err == nil {
		t.Fatal("expected an error for an empty project directory, got nil")
	}
}
