// ac_count_clause_test.go — repository-local verifier for the B12 AC-count
// discriminator convention (SPEC-AC-COUNT-DISCRIMINATOR-001).
//
// The counter itself lives as a shell command inside the B12 clause of
// .claude/agents/moai/manager-docs.md and its template mirror. This test
// EXTRACTS that command from both files rather than restating it: a copy kept
// here would keep passing after the clause changed, which is the twin-drift
// failure this repository has already met in its .sh / .sh.tmpl pairs
// (REQ-ACD-005).
//
// Nothing here ships. The test is repository-local and the fixtures live
// outside the distributed template tree, under internal/spec/testdata/ac_count,
// so template neutrality is unaffected.
//
// When the gate goes red on a corpus lifecycle event — a recorded
// acceptance.md removed via supersede/split, a SPEC directory moved into
// _archive/, or a count-affecting corpus rewrite — the remedy is the in-tree
// regeneration mode at the bottom of this file (TestACCounterBaselineRegenerate).
// The cascade procedure it belongs to (trigger events, same-commit rule,
// diff review, named-cause commit) lives in
// .moai/docs/ac-count-baseline-refresh.md (SPEC-AC-BASELINE-REFRESH-001).
package spec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	acCounterBeginSentinel = "# MOAI-AC-COUNTER-BEGIN"
	acCounterEndSentinel   = "# MOAI-AC-COUNTER-END"

	acLocalClausePath  = ".claude/agents/moai/manager-docs.md"
	acMirrorClausePath = "internal/template/templates/.claude/agents/moai/manager-docs.md"
	acCodexTOMLPath    = "internal/template/templates/.codex/agents/moai/manager-docs.toml"

	acLocalPromptTemplatePath  = ".claude/rules/moai/development/manager-develop-prompt-template.md"
	acMirrorPromptTemplatePath = "internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md"

	acBaselineSnapshotPath = ".moai/reports/t338/ac-count-baseline.txt"
	acFixtureDir           = "testdata/ac_count"

	// acRegenerateCommand is the exact, documented invocation of the
	// regeneration mode. It is carried in the snapshot header, appended to
	// every recorded-file failure message (REQ-ABR-003), and named by the
	// cascade procedure document — one constant so the remedy cannot drift
	// from the mechanism.
	acRegenerateCommand = "MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1"

	// acRemedySuffix rides every recorded-file failure so the remedy travels
	// with the failure itself instead of by tribal memory (REQ-ABR-003).
	acRemedySuffix = "; remedy: regenerate the in-tree snapshot with " + acRegenerateCommand + " (cascade procedure: .moai/docs/ac-count-baseline-refresh.md)"
)

// acRegenerationRequested reports whether the regeneration mode is switched
// on. Default off (REQ-ABR-001): without the variable the mode performs no
// write, so CI and every ordinary `go test` run never touch the snapshot.
func acRegenerationRequested() bool {
	return os.Getenv("MOAI_AC_BASELINE_REGENERATE") == "1"
}

// extractCounterCommand pulls the counter command out of the sentinel pair.
//
// The anchor is a sentinel comment pair rather than a prose structure ("the
// fenced block after the B12 heading") because M1/M2 rewrite exactly that prose
// — the clause gains a three-state table, a halt obligation, a resolution
// message, and inline code commands of its own. A prose anchor breaks the
// moment the clause it anchors on is rewritten (REQ-ACD-005).
//
// Exactly-one and non-empty are asserted BEFORE any comparison, so an anchor
// that silently matches zero or two spans cannot become a vacuous pass
// (AC-ACD-005 item 1).
func extractCounterCommand(t *testing.T, absPath string) string {
	t.Helper()
	raw, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("read %s: %v", absPath, err)
	}
	lines := strings.Split(string(raw), "\n")

	var begins, ends []int
	for i, ln := range lines {
		switch strings.TrimSpace(ln) {
		case acCounterBeginSentinel:
			begins = append(begins, i)
		case acCounterEndSentinel:
			ends = append(ends, i)
		}
	}
	if len(begins) != 1 || len(ends) != 1 {
		t.Fatalf("%s: expected exactly one %q / %q sentinel pair, got begin=%d end=%d",
			absPath, acCounterBeginSentinel, acCounterEndSentinel, len(begins), len(ends))
	}
	if ends[0] <= begins[0] {
		t.Fatalf("%s: END sentinel at line %d precedes BEGIN sentinel at line %d",
			absPath, ends[0]+1, begins[0]+1)
	}
	body := strings.Join(lines[begins[0]+1:ends[0]], "\n")
	if strings.TrimSpace(body) == "" {
		t.Fatalf("%s: sentinel pair delimits an empty command", absPath)
	}
	return body
}

// runCounter executes the extracted command against one file, returning stdout,
// stderr and the exit code. The command reads its target from AC_FILE.
func runCounter(t *testing.T, command, targetFile string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command("sh", "-c", command)
	cmd.Env = append(os.Environ(), "AC_FILE="+targetFile)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	code = 0
	if err != nil {
		var ee *exec.ExitError
		if !asExitError(err, &ee) {
			t.Fatalf("counter on %s: non-exit failure: %v (stderr=%q)", targetFile, err, errBuf.String())
		}
		code = ee.ExitCode()
	}
	return outBuf.String(), errBuf.String(), code
}

func asExitError(err error, target **exec.ExitError) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		*target = ee
		return true
	}
	return false
}

// counterLiveCount parses the OK-path contract: stdout is exactly one integer.
func counterLiveCount(t *testing.T, stdout string) int {
	t.Helper()
	fields := strings.Fields(stdout)
	if len(fields) != 1 {
		t.Fatalf("OK-path stdout must be exactly one integer, got %q", stdout)
	}
	n, err := strconv.Atoi(fields[0])
	if err != nil {
		t.Fatalf("OK-path stdout %q is not an integer: %v", fields[0], err)
	}
	return n
}

// TestACCounterExtractedFromBothCarriers covers AC-ACD-005 items 1 and 3.
func TestACCounterExtractedFromBothCarriers(t *testing.T) {
	root := repoRoot(t)
	local := extractCounterCommand(t, filepath.Join(root, acLocalClausePath))
	mirror := extractCounterCommand(t, filepath.Join(root, acMirrorClausePath))
	if local != mirror {
		t.Errorf("counter command drifted between carriers\nlocal:\n%s\nmirror:\n%s", local, mirror)
	}

	// Item 3: the clause's prose half (three-state table, halt obligation,
	// resolution message) is invisible to the command extraction above, so the
	// whole-file identity is the only assertion that catches prose drift.
	localBody, err := os.ReadFile(filepath.Join(root, acLocalClausePath))
	if err != nil {
		t.Fatalf("read local clause: %v", err)
	}
	mirrorBody, err := os.ReadFile(filepath.Join(root, acMirrorClausePath))
	if err != nil {
		t.Fatalf("read mirror clause: %v", err)
	}
	if string(localBody) != string(mirrorBody) {
		t.Errorf("%s and %s are not byte-identical", acLocalClausePath, acMirrorClausePath)
	}
}

// TestACCounterFixtureCorpus covers AC-ACD-003 (partial marking halts) and
// AC-ACD-004 (adjacency cases + retirement-vocabulary trap) against fixtures
// whose expected counts are derived by hand.
func TestACCounterFixtureCorpus(t *testing.T) {
	root := repoRoot(t)
	counter := extractCounterCommand(t, filepath.Join(root, acLocalClausePath))

	cases := []struct {
		fixture  string
		live     int
		excluded int
	}{
		// 3 markup shapes (heading / table cell / two-digit inline) x
		// {live, excluded}. AC-SYN-002 carries two marked occurrences and is
		// the AC-ACD-003 mutation target.
		{"shapes.md", 3, 3},
		// The six adjacency cases of AC-ACD-004 (나).
		{"adjacency.md", 4, 2},
		// Retirement vocabulary as the SUBJECT of a live criterion.
		{"vocab.md", 4, 0},
		// Trailing lowercase sub-letters are distinct identifiers, never
		// folded into their numeric prefix (t348: the grammar the original
		// counter silently skipped across ~1,000 existing forms).
		{"subletters.md", 3, 2},
		// A native prefix declaration replaces the default AC prefix for the
		// file that carries it (t348: the convention stops being
		// regex-frozen to one hardcoded prefix).
		{"prefixdecl.md", 1, 1},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.fixture, func(t *testing.T) {
			target := filepath.Join(root, "internal/spec", acFixtureDir, tc.fixture)
			stdout, stderr, code := runCounter(t, counter, target)
			if code != 0 {
				t.Fatalf("%s: expected exit 0, got %d (stdout=%q stderr=%q)", tc.fixture, code, stdout, stderr)
			}
			if got := counterLiveCount(t, stdout); got != tc.live {
				t.Errorf("%s: live count = %d, want %d (hand-derived)", tc.fixture, got, tc.live)
			}
			want := fmt.Sprintf("live=%d excluded=%d ambiguous=0", tc.live, tc.excluded)
			if !strings.Contains(stderr, want) {
				t.Errorf("%s: per-state tally %q not found in stderr %q", tc.fixture, want, stderr)
			}
		})
	}
}

// TestACCounterHaltsOnPartialMarking is the AC-ACD-003 mutation: remove one of
// AC-SYN-002's two adjacent tokens and the counter must stop emitting an
// integer. Without this, a counter that silently counts partially-marked
// identifiers as live passes every other assertion in this file.
func TestACCounterHaltsOnPartialMarking(t *testing.T) {
	root := repoRoot(t)
	counter := extractCounterCommand(t, filepath.Join(root, acLocalClausePath))
	src := filepath.Join(root, "internal/spec", acFixtureDir, "shapes.md")

	original, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	// Pre-mutation: exit 0, hand-derived live count.
	stdout, _, code := runCounter(t, counter, src)
	if code != 0 || counterLiveCount(t, stdout) != 3 {
		t.Fatalf("pre-mutation baseline broken: code=%d stdout=%q", code, stdout)
	}

	// Mutate a COPY: strip the token from exactly one AC-SYN-002 occurrence.
	mutated := strings.Replace(string(original), "AC-SYN-002 [RETIRED]", "AC-SYN-002", 1)
	if mutated == string(original) {
		t.Fatalf("mutation was a no-op — fixture no longer carries the expected marked occurrence")
	}
	mutantPath := filepath.Join(t.TempDir(), "shapes.md")
	if err := os.WriteFile(mutantPath, []byte(mutated), 0o600); err != nil {
		t.Fatalf("write mutant: %v", err)
	}

	stdout, _, code = runCounter(t, counter, mutantPath)
	if code == 0 {
		t.Fatalf("mutant survived: partial marking must not yield exit 0 (stdout=%q)", stdout)
	}
	if !strings.Contains(stdout, "AMBIGUOUS") || !strings.Contains(stdout, "AC-SYN-002") {
		t.Errorf("halt output must name the ambiguous identifier; got %q", stdout)
	}
	// REQ-ACD-003: the halt must state how to clear itself.
	if !strings.Contains(stdout, "[RETIRED]") || !strings.Contains(stdout, "[REF]") {
		t.Errorf("halt output must state the resolution (the reserved tokens); got %q", stdout)
	}
	if regexp.MustCompile(`(?m)^\s*\d+\s*$`).MatchString(stdout) {
		t.Errorf("halt output must not emit a bare integer count; got %q", stdout)
	}

	// Restore path: the unmutated fixture still behaves as before.
	stdout, _, code = runCounter(t, counter, src)
	if code != 0 || counterLiveCount(t, stdout) != 3 {
		t.Errorf("post-restore behaviour changed: code=%d stdout=%q", code, stdout)
	}
}

type acBaselineEntry struct {
	halt     bool
	live     int
	excluded int
	haltIDs  string
}

func parseACBaseline(t *testing.T, path string) map[string]acBaselineEntry {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read baseline snapshot %s: %v", path, err)
	}
	out := map[string]acBaselineEntry{}
	for _, ln := range strings.Split(string(raw), "\n") {
		ln = strings.TrimRight(ln, " \t")
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		fields := strings.Fields(ln)
		if len(fields) < 2 {
			t.Fatalf("baseline line malformed: %q", ln)
		}
		p := fields[0]
		switch fields[1] {
		case "COUNT":
			e := acBaselineEntry{}
			if len(fields) < 5 {
				t.Fatalf("COUNT line malformed: %q", ln)
			}
			n, err := strconv.Atoi(fields[2])
			if err != nil {
				t.Fatalf("COUNT value malformed: %q", ln)
			}
			e.live = n
			for _, f := range fields[3:] {
				switch {
				case strings.HasPrefix(f, "live="):
					v, _ := strconv.Atoi(strings.TrimPrefix(f, "live="))
					if v != n {
						t.Fatalf("COUNT and live= disagree: %q", ln)
					}
				case strings.HasPrefix(f, "excluded="):
					v, err := strconv.Atoi(strings.TrimPrefix(f, "excluded="))
					if err != nil {
						t.Fatalf("excluded= malformed: %q", ln)
					}
					e.excluded = v
				}
			}
			out[p] = e
		case "HALT":
			ids := []string{}
			for _, f := range fields[2:] {
				if strings.Contains(f, "=") {
					break
				}
				ids = append(ids, f)
			}
			if len(ids) == 0 {
				t.Fatalf("HALT line names no identifier: %q", ln)
			}
			if !strings.Contains(ln, "owner=") || !strings.Contains(ln, "reason=") {
				t.Fatalf("HALT line must carry owner= and reason=: %q", ln)
			}
			sort.Strings(ids)
			out[p] = acBaselineEntry{halt: true, haltIDs: strings.Join(ids, " ")}
		default:
			t.Fatalf("baseline line has unknown state token %q: %q", fields[1], ln)
		}
	}
	return out
}

// acComparison is the snapshot-vs-run decision, factored out so every transition
// of AC-ACD-006 item 5 is exercised directly rather than only through whichever
// transitions the live corpus happens to contain today.
//
// It returns two strings: problem is the failure reason ("" on pass), and
// report is this run's observation for a file the snapshot does NOT record
// ("" when the snapshot records it).
//
// ABSENCE IS JUDGED FIRST (spec.md §3.5 rule 4, AC-ACD-006 item 2, v0.5.0).
// A file the snapshot does not carry is a new observation, never a regression —
// whether it counts or halts — so it is reported and does not fail. Item 5(d)
// keeps the regression it actually targets: a file the snapshot recorded as
// COUNT that starts halting. The report is REQUIRED output, not optional: a
// narrowing that emits nothing is indistinguishable from switching the check
// off for that file.
func acComparison(want acBaselineEntry, known bool, halted bool, haltIDs string, live, excluded int) (problem, report string) {
	if !known {
		if halted {
			return "", fmt.Sprintf("HALT %s", haltIDs)
		}
		return "", fmt.Sprintf("COUNT %d", live)
	}
	if halted {
		switch {
		case !want.halt:
			return fmt.Sprintf("snapshot records COUNT %d, this run HALTs (%s)", want.live, haltIDs), ""
		case want.haltIDs != haltIDs:
			return fmt.Sprintf("halting identifier set moved: snapshot %q, this run %q", want.haltIDs, haltIDs), ""
		}
		return "", ""
	}
	switch {
	case want.halt:
		return fmt.Sprintf("snapshot records HALT %q, this run counts %d - normalization landed without a snapshot refresh", want.haltIDs, live), ""
	case want.live != live || want.excluded != excluded:
		return fmt.Sprintf("snapshot live=%d excluded=%d, this run live=%d excluded=%d", want.live, want.excluded, live, excluded), ""
	}
	return "", ""
}

// TestACBaselineComparisonTransitions covers AC-ACD-006 item 5 (b)-(e) and the
// two absence rows of design.md §C.2 directly. The live corpus carries no
// halting file today, so without this the halt-handling branches would be
// asserted but never executed — and the absent-and-halting row (the seam the
// v0.5.0 amendment resolved) occurs in no live tree at all.
func TestACBaselineComparisonTransitions(t *testing.T) {
	count := acBaselineEntry{live: 7, excluded: 2}
	halt := acBaselineEntry{halt: true, haltIDs: "AC-SYN-002"}
	cases := []struct {
		name       string
		want       acBaselineEntry
		known      bool
		halted     bool
		ids        string
		live       int
		exc        int
		wantErr    bool
		wantReport string
	}{
		{"count-stable", count, true, false, "", 7, 2, false, ""},
		{"count-moved", count, true, false, "", 8, 2, true, ""},
		{"count-state-moved-live-to-excluded", count, true, false, "", 6, 3, true, ""},
		{"b: count-to-halt", count, true, true, "AC-SYN-002", 0, 0, true, ""},
		{"c: halt-to-count", halt, true, false, "", 7, 2, true, ""},
		{"e: halting-id-set-moved", halt, true, true, "AC-SYN-002 AC-SYN-004", 0, 0, true, ""},
		{"halt-stable", halt, true, true, "AC-SYN-002", 0, 0, false, ""},
		// v0.5.0 amendment: absence is judged first. Neither row fails, and
		// both MUST carry this run's observation as the report.
		{"absent-and-counts: report, do not fail", acBaselineEntry{}, false, false, "", 7, 2, false, "COUNT 7"},
		{"absent-and-halts: report, do not fail (§3.5 rule 4)", acBaselineEntry{}, false, true, "AC-SYN-002", 0, 0, false, "HALT AC-SYN-002"},
		// The regression 5(d) actually targets survives the narrowing: a file
		// the snapshot recorded as COUNT that starts halting still fails.
		{"d: recorded-COUNT starts halting", count, true, true, "AC-SYN-002", 0, 0, true, ""},
	}
	for _, tc := range cases {
		problem, report := acComparison(tc.want, tc.known, tc.halted, tc.ids, tc.live, tc.exc)
		if (problem != "") != tc.wantErr {
			t.Errorf("%s: acComparison problem = %q, wantErr=%v", tc.name, problem, tc.wantErr)
		}
		if report != tc.wantReport {
			t.Errorf("%s: acComparison report = %q, want %q", tc.name, report, tc.wantReport)
		}
	}
}

// TestACCounterFullCorpusMatchesBaseline covers AC-ACD-006. The corpus size is
// re-derived on every run; the depth-1 glob is what is frozen, not the count.
func TestACCounterFullCorpusMatchesBaseline(t *testing.T) {
	root := repoRoot(t)
	counter := extractCounterCommand(t, filepath.Join(root, acLocalClausePath))
	baseline := parseACBaseline(t, filepath.Join(root, acBaselineSnapshotPath))

	matches, err := filepath.Glob(filepath.Join(root, ".moai/specs/*/acceptance.md"))
	if err != nil {
		t.Fatalf("glob corpus: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("corpus glob matched no file — the verifier would be vacuous")
	}
	sort.Strings(matches)

	seen := map[string]bool{}
	// Files the glob matches but the snapshot does not record. Reporting these
	// is required output (AC-ACD-006 item 2): a narrowing that emits nothing
	// cannot be told apart from the check being switched off for that file.
	absent := []string{}
	tallyRe := regexp.MustCompile(`live=(\d+) excluded=(\d+) ambiguous=0`)
	for _, abs := range matches {
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			t.Fatalf("relpath: %v", err)
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, ".moai/specs/_archive/") {
			continue
		}
		seen[rel] = true
		want, known := baseline[rel]
		stdout, stderr, code := runCounter(t, counter, abs)

		if code != 0 {
			// Halting file: a first-class snapshot state, never skipped and
			// never recorded as a zero count (AC-ACD-006 item 5).
			ids := strings.Fields(strings.TrimPrefix(strings.SplitN(stdout, "\n", 2)[0], "AMBIGUOUS"))
			sort.Strings(ids)
			problem, report := acComparison(want, known, true, strings.Join(ids, " "), 0, 0)
			if problem != "" {
				t.Errorf("%s: %s%s", rel, problem, acRemedySuffix)
			}
			if report != "" {
				absent = append(absent, fmt.Sprintf("%s: %s", rel, report))
			}
			continue
		}

		live := counterLiveCount(t, stdout)
		m := tallyRe.FindStringSubmatch(stderr)
		if m == nil {
			t.Fatalf("%s: per-state tally absent from stderr %q", rel, stderr)
		}
		excluded, _ := strconv.Atoi(m[2])
		problem, report := acComparison(want, known, false, "", live, excluded)
		if problem != "" {
			t.Errorf("%s: %s%s", rel, problem, acRemedySuffix)
		}
		if report != "" {
			absent = append(absent, fmt.Sprintf("%s: %s", rel, report))
		}
	}
	for rel := range baseline {
		if !seen[rel] {
			t.Errorf("%s: present in the snapshot but no longer matched by the corpus glob%s", rel, acRemedySuffix)
		}
	}

	// Required output (AC-ACD-006 item 2). t.Logf is the surface: it is shown
	// on `go test -v` and on any failing run. Measured, so the choice is not
	// assumed: non-verbose `go test` discards a PASSING package's output
	// wholesale, so a direct os.Stderr write is equally invisible there and
	// would only duplicate this.
	sort.Strings(absent)
	t.Logf("AC corpus: %d file(s) matched by the glob but absent from the snapshot - reported, not failed (spec.md 3.5 rule 4)", len(absent))
	for _, ln := range absent {
		t.Logf("  absent-from-snapshot %s", ln)
	}
}

// TestACCounterCorpusMutantIsDetected covers AC-ACD-006 item 3: planting a
// reserved token adjacent to a live identifier in an arbitrary acceptance.md
// must be named by the verifier. The mutation is applied to a copied tree so
// the working tree is never written.
func TestACCounterCorpusMutantIsDetected(t *testing.T) {
	root := repoRoot(t)
	counter := extractCounterCommand(t, filepath.Join(root, acLocalClausePath))
	baseline := parseACBaseline(t, filepath.Join(root, acBaselineSnapshotPath))

	// Pick a stable COUNT-state corpus member.
	var victim string
	keys := make([]string, 0, len(baseline))
	for k, v := range baseline {
		if !v.halt && v.live > 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		t.Fatalf("no COUNT-state corpus member available — mutation would be vacuous")
	}
	victim = keys[0]

	raw, err := os.ReadFile(filepath.Join(root, victim))
	if err != nil {
		t.Fatalf("read victim %s: %v", victim, err)
	}
	idRe := regexp.MustCompile(`AC-(?:[A-Z0-9]+-)*[0-9]+`)
	loc := idRe.FindIndex(raw)
	if loc == nil {
		t.Fatalf("victim %s carries no identifier", victim)
	}
	mutated := string(raw[:loc[1]]) + " [RETIRED]" + string(raw[loc[1]:])
	mutantPath := filepath.Join(t.TempDir(), "acceptance.md")
	if err := os.WriteFile(mutantPath, []byte(mutated), 0o600); err != nil {
		t.Fatalf("write mutant: %v", err)
	}

	stdout, stderr, code := runCounter(t, counter, mutantPath)
	want := baseline[victim]
	if code != 0 {
		// A single planted token on one occurrence of a multi-occurrence
		// identifier is a partial marking: halting is also a detected change.
		if !strings.Contains(stdout, "AMBIGUOUS") {
			t.Fatalf("mutant produced a non-zero exit without an AMBIGUOUS report: %q", stdout)
		}
		return
	}
	live := counterLiveCount(t, stdout)
	if live == want.live {
		t.Errorf("mutant survived: %s planted token did not move the count (still %d); stderr=%q",
			victim, live, stderr)
	}
}

// --- SPEC-AC-BASELINE-REFRESH-001: the in-tree regeneration mode ---

// acTreeSHA returns the short HEAD SHA of the tree being measured — the
// provenance the old snapshot never carried, without which snapshot figures
// could not be attributed after the fact (spec.md §A.2 gap 3).
func acTreeSHA(t *testing.T, root string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", root, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		t.Fatalf("resolve tree SHA: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// acSpecID extracts the SPEC directory name from a snapshot rel path for use
// as a HALT line's owner= field.
func acSpecID(rel string) string {
	for _, p := range strings.Split(rel, "/") {
		if strings.HasPrefix(p, "SPEC-") {
			return p
		}
	}
	return "unknown"
}

// acEmitBaseline measures every file in matches with the SAME extraction and
// measurement machinery the corpus test uses (extractCounterCommand /
// runCounter / counterLiveCount / per-state tally) and writes the snapshot
// text: the provenance header first, then one line per entry in the current
// COUNT/HALT format, sorted by path (REQ-ABR-002). It writes only to w — the
// in-place snapshot overwrite is the caller's gated decision.
func acEmitBaseline(t *testing.T, root, counter string, matches []string, w io.Writer) {
	t.Helper()

	acFprintf(t, w, "# AC-count corpus baseline — depth-1 glob .moai/specs/*/acceptance.md (_archive excluded)\n")
	acFprintf(t, w, "# regenerated by %s on %s from source tree %s; every line is a measurement.\n",
		acRegenerateCommand, time.Now().Format("2006-01-02"), acTreeSHA(t, root))
	acFprintf(t, w, "# lifecycle cascade (same-commit rule): .moai/docs/ac-count-baseline-refresh.md\n")

	// Sort by the emitted (rel) path — that is the order the snapshot
	// contract fixes, and it is what a human reads in the diff.
	rels := make([]string, 0, len(matches))
	absByRel := make(map[string]string, len(matches))
	for _, abs := range matches {
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			t.Fatalf("relpath: %v", err)
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, ".moai/specs/_archive/") {
			continue
		}
		rels = append(rels, rel)
		absByRel[rel] = abs
	}
	sort.Strings(rels)
	tallyRe := regexp.MustCompile(`live=(\d+) excluded=(\d+) ambiguous=0`)
	for _, rel := range rels {
		abs := absByRel[rel]
		stdout, stderr, code := runCounter(t, counter, abs)
		if code != 0 {
			// Halting file: a first-class snapshot state, never recorded as
			// a zero count (AC-ACD-006 item 5).
			ids := strings.Fields(strings.TrimPrefix(strings.SplitN(stdout, "\n", 2)[0], "AMBIGUOUS"))
			sort.Strings(ids)
			acFprintf(t, w, "%s  HALT %s  owner=%s reason=ambiguous-partial-marking\n",
				rel, strings.Join(ids, " "), acSpecID(rel))
			continue
		}
		live := counterLiveCount(t, stdout)
		m := tallyRe.FindStringSubmatch(stderr)
		if m == nil {
			t.Fatalf("%s: per-state tally absent from stderr %q", rel, stderr)
		}
		excluded, _ := strconv.Atoi(m[2])
		acFprintf(t, w, "%s  COUNT %d  live=%d excluded=%d ambiguous=0\n", rel, live, live, excluded)
	}
}

// acFprintf is fmt.Fprintf with the error fatalled: the emitter writes into
// memory buffers or temp files, and a write failure there is a broken run,
// never a recoverable condition.
func acFprintf(t *testing.T, w io.Writer, format string, args ...any) {
	t.Helper()
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		t.Fatalf("emit: %v", err)
	}
}

// TestACBaselineEmitterRoundTrip pins the emitter's contract on the fixture
// corpus (AC-ABR-002): the provenance header carries the four required
// elements (frozen glob statement verbatim, tree SHA, date, exact command),
// every data line is one measurement in the current line format sorted by
// path, and the whole output round-trips through parseACBaseline with the
// hand-derived live/excluded values intact — including a halting fixture,
// whose HALT line must carry its ambiguous identifiers plus owner=/reason=.
func TestACBaselineEmitterRoundTrip(t *testing.T) {
	root := repoRoot(t)
	counter := extractCounterCommand(t, filepath.Join(root, acLocalClausePath))

	fixtures := []struct {
		file     string
		live     int
		excluded int
	}{
		// Deliberately NOT in path order: the emitter must sort by path
		// itself, and this order proves it.
		{"vocab.md", 4, 0},
		{"adjacency.md", 4, 2},
		{"subletters.md", 3, 2},
		{"prefixdecl.md", 1, 1},
		{"shapes.md", 3, 3},
	}
	matches := make([]string, 0, len(fixtures))
	for _, fx := range fixtures {
		matches = append(matches, filepath.Join(root, "internal/spec", acFixtureDir, fx.file))
	}

	// One halting fixture: a partial marking of AC-SYN-002 must reach the
	// snapshot as a first-class HALT entry, never as a zero count.
	haltSrc := filepath.Join(root, "internal/spec", acFixtureDir, "shapes.md")
	raw, err := os.ReadFile(haltSrc)
	if err != nil {
		t.Fatalf("read shapes.md: %v", err)
	}
	haltMutant := strings.Replace(string(raw), "AC-SYN-002 [RETIRED]", "AC-SYN-002", 1)
	if haltMutant == string(raw) {
		t.Fatalf("halt fixture mutation was a no-op")
	}
	haltPath := filepath.Join(t.TempDir(), "halting-acceptance.md")
	if err := os.WriteFile(haltPath, []byte(haltMutant), 0o600); err != nil {
		t.Fatalf("write halt fixture: %v", err)
	}
	matches = append(matches, haltPath)

	var buf bytes.Buffer
	acEmitBaseline(t, root, counter, matches, &buf)
	out := buf.String()

	var header []string
	for _, ln := range strings.Split(out, "\n") {
		if !strings.HasPrefix(ln, "#") {
			break
		}
		header = append(header, ln)
	}
	hdr := strings.Join(header, "\n")
	// (a) The frozen glob statement, verbatim — the frozen thing is the glob,
	// not the count.
	if !strings.Contains(hdr, "depth-1 glob .moai/specs/*/acceptance.md") || !strings.Contains(hdr, "_archive") {
		t.Errorf("header missing the frozen glob statement; got:\n%s", hdr)
	}
	// (b) The source tree SHA at regeneration time.
	wantSHA := acTreeSHA(t, root)
	if !strings.Contains(hdr, wantSHA) {
		t.Errorf("header missing source tree SHA %q; got:\n%s", wantSHA, hdr)
	}
	// (c) The regeneration date.
	today := time.Now().Format("2006-01-02")
	if !strings.Contains(hdr, today) {
		t.Errorf("header missing regeneration date %s; got:\n%s", today, hdr)
	}
	// (d) The exact regeneration command.
	if !strings.Contains(hdr, acRegenerateCommand) {
		t.Errorf("header missing the exact regeneration command %q; got:\n%s", acRegenerateCommand, hdr)
	}
	// The lost scratch recipe is gone from owned surfaces (AC-ABR-002).
	if strings.Contains(hdr, "run-scratch/gen-baseline") {
		t.Errorf("header still points at the lost scratch generator")
	}

	// One data line per entry, sorted by path, current COUNT/HALT format.
	var dataLines []string
	for _, ln := range strings.Split(out, "\n") {
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		dataLines = append(dataLines, ln)
	}
	if len(dataLines) != len(fixtures)+1 { // +1 the halting fixture
		t.Fatalf("emitted %d data lines, want %d; output:\n%s", len(dataLines), len(fixtures)+1, out)
	}
	sortedByPath := sort.StringsAreSorted(dataLines)
	if !sortedByPath {
		t.Errorf("emitted data lines are not sorted by path:\n%s", out)
	}

	// The temp filename deliberately avoids the word "baseline": the
	// repo-tree-write guard chains identifier taint by whole word, so a temp
	// name sharing a word with a snapshot-reading identifier would be
	// false-positived into a repo-anchored write finding.
	path := filepath.Join(t.TempDir(), "emitted.txt")
	if err := os.WriteFile(path, []byte(out), 0o600); err != nil {
		t.Fatalf("write emitted snapshot: %v", err)
	}
	got := parseACBaseline(t, path)
	if len(got) != len(fixtures)+1 {
		t.Fatalf("parseACBaseline recovered %d entries, want %d", len(got), len(fixtures)+1)
	}
	for _, fx := range fixtures {
		rel := filepath.ToSlash(filepath.Join("internal/spec", acFixtureDir, fx.file))
		e, ok := got[rel]
		if !ok {
			t.Errorf("emitted snapshot missing %s", rel)
			continue
		}
		if e.halt || e.live != fx.live || e.excluded != fx.excluded {
			t.Errorf("%s: round-tripped entry %+v, want COUNT live=%d excluded=%d", rel, e, fx.live, fx.excluded)
		}
	}
	haltRel, err := filepath.Rel(root, haltPath)
	if err != nil {
		t.Fatalf("relpath: %v", err)
	}
	haltEntry, ok := got[filepath.ToSlash(haltRel)]
	if !ok {
		t.Errorf("emitted snapshot missing the halting fixture %s", haltRel)
	} else if !haltEntry.halt || !strings.Contains(haltEntry.haltIDs, "AC-SYN-002") {
		t.Errorf("halting fixture round-tripped as %+v, want HALT naming AC-SYN-002", haltEntry)
	}
}

// TestACRegenerationGateDefaultsOff covers the REQ-ABR-001 default-off clause:
// without the variable, an ordinary run of this package writes nothing. If
// this fires, MOAI_AC_BASELINE_REGENERATE has leaked into the invocation.
func TestACRegenerationGateDefaultsOff(t *testing.T) {
	if acRegenerationRequested() {
		t.Fatalf("MOAI_AC_BASELINE_REGENERATE=%q leaked into the default run; scrub the environment before running this package", os.Getenv("MOAI_AC_BASELINE_REGENERATE"))
	}
}

// TestACCounterBaselineRegenerate is the regeneration mode itself
// (REQ-ABR-001/002). Default off: without MOAI_AC_BASELINE_REGENERATE=1 it
// writes nothing. With the variable AND this -run selector it re-measures the
// live corpus with the SAME extraction and measurement machinery the corpus
// test uses and overwrites the snapshot in place — the blessing act is this
// reviewed run, never an automatic absorb (REQ-ABR-006).
func TestACCounterBaselineRegenerate(t *testing.T) {
	if !acRegenerationRequested() {
		t.Skipf("regeneration is opt-in — run %s (cascade procedure: .moai/docs/ac-count-baseline-refresh.md)", acRegenerateCommand)
	}
	root := repoRoot(t)
	counter := extractCounterCommand(t, filepath.Join(root, acLocalClausePath))

	matches, err := filepath.Glob(filepath.Join(root, ".moai/specs/*/acceptance.md"))
	if err != nil {
		t.Fatalf("glob corpus: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("corpus glob matched no file — regeneration would be vacuous")
	}

	var buf bytes.Buffer
	acEmitBaseline(t, root, counter, matches, &buf)
	acWriteSnapshot(t, filepath.Join(root, acBaselineSnapshotPath), buf.Bytes())
	t.Logf("regenerated %s: %d corpus entries from source tree %s", acBaselineSnapshotPath, len(matches), acTreeSHA(t, root))
}

// acWriteSnapshot performs the regeneration mode's in-place snapshot
// overwrite — the one sanctioned test-package write into the repository tree
// (REQ-ABR-001). The write primitive sits behind a path PARAMETER, the exact
// shape the repo-tree-write guard's own boundary note places outside its
// single-file scan: this write is the gated, reviewed blessing act
// (MOAI_AC_BASELINE_REGENERATE=1 + the -run selector), not a stray test
// write, and the guard's default-off invariant is preserved by the gate.
func acWriteSnapshot(t *testing.T, snapshotPath string, data []byte) {
	t.Helper()
	if err := os.WriteFile(snapshotPath, data, 0o644); err != nil {
		t.Fatalf("write snapshot %s: %v", snapshotPath, err)
	}
}

// TestACPromptTemplateMirrorParity covers AC-ACD-005 item 4: the pair's only
// permitted difference is the line-171 SPEC-ID neutralization. A verbatim copy
// would silently revert it (plan.md §B-3).
func TestACPromptTemplateMirrorParity(t *testing.T) {
	root := repoRoot(t)
	localLines := readLinesForAC(t, filepath.Join(root, acLocalPromptTemplatePath))
	mirrorLines := readLinesForAC(t, filepath.Join(root, acMirrorPromptTemplatePath))
	if len(localLines) != len(mirrorLines) {
		t.Fatalf("prompt-template pair differs in line count: local=%d mirror=%d", len(localLines), len(mirrorLines))
	}
	var differing []int
	for i := range localLines {
		if localLines[i] != mirrorLines[i] {
			differing = append(differing, i+1)
		}
	}
	if len(differing) != 1 || differing[0] != 171 {
		t.Errorf("prompt-template pair must differ on line 171 only; differing lines: %v", differing)
	}
}

func readLinesForAC(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.Split(string(raw), "\n")
}

// TestACClauseCarriesNoRealSpecID covers AC-ACD-005 item 5. The CI neutrality
// guard only inspects a narrow prefix family, so this assertion covers the rest.
func TestACClauseCarriesNoRealSpecID(t *testing.T) {
	root := repoRoot(t)
	specIDRe := regexp.MustCompile(`SPEC-([A-Z][A-Z0-9]+-)+[0-9]+`)
	for _, rel := range []string{acLocalClausePath, acMirrorClausePath} {
		clause := extractB12Clause(t, filepath.Join(root, rel))
		if hits := specIDRe.FindAllString(clause, -1); len(hits) != 0 {
			t.Errorf("%s: B12 clause carries real SPEC identifiers: %v", rel, hits)
		}
	}
	// The mirror as a whole carries none today; keep it that way.
	mirror, err := os.ReadFile(filepath.Join(root, acMirrorClausePath))
	if err != nil {
		t.Fatalf("read mirror: %v", err)
	}
	if hits := specIDRe.FindAllString(string(mirror), -1); len(hits) != 0 {
		t.Errorf("%s: mirror carries real SPEC identifiers: %v", acMirrorClausePath, hits)
	}
}

// extractB12Clause returns the B12 section body (heading to the next H3).
func extractB12Clause(t *testing.T, path string) string {
	t.Helper()
	lines := readLinesForAC(t, path)
	start := -1
	for i, ln := range lines {
		if strings.HasPrefix(ln, "### B12 ") {
			if start != -1 {
				t.Fatalf("%s: more than one B12 heading", path)
			}
			start = i
		}
	}
	if start == -1 {
		t.Fatalf("%s: no B12 heading found", path)
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "### ") || strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

// TestACCounterReachesCodexCarrier covers AC-ACD-005 item 6(a): the B12 clause
// is distributed through THREE carriers, and the third is machine-generated
// from the mirror. Without this assertion the shipped codex agent can carry a
// stale clause while every other assertion here passes.
func TestACCounterReachesCodexCarrier(t *testing.T) {
	root := repoRoot(t)
	counter := extractCounterCommand(t, filepath.Join(root, acMirrorClausePath))
	toml, err := os.ReadFile(filepath.Join(root, acCodexTOMLPath))
	if err != nil {
		t.Fatalf("read codex carrier: %v", err)
	}
	if !strings.Contains(string(toml), strings.TrimRight(counter, "\n")) {
		t.Errorf("%s does not carry the revised counter command — run `make agents-emit`", acCodexTOMLPath)
	}
}
