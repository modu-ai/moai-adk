package graph

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// M3 acceptance tests (SPEC-MX-TAG-EDGES-001): the graph-backed fan-in
// evidence source. AC-MTE-009 (parity fixture + direction) and AC-MTE-012
// (hub exclusion). The textual reference is an IN-TEST counter replicating
// hook/mx fanInIndex semantics over the fixture bytes — test-only, never a
// production seam (plan.md §C: this is the one package where both the graph
// source and the textual semantics are observable without breaking the
// layering lock).

// parityFixture builds the AC-MTE-009 fixture: symbol S declared in
// internal/core/core.go, called from 3 distinct evidence-backed EXTERNAL
// caller files, plus 2 comment/string mentions, 2 same-file call sites,
// 1 inferred-only caller file, and 3 test-file callers.
func parityFixture(t *testing.T) string {
	t.Helper()
	return writeParityHubFixture(t, map[string]string{
		"internal/core/core.go": `package core

// S is referenced in prose twice here: S and S again.
const note = "mention S inside a string"

func S() {}

func Local() {
	S()
	S()
}
`,
		"internal/a/a.go": `package a

import "example.com/parity/internal/core"

func A() {
	core.S()
}
`,
		"internal/b/b.go": `package b

import "example.com/parity/internal/core"

func B() {
	core.S()
}
`,
		"internal/core/core2.go": `package core

func C2() {
	S()
}
`,
		"internal/x/x.go": `package x

func X() {
	S()
}
`,
		"internal/core/core_test.go": `package core

import "testing"

func TestS(t *testing.T) {
	S()
}
`,
		"internal/tests/api_test.go": `package tests

import (
	"testing"

	"example.com/parity/internal/core"
)

func TestSApi(t *testing.T) {
	core.S()
}
`,
		"internal/fixtures/fix_test.go": `package fixtures

import "example.com/parity/internal/core"

func TestSFix(t *testing.T) {
	core.S()
}
`,
	})
}

// hubFixture builds the AC-MTE-012 fixture: HS has 3 evidence-backed source
// callers AND 3 test-file callers; HOnly has ONLY test-file callers.
func hubFixture(t *testing.T) string {
	t.Helper()
	return writeParityHubFixture(t, map[string]string{
		"internal/core/hub.go": `package core

func HS() {}

func HOnly() {}
`,
		"internal/a/a.go": `package a

import "example.com/parity/internal/core"

func A() {
	core.HS()
}
`,
		"internal/b/b.go": `package b

import "example.com/parity/internal/core"

func B() {
	core.HS()
}
`,
		"internal/c/c.go": `package c

import "example.com/parity/internal/core"

func C() {
	core.HS()
}
`,
		"internal/core/hub_test.go": `package core

import "testing"

func TestHS(t *testing.T) {
	HS()
}

func TestHOnly(t *testing.T) {
	HOnly()
}
`,
		"internal/tests/api_test.go": `package tests

import (
	"testing"

	"example.com/parity/internal/core"
)

func TestHSApi(t *testing.T) {
	core.HS()
}

func TestHOnlyApi(t *testing.T) {
	core.HOnly()
}
`,
		"internal/fixtures/fix_test.go": `package fixtures

import "example.com/parity/internal/core"

func TestHSFix(t *testing.T) {
	core.HS()
}

func TestHOnlyFix(t *testing.T) {
	core.HOnly()
}
`,
	})
}

// writeParityHubFixture materializes a Go module from rel→src pairs and
// builds + writes the edges artifact the edge-backed source reads.
func writeParityHubFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, src string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/parity\n\ngo 1.26\n")
	for rel, src := range files {
		write(rel, src)
	}

	edges, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers: %v", err)
	}
	dir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSONL(filepath.Join(dir, "edges.jsonl"), edges); err != nil {
		t.Fatalf("WriteJSONL: %v", err)
	}
	return root
}

// textualReference is the in-test textual-matched reference counter: it
// replicates hook/mx fanInIndex semantics (word-boundary identifier
// occurrences across all .go files, subtracting the declaration occurrence
// in currentFile) over the fixture tree. TEST-ONLY.
func textualReference(t *testing.T, root, funcName, currentFileRel string) int {
	t.Helper()
	counts := map[string]int{}
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == "vendor" || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		counts[path] = countWordBoundary(string(data), funcName)
		return nil
	})
	if walkErr != nil {
		t.Fatal(walkErr)
	}
	total := 0
	for path, n := range counts {
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			t.Fatal(relErr)
		}
		if rel == currentFileRel {
			total += n - 1 // subtract the declaration itself
			continue
		}
		total += n
	}
	if total < 0 {
		total = 0
	}
	return total
}

// countWordBoundary counts ASCII word-boundary occurrences of name — the
// same token class fanInIndex counts (isIdentByte).
func countWordBoundary(s, name string) int {
	isIdent := func(c byte) bool {
		return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
	}
	count := 0
	for i := 0; i < len(s); {
		if !isIdent(s[i]) {
			i++
			continue
		}
		j := i
		for j < len(s) && isIdent(s[j]) {
			j++
		}
		if s[i:j] == name {
			count++
		}
		i = j
	}
	return count
}

// edgeSourceFor builds an edge-backed source over the fixture artifact.
func edgeSourceFor(t *testing.T, root string) *EdgeFanInSource {
	t.Helper()
	return NewEdgeFanInSource(root)
}

// fanInResult is one EvidenceBacked answer, unpacked for assertions.
type fanInResult struct {
	evidence int
	inferred int
	label    string
}

// evidenceFor answers EvidenceBacked and fails the test on error.
func evidenceFor(t *testing.T, src *EdgeFanInSource, funcName, currentFile string) fanInResult {
	t.Helper()
	ev, inf, label, err := src.EvidenceBacked(context.Background(), funcName, currentFile)
	if err != nil {
		t.Fatalf("EvidenceBacked(%s): %v", funcName, err)
	}
	return fanInResult{evidence: ev, inferred: inf, label: label}
}

// AC-MTE-009 — the parity fixture: graph source blocking count = 3
// (same-file calls and test callers excluded, inferred itemized separately),
// textual reference returns its larger occurrence count, delta direction
// graph <= textual asserted as the DOCUMENTED outcome.
func TestGraphFanInParityFixture(t *testing.T) {
	requireCodeExtraction(t)
	root := parityFixture(t)
	src := edgeSourceFor(t, root)

	res := evidenceFor(t, src, "S", filepath.Join(root, "internal", "core", "core.go"))
	if res.label == "" {
		t.Error("EvidenceBacked returned an empty source label — the verdict would be unlabeled")
	}
	if res.evidence != 3 {
		t.Errorf("graph blocking count = %d, want 3 (a.go + b.go extracted, core2.go intra-package; same-file and test callers excluded)", res.evidence)
	}
	if res.inferred != 1 {
		t.Errorf("graph inferred-only count = %d, want 1 (x.go)", res.inferred)
	}

	// Delta direction is the documented outcome, not a failure.
	textual := textualReference(t, root, "S", "internal/core/core.go")
	if textual <= res.evidence {
		t.Errorf("documented delta direction violated: textual=%d must exceed graph evidence=%d", textual, res.evidence)
	}
}

// AC-MTE-012 — hub exclusion: test-file callers never reach the blocking
// count (REQ-SPC-004-040 hard-coded fallback patterns; mx.yaml test_paths
// NOT honored — accepted divergence).
func TestHubExclusionTestCallers(t *testing.T) {
	requireCodeExtraction(t)
	root := hubFixture(t)
	src := edgeSourceFor(t, root)

	hs := evidenceFor(t, src, "HS", filepath.Join(root, "internal", "core", "hub.go"))
	if hs.label == "" {
		t.Error("EvidenceBacked(HS) returned an empty source label")
	}
	if hs.evidence != 3 {
		t.Errorf("HS blocking count = %d, want 3 source callers (3 test-file callers excluded)", hs.evidence)
	}
	if hs.inferred != 0 {
		t.Errorf("HS inferred-only count = %d, want 0", hs.inferred)
	}

	// Only-test-file callers → blocking count 0.
	hOnly := evidenceFor(t, src, "HOnly", filepath.Join(root, "internal", "core", "hub.go"))
	if hOnly.label == "" {
		t.Error("EvidenceBacked(HOnly) returned an empty source label")
	}
	if hOnly.evidence != 0 {
		t.Errorf("HOnly blocking count = %d, want 0 (test-file callers only)", hOnly.evidence)
	}
	if hOnly.inferred != 0 {
		t.Errorf("HOnly inferred-only count = %d, want 0", hOnly.inferred)
	}
}
