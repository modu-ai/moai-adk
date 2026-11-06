package homestate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4 RED, sub-step S2c — the factory card reader
// accepts `verdict: INCONCLUSIVE` and keeps the card in audit as it does for FAIL,
// with the error class in its reason (AC-AEC-002b; REQ-AEC-007; card t2017).
//
// Today the E-VERDICT reader's pattern knows three labels, so an INCONCLUSIVE file
// reads as "no verdict line" and the refusal never names the unrun backend.

const inconclusiveSHA = "0123456789abcdef0123456789abcdef01234567"

// inconclusiveBody is an exported INCONCLUSIVE verdict file whose codex leg hit a
// quota.
func inconclusiveBody(sha string) string {
	return "# audit report\n\nverdict: INCONCLUSIVE\naudited_sha: " + sha + "\n" +
		"convergence_overall: inconclusive\n" +
		"required_backend: claude pass\n" +
		"required_backend: codex inconclusive quota\n"
}

// The reader (E-VERDICT, T6/T7/T12/T13) extracts the label and the audited SHA of
// an INCONCLUSIVE file; the pattern stays strict about the label.
func TestReadAuditVerdict_AcceptsAnInconclusiveLabel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".moai", "reports", "t2017", "sync-audit.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(inconclusiveBody(inconclusiveSHA)), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := readAuditVerdict(dir, "t2017", "sync-audit", inconclusiveSHA)
	if err != nil {
		t.Fatalf("readAuditVerdict refused an INCONCLUSIVE file: %v", err)
	}
	if v.Verdict != "INCONCLUSIVE" || v.AuditedSHA != inconclusiveSHA {
		t.Errorf("read %+v, want INCONCLUSIVE at %s", v, inconclusiveSHA)
	}

	label, sha, perr := ParseAuditVerdictFile("sync-audit.md", []byte(inconclusiveBody(inconclusiveSHA)), inconclusiveSHA)
	if perr != nil || label != "INCONCLUSIVE" || sha != inconclusiveSHA {
		t.Errorf("ParseAuditVerdictFile = (%q, %q, %v), want INCONCLUSIVE", label, sha, perr)
	}
}

// Controls (pass today): the three existing labels still read, and a lower-case or
// suffixed label is not a label.
func TestParseAuditVerdictFile_LabelsStayStrict_Control(t *testing.T) {
	for _, label := range []string{"PASS", "PASS-WITH-DEBT", "FAIL"} {
		got, _, err := ParseAuditVerdictFile("x.md", []byte("verdict: "+label+"\naudited_sha: "+inconclusiveSHA+"\n"), inconclusiveSHA)
		if err != nil || got != label {
			t.Errorf("label %s read as (%q, %v)", label, got, err)
		}
	}
	for _, bad := range []string{"inconclusive", "INCONCLUSIVE-X", "Inconclusive", "MAYBE"} {
		if _, _, err := ParseAuditVerdictFile("x.md", []byte("verdict: "+bad+"\naudited_sha: "+inconclusiveSHA+"\n"), inconclusiveSHA); err == nil {
			t.Errorf("label %q was accepted", bad)
		}
	}
}

// The card-transition guard holds the card in audit on an INCONCLUSIVE plan-audit
// verdict, and the refusal names the unrun backend and its class. The record is
// unchanged by the refused exit.
func TestFR_M4_InconclusivePlanAuditVerdict_HoldsTheCardInAuditNamingTheClass(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()

	c := frLeasedCard(repo, "inc-plan", CardPlanAudit)
	c.EvidenceSHA = repo.Commit
	c.EvidencePath = repo.Artifact
	frPlace(t, db, c)
	frWrite(t, filepath.Join(repo.Dir, ".moai", "reports", "inc-plan", "plan-audit.md"), inconclusiveBody(repo.Commit))

	before := frRowDump(t, db, frRun, "inc-plan")
	_, err := db.Transition(ctx, frHolderRequest(c, CardKickoff))
	if !errors.Is(err, ErrEvidence) {
		t.Fatalf("err = %v, want ErrEvidence", err)
	}
	for _, want := range []string{"codex", "quota"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal %q does not name %q", err.Error(), want)
		}
	}
	if strings.Contains(err.Error(), "carries no `verdict:` line") {
		t.Errorf("the refusal reads the INCONCLUSIVE file as having no verdict line: %v", err)
	}
	if after := frRowDump(t, db, frRun, "inc-plan"); after != before {
		t.Errorf("the refused exit changed the record")
	}
}

// Through admitCardVerdict (the sync phase reads the label alone, then the shared
// predicate): the class-naming reason reaches the card seam. This arm already holds
// once the shared admission rule refuses an INCONCLUSIVE overall; it guards the
// propagation.
func TestAdmitCardVerdict_InconclusiveSyncVerdict_NamesTheClass(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sync-audit.md")
	if err := os.WriteFile(path, []byte(inconclusiveBody(inconclusiveSHA)), 0o600); err != nil {
		t.Fatal(err)
	}
	card := Card{State: CardSyncAudit, WorktreePath: t.TempDir()}
	ok, reason := admitCardVerdict(card, path)
	if ok {
		t.Fatal("a sync-audit INCONCLUSIVE verdict was admitted")
	}
	for _, want := range []string{"codex", "quota"} {
		if !strings.Contains(reason, want) {
			t.Errorf("reason %q does not name %q", reason, want)
		}
	}
}
