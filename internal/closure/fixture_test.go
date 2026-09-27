package closure

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// Fixture identifiers re-exported for the closure tests.
const (
	fixCard   = closuretest.Card
	fixSpecID = closuretest.SpecID
)

// acFixture is the acceptance.md §A repository, backed by closuretest.
type acFixture struct {
	*closuretest.Fixture
	t *testing.T
}

func newACFixture(t *testing.T) *acFixture {
	t.Helper()
	f := &acFixture{Fixture: closuretest.New(t), t: t}
	return f
}

// facts returns GitFacts over the card worktree via gitio.
func (f *acFixture) facts() GitFacts {
	return GitFacts{
		IsAncestor:      func(a, b string) (bool, error) { return gitio.IsAncestor(f.CardDir, a, b) },
		NonMergeCommits: func(from, to string) ([]CommitPaths, error) { return gitio.NonMergeCommits(f.CardDir, from, to) },
	}
}

// input assembles a BuildInput over the fixture's current files.
func (f *acFixture) input() BuildInput {
	f.t.Helper()
	rep, progress, acceptance, ac, acOK, err := LoadBuildSideFiles(f.CardDir, fixSpecID, closuretest.Autonomy(), nil, nil)
	if err != nil {
		f.t.Fatalf("load+verify: %v", err)
	}
	receipt, receiptOK, _ := readFileOrEmpty(
		filepath.Join(f.SpecDir, contract.ReceiptFile))
	ev := EvidenceFor(f.CardDir, fixCard)
	in := BuildInput{
		Card: fixCard, SpecID: fixSpecID, Home: f.CardDir,
		Mode: "contract", SecondReviewPolicy: "required",
		GeneratedAt:    time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC),
		Verify:         rep,
		Receipt:        receipt,
		ReceiptPresent: receiptOK,
		ProgressMD:     progress,
		AcceptanceMD:   acceptance,
		MeasuredAC:     ac,
		ACAvailable:    acOK,
		Facts:          f.facts(),
		EvalCommit:     f.Head(),
	}
	in.PreviousReportHash = PreviousReportHashOf(ev.ReportJSON)
	in.Records, in.UnreadableRecords = mustLoadA2Records(f.t, f.CardDir, fixCard)
	return in
}

// mustLoadA2Records loads the card's escalation records through the
// production loader.
func mustLoadA2Records(t *testing.T, home, card string) ([]escalation.Record, []string) {
	t.Helper()
	recs, unreadable, err := LoadA2Records(home, card)
	if err != nil {
		t.Fatalf("load A2 records: %v", err)
	}
	return recs, unreadable
}

// writeA2Record writes one escalation record file into the card's record
// directory (the records' storage shape is A2's: Markdown + frontmatter).
func (f *acFixture) writeA2Record(rec escalation.Record) {
	f.t.Helper()
	data, err := rec.Marshal()
	if err != nil {
		f.t.Fatalf("marshal record: %v", err)
	}
	name := rec.Class + "-" + rec.Fingerprint
	if rec.Occurrences > 1 {
		name += fmt.Sprintf("-%d", rec.Occurrences)
	}
	f.Write(filepath.Join(f.EscDir, name+".md"), string(data))
}

// writeA2Raw writes an arbitrary (possibly malformed) record file.
func (f *acFixture) writeA2Raw(name, content string) {
	f.t.Helper()
	f.Write(filepath.Join(f.EscDir, name), content)
}

// writeCardPlanAudit writes a plan-audit report into the card evidence
// directory.
func (f *acFixture) writeCardPlanAudit(name, lastLine string) {
	f.t.Helper()
	f.Write(filepath.Join(f.EvidenceDir, name), "# plan audit\n\n"+lastLine+"\n")
}

// writeGlobalPlanAudit writes a plan-audit report into the primary stream.
func (f *acFixture) writeGlobalPlanAudit(name, lastLine string) {
	f.t.Helper()
	f.Write(filepath.Join(f.CardDir, ".moai", "reports", "plan-audit", name),
		"# plan audit\n\n"+lastLine+"\n")
}

// WriteSecondReviewLine appends one second-review record line to the card
// evidence directory's second-review.jsonl.
func (f *acFixture) WriteSecondReviewLine(rec SecondReviewRecord) {
	f.t.Helper()
	data, err := json.Marshal(rec)
	if err != nil {
		f.t.Fatalf("marshal second-review record: %v", err)
	}
	f.Write(filepath.Join(f.EvidenceDir, SecondReviewFile), string(data)+"\n")
}
