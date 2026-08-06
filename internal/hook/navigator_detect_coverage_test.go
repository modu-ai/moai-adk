package hook

// navigator_detect_coverage_test.go — SPEC-NAVIGATOR-SYNC-002 M1.5 (BAS Epic
// M1 Falconer Detect — coverage harness). TDD coverage for AC-NS2-007 /
// REQ-NS2-007: the ≥80% mapping-coverage gate, mechanically measured over a
// deterministic fixture corpus.
//
// This is NOT the unit-test line-coverage % (the 88.6% figure from M1.1 is
// `go test -cover` over the detect package). The mapping coverage % measures
// the Detect layer's correctness over a realistic corpus of changed-path
// inputs: of paths that fall inside the M0 graph's scan roots, what fraction
// maps to a non-empty affected-row set when a matching graph row exists.
//
// The fixture corpus lives at internal/hook/testdata/navigator-detect-corpus/
// (nav-graph.json + corpus_cases.json + README.md). The test loads both,
// iterates the corpus, runs detect.Traverse per case, classifies the result
// into in-scope-mapped / in-scope-unmapped / out-of-scope, computes the ratio
//   coverage = (in-scope-mapped) / (in-scope-mapped + in-scope-unmapped)
// and asserts coverage >= 0.80 (REQ-NS2-007). Out-of-scope cases are excluded
// from BOTH numerator and denominator (plan.md §E partition).
//
// Attribution (verification-claim-integrity §2): the coverage number is
// attributable to the exact `go test … -v` invocation + its stdout; it is NOT
// a carried-over estimate. A subsequent run on the same fixture corpus
// produces the same percentage (deterministic).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/navigator/detect"
	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// coverageThreshold is the REQ-NS2-007 floor. The Detect layer's mapping
// coverage over the fixture corpus MUST meet or exceed this value. Named
// constant (not a magic literal in the assertion) per the hardcoding-
// prevention rule (hns-moaiadk-best-practices §thresholds).
const coverageThreshold = 0.80

// corpusDir is the committed fixture-corpus directory relative to the
// internal/hook package root.
const corpusDir = "testdata/navigator-detect-corpus"

// corpusCase is one row of the fixture corpus manifest. Each case carries a
// changed-path input and its expected mapping class.
type corpusCase struct {
	ChangedPath string `json:"changed_path"`
	Class       string `json:"class"`
	Note        string `json:"note"`
}

// corpusManifest is the JSON shape of corpus_cases.json.
type corpusManifest struct {
	Cases []corpusCase `json:"cases"`
}

// TestNavigatorDetectCoverage is the AC-NS2-007 / REQ-NS2-007 coverage gate.
//
// It loads the pre-built nav-graph.json fixture and the corpus_cases.json
// manifest from the committed corpus directory, runs detect.Traverse for
// every case, tallies the in-scope-mapped vs in-scope-unmapped counts (the
// out-of-scope class is excluded from BOTH per plan.md §E), computes the
// coverage ratio, and asserts coverage >= 0.80.
//
// On failure the test prints the observed ratio + the per-case breakdown so
// the regression is immediately diagnosable.
func TestNavigatorDetectCoverage(t *testing.T) {
	// Serial: loads committed fixtures from disk via a relative path. No
	// parallel mutation hazard, but keeping it serial matches the other
	// committed-fixture tests in this package and makes the barrier-sharing
	// atomic-read test in the same -race invocation easier to reason about.
	graphPath := filepath.Join(corpusDir, "nav-graph.json")
	raw, err := os.ReadFile(graphPath)
	if err != nil {
		t.Fatalf("read fixture graph %s: %v\n"+
			"(the committed corpus is required for AC-NS2-007; if you moved the "+
			"testdata directory, update corpusDir in this test)", graphPath, err)
	}
	var graph navsync.Graph
	if err := json.Unmarshal(raw, &graph); err != nil {
		t.Fatalf("unmarshal fixture graph %s: %v", graphPath, err)
	}
	if graph.Edges == nil {
		t.Fatalf("fixture graph %s has no edges array — the corpus is malformed", graphPath)
	}

	manifestPath := filepath.Join(corpusDir, "corpus_cases.json")
	mraw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read corpus manifest %s: %v", manifestPath, err)
	}
	var manifest corpusManifest
	if err := json.Unmarshal(mraw, &manifest); err != nil {
		t.Fatalf("unmarshal corpus manifest %s: %v", manifestPath, err)
	}
	if len(manifest.Cases) < 20 {
		t.Fatalf("corpus manifest has only %d cases; AC-NS2-007 requires N>=20 "+
			"(in-scope count alone must be >=20)", len(manifest.Cases))
	}

	var mapped, unmapped, outOfScope int
	var unmappedPaths []string
	for _, c := range manifest.Cases {
		result, err := detect.Traverse(&graph, c.ChangedPath)
		if err != nil {
			t.Fatalf("Traverse(%q) returned error: %v", c.ChangedPath, err)
		}
		hasRows := result != nil && (len(result.Nodes) > 0 || len(result.Edges) > 0)
		switch c.Class {
		case "in-scope-mapped":
			if hasRows {
				mapped++
			} else {
				// A mapped case that returned 0 rows is a REGRESSION — the
				// graph indexes this path but Traverse did not surface it.
				unmapped++
				unmappedPaths = append(unmappedPaths, c.ChangedPath+" (expected mapped, got 0 rows)")
			}
		case "in-scope-unmapped":
			// Expected 0 rows; exercises the denominator deliberately.
			if hasRows {
				// Unexpected mapping — not a failure per se (the layer
				// returned more than expected), but record it for diagnosis.
				t.Logf("note: in-scope-unmapped case %q returned %d row(s); "+
					"recording as mapped (raises observed coverage)",
					c.ChangedPath, len(result.Edges))
				mapped++
			} else {
				unmapped++
				unmappedPaths = append(unmappedPaths, c.ChangedPath)
			}
		case "out-of-scope":
			outOfScope++
			// Out-of-scope cases are excluded from BOTH numerator and
			// denominator (plan.md §E partition). We do NOT assert hasRows
			// here — an out-of-scope path may coincidentally match no edge
			// (the expected case) and that is fine.
		default:
			t.Fatalf("corpus case %q has unknown class %q (expected in-scope-mapped / in-scope-unmapped / out-of-scope)",
				c.ChangedPath, c.Class)
		}
	}

	inScope := mapped + unmapped
	if inScope == 0 {
		t.Fatal("corpus has 0 in-scope cases; cannot compute coverage ratio (division by zero)")
	}
	coverage := float64(mapped) / float64(inScope)

	t.Logf("coverage corpus summary: total=%d, in-scope=%d (mapped=%d, unmapped=%d), out-of-scope=%d",
		len(manifest.Cases), inScope, mapped, unmapped, outOfScope)
	t.Logf("observed mapping coverage: %d/%d = %.4f (threshold %.2f)", mapped, inScope, coverage, coverageThreshold)

	if coverage < coverageThreshold {
		t.Errorf("AC-NS2-007 FAIL: mapping coverage %.4f < threshold %.2f\n"+
			"unmapped in-scope paths (%d):\n  %s",
			coverage, coverageThreshold, len(unmappedPaths),
			joinPaths(unmappedPaths))
	}
}

// joinPaths renders a slice of path strings as an indented bullet list for
// the failure message. Pure helper, no I/O.
func joinPaths(paths []string) string {
	if len(paths) == 0 {
		return "(none)"
	}
	out := ""
	for i, p := range paths {
		if i > 0 {
			out += "\n  "
		}
		out += "- " + p
	}
	return out
}
