package graph

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// M2 regression locks (SPEC-MX-TAG-EDGES-001): the artifact contract the
// tag layer must never break. AC-MTE-001 (determinism incl. mx-* lines),
// AC-MTE-005 (legacy artifact without mx-* lines), AC-MTE-007 (metadata
// stays scanner-side), AC-MTE-008 (traversal additivity + reverse-only).

// AC-MTE-001 — two builds over fresh copies of the same tree write
// byte-identical edges.jsonl INCLUDING all mx-* lines (REQ-MTE-003: no
// wall-clock, no scan-order dependence).
func TestEdgesJSONLDeterministicWithTags(t *testing.T) {
	paths := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		root := tagFixture(t)
		path := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
		edges, err := Build(root)
		if err != nil {
			t.Fatalf("build %d: %v", i, err)
		}
		hasTagEdge := false
		for _, e := range mxEdgesOf(edges) {
			if e.Kind == KindMXDebt || e.Kind == KindMXAnchor {
				hasTagEdge = true
			}
		}
		if !hasTagEdge {
			t.Fatalf("build %d produced no mx-debt/mx-anchor lines — nothing was compared", i)
		}
		if err := WriteJSONL(path, edges); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
		paths = append(paths, path)
	}
	base, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	for i, path := range paths[1:] {
		other, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(base, other) {
			t.Fatalf("edges.jsonl not byte-identical: build 0 vs build %d\nA=%s\nB=%s", i+1, base, other)
		}
	}
	if !strings.Contains(string(base), `"kind":"mx-debt"`) {
		t.Fatal("artifact carries no mx-debt line — the determinism lock swept nothing")
	}
}

// AC-MTE-005 — an edges.jsonl written WITHOUT any mx-* lines (pre-upgrade
// artifact) loads and serves queries without error: absent kinds are absent
// edges, never a failure (REQ-MTE-005). Named so the AC's
// `-run TestLegacyArtifactLoad` selector sweeps it.
func TestLegacyArtifactLoadMxAbsent(t *testing.T) {
	root := tagFixture(t)
	dir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := `{"kind":"import","source":"internal/app","target":"internal/helper"}
{"kind":"mx-spec","source":"internal/app/debts.go","target":"SPEC-OLD-001","line":4}
{"kind":"spec-depends","source":"SPEC-OLD-001","target":"SPEC-OLD-DEP-001"}
`
	path := filepath.Join(dir, "edges.jsonl")
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadJSONL(path)
	if err != nil {
		t.Fatalf("mx-less artifact must load: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("loaded %d edges, want 3", len(loaded))
	}
	for _, e := range mxEdgesOf(loaded) {
		t.Errorf("mx-kind edge in a pre-upgrade artifact: %+v", e)
	}
	if got := FindCallers(loaded, "SPEC-OLD-001"); len(got) != 1 {
		t.Errorf("FindCallers over mx-less artifact = %v, want the single mx-spec source", got)
	}
}

// AC-MTE-007 — the DEBT pair (with vs without @MX:UPGRADE) emits edges with
// IDENTICAL key sets, and neither RotRisk state nor scanner wall-clock
// reaches any mx-* line (REQ-MTE-003, REQ-MTE-007).
func TestTagEdgesCarryNoMetadata(t *testing.T) {
	root := tagFixture(t)
	edges, err := Build(root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	debts := map[string]Edge{}
	for _, e := range mxEdgesOf(edges) {
		if e.Kind == KindMXDebt {
			debts[e.Source] = e
		}
	}
	triggered, okT := debts["internal/app/debts.go"]
	rotting, okR := debts["internal/app/rot.go"]
	if !okT || !okR {
		t.Fatalf("DEBT pair missing: %+v", debts)
	}

	// Identical KEY SETS: serialize both and compare the key universe.
	keysOf := func(t *testing.T, e Edge) []string {
		t.Helper()
		line, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(line, &m); err != nil {
			t.Fatal(err)
		}
		out := make([]string, 0, len(m))
		for k := range m {
			out = append(out, k)
		}
		return out
	}
	tKeys, rKeys := keysOf(t, triggered), keysOf(t, rotting)
	if len(tKeys) != len(rKeys) {
		t.Errorf("DEBT pair key sets differ: %v vs %v", tKeys, rKeys)
	}
	for _, k := range tKeys {
		found := false
		for _, rk := range rKeys {
			if rk == k {
				found = true
			}
		}
		if !found {
			t.Errorf("triggered DEBT key %q absent from the rot-risk DEBT edge", k)
		}
	}

	// No scanner-mutable state anywhere on an mx-* line: the raw lines are
	// scanned for the rot flag, provenance, and wall-clock keys.
	for _, e := range mxEdgesOf(edges) {
		line, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		for _, banned := range []string{"rotRisk", "rot_risk", "lastSeenAt", "createdBy", "contentHash", "body", "reason"} {
			if strings.Contains(string(line), banned) {
				t.Errorf("mx-* line carries scanner metadata key %q: %s", banned, line)
			}
		}
	}
}

// AC-MTE-008 — mx-* edges propagate REVERSE-ONLY in FindCallers and
// BlastRadius (never bidirectional like mx-spec), and a graph with the mx-*
// edges zeroed returns the pre-change traversal results (REQ-MTE-008).
func TestTraversalAdditivityWithTags(t *testing.T) {
	requireCodeExtraction(t) // the mx-debt symbol target needs the range join
	root := tagFixture(t)
	full, _, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers: %v", err)
	}
	var zeroed []Edge
	for _, e := range full {
		if strings.HasPrefix(e.Kind, "mx-") && e.Kind != KindMXSpec {
			continue
		}
		zeroed = append(zeroed, e)
	}

	const debtTarget = "Debted"
	const debtFile = "internal/app/debts.go"

	// (a) zeroed set = pre-change behavior: FindCallers of the symbol is
	// empty, and the file's blast radius is EXACTLY the pre-change result
	// (the comparison below asserts the full set adds nothing to it).
	if got := FindCallers(zeroed, debtTarget); len(got) != 0 {
		t.Errorf("pre-change FindCallers(%q) = %v, want empty", debtTarget, got)
	}
	preBlast := BlastRadius(zeroed, debtFile) // mx-spec bidirectional reach, unchanged by the tag layer

	// (b) full set: reverse propagation only. The tag's file gains the
	// debt target's caller set...
	callers := FindCallers(full, debtTarget)
	found := false
	for _, c := range callers {
		if c == debtFile {
			found = true
		}
	}
	if !found {
		t.Errorf("FindCallers(%q) = %v, want the tag's file %q", debtTarget, callers, debtFile)
	}
	if got := BlastRadius(full, debtTarget); len(got) == 0 || !contains(got, debtFile) {
		t.Errorf("BlastRadius(%q) = %v, want %q reachable (reverse-only propagation)", debtTarget, got, debtFile)
	}
	// ...and the tag's file's blast radius does NOT gain the target's
	// callers (mx-* is NOT bidirectional — no forward edge file→symbol); it
	// stays byte-identical to the zeroed set's result.
	fullBlast := BlastRadius(full, debtFile)
	if contains(fullBlast, debtTarget) {
		t.Errorf("BlastRadius(%q) = %v — mx-debt propagated forward, want reverse-only", debtFile, fullBlast)
	}
	if strings.Join(fullBlast, "\x00") != strings.Join(preBlast, "\x00") {
		t.Errorf("BlastRadius(%q) changed with the tag layer: pre-change %v vs full %v", debtFile, preBlast, fullBlast)
	}
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
