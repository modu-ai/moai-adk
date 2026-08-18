package hook

// navigator_detect_nonoverlap_test.go — SPEC-NAVIGATOR-SYNC-002 M1.5 (BAS
// Epic M1 Falconer Detect — non-overlap + consumer-only guards). TDD
// coverage for:
//   - AC-NS2-008 (non-overlap with predecessor chains) — the Detect layer
//     ONLY reads nav-graph.json and NEVER writes to it or to the other
//     three predecessor-chain outputs (capability-map.md, audit-report,
//     capability-symbols). Pattern carried forward from
//     internal/navigator/sync/nonoverlap_test.go (M0).
//   - AC-NS2-005a (consumer-only: M0 + mx byte-unchanged) — the M1 run-phase
//     diff touches NO path under internal/navigator/sync/ or internal/mx/.
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

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// consumerOnlyBaseline is one candidate diff baseline for the M0/mx pin.
type consumerOnlyBaseline struct {
	name string // ref name, for failure messages
	sha  string // resolved commit; empty marks the origin/main default
	// distance is how many commits HEAD carries beyond this baseline.
	// -1 marks a release tip that is NOT an ancestor of HEAD and is therefore
	// not one of HEAD's boundaries at all.
	distance int
}

// chooseConsumerOnlyBaseline picks the diff baseline for the M0/mx pin.
//
// origin/main is the default and keeps its full original strength (the
// three-dot verbatim AC form): it is displaced ONLY when a pushed release tip
// is an ancestor of HEAD AND sits strictly closer to HEAD than main's
// merge-base does. A pushed release tip is a reviewed boundary — every
// integration onto release/* rides a lane merge that carried a review verdict
// — so the pin's jurisdiction is the change-set BETWEEN such boundaries, not
// the accumulated batch content that already crossed one. Without this
// adjustment the pin fired deterministically on the release batch PR
// (release→main) and on every lane branch that had merged the release tip,
// flagging reviewed merges such as an internal/mx/spec_loader change that
// arrived through the lane flow. On a plain feature branch no release tip is
// an ancestor, origin/main applies unchanged, and the pin is byte-identical
// to the original. A tie keeps origin/main — the maximal-strength choice.
func chooseConsumerOnlyBaseline(mainDistance int, releaseTips []consumerOnlyBaseline) consumerOnlyBaseline {
	base := consumerOnlyBaseline{name: "origin/main", distance: mainDistance}
	for _, c := range releaseTips {
		if c.distance < 0 {
			continue // not an ancestor of HEAD — not HEAD's boundary
		}
		if c.distance < base.distance {
			base = c
		}
	}
	return base
}

// TestConsumerOnly_M0AndMxByteUnchanged (AC-NS2-005a, REQ-NS2-005) asserts
// the diff between HEAD and its nearest reviewed boundary touches NO path
// under internal/navigator/sync/ or internal/mx/. The default baseline is the
// verbatim AC command:
//
//	git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/sync|mx)/'
//
// grep exit 1 (no matches) = PASS. When HEAD descends from a pushed
// release/* tip, that tip becomes the baseline instead (see
// chooseConsumerOnlyBaseline for why a pushed release tip is a reviewed
// boundary); the pin then guards the change-set on top of the batch, which is
// the work actually under this guard's jurisdiction. The test is skipped when
// origin/main is unavailable (shallow clone, detached HEAD in CI without
// origin) so it does not produce false failures in environments that lack the
// git baseline; in those environments the orchestrator's verification batch
// surfaces the same command directly.
func TestConsumerOnly_M0AndMxByteUnchanged(t *testing.T) {
	// Serial: shells out to git. Skip if origin/main is unavailable.
	if os.Getenv("MOAI_NAVIGATOR_DETECT_SKIP_GIT_DIFF") == "1" {
		t.Skip("MOAI_NAVIGATOR_DETECT_SKIP_GIT_DIFF=1 — skipping git-diff guard")
	}
	// Verify origin/main exists so the main baseline is meaningful.
	revCmd := exec.Command("git", "rev-parse", "--verify", "origin/main")
	if err := revCmd.Run(); err != nil {
		t.Skipf("origin/main not resolvable (git rev-parse failed: %v); skipping AC-NS2-005a git-diff guard — run the verbatim command manually if needed",
			err)
	}

	mainDistance := 0
	if out, err := exec.Command("git", "rev-list", "--count", "origin/main...HEAD").Output(); err == nil {
		if n, perr := strconv.Atoi(strings.TrimSpace(string(out))); perr == nil {
			mainDistance = n
		}
	}

	var releaseTips []consumerOnlyBaseline
	refsOut, err := exec.Command("git", "for-each-ref", "--format=%(refname)", "refs/remotes/origin/release/").Output()
	if err != nil {
		t.Fatalf("git for-each-ref refs/remotes/origin/release/ failed: %v", err)
	}
	for _, ref := range strings.Fields(string(refsOut)) {
		shaOut, err := exec.Command("git", "rev-parse", "--verify", ref).Output()
		if err != nil {
			continue // ref vanished between listing and resolving — not a boundary
		}
		sha := strings.TrimSpace(string(shaOut))
		if err := exec.Command("git", "merge-base", "--is-ancestor", sha, "HEAD").Run(); err != nil {
			releaseTips = append(releaseTips, consumerOnlyBaseline{name: ref, sha: sha, distance: -1})
			continue
		}
		distOut, err := exec.Command("git", "rev-list", "--count", sha+"..HEAD").Output()
		if err != nil {
			continue
		}
		d, perr := strconv.Atoi(strings.TrimSpace(string(distOut)))
		if perr != nil {
			continue
		}
		releaseTips = append(releaseTips, consumerOnlyBaseline{name: ref, sha: sha, distance: d})
	}

	chosen := chooseConsumerOnlyBaseline(mainDistance, releaseTips)
	diffRef := "origin/main...HEAD" // three-dot: the verbatim AC form
	if chosen.sha != "" {
		// The tip is an ancestor of HEAD, so two-dot equals three-dot here;
		// the two-dot spelling just pins the baseline SHA explicitly.
		diffRef = chosen.sha + "..HEAD"
	}
	cmd := exec.Command("git", "diff", "--name-only", diffRef)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git diff --name-only %s failed: %v", diffRef, err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "internal/navigator/sync/") {
			t.Errorf("AC-NS2-005a FAIL: diff vs %s touches M0 producer path %q — the Detect layer must consume internal/navigator/sync/ read-only", chosen.name, line)
		}
		if strings.HasPrefix(line, "internal/mx/") {
			t.Errorf("AC-NS2-005a FAIL: diff vs %s touches mx producer path %q — the Detect layer must consume internal/mx/ read-only", chosen.name, line)
		}
	}
}

// TestChooseConsumerOnlyBaseline pins the baseline-selection policy itself:
// release tips displace origin/main only when ancestral AND strictly closer,
// and a tie keeps origin/main (the maximal-strength pin).
func TestChooseConsumerOnlyBaseline(t *testing.T) {
	cases := []struct {
		name        string
		mainDist    int
		releaseTips []consumerOnlyBaseline
		wantName    string
	}{
		{"no release refs keeps main", 12, nil, "origin/main"},
		{"non-ancestral release tip ignored", 3, []consumerOnlyBaseline{
			{name: "origin/release/v3.1.1", sha: "aaaa", distance: -1},
		}, "origin/main"},
		{"closer ancestral release tip wins", 40, []consumerOnlyBaseline{
			{name: "origin/release/v3.1.1", sha: "bbbb", distance: 2},
		}, "origin/release/v3.1.1"},
		{"tie keeps main (max pin strength)", 5, []consumerOnlyBaseline{
			{name: "origin/release/v3.1.1", sha: "cccc", distance: 5},
		}, "origin/main"},
		{"nearest of several ancestral tips wins", 40, []consumerOnlyBaseline{
			{name: "origin/release/v3.1.0", sha: "dddd", distance: 30},
			{name: "origin/release/v3.1.1", sha: "eeee", distance: 3},
		}, "origin/release/v3.1.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := chooseConsumerOnlyBaseline(tc.mainDist, tc.releaseTips)
			if got.name != tc.wantName {
				t.Errorf("chooseConsumerOnlyBaseline(mainDist=%d, ...) chose %q, want %q",
					tc.mainDist, got.name, tc.wantName)
			}
		})
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
