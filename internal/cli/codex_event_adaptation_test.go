package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2e — event adaptation (card t1099).
// PreCompact, PostCompact, PermissionRequest, and Interrupt are adapted on the
// Codex path. These are the golden/unit legs of AC-HPR-009, AC-HPR-010, and
// AC-HPR-011; the live legs (TestLiveCodexCompactFires,
// TestLiveCodexPermissionRequestFires, TestLiveCodexInterruptFires) are not run
// in this SPEC (operator decision Q5).

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// adaptationProtocol serves a canned input and serializes the output through
// the REAL hook protocol, so the Codex path maps exactly the bytes a Claude
// run would write.
type adaptationProtocol struct {
	input *hook.HookInput
	real  hook.Protocol
}

func (p *adaptationProtocol) ReadInput(_ io.Reader) (*hook.HookInput, error) { return p.input, nil }
func (p *adaptationProtocol) WriteOutput(w io.Writer, o *hook.HookOutput) error {
	return p.real.WriteOutput(w, o)
}

// capturingRegistry dispatches through real handlers and keeps the last
// handler output, so a test can compare what MoAI produced with what Codex
// was handed after mapping.
type capturingRegistry struct {
	inner hook.Registry
	last  *hook.HookOutput
}

func (r *capturingRegistry) Register(h hook.Handler)                   { r.inner.Register(h) }
func (r *capturingRegistry) Handlers(ev hook.EventType) []hook.Handler { return r.inner.Handlers(ev) }
func (r *capturingRegistry) Dispatch(ctx context.Context, ev hook.EventType, in *hook.HookInput) (*hook.HookOutput, error) {
	out, err := r.inner.Dispatch(ctx, ev, in)
	r.last = out
	return out, err
}

// runCodexSubcommandAt runs one hook subcommand with the given harness flag
// value ("" = the Claude default) against root, with deps pointing at reg.
func runCodexSubcommandAt(t *testing.T, root, subcommand, harness string, input *hook.HookInput, reg hook.Registry) (string, error) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	origDeps := deps
	deps = &Dependencies{HookRegistry: reg, HookProtocol: &adaptationProtocol{input: input, real: hook.NewProtocol()}}
	t.Cleanup(func() { deps = origDeps })

	var sub *cobra.Command
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == subcommand {
			sub = cmd
			break
		}
	}
	if sub == nil {
		t.Fatalf("hook subcommand %q not found", subcommand)
	}
	args := []string{}
	if harness != "" {
		args = []string{"--harness", harness}
	}
	if err := sub.ParseFlags(args); err != nil {
		t.Fatalf("parse %v: %v", args, err)
	}
	t.Cleanup(func() { _ = sub.Flags().Set("harness", "") })
	sub.SetContext(context.Background())

	var runErr error
	stdout := captureStdoutDuring(t, func() { runErr = sub.RunE(sub, []string{}) })
	return stdout, runErr
}

// newMoaiRoot returns a temp project root carrying a .moai directory, which the
// hook write-side resolver requires before it writes anything.
func newMoaiRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestCodexCompactCheckpointRoundTrip is the AC-HPR-009 golden leg: a
// PreCompact payload followed by a PostCompact payload on the Codex path runs
// the same checkpoint-save and checkpoint-restore handlers Claude runs, and
// the restored memo equals the saved one.
func TestCodexCompactCheckpointRoundTrip(t *testing.T) {
	root := newMoaiRoot(t)
	reg := &capturingRegistry{inner: hook.NewRegistry(nil)}
	reg.Register(hook.NewCompactHandler())
	reg.Register(hook.NewPostCompactHandler())

	pre := &hook.HookInput{HookEventName: "PreCompact", SessionID: "codex-compact-1", CWD: root, Trigger: "auto"}
	stdout, err := runCodexSubcommandAt(t, root, "compact", "codex", pre, reg)
	if err != nil {
		t.Fatalf("moai hook compact --harness codex: %v (stdout %q)", err, stdout)
	}
	if !json.Valid([]byte(strings.TrimSpace(stdout))) {
		t.Fatalf("PreCompact Codex stdout is not JSON: %q", stdout)
	}
	saved, err := os.ReadFile(filepath.Join(root, ".moai", "state", "session-memo.md"))
	if err != nil {
		t.Fatalf("PreCompact on the Codex path wrote no memo: %v", err)
	}
	if !strings.Contains(string(saved), "codex-compact-1") {
		t.Fatalf("saved memo does not carry the session: %q", saved)
	}

	post := &hook.HookInput{HookEventName: "PostCompact", SessionID: "codex-compact-1", CWD: root}
	stdout, err = runCodexSubcommandAt(t, root, "post-compact", "codex", post, reg)
	if err != nil {
		t.Fatalf("moai hook post-compact --harness codex: %v (stdout %q)", err, stdout)
	}
	if reg.last == nil {
		t.Fatal("PostCompact handler produced no output")
	}
	restored := strings.TrimPrefix(reg.last.SystemMessage, "[Session Memo - Restored after context compaction]\n\n")
	if strings.TrimSpace(restored) != strings.TrimSpace(string(saved)) {
		t.Fatalf("restored memo differs from the saved memo:\nsaved:\n%s\nrestored:\n%s", saved, restored)
	}

	// PostCompact has no measured Codex delivery channel for the restored
	// text, so the adapter must say so rather than drop it silently.
	var recorded bool
	for _, d := range sinkRecords(t, root) {
		if d.Event == hook.EventPostCompact && d.Key == "systemMessage" && d.ContentLength == len(reg.last.SystemMessage) {
			recorded = true
		}
	}
	if !recorded {
		t.Fatalf("the undelivered PostCompact memo left no discard record; sink = %+v", sinkRecords(t, root))
	}
}

// TestCodexPermissionRequestDenyPreserved is the AC-HPR-010 golden leg: the
// same permission logic Claude runs denies a tool input carrying the
// updated-input marker, and Codex receives that deny with a non-empty reason.
// A request the logic has no opinion on stays without an opinion — Codex, like
// Claude, then applies its own approval flow (REQ-HPR-011, same logic).
func TestCodexPermissionRequestDenyPreserved(t *testing.T) {
	newReg := func() *capturingRegistry {
		r := &capturingRegistry{inner: hook.NewRegistry(nil)}
		r.Register(hook.NewPermissionRequestHandler())
		return r
	}

	t.Run("marker denies with a reason", func(t *testing.T) {
		root := newMoaiRoot(t)
		in := &hook.HookInput{
			HookEventName: "PermissionRequest", SessionID: "codex-perm-1", ToolName: "Bash",
			ToolInput: json.RawMessage(`{"command":"ls","__updated_input_marker__":true}`),
		}
		reg := newReg()
		stdout, err := runCodexSubcommandAt(t, root, "permission-request", "codex", in, reg)
		if err != nil {
			t.Fatalf("moai hook permission-request --harness codex: %v", err)
		}
		if reg.last == nil || reg.last.HookSpecificOutput == nil || reg.last.HookSpecificOutput.Decision == nil ||
			reg.last.HookSpecificOutput.Decision.Behavior != "deny" {
			t.Fatalf("premise: the Claude permission logic did not deny: %+v", reg.last)
		}
		assertCodexDeny(t, hook.EventPermissionRequest, stdout)
		if !strings.Contains(stdout, "updatedInput re-verification") {
			t.Fatalf("Codex deny does not carry the handler's reason: %s", stdout)
		}
	})

	t.Run("no opinion stays no opinion", func(t *testing.T) {
		root := newMoaiRoot(t)
		in := &hook.HookInput{
			HookEventName: "PermissionRequest", SessionID: "codex-perm-2", ToolName: "Bash",
			ToolInput: json.RawMessage(`{"command":"ls"}`),
		}
		stdout, err := runCodexSubcommandAt(t, root, "permission-request", "codex", in, newReg())
		if err != nil {
			t.Fatalf("moai hook permission-request --harness codex: %v", err)
		}
		if strings.Contains(stdout, `"deny"`) || strings.Contains(stdout, `"allow"`) {
			t.Fatalf("a request the logic has no opinion on gained a decision on Codex: %s", stdout)
		}
	})

	t.Run("a handler fault on the real subcommand is fail-closed", func(t *testing.T) {
		root := newMoaiRoot(t)
		in := &hook.HookInput{HookEventName: "PermissionRequest", SessionID: "codex-perm-3", ToolName: "Bash"}
		stdout, err := runCodexSubcommandAt(t, root, "permission-request", "codex", in,
			&faultRegistry{err: hook.ErrHookTimeout})
		if err != nil {
			t.Fatalf("a fault must be answered on stdout, not as an error: %v", err)
		}
		assertCodexDeny(t, hook.EventPermissionRequest, stdout)
	})
}

// TestCodexInterruptRecordsCancellation is the AC-HPR-011 unit leg: a Codex
// Interrupt payload fed to `moai hook interrupt --harness codex` writes a
// cancellation record bound to the session and its goal run, and the armed
// goal reads `cancelled`. internal/hook carries no Interrupt constant.
func TestCodexInterruptRecordsCancellation(t *testing.T) {
	t.Run("records the cancellation", func(t *testing.T) {
		root := newMoaiRoot(t)
		g := goal.NewGoal("codex-int-1", "tests pass", []goal.Condition{{Type: goal.ConditionMechanical, Cmd: "false"}})
		g.CreatedAt = "2026-09-26T00:00:00Z"
		if err := goal.SaveGoal(root, g); err != nil {
			t.Fatal(err)
		}
		other := goal.NewGoal("codex-int-other", "other", nil)
		if err := goal.SaveGoal(root, other); err != nil {
			t.Fatal(err)
		}

		in := &hook.HookInput{HookEventName: "Interrupt", SessionID: "codex-int-1", TranscriptPath: "/tmp/rollout-1.jsonl"}
		stdout, err := runCodexSubcommandAt(t, root, "interrupt", "codex", in, &faultRegistry{})
		if err != nil {
			t.Fatalf("moai hook interrupt --harness codex: %v (stdout %q)", err, stdout)
		}

		recs, err := readInterruptRecords(root, "codex-int-1")
		if err != nil || len(recs) != 1 {
			t.Fatalf("cancellation records = %v (err %v), want exactly one", recs, err)
		}
		rec := recs[0]
		if rec.SessionID != "codex-int-1" || rec.TranscriptPath != "/tmp/rollout-1.jsonl" ||
			rec.GoalCreatedAt != g.CreatedAt || rec.GoalStatusBefore != goal.StatusArmed || rec.RecordedAt == "" {
			t.Fatalf("record is not bound to the session and run: %+v", rec)
		}

		got, err := goal.LoadGoal(root, "codex-int-1")
		if err != nil || got == nil || got.Status != goal.StatusCancelled {
			t.Fatalf("goal after interrupt = %+v (err %v), want status cancelled", got, err)
		}
		// Another session's goal is untouched (acceptance.md §E).
		if o, _ := goal.LoadGoal(root, "codex-int-other"); o == nil || o.Status != goal.StatusArmed {
			t.Fatalf("another session's goal changed: %+v", o)
		}
	})

	t.Run("records the cancellation when no goal is armed", func(t *testing.T) {
		root := newMoaiRoot(t)
		in := &hook.HookInput{HookEventName: "Interrupt", SessionID: "codex-int-2"}
		if _, err := runCodexSubcommandAt(t, root, "interrupt", "codex", in, &faultRegistry{}); err != nil {
			t.Fatalf("moai hook interrupt --harness codex: %v", err)
		}
		recs, err := readInterruptRecords(root, "codex-int-2")
		if err != nil || len(recs) != 1 || recs[0].GoalStatusBefore != "" {
			t.Fatalf("records = %+v (err %v), want one record with no goal status", recs, err)
		}
		if g, _ := goal.LoadGoal(root, "codex-int-2"); g != nil {
			t.Fatalf("an interrupt without a goal created goal state: %+v", g)
		}
	})

	t.Run("refused without --harness codex", func(t *testing.T) {
		root := newMoaiRoot(t)
		in := &hook.HookInput{HookEventName: "Interrupt", SessionID: "codex-int-3"}
		if _, err := runCodexSubcommandAt(t, root, "interrupt", "", in, &faultRegistry{}); err == nil {
			t.Fatal("moai hook interrupt must refuse the Claude harness: Interrupt has no Claude-side counterpart")
		}
		if recs, _ := readInterruptRecords(root, "codex-int-3"); len(recs) != 0 {
			t.Fatalf("a refused interrupt wrote records: %+v", recs)
		}
	})

	t.Run("internal/hook carries no Interrupt constant", func(t *testing.T) {
		out, err := exec.Command("grep", "-rn", "EventInterrupt", filepath.Join("..", "hook")).CombinedOutput()
		if err == nil || len(strings.TrimSpace(string(out))) != 0 {
			t.Fatalf("grep -rn EventInterrupt internal/hook printed %q (err %v); want nothing", out, err)
		}
	})
}
