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
package spec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
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
)

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
				t.Errorf("%s: %s", rel, problem)
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
			t.Errorf("%s: %s", rel, problem)
		}
		if report != "" {
			absent = append(absent, fmt.Sprintf("%s: %s", rel, report))
		}
	}
	for rel := range baseline {
		if !seen[rel] {
			t.Errorf("%s: present in the snapshot but no longer matched by the corpus glob", rel)
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
