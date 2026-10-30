package web

// context_model_source_test.go — ContextModelFrom against a real claim.Store
// (Issue #1948).
//
// Defect answer: these tests catch a ledger stored by ctxmodel.Update that the
// panel does not show, one Issue's ledger shown on another's row, a never-written
// ledger shown as an error or as data, a malformed stored document shown as a
// ledger, a nil reader that does not fail, and a source that can write.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

func csStore(t *testing.T) *claim.Store {
	t.Helper()
	s, err := claim.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return s
}

func csPut(t *testing.T, s *claim.Store, issue int, items ...ctxmodel.Item) {
	t.Helper()
	if _, err := ctxmodel.Update(s, issue, func(l *ctxmodel.Ledger) error {
		for _, it := range items {
			if err := l.Add(it); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("store ledger of #%d: %v", issue, err)
	}
}

var (
	csPlan  = ctxmodel.Item{Section: ctxmodel.SectionPlan, Title: "plan", Path: ".moai/specs/SPEC-X/plan.md", Commit: cmSHA, SHA256: cmHash, Owner: "lane-18"}
	csAudit = ctxmodel.Item{Section: ctxmodel.SectionAudit, Title: "audit", Owner: "lane-20", Pending: true}
)

func TestContextModelFrom_StoredLedgerReachesThePanel(t *testing.T) {
	s := csStore(t)
	csPut(t, s, 7, csPlan, csAudit)
	row := cmRow(t, cmApp(t, ContextModelFrom(s)))
	if !row.Ledger.Known || row.Ledger.Note != "" || len(row.Ledger.Items) != 2 {
		t.Fatalf("ledger = %+v", row.Ledger)
	}
	plan, audit := row.Ledger.Items[0], row.Ledger.Items[1]
	if plan.Title != "plan" || plan.Path != csPlan.Path || plan.Commit != cmSHA[:12] || plan.Status == "available" {
		t.Errorf("plan item = %+v", plan)
	}
	if audit.Status != "pending" || audit.Path != prGap {
		t.Errorf("audit item = %+v", audit)
	}
}

func TestContextModelFrom_OtherIssuesLedgerIsNotShown(t *testing.T) {
	s := csStore(t)
	csPut(t, s, 8, csPlan)
	row := cmRow(t, cmApp(t, ContextModelFrom(s)))
	if !row.Ledger.Known || len(row.Ledger.Items) != 0 {
		t.Fatalf("issue 8's ledger leaked onto issue 7: %+v", row.Ledger)
	}
}

func TestContextModelFrom_NeverWrittenReadsAsEmptyNotAsError(t *testing.T) {
	row := cmRow(t, cmApp(t, ContextModelFrom(csStore(t))))
	if !row.Ledger.Known || row.Ledger.Note != "" || len(row.Ledger.Items) != 0 {
		t.Fatalf("ledger = %+v", row.Ledger)
	}
}

func TestContextModelFrom_MalformedStoredDocumentIsUnreadable(t *testing.T) {
	s := csStore(t)
	if _, err := s.PutDoc("context", "issue-7", 0, []byte(`[]`)); err != nil {
		t.Fatalf("seed: %v", err)
	}
	row := cmRow(t, cmApp(t, ContextModelFrom(s)))
	if row.Ledger.Known || !strings.HasPrefix(row.Ledger.Note, "model unreadable") {
		t.Fatalf("a document that is not a ledger rendered as one: %+v", row.Ledger)
	}
}

func TestContextModelFrom_NilReaderFails(t *testing.T) {
	if _, err := ContextModelFrom(nil)(context.Background(), 7); err == nil {
		t.Fatal("a nil reader returned a model")
	}
	row := cmRow(t, cmApp(t, ContextModelFrom(nil)))
	if row.Ledger.Known || !strings.HasPrefix(row.Ledger.Note, "model unreadable") {
		t.Fatalf("ledger = %+v", row.Ledger)
	}
}

// The source holds a reader, and what it hands the ledger loader refuses every
// write: reading a ledger changes no stored revision.
func TestContextModelFrom_NeverWrites(t *testing.T) {
	s := csStore(t)
	csPut(t, s, 7, csPlan)
	before, err := s.GetDoc("context", "issue-7")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ContextModelFrom(s)(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if _, err := (readOnlyDocs{s}).UpdateDoc("context", "issue-7", func(cur []byte) ([]byte, error) { return cur, nil }); err == nil {
		t.Fatal("the read-only adapter accepted a write")
	}
	after, err := s.GetDoc("context", "issue-7")
	if err != nil {
		t.Fatal(err)
	}
	if before.Rev != after.Rev || string(before.Data) != string(after.Data) {
		t.Fatalf("stored ledger changed: rev %d -> %d", before.Rev, after.Rev)
	}
}
