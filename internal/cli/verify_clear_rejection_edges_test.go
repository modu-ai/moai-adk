package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M2 — edge cases of the clear-rejection verb
// beyond the AC rows (REQ-AEC-012; card t2017). Like verify_clear_rejection_test.go
// every test works in a t.TempDir() tree and never touches a real
// .moai/state/audit-receipts.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// Evidence the verb cannot read in full, or that is not a file, is refused with
// the evidence-path code and nothing changes.
func TestClearRejection_UnusableEvidence_IsRefused_NothingChanges(t *testing.T) {
	rows := map[string]func(w *crWorld) string{
		"the evidence file does not exist": func(w *crWorld) string {
			w.evidence("other.md", "x\n") // creates .moai/reports/tcard
			return filepath.Join(w.tree, ".moai", "reports", "tcard", "absent.md")
		},
		"the evidence path is a directory": func(w *crWorld) string {
			w.evidence("other.md", "x\n")
			d := filepath.Join(w.tree, ".moai", "reports", "tcard", "adir")
			if err := os.MkdirAll(d, 0o755); err != nil {
				w.t.Fatalf("mkdir: %v", err)
			}
			return d
		},
		"the evidence file is larger than the read limit": func(w *crWorld) string {
			return w.evidence("big.md", crSpecA+"\n"+strings.Repeat("x", clearEvidenceMaxBytes+16))
		},
		"the tree has no .moai/reports/ directory": func(w *crWorld) string {
			return filepath.Join(w.tree, ".moai", "reports", "tcard", "verdict.md")
		},
	}
	for name, mk := range rows {
		t.Run(name, func(t *testing.T) {
			w := newCRWorld(t)
			w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
			ev := mk(w)
			before := w.snapshot()
			out, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, ev)
			wantRefusal(t, out, err, crRefEvidencePath)
			assertStoreUnchanged(t, before, w.snapshot())
		})
	}
}

// The SPEC ID must appear as a whole identifier: SPEC-ALPHA-0012 is not
// SPEC-ALPHA-001, and a longer prefix such as XSPEC-ALPHA-001 does not name it.
func TestClearRejection_EvidenceNamingASimilarID_IsRefused(t *testing.T) {
	for name, body := range map[string]string{
		"a longer ID that starts with it":        "re-audited " + crSpecA + "2 only\n",
		"an ID that ends with it after a letter": "re-audited X" + crSpecA + " only\n",
	} {
		t.Run(name, func(t *testing.T) {
			w := newCRWorld(t)
			w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
			ev := w.evidence("verdict.md", body)
			before := w.snapshot()
			out, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, ev)
			wantRefusal(t, out, err, crRefEvidenceSpec)
			assertStoreUnchanged(t, before, w.snapshot())
		})
	}
	t.Run("the ID inside prose and punctuation is named", func(t *testing.T) {
		w := newCRWorld(t)
		markA := w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
		ev := w.evidence("verdict.md", "(see "+crSpecA+"), re-audited.\n")
		if out, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, ev); err != nil {
			t.Fatalf("refused: %v\n%s", err, out)
		}
		if _, statErr := os.Stat(markA); !os.IsNotExist(statErr) {
			t.Errorf("the mark must be cleared")
		}
	})
}

// A missing git identity refuses after the typed confirmation, before any write.
func TestClearRejection_MissingGitIdentity_IsRefused_NothingChanges(t *testing.T) {
	w := newCRWorld(t)
	w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
	ev := w.evidence("verdict.md", "evidence for "+crSpecA+"\n")
	saved := contractGitIdentityFn
	contractGitIdentityFn = func(string) (string, string, error) { return "", "", errors.New("git user.name is not set") }
	t.Cleanup(func() { contractGitIdentityFn = saved })
	before := w.snapshot()
	out, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, ev)
	wantRefusal(t, out, err, "git_identity_missing")
	assertStoreUnchanged(t, before, w.snapshot())
}

// A relative evidence path is resolved against the tree, not the process cwd.
func TestClearRejection_RelativeEvidencePath_ResolvesAgainstTheTree(t *testing.T) {
	w := newCRWorld(t)
	markA := w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
	w.evidence("verdict.md", "evidence for "+crSpecA+"\n")
	rel := filepath.Join(".moai", "reports", "tcard", "verdict.md")
	if out, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, rel); err != nil {
		t.Fatalf("refused a tree-relative evidence path: %v\n%s", err, out)
	}
	if _, statErr := os.Stat(markA); !os.IsNotExist(statErr) {
		t.Errorf("the mark must be cleared")
	}
}

// A missing --evidence is a usage error (exit 2), not a refusal, and changes nothing.
func TestClearRejection_MissingEvidenceFlag_IsAUsageError(t *testing.T) {
	w := newCRWorld(t)
	w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
	before := w.snapshot()
	rootCmd.SetIn(strings.NewReader(typed(crAgent, crSpecA)))
	t.Cleanup(func() { rootCmd.SetIn(nil) })
	_, err := execRoot(t, "verify", crVerb, crAgent, crSpecA, "--project-root", w.tree)
	var ee *exitCodeError
	if !errors.As(err, &ee) || ee.code != 2 {
		t.Errorf("want a usage error with exit 2, got %v", err)
	}
	assertStoreUnchanged(t, before, w.snapshot())
}

// A mark file that cannot be parsed is neither "absent" nor clearable: the verb
// reports an I/O-class error (exit 2) and removes nothing.
func TestClearRejection_UnreadableMarkFile_IsAnErrorNotAClear(t *testing.T) {
	w := newCRWorld(t)
	markA := w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
	if err := os.WriteFile(markA, []byte("{ not json"), 0o644); err != nil {
		t.Fatalf("corrupt the mark: %v", err)
	}
	ev := w.evidence("verdict.md", "evidence for "+crSpecA+"\n")
	before := w.snapshot()
	_, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, ev)
	var ee *exitCodeError
	if !errors.As(err, &ee) || ee.code != 2 {
		t.Errorf("want exit 2 for an unreadable mark, got %v", err)
	}
	assertStoreUnchanged(t, before, w.snapshot())
}
