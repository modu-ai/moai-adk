package rosterguard

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// ── M2: noun class, adjacency, numeral selection ─────────────────────────────

// TestNumeralSelectionIsNearestPreceding pins the selection rule that decides
// what a hit MEANS.
//
// The first row is the LIVE string, read in this tree at
// internal/harness/delegationmap/types.go:73. It is the load-bearing case
// because it is the shape that actually occurs: a leftmost-numeral rule reads
// the section marker and reports 4, and the registry's own CountPattern comment
// cites this same string as the reason that pattern is a regexp rather than a
// line anchor. The synthetic variant is kept as an additional row because it
// differs in the section marker and the intervening word count, so it exercises
// a different adjacency.
func TestNumeralSelectionIsNearestPreceding(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
		axis string
	}{
		{
			name: "live: CLAUDE.md section-4 citation",
			line: "// CLAUDE.md §4 (the 13 retained agents — 12 MoAI-custom plus the",
			want: "13",
			axis: "digit",
		},
		{
			name: "synthetic: no section marker",
			line: "CLAUDE.md 4 (13 retained agents)",
			want: "13",
			axis: "digit",
		},
		{
			name: "attached numeral: N-agent catalog",
			line: "the 11-agent catalog is stale",
			want: "11",
			axis: "digit",
		},
		{
			name: "noun inside a longer word does not match",
			line: "see the retained agents-reference for detail",
			want: "",
		},
		{
			name: "a sentence boundary blocks the adjacency",
			line: "there are 13. The agent roster lives elsewhere",
			want: "",
		},
		{
			name: "a numeral beyond the adjacency window does not reach the noun",
			line: "13 aaaaaaaaaaaaaaaaaaaaaa retained agents",
			want: "",
		},
		{
			name: "word axis fires on synthetic input",
			line: "thirteen retained agents are listed",
			want: "thirteen",
			axis: "word",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ScanNumeralLine(tc.line)
			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("want no hit, got %+v", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("want exactly one hit, got %+v", got)
			}
			if got[0].Numeral != tc.want {
				t.Errorf("selected numeral = %q, want %q (nearest preceding, not leftmost)", got[0].Numeral, tc.want)
			}
			if got[0].Axis != tc.axis {
				t.Errorf("axis = %q, want %q", got[0].Axis, tc.axis)
			}
		})
	}
}

// TestNeutralisationRunsBeforeMatching is AC-RNA-005(a) and (b).
//
// Each shape gets its own case rather than being an accident of the window
// width: a selector predicates over ONE member of a set, and a subset
// predication states that some members have a property — neither states a size,
// and under the REQ-RNA-007 order neither reaches the match stage at all, so
// neither enters the breadth set.
func TestNeutralisationRunsBeforeMatching(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		wantHit bool
	}{
		{
			name: "selector: one of the N retained agents",
			line: "manager-spec is one of the 11 retained agents",
		},
		{
			name: "subset predication, live shape at .claude/agents/harness/workflow-specialist.md:52",
			line: "All four are retained agents.",
		},
		{
			name: "word-form subset predication",
			line: "four are retained agents",
		},
		{
			name:    "positive control: a bare size claim still fires",
			line:    "the 11 retained agents are listed below",
			wantHit: true,
		},
		{
			name:    "positive control: the word axis is alive",
			line:    "thirteen retained agents",
			wantHit: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ScanNumeralLine(tc.line)
			if tc.wantHit && len(got) == 0 {
				t.Fatalf("want a hit (an empty result here is indistinguishable from a dead regexp), got none")
			}
			if !tc.wantHit && len(got) != 0 {
				t.Fatalf("want no hit — neutralisation runs BEFORE matching, so this phrase must enter neither set; got %+v", got)
			}
		})
	}
}

// ── M1: the discharge model ──────────────────────────────────────────────────

// TestNumeralDischargeFollowsTheD2Rule is AC-RNA-002 and AC-RNA-004.
func TestNumeralDischargeFollowsTheD2Rule(t *testing.T) {
	hit := NumeralHit{Path: "some/file.md", Line: 7, Numeral: "11", Phrase: "11 retained agents", Axis: "digit"}

	countRow := Site{ID: "row", Path: "some/file.md", Axis: AxisRetainedRoster, Claims: ClaimCount, CountPattern: `(\d+) retained agents`}
	membershipRow := Site{ID: "row", Path: "some/file.md", Axis: AxisRetainedRoster, Claims: ClaimMembership}

	cases := []struct {
		name    string
		sites   []Site
		exempts []NumeralExempt
		want    string
	}{
		{
			name:  "a ClaimCount row for the path discharges",
			sites: []Site{countRow},
			want:  "",
		},
		{
			name:    "a numeral-exempt declaration with a reason discharges",
			exempts: []NumeralExempt{{ID: "x", Path: "some/file.md", Reason: "a historical citation"}},
			want:    "",
		},
		{
			name:  "a membership-only row does NOT discharge a count claim",
			sites: []Site{membershipRow},
			want:  "a membership registration does not discharge a count claim",
		},
		{
			name: "an unregistered, unexempted path is an undeclared count claim",
			want: "undeclared count claim",
		},
		{
			name:    "an empty exempt reason is a finding, never a suppression",
			exempts: []NumeralExempt{{ID: "x", Path: "some/file.md", Reason: "   "}},
			want:    "declaration is incomplete",
		},
		{
			name:    "a count row on ANOTHER path does not discharge this one",
			sites:   []Site{{ID: "other", Path: "other/file.md", Claims: ClaimCount, CountPattern: `(\d+)`}},
			exempts: nil,
			want:    "undeclared count claim",
		},
		{
			name:    "a membership row and a separate count row on one path: the count row discharges",
			sites:   []Site{membershipRow, countRow},
			exempts: nil,
			want:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DischargeNumeralHits([]NumeralHit{hit}, tc.sites, tc.exempts)
			joined := strings.Join(got, " | ")
			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("want the hit discharged (breadth set only, no finding), got: %s", joined)
				}
				return
			}
			if !strings.Contains(joined, tc.want) {
				t.Fatalf("want a finding containing %q, got: %s", tc.want, joined)
			}
		})
	}
}

// TestUndeclaredCountClaimNamesBothDischargePaths is AC-RNA-001: the message
// must name the observed numeral, the matched phrase, and BOTH ways out.
func TestUndeclaredCountClaimNamesBothDischargePaths(t *testing.T) {
	hit := NumeralHit{Path: "new/file.md", Line: 3, Numeral: "9", Phrase: "9 retained agents", Axis: "digit"}
	got := strings.Join(DischargeNumeralHits([]NumeralHit{hit}, nil, nil), " | ")
	for _, want := range []string{"new/file.md", "9 retained agents", "ClaimCount", "numeral exemption"} {
		if !strings.Contains(got, want) {
			t.Errorf("finding does not mention %q: %s", want, got)
		}
	}
}

// ── M2: layer scope ──────────────────────────────────────────────────────────

// TestNumeralScopeIsTheSweepScopePlusResearch is AC-RNA-015. The research file
// is asserted OUT OF SCOPE rather than exempted: an exempt row would make it a
// hit that happens to be discharged, which is a different claim.
func TestNumeralScopeIsTheSweepScopePlusResearch(t *testing.T) {
	out := []string{
		".moai/research/anthropic-best-practices-2026-05-24.md",
		"docs-site/content/en/page.md",
		".moai/specs/SPEC-X/spec.md",
		"CHANGELOG.md",
	}
	for _, rel := range out {
		if InNumeralScope(rel) {
			t.Errorf("%s must be OUT of the numeral layer's scope", rel)
		}
	}
	in := []string{"CLAUDE.md", ".claude/rules/moai/development/model-policy.md", "internal/web/agentfm.go"}
	for _, rel := range in {
		if !InNumeralScope(rel) {
			t.Errorf("%s must be IN the numeral layer's scope", rel)
		}
	}
	// The exclusion set is the sweep's own list plus exactly one addition; a
	// second, independently-typed list is what this asserts against.
	got := numeralSkipPrefixes()
	if len(got) != len(sweepSkipPrefixes)+1 {
		t.Errorf("numeral exclusions = %d entries, want the sweep's %d plus exactly one (.moai/research/)", len(got), len(sweepSkipPrefixes))
	}
}

// ── M4: the live layer ───────────────────────────────────────────────────────

// liveHits memoises the tree walk. Four tests read the same breadth set, and
// walking ~6k files once per test dominated the package's runtime.
var (
	liveHits []NumeralHit
	liveErr  error
	liveDone bool
)

func scanLive(t *testing.T) []NumeralHit {
	t.Helper()
	if !liveDone {
		liveHits, liveErr = ScanNumeralAxis(repoRoot(t))
		liveDone = true
	}
	if liveErr != nil {
		t.Fatalf("numeral scan: %v", liveErr)
	}
	return liveHits
}

// TestLiveWordAxisBreadthSetIsEmptyOutsideThisPackage is AC-RNA-005(c).
//
// The word axis is a STRUCTURAL hole: on this tree its live breadth set is
// empty outside this package's own fixtures, and the run evidence names the
// cause rather than an expectation. The single live candidate is
// .claude/agents/harness/workflow-specialist.md:52 — "All four are retained
// agents." — a SUBSET PREDICATION the neutraliser removes BEFORE the axes
// match, so it enters neither set. That line is read from disk here rather
// than quoted, so the criterion fails if the line changes shape.
//
// This package's own files are excluded from the count because a guard whose
// subject matter is roster count claims necessarily quotes them; they carry
// declared exemptions for exactly that reason.
func TestLiveWordAxisBreadthSetIsEmptyOutsideThisPackage(t *testing.T) {
	const liveCandidate = ".claude/agents/harness/workflow-specialist.md"
	const liveCandidateLine = 52

	body, err := os.ReadFile(filepath.Join(repoRoot(t), liveCandidate))
	if err != nil {
		t.Fatalf("read %s: %v", liveCandidate, err)
	}
	lines := strings.Split(string(body), "\n")
	if len(lines) < liveCandidateLine {
		t.Fatalf("%s has %d lines, fewer than %d — re-measure the live candidate", liveCandidate, len(lines), liveCandidateLine)
	}
	line := lines[liveCandidateLine-1]
	if !strings.Contains(line, "are retained agents") {
		t.Fatalf("%s:%d no longer reads as a subset predication (%q) — re-measure the cause of the empty word-axis result", liveCandidate, liveCandidateLine, line)
	}
	if hits := ScanNumeralLine(line); len(hits) != 0 {
		t.Errorf("%s:%d %q produced %+v; the neutraliser must remove a subset predication BEFORE the axes match", liveCandidate, liveCandidateLine, line, hits)
	}
	t.Logf("live word-axis candidate %s:%d = %q — removed by subset-predication neutralisation, so it enters neither set",
		liveCandidate, liveCandidateLine, strings.TrimSpace(line))

	for _, h := range scanLive(t) {
		if h.Axis != "word" {
			continue
		}
		if strings.HasPrefix(h.Path, "internal/harness/rosterguard/") {
			continue
		}
		t.Errorf("unexpected live word-axis hit %s:%d %q — the axis was measured empty outside this package; adjudicate it rather than widening an exemption", h.Path, h.Line, h.Phrase)
	}
}

// TestNumeralAxisFindsNoUndeclaredCountClaim is the live half of the layer.
//
// The BREADTH set is printed — every post-neutralisation match, discharged or
// not — one line per hit at column 0, so the set can be compared against the
// enumeration rather than merely counted. Printing only the finding set would
// make a green tree print nothing, and nothing would then constrain breadth: a
// layer implementing one noun form passes every shell-level criterion while
// missing most sites.
//
// fmt.Println rather than t.Logf on purpose: AC-RNA-006(a) greps the output
// anchored at column 0 ('^digit-axis hit '), and t.Logf indents its output and
// prefixes it with file:line.
func TestNumeralAxisFindsNoUndeclaredCountClaim(t *testing.T) {
	hits := scanLive(t)
	for _, h := range hits {
		fmt.Printf("%s-axis hit %s:%d %s\n", h.Axis, h.Path, h.Line, h.Phrase)
	}

	// AC-RNA-008: an empty breadth set is a measurement failure, not a clean
	// tree. The registry is non-empty by construction and carries count claims,
	// so the layer must at minimum rediscover them.
	if len(hits) == 0 {
		t.Fatal("the numeral layer produced an EMPTY breadth set in a tree that demonstrably carries roster count claims — the walk, the noun class, the neutraliser or the exclusion list is broken. This is a measurement failure, not a clean tree")
	}

	for _, v := range DischargeNumeralHits(hits, Registry(), NumeralExemptions()) {
		t.Error(v)
	}
}

// TestNumeralBreadthSetEqualsTheDeclaredUnion is AC-RNA-006(b).
//
// Both terms of the right-hand side are REGISTRY-derived and therefore
// independent of the layer: a layer-produced term (the neutralised paths, for
// instance) must not appear, because an over-eager neutraliser would then
// shrink both sides together and the equality would hide the failure it exists
// to expose. Comparing the printed count against the layer's own count is
// circular, and a mutant printing only the first hit passes it.
//
// The subtracted term is the declared NumeralUnreachable set — the converse
// declaration, in the shape Site.SweepUnreachable already uses: a registered
// count row whose claim the layer is EXPECTED not to reach, each carrying its
// own reason. Without it the equality is unsatisfiable rather than strict (see
// the field's doc comment).
func TestNumeralBreadthSetEqualsTheDeclaredUnion(t *testing.T) {
	hits := scanLive(t)

	breadth := map[string]bool{}
	for _, h := range hits {
		breadth[h.Path] = true
	}

	unreachable := map[string]string{}
	for _, s := range Registry() {
		if s.NumeralUnreachable != "" {
			unreachable[s.Path] = s.NumeralUnreachable
		}
	}

	want := map[string]bool{}
	for _, s := range Registry() {
		if s.Claims.Has(ClaimCount) && unreachable[s.Path] == "" {
			want[s.Path] = true
		}
	}
	for _, e := range NumeralExemptions() {
		want[e.Path] = true
	}

	for p := range want {
		if !breadth[p] {
			t.Errorf("declared path %s is not in the breadth set — the row or exemption was written against text the layer no longer reaches; re-measure it, or declare NumeralUnreachable with a reason", p)
		}
	}
	for p := range breadth {
		if !want[p] {
			t.Errorf("breadth-set path %s is neither a registered ClaimCount row nor a numeral exemption", p)
		}
	}
	// The converse of the unreachable declaration: a path declared out of the
	// layer's reach that IS reached has changed shape.
	for p, why := range unreachable {
		if breadth[p] {
			t.Errorf("registered site %s declares NumeralUnreachable (%s) but the layer DID reach it — drop the declaration", p, why)
		}
	}

	t.Logf("breadth set: %d paths, %d hits", len(breadth), len(hits))
}

// TestNumeralExemptionsAreWellFormed keeps the exemption list a set of
// declarations a reviewer can disagree with rather than an inference.
func TestNumeralExemptionsAreWellFormed(t *testing.T) {
	root := repoRoot(t)
	seen := map[string]bool{}
	for _, e := range NumeralExemptions() {
		if e.ID == "" {
			t.Errorf("numeral exemption for %s has no ID", e.Path)
		}
		if seen[e.ID] {
			t.Errorf("duplicate numeral-exemption ID %q", e.ID)
		}
		seen[e.ID] = true
		if strings.TrimSpace(e.Reason) == "" {
			t.Errorf("numeral exemption %s carries no reason — an unreasoned exemption is a mute", e.ID)
		}
		if _, err := os.Stat(filepath.Join(root, e.Path)); err != nil {
			t.Errorf("numeral exemption %s: path %s does not exist: %v", e.ID, e.Path, err)
		}
	}
	for _, s := range Registry() {
		if s.NumeralUnreachable != "" && !s.Claims.Has(ClaimCount) {
			t.Errorf("site %s declares NumeralUnreachable but carries no ClaimCount — the declaration has nothing to except", s.ID)
		}
	}
}

// TestHistoricalCitationsAreNotFindings is AC-RNA-014: the historical-citation
// sites produce no finding and are not repaired. Whichever route run-phase
// took — a neutralising mechanism or a per-path exempt declaration — the
// OUTCOME is asserted here.
func TestHistoricalCitationsAreNotFindings(t *testing.T) {
	root := repoRoot(t)
	hits := scanLive(t)
	findings := strings.Join(DischargeNumeralHits(hits, Registry(), NumeralExemptions()), "\n")

	historical := []string{
		".claude/agents/moai/manager-docs.md",
		".claude/agents/moai/manager-spec.md",
		"internal/template/templates/.claude/agents/moai/manager-docs.md",
		"internal/template/templates/.claude/agents/moai/manager-spec.md",
	}
	for _, rel := range historical {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("historical-citation site %s no longer exists: %v (re-measure this list)", rel, err)
			continue
		}
		if strings.Contains(findings, rel) {
			t.Errorf("historical citation %s entered the finding set; it describes the roster AS IT WAS and is not drift", rel)
		}
	}
}

// ── Control probe ────────────────────────────────────────────────────────────

// TestNumeralGuardFiresOnDeliberatelyWrongInput is the control probe.
//
// A passing layer with no control probe is indistinguishable from a dead one:
// the live test above reports exactly the same green output if ScanNumeralLine
// returned nil unconditionally, if the noun class matched nothing, or if the
// neutraliser blanked every line. Each case below feeds input that MUST produce
// a violation.
func TestNumeralGuardFiresOnDeliberatelyWrongInput(t *testing.T) {
	cases := []struct {
		name    string
		hits    []NumeralHit
		sites   []Site
		exempts []NumeralExempt
		want    string
	}{
		{
			name: "an undeclared count claim is reported",
			hits: []NumeralHit{{Path: "probe.md", Line: 1, Numeral: "11", Phrase: "11 retained agents", Axis: "digit"}},
			want: "undeclared count claim",
		},
		{
			name:  "a membership-only registration does not silence a count claim",
			hits:  []NumeralHit{{Path: "probe.md", Line: 1, Numeral: "11", Phrase: "11 retained agents", Axis: "digit"}},
			sites: []Site{{ID: "p", Path: "probe.md", Claims: ClaimMembership}},
			want:  "does not discharge a count claim",
		},
		{
			name:    "an exemption with a blank reason does not suppress",
			hits:    []NumeralHit{{Path: "probe.md", Line: 1, Numeral: "11", Phrase: "11 retained agents", Axis: "digit"}},
			exempts: []NumeralExempt{{ID: "p", Path: "probe.md", Reason: "\t "}},
			want:    "declaration is incomplete",
		},
		{
			name: "a word-axis hit is reported like a digit-axis one",
			hits: []NumeralHit{{Path: "probe.md", Line: 1, Numeral: "eleven", Phrase: "eleven retained agents", Axis: "word"}},
			want: "eleven",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.Join(DischargeNumeralHits(tc.hits, tc.sites, tc.exempts), " | ")
			if !strings.Contains(got, tc.want) {
				t.Fatalf("want a finding containing %q, got: %s", tc.want, got)
			}
		})
	}

	// The scanner half of the probe: a synthetic tree carrying a count-only
	// claim that names ZERO agents — the shape no enumeration threshold reaches.
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "sized.md"), []byte("a 33-cell matrix: 11 retained agents x 3 model tiers\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".moai", "research"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".moai", "research", "dated.md"), []byte("the 11 retained agents\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ScanNumeralAxis(tmp)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, h := range got {
		paths = append(paths, h.Path)
	}
	sort.Strings(paths)
	if len(paths) != 1 || paths[0] != "sized.md" {
		t.Fatalf("scan = %v, want exactly [sized.md] (a count-only claim naming zero agents must be reached, and .moai/research/ must stay excluded)", paths)
	}
	if got[0].Numeral != "11" {
		t.Errorf("selected numeral = %q, want 11 (nearest preceding — 33 is the cell count, not the roster size)", got[0].Numeral)
	}
}
