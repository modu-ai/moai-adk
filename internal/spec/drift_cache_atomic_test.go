package spec

// drift_cache_atomic_test.go — AC-DCF-004(b): the drift cache write is atomic
// THROUGH internal/atomicfile.Replace, not a hand-rolled temp-plus-rename.
//
// Why this matters here and not elsewhere: the out-of-band fill child is
// deliberately bounded by a deadline engineered to fire NEAR completion, which
// is exactly the moment a non-atomic os.WriteFile of an ~869-record JSON file
// is most likely to be in flight. loadDriftCache fails open on a truncated
// file, but a torn cache is prevented rather than tolerated.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSaveDriftCacheLeavesPreviousFileIntactOnInterruptedWrite asserts the
// behavioural half: when the final replace step fails (the write is
// interrupted at the one point where a non-atomic writer would already have
// truncated the destination), the PREVIOUS cache file is byte-identical and no
// temporary artefact is left behind.
func TestSaveDriftCacheLeavesPreviousFileIntactOnInterruptedWrite(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}

	// A complete previous cache, written through the real path.
	saveDriftCache(dir, "head-A", &DriftReport{Records: []DriftRecord{{SPECID: "SPEC-A-001"}}, Count: 1})
	before, err := os.ReadFile(driftCachePath(dir))
	if err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	// Interrupt the write exactly at the replace step.
	orig := driftCacheReplaceFn
	driftCacheReplaceFn = func(string, string) error { return os.ErrPermission }
	t.Cleanup(func() { driftCacheReplaceFn = orig })

	saveDriftCache(dir, "head-B", &DriftReport{Records: []DriftRecord{{SPECID: "SPEC-B-001"}}, Count: 99})

	after, err := os.ReadFile(driftCachePath(dir))
	if err != nil {
		t.Fatalf("read cache after interrupted write: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("interrupted write mutated the previous cache file\nbefore:\n%s\nafter:\n%s", before, after)
	}

	entries, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("read state dir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != driftCacheFilename {
			t.Fatalf("interrupted write left a residual artefact: %q", e.Name())
		}
	}
}

// TestSaveDriftCacheLeavesNoTempArtefactOnSuccess is the positive control: a
// successful write leaves the cache and nothing else.
func TestSaveDriftCacheLeavesNoTempArtefactOnSuccess(t *testing.T) {
	dir := t.TempDir()
	saveDriftCache(dir, "head-A", &DriftReport{Records: []DriftRecord{{SPECID: "SPEC-A-001"}}, Count: 1})

	raw, err := os.ReadFile(driftCachePath(dir))
	if err != nil {
		t.Fatalf("read cache: %v", err)
	}
	var payload driftCacheFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("cache does not parse: %v", err)
	}
	if payload.HeadSHA != "head-A" {
		t.Fatalf("head_sha = %q, want %q", payload.HeadSHA, "head-A")
	}

	entries, err := os.ReadDir(filepath.Join(dir, ".moai", "state"))
	if err != nil {
		t.Fatalf("read state dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != driftCacheFilename {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("state dir = %v, want exactly [%s]", names, driftCacheFilename)
	}
}

// TestSaveDriftCacheRoutesThroughAtomicfileReplace is the structural half of
// AC-DCF-004(b) and its continued-firing carrier: the assertion is on the
// WRITER itself, so replacing it with a plain in-place os.WriteFile turns this
// test red rather than quietly removing a case.
func TestSaveDriftCacheRoutesThroughAtomicfileReplace(t *testing.T) {
	src, err := os.ReadFile("drift_cache.go")
	if err != nil {
		t.Fatalf("read drift_cache.go: %v", err)
	}
	body := string(src)

	if !strings.Contains(body, "atomicfile.Replace") {
		t.Fatal("drift_cache.go does not reference atomicfile.Replace — the cache write is not routed through the in-tree atomic primitive")
	}
	// The destination must never be written in place. A temp file inside the
	// same directory is fine; `os.WriteFile(path,` — the destination itself —
	// is the non-atomic shape this criterion excludes.
	if strings.Contains(body, "os.WriteFile(path,") {
		t.Fatal("drift_cache.go writes the destination in place with os.WriteFile — a torn cache is possible")
	}
}
