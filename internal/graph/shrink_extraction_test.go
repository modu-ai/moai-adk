//go:build cgo

package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// D23 fixture discipline: the partial-failure fixture uses REAL extraction
// output — genuine `file:function` Source shapes from BuildWithCodeLayers —
// never hand-written bare-path code-call Sources.
func shrinkExtractionFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	src := filepath.Join(root, "internal", "demo")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "calls.go"),
		[]byte("package demo\n\nfunc Calls() string { return Helper() }\n\nfunc Helper() string { return \"x\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// REQ-GR-008 end-to-end on real extraction shapes: a rebuild that loses one
// scanned file's edges (the file still on disk, dropped from the scanned set
// — the partial-failure shape) refuses, naming the edge with its real
// compound Source and the unscanned file.
func TestDetectUnexplainedShrink_RealExtractionPartialFailure(t *testing.T) {
	root := shrinkExtractionFixture(t)
	existing, scanned, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	var callEdge Edge
	found := false
	for _, e := range existing {
		if e.Kind == KindCodeCall && strings.Contains(e.Source, "calls.go:") {
			callEdge, found = e, true
			break
		}
	}
	if !found {
		t.Fatalf("fixture must yield a real code-call edge with a compound source, got: %+v", existing)
	}
	if !containsString(scanned, "internal/demo/calls.go") {
		t.Fatalf("fixture precondition: calls.go must be in the scanned set, got: %v", scanned)
	}

	// Partial failure: the rebuild drops calls.go entirely — its edges gone,
	// the file excluded from the scanned set while still on disk.
	var rebuilt []Edge
	for _, e := range existing {
		if e.Kind == KindCodeCall || e.Kind == KindCodeImport {
			if file, _ := splitCodeNode(e.Source); file == "internal/demo/calls.go" || e.Source == "internal/demo/calls.go" {
				continue
			}
		}
		rebuilt = append(rebuilt, e)
	}
	scannedMinus := make([]string, 0, len(scanned))
	for _, f := range scanned {
		if f != "internal/demo/calls.go" {
			scannedMinus = append(scannedMinus, f)
		}
	}
	set := map[string]bool{}
	for _, f := range scannedMinus {
		set[f] = true
	}

	report := DetectUnexplainedShrink(existing, rebuilt, set, root)
	if report.Empty() {
		t.Fatal("a real-shape partial failure must refuse the overwrite")
	}
	msg := report.Describe()
	if !strings.Contains(msg, callEdge.Source) {
		t.Errorf("refusal must name the removed edge's real compound source %q, got:\n%s", callEdge.Source, msg)
	}
	if !strings.Contains(msg, "internal/demo/calls.go") {
		t.Errorf("refusal must name the unscanned source file, got:\n%s", msg)
	}

	// Control: the FULL real rebuild against itself never refuses.
	fullSet := map[string]bool{}
	for _, f := range scanned {
		fullSet[f] = true
	}
	if report := DetectUnexplainedShrink(existing, existing, fullSet, root); !report.Empty() {
		t.Errorf("a faithful rebuild must never refuse, got:\n%s", report.Describe())
	}
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
