//go:build !cgo

package graph

import (
	"os"
	"path/filepath"
	"testing"
)

// B5 / nocgo deliberate semantics (SPEC-GRAPH-REPORT-001 M3): under a
// CGO-disabled build the code layer is absent — the extraction walk scans
// nothing, so an existing artifact built by a previous cgo run (or on a cgo
// leg) shrinks to its doc edges on the next rebuild. The discriminator
// decides, not the count: the removed edges' source files still exist on
// disk and lie outside the (empty) scanned set, so the guard REFUSES — the
// write path answers from the existing artifact instead of silently
// publishing a codeless graph as if it were the whole truth. This refusal is
// the DELIBERATE semantics, pinned here so a nocgo leg cannot mistake it
// for a regression.
func TestDetectUnexplainedShrink_NoCGOEmptyScannedSetRefuses(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "internal", "demo")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "calls.go"),
		[]byte("package demo\n\nfunc Calls() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	existing := []Edge{{Kind: KindCodeCall, Source: "internal/demo/calls.go:Calls", Target: "Helper"}}
	var rebuilt []Edge // the nocgo rebuild carries doc edges only
	var scanned []string

	if report := DetectUnexplainedShrink(existing, rebuilt, setOf(scanned), root); report.Empty() {
		t.Fatal("nocgo empty-scanned rebuild over an existing cgo artifact must refuse — the discriminator (file exists AND unscanned) holds")
	}
}

// The doc-edge half of a nocgo rebuild passes untouched: non-file-sourced
// kinds are out of the guard's scope, so a doc-only artifact overwrites a
// doc-only artifact without refusal.
func TestDetectUnexplainedShrink_NoCGODocEdgesUnaffected(t *testing.T) {
	root := t.TempDir()
	existing := []Edge{
		{Kind: KindImport, Source: "internal/demo", Target: "internal/config"},
		{Kind: KindSpecDepends, Source: "SPEC-A-001", Target: "SPEC-B-001"},
	}
	if report := DetectUnexplainedShrink(existing, nil, map[string]bool{}, root); !report.Empty() {
		t.Fatalf("doc-only shrink must not refuse, got:\n%s", report.Describe())
	}
}

func setOf(files []string) map[string]bool {
	out := map[string]bool{}
	for _, f := range files {
		out[f] = true
	}
	return out
}
