package hook

// navigator_detect_nonoverlap_test.go — SPEC-NAVIGATOR-SYNC-002 M1.5 (BAS
// Epic M1 Falconer Detect — non-overlap + consumer-only guards). TDD
// coverage for:
//   - AC-NS2-008 (non-overlap with predecessor chains) — the Detect layer
//     ONLY reads nav-graph.json and NEVER writes to it or to the other
//     three predecessor-chain outputs (capability-map.md, audit-report,
//     capability-symbols). Pattern carried forward from
//     internal/navigator/sync/nonoverlap_test.go (M0).
//   - AC-NS2-005b (consumer-only: read via public API) — navigator_detect.go
//     has NO write/rename call targeting a predecessor surface; the M0 types
//     are consumed via the public sync.Graph / sync.Edge / sync.Node API only
//     (the unexported nodeKey helper at internal/navigator/sync/schema.go:97
//     was re-declared inside internal/navigator/detect/traverse.go precisely
//     for this reason — see REQ-NS2-005 / plan.md §C.7 asset-reuse map).
//
// These guards mechanically enforce the bridge-not-absorb invariant: the
// Detect layer sits ON TOP of the M0 producer + the mx SpecAssociator and
// consumes their outputs read-only. A regression that adds a write to a
// predecessor surface would silently corrupt the producer's output and is
// the single most dangerous scope-creep failure mode for this layer.
//
// All guards here are source-content checks, safe to run on any branch. A
// prior AC-NS2-005a guard instead diffed the working branch against
// origin/main and failed whenever the diff touched internal/navigator/sync/
// or internal/mx/. That assertion was one-shot — it held for the branch that
// introduced this layer and was satisfied at merge — but as a permanent test
// it failed every later branch legitimately editing those packages. Do not
// reintroduce a git-diff-based guard here; scope such an assertion to the
// branch that needs it.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// predecessorSurfaces lists the output artifacts of the three existing
// Navigator chains + the M0 graph + the mx layer. The Detect layer MUST NOT
// write to any of these; it consumes them read-only (REQ-NS2-005, REQ-NS2-008).
var predecessorSurfaces = []string{
	"nav-graph.json",
	"capability-map.md",
	"audit-report.md",
	"audit-report.json",
	"capability-symbols.md",
	"capability-symbols.json",
}

// writeVerbs are the filesystem-mutation API surface we forbid on predecessor
// surfaces. Read references (os.ReadFile, os.Open, json.Unmarshal on a
// already-read buffer) do NOT count — they are the read-only consumption the
// bridge-not-absorb pattern permits.
var writeVerbs = []string{
	"os.WriteFile",
	"os.Rename",
	"os.Create",
	"ioutil.WriteFile",
}

// detectSourceGlob returns the non-test Go source files in the hook package
// that make up the Detect layer production surface.
func detectSourceGlob(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob("navigator_detect*.go")
	if err != nil {
		t.Fatalf("glob navigator_detect*.go: %v", err)
	}
	var out []string
	for _, m := range matches {
		// Skip test files — the assertions live here and they necessarily
		// mention the forbidden fragments as the assertion targets.
		if strings.HasSuffix(m, "_test.go") {
			continue
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		t.Fatal("no navigator_detect*.go production source files found (test running in the wrong directory?)")
	}
	return out
}

// TestNonOverlap_DetectReadsNavGraphOnly (AC-NS2-008, REQ-NS2-008) asserts
// the bridge-not-absorb invariant for the M0 graph: navigator_detect.go may
// reference nav-graph.json ONLY in a read-or-neutral context — a path
// constant declaration, a read-shaped verb (os.ReadFile / os.Open), or a
// comment. The forbidden case is a WRITE verb (os.WriteFile / os.Rename /
// os.Create / ioutil.WriteFile) on a line that also mentions nav-graph.json.
// Other predecessor surfaces (capability-map.md / audit-report /
// capability-symbols) must have ZERO references in the Detect source at all
// — the Detect layer does not even read those, only the M0 joined graph.
func TestNonOverlap_DetectReadsNavGraphOnly(t *testing.T) {
	// Serial: reads committed source files via relative globs. No parallel
	// mutation hazard; kept serial for consistency with the M0 nonoverlap
	// test it carries forward from.
	for _, src := range detectSourceGlob(t) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		body := string(b)

		// (1) nav-graph.json references: a write verb co-occurring with
		// nav-graph.json on the same non-comment line is a bridge-not-absorb
		// violation. Read references and path-constant declarations are
		// allowed (the const navGraphRelPath is the path passed to
		// os.ReadFile inside detectForChangedPath).
		for _, line := range strings.Split(body, "\n") {
			if !strings.Contains(line, "nav-graph.json") {
				continue
			}
			if isCommentLine(line) {
				continue
			}
			if hasWriteVerb(line) {
				t.Errorf("%s: nav-graph.json appears on a write-verb line — the Detect layer must read nav-graph.json only, never write it: %s",
					src, strings.TrimSpace(line))
			}
		}

		// (2) Other predecessor surfaces: ZERO references allowed. The
		// Detect layer does not even READ capability-map / audit-report /
		// capability-symbols — it consumes only the M0 joined graph.
		for _, surface := range predecessorSurfaces {
			if surface == "nav-graph.json" {
				continue
			}
			if strings.Contains(body, surface) {
				t.Errorf("%s: forbidden predecessor-surface reference %q appears in source (Detect must not touch the three predecessor chains' outputs, only nav-graph.json)",
					src, surface)
			}
		}
	}
}

// TestNonOverlap_DetectDoesNotWritePredecessorSurfaces (AC-NS2-008 +
// AC-NS2-005b, REQ-NS2-008) asserts that no line in the Detect source
// carries a write-shaped verb (os.WriteFile / os.Rename / os.Create /
// ioutil.WriteFile) targeting any predecessor surface. This is the write-
// verb proxy carried forward from
// internal/navigator/sync/nonoverlap_test.go:TestNonOverlap_SourceGrepForbiddenWriteSurfaces.
func TestNonOverlap_DetectDoesNotWritePredecessorSurfaces(t *testing.T) {
	for _, src := range detectSourceGlob(t) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		for _, line := range strings.Split(string(b), "\n") {
			if isCommentLine(line) {
				continue
			}
			if !hasWriteVerb(line) {
				continue
			}
			for _, surface := range predecessorSurfaces {
				if strings.Contains(line, surface) {
					t.Errorf("%s: write verb targets forbidden predecessor surface %q: %s",
						src, surface, strings.TrimSpace(line))
				}
			}
		}
	}
}

// TestNonOverlap_DetectNeverWritesToSyncOrMxPaths (AC-NS2-005b,
// REQ-NS2-005) asserts the Detect source never writes to a path under
// internal/navigator/sync/ or internal/mx/. Those packages are the M0
// producer + the SpecAssociator; the Detect layer MUST consume them via
// their public API only. The only writes navigator_detect.go performs are
// the JSONL append under .moai/state/navigator-detect/ and the log under
// .moai/logs/ — both M1-owned advisory surfaces, never predecessor surfaces.
func TestNonOverlap_DetectNeverWritesToSyncOrMxPaths(t *testing.T) {
	forbiddenPathFragments := []string{
		"internal/navigator/sync/",
		"internal/mx/",
	}
	for _, src := range detectSourceGlob(t) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		for _, line := range strings.Split(string(b), "\n") {
			if isCommentLine(line) {
				continue
			}
			if !hasWriteVerb(line) {
				continue
			}
			for _, frag := range forbiddenPathFragments {
				if strings.Contains(line, frag) {
					t.Errorf("%s: write verb targets M0/mx producer path %q: %s",
						src, frag, strings.TrimSpace(line))
				}
			}
		}
	}
}

// --- helpers ---

// isCommentLine reports whether the line is a Go comment (// or /* */ or
// the leading godoc). Comment references to nav-graph.json are allowed —
// they document the read-only consumption, not perform it.
func isCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*")
}

// hasWriteVerb reports whether the line mentions a write-shaped filesystem verb.
func hasWriteVerb(line string) bool {
	for _, v := range writeVerbs {
		if strings.Contains(line, v) {
			return true
		}
	}
	return false
}
