package spec

import (
	"strings"
	"testing"
)

// drift_close_body_test.go — SPEC-DRIFT-CLOSE-BODY-001 (card t410).
//
// The defect: a close commit whose SUBJECT cannot carry the target SPEC-ID
// (`chore(SPEC group C): Mx-phase close (...)`) is dropped by the stage-2 subject
// re-filter of inMemImpliedStatus, so the walker keeps descending and adopts an
// older `docs(...)` commit as `in-progress`. The SPEC is genuinely closed, and the
// drift table reports a false positive.
//
// The fixtures below are BUILT FROM MEASURED COMMIT BODIES, not invented. Every
// line in the 12-line predicate table is quoted verbatim from the commit named in
// its `source` field (spec.md §5.3). They are the predicate's judgment basis.
//
// The tests drive detectDrift through the driftDeps seam (fakeGit, drift_seam_test.go)
// so the commit index is a commitRecord fixture rather than real git history: the
// body-line predicate is what is under test, not git's log format.

// closeBodySpecFixture writes a V3R6-era SPEC whose frontmatter carries status.
func closeBodySpecFixture(t *testing.T, root, specID, status string) {
	t.Helper()
	writeSpecFixture(t, root, specID, status, "2026-05-01", progressV3R6)
}

// runCloseBodyDrift runs the drift computation over an injected commit index.
func runCloseBodyDrift(t *testing.T, root string, commits []commitRecord) *DriftReport {
	t.Helper()
	fg := &fakeGit{head: "head-close-body", commits: commits}
	report, err := detectDrift(root, fg.deps(false)) // --no-cache path
	if err != nil {
		t.Fatalf("detectDrift: %v", err)
	}
	return report
}

// ---------------------------------------------------------------------------
// AC-DCB-001 — a body-declared close is recognized.
// ---------------------------------------------------------------------------

// alphaCloseSubject reproduces the measured close subject of e979a4d13: an
// arbitrary combined scope that names no SPEC-ID token at all, so
// commitMatchesSPECID drops it in stage 2.
const alphaCloseSubject = "chore(SPEC group C): Mx-phase close (status implemented→completed)"

// alphaCommits is the AC-DCB-001 fixture: a body-declaring close commit above an
// older docs commit for the same SPEC.
func alphaCommits() []commitRecord {
	closeBody := alphaCloseSubject + "\n\n" +
		"- SPEC-FIX-ALPHA-001: Mx verdict EVALUATE-PASS\n"
	olderSubject := "docs(SPEC-FIX-ALPHA-001): /moai mx Step C"

	return []commitRecord{ // newest-first, as git emits
		{subject: alphaCloseSubject, fullMsg: closeBody},
		{subject: olderSubject, fullMsg: olderSubject},
	}
}

// TestDriftCloseBody_PrimaryWalkYieldsInProgress is the vacuity guard AC-DCB-001
// requires: without it, a fixture that never reproduced the defect would still go
// green once the fallback landed.
func TestDriftCloseBody_PrimaryWalkYieldsInProgress(t *testing.T) {
	got, err := inMemImpliedStatus(alphaCommits(), "SPEC-FIX-ALPHA-001")
	if err != nil {
		t.Fatalf("inMemImpliedStatus: %v", err)
	}
	if got != "in-progress" {
		t.Fatalf("primary walk = %q, want %q — the fixture must reproduce the defect before the fallback can be said to repair it", got, "in-progress")
	}
}

// TestDriftCloseBody_BodyDeclaredCloseRecognized is AC-DCB-001.
func TestDriftCloseBody_BodyDeclaredCloseRecognized(t *testing.T) {
	root := t.TempDir()
	closeBodySpecFixture(t, root, "SPEC-FIX-ALPHA-001", "completed")

	report := runCloseBodyDrift(t, root, alphaCommits())

	rec, ok := findRecord(report, "SPEC-FIX-ALPHA-001")
	if !ok {
		t.Fatalf("record for SPEC-FIX-ALPHA-001 missing")
	}
	if rec.GitImpliedStatus != "completed" {
		t.Errorf("GitImpliedStatus = %q, want %q — the body line `- SPEC-FIX-ALPHA-001: Mx verdict ...` declares the close", rec.GitImpliedStatus, "completed")
	}
	if rec.Drifted {
		t.Errorf("Drifted = true, want false — a body-declared close must clear the false positive")
	}
}

// ---------------------------------------------------------------------------
// AC-DCB-002 — a mention is not a close.
// ---------------------------------------------------------------------------

// gammaCommits reproduces a83934d55: a close-infix-bearing subject for one SPEC
// whose body MENTIONS a different SPEC as a dependency.
func gammaCommits() []commitRecord {
	subject := "feat(SPEC-HIER-BETA-001): dependency-aware wiring (Tier M, 3-phase close)"
	body := subject + "\n\n" +
		"depends_on: SPEC-DEP-GAMMA-001 (completed). related: SPEC-GOAL-HTML-WIRING-001,\n"
	olderSubject := "docs(SPEC-DEP-GAMMA-001): M1 notes"

	return []commitRecord{
		{subject: subject, fullMsg: body},
		{subject: olderSubject, fullMsg: olderSubject},
	}
}

// TestDriftCloseBody_GammaPrimaryWalkPrecondition is the D7 reachability guard: the
// mention counter-example only proves anything if it actually reaches the fallback.
func TestDriftCloseBody_GammaPrimaryWalkPrecondition(t *testing.T) {
	got, err := inMemImpliedStatus(gammaCommits(), "SPEC-DEP-GAMMA-001")
	if err != nil {
		t.Fatalf("inMemImpliedStatus: %v", err)
	}
	if got == "completed" || isTerminalStatus(got) {
		t.Fatalf("primary walk = %q — the fixture must be non-completed and non-terminal, or the FALLBACK-ONLY gate never opens and no mutant can be killed", got)
	}
}

// TestDriftCloseBody_MentionIsNotClose is AC-DCB-002.
func TestDriftCloseBody_MentionIsNotClose(t *testing.T) {
	root := t.TempDir()
	closeBodySpecFixture(t, root, "SPEC-DEP-GAMMA-001", "completed")

	report := runCloseBodyDrift(t, root, gammaCommits())

	rec, ok := findRecord(report, "SPEC-DEP-GAMMA-001")
	if !ok {
		t.Fatalf("record for SPEC-DEP-GAMMA-001 missing")
	}
	if rec.GitImpliedStatus == "completed" {
		t.Errorf("GitImpliedStatus = %q, want NOT completed — `depends_on: SPEC-DEP-GAMMA-001 (completed)` is a mention, not a close declaration (REQ-DCB-004)", rec.GitImpliedStatus)
	}
	if !rec.Drifted {
		t.Errorf("Drifted = false, want true — a dependency mention must not clear drift")
	}
}

// ---------------------------------------------------------------------------
// The 12-line predicate table (spec.md §5.3). This IS the predicate's judgment
// basis: 12/12 or the predicate is wrong.
// ---------------------------------------------------------------------------

// closeBearingSubject is the subject the line-level table tests run under. Shape A
// is gated on the containing commit's subject carrying a close signal, so the table
// — which exercises the LINE predicate — supplies a subject that clears that gate.
// The gate itself is measured separately, against the commit that motivated it.
const closeBearingSubject = alphaCloseSubject

type bodyLineCase struct {
	n      int
	source string
	specID string
	line   string
	want   bool
	why    string
}

// closeBodyFixtureLines are quoted verbatim from the commits named in `source`.
var closeBodyFixtureLines = []bodyLineCase{
	{
		n: 1, source: "e979a4d13", specID: "SPEC-V3R6-SESSION-HANDOFF-AUTO-001",
		line: "- SPEC-V3R6-SESSION-HANDOFF-AUTO-001: Mx verdict EVALUATE-PASS, 1 @MX:TODO deferred",
		want: true, why: "shape A — list marker stripped, line starts with the full ID followed by ':'",
	},
	{
		n: 2, source: "e979a4d13", specID: "SPEC-V3R6-PROMPT-CACHE-001",
		line: "- SPEC-V3R6-PROMPT-CACHE-001: Mx verdict EVALUATE-PASS, 10 @MX:ANCHOR tags",
		want: true, why: "shape A",
	},
	{
		n: 3, source: "2f449e189", specID: "SPEC-GLM-KEY-INPUT-001",
		line: "* docs(SPEC-GLM-KEY-INPUT-001): sync-phase artifacts — 3-phase close",
		want: true, why: "shape B — conventional-commit subject; the chain returns completed",
	},
	{
		n: 4, source: "7beda68a5", specID: "SPEC-V3R6-CODERABBIT-ADOPTION-001",
		line: "* chore(SPEC-V3R6-CODERABBIT-ADOPTION-001): sync-phase artifacts — 3-phase close (4be491a0b)",
		want: true, why: "shape B — close-infix beats the chore prefix rule",
	},
	{
		n: 5, source: "a83934d55", specID: "SPEC-AUTONOMY-TIERS-001",
		line: "depends_on: SPEC-AUTONOMY-TIERS-001 (completed). related: SPEC-GOAL-HTML-WIRING-001,",
		want: false, why: "mention — dependency key (REQ-DCB-004); measured TIGHT false positive",
	},
	{
		n: 6, source: "fb8aff006", specID: "SPEC-WORKTREE-BRANCH-GUARD-001",
		line: "Carries the merged in-progress → implemented → completed status transition on spec.md frontmatter (sole YAML-frontmatter artifact), the progress.md §E.4 Sync-phase Audit-Ready Signal block (sync_commit_sha=61e2c3bc0, sync_status=audit-ready), and the CHANGELOG [Unreleased] entry. Zero code changes — sync is markdown-only. AC-WBG-D-001..008 (8/8 MUST) all PASS against run-phase PR #1488 squash 61e2c3bc0 on origin/main. Depends on SPEC-WORKTREE-BRANCH-GUARD-001 and SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 (both completed). Route B (PR-mandatory, enforce_admins branch-protection override).",
		want: false, why: "mention — prose, not a conventional-commit subject; measured TIGHT false positive",
	},
	{
		n: 7, source: "80dea9684", specID: "SPEC-INTERNAL-TEST-002",
		line: "spec/plan/acceptance/progress frontmatter status in-progress -> completed (merged 3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-008/009). progress.md §E.4 audit-ready signal block (sync_status: PASS-WITH-DEBT, 3 residual_debt items owned by follow-up SPEC-INTERNAL-TEST-002). CHANGELOG [Unreleased] > Fixed entry.",
		want: false, why: "mention — names a FOLLOW-UP SPEC. Prose carrying the close-infix: the shape-B structural precondition is what rejects it",
	},
	{
		n: 8, source: "51d18d3fe", specID: "SPEC-INTERNAL-SECURITY-001",
		line: "SPEC-INTERNAL-SECURITY-001 f3193bac8 / SPEC-HANDOFF-GOALFIX-001",
		want: false, why: "mention — a 72-column wrap artifact; starts with the full ID but no ':' follows",
	},
	{
		n: 9, source: "2f449e189", specID: "SPEC-GLM-KEY-INPUT-001",
		line: "- CHANGELOG [Unreleased]: Added entry for SPEC-GLM-KEY-INPUT-001",
		want: false, why: "mention — list marker followed by a different key",
	},
	{
		n: 10, source: "2f449e189", specID: "SPEC-GLM-KEY-INPUT-001",
		line: "- SPEC-GLM-KEY-INPUT-001         → d06771f07",
		want: false, why: "mention — no ':' after the ID",
	},
	{
		n: 11, source: "7beda68a5", specID: "SPEC-WORKTREE-ENTRY-STRATEGY-001",
		line: "* fix(SPEC-WORKTREE-ENTRY-STRATEGY-001): M1 web auto-toggles default OFF (AutoCleanup+AutoMerge true→false)",
		want: false, why: "shape B but the chain returns implemented, not completed",
	},
	{
		n: 12, source: "7beda68a5", specID: "SPEC-WORKTREE-ENTRY-STRATEGY-001",
		line: "* docs(SPEC-WORKTREE-ENTRY-STRATEGY-001): sync-phase artifacts — 3-phase close (CHANGELOG + completed transition)",
		want: true, why: "shape B — the real close, 116 lines below line 11 in the source body",
	},
}

// TestDriftCloseBody_PredicateTable asserts the predicate against all 12 measured
// lines. Fewer lines would leave a trap uncovered: wrap artifact (8), missing colon
// (10), a different key behind the marker (9), close-infix-bearing prose (7), and a
// non-close shape-B line (11).
func TestDriftCloseBody_PredicateTable(t *testing.T) {
	for _, tc := range closeBodyFixtureLines {
		got := bodyDeclaresClose(closeBearingSubject, tc.line, tc.specID)
		if got != tc.want {
			t.Errorf("line %d (%s, %s): bodyDeclaresClose = %v, want %v — %s\n  line: %s",
				tc.n, tc.source, tc.specID, got, tc.want, tc.why, tc.line)
		}
	}
}

// TestDriftCloseBody_FullBodyScan is the ordered 11·12 pair: both name the same
// SPEC-ID as shape B, 11 is non-close and comes first. An implementation that
// returns at the first shape-B line stops at 11 and never sees 12.
func TestDriftCloseBody_FullBodyScan(t *testing.T) {
	var line11, line12 string
	for _, tc := range closeBodyFixtureLines {
		switch tc.n {
		case 11:
			line11 = tc.line
		case 12:
			line12 = tc.line
		}
	}
	body := "chore(release): squash merge\n\n" + line11 + "\n" + line12 + "\n"

	if !bodyDeclaresClose(closeBearingSubject, body, "SPEC-WORKTREE-ENTRY-STRATEGY-001") {
		t.Errorf("bodyDeclaresClose = false, want true — the scan must sweep the whole body; line 11 is non-close and precedes the real close on line 12 (REQ-DCB-003)")
	}
}

// ---------------------------------------------------------------------------
// AC-DCB-003 — the FALLBACK-ONLY input gate and the completed-or-nothing output.
// ---------------------------------------------------------------------------

// TestDriftCloseBody_FallbackOnlyInputGate is AC-DCB-003 (a)(b)(c): with the same
// body-declared close evidence present, a frontmatter status other than `completed`
// must never be overwritten.
func TestDriftCloseBody_FallbackOnlyInputGate(t *testing.T) {
	cases := []struct {
		name        string
		frontmatter string
		wantGit     string
	}{
		{"(a) in-progress", "in-progress", "in-progress"},
		{"(b) superseded", "superseded", "terminal-exempt"},
		{"(c) draft", "draft", "in-progress"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			closeBodySpecFixture(t, root, "SPEC-FIX-ALPHA-001", tc.frontmatter)

			report := runCloseBodyDrift(t, root, alphaCommits())

			rec, ok := findRecord(report, "SPEC-FIX-ALPHA-001")
			if !ok {
				t.Fatalf("record missing")
			}
			if rec.GitImpliedStatus != tc.wantGit {
				t.Errorf("GitImpliedStatus = %q, want %q — the body lookup fires only while frontmatter is completed (REQ-DCB-002)", rec.GitImpliedStatus, tc.wantGit)
			}
		})
	}
}

// deltaCommits carries ONLY the non-close shape-B line 11 in its body — the 11·12
// pair with 12 removed. The chain classifies that line as `implemented`; the output
// gate must discard it.
func deltaCommits() []commitRecord {
	subject := "chore(release): squash merge of the entry-strategy branch"
	body := subject + "\n\n" +
		"* fix(SPEC-ENTRY-DELTA-001): M1 web auto-toggles default OFF (AutoCleanup+AutoMerge true→false)\n"
	olderSubject := "docs(SPEC-ENTRY-DELTA-001): M2 doc alignment"

	return []commitRecord{
		{subject: subject, fullMsg: body},
		{subject: olderSubject, fullMsg: olderSubject},
	}
}

// TestDriftCloseBody_DeltaPrimaryWalkPrecondition is the (d) vacuity guard: if the
// primary walk already said `completed`, case (d) would distinguish nothing.
func TestDriftCloseBody_DeltaPrimaryWalkPrecondition(t *testing.T) {
	got, err := inMemImpliedStatus(deltaCommits(), "SPEC-ENTRY-DELTA-001")
	if err != nil {
		t.Fatalf("inMemImpliedStatus: %v", err)
	}
	if got == "completed" {
		t.Fatalf("primary walk = %q — case (d) asserts nothing unless the primary walk is non-completed", got)
	}
	if got != "in-progress" {
		t.Fatalf("primary walk = %q, want %q", got, "in-progress")
	}
}

// TestDriftCloseBody_OutputIsCompletedOrNothing is AC-DCB-003 (d) / REQ-DCB-005.
// §5.1's structural argument for this property died when shape B started feeding
// body lines to ClassifyPRTitle, which returns implemented / in-progress / draft.
// The guarantee is now behavioral, so it is measured rather than argued.
func TestDriftCloseBody_OutputIsCompletedOrNothing(t *testing.T) {
	root := t.TempDir()
	closeBodySpecFixture(t, root, "SPEC-ENTRY-DELTA-001", "completed")

	report := runCloseBodyDrift(t, root, deltaCommits())

	rec, ok := findRecord(report, "SPEC-ENTRY-DELTA-001")
	if !ok {
		t.Fatalf("record missing")
	}
	if rec.GitImpliedStatus != "in-progress" {
		t.Errorf("GitImpliedStatus = %q, want %q — a non-close shape-B body line must leave the primary walk's value untouched; the body lookup emits completed or nothing (REQ-DCB-005)", rec.GitImpliedStatus, "in-progress")
	}
}

// ---------------------------------------------------------------------------
// Candidate-window gate (a): a subject that names the SPEC-ID belongs to the
// primary walk (REQ-DCB-006 / §5.4 condition 2).
// ---------------------------------------------------------------------------

// TestDriftCloseBody_SubjectNamingSpecIDIsPrimaryWalkTerritory asserts the body
// lookup declines a commit whose subject already carries the full ID, so the new
// axis cannot re-decide what the primary walk decided.
func TestDriftCloseBody_SubjectNamingSpecIDIsPrimaryWalkTerritory(t *testing.T) {
	subject := "docs(SPEC-SUBJ-OWNED-001): M2 doc alignment"
	body := subject + "\n\n" +
		"* docs(SPEC-SUBJ-OWNED-001): sync-phase artifacts — 3-phase close\n"

	commits := []commitRecord{{subject: subject, fullMsg: body}}

	if inMemBodyDeclaredClose(commits, "SPEC-SUBJ-OWNED-001") {
		t.Errorf("inMemBodyDeclaredClose = true, want false — the subject names the full ID, so this commit is the primary walk's to classify (§5.4 condition 2)")
	}
}

// ---------------------------------------------------------------------------
// The shape-A subject gate — found by the M3 corpus sweep, not by the 12-line
// fixture. spec.md §5.3 anticipates exactly this: a counter-example outside the
// fixture is answered by NARROWING the predicate, never by widening it.
// ---------------------------------------------------------------------------

// amendmentCommit reproduces ac3e38a0b: a shape-A-looking line that records an
// AMENDMENT, inside a commit that closes nothing. Its own body states that the
// status was preserved rather than transitioned.
func amendmentCommit() commitRecord {
	subject := "feat(SPEC-MCP-DEFAULT-ON-001): moai MCP server as first-class default (plan+amendment+run)"
	body := subject + "\n\n" +
		"Both completed SPECs amended per SPEC-MCP-DEFAULT-ON-001 REQ-A-6/REQ-A-7:\n" +
		"- SPEC-AMEND-TARGET-001: REQ-MCP-002 opt-in->default-on, REQ-MCP-015 opt-out flag, AC-MCP-002/006 amended (0.1.0 -> 0.2.0)\n\n" +
		"status: completed preserved (owner direction); amendment_of omitted.\n"
	return commitRecord{subject: subject, fullMsg: body}
}

// TestDriftCloseBody_AmendmentRecordIsNotClose is the measured counter-example.
// Without the subject gate the amendment line reads as a close and clears a row
// whose close is not on the judged branch at all.
func TestDriftCloseBody_AmendmentRecordIsNotClose(t *testing.T) {
	c := amendmentCommit()

	if bodyDeclaresClose(c.subject, c.fullMsg, "SPEC-AMEND-TARGET-001") {
		t.Errorf("bodyDeclaresClose = true, want false — the line records an amendment (REQ/AC edits, 0.1.0 -> 0.2.0) inside a commit that closes nothing; shape A must be gated on the subject carrying a close signal")
	}

	// Converse: the same line inside a commit that IS a close still counts, so the
	// gate narrows shape A rather than disabling it.
	closeSubject := "docs: close out 4 A-tier SPECs with sync-phase"
	if !bodyDeclaresClose(closeSubject, c.fullMsg, "SPEC-AMEND-TARGET-001") {
		t.Errorf("bodyDeclaresClose = false, want true — the gate must narrow shape A, not disable it")
	}
}

// TestDriftCloseBody_SubjectCloseSignalCoversMeasuredCloses discharges spec.md
// §5.4 condition 3: a subject-side signal used as a gate must admit every measured
// close subject. closeInfixMatch failed this (it drops 2 of 6, this card's own
// target among them), which is why the gate is a broader `close` substring.
func TestDriftCloseBody_SubjectCloseSignalCoversMeasuredCloses(t *testing.T) {
	measured := []string{
		"Close out 2 SPECs with 3-phase lifecycle completion (doc-only) (#1210)",                                                                            // 7beda68a5
		"chore(SPEC group C): Mx-phase close (status implemented→completed, 2026-06-02)",                                                                    // e979a4d13
		"docs(specs): batch sync-phase close — 5 B-grade SPECs (3-phase close) (#1240)",                                                                     // 2f449e189
		"docs: close out 4 A-tier SPECs with sync-phase (CLI-TUI-MODERNIZE-001 · INVOCATION-MODEL-002 · WORKFLOW-CACHE-OPT-001 · ASTGREP-EDIT-001) (#1215)", // cd21df594
		"chore(spec): close KANBAN-RENAME-001 + AGENT-MODEL-ENFORCE-001, supersede CONFIG-TIER-PERSIST-001 (#1516)",                                         // cd80f0644
		"docs(SPEC-INTERNAL-TEST-001): sync-phase artifacts + 3-phase close",
		"feat(SPEC-HIERARCHICAL-TEAM-001): hierarchical team wiring (Tier M, 3-phase close) (#1394)",
	}
	for _, s := range measured {
		if !subjectCloseSignal(s) {
			t.Errorf("subjectCloseSignal = false, want true — §5.4 condition 3 requires every measured close subject to pass\n  subject: %s", s)
		}
	}

	// Two of these are exactly the subjects closeInfixMatch drops (§5.4), which is
	// the reason the gate is not closeInfixMatch.
	for _, s := range []string{measured[0], measured[1]} {
		if closeInfixMatch(strings.ToLower(s)) {
			t.Errorf("closeInfixMatch = true, want false — the §5.4 rejection measurement no longer reproduces; re-derive the gate\n  subject: %s", s)
		}
	}
}

// TestDriftCloseBody_ExistingFallbackStillFirst pins REQ-DCB-006: the combined-scope
// fallback keeps deciding the cases it already decided.
func TestDriftCloseBody_ExistingFallbackStillFirst(t *testing.T) {
	root := t.TempDir()
	closeBodySpecFixture(t, root, "SPEC-ABC-FOO-001", "completed")

	closeSubject := "chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)"
	commits := []commitRecord{
		{subject: closeSubject, fullMsg: closeSubject},
		{subject: "feat(SPEC-ABC-FOO-001): M1 implementation", fullMsg: "feat(SPEC-ABC-FOO-001): M1 implementation"},
	}

	report := runCloseBodyDrift(t, root, commits)

	rec, ok := findRecord(report, "SPEC-ABC-FOO-001")
	if !ok {
		t.Fatalf("record missing")
	}
	if rec.Drifted {
		t.Errorf("Drifted = true, want false — the combined-scope fallback must still resolve this (REQ-DCB-006). GitImpliedStatus=%q", rec.GitImpliedStatus)
	}
}

// TestDriftCloseBody_ListMarkersStripped pins the marker set the shape predicates
// tolerate, so a future reader does not have to infer it from the fixture lines.
func TestDriftCloseBody_ListMarkersStripped(t *testing.T) {
	const id = "SPEC-MARKER-CASE-001"
	for _, marker := range []string{"", "- ", "* ", "+ ", "  - ", "\t* "} {
		line := marker + id + ": Mx verdict EVALUATE-PASS"
		if !bodyDeclaresClose(closeBearingSubject, line, id) {
			t.Errorf("marker %q: bodyDeclaresClose = false, want true\n  line: %s", marker, line)
		}
	}
	// A marker that is not a list marker must not be stripped into a shape-A match.
	if bodyDeclaresClose(closeBearingSubject, "> "+id+": quoted", id) {
		t.Errorf("blockquote marker was stripped — the marker set must stay narrow")
	}
	if !strings.Contains("> "+id, id) {
		t.Fatal("fixture sanity")
	}
}
