package guardstate

import (
	"os"
	"path/filepath"
	"testing"
)

// repoRoot walks up from the test's working directory to the module root. The
// census criteria are about THIS repository's manifest against THIS
// repository's workflow files, so a repo-relative resolution is the subject
// rather than a convenience — but it is resolved by finding go.mod rather than
// by counting `..` segments, which is what makes it survive a package move.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod above %s", dir)
		}
		dir = parent
	}
}

func loadRepoManifest(t *testing.T) (*Manifest, string) {
	t.Helper()
	root := repoRoot(t)
	m, err := LoadManifest(filepath.Join(root, ManifestPath))
	if err != nil {
		t.Fatalf("load %s: %v", ManifestPath, err)
	}
	return m, root
}

// AC-GSM-001 — the manifest exists and its census is complete: an entry exists
// for every workflow file, with an EMPTY SET-DIFFERENCE IN BOTH DIRECTIONS.
//
// Mutant this kills: a count-equality check is satisfied by 18 entries all
// naming the same file. The assertion below is a set comparison in both
// directions for exactly that reason — the duplicate-naming mutant collapses
// the manifest-side set and leaves 17 files undeclared.
func TestCensus_SetDifferenceEmptyBothDirections(t *testing.T) {
	m, root := loadRepoManifest(t)

	onDisk := map[string]bool{}
	for _, pat := range []string{"*.yml", "*.yaml"} {
		matches, err := filepath.Glob(filepath.Join(root, ".github", "workflows", pat))
		if err != nil {
			t.Fatalf("glob %s: %v", pat, err)
		}
		for _, abs := range matches {
			rel, err := filepath.Rel(root, abs)
			if err != nil {
				t.Fatalf("rel %s: %v", abs, err)
			}
			onDisk[filepath.ToSlash(rel)] = true
		}
	}
	if len(onDisk) == 0 {
		t.Fatalf("the disk enumeration returned nothing; an empty comparison would pass vacuously")
	}

	declared := map[string]bool{}
	for _, e := range m.Entries {
		if declared[e.Locator] {
			t.Fatalf("locator %q is declared twice; a duplicate collapses the manifest-side set and is the count-equality mutant", e.Locator)
		}
		declared[e.Locator] = true
	}

	for path := range onDisk {
		if !declared[path] {
			t.Errorf("disk\\manifest: %s exists on disk with no manifest entry", path)
		}
	}
	for path := range declared {
		if !onDisk[path] {
			t.Errorf("manifest\\disk: entry names %s, which is not on disk", path)
		}
	}
	if len(declared) != len(onDisk) {
		t.Errorf("declared %d entries against %d workflow files", len(declared), len(onDisk))
	}
}

// AC-GSM-002 (a) over the populated census — no entry is missing any of the
// five fields. The per-field named rejection is in manifest_test.go; this is
// the same check applied to the shipped artifact rather than to a fixture.
func TestCensus_EveryEntryCarriesFiveFields(t *testing.T) {
	m, _ := loadRepoManifest(t)
	if len(m.Entries) == 0 {
		t.Fatalf("the census is empty; per-entry assertions would pass vacuously")
	}
	for _, e := range m.Entries {
		if err := e.Validate(); err != nil {
			t.Errorf("entry %q: %v", e.Locator, err)
		}
	}
}

// AC-GSM-004 — a legitimately quiet subject declares its condition. Both
// release-cycle subjects carry an explicit conditional expectation and neither
// is absent. Omission is the tempting shortcut for exactly these two entries,
// and omitting them would also make state-table row 5 unreachable.
func TestCensus_ReleaseOnlySubjectsDeclareTheirCondition(t *testing.T) {
	m, root := loadRepoManifest(t)

	subjects := []string{
		".github/workflows/spec-status-auto-sync.yml",
		".github/workflows/release.yml",
	}
	byLocator := map[string]Entry{}
	for _, e := range m.Entries {
		byLocator[e.Locator] = e
	}
	for _, s := range subjects {
		if _, err := os.Stat(filepath.Join(root, s)); err != nil {
			t.Fatalf("subject %s is not on disk: %v — the criterion's premise has moved", s, err)
		}
		e, ok := byLocator[s]
		if !ok {
			t.Errorf("%s has no manifest entry; a quiet subject must be DECLARED quiet, not omitted", s)
			continue
		}
		if !e.IsConditional() {
			t.Errorf("%s carries no expected_when; without it row 5 collapses into row 6 and a correctly-quiet subject is reported as an anomaly on every sweep", s)
		}
	}
}
