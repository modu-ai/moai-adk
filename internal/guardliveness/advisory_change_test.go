package guardliveness

import (
	"strings"
	"testing"
	"time"
)

// firstSubjectLine returns the first individually-rendered entry line, which is
// the advisory's leading position. Asserting against the rendered text rather
// than against an internal ordering is what makes clause (a) a statement about
// what a reader sees.
func firstSubjectLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "- ") {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// AC-GDL-007 — the advisory leads with changes and survives a standing
// non-clean neighbour.
//
// Two consecutive evaluations: S = {subject-2, subject-3} is non-clean in both,
// and T = subject-1 newly became non-clean in the second. BOTH clauses are
// asserted, because clause (a) alone is satisfied by a renderer printing the
// full standing list with T at the top — the block a reader learns to skip
// after the third session, and how a new advisory inherits the filter an
// always-red neighbour has already trained (spec.md §A.8).
func TestAdvisoryLeadsWithChangesAndCollapsesStandingEntries(t *testing.T) {
	now := time.Now()
	first := Result{
		Clean: designate("alpha"),
		Entries: []Entry{
			entry("subject-1", "alpha", "settled"),
			entry("subject-2", "beta", "settled"),
			entry("subject-3", "gamma", "unsettled"),
		},
	}
	second := Result{
		Clean: designate("alpha"),
		Entries: []Entry{
			entry("subject-1", "gamma", "unsettled"),
			entry("subject-2", "beta", "settled"),
			entry("subject-3", "gamma", "unsettled"),
		},
	}

	firstText, rec := Advisory(snapshotAt(first, now.Add(-2*time.Hour)), RenderRecord{}, now)
	if rec == nil {
		t.Fatal("the first render recorded nothing, so the second has no previous result to diff against")
	}
	for _, subject := range []string{"subject-2", "subject-3"} {
		if !strings.Contains(firstText, subject) {
			t.Fatalf("the first render omits %q, so it never announced what the second must treat as standing:\n%s", subject, firstText)
		}
	}

	text, next := Advisory(snapshotAt(second, now.Add(-5*time.Minute)), *rec, now)
	if text == "" {
		t.Fatal("the second render was silent while three subjects were not reporting clean")
	}
	if next == nil {
		t.Fatal("the second render recorded nothing")
	}

	// (a) the changed entry occupies the leading position.
	if lead := firstSubjectLine(text); !strings.Contains(lead, "subject-1") {
		t.Errorf("the newly non-clean subject does not lead; leading line is %q:\n%s", lead, text)
	}

	// (b) the standing members are a count, not re-rendered entries.
	for _, standing := range []string{"subject-2", "subject-3"} {
		if strings.Contains(text, standing) {
			t.Errorf("standing subject %q is re-rendered individually instead of being counted:\n%s", standing, text)
		}
	}
	if !strings.Contains(text, "2") {
		t.Errorf("the two standing subjects are not represented as a count:\n%s", text)
	}
}

// A first render has no previous result, so every non-clean entry is a change.
// Without this, "lead with what changed" is satisfied by a renderer that says
// nothing on the session where the reader has seen nothing yet.
func TestAdvisoryFirstRenderTreatsEveryNonCleanEntryAsChanged(t *testing.T) {
	now := time.Now()
	text, rec := Advisory(snapshotAt(resultA(), now.Add(-time.Minute)), RenderRecord{}, now)
	for _, subject := range []string{"subject-2", "subject-3"} {
		if !strings.Contains(text, subject) {
			t.Errorf("first render omits non-clean subject %q:\n%s", subject, text)
		}
	}
	if rec == nil {
		t.Fatal("first render recorded nothing")
	}
}

// When nothing changed, the advisory still speaks — REQ-GDL-004 fires on any
// non-clean entry — but it speaks as a count. A renderer that goes silent here
// would satisfy the collapse and drop the trigger.
func TestAdvisoryRendersAStandingCountWhenNothingChanged(t *testing.T) {
	now := time.Now()
	_, rec := Advisory(snapshotAt(resultA(), now.Add(-time.Hour)), RenderRecord{}, now)
	if rec == nil {
		t.Fatal("first render recorded nothing")
	}

	text, _ := Advisory(snapshotAt(resultA(), now.Add(-time.Minute)), *rec, now)
	if text == "" {
		t.Fatal("the advisory went silent while two subjects were still not reporting clean")
	}
	for _, standing := range []string{"subject-2", "subject-3"} {
		if strings.Contains(text, standing) {
			t.Errorf("standing subject %q is re-rendered on an unchanged result:\n%s", standing, text)
		}
	}
	if !strings.Contains(text, "2") {
		t.Errorf("the standing count is absent:\n%s", text)
	}
}

// An entry moving from one non-clean classification to another has CHANGED.
// A renderer diffing only membership of the non-clean set treats it as
// standing and never announces the move.
func TestAdvisoryTreatsAReclassifiedEntryAsChanged(t *testing.T) {
	now := time.Now()
	before := Result{
		Clean:   designate("alpha"),
		Entries: []Entry{entry("subject-2", "beta", "settled"), entry("subject-3", "gamma", "unsettled")},
	}
	after := Result{
		Clean:   designate("alpha"),
		Entries: []Entry{entry("subject-2", "gamma", "unsettled"), entry("subject-3", "gamma", "unsettled")},
	}

	_, rec := Advisory(snapshotAt(before, now.Add(-time.Hour)), RenderRecord{}, now)
	if rec == nil {
		t.Fatal("first render recorded nothing")
	}
	text, _ := Advisory(snapshotAt(after, now.Add(-time.Minute)), *rec, now)
	if lead := firstSubjectLine(text); !strings.Contains(lead, "subject-2") {
		t.Errorf("the reclassified subject does not lead; leading line is %q:\n%s", lead, text)
	}
	if strings.Contains(text, "subject-3") {
		t.Errorf("the unchanged subject is re-rendered individually:\n%s", text)
	}
}

// A result the partition could not be computed over yields no record. Recording
// one would let a later render treat entries it never read as standing, and a
// standing entry is one the reader is told about only as a number.
func TestAdvisoryRecordsNothingForAContractViolatingResult(t *testing.T) {
	now := time.Now()
	violating := Result{Clean: nil, Entries: []Entry{entry("subject-1", "alpha", "settled")}}
	text, rec := Advisory(snapshotAt(violating, now.Add(-time.Minute)), RenderRecord{}, now)
	if !strings.Contains(text, contractViolationMarker) {
		t.Fatalf("a contract-violating result did not render the violation:\n%s", text)
	}
	if rec != nil {
		t.Fatalf("a contract-violating result produced a render record: %+v", rec)
	}
}

// A snapshot with no recorded timestamp is not a measurement, so it records
// nothing either — otherwise a non-measurement would silence the next render.
func TestAdvisoryRecordsNothingWithoutARecordedTimestamp(t *testing.T) {
	text, rec := Advisory(Snapshot{Result: resultA()}, RenderRecord{}, time.Now())
	if text != "" || rec != nil {
		t.Fatalf("a timestamp-less snapshot rendered %q and recorded %+v", text, rec)
	}
}

// An all-clean result records the clean state, so an entry that goes non-clean
// again afterwards is announced rather than counted.
func TestAdvisoryRecordsAnAllCleanResultSoARelapseIsAnnounced(t *testing.T) {
	now := time.Now()
	allClean := Result{
		Clean:   designate("alpha"),
		Entries: []Entry{entry("subject-2", "alpha", "settled")},
	}
	relapsed := Result{
		Clean:   designate("alpha"),
		Entries: []Entry{entry("subject-2", "beta", "settled")},
	}

	text, rec := Advisory(snapshotAt(allClean, now.Add(-time.Hour)), RenderRecord{}, now)
	if text != "" {
		t.Fatalf("advisory spoke on an all-clean result: %q", text)
	}
	if rec == nil {
		t.Fatal("an all-clean render recorded nothing, so a relapse would read as standing")
	}
	next, _ := Advisory(snapshotAt(relapsed, now.Add(-time.Minute)), *rec, now)
	if !strings.Contains(next, "subject-2") {
		t.Errorf("a subject that went non-clean again was not announced:\n%s", next)
	}
}
