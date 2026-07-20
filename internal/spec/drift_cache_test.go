package spec

import (
	"os"
	"path/filepath"
	"testing"
)

// Unit tests for the HEAD-SHA drift cache (SPEC-SESSIONSTART-PERF-001 M1 —
// REQ-SSP-004 / REQ-SSP-006).
//
// The cache is FAIL-OPEN by contract: every failure mode must degrade to "recompute",
// never to an error and never to a wrong answer. These tests pin each failure path,
// because a cache that fails closed would break the session-start critical path it
// exists to protect.

// writeRawCache writes literal bytes to the cache path, bypassing saveDriftCache, so
// corrupt-file paths can be exercised.
func writeRawCache(t *testing.T, baseDir, content string) {
	t.Helper()

	path := driftCachePath(baseDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}
}

func TestDriftCache_RoundTrip(t *testing.T) {
	baseDir := t.TempDir()

	report := &DriftReport{
		Records: []DriftRecord{
			{SPECID: "SPEC-A-001", FrontmatterStatus: "completed", GitImpliedStatus: "completed", Drifted: false},
			{SPECID: "SPEC-B-001", FrontmatterStatus: "in-progress", GitImpliedStatus: "implemented", Drifted: true},
		},
		Count: 1,
	}

	saveDriftCache(baseDir, "head-abc", report)

	got, ok := loadDriftCache(baseDir, "head-abc")
	if !ok {
		t.Fatal("loadDriftCache after save = not ok, want a cache hit")
	}
	if got.Count != report.Count {
		t.Errorf("Count = %d, want %d", got.Count, report.Count)
	}
	if len(got.Records) != len(report.Records) {
		t.Fatalf("len(Records) = %d, want %d", len(got.Records), len(report.Records))
	}
	for i := range report.Records {
		if got.Records[i] != report.Records[i] {
			t.Errorf("Records[%d] = %+v, want %+v", i, got.Records[i], report.Records[i])
		}
	}
}

func TestDriftCache_FailOpenPaths(t *testing.T) {
	tests := []struct {
		name string
		// setup prepares the cache state; returns the head to query with.
		setup func(t *testing.T, baseDir string) string
	}{
		{
			name: "absent cache file",
			setup: func(_ *testing.T, _ string) string {
				return "head-abc"
			},
		},
		{
			name: "stale head key",
			setup: func(_ *testing.T, baseDir string) string {
				saveDriftCache(baseDir, "head-OLD", &DriftReport{Count: 42})
				return "head-NEW" // HEAD advanced since the entry was written
			},
		},
		{
			name: "corrupt JSON",
			setup: func(t *testing.T, baseDir string) string {
				writeRawCache(t, baseDir, "{ this is not valid json")
				return "head-abc"
			},
		},
		{
			name: "empty head key (non-git checkout)",
			setup: func(_ *testing.T, baseDir string) string {
				saveDriftCache(baseDir, "head-abc", &DriftReport{Count: 7})
				return "" // git rev-parse HEAD failed
			},
		},
		{
			name: "cache entry with empty head_sha",
			setup: func(t *testing.T, baseDir string) string {
				writeRawCache(t, baseDir, `{"head_sha":"","count":99,"records":[]}`)
				return "head-abc"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			head := tt.setup(t, baseDir)

			got, ok := loadDriftCache(baseDir, head)
			if ok {
				t.Errorf("loadDriftCache = (%+v, true), want a miss so the caller recomputes", got)
			}
		})
	}
}

// TestDriftCache_SaveIsBestEffort pins that a save which cannot proceed is silent —
// it must never panic or surface an error into drift detection.
func TestDriftCache_SaveIsBestEffort(t *testing.T) {
	baseDir := t.TempDir()

	// Empty head: nothing to key on, so the save is a no-op rather than a failure.
	saveDriftCache(baseDir, "", &DriftReport{Count: 1})
	if _, err := os.Stat(driftCachePath(baseDir)); !os.IsNotExist(err) {
		t.Error("saveDriftCache with an empty head wrote a cache file; want no-op")
	}

	// Nil report: also a no-op.
	saveDriftCache(baseDir, "head-abc", nil)
	if _, err := os.Stat(driftCachePath(baseDir)); !os.IsNotExist(err) {
		t.Error("saveDriftCache with a nil report wrote a cache file; want no-op")
	}
}

// TestDriftCachePath pins the cache location under the gitignored .moai/state/
// runtime-state directory (the same family as context-usage.json).
func TestDriftCachePath(t *testing.T) {
	got := driftCachePath("/project")
	want := filepath.Join("/project", ".moai", "state", driftCacheFilename)

	if got != want {
		t.Errorf("driftCachePath = %q, want %q", got, want)
	}
}
