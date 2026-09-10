package mx

import (
	"os"
	"path/filepath"
	"testing"
)

// fpFixtureWrite writes a repo-relative slash path under root.
func fpFixtureWrite(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// AC-GFC-004 — the codemaps content fingerprint's producer (StampCodemaps)
// and its consumer (the recompute checkCodemaps compares against) apply the
// SAME predicate, so a dirty-stamped tree is judged against a fingerprint
// computed the way it was stamped.
//
// The consumer side is reproduced here as the equality test checkCodemaps
// performs (check.go recomputes the aggregate and compares it to the stored
// ContentFingerprint); internal/graph imports this package, so the check
// cannot be called from here without an import cycle.
//
// Mutant this kills: a filtered checker paired with an unfiltered codemaps
// stamp writer — the first assertion fails immediately, because the tree is
// stale against its own fresh stamp.
func TestCodemapsFingerprint_ProducerConsumer(t *testing.T) {
	root := t.TempDir()
	stampGit(t, root, "init", "-q")
	stampGit(t, root, "config", "user.email", "fixture@example.com")
	stampGit(t, root, "config", "user.name", "Fixture")
	fpFixtureWrite(t, root, "internal/alpha/alpha.go", "package alpha\n")
	fpFixtureWrite(t, root, "internal/alpha/testdata/fixture.go", "package testdata\n")
	fpFixtureWrite(t, root, "internal/alpha/alpha_test.go", "package alpha\n")
	stampGit(t, root, "add", "-A")
	stampGit(t, root, "commit", "-q", "-m", "base")

	// Uncommitted described-source change → the stamp takes the dirty branch.
	fpFixtureWrite(t, root, "internal/alpha/alpha.go", "package alpha\n\nfunc A() {}\n")

	pv, err := StampCodemaps(root, "")
	if err != nil {
		t.Fatalf("StampCodemaps: %v", err)
	}
	if !pv.Dirty || pv.ContentFingerprint == "" {
		t.Fatalf("fixture did not produce a dirty stamp: dirty=%v fingerprint=%q", pv.Dirty, pv.ContentFingerprint)
	}

	consumer := func() string {
		t.Helper()
		fp, err := AggregateDescribedFingerprintFiltered(root, DefaultDescribedRoots)
		if err != nil {
			t.Fatalf("AggregateDescribedFingerprintFiltered: %v", err)
		}
		return fp
	}

	// Read back with no intervening edit → fresh.
	if consumer() != pv.ContentFingerprint {
		t.Fatalf("stale against its own fresh stamp — the producer and the consumer disagree")
	}

	// A testdata edit is not described-worthy → still fresh.
	fpFixtureWrite(t, root, "internal/alpha/testdata/fixture.go", "package testdata // edited\n")
	if consumer() != pv.ContentFingerprint {
		t.Errorf("a testdata edit moved the codemaps fingerprint")
	}

	// A _test.go edit is not described-worthy → still fresh.
	fpFixtureWrite(t, root, "internal/alpha/alpha_test.go", "package alpha // edited\n")
	if consumer() != pv.ContentFingerprint {
		t.Errorf("a _test.go edit moved the codemaps fingerprint")
	}

	// A production .go edit is described-worthy → stale.
	fpFixtureWrite(t, root, "internal/alpha/alpha.go", "package alpha\n\nfunc A() { _ = 1 }\n")
	if consumer() == pv.ContentFingerprint {
		t.Errorf("a production .go edit left the codemaps fingerprint unmoved")
	}
}
