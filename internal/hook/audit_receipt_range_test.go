package hook

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// Receipt freshness after the audit (SPEC-MCP-NEUTRALITY-001 REQ-MN-020; card t2017).
//
// A receipt judged at commit C is reused by an auditor's PASS citing it. Any later
// commit moves both the head and the digest, so the receipt is stale whatever the
// commit touched; SPEC-AUDIT-ENGINE-CONVERGENCE-001 excludes the evidence-only
// relaxation (B3) by operator decision. The cases below pin that outcome, and
// audit_receipt_stale_test.go pins the evidence-only commits in particular.
//
// The fixtures are real git repositories under t.TempDir() and the check is driven
// through the auditor's SubagentStop step, the one production caller.

const (
	rangeSpec      = "SPEC-X-001"
	rangeOtherSpec = "SPEC-Y-001"
	rangeCard      = "t1"
)

const rangeProgress = `# Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready

## §E.2 Run-phase Evidence

| AC | Status |
|---|---|
| AC-1 | PASS |

## §E.3 Run-phase Audit-Ready Signal

run_status: pending

## §E.4 Sync-phase Audit-Ready Signal

sync_status: pending
`

type rangeRepo struct {
	root string // canonical tree root, also the store root
	rid  string // the receipt judged at the base commit
	base string // the commit the receipt judged
}

func rangeGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.test", "-c", "commit.gpgsign=false"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v (%s)", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func rangeWrite(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func rangeRead(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

func rangeCommit(t *testing.T, root, msg string) {
	t.Helper()
	rangeGit(t, root, "add", "-A")
	rangeGit(t, root, "commit", "-q", "-m", msg)
}

// newRangeRepo builds a repository whose tree holds an audited SPEC, a second SPEC,
// a verdict file, an evidence ledger and one source file; commits it; and records a
// receipt for an audit that judged that commit.
func newRangeRepo(t *testing.T) rangeRepo {
	t.Helper()
	root := newGateTree(t, "required")
	rangeWrite(t, root, ".gitignore", ".moai/state/\n")
	rangeWrite(t, root, ".moai/specs/"+rangeSpec+"/progress.md", rangeProgress)
	rangeWrite(t, root, ".moai/specs/"+rangeSpec+"/spec.md", "# spec\n")
	rangeWrite(t, root, ".moai/specs/"+rangeOtherSpec+"/progress.md", rangeProgress)
	rangeWrite(t, root, ".moai/reports/t1/verdict.md", "verdict: PASS\n")
	rangeWrite(t, root, ".moai/reports/t1/evidence-ledger.md", "ledger\n")
	rangeWrite(t, root, "internal/a.go", "package internal\n")
	rangeCommit(t, root, "base")
	seedStart(t, root, "a1", auditreceipt.AgentPlanAuditor, time.Now().UTC().Add(-time.Hour))
	rid := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolAuditMulti, TreeRoot: root, CodexVerdict: "pass"})
	return rangeRepo{root: root, rid: rid, base: rangeGit(t, root, "rev-parse", "HEAD")}
}

// rangeStop runs the auditor's Stop step with a PASS citing the receipt and the
// given spec= value, and reports whether the PASS was accepted and the causes of
// any refusal it recorded.
func rangeStop(t *testing.T, r rangeRepo, spec string) (accepted bool, causes []string) {
	t.Helper()
	line := "AUDIT-VERDICT: PASS spec=" + spec + " receipts=" + r.rid
	out := runStop(t, stopInput(r.root, auditreceipt.AgentPlanAuditor, "a1", "audited\n\n"+line, false))
	rej, err := auditreceipt.ListRejections(r.root)
	if err != nil {
		t.Fatalf("ListRejections: %v", err)
	}
	for _, x := range rej {
		causes = append(causes, x.Cause)
	}
	return out.Decision != "block" && len(rej) == 0, causes
}

func requireFresh(t *testing.T, r rangeRepo, spec string) {
	t.Helper()
	if ok, causes := rangeStop(t, r, spec); !ok {
		t.Errorf("the receipt judged at %s was refused after evidence-only commits (spec=%s): %v", r.base[:10], spec, causes)
	}
}

func requireStale(t *testing.T, r rangeRepo, spec string) {
	t.Helper()
	ok, causes := rangeStop(t, r, spec)
	if ok {
		t.Fatalf("the receipt judged at %s was accepted (spec=%s), want a refusal", r.base[:10], spec)
	}
	if len(causes) != 1 || causes[0] != auditreceipt.CauseReceiptTreeStateChanged {
		t.Errorf("refusal causes = %v, want exactly %q", causes, auditreceipt.CauseReceiptTreeStateChanged)
	}
}

func replaceOnce(t *testing.T, body, old, repl string) string {
	t.Helper()
	if !strings.Contains(body, old) {
		t.Fatalf("fixture text %q not found", old)
	}
	return strings.Replace(body, old, repl, 1)
}

// Any change after the audit makes the receipt stale, with
// CauseReceiptTreeStateChanged. Each row is one commit on top of the judged one.
func TestAEC_AnyChangeAfterTheAudit_MakesTheReceiptStale(t *testing.T) {
	progress := ".moai/specs/" + rangeSpec + "/progress.md"
	otherProgress := ".moai/specs/" + rangeOtherSpec + "/progress.md"

	rows := []struct {
		name string
		spec string
		do   func(t *testing.T, r rangeRepo)
	}{
		{"a .go file", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, "internal/a.go", "package internal\n\nvar X = 1\n")
		}},
		{"spec.md", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, ".moai/specs/"+rangeSpec+"/spec.md", "# spec\nchanged\n")
		}},
		{"a report outside the audit-artifact family", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, ".moai/reports/t1/evidence-ledger.md", "ledger\nedited\n")
		}},
		{"a path that only begins with the allowed prefix text", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, ".moai/reports-x/a", "x\n")
		}},
		{"exactly one line under E.2", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, progress, replaceOnce(t, rangeRead(t, r.root, progress), "| AC-1 | PASS |", "| AC-1 | FAIL |"))
		}},
		{"the E.2 heading deleted", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, progress, replaceOnce(t, rangeRead(t, r.root, progress), "## §E.2 Run-phase Evidence\n\n", ""))
		}},
		{"the E.2 heading renamed, keeping its token", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, progress, replaceOnce(t, rangeRead(t, r.root, progress), "## §E.2 Run-phase Evidence", "## §E.2 Run-phase Results"))
		}},
		{"the E.2 heading renamed, losing its token", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, progress, replaceOnce(t, rangeRead(t, r.root, progress), "## §E.2 Run-phase Evidence", "## Run-phase Evidence"))
		}},
		{"the E.2 section moved below E.3", rangeSpec, func(t *testing.T, r rangeRepo) {
			body := rangeRead(t, r.root, progress)
			e2, e3, e4 := strings.Index(body, "## §E.2"), strings.Index(body, "## §E.3"), strings.Index(body, "## §E.4")
			rangeWrite(t, r.root, progress, body[:e2]+body[e3:e4]+body[e2:e3]+body[e4:])
		}},
		{"the progress.md of another SPEC", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, otherProgress, replaceOnce(t, rangeRead(t, r.root, otherProgress), "plan_status: audit-ready", "plan_status: audit-ready (edited)"))
		}},
		{"a progress.md while the citation names a card id", rangeCard, func(t *testing.T, r rangeRepo) {
			rangeWrite(t, r.root, progress, replaceOnce(t, rangeRead(t, r.root, progress), "plan_status: audit-ready", "plan_status: audit-ready (edited)"))
		}},
		{"a report renamed into internal/", rangeSpec, func(t *testing.T, r rangeRepo) {
			if err := os.MkdirAll(filepath.Join(r.root, "internal"), 0o755); err != nil {
				t.Fatal(err)
			}
			rangeGit(t, r.root, "mv", ".moai/reports/t1/verdict.md", "internal/verdict.md")
		}},
		{"a source file renamed into the report family", rangeSpec, func(t *testing.T, r rangeRepo) {
			rangeGit(t, r.root, "mv", "internal/a.go", ".moai/reports/t1/verdict-a.md")
		}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := newRangeRepo(t)
			row.do(t, r)
			rangeCommit(t, r.root, row.name)
			requireStale(t, r, row.spec)
		})
	}

	t.Run("HEAD is not a descendant of the judged commit, whatever its tree", func(t *testing.T) {
		r := newRangeRepo(t)
		tree := rangeGit(t, r.root, "rev-parse", "HEAD^{tree}")
		orphan := rangeGit(t, r.root, "commit-tree", tree, "-m", "same tree, other history")
		rangeGit(t, r.root, "checkout", "-q", "--detach", orphan)
		requireStale(t, r, rangeSpec)
	})

	t.Run("an uncommitted code edit exists after an evidence-only commit", func(t *testing.T) {
		r := newRangeRepo(t)
		rangeWrite(t, r.root, ".moai/reports/t1/verdict.md", "verdict: PASS\nfiled\n")
		rangeCommit(t, r.root, "file the verdict")
		rangeWrite(t, r.root, "internal/a.go", "package internal\n\nvar Y = 2\n")
		requireStale(t, r, rangeSpec)
	})
}

// legacyReceipt rewrites the receipt as a record written before the change: only the
// keys that existed then, whatever keys the writer adds now.
func legacyReceipt(t *testing.T, r rangeRepo) {
	t.Helper()
	path := filepath.Join(auditreceipt.StateDir(r.root), "receipts", r.rid+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read receipt: %v", err)
	}
	var all map[string]any
	if err := json.Unmarshal(raw, &all); err != nil {
		t.Fatalf("decode receipt: %v", err)
	}
	known := map[string]bool{
		"receipt_id": true, "tool": true, "tree_root": true, "root_source": true, "created_at": true,
		"codex_verdict": true, "gate_unmet": true, "codex_error": true, "error_class": true,
		"head_commit": true, "tree_digest": true, "bind_capture_failed": true,
	}
	for k := range all {
		if !known[k] {
			delete(all, k)
		}
	}
	out, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatalf("write receipt: %v", err)
	}
}

// A receipt whose head equals the current HEAD is judged exactly as before. Controls:
// they pass today and must keep passing.
func TestAEC_HeadUnchanged_IsJudgedAsBefore_Control(t *testing.T) {
	t.Run("nothing changed", func(t *testing.T) {
		r := newRangeRepo(t)
		requireFresh(t, r, rangeSpec)
	})
	t.Run("an uncommitted code edit at the same head", func(t *testing.T) {
		r := newRangeRepo(t)
		rangeWrite(t, r.root, "internal/a.go", "package internal\n\nvar Y = 2\n")
		requireStale(t, r, rangeSpec)
	})
	t.Run("a legacy receipt at the same head", func(t *testing.T) {
		r := newRangeRepo(t)
		legacyReceipt(t, r)
		requireFresh(t, r, rangeSpec)
	})
}
