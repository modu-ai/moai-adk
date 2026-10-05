package hook

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
)

// newServedGateTree creates a git-backed project tree whose workflow config
// sets the served-model gate and, when codex is non-empty, the codex audit
// gate. servedGate "" writes no served_model_gate key at all.
func newServedGateTree(t *testing.T, servedGate, codex string) string {
	t.Helper()
	root := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Skipf("git unavailable: %v (%s)", err, out)
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "workflow:\n"
	if servedGate != "" {
		body += "  served_model_gate:\n    enabled: " + servedGate + "\n"
	}
	if codex != "" {
		body += "  audit:\n    gates:\n      codex: " + codex + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return auditreceipt.Canonical(root)
}

// rejectionsOfKind lists the tree's own refusal records of one kind.
func rejectionsOfKind(t *testing.T, root, kind string) []auditreceipt.Rejection {
	t.Helper()
	all, err := auditreceipt.ListRejectionsForTree(root, root)
	if err != nil {
		t.Fatalf("ListRejectionsForTree: %v", err)
	}
	var out []auditreceipt.Rejection
	for _, r := range all {
		if auditreceipt.RejectionKind(r) == kind {
			out = append(out, r)
		}
	}
	return out
}

// servedStopInput is a SubagentStop input for a gate-tree auditor whose
// transcript lives outside the tree.
func servedStopInput(root, agentType, agentID, transcript, message string) *HookInput {
	return &HookInput{CWD: root, SessionID: "s-g", AgentID: agentID, AgentType: agentType,
		AgentTranscriptPath: transcript, LastAssistantMessage: message, HookEventName: string(EventSubagentStop)}
}

func driftTranscript(t *testing.T, agentType, agentID string) string {
	t.Helper()
	return writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), agentID,
		repeatRows(assistantRow("glm-5.3-flash"), 2), map[string]string{"agentType": agentType, "model": "opus"})
}

func okTranscript(t *testing.T, agentType, agentID string) string {
	t.Helper()
	return writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), agentID,
		repeatRows(assistantRow("claude-opus-5-5"), 2), map[string]string{"agentType": agentType, "model": "opus"})
}

func missingTranscript(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "subagents", "agent-gone.jsonl")
}

const passVerdict = "AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none"

// AC-SMA-005 — gate off: drift and unknown warn, nothing is refused or denied;
// the positive control flips only the gate and a served refusal appears.
func TestServedModel_GateOffWarnsOnly(t *testing.T) {
	cases := []struct {
		name       string
		gate       string
		transcript func(t *testing.T) string
		wantRefuse bool
	}{
		{"a_gate_off_drift", "false", func(t *testing.T) string { return driftTranscript(t, "plan-auditor", "g1") }, false},
		{"b_gate_absent_unknown", "", missingTranscript, false},
		{"c_control_gate_on_drift", "true", func(t *testing.T) string { return driftTranscript(t, "plan-auditor", "g1") }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newServedGateTree(t, tc.gate, "")
			out := runServedStop(t, staticConfigProvider{cfg: &config.Config{}}, servedStopInput(root, "plan-auditor", "g1", tc.transcript(t), passVerdict))
			if out == nil || out.Decision != "" {
				t.Fatalf("stop output = %+v, want no decision", out)
			}
			// The missing-transcript case is `unknown` (transcript unreadable)
			// — its warning names the agent and the cause, not a declared
			// model or served set. The drift cases name both.
			if tc.name == "b_gate_absent_unknown" {
				for _, want := range []string{"plan-auditor", "transcript absent or unreadable"} {
					if !strings.Contains(out.SystemMessage, want) {
						t.Fatalf("warning %q does not name %q", out.SystemMessage, want)
					}
				}
			} else {
				for _, want := range []string{"plan-auditor", "opus"} {
					if !strings.Contains(out.SystemMessage, want) {
						t.Fatalf("warning %q does not name %q", out.SystemMessage, want)
					}
				}
				if !strings.Contains(out.SystemMessage, "served") {
					t.Fatalf("warning %q does not name the served set", out.SystemMessage)
				}
			}
			got := rejectionsOfKind(t, root, auditreceipt.KindServed)
			if tc.wantRefuse {
				if len(got) != 1 {
					t.Fatalf("served refusals = %+v, want exactly 1 with the gate on", got)
				}
				return
			}
			if len(got) != 0 {
				t.Fatalf("served refusals = %+v, want none with the gate off", got)
			}
			if ok, reason := denied(runSpawn(t, receiptSpawnInput(root, "manager-develop", "Agent"))); ok {
				t.Fatalf("manager-develop spawn denied with the gate off: %q", reason)
			}
		})
	}
}

// snapshotTree maps every file under root (outside .git) to its bytes.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return out
}

// AC-SMA-006 — gate on: a gate auditor's drift or unknown is refused adoption,
// a non-auditor never is, and nothing the audit produced is touched.
func TestServedModel_GateOnAdoptionRefusal(t *testing.T) {
	cases := []struct {
		name        string
		agentType   string
		transcript  func(t *testing.T) string
		message     string
		wantRefusal bool
		wantSpec    string
		wantCause   string
	}{
		{"a_sync_auditor_drift", "sync-auditor", func(t *testing.T) string { return driftTranscript(t, "sync-auditor", "g6") },
			passVerdict, true, "SPEC-X-001", auditreceipt.CauseServedModelDrift},
		{"b_plan_auditor_unknown", "plan-auditor", missingTranscript,
			"no verdict line here", true, auditreceipt.UnknownSpec, auditreceipt.CauseServedModelUnknown},
		{"c_manager_develop_drift_never_refused", "manager-develop", func(t *testing.T) string { return driftTranscript(t, "manager-develop", "g6") },
			"done", false, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newServedGateTree(t, "true", "")
			if err := os.WriteFile(filepath.Join(root, "audit-report.md"), []byte("auditor-model: opus\n# report\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			specDir := filepath.Join(root, ".moai", "specs", "SPEC-X-001")
			if err := os.MkdirAll(specDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# spec\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			before := snapshotTree(t, root)

			out := runServedStop(t, nil, servedStopInput(root, tc.agentType, "g6", tc.transcript(t), tc.message))
			if out == nil || out.Decision != "" {
				t.Fatalf("stop output = %+v, want no decision (never block)", out)
			}

			after := snapshotTree(t, root)
			for path, body := range before {
				if after[path] != body {
					t.Fatalf("existing file %s changed", path)
				}
			}
			var added []string
			for path := range after {
				if _, ok := before[path]; !ok {
					added = append(added, path)
				}
			}
			sort.Strings(added)
			for _, p := range added {
				if p != ".moai/logs/"+agentModelAuditFileName && !strings.HasPrefix(p, ".moai/state/audit-receipts/rejections/") {
					t.Fatalf("unexpected new file %s (added=%v)", p, added)
				}
			}

			rows, _ := readServedAuditRows(t, root)
			if len(rows) != 1 {
				t.Fatalf("audit rows = %d, want 1", len(rows))
			}
			got := rejectionsOfKind(t, root, auditreceipt.KindServed)
			if !tc.wantRefusal {
				if len(got) != 0 {
					t.Fatalf("non-auditor refusals = %+v, want none", got)
				}
				if rows[0]["verdict"] != ServedVerdictDrift || !strings.Contains(out.SystemMessage, "served-model") {
					t.Fatalf("non-auditor drift must still be recorded and warned: row=%v msg=%q", rows[0], out.SystemMessage)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("served refusals = %+v, want exactly 1", got)
			}
			r := got[0]
			if r.AgentType != tc.agentType || r.SpecID != tc.wantSpec || r.Kind != auditreceipt.KindServed || !strings.Contains(r.Cause, tc.wantCause) {
				t.Fatalf("refusal = %+v, want role %s spec %s kind served cause containing %q", r, tc.agentType, tc.wantSpec, tc.wantCause)
			}
		})
	}
}

// AC-SMA-007 — an outstanding served refusal denies exactly the phase-entry
// spawns, under its own sentinel; mixed with a receipt refusal both appear.
func TestServedModel_RefusalDeniesPhaseEntry(t *testing.T) {
	seedServed := func(t *testing.T, root string) {
		t.Helper()
		r := auditreceipt.Rejection{AgentType: "sync-auditor", SpecID: "SPEC-X-001", Kind: auditreceipt.KindServed,
			Cause: auditreceipt.CauseServedModelDrift + ": expected opus, served [glm-5.3-flash]", TreeRoot: root}
		if err := auditreceipt.WriteRejectionIn(root, &r); err != nil {
			t.Fatalf("WriteRejectionIn: %v", err)
		}
	}
	phase := []string{"manager-develop", "manager-docs", "manager-git"}

	t.Run("a_served_refusal_only", func(t *testing.T) {
		root := newServedGateTree(t, "true", "")
		seedServed(t, root)
		for _, agent := range phase {
			ok, reason := denied(runSpawn(t, receiptSpawnInput(root, agent, "Agent")))
			if !ok || !strings.HasPrefix(reason, servedModelViolation) {
				t.Fatalf("%s: denied=%v reason=%q, want a deny starting with %s", agent, ok, reason, servedModelViolation)
			}
			for _, want := range []string{"sync-auditor", "SPEC-X-001", "glm-5.3-flash"} {
				if !strings.Contains(reason, want) {
					t.Fatalf("%s: reason %q lacks %q", agent, reason, want)
				}
			}
		}
		if ok, reason := denied(runSpawn(t, receiptSpawnInput(root, "Explore", "Agent"))); ok {
			t.Fatalf("Explore denied: %q", reason)
		}
	})

	t.Run("b_served_and_receipt_refusals_together", func(t *testing.T) {
		root := newServedGateTree(t, "true", "required")
		seedServed(t, root)
		legacy := auditreceipt.Rejection{AgentType: "plan-auditor", SpecID: "SPEC-Y-002", Cause: auditreceipt.CauseNoReceiptCited}
		if err := auditreceipt.WriteRejection(root, &legacy); err != nil {
			t.Fatalf("WriteRejection: %v", err)
		}
		ok, reason := denied(runSpawn(t, receiptSpawnInput(root, "manager-develop", "Agent")))
		if !ok {
			t.Fatal("manager-develop not denied with both refusal kinds outstanding")
		}
		iReceipt := strings.Index(reason, auditReceiptViolation)
		iServed := strings.Index(reason, servedModelViolation)
		if iReceipt < 0 || iServed < 0 {
			t.Fatalf("reason %q must carry both sentinels", reason)
		}
		receiptPart, servedPart := reason[iReceipt:iServed], reason[iServed:]
		if iServed < iReceipt {
			servedPart, receiptPart = reason[iServed:iReceipt], reason[iReceipt:]
		}
		if !strings.Contains(receiptPart, "SPEC-Y-002") || strings.Contains(receiptPart, "SPEC-X-001") {
			t.Fatalf("receipt part %q must list only the receipt record", receiptPart)
		}
		if !strings.Contains(servedPart, "SPEC-X-001") || strings.Contains(servedPart, "SPEC-Y-002") {
			t.Fatalf("served part %q must list only the served record", servedPart)
		}
	})

	t.Run("c_control_gate_off", func(t *testing.T) {
		root := newServedGateTree(t, "false", "")
		seedServed(t, root)
		for _, agent := range append(phase, "Explore") {
			if ok, reason := denied(runSpawn(t, receiptSpawnInput(root, agent, "Agent"))); ok {
				t.Fatalf("%s denied with the served gate off: %q", agent, reason)
			}
		}
	})
}

// AC-SMA-008 — clearing is kind-scoped, a kind-less record reads as receipt
// kind, and the two kinds for one role and SPEC never overwrite each other.
func TestServedModel_ClearingIsKindScoped(t *testing.T) {
	seedBoth := func(t *testing.T, root string) {
		t.Helper()
		served := auditreceipt.Rejection{AgentType: "plan-auditor", SpecID: "SPEC-X-001", Kind: auditreceipt.KindServed,
			Cause: auditreceipt.CauseServedModelDrift, TreeRoot: root}
		if err := auditreceipt.WriteRejectionIn(root, &served); err != nil {
			t.Fatalf("WriteRejectionIn: %v", err)
		}
		legacy := auditreceipt.Rejection{AgentType: "plan-auditor", SpecID: "SPEC-X-001", Cause: auditreceipt.CauseNoReceiptCited}
		if err := auditreceipt.WriteRejection(root, &legacy); err != nil {
			t.Fatalf("WriteRejection: %v", err)
		}
		entries, err := os.ReadDir(filepath.Join(auditreceipt.StateDir(root), "rejections"))
		if err != nil || len(entries) != 2 {
			t.Fatalf("setup: want 2 separate record files, got %d (err %v)", len(entries), err)
		}
		if len(rejectionsOfKind(t, root, auditreceipt.KindServed)) != 1 || len(rejectionsOfKind(t, root, auditreceipt.KindReceipt)) != 1 {
			t.Fatal("setup: want one served and one kind-less (receipt) record")
		}
	}

	t.Run("a_served_ok_receipt_unproven_clears_served_only", func(t *testing.T) {
		root := newServedGateTree(t, "true", "required")
		seedBoth(t, root)
		runServedStop(t, nil, servedStopInput(root, "plan-auditor", "k1", okTranscript(t, "plan-auditor", "k1"), passVerdict))
		if got := rejectionsOfKind(t, root, auditreceipt.KindServed); len(got) != 0 {
			t.Fatalf("served refusals after a served ok = %+v, want none", got)
		}
		if got := rejectionsOfKind(t, root, auditreceipt.KindReceipt); len(got) != 1 {
			t.Fatalf("receipt refusals = %+v, want the kind-less record kept", got)
		}
	})

	t.Run("b_receipt_proven_served_drift_clears_receipt_only", func(t *testing.T) {
		root := newServedGateTree(t, "true", "required")
		seedBoth(t, root)
		start := time.Now().UTC()
		seedStart(t, root, "k2", "plan-auditor", start)
		rc := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: start.Add(time.Second)})
		out := runServedStop(t, nil, servedStopInput(root, "plan-auditor", "k2", driftTranscript(t, "plan-auditor", "k2"),
			"AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts="+rc))
		if out.Decision != "" {
			t.Fatalf("receipt-proven PASS decision = %q, want none", out.Decision)
		}
		if got := rejectionsOfKind(t, root, auditreceipt.KindReceipt); len(got) != 0 {
			t.Fatalf("receipt refusals after a proven PASS = %+v, want none", got)
		}
		if got := rejectionsOfKind(t, root, auditreceipt.KindServed); len(got) != 1 {
			t.Fatalf("served refusals = %+v, want the served record kept", got)
		}
	})
}

// AC-SMA-009 — enabling the served gate alone never arms the receipt guard;
// the control tree that declares the codex gate still gets the receipt block.
func TestServedModel_GateDoesNotArmReceiptGuard(t *testing.T) {
	run := func(t *testing.T, root string) (*HookOutput, bool) {
		t.Helper()
		if _, err := NewSubagentStartHandler().Handle(context.Background(),
			&HookInput{CWD: root, AgentID: "r1", AgentType: "plan-auditor", HookEventName: string(EventSubagentStart)}); err != nil {
			t.Fatalf("SubagentStart: %v", err)
		}
		out := runServedStop(t, nil, servedStopInput(root, "plan-auditor", "r1", okTranscript(t, "plan-auditor", "r1"), passVerdict))
		ok, _ := denied(runSpawn(t, receiptSpawnInput(root, "manager-develop", "Agent")))
		return out, ok
	}

	t.Run("served_gate_only", func(t *testing.T) {
		root := newServedGateTree(t, "true", "")
		out, spawnDenied := run(t, root)
		if out.Decision != "" {
			t.Fatalf("stop decision = %q, want none", out.Decision)
		}
		if _, err := os.Stat(filepath.Join(auditreceipt.StateDir(root), "starts")); !os.IsNotExist(err) {
			t.Fatalf("a start marker directory exists (err %v), want none", err)
		}
		if all, _ := auditreceipt.ListRejections(root); len(all) != 0 {
			t.Fatalf("refusals = %+v, want none", all)
		}
		if spawnDenied {
			t.Fatal("manager-develop spawn denied with only the served gate on")
		}
	})

	t.Run("control_codex_gate_required", func(t *testing.T) {
		root := newServedGateTree(t, "true", "required")
		out, _ := run(t, root)
		if out.Decision != "block" {
			t.Fatalf("stop decision = %q, want the receipt guard's block", out.Decision)
		}
		if got := rejectionsOfKind(t, root, auditreceipt.KindReceipt); len(got) != 1 {
			t.Fatalf("receipt refusals = %+v, want 1", got)
		}
	})
}
