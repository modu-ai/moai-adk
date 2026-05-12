// Package template — SPEC-V3R4-CATALOG-002 audit suite.
//
// Audit invariants for the SlimFS layer over the REAL production catalog
// (loaded from the embedded catalog.yaml). Complementary to slim_fs_test.go
// which uses synthetic catalogs for unit-level coverage.
//
// REQ-014 / REQ-015 / REQ-016 / REQ-017 / REQ-003 invariants are enforced
// here as audit tests. All sentinel emissions use t.Errorf (NOT t.Logf) —
// CATALOG-001 eval-1 EC3 hash sentinel lesson.
//
// @MX:NOTE: [AUTO] CATALOG-002 audit suite — REQ-014/015/016/017/003 invariants over the real production catalog.
package template

import (
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// loadSlimFS loads the embedded catalog and wraps embeddedRaw in SlimFS.
// Returns both the wrapped fs.FS (caller-view, "templates/"-prefix stripped)
// and the underlying Catalog for cross-checks.
//
// Audit tests use the real production catalog — they verify the slim filter
// against the actual 65-entry manifest, not a synthetic one.
func loadSlimFS(t *testing.T) (fs.FS, *Catalog) {
	t.Helper()
	cat, err := LoadCatalog(embeddedRaw)
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	slim, err := SlimFS(embeddedRaw, cat)
	if err != nil {
		t.Fatalf("wrap SlimFS: %v", err)
	}
	return slim, cat
}

// TestSlimFS_HidesNonCoreEntries verifies REQ-014: every non-core catalog
// entry MUST be invisible through the slim FS (Stat returns fs.ErrNotExist).
//
// Non-core entries include optional-pack:* tiers and harness-generated.
// Expected hidden count: 25 (17 optional skills + 7 optional agents + 1 harness-generated).
//
// Sentinel: CATALOG_SLIM_LEAK
func TestSlimFS_HidesNonCoreEntries(t *testing.T) {
	t.Parallel()
	slim, cat := loadSlimFS(t)

	audited := 0
	for _, e := range cat.AllEntries() {
		if e.Tier == TierCore {
			continue
		}
		// entry.Path = "templates/.claude/skills/foo/" → caller view = ".claude/skills/foo/"
		callerPath := strings.TrimPrefix(strings.TrimSuffix(e.Path, "/"), "templates/")
		_, err := fs.Stat(slim, callerPath)
		if err == nil {
			t.Errorf("CATALOG_SLIM_LEAK: %s tier=%s (Stat succeeded, expected fs.ErrNotExist)", callerPath, e.Tier)
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("CATALOG_SLIM_LEAK: %s tier=%s (got err=%v, want fs.ErrNotExist)", callerPath, e.Tier, err)
		}
		audited++
	}
	t.Logf("audited %d non-core entries", audited)
}

// TestSlimFS_PreservesCoreEntries verifies REQ-015: every core catalog entry
// MUST be reachable through the slim FS (Stat returns no error).
//
// EC4 sub-assertion: a nested path inside a core skill directory must also
// be reachable, validating that prefix-based filtering does not over-block.
//
// Expected reachable count: 40 (20 core skills + 20 core agents) at directory
// granularity, plus the nested EC4 path.
//
// Sentinel: CATALOG_SLIM_CORE_MISSING
func TestSlimFS_PreservesCoreEntries(t *testing.T) {
	t.Parallel()
	slim, cat := loadSlimFS(t)

	audited := 0
	for _, e := range cat.AllEntries() {
		if e.Tier != TierCore {
			continue
		}
		callerPath := strings.TrimPrefix(strings.TrimSuffix(e.Path, "/"), "templates/")
		_, err := fs.Stat(slim, callerPath)
		if err != nil {
			t.Errorf("CATALOG_SLIM_CORE_MISSING: %s (got err=%v, want success)", callerPath, err)
		}
		audited++
	}
	t.Logf("audited %d core entries", audited)

	// EC4 sub-assertion: nested path under a core skill.
	// The "moai" skill is tier=core; its sub-file "workflows/plan.md" must be reachable.
	t.Run("nested_moai_workflows_plan", func(t *testing.T) {
		const nestedPath = ".claude/skills/moai/workflows/plan.md"
		_, err := fs.Stat(slim, nestedPath)
		if err != nil {
			t.Errorf("CATALOG_SLIM_CORE_MISSING: nested %s: %v", nestedPath, err)
		}
	})
}

// TestSlimFS_PreservesNonCatalogFiles verifies REQ-016: files that are NOT
// enumerated in catalog.yaml's skills/agents sections (e.g. rules, config,
// root files) must remain reachable through the slim FS.
//
// These paths are outside the catalog's skill/agent lists and must pass
// through the filter unchanged.
//
// T3.4 path verification (pre-checked against templates/ tree):
//   - .claude/rules/moai/core/zone-registry.md  → OK
//   - .claude/output-styles/moai/moai.md         → OK
//   - .moai/config/sections/harness.yaml         → OK
//   - CLAUDE.md                                  → OK
//   - .gitignore                                 → OK
//
// Note: .moai/config/sections/quality.yaml is only present as quality.yaml.tmpl
// in the templates tree; harness.yaml is used instead.
//
// Sentinel: CATALOG_SLIM_OVER_FILTER
func TestSlimFS_PreservesNonCatalogFiles(t *testing.T) {
	t.Parallel()
	slim, _ := loadSlimFS(t)

	// Non-catalog paths — these are NOT enumerated in catalog.yaml's
	// skills+agents sections but are part of the templates/ tree.
	nonCatalogPaths := []string{
		".claude/rules/moai/core/zone-registry.md",
		".claude/output-styles/moai/moai.md",
		".moai/config/sections/harness.yaml",
		"CLAUDE.md",
		".gitignore",
	}

	for _, p := range nonCatalogPaths {
		_, err := fs.Stat(slim, p)
		if err != nil {
			t.Errorf("CATALOG_SLIM_OVER_FILTER: %s (got err=%v, want success)", p, err)
		}
	}
}

// TestSlimFS_WalkDirNoLeak verifies REQ-017: a full walk of the slim FS must
// not encounter any path that matches a denied (non-core) catalog entry.
//
// The deny set is reconstructed in caller-view namespace ("templates/" stripped)
// and checked against every visited path using both exact and prefix matching.
//
// Sentinel: CATALOG_SLIM_WALK_LEAK
func TestSlimFS_WalkDirNoLeak(t *testing.T) {
	t.Parallel()
	slim, cat := loadSlimFS(t)

	// Reconstruct deny set in caller-view namespace (templates/ stripped).
	denyCallerView := make(map[string]struct{})
	for _, e := range cat.AllEntries() {
		if e.Tier == TierCore {
			continue
		}
		callerPath := strings.TrimPrefix(strings.TrimSuffix(e.Path, "/"), "templates/")
		denyCallerView[callerPath] = struct{}{}
	}

	visited := 0
	err := fs.WalkDir(slim, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		visited++
		// Verify path does NOT match any deny entry as prefix.
		for denied := range denyCallerView {
			if path == denied {
				t.Errorf("CATALOG_SLIM_WALK_LEAK: %s (exact match with denied %s)", path, denied)
				return nil
			}
			if strings.HasPrefix(path, denied+"/") {
				t.Errorf("CATALOG_SLIM_WALK_LEAK: %s (prefix match with denied %s)", path, denied)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if visited == 0 {
		t.Errorf("CATALOG_SLIM_WALK_LEAK: walk visited 0 paths (slim FS appears empty)")
	}
	t.Logf("walk visited %d paths", visited)
}

// TestSlimFS_ReadOnlyInvariant verifies REQ-003: the slimFS wrapper must be
// structurally read-only (no sync.* fields, no channels) and must handle
// concurrent reads without data races.
//
// Two sub-tests:
//   (a) reflective_struct_check — inspects slimFS field types via reflect.
//   (b) concurrent_reads_race_clean — 32-goroutine fan-out to expose races
//       when run with `go test -race`.
func TestSlimFS_ReadOnlyInvariant(t *testing.T) {
	t.Parallel()

	t.Run("reflective_struct_check", func(t *testing.T) {
		// Inspect slimFS struct type at the type level — no instance needed
		// because the read-only invariant is a property of the struct shape,
		// not of any particular constructed value.
		ty := reflect.TypeFor[slimFS]()
		for f := range ty.Fields() {
			// No sync.* types — slimFS must be stateless and goroutine-safe
			// without internal locks (immutable-after-construction design).
			if strings.HasPrefix(f.Type.String(), "sync.") {
				t.Errorf("CATALOG_SLIM_NOT_READONLY: field=%s type=%s (sync.* forbidden)", f.Name, f.Type)
			}
			// No channels — synchronous, non-blocking operations only.
			if f.Type.Kind() == reflect.Chan {
				t.Errorf("CATALOG_SLIM_NOT_READONLY: field=%s type=%s (chan forbidden)", f.Name, f.Type)
			}
		}
		// denySet immutability is documented (godoc-only invariant per P3-3 finding):
		// "denySet is set once at construction and used read-only thereafter."
	})

	// 32 goroutines × 50 iterations gives ~1600 concurrent reads, sufficient to
	// surface most data races in production wrappers. See spec.md EC5(b) and
	// the discussion of goroutine count rationale.
	t.Run("concurrent_reads_race_clean", func(t *testing.T) {
		slim, cat := loadSlimFS(t)

		// Build a path pool from real catalog entries (mix of core + non-core).
		var pool []string
		for _, e := range cat.AllEntries() {
			callerPath := strings.TrimPrefix(strings.TrimSuffix(e.Path, "/"), "templates/")
			pool = append(pool, callerPath)
		}
		if len(pool) == 0 {
			t.Fatalf("empty path pool")
		}

		const goroutines = 32
		const iterations = 50
		var wg sync.WaitGroup
		var errCount int64
		wg.Add(goroutines)
		for g := range goroutines {
			go func(seed int) {
				defer wg.Done()
				for i := range iterations {
					path := pool[(seed*31+i)%len(pool)]
					_, statErr := fs.Stat(slim, path)
					_ = statErr // ignore — Stat may return ErrNotExist for non-core, both are fine
					if i%5 == 0 {
						_ = fs.WalkDir(slim, ".", func(p string, d fs.DirEntry, walkErr error) error {
							if walkErr != nil {
								atomic.AddInt64(&errCount, 1)
								return walkErr
							}
							// briefly descend into directories
							return nil
						})
					}
					if i%3 == 0 {
						if dirEntries, dirErr := fs.ReadDir(slim, "."); dirErr == nil {
							_ = dirEntries
						}
					}
				}
			}(g)
		}
		wg.Wait()
		if atomic.LoadInt64(&errCount) > 0 {
			t.Errorf("CATALOG_SLIM_NOT_READONLY: %d errors during concurrent reads (potential data race or state mutation)", errCount)
		}
		// The real race detector verdict comes from `go test -race`. This test
		// ensures the race detector has enough concurrent traffic to surface
		// any data race. If -race is not enabled, this test still passes —
		// it serves as a load test for the wrapper.
	})
}
