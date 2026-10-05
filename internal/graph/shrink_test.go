package graph

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// shrinkFixtureFile writes one on-disk file under root (rel is slash-separated
// and project-relative, the extraction's own path domain) so the guard's
// disk-existence test sees it.
func shrinkFixtureFile(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// REQ-GR-008 (AC-GR-012 core): a removed file-sourced edge whose file still
// exists on disk AND lies outside the scanned set is an unexplained shrink —
// the report names the edge and the unscanned source file.
func TestDetectUnexplainedShrink_RefusesExistingUnscanned(t *testing.T) {
	root := t.TempDir()
	shrinkFixtureFile(t, root, "internal/demo/calls.go", "package demo\n")

	existing := []Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}}
	rebuilt := []Edge{} // partial failure: the call edge vanished
	scanned := map[string]bool{}

	report := DetectUnexplainedShrink(existing, rebuilt, scanned, root)
	if report.Empty() {
		t.Fatal("existing-but-unscanned source must produce a non-empty report")
	}
	msg := report.Describe()
	for _, want := range []string{
		"internal/demo/calls.go:Calls", // the removed edge's Source
		"Helper",                       // the removed edge's Target
		"internal/demo/calls.go",       // the unscanned source FILE
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("report must name %q, got:\n%s", want, msg)
		}
	}
}

// REQ-GR-008 existence discriminator: a source file ABSENT from disk is a
// genuine deletion — its edges may legitimately vanish, never a defect.
func TestDetectUnexplainedShrink_GenuineDeletionIsEmpty(t *testing.T) {
	root := t.TempDir() // internal/demo/calls.go never created

	existing := []Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}}
	report := DetectUnexplainedShrink(existing, nil, map[string]bool{}, root)
	if !report.Empty() {
		t.Fatalf("deleted source must not be reported, got:\n%s", report.Describe())
	}
}

// A removed edge whose file was INSIDE the scanned set is explained by the
// scan itself (the extractor saw the file; the relationship changed) — no
// refusal.
func TestDetectUnexplainedShrink_ScannedSourceIsExplained(t *testing.T) {
	root := t.TempDir()
	shrinkFixtureFile(t, root, "internal/demo/calls.go", "package demo\n")

	existing := []Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}}
	scanned := map[string]bool{"internal/demo/calls.go": true}
	report := DetectUnexplainedShrink(existing, nil, scanned, root)
	if !report.Empty() {
		t.Fatalf("scanned source must not be reported, got:\n%s", report.Describe())
	}
}

// REQ-GR-008 set-difference trigger (D25): the trigger is `existing − rebuilt`
// regardless of totals — an equal-cardinality remove+add mutant (one unscanned
// edge lost, one unrelated edge gained, same total count) still refuses.
func TestDetectUnexplainedShrink_EqualCardinalityMutant(t *testing.T) {
	root := t.TempDir()
	shrinkFixtureFile(t, root, "internal/demo/calls.go", "package demo\n")

	existing := []Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}}
	rebuilt := []Edge{{Kind: KindCodeCall, Source: "internal/demo/other.go:Other", Target: "Helper"}}
	if len(existing) != len(rebuilt) {
		t.Fatalf("fixture must keep equal cardinality: %d vs %d", len(existing), len(rebuilt))
	}
	report := DetectUnexplainedShrink(existing, rebuilt, map[string]bool{}, root)
	if report.Empty() {
		t.Fatal("equal-cardinality substitution must not evade the set-difference trigger")
	}
}

// REQ-GR-008 kind scope: non-file-sourced kinds — doc-import (Source is a
// package/directory name) and spec-depends (Source is a SPEC ID) — are never
// evaluated against a file test, even when a DIRECTORY of that name exists on
// disk (a stat would succeed; the guard must never perform it).
func TestDetectUnexplainedShrink_KindScopeSkipsNonFileKinds(t *testing.T) {
	root := t.TempDir()
	// internal/demo exists as a REAL directory: a mis-scoped guard would stat
	// it, find it existing + unscanned, and refuse.
	shrinkFixtureFile(t, root, "internal/demo/demo.go", "package demo\n")
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-GRAPH-SHRINK-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	existing := []Edge{
		{Kind: KindImport, Source: "internal/demo", Target: "internal/config"},
		{Kind: KindSpecDepends, Source: "SPEC-GRAPH-SHRINK-001", Target: "SPEC-OTHER-001"},
	}
	report := DetectUnexplainedShrink(existing, nil, map[string]bool{}, root)
	if !report.Empty() {
		t.Fatalf("non-file-sourced kinds must be skipped by the guard, got:\n%s", report.Describe())
	}
}

// D23 decode step: the existence test runs on the FILE part of the compound
// code-call Source (splitCodeNode), never the undecoded `file:function`
// string — proven by refusal when the FILE exists, and by no report when a
// file named like the undecoded string exists but the decoded one does not.
// The bare-path code-import Source is the real shape for its kind and is in
// scope directly.
func TestDetectUnexplainedShrink_DecodeStep(t *testing.T) {
	root := t.TempDir()
	// A file literally named "calls.go:Calls" (colon in the filename) — a
	// guard stat'ing the UNDECODED string would find it; the decoded file
	// internal/demo/calls.go does not exist, so the correct verdict is empty.
	shrinkFixtureFile(t, root, "internal/demo/calls.go:Calls", "package demo\n")

	report := DetectUnexplainedShrink(
		[]Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}},
		nil, map[string]bool{}, root)
	if !report.Empty() {
		t.Fatalf("guard must stat the DECODED file, not the compound source, got:\n%s", report.Describe())
	}

	// Control: the decoded file exists → refusal (the decode found it).
	shrinkFixtureFile(t, root, "internal/demo/calls.go", "package demo\n")
	report = DetectUnexplainedShrink(
		[]Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}},
		nil, map[string]bool{}, root)
	if report.Empty() {
		t.Fatal("existing decoded file outside the scanned set must refuse")
	}
}

// Code-import edges are the other file-sourced kind: a bare project-relative
// file path as Source, refused under the same discriminator.
func TestDetectUnexplainedShrink_CodeImportInScope(t *testing.T) {
	root := t.TempDir()
	shrinkFixtureFile(t, root, "internal/demo/imports.go", "package demo\n")

	report := DetectUnexplainedShrink(
		[]Edge{{Kind: KindCodeImport, Source: "internal/demo/imports.go", Target: "internal/config"}},
		nil, map[string]bool{}, root)
	if report.Empty() {
		t.Fatal("removed code-import edge with existing unscanned source must refuse")
	}
}

// REQ-GR-008 path validation: a Source that is not a safe project-relative
// path (absolute, `..` traversal, or a symlink escaping the project root) is
// never stat'ed — the guard skips it rather than probing outside the tree.
func TestDetectUnexplainedShrink_PathValidationSkips(t *testing.T) {
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "outside.go")
	if err := os.WriteFile(outsideFile, []byte("package outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	shrinkFixtureFile(t, root, "internal/demo/link.go", "package demo\n")
	// Symlink INSIDE the root pointing OUTSIDE it.
	if err := os.Symlink(outsideFile, filepath.Join(root, "internal", "demo", "escape.go")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	existing := []Edge{
		{Kind: KindCodeImport, Source: "/etc/hosts", Target: "x"},                            // absolute
		{Kind: KindCodeImport, Source: "../outside.go", Target: "x"},                         // .. traversal
		{Kind: KindCodeImport, Source: "internal/demo/escape.go", Target: "internal/config"}, // symlink escape
	}
	report := DetectUnexplainedShrink(existing, nil, map[string]bool{}, root)
	if !report.Empty() {
		t.Fatalf("unsafe source paths must be skipped, never stat'ed, got:\n%s", report.Describe())
	}
}

// No shrink (rebuilt a superset or equal) → empty report.
func TestDetectUnexplainedShrink_NoShrink(t *testing.T) {
	root := t.TempDir()
	shrinkFixtureFile(t, root, "internal/demo/calls.go", "package demo\n")
	base := []Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}}
	added := []Edge{base[0], {Kind: KindCodeCall, Source: "internal/demo/calls.go:Other", Target: "Helper"}}
	scanned := map[string]bool{"internal/demo/calls.go": true}

	if report := DetectUnexplainedShrink(base, added, scanned, root); !report.Empty() {
		t.Errorf("superset rebuild must not report, got:\n%s", report.Describe())
	}
	if report := DetectUnexplainedShrink(base, base, scanned, root); !report.Empty() {
		t.Errorf("identical rebuild must not report, got:\n%s", report.Describe())
	}
}

// The typed refusal carries the report: errors.As recovers it at the write
// paths, and its message is the report's own naming.
func TestShrinkRefusalError_CarriesReport(t *testing.T) {
	root := t.TempDir()
	shrinkFixtureFile(t, root, "internal/demo/calls.go", "package demo\n")
	report := DetectUnexplainedShrink(
		[]Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}},
		nil, map[string]bool{}, root)
	if report.Empty() {
		t.Fatal("fixture must produce a non-empty report")
	}

	err := error(&ShrinkRefusalError{Report: report})
	var refuse *ShrinkRefusalError
	if !errors.As(err, &refuse) {
		t.Fatal("errors.As must recover *ShrinkRefusalError")
	}
	if refuse.Report.Describe() != report.Describe() {
		t.Error("refusal must carry the report verbatim")
	}
	if !strings.Contains(err.Error(), "internal/demo/calls.go") {
		t.Errorf("refusal message must name the unscanned source, got: %v", err)
	}
}
