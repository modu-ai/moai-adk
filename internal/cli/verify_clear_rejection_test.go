package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M2 RED — the one operator path that clears a
// receipt-kind rejection mark (AC-AEC-004a, AC-AEC-004b, AC-AEC-004c;
// REQ-AEC-011, REQ-AEC-012, REQ-AEC-013; card t2017).
//
// Every test builds its own tree under t.TempDir() and points the verb at it with
// --project-root. NO test in this file reads or writes the repository's real
// .moai/state/audit-receipts: the tree, its store, the evidence file, the marks
// and the approval log all live in the temporary directory.
//
// The verb does not exist yet. cobra answers an unknown verb under `moai verify`
// with the help text and a nil error, so each positive test fails because the mark
// is still there, and each refusal test fails because no `refused <code>` exit
// error came back. None of this needs a production symbol.
//
// ASSUMED NAMES. The SPEC fixes behavior, not spelling. Everything this file
// commits to is in the constants block below; change it there and nowhere else.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/hook"
)

const (
	crVerb         = "clear-rejection" // moai verify clear-rejection <agent-type> <spec-id>
	crEvidenceFlag = "--evidence"

	// Refusal codes: the exit error reads "refused <code>: <detail>", exit code 1,
	// the shape `moai contract verdict` already uses.
	crRefAgentMarker  = "agent_marker"
	crRefNotTTY       = "not_tty"
	crRefMarkNotFound = "mark_not_found" // absent from this tree, or recorded for another tree
	crRefEvidencePath = "evidence_path"  // outside .moai/reports/
	crRefEvidenceSpec = "evidence_spec"  // does not name the mark's SPEC ID
	crRefConfirmation = "confirmation_mismatch"

	// crApprovalsFile is the append-only approval log inside the audit-receipt
	// store (REQ-AEC-013). Only the hand-written-line guard names it.
	crApprovalsFile = "approvals.jsonl"

	// crStoreFunc is the store function the verb alone may call (AC-AEC-004c).
	crStoreFunc = "ClearRejectionOnApproval"
	crVerbFile  = "verify_clear_rejection.go"

	crAgent = auditreceipt.AgentPlanAuditor
	crSpecA = "SPEC-ALPHA-001"
	crSpecB = "SPEC-BETA-002"
)

// crWorld is one throwaway project tree with an audit-receipt store.
type crWorld struct {
	t     *testing.T
	tree  string // canonical tree root; also the store root
	env   map[string]string
	isTTY bool
	n     int
}

func newCRWorld(t *testing.T) *crWorld {
	t.Helper()
	root := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root, "-c", "user.name=Fixture", "-c", "user.email=f@example.test"}, args...)...)
		cmd.Env = closuretest.ScrubbedEnv()
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Skipf("git unavailable or failed: git %s: %v (%s)", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q")
	if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// The codex audit gate is declared required: the spawn check then enforces.
	body := "workflow:\n  audit:\n    gates:\n      codex: required\n"
	if err := os.WriteFile(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	run("add", "seed.txt")
	run("commit", "-q", "-m", "seed")

	w := &crWorld{t: t, tree: auditreceipt.Canonical(root), env: map[string]string{}, isTTY: true}

	// Seams: a scrubbed environment, a terminal on standard input, a fixed git
	// identity. The runtime's own environment never reaches the verb.
	savedEnv, savedTTY, savedID := contractGetenvFn, contractStdinIsTerminalFn, contractGitIdentityFn
	contractGetenvFn = func(k string) string { return w.env[k] }
	contractStdinIsTerminalFn = func() bool { return w.isTTY }
	contractGitIdentityFn = func(string) (string, string, error) { return "Fixture Operator", "operator@example.test", nil }
	t.Cleanup(func() {
		contractGetenvFn, contractStdinIsTerminalFn, contractGitIdentityFn = savedEnv, savedTTY, savedID
	})
	return w
}

func (w *crWorld) head() string {
	w.t.Helper()
	cmd := exec.Command("git", "-C", w.tree, "rev-parse", "HEAD")
	cmd.Env = closuretest.ScrubbedEnv()
	out, err := cmd.Output()
	if err != nil {
		w.t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// mark writes one receipt-kind rejection mark for this tree and returns its path.
func (w *crWorld) mark(agent, spec, cause string) string {
	w.t.Helper()
	r := auditreceipt.Rejection{AgentType: agent, SpecID: spec, Cause: cause, TreeRoot: w.tree}
	if err := auditreceipt.WriteRejection(w.tree, &r); err != nil {
		w.t.Fatalf("WriteRejection: %v", err)
	}
	return filepath.Join(auditreceipt.StateDir(w.tree), "rejections", agent+"--"+spec+".json")
}

// evidence writes an evidence file under .moai/reports/ and returns its path.
func (w *crWorld) evidence(name, body string) string {
	w.t.Helper()
	p := filepath.Join(w.tree, ".moai", "reports", "tcard", name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		w.t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		w.t.Fatalf("write evidence: %v", err)
	}
	return p
}

// verb runs `moai verify clear-rejection <agent> <spec> --evidence <path>` with
// stdin as typed. A fresh rootCmd run each time (execRoot resets flag state).
func (w *crWorld) verb(stdin, agent, spec, evidence string) (string, error) {
	w.t.Helper()
	rootCmd.SetIn(strings.NewReader(stdin))
	w.t.Cleanup(func() { rootCmd.SetIn(nil) })
	return execRoot(w.t, "verify", crVerb, agent, spec, crEvidenceFlag, evidence, "--project-root", w.tree)
}

// snapshot returns every file under the tree's audit-receipt store, by relative
// path, with its bytes.
func (w *crWorld) snapshot() map[string]string {
	w.t.Helper()
	out := map[string]string{}
	base := auditreceipt.StateDir(w.tree)
	_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			w.t.Fatalf("read %s: %v", p, rerr)
		}
		rel, _ := filepath.Rel(base, p)
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	return out
}

// spawnDeny drives the real PreToolUse spawn check for a phase-entry agent in
// this tree and returns whether it denied, and the reason.
func (w *crWorld) spawnDeny() (bool, string) {
	w.t.Helper()
	w.t.Setenv("CLAUDE_PROJECT_DIR", w.tree) // the spawn path also writes its model audit log here
	raw, _ := json.Marshal(map[string]string{"subagent_type": "manager-develop", "prompt": "go"})
	out, err := hook.NewPreToolHandler(nil, hook.DefaultSecurityPolicy()).Handle(context.Background(), &hook.HookInput{
		CWD: w.tree, ToolName: "Agent", ToolInput: raw, HookEventName: "PreToolUse",
	})
	if err != nil {
		w.t.Fatalf("PreToolUse Handle: %v", err)
	}
	if out == nil || out.HookSpecificOutput == nil {
		return false, ""
	}
	h := out.HookSpecificOutput
	return h.PermissionDecision == hook.DecisionDeny, h.PermissionDecisionReason
}

// wantRefusal asserts err is the verb's refusal: exit code 1, "refused <code>".
func wantRefusal(t *testing.T, out string, err error, code string) {
	t.Helper()
	var ee *exitCodeError
	if !errors.As(err, &ee) {
		t.Errorf("want a refusal %q (exit error), got err=%v\noutput: %s", code, err, out)
		return
	}
	if ee.code != 1 || !strings.HasPrefix(ee.msg, "refused "+code) {
		t.Errorf("want exit 1 %q, got exit %d %q", "refused "+code, ee.code, ee.msg)
	}
}

func typed(agent, spec string) string { return agent + " " + spec + "\n" }

// ─── AC-AEC-004a ───

// AC-AEC-004a — an operator clears one named mark: one approval record carries
// the operator, the mark as read, the evidence path and SHA-256, HEAD and a UTC
// time; the named mark is gone; the other mark is byte-identical; and the spawn
// check no longer lists the cleared mark.
func TestAEC004a_ClearRejection_OperatorApproval_ClearsExactlyTheNamedMark(t *testing.T) {
	w := newCRWorld(t)
	markA := w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
	markB := w.mark(crAgent, crSpecB, auditreceipt.CauseReceiptAuditNotRun)
	body := "# audit evidence\n\nre-audited " + crSpecA + " through codex_audit; verdict PASS\n"
	ev := w.evidence("verdict.md", body)

	// Control (passes today): the fixture is faithful. Both marks stand and the
	// spawn check lists both.
	if denied, reason := w.spawnDeny(); !denied || !strings.Contains(reason, crSpecA) || !strings.Contains(reason, crSpecB) {
		t.Fatalf("control: the spawn check must deny and list both marks; denied=%v reason=%q", denied, reason)
	}
	bBefore, err := os.ReadFile(markB)
	if err != nil {
		t.Fatalf("read mark B: %v", err)
	}
	before := w.snapshot()

	out, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, ev)
	if err != nil {
		t.Fatalf("the verb refused a correct approval: %v\noutput: %s", err, out)
	}

	if _, statErr := os.Stat(markA); !os.IsNotExist(statErr) {
		t.Errorf("the SPEC-A mark must be gone, stat err = %v", statErr)
	}
	if bAfter, rerr := os.ReadFile(markB); rerr != nil || !bytes.Equal(bAfter, bBefore) {
		t.Errorf("the SPEC-B mark must be byte-identical (read err %v)", rerr)
	}

	// Exactly one new file under the store: the approval log, with exactly one line.
	after := w.snapshot()
	var added []string
	for k := range after {
		if _, had := before[k]; !had {
			added = append(added, k)
		}
	}
	if len(added) != 1 {
		t.Fatalf("exactly one file (the approval log) must be added to the store, got %v", added)
	}
	lines := strings.Split(strings.TrimRight(after[added[0]], "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("the approval log must hold exactly one record, got %d lines", len(lines))
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &rec); err != nil {
		t.Fatalf("the approval record is not one JSON object: %v\n%s", err, lines[0])
	}

	sum := aecFileSHA256(t, ev)
	for what, want := range map[string]string{
		"the operator name":      "Fixture Operator",
		"the git identity":       "operator@example.test",
		"the mark's agent type":  crAgent,
		"the mark's SPEC ID":     crSpecA,
		"the mark's cause":       auditreceipt.CauseNoReceiptCited,
		"the evidence path":      ".moai/reports/tcard/verdict.md",
		"the evidence SHA-256":   sum,
		"the tree's HEAD":        w.head(),
		"the typed confirmation": crAgent + " " + crSpecA,
	} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("the approval record does not carry %s (%q): %s", what, want, lines[0])
		}
	}
	if strings.Contains(lines[0], crSpecB) {
		t.Errorf("the approval record names the OTHER mark: %s", lines[0])
	}
	if !hasUTCTime(rec) {
		t.Errorf("the approval record carries no UTC (RFC 3339, Z) time: %s", lines[0])
	}

	// The spawn check no longer lists SPEC-A, and still lists SPEC-B.
	denied, reason := w.spawnDeny()
	if strings.Contains(reason, crSpecA) {
		t.Errorf("the spawn check still lists the cleared mark: %s", reason)
	}
	if !denied || !strings.Contains(reason, crSpecB) {
		t.Errorf("the spawn check must still deny for the other mark; denied=%v reason=%q", denied, reason)
	}
}

// hasUTCTime reports whether any string value of rec parses as RFC 3339 in UTC.
func hasUTCTime(rec map[string]any) bool {
	for _, v := range rec {
		if s, ok := v.(string); ok && strings.HasSuffix(s, "Z") {
			if _, err := time.Parse(time.RFC3339Nano, s); err == nil {
				return true
			}
		}
	}
	return false
}

// AC-AEC-004a (edge, acceptance.md section E) — an UnknownSpec mark is clearable
// only by naming `unknown` exactly; the evidence check then requires the agent
// type in the evidence file.
func TestAEC004a_ClearRejection_UnknownSpecMark_ByNamingUnknown(t *testing.T) {
	w := newAECCRUnknown(t)
	ev := w.evidence("verdict.md", "re-audit by "+crAgent+" ended with a verdict line\n")
	out, err := w.verb(typed(crAgent, "unknown"), crAgent, "unknown", ev)
	if err != nil {
		t.Fatalf("the verb refused a correct unknown-spec approval: %v\noutput: %s", err, out)
	}
	if _, statErr := os.Stat(filepath.Join(auditreceipt.StateDir(w.tree), "rejections", crAgent+"--"+auditreceipt.UnknownSpec+".json")); !os.IsNotExist(statErr) {
		t.Errorf("the unknown-spec mark must be gone, stat err = %v", statErr)
	}
}

func newAECCRUnknown(t *testing.T) *crWorld {
	t.Helper()
	w := newCRWorld(t)
	w.mark(crAgent, "", auditreceipt.CauseVerdictLineMissing) // an empty SPEC ID is keyed unknown-spec
	return w
}

// AC-AEC-004a (edge) — the unknown-spec evidence check requires the agent type;
// an evidence file that does not name it is refused and nothing changes.
func TestAEC004a_ClearRejection_UnknownSpecMark_EvidenceMustNameTheAgentType(t *testing.T) {
	w := newAECCRUnknown(t)
	ev := w.evidence("verdict.md", "some unrelated notes\n")
	before := w.snapshot()
	out, err := w.verb(typed(crAgent, "unknown"), crAgent, "unknown", ev)
	wantRefusal(t, out, err, crRefEvidenceSpec)
	assertStoreUnchanged(t, before, w.snapshot())
}

// ─── AC-AEC-004b ───

func assertStoreUnchanged(t *testing.T, before, after map[string]string) {
	t.Helper()
	if len(before) != len(after) {
		t.Errorf("the store changed: %d files before, %d after", len(before), len(after))
	}
	for k, v := range before {
		if got, ok := after[k]; !ok {
			t.Errorf("a refused call removed %s", k)
		} else if got != v {
			t.Errorf("a refused call changed %s", k)
		}
	}
	for k := range after {
		if _, ok := before[k]; !ok {
			t.Errorf("a refused call wrote %s", k)
		}
	}
}

// AC-AEC-004b — every refusal leaves the approvals log and every mark unchanged.
func TestAEC004b_ClearRejection_Refusals_ChangeNothing(t *testing.T) {
	type row struct {
		name  string
		setup func(w *crWorld) (stdin, agent, spec, evidence string)
		code  string
	}
	good := func(w *crWorld) (string, string, string, string) {
		return typed(crAgent, crSpecA), crAgent, crSpecA, w.evidence("verdict.md", "evidence for "+crSpecA+"\n")
	}
	rows := []row{
		{"CLAUDECODE is set", func(w *crWorld) (string, string, string, string) {
			w.env["CLAUDECODE"] = "1"
			return good(w)
		}, crRefAgentMarker},
		{"CLAUDE_CODE_SESSION_ID is set", func(w *crWorld) (string, string, string, string) {
			w.env["CLAUDE_CODE_SESSION_ID"] = "sess-1"
			return good(w)
		}, crRefAgentMarker},
		{"standard input is not a terminal", func(w *crWorld) (string, string, string, string) {
			w.isTTY = false
			return good(w)
		}, crRefNotTTY},
		{"the mark does not exist", func(w *crWorld) (string, string, string, string) {
			ev := w.evidence("verdict.md", "evidence for SPEC-NOPE-999\n")
			return typed(crAgent, "SPEC-NOPE-999"), crAgent, "SPEC-NOPE-999", ev
		}, crRefMarkNotFound},
		{"the mark was recorded for another tree in the same store", func(w *crWorld) (string, string, string, string) {
			other := auditreceipt.Rejection{AgentType: crAgent, SpecID: "SPEC-OTHER-003", Cause: auditreceipt.CauseNoReceiptCited, TreeRoot: filepath.Join(w.tree, "elsewhere")}
			if err := auditreceipt.WriteRejectionIn(w.tree, &other); err != nil {
				w.t.Fatalf("WriteRejectionIn: %v", err)
			}
			ev := w.evidence("verdict.md", "evidence for SPEC-OTHER-003\n")
			return typed(crAgent, "SPEC-OTHER-003"), crAgent, "SPEC-OTHER-003", ev
		}, crRefMarkNotFound},
		{"a mark file at this tree's path names another tree", func(w *crWorld) (string, string, string, string) {
			// A copied file: it sits where this tree's mark would, but its own
			// tree_root names a different tree.
			p := w.mark(crAgent, "SPEC-COPIED-004", auditreceipt.CauseNoReceiptCited)
			b, err := os.ReadFile(p)
			if err != nil {
				w.t.Fatalf("read: %v", err)
			}
			forged := strings.Replace(string(b), w.tree, filepath.Join(w.tree, "elsewhere"), 1)
			if err := os.WriteFile(p, []byte(forged), 0o644); err != nil {
				w.t.Fatalf("write: %v", err)
			}
			ev := w.evidence("verdict.md", "evidence for SPEC-COPIED-004\n")
			return typed(crAgent, "SPEC-COPIED-004"), crAgent, "SPEC-COPIED-004", ev
		}, crRefMarkNotFound},
		{"the evidence file is outside .moai/reports/", func(w *crWorld) (string, string, string, string) {
			p := filepath.Join(w.tree, "NOTES.md")
			if err := os.WriteFile(p, []byte("evidence for "+crSpecA+"\n"), 0o644); err != nil {
				w.t.Fatalf("write: %v", err)
			}
			return typed(crAgent, crSpecA), crAgent, crSpecA, p
		}, crRefEvidencePath},
		{"the evidence path climbs out of .moai/reports/ with ..", func(w *crWorld) (string, string, string, string) {
			p := filepath.Join(w.tree, "NOTES.md")
			if err := os.WriteFile(p, []byte("evidence for "+crSpecA+"\n"), 0o644); err != nil {
				w.t.Fatalf("write: %v", err)
			}
			climb := filepath.Join(w.tree, ".moai", "reports", "..", "..", "NOTES.md")
			return typed(crAgent, crSpecA), crAgent, crSpecA, climb
		}, crRefEvidencePath},
		{"the evidence path is a prefix trick (.moai/reports-x)", func(w *crWorld) (string, string, string, string) {
			p := filepath.Join(w.tree, ".moai", "reports-x", "v.md")
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				w.t.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(p, []byte("evidence for "+crSpecA+"\n"), 0o644); err != nil {
				w.t.Fatalf("write: %v", err)
			}
			return typed(crAgent, crSpecA), crAgent, crSpecA, p
		}, crRefEvidencePath},
		{"the evidence file is a symlink out of .moai/reports/", func(w *crWorld) (string, string, string, string) {
			target := filepath.Join(w.tree, "NOTES.md")
			if err := os.WriteFile(target, []byte("evidence for "+crSpecA+"\n"), 0o644); err != nil {
				w.t.Fatalf("write: %v", err)
			}
			link := filepath.Join(w.tree, ".moai", "reports", "tcard", "link.md")
			if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
				w.t.Fatalf("mkdir: %v", err)
			}
			if err := os.Symlink(target, link); err != nil {
				w.t.Skipf("symlinks unavailable here: %v", err)
			}
			return typed(crAgent, crSpecA), crAgent, crSpecA, link
		}, crRefEvidencePath},
		{"the evidence file does not name the mark's SPEC ID", func(w *crWorld) (string, string, string, string) {
			return typed(crAgent, crSpecA), crAgent, crSpecA, w.evidence("verdict.md", "evidence for "+crSpecB+" only\n")
		}, crRefEvidenceSpec},
		{"the typed confirmation names another agent type", func(w *crWorld) (string, string, string, string) {
			_, a, s, e := good(w)
			return typed(auditreceipt.AgentSyncAuditor, crSpecA), a, s, e
		}, crRefConfirmation},
		{"the typed confirmation names another SPEC", func(w *crWorld) (string, string, string, string) {
			_, a, s, e := good(w)
			return typed(crAgent, crSpecB), a, s, e
		}, crRefConfirmation},
		{"nothing is typed", func(w *crWorld) (string, string, string, string) {
			_, a, s, e := good(w)
			return "", a, s, e
		}, crRefConfirmation},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			w := newCRWorld(t)
			w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
			w.mark(crAgent, crSpecB, auditreceipt.CauseReceiptAuditNotRun)
			stdin, agent, spec, ev := r.setup(w)
			before := w.snapshot()
			out, err := w.verb(stdin, agent, spec, ev)
			wantRefusal(t, out, err, r.code)
			assertStoreUnchanged(t, before, w.snapshot())
		})
	}
}

// AC-AEC-004b — the marker set the verb checks equals the contract signing
// path's set. Observed by behavior: every member of contractAgentMarkers() makes
// the verb refuse on its own, and a marker outside the set does not (a Codex
// session sets neither marker and is not detected, spec.md section E B4).
func TestAEC004b_ClearRejection_MarkerSetEqualsContractSigningPathSet(t *testing.T) {
	for _, m := range contractAgentMarkers() {
		t.Run("refuses on "+m, func(t *testing.T) {
			w := newCRWorld(t)
			w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
			ev := w.evidence("verdict.md", "evidence for "+crSpecA+"\n")
			w.env[m] = "set"
			before := w.snapshot()
			out, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, ev)
			wantRefusal(t, out, err, crRefAgentMarker)
			assertStoreUnchanged(t, before, w.snapshot())
		})
	}
	t.Run("a marker outside the set is not a marker", func(t *testing.T) {
		w := newCRWorld(t)
		markA := w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
		ev := w.evidence("verdict.md", "evidence for "+crSpecA+"\n")
		w.env["CODEX_SESSION_ID"] = "set"
		if out, err := w.verb(typed(crAgent, crSpecA), crAgent, crSpecA, ev); err != nil {
			t.Fatalf("a variable outside the signing path's set must not refuse: %v\n%s", err, out)
		}
		if _, statErr := os.Stat(markA); !os.IsNotExist(statErr) {
			t.Errorf("the mark must be cleared, stat err = %v", statErr)
		}
	})
}

// AC-AEC-004b — control (passes today, EV-19): the set the verb must equal is the
// two Claude Code markers.
func TestAEC004b_ContractAgentMarkers_AreTheTwoClaudeCodeMarkers_Control(t *testing.T) {
	got := contractAgentMarkers()
	if len(got) != 2 || got[0] != "CLAUDECODE" || got[1] != "CLAUDE_CODE_SESSION_ID" {
		t.Errorf("contractAgentMarkers() = %v, want [CLAUDECODE CLAUDE_CODE_SESSION_ID]", got)
	}
}

// ─── AC-AEC-004c ───

// AC-AEC-004c — an approval line written into the log by hand clears nothing: no
// process reads the log as authority, and the spawn check still denies. Passes
// today (the log is read by nobody); it guards the "honour a hand-written line"
// mutant once the log exists.
func TestAEC004c_HandWrittenApprovalLine_DoesNotClearTheSpawnDeny(t *testing.T) {
	w := newCRWorld(t)
	markA := w.mark(crAgent, crSpecA, auditreceipt.CauseNoReceiptCited)
	line := `{"operator":"Fixture Operator","email":"operator@example.test","agent_type":"` + crAgent +
		`","spec_id":"` + crSpecA + `","evidence_path":".moai/reports/tcard/verdict.md","approved_at":"2026-11-03T00:00:00Z"}` + "\n"
	logPath := filepath.Join(auditreceipt.StateDir(w.tree), crApprovalsFile)
	if err := os.WriteFile(logPath, []byte(line), 0o644); err != nil {
		t.Fatalf("write approval line: %v", err)
	}
	denied, reason := w.spawnDeny()
	if !denied || !strings.Contains(reason, crSpecA) {
		t.Errorf("a hand-written approval line must not clear the spawn deny; denied=%v reason=%q", denied, reason)
	}
	if _, err := os.Stat(markA); err != nil {
		t.Errorf("the mark must still exist: %v", err)
	}
}

// AC-AEC-004c — the built server lists no tool that clears a rejection. Passes
// today; it holds the MCP-path exclusion in place (REQ-AEC-011).
func TestAEC004c_NoMCPToolClearsARejection(t *testing.T) {
	for name, tool := range selfReviewListTools(t) {
		lower := strings.ToLower(name)
		desc, _ := tool["description"].(string)
		d := strings.ToLower(desc)
		if strings.Contains(lower, "reject") || strings.Contains(lower, "approv") ||
			(strings.Contains(d, "rejection") && (strings.Contains(d, "clear") || strings.Contains(d, "remov"))) {
			t.Errorf("MCP tool %q reads as a rejection-clearing path: %s", name, desc)
		}
	}
}

// AC-AEC-004c — the store function that clears a mark on approval has exactly one
// non-test caller, and it is the verb. A source sweep over the Go files of the
// module: any call, qualified or not, of crStoreFunc outside _test.go files.
func TestAEC004c_StoreClearFunctionHasExactlyOneNonTestCaller_TheVerb(t *testing.T) {
	var callers []string
	fset := token.NewFileSet()
	for _, top := range []string{"internal", "cmd", "pkg"} {
		root := filepath.Join("..", "..", top)
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if d.Name() == "testdata" || d.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			f, perr := parser.ParseFile(fset, p, nil, parser.SkipObjectResolution)
			if perr != nil {
				return nil
			}
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				switch fn := call.Fun.(type) {
				case *ast.SelectorExpr:
					if fn.Sel.Name == crStoreFunc {
						callers = append(callers, filepath.ToSlash(p))
					}
				case *ast.Ident:
					if fn.Name == crStoreFunc {
						callers = append(callers, filepath.ToSlash(p))
					}
				}
				return true
			})
			return nil
		})
	}
	if len(callers) != 1 {
		t.Fatalf("%s must have exactly one non-test caller, found %d: %v", crStoreFunc, len(callers), callers)
	}
	if filepath.Base(callers[0]) != crVerbFile {
		t.Errorf("the one caller of %s must be %s, found %s", crStoreFunc, crVerbFile, callers[0])
	}
}
