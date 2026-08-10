package hook

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/routing"
)

// newRoutingSeamRoot builds a temp project root with BOTH harness observation
// gates open, and points CLAUDE_PROJECT_DIR at it.
//
// Every gate-open fixture sets hook.opt_in.enabled explicitly. Relying on the
// default would silently exercise the dormant path — that gate is fail-CLOSED,
// so an omitted key means OFF and every assertion below would pass vacuously.
func newRoutingSeamRoot(t *testing.T) string {
	t.Helper()
	return newRoutingSeamRootWith(t,
		"hook:\n  opt_in:\n    enabled: true\n",
		"learning:\n  enabled: true\n",
	)
}

func newRoutingSeamRootWith(t *testing.T, systemYAML, harnessYAML string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "system.yaml"), []byte(systemYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "harness.yaml"), []byte(harnessYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	return root
}

func seamStore(root string) *routing.Store {
	return routing.NewStore(filepath.Join(root, ".moai", "state"))
}

func seamPending(t *testing.T, root, sessionID string) routing.PendingRow {
	t.Helper()
	p, ok, err := seamStore(root).LoadPending(sessionID)
	if err != nil {
		t.Fatalf("load pending: %v", err)
	}
	if !ok {
		t.Fatalf("expected a pending row for session %q, found none", sessionID)
	}
	return p
}

func seamLedgerRows(t *testing.T, root string) []routing.Row {
	t.Helper()
	path := filepath.Join(root, ".moai", "state", routing.LedgerFileName)
	rows, _, err := routing.NewReader(path).Read(routing.Filter{})
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	return rows
}

// TestRoutingSeam_UserPromptSubmit_CreatesPending — AC-HLE-006 (REQ-HLE-001/002):
// a prompt creates a pending row mechanically, with the digest and class derived
// through the existing functions rather than supplied by an instruction.
func TestRoutingSeam_UserPromptSubmit_CreatesPending(t *testing.T) {
	root := newRoutingSeamRoot(t)
	const prompt = "add OAuth2 support to the session layer"

	RoutingSeamUserPromptSubmit(&HookInput{SessionID: "seam-a1", Prompt: prompt, CWD: root})

	p := seamPending(t, root, "seam-a1")
	if p.RequestDigest != routing.RequestDigest(prompt) {
		t.Errorf("request_digest = %q, want %q", p.RequestDigest, routing.RequestDigest(prompt))
	}
	if p.RequestClass != routing.ClassifyRequest(prompt) {
		t.Errorf("request_class = %q, want %q", p.RequestClass, routing.ClassifyRequest(prompt))
	}
	if p.SessionID != "seam-a1" {
		t.Errorf("session_id = %q, want seam-a1", p.SessionID)
	}
}

// TestRoutingSeam_NoRawPromptPersisted — AC-HLE-007 (REQ-HLE-016): the canary
// literal must not reach disk anywhere under the state dir.
func TestRoutingSeam_NoRawPromptPersisted(t *testing.T) {
	root := newRoutingSeamRoot(t)
	const canary = "ZZQX-CANARY-PROMPT"

	RoutingSeamUserPromptSubmit(&HookInput{SessionID: "seam-priv", Prompt: canary + " please implement it", CWD: root})

	stateDir := filepath.Join(root, ".moai", "state")
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("read state dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected the seam to have written something; an empty state dir makes this assertion vacuous")
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(stateDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte(canary)) {
			t.Fatalf("verbatim prompt text leaked into %s", e.Name())
		}
	}
}

// TestRoutingSeam_MultiTurnNoReroute — AC-HLE-008 (REQ-HLE-003): the seam runs
// once per prompt, and a multi-turn pipeline spans many prompts. Repeated
// prompts must not close the row.
func TestRoutingSeam_MultiTurnNoReroute(t *testing.T) {
	root := newRoutingSeamRoot(t)
	in := &HookInput{SessionID: "seam-multi", Prompt: "keep going", CWD: root}

	RoutingSeamUserPromptSubmit(in)
	RoutingSeamUserPromptSubmit(in)
	RoutingSeamUserPromptSubmit(in)
	RoutingSeamUserPromptSubmit(in)

	if rows := seamLedgerRows(t, root); len(rows) != 0 {
		t.Fatalf("got %d ledger rows, want 0 — a per-prompt reroute is the exact failure this SPEC repairs: %+v", len(rows), rows)
	}
	seamPending(t, root, "seam-multi") // still open
}

// TestRoutingSeam_LiteralSubcommandFirstWriterWins — AC-HLE-009 (REQ-HLE-006,
// plan.md §E D1): the subcommand is taken from a literal `/moai <sub>` prefix
// only, and only by the first writer.
func TestRoutingSeam_LiteralSubcommandFirstWriterWins(t *testing.T) {
	root := newRoutingSeamRoot(t)

	for _, prompt := range []string{
		"/moai plan add auth",
		"please run the auth work",
		"/moai run SPEC-X",
	} {
		RoutingSeamUserPromptSubmit(&HookInput{SessionID: "seam-sub", Prompt: prompt, CWD: root})
	}
	if got := seamPending(t, root, "seam-sub").MatchedSubcommand; got != "plan" {
		t.Errorf("matched_subcommand = %q, want %q — the natural-language prompt must set nothing and the second literal must not relabel", got, "plan")
	}

	// A fresh session that never carries a literal prefix stays unlabelled: the
	// seam must not guess a subcommand from natural-language text.
	RoutingSeamUserPromptSubmit(&HookInput{SessionID: "seam-nl", Prompt: "please plan the auth work", CWD: root})
	if got := seamPending(t, root, "seam-nl").MatchedSubcommand; got != "" {
		t.Errorf("matched_subcommand = %q, want empty — natural language must not be guessed into a subcommand", got)
	}
}

// TestRoutingSeam_SubagentStop_AgentTypeVerbatim — AC-HLE-010 (REQ-HLE-007):
// agent_type is recorded verbatim, including a named-spawn value that is not a
// retained-catalog agent name, and the derived subject field is never used.
func TestRoutingSeam_SubagentStop_AgentTypeVerbatim(t *testing.T) {
	root := newRoutingSeamRoot(t)
	RoutingSeamUserPromptSubmit(&HookInput{SessionID: "seam-b1", Prompt: "/moai run SPEC-X", CWD: root})

	RoutingSeamSubagentStop(&HookInput{SessionID: "seam-b1", AgentType: "manager-develop", CWD: root})
	// The observed named-spawn shape: subagent_type plan-auditor spawned with
	// name audit-hle reports agent_type "audit-hle".
	RoutingSeamSubagentStop(&HookInput{SessionID: "seam-b1", AgentType: "audit-hle", AgentName: "audit-hle", CWD: root})

	dels := seamPending(t, root, "seam-b1").Delegations
	if len(dels) != 2 {
		t.Fatalf("got %d delegations, want 2", len(dels))
	}
	if dels[0].Agent != "manager-develop" {
		t.Errorf("delegations[0].agent = %q, want manager-develop", dels[0].Agent)
	}
	if dels[1].Agent != "audit-hle" {
		t.Errorf("delegations[1].agent = %q, want audit-hle stored verbatim (no normalization onto a catalog name)", dels[1].Agent)
	}
}

// TestRoutingSeam_AbsentAgentTypeMarker — AC-HLE-011 (REQ-HLE-008): an absent
// identity is recorded under a distinguishable marker so it stays countable.
func TestRoutingSeam_AbsentAgentTypeMarker(t *testing.T) {
	root := newRoutingSeamRoot(t)
	RoutingSeamUserPromptSubmit(&HookInput{SessionID: "seam-b2", Prompt: "/moai run SPEC-X", CWD: root})

	RoutingSeamSubagentStop(&HookInput{SessionID: "seam-b2", CWD: root}) // no agent_type at all

	dels := seamPending(t, root, "seam-b2").Delegations
	if len(dels) != 1 {
		t.Fatalf("got %d delegations, want 1 — an unattributed delegation must still be appended", len(dels))
	}
	if dels[0].Agent == "" {
		t.Fatal("agent is empty; an absent identity must be distinguishable from an attributed one")
	}
	if dels[0].Agent != routing.AgentUnattributed {
		t.Errorf("agent = %q, want the declared marker %q", dels[0].Agent, routing.AgentUnattributed)
	}
}

// TestRoutingSeam_UnknownOutcomeNotSuccess — AC-HLE-012 (REQ-HLE-009): with no
// observable outcome, the entry records the unknown marker and a null blocker.
// Absence of a failure signal is not evidence of success (plan.md §G AP-3).
func TestRoutingSeam_UnknownOutcomeNotSuccess(t *testing.T) {
	root := newRoutingSeamRoot(t)
	RoutingSeamUserPromptSubmit(&HookInput{SessionID: "seam-b3", Prompt: "/moai run SPEC-X", CWD: root})

	RoutingSeamSubagentStop(&HookInput{SessionID: "seam-b3", AgentType: "manager-docs", CWD: root})

	dels := seamPending(t, root, "seam-b3").Delegations
	if len(dels) != 1 {
		t.Fatalf("got %d delegations, want 1", len(dels))
	}
	if dels[0].Outcome != routing.OutcomeUnknownDelegation {
		t.Errorf("outcome = %q, want %q", dels[0].Outcome, routing.OutcomeUnknownDelegation)
	}
	if dels[0].Outcome == "success" {
		t.Error("an unobserved outcome must never be recorded as success")
	}
	if dels[0].Blocker != nil {
		t.Errorf("blocker = %v, want null", *dels[0].Blocker)
	}

	// Belt and braces: the literal must not appear anywhere in the entry.
	data, err := os.ReadFile(filepath.Join(root, ".moai", "state", "routing-pending-seam-b3.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"outcome":"success"`) {
		t.Fatalf("the pending row claims success for an unobserved delegation: %s", data)
	}
}

// TestRoutingSeam_FailOpen — AC-HLE-015 (REQ-HLE-012): every seam returns
// without error and writes a diagnostic when the state dir cannot be written or
// the pending file is malformed. The seams return nothing, so "returns without
// error" is a structural property; what this test proves is that they neither
// panic nor abort the handler.
func TestRoutingSeam_FailOpen(t *testing.T) {
	t.Run("unwritable state dir", func(t *testing.T) {
		root := newRoutingSeamRoot(t)
		// Make .moai/state a regular FILE so every write beneath it fails.
		if err := os.WriteFile(filepath.Join(root, ".moai", "state"), []byte("not a dir"), 0o644); err != nil {
			t.Fatal(err)
		}
		var sink bytes.Buffer
		SetRoutingSeamSinkForTesting(&sink)
		t.Cleanup(func() { SetRoutingSeamSinkForTesting(nil) })

		in := &HookInput{SessionID: "seam-fo1", Prompt: "/moai run SPEC-X", AgentType: "manager-develop", CWD: root}
		RoutingSeamUserPromptSubmit(in)
		RoutingSeamSubagentStop(in)
		RoutingSeamStopEvidence(root, in.SessionID)

		if sink.Len() == 0 {
			t.Fatal("a failing seam must surface a diagnostic to the sink")
		}
	})

	t.Run("malformed pending file", func(t *testing.T) {
		root := newRoutingSeamRoot(t)
		stateDir := filepath.Join(root, ".moai", "state")
		if err := os.MkdirAll(stateDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(stateDir, "routing-pending-seam-fo2.json"), []byte("{not json"), 0o644); err != nil {
			t.Fatal(err)
		}
		var sink bytes.Buffer
		SetRoutingSeamSinkForTesting(&sink)
		t.Cleanup(func() { SetRoutingSeamSinkForTesting(nil) })

		in := &HookInput{SessionID: "seam-fo2", Prompt: "/moai run SPEC-X", AgentType: "manager-develop", CWD: root}
		RoutingSeamUserPromptSubmit(in) // drops the malformed file and starts fresh
		RoutingSeamSubagentStop(in)
		RoutingSeamStopEvidence(root, in.SessionID)

		// The handler survived; a usable row exists again.
		seamPending(t, root, "seam-fo2")
	})
}

// TestRoutingSeam_GatedOffWritesNothing — AC-HLE-016 (REQ-HLE-013): with either
// gate closed, no ledger file and no pending file appear.
func TestRoutingSeam_GatedOffWritesNothing(t *testing.T) {
	cases := []struct {
		name        string
		systemYAML  string
		harnessYAML string
	}{
		{"hook opt-in closed", "hook:\n  opt_in:\n    enabled: false\n", "learning:\n  enabled: true\n"},
		{"hook opt-in absent (fail-closed default)", "other: 1\n", "learning:\n  enabled: true\n"},
		{"learning closed", "hook:\n  opt_in:\n    enabled: true\n", "learning:\n  enabled: false\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newRoutingSeamRootWith(t, tc.systemYAML, tc.harnessYAML)
			in := &HookInput{SessionID: "seam-gated", Prompt: "/moai run SPEC-X", AgentType: "manager-develop", CWD: root}

			RoutingSeamUserPromptSubmit(in)
			RoutingSeamSubagentStop(in)
			RoutingSeamStopEvidence(root, in.SessionID)

			stateDir := filepath.Join(root, ".moai", "state")
			entries, err := os.ReadDir(stateDir)
			if err == nil && len(entries) > 0 {
				names := make([]string, 0, len(entries))
				for _, e := range entries {
					names = append(names, e.Name())
				}
				t.Fatalf("a closed gate must write nothing, found: %v", names)
			}
		})
	}
}
