package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// goalVerbOutcome runs `moai goal <args...> --session <session>` against a
// temporary project root and reports whether a goal state file was written.
// It never touches a real session's .moai/state/goal directory.
func goalVerbOutcome(t *testing.T, session string, args ...string) (armed bool, out string, err error) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	// Fail closed if the seam does not hold: every goal verb resolves its state
	// root through goalProjectRoot, so a mismatch here would mean the run could
	// write into a real checkout's .moai/state/goal.
	if got := goalProjectRoot(); got != root {
		t.Fatalf("goal state root = %q, want temp root %q", got, root)
	}

	rc, buf := newGoalTestRoot()
	rc.SetArgs(append(append([]string{"goal"}, args...), "--session", session))
	err = rc.Execute()
	_, statErr := os.Stat(filepath.Join(root, goal.StateDir, session+".json"))
	return statErr == nil, buf.String(), err
}

// TestGoalVerbControls pins the four control cells for card t605 (an
// unregistered verb falling through to goal arming). A reproduction that
// measures only the typo cell cannot tell "everything arms" from "everything
// is refused", so the registered verbs, a real condition, and the empty
// argument are measured beside it.
func TestGoalVerbControls(t *testing.T) {
	t.Run("registered verb status arms nothing", func(t *testing.T) {
		armed, out, err := goalVerbOutcome(t, "VSTATUS", "status")
		if err != nil || armed || !strings.Contains(out, "no armed goal") {
			t.Fatalf("status: armed=%v err=%v out=%q", armed, err, out)
		}
	})
	t.Run("registered verb clear arms nothing", func(t *testing.T) {
		armed, out, err := goalVerbOutcome(t, "VCLEAR", "clear")
		if err != nil || armed {
			t.Fatalf("clear: armed=%v err=%v out=%q", armed, err, out)
		}
	})
	t.Run("unregistered verb statuss is refused", func(t *testing.T) {
		armed, out, err := goalVerbOutcome(t, "VTYPO", "statuss")
		t.Logf("statuss: armed=%v err=%v out=%q", armed, err, out)
		if err == nil || armed {
			t.Fatalf("statuss was armed (armed=%v err=%v)", armed, err)
		}
	})
	t.Run("real condition arms", func(t *testing.T) {
		armed, out, err := goalVerbOutcome(t, "VREAL", "go test ./... exits 0")
		if err != nil || !armed {
			t.Fatalf("real condition: armed=%v err=%v out=%q", armed, err, out)
		}
	})
	t.Run("no argument prints help and arms nothing", func(t *testing.T) {
		armed, out, err := goalVerbOutcome(t, "VNOARG")
		if err != nil || armed || !strings.Contains(out, "Verbs:") {
			t.Fatalf("no arg: armed=%v err=%v out=%q", armed, err, out)
		}
	})
	t.Run("empty string argument is refused", func(t *testing.T) {
		armed, out, err := goalVerbOutcome(t, "VEMPTY", "")
		if err == nil || armed {
			t.Fatalf("empty arg: armed=%v err=%v out=%q", armed, err, out)
		}
	})
}

// TestGoalVerbResolvingTokens measures the shape the controls leave open: a
// single verb-like word that the shell DOES resolve passes the runnability
// probe and is too short for the prose-shape check. Measurement only — it
// logs each outcome and asserts nothing, so the fix scope is decided on
// observed behaviour rather than on the static reading.
func TestGoalVerbResolvingTokens(t *testing.T) {
	for _, tok := range []string{"stat", "ls", "reset", "done", "cancel", "rm", "help", "list", "show"} {
		t.Run(tok, func(t *testing.T) {
			armed, _, err := goalVerbOutcome(t, "VTOK", tok)
			t.Logf("token=%q armed=%v err=%v", tok, armed, err)
		})
	}
}

// TestGoalVerbMisreadWordsRefused pins the fix for card t605: a single word a
// user would plausibly type as a goal verb is refused with the verb it meant,
// instead of being armed as a condition. Several of these words resolve as
// shell commands (reset, done, cancel, stat), so the arm-time runnability gate
// alone lets them through and the goal blocks every turn-end.
func TestGoalVerbMisreadWordsRefused(t *testing.T) {
	for _, tc := range []struct{ word, verb string }{
		{"cancel", "clear"},
		{"reset", "clear"},
		{"stop", "clear"},
		{"done", "clear"},
		{"STOP", "clear"},
		{"show", "status"},
		{"list", "status"},
		{"info", "status"},
		{"stat", "status"},
	} {
		t.Run(tc.word, func(t *testing.T) {
			armed, out, err := goalVerbOutcome(t, "VMISREAD", tc.word)
			if err == nil || armed {
				t.Fatalf("%q was armed (armed=%v err=%v out=%q)", tc.word, armed, err, out)
			}
			if !strings.Contains(err.Error(), "moai goal "+tc.verb) {
				t.Errorf("refusal for %q does not suggest %q: %v", tc.word, tc.verb, err)
			}
			if !strings.Contains(err.Error(), "cmd:") {
				t.Errorf("refusal for %q does not name the cmd: escape: %v", tc.word, err)
			}
		})
	}
}

// TestGoalVerbHelpWordShowsHelp: "help" as the single argument is a request for
// help, not a condition.
func TestGoalVerbHelpWordShowsHelp(t *testing.T) {
	armed, out, err := goalVerbOutcome(t, "VHELP", "help")
	if err != nil || armed || !strings.Contains(out, "Verbs:") {
		t.Fatalf("help: armed=%v err=%v out=%q", armed, err, out)
	}
}

// TestGoalVerbMisreadCheckStaysNarrow is the over-refusal guard: the check
// matches a single whole argument only, so an explicit arm verb, a declared
// cmd: condition, a multi-word condition, and an unlisted command still arm.
func TestGoalVerbMisreadCheckStaysNarrow(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"explicit arm verb", []string{"arm", "reset"}},
		{"cmd prefix", []string{"cmd: reset"}},
		{"multi-word condition", []string{"reset && true"}},
		{"unlisted command", []string{"ls"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			armed, out, err := goalVerbOutcome(t, "VNARROW", tc.args...)
			if err != nil || !armed {
				t.Fatalf("%v was not armed (armed=%v err=%v out=%q)", tc.args, armed, err, out)
			}
		})
	}
}
