// autodone_scan_test.go — SPEC-TODO-LAND-AUTO-DONE-001 M1: the guard
// predicates and the scan-decision pure function.
//
// Every criterion here is asserted in BOTH directions where a one-directional
// suite could pass a mutant: the comma form attributes, and a group carrying
// two card tokens or a non-card-leading group still does not; the negation
// marker excludes a subject, and the same tokens on a clean subject still
// attribute; the collision gate skips on subject evidence, and the recorded
// SHA still closes through it.
package kanban

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestLandedPredicate_NegationAttributesNothing — REQ-AD-008 (guard M3): a
// subject carrying an explicit non-landing declaration attributes NOTHING.
// Both tokens, both cases; and the control proves the exclusion is the
// negation and not the parenthetical.
func TestLandedPredicate_NegationAttributesNothing(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "fix(t903): attempt (not merged)"},
		formCommit{subject: "docs(t904): notes (NOT landed)"},
		formCommit{subject: "docs(t905): genuinely delivered"},
	)
	q := GitLandedQuerier{Run: gitIn(dir), Ref: "origin/develop"}

	for _, card := range []string{"t903", "t904"} {
		if got, err := q.Landed(card); err != nil {
			t.Fatalf("Landed(%s): %v", card, err)
		} else if got != LandingNotLanded {
			t.Errorf("%s = %q, want %q — a non-landing declaration attributes nothing", card, got, LandingNotLanded)
		}
	}
	if got, err := q.Landed("t905"); err != nil {
		t.Fatalf("Landed(t905): %v", err)
	} else if got != LandingLanded {
		t.Errorf("t905 = %q, want %q — the control: a clean subject still attributes", got, LandingLanded)
	}
}

// TestLandedPredicate_NegationIsSubjectStreamOnly — the §D.1 edge case: a
// negation marker in a commit BODY with a clean attributing SUBJECT must not
// negate the subject's attribution (the predicate is subject-stream-only).
func TestLandedPredicate_NegationIsSubjectStreamOnly(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{
			subject: "docs(t906): the real delivery",
			body:    "an earlier attempt was not merged; this one was",
		},
	)
	if got, err := (GitLandedQuerier{Run: gitIn(dir), Ref: "origin/develop"}).Landed("t906"); err != nil {
		t.Fatalf("Landed(t906): %v", err)
	} else if got != LandingLanded {
		t.Errorf("t906 = %q, want %q — a body negation is out of scope and must not negate the subject", got, LandingLanded)
	}
}

// TestLandedPredicate_CommaFormTrailingGroup — AC-AD-017 Shape A: a
// comma-form trailing parenthetical attributes to the card the group opens
// with. The negatives pin the boundaries: a group carrying two distinct card
// tokens attributes nothing (§D), and a group whose card token does not OPEN
// the group still attributes nothing (the pinned t80 branch-name control's
// shape).
func TestLandedPredicate_CommaFormTrailingGroup(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "fix(hooks): sync-phase 게이트의 C++ 검사 복구 (t603, H08)"},
		formCommit{subject: "docs: two cards in one group (t811, t812)"},
		formCommit{subject: "docs: card not leading the group (work for t813)"},
	)
	q := GitLandedQuerier{Run: gitIn(dir), Ref: "origin/develop"}

	if got, err := q.Landed("t603"); err != nil {
		t.Fatalf("Landed(t603): %v", err)
	} else if got != LandingLanded {
		t.Errorf("t603 = %q, want %q — the comma-form group attributes to the id it opens with", got, LandingLanded)
	}
	for _, card := range []string{"t811", "t812"} {
		if got, err := q.Landed(card); err != nil {
			t.Fatalf("Landed(%s): %v", card, err)
		} else if got != LandingNotLanded {
			t.Errorf("%s = %q, want %q — a group carrying two distinct card tokens attributes nothing (§D)", card, got, LandingNotLanded)
		}
	}
	if got, err := q.Landed("t813"); err != nil {
		t.Fatalf("Landed(t813): %v", err)
	} else if got != LandingNotLanded {
		t.Errorf("t813 = %q, want %q — a card that does not open the group is not attributed by it", got, LandingNotLanded)
	}
}

// TestAutoDoneDistinctTexts — the M1 collision gate's counter: an id carried
// by more than one DISTINCT text across the live queue and the archive.
func TestAutoDoneDistinctTexts(t *testing.T) {
	rec := &BacklogRecord{}
	seed := func(id, text string, state BacklogState, archived bool) {
		item := BacklogItem{ID: id, Text: text, State: state}
		if archived {
			rec.Archived = append(rec.Archived, BacklogArchiveEntry{Item: item, Position: 0, Findings: []BacklogArchivedFinding{}})
			return
		}
		rec.Items = append(rec.Items, item)
	}

	seed("t902", "original work", BacklogStateQueued, true)
	if got := AutoDoneDistinctTexts(rec, "t902"); got != 1 {
		t.Errorf("archived predecessor alone = %d distinct text, want 1", got)
	}
	seed("t902", "reissued different work", BacklogStateQueued, false)
	if got := AutoDoneDistinctTexts(rec, "t902"); got != 2 {
		t.Errorf("reissued id = %d distinct texts, want 2 — the collision gate's input", got)
	}
	// The same text re-recorded twice is NOT a collision.
	rec2 := &BacklogRecord{}
	seed2 := func(id, text string, archived bool) {
		item := BacklogItem{ID: id, Text: text, State: BacklogStateQueued}
		if archived {
			rec2.Archived = append(rec2.Archived, BacklogArchiveEntry{Item: item, Position: 0, Findings: []BacklogArchivedFinding{}})
			return
		}
		rec2.Items = append(rec2.Items, item)
	}
	seed2("t801", "same text", true)
	seed2("t801", "same text", false)
	if got := AutoDoneDistinctTexts(rec2, "t801"); got != 1 {
		t.Errorf("identical text in archive and live = %d, want 1 — same text is not a reissue", got)
	}
	if got := AutoDoneDistinctTexts(rec, "t777"); got != 0 {
		t.Errorf("absent id = %d, want 0", got)
	}
}

// scanRepo builds a repository whose origin/develop history is exactly the
// given subjects (oldest first) and returns the dir.
func scanRepo(t *testing.T, subjects ...string) string {
	t.Helper()
	entries := make([]formCommit, 0, len(subjects)+1)
	entries = append(entries, formCommit{subject: "chore: seed the tree"})
	for _, s := range subjects {
		entries = append(entries, formCommit{subject: s})
	}
	return formRepo(t, "develop", entries...)
}

// TestScanLandedSubjects_LandedAttributions — the one-query subject stream
// the scan evaluates: malformed lines are reported, not silently dropped, and
// the attribution map carries the FIRST (newest) attributing commit per id.
func TestScanLandedSubjects_LandedAttributions(t *testing.T) {
	dir := scanRepo(t,
		"Merge WT-fixture into develop (card t901)",
		"fix(t901): earlier step",
	)
	commits, err := ScanLandedSubjects(gitIn(dir), "origin/develop")
	if err != nil {
		t.Fatalf("ScanLandedSubjects: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("commits = %d, want 3 (seed + two)", len(commits))
	}
	// Newest first — git log order.
	if commits[0].Subject != "fix(t901): earlier step" {
		t.Errorf("commits[0].Subject = %q, want the newest subject", commits[0].Subject)
	}
	if commits[0].SHA == "" || len(commits[0].SHA) != 40 {
		t.Errorf("commits[0].SHA = %q, want a full SHA", commits[0].SHA)
	}

	attr := LandedAttributions(commits, "develop")
	hit, ok := attr["t901"]
	if !ok {
		t.Fatalf("t901 missing from attributions %v", attr)
	}
	if hit.Subject != "fix(t901): earlier step" {
		t.Errorf("t901 attributed to %q, want the FIRST (newest) attributing commit", hit.Subject)
	}
	if _, ok := attr["t777"]; ok {
		t.Error("t777 attributed by no subject but present in the map")
	}

	// An unanswerable ref is an error, never an empty result that reads as
	// "nothing landed".
	empty := t.TempDir()
	cmd := exec.Command("git", "-C", empty, "init", "-q")
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	if err := cmd.Run(); err != nil {
		t.Fatalf("init empty fixture: %v", err)
	}
	if _, err := ScanLandedSubjects(gitIn(empty), "origin/develop"); err == nil {
		t.Error("ScanLandedSubjects on an unresolvable ref returned no error — the inconclusive answer collapsed into not-landed")
	}
}

// TestScanLandedSubjects_CarriesCommitTime — the t684 generation boundary's
// input: the stream's rows carry a parseable unix committer time, and a
// malformed time field is an error, not a zero.
func TestScanLandedSubjects_CarriesCommitTime(t *testing.T) {
	dir := scanRepo(t, "fix(t901): timed landing")
	commits, err := ScanLandedSubjects(gitIn(dir), "origin/develop")
	if err != nil {
		t.Fatalf("ScanLandedSubjects: %v", err)
	}
	if len(commits) == 0 {
		t.Fatal("empty stream")
	}
	for _, c := range commits {
		if c.CommitTime <= 0 {
			t.Errorf("commit %s carries CommitTime %d, want a positive unix timestamp", c.SHA, c.CommitTime)
		}
	}
	// The newest commit's time is within a minute of now — a real parse of
	// %ct, not a zero from a missing field.
	newest := time.Now().Unix()
	if d := newest - commits[0].CommitTime; d < 0 || d > 60 {
		t.Errorf("newest CommitTime %d is %ds from now, want within a minute", commits[0].CommitTime, d)
	}
}

// TestAutoDoneSubjectFresh — the t684 generation boundary in both directions:
// a commit older than the card is stale (the old generation's landing must
// not close the reissued card), a commit at or after creation is fresh, and
// an undecidable card time fails CLOSED.
func TestAutoDoneSubjectFresh(t *testing.T) {
	cardAt := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	addedAt := cardAt.Format(time.RFC3339)
	hit := LandedCommit{SHA: "abc", Subject: "fix(t1): x", CommitTime: cardAt.Unix()}

	if AutoDoneSubjectFresh(LandedCommit{SHA: "abc", CommitTime: cardAt.Add(-24 * time.Hour).Unix()}, addedAt) {
		t.Error("a day-older commit reads fresh — the old generation would close the reissued card")
	}
	if !AutoDoneSubjectFresh(hit, addedAt) {
		t.Error("a same-second commit reads stale — the boundary is generation separation, not second-level forensics")
	}
	if !AutoDoneSubjectFresh(LandedCommit{SHA: "abc", CommitTime: cardAt.Add(time.Hour).Unix()}, addedAt) {
		t.Error("a commit after card creation reads stale — the boundary over-blocks real landings")
	}
	if AutoDoneSubjectFresh(hit, "not-a-timestamp") {
		t.Error("an unparseable added_at fails OPEN — no subject close without a decidable boundary")
	}
	if AutoDoneSubjectFresh(hit, "") {
		t.Error("an empty added_at fails OPEN — no subject close without a decidable boundary")
	}
}

// shaReachableFacts builds facts with a recorded SHA and a given reachability.
func shaReachableFacts(sha string, reachable AutoDoneTri) AutoDoneFacts {
	return AutoDoneFacts{RecordedSHA: sha, SHAReachable: reachable, SubjectKnown: true, DistinctTexts: 1, SpecSyncGate: AutoDoneYes}
}

// TestAutoDoneDecide — the whole scan policy, table-driven. Every skip token
// and both close forms are exercised, including the precedence pairs the ACs
// pin: the sync gate outranks evidence (AC-AD-006), the recorded SHA closes
// through a collision (AC-AD-005), and the collision gate blocks subject
// evidence (AC-AD-004).
func TestAutoDoneDecide(t *testing.T) {
	hit := &LandedCommit{SHA: "abc123", Subject: "fix(t901): work (t901)"}
	tests := []struct {
		name       string
		facts      AutoDoneFacts
		wantClose  bool
		wantForm   string
		wantReason string
	}{
		{"form 1 closes", shaReachableFacts("abc123", AutoDoneYes), true, AutoDoneFormSHA, ""},
		{"form 2 closes", AutoDoneFacts{SubjectKnown: true, SubjectHit: hit, DistinctTexts: 1, SpecSyncGate: AutoDoneYes}, true, AutoDoneFormSubject, ""},
		{"recorded SHA unreachable falls to form 2", AutoDoneFacts{RecordedSHA: "abc123", SHAReachable: AutoDoneNo, SubjectKnown: true, SubjectHit: hit, DistinctTexts: 1, SpecSyncGate: AutoDoneYes}, true, AutoDoneFormSubject, ""},
		{"recorded SHA unreachable and no subject is not-landed", AutoDoneFacts{RecordedSHA: "abc123", SHAReachable: AutoDoneNo, SubjectKnown: true, DistinctTexts: 1, SpecSyncGate: AutoDoneYes}, false, "", AutoDoneSkipNotLanded},
		{"recorded SHA reachability unanswerable is inconclusive", shaReachableFacts("abc123", AutoDoneUnknown), false, "", AutoDoneSkipQueryInconclusive},
		{"subject query unanswerable is inconclusive", AutoDoneFacts{SubjectKnown: false, DistinctTexts: 1, SpecSyncGate: AutoDoneYes}, false, "", AutoDoneSkipQueryInconclusive},
		{"collision on subject evidence alone skips ambiguous", AutoDoneFacts{SubjectKnown: true, SubjectHit: hit, DistinctTexts: 2, SpecSyncGate: AutoDoneYes}, false, "", AutoDoneSkipAmbiguousID},
		{"recorded SHA closes through a collision", shaReachableFacts("abc123", AutoDoneYes), true, AutoDoneFormSHA, ""},
		{"no evidence is not-landed", AutoDoneFacts{SubjectKnown: true, DistinctTexts: 1, SpecSyncGate: AutoDoneYes}, false, "", AutoDoneSkipNotLanded},
		{"sync gate outranks subject evidence", AutoDoneFacts{SubjectKnown: true, SubjectHit: hit, DistinctTexts: 1, SpecSyncGate: AutoDoneNo}, false, "", AutoDoneSkipSpecNotCompleted},
		{"sync gate outranks a recorded SHA", func() AutoDoneFacts {
			f := shaReachableFacts("abc123", AutoDoneYes)
			f.SpecSyncGate = AutoDoneNo
			return f
		}(), false, "", AutoDoneSkipSpecNotCompleted},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AutoDoneDecide(tc.facts)
			if got.Close != tc.wantClose {
				t.Fatalf("Close = %v, want %v (decision %+v)", got.Close, tc.wantClose, got)
			}
			if tc.wantClose && got.Form != tc.wantForm {
				t.Errorf("Form = %q, want %q", got.Form, tc.wantForm)
			}
			if !tc.wantClose && got.Reason != tc.wantReason {
				t.Errorf("Reason = %q, want %q", got.Reason, tc.wantReason)
			}
		})
	}
}

// TestAutoDoneSkipVocabularyClosed — the four-token skip vocabulary is the
// closed set REQ-AD-010 canonically enumerates. A fifth token, a renamed
// token, or a missing one breaks every reader that renders skip reasons.
func TestAutoDoneSkipVocabularyClosed(t *testing.T) {
	want := []string{"ambiguous-id", "spec-not-completed", "not-landed", "query-inconclusive"}
	if got := AutoDoneSkipReasons(); !slicesEqual(got, want) {
		t.Errorf("skip vocabulary = %v, want %v (closed set, canonical order)", got, want)
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestLandedScanArgsShape — the scan query is a subject stream keyed by SHA,
// never a whole-message stream: the format separator keeps SHA and subject
// separable, and the argv names the ref as an input.
func TestLandedScanArgsShape(t *testing.T) {
	args := LandedScanArgs("origin/develop")
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "origin/develop") {
		t.Errorf("argv %v does not name the ref", args)
	}
	if strings.Contains(joined, "%B") {
		t.Errorf("argv %v carries a whole-message format — a %%B stream feeds commit bodies to the matcher as subjects", args)
	}
	if !strings.Contains(joined, "%H") {
		t.Errorf("argv %v carries no SHA field — the log row needs the attributing commit's SHA", args)
	}
	if !strings.Contains(joined, "%ct") {
		t.Errorf("argv %v carries no committer-time field — the t684 generation boundary needs %%%%ct to judge added_at", args)
	}
}
