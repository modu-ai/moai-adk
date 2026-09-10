package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// TestGoalArm_RefusesUnrunnableMechanicalCondition is the arm-time gate: prose
// that classified mechanical would be handed to `sh -c`, where its first word
// is not a command — exit 127, forever. Arming it silently (exit 0, "armed
// goal ... (mechanical condition ...)") is the defect; the arm MUST refuse and
// name the `model:` escape hatch.
func TestGoalArm_RefusesUnrunnableMechanicalCondition(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", koreanProse, "--session", "REFUSE"})
	err := rc.Execute()
	if err == nil {
		t.Fatalf("goal arm accepted unrunnable prose (out=%s)", buf.String())
	}

	combined := buf.String() + err.Error()
	// (a) the offending first word.
	if !strings.Contains(combined, "모든") {
		t.Errorf("refusal does not name the offending first word; got:\n%s", combined)
	}
	// (b) says it was treated as a shell command.
	if !strings.Contains(combined, "shell command") {
		t.Errorf("refusal does not say the string was treated as a shell command; got:\n%s", combined)
	}
	// (c) names the model: prefix as the remedy.
	if !strings.Contains(combined, "model:") {
		t.Errorf("refusal does not name the model: prefix; got:\n%s", combined)
	}

	// No goal state may be written on refusal.
	if _, statErr := os.Stat(filepath.Join(root, goal.StateDir, "REFUSE.json")); statErr == nil {
		t.Errorf("refused arm still wrote a goal state file")
	}
}

// TestGoalArm_ModelPrefixEscapesTheRefusal proves the remedy the refusal names
// actually works: the same prose, declared explicitly, arms as a model condition.
func TestGoalArm_ModelPrefixEscapesTheRefusal(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", "model: " + koreanProse, "--session", "OK"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("model-prefixed prose refused: %v (out=%s)", err, buf.String())
	}
	g, err := goal.LoadGoal(root, "OK")
	if err != nil || g == nil {
		t.Fatalf("LoadGoal after model-prefixed arm: g=%+v err=%v", g, err)
	}
	if len(g.Conditions) != 1 || g.Conditions[0].Type != goal.ConditionModel {
		t.Fatalf("conditions = %+v, want one model condition", g.Conditions)
	}
	if g.Conditions[0].Claim != koreanProse {
		t.Errorf("claim = %q, want the prefix-stripped prose", g.Conditions[0].Claim)
	}
}

// TestGoalArm_NoFalseRefusals is the fail-open half. The gate refuses only on
// POSITIVE evidence that the first token resolves to nothing; every shape the
// naive first-token extraction cannot judge must be allowed through, because a
// false refusal is worse than a missed catch.
func TestGoalArm_NoFalseRefusals(t *testing.T) {
	cases := []struct {
		name string
		cond string
	}{
		{"plain command", "go test ./internal/cli/..."},
		{"shell builtin only", "true"},
		{"builtin cd with an operator", "cd /tmp && ls"},
		{"env assignment prefix", "FOO=bar env"},
		{"relative script path", "./scripts/does-not-exist-yet.sh"},
		{"absolute path", "/usr/bin/env true"},
		{"subshell", "( true )"},
		{"negation", "! false"},
		{"shell keyword", "if true; then exit 0; fi"},
		{"variable expansion", "$SHELL -c true"},
		{"trailing exits clause", "go build ./... exits 0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("CLAUDE_PROJECT_DIR", root)
			rc, buf := newGoalTestRoot()
			rc.SetArgs([]string{"goal", "arm", tc.cond, "--session", "S"})
			if err := rc.Execute(); err != nil {
				t.Fatalf("false refusal for %q: %v (out=%s)", tc.cond, err, buf.String())
			}
			if _, statErr := os.Stat(filepath.Join(root, goal.StateDir, "S.json")); statErr != nil {
				t.Errorf("accepted arm wrote no state file: %v", statErr)
			}
		})
	}
}

// TestGoalArm_ModelConditionSkipsTheGate: a model condition has no command, so
// the runnability gate must not run against its claim text.
func TestGoalArm_ModelConditionSkipsTheGate(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", oneLineWithReferent, "--session", "M"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("model condition refused by the runnability gate: %v (out=%s)", err, buf.String())
	}
}

// TestGoalArm_CmdPrefixExemptsTheRunnabilityGate pins the escape hatch the
// refusal message advertises. The message says "if it really is a command,
// declare that with the cmd: prefix" — so `cmd:` MUST actually open that door,
// or the message names a remedy that does not exist.
//
// The exemption is not merely permissive. `command -v` probes the ARMING
// environment, so a condition naming a tool that exists at eval time but not at
// arm time (a different PATH, a container, a binary the goal itself builds) is a
// legitimate goal the gate would otherwise refuse with no way to override it. A
// deliberate `cmd:` on genuine prose is still caught by the eval-time exit-127
// backstop, one turn in rather than thirty.
func TestGoalArm_CmdPrefixExemptsTheRunnabilityGate(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	// A command that positively resolves to nothing at arm time — the exact
	// shape the gate refuses without the prefix.
	const notYetBuilt = "cmd: t436-tool-that-does-not-exist --check"

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", notYetBuilt, "--session", "CMDPREFIX"})
	if err := rc.Execute(); err != nil {
		t.Fatalf("cmd:-declared condition was refused: %v\nout=%s", err, buf.String())
	}

	g, err := goal.LoadGoal(root, "CMDPREFIX")
	if err != nil {
		t.Fatalf("LoadGoal: %v", err)
	}
	if len(g.Conditions) != 1 || g.Conditions[0].Type != goal.ConditionMechanical {
		t.Fatalf("conditions = %+v, want one mechanical condition", g.Conditions)
	}
	if g.Conditions[0].Cmd != "t436-tool-that-does-not-exist --check" {
		t.Errorf("cmd = %q, want the prefix stripped", g.Conditions[0].Cmd)
	}
}

// TestGoalArm_UndeclaredProseStillRefused is the other half: the exemption keys
// on the EXPLICIT prefix, not on the mechanical tier, so a condition that
// reached that tier through the substring fallback is still refused.
func TestGoalArm_UndeclaredProseStillRefused(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	rc, buf := newGoalTestRoot()
	rc.SetArgs([]string{"goal", "arm", "t436-tool-that-does-not-exist --check", "--session", "NOPREFIX"})
	if err := rc.Execute(); err == nil {
		t.Fatalf("undeclared unrunnable condition was armed (out=%s)", buf.String())
	}
}

// TestDeclaredMechanical pins the predicate itself: only an explicit `cmd:`
// counts — `model:` does not, and neither does ordinary text that happens to
// carry a colon.
func TestDeclaredMechanical(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"cmd: go test ./...", true},
		{"CMD:go test ./...", true},
		{"  cmd : go test ./...  ", true},
		{"model: all AC rows PASS", false},
		{"go test ./...", false},
		{"cmdline: go test ./...", false},
		{"run this: go test ./...", false},
		{"", false},
	} {
		if got := declaredMechanical(tc.in); got != tc.want {
			t.Errorf("declaredMechanical(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
