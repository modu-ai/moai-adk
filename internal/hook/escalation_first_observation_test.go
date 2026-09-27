package hook

// AC-AE-025 (SPEC-AUTONOMY-ESCALATION-001, REQ-AE-023): on a card never armed,
// the first hook event — a PreToolUse Write outside ownership.write — verifies
// the sole candidate contract in process on that same call. Signed-invalid:
// not-armed plus a warning with the reason codes, no ownership-move and no
// detection-disarmed record, and a later over-budget operation still trips
// against budget_default. Signed-valid: that call arms the card, appends the
// armed entry, writes the state file and one ownership-move record, and starts
// no subprocess.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// escalationHookFixture builds a contract-mode worktree and handlers bound to
// its configuration, with the store isolated under a per-test MOAI_HOME.
func escalationHookFixture(t *testing.T, autonomyExtra string) *escalationtest.Worktree {
	t.Helper()
	t.Setenv(config.EnvHome, t.TempDir())
	w := escalationtest.NewWorktree(t, "t9001")
	w.WriteContractMode(autonomyExtra)
	w.Write("internal/fixture/y.go", "package fixture\n")
	t.Setenv(config.EnvClaudeProjectDir, w.Root)
	return w
}

func escalationHandlers(t *testing.T, w *escalationtest.Worktree) (Handler, Handler) {
	t.Helper()
	return escalationHandlersMode(t, w, "")
}

// escalationHandlersMode builds handlers from the worktree's configuration,
// with the autonomy mode overridden when mode is non-empty.
func escalationHandlersMode(t *testing.T, w *escalationtest.Worktree, mode string) (Handler, Handler) {
	t.Helper()
	cfg, err := config.NewLoader().Load(w.Path(".moai"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if mode != "" {
		cfg.Workflow.Autonomy.Mode = mode
	}
	provider := staticConfigProvider{cfg: cfg}
	pre := NewPreToolHandler(provider, DefaultSecurityPolicy())
	post := WithEscalationConfig(NewPostToolHandlerWithMxValidatorAndTimeout(nil, nil, "", 0), provider)
	return pre, post
}

func escalationSettings(t *testing.T, w *escalationtest.Worktree) config.AutonomySettings {
	t.Helper()
	cfg, err := config.NewLoader().Load(w.Path(".moai"))
	if err != nil {
		t.Fatal(err)
	}
	return config.ResolveAutonomy(cfg.Workflow)
}

func escalationInput(w *escalationtest.Worktree, tool string, in map[string]string) *HookInput {
	raw, _ := json.Marshal(in)
	return &HookInput{SessionID: "s", CWD: w.Path("internal"), PermissionMode: "default", ToolName: tool, ToolInput: raw}
}

// installGitTrap puts a fake git first on PATH that appends one line per
// invocation to a marker file; gitTrapCount reads the count.
func installGitTrap(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	marker := filepath.Join(dir, "git-invocations")
	script := "#!/bin/sh\necho x >> '" + marker + "'\nexit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return marker
}

func gitTrapCount(marker string) int {
	data, err := os.ReadFile(marker)
	if err != nil {
		return 0
	}
	return strings.Count(string(data), "\n")
}

func escalationRecordNames(t *testing.T, w *escalationtest.Worktree) []string {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join(escalation.RecordDir(w.Root, w.Card), "*.md"))
	var out []string
	for _, p := range paths {
		out = append(out, filepath.Base(p))
	}
	return out
}

func escalationCardLog(t *testing.T, w *escalationtest.Worktree) escalation.CardLog {
	t.Helper()
	files, err := escalation.CardFilesFor(w.Root, w.Card)
	if err != nil {
		t.Fatal(err)
	}
	lg, err := escalation.ReadCardLog(files.Log)
	if err != nil {
		t.Fatal(err)
	}
	return lg
}

func TestFirstObservationVerifiedAtPreToolUse(t *testing.T) {
	t.Run("signed-invalid", func(t *testing.T) {
		w := escalationHookFixture(t, "escalation:\n  budget_default:\n    operations: 1")
		w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
		w.Replace(".moai/specs/SPEC-A-001/contract.yaml", "verdict: PASS", "verdict: FAIL")
		pre, post := escalationHandlers(t, w)

		out, err := pre.Handle(context.Background(), escalationInput(w, "Write",
			map[string]string{"file_path": w.Path("internal/bar/x.go"), "content": "package bar\n"}))
		if err != nil || out.HookSpecificOutput != nil && out.HookSpecificOutput.PermissionDecision != "" {
			t.Fatalf("PreToolUse output = %+v, %v; want no decision", out, err)
		}
		if got := escalationRecordNames(t, w); len(got) != 0 {
			t.Fatalf("records after first observation = %v, want none", got)
		}
		lg := escalationCardLog(t, w)
		if lg.Armed() {
			t.Fatal("signed-invalid contract armed")
		}
		var notArmed, warnings int
		warnOK := false
		for _, e := range lg.Entries {
			switch e.Kind {
			case escalation.LineNotArmed:
				notArmed++
			case escalation.LineWarning:
				warnings++
				warnOK = warnOK || slices.Contains(e.Reasons, "plan_audit_not_passing")
			}
		}
		if notArmed != 1 || warnings != 1 || !warnOK {
			t.Errorf("log = %+v, want one not-armed and one warning carrying plan_audit_not_passing", lg.Entries)
		}

		// Two operations against budget_default.operations = 1.
		for i := 0; i < 2; i++ {
			in := escalationInput(w, "Bash", map[string]string{"command": "ls"})
			in.ToolResponse = json.RawMessage(`{"stdout":"","stderr":""}`)
			if _, err := post.Handle(context.Background(), in); err != nil {
				t.Fatal(err)
			}
		}
		names := escalationRecordNames(t, w)
		if len(names) != 1 || !strings.HasPrefix(names[0], escalation.ClassBudgetExceeded+"-") {
			t.Fatalf("records = %v, want one budget-exceeded", names)
		}
		data, _ := os.ReadFile(filepath.Join(escalation.RecordDir(w.Root, w.Card), names[0]))
		r, err := escalation.ParseRecord(data)
		if err != nil || r.ContractRef != "config:workflow.autonomy.escalation.budget_default.operations" ||
			!strings.Contains(string(data), "operations") {
			t.Errorf("budget record = %+v (%v)", r, err)
		}
	})

	t.Run("signed-valid", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("the subprocess trap is a POSIX shell script")
		}
		w := escalationHookFixture(t, "")
		w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
		marker := installGitTrap(t)
		write := func() *HookInput {
			return escalationInput(w, "Write",
				map[string]string{"file_path": w.Path("internal/bar/x.go"), "content": "package bar\n"})
		}

		// Baseline: the same call with the detector inert (guided). Existing
		// PreToolUse steps may start git themselves; the detector must add
		// no invocation on top of them.
		guidedPre, _ := escalationHandlersMode(t, w, "guided")
		if _, err := guidedPre.Handle(context.Background(), write()); err != nil {
			t.Fatal(err)
		}
		baseline := gitTrapCount(marker)
		t.Logf("guided-mode baseline: %d git invocation(s) by existing PreToolUse steps", baseline)
		if got := escalationRecordNames(t, w); len(got) != 0 {
			t.Fatalf("guided call wrote records: %v", got)
		}

		pre, _ := escalationHandlers(t, w)
		out, err := pre.Handle(context.Background(), write())
		if err != nil || out.HookSpecificOutput != nil && out.HookSpecificOutput.PermissionDecision != "" {
			t.Fatalf("PreToolUse output = %+v, %v; want no decision", out, err)
		}
		if added := gitTrapCount(marker) - 2*baseline; added != 0 {
			t.Errorf("the contract-mode call started %d git processes beyond the guided baseline of %d", added, baseline)
		}
		// A direct detector call on the now-armed card starts no process at all.
		before := gitTrapCount(marker)
		escalation.Observe(escalationSettings(t, w), escalation.Event{Hook: escalation.HookPreToolUse,
			CWD: w.Path("internal"), ToolName: "Write", FilePath: w.Path("internal/fixture/y.go")})
		if gitTrapCount(marker) != before {
			t.Error("escalation.Observe started git")
		}
		lg := escalationCardLog(t, w)
		if !lg.Armed() {
			t.Fatalf("card not armed on first observation: %+v", lg.Entries)
		}
		files, _ := escalation.CardFilesFor(w.Root, w.Card)
		if _, ok, err := escalation.ReadCardState(files.State); !ok || err != nil {
			t.Errorf("card state file not written: %v", err)
		}
		names := escalationRecordNames(t, w)
		if len(names) != 1 || !strings.HasPrefix(names[0], escalation.ClassOwnershipMove+"-") {
			t.Errorf("records = %v, want exactly one ownership-move", names)
		}
	})
}
