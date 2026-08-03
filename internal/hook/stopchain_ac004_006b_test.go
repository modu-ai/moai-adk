package hook

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestAC006b_LifecycleDormantAtFullyAutonomous (A11 / AC-STOPCHAIN-TRIM-006b):
// at MOAI_AUTONOMY_TIER=fully-autonomous, each of the three subagent lifecycle
// hooks (SubagentStop, TeammateIdle, TaskCompleted) is DORMANT — observe-only.
// The hook (a) exits 0 (non-blocking), (b) does NOT invoke the moai binary
// (zero cold-start), and (c) writes an audit-log entry (observe-only is NOT
// silent). At semi-auto the dormant guard does NOT fire (regression guard — the
// active blocking/reject behavior must return).
func TestAC006b_LifecycleDormantAtFullyAutonomous(t *testing.T) {
	requireBash(t)
	repoRoot := filepath.Join(mustGetwd(t), "..", "..")
	wrappers := []struct {
		name    string
		rel     string
		extra   []string // extra args (handle-agent-hook.sh takes the action arg)
	}{
		{"SubagentStop (develop-completion)", ".claude/hooks/moai/handle-agent-hook.sh", []string{"develop-completion"}},
		{"TeammateIdle", ".claude/hooks/moai/handle-teammate-idle.sh", nil},
		{"TaskCompleted", ".claude/hooks/moai/handle-task-completed.sh", nil},
	}

	for _, w := range wrappers {
		t.Run(w.name+"/fully-autonomous", func(t *testing.T) {
			tmp := t.TempDir()
			stubBin := filepath.Join(tmp, "bin")
			_ = os.MkdirAll(stubBin, 0o755)
			homeDir := filepath.Join(tmp, "home")
			_ = os.MkdirAll(homeDir, 0o755)
			counter := filepath.Join(tmp, "moai.count")
			writeCountingStub(t, stubBin, "moai", counter)

			// Point CLAUDE_PROJECT_DIR at the tmp dir so the audit log lands there.
			projectDir := tmp
			args := append([]string{filepath.Join(repoRoot, w.rel)}, w.extra...)
			cmd := exec.Command("bash", args...)
			cmd.Stdin = strings.NewReader(`{"session_id":"sess-006b","agent_id":"a1"}`)
			cmd.Dir = projectDir
			cmd.Env = []string{
				"PATH=" + stubBin + string(os.PathListSeparator) + os.Getenv("PATH"),
				"HOME=" + homeDir,
				"CLAUDE_PROJECT_DIR=" + projectDir,
				config.EnvAutonomyTier + "=" + config.AutonomyTierFullyAutonomous,
			}
			out, err := cmd.CombinedOutput()
			code := 0
			if err != nil {
				if ex, ok := err.(*exec.ExitError); ok {
					code = ex.ExitCode()
				} else {
					t.Fatalf("run %s: %v; out=%s", w.name, err, out)
				}
			}
			if code != 0 {
				t.Fatalf("AC-006b: %s at fully-autonomous exited %d (must exit 0 — dormant/non-blocking). output=%s", w.name, code, out)
			}
			if n := readCounter(t, counter); n != 0 {
				t.Fatalf("AC-006b: %s at fully-autonomous invoked the moai binary %d time(s) — dormant must pay ZERO cold-starts. guard missing or misplaced?", w.name, n)
			}
			auditLog := filepath.Join(projectDir, ".moai", "logs", "lifecycle-dormant.log")
			if b, err := os.ReadFile(auditLog); err != nil || len(b) == 0 {
				t.Fatalf("AC-006b: %s at fully-autonomous did NOT write an audit-log entry (%s) — observe-only is NOT silent. err=%v", w.name, auditLog, err)
			}
		})

		t.Run(w.name+"/semi-auto-regression", func(t *testing.T) {
			// At semi-auto the dormant guard must NOT fire → moai binary IS invoked.
			tmp := t.TempDir()
			stubBin := filepath.Join(tmp, "bin")
			_ = os.MkdirAll(stubBin, 0o755)
			homeDir := filepath.Join(tmp, "home")
			_ = os.MkdirAll(homeDir, 0o755)
			counter := filepath.Join(tmp, "moai.count")
			writeCountingStub(t, stubBin, "moai", counter)
			projectDir := tmp
			args := append([]string{filepath.Join(repoRoot, w.rel)}, w.extra...)
			cmd := exec.Command("bash", args...)
			cmd.Stdin = strings.NewReader(`{"session_id":"sess-006b-semi","agent_id":"a1"}`)
			cmd.Dir = projectDir
			cmd.Env = []string{
				"PATH=" + stubBin + string(os.PathListSeparator) + os.Getenv("PATH"),
				"HOME=" + homeDir,
				"CLAUDE_PROJECT_DIR=" + projectDir,
				config.EnvAutonomyTier + "=" + config.AutonomyTierSemiAuto,
			}
			_, _ = cmd.CombinedOutput()
			if n := readCounter(t, counter); n != 1 {
				t.Fatalf("AC-006b regression: %s at semi-auto invoked moai %d time(s); want exactly 1 (dormant guard must NOT fire at semi-auto — active behavior must return).", w.name, n)
			}
		})
	}
}

// TestAC004_SyncGateAdvisoryAtFullyAutonomous (A11 / AC-STOPCHAIN-TRIM-004):
// at MOAI_AUTONOMY_TIER=fully-autonomous, a FAILING sync-gate decision emits
// advisory output only (systemMessage), with NO decision:block. We force a
// build failure by pointing the gate at a fixture whose HEAD is a sync commit
// but whose go.mod is broken, then assert the gate's stdout JSON lacks
// "decision":"block" at fully-autonomous and contains it at semi-auto.
func TestAC004_SyncGateAdvisoryAtFullyAutonomous(t *testing.T) {
	requireBash(t)
	requireGit(t)
	repoRoot := filepath.Join(mustGetwd(t), "..", "..")
	script := filepath.Join(repoRoot, ".claude", "hooks", "moai", "sync-phase-quality-gate.sh")

	// Fixture: a go project whose `go build ./...` fails (syntax error).
	runGate := func(t *testing.T, tier string, setTier bool) (stdout string, exitCode int) {
		t.Helper()
		fixture := t.TempDir()
		initGoFixtureRepo(t, fixture, "docs(SPEC-X): sync-phase artifacts")
		// Break the build so C2 (go build) fails.
		if err := os.WriteFile(filepath.Join(fixture, "main.go"), []byte("package main\nfunc broken( {\n}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		mustRunGit(t, fixture, "add", "main.go")
		mustRunGit(t, fixture, "commit", "-m", "docs(SPEC-X): sync-phase artifacts (broken build)")
		// Clear the once-per-commit sentinel.
		_ = os.RemoveAll(filepath.Join(fixture, ".moai", "state"))

		env := []string{
			"PATH=" + os.Getenv("PATH"),
			"HOME=" + t.TempDir(),
			"CLAUDE_PROJECT_DIR=" + fixture,
		}
		if setTier {
			env = append(env, config.EnvAutonomyTier+"="+tier)
		} else {
			env = append(env, config.EnvAutonomyTier+"=")
		}
		cmd := exec.Command("bash", script)
		cmd.Stdin = strings.NewReader(`{}`)
		cmd.Dir = fixture
		cmd.Env = env
		var sb strings.Builder
		cmd.Stdout = &sb
		_ = cmd.Run()
		// exit code is always 0 per the script contract; capture stdout.
		return sb.String(), 0
	}

	t.Run("fully-autonomous → advisory only (no decision:block)", func(t *testing.T) {
		out, _ := runGate(t, config.AutonomyTierFullyAutonomous, true)
		if strings.Contains(out, `"decision":"block"`) {
			t.Fatalf("AC-004: fully-autonomous gate emitted decision:block — must be advisory only. stdout=%s", out)
		}
	})
	t.Run("semi-auto → retains decision:block (regression guard)", func(t *testing.T) {
		out, _ := runGate(t, config.AutonomyTierSemiAuto, true)
		if !strings.Contains(out, `"decision":"block"`) {
			t.Fatalf("AC-004 regression: semi-auto gate did NOT emit decision:block on a failing build — advisory downgrade leaked into semi-auto. stdout=%s", out)
		}
	})
}

// TestAC007_UnsetTokenSemiAutoBehavior (REQ-003 / AC-STOPCHAIN-TRIM-007):
// an unset/empty MOAI_AUTONOMY_TIER resolves to semi-auto at every Go reader,
// preserving today's behavior (the IsGitCommit gate runs; the lifecycle hooks
// are active; the sync-gate full-blocks). This is the backward-compat sweep.
// The reader-level fallback is covered by TestAutonomyTierReader (M1); here we
// verify the PRE_TOOL integration: with the token unset, the commit gate is
// active (gate would run when Enabled). Verified structurally via the predicate.
func TestAC007_UnsetTokenKeepsGateActive(t *testing.T) {
	// Unset → AutonomyTier() returns semi-auto → IsAutonomyTierCommitGateOff=false.
	_ = os.Unsetenv(config.EnvAutonomyTier)
	tier := config.AutonomyTier()
	if tier != config.AutonomyTierSemiAuto {
		t.Fatalf("AC-007: unset token resolved to %q, want %q (backward compat)", tier, config.AutonomyTierSemiAuto)
	}
	if config.IsAutonomyTierCommitGateOff(tier) {
		t.Fatalf("AC-007: unset token turned the commit gate OFF — must stay ON (backward compat)")
	}
	if config.IsAutonomyTierLifecycleDormant(tier) {
		t.Fatalf("AC-007: unset token made lifecycle hooks dormant — must stay active (backward compat)")
	}
	// Also exercise the JSON stdin helper to keep the import honest.
	_ = jsonStdin(t, map[string]any{"k": "v"})
	_ = json.RawMessage(`{}`)
}
