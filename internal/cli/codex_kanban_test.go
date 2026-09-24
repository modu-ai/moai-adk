package cli

// codex_kanban_test.go — `moai codex -k` enters Kanban Mode on the same terms
// as `moai cc -k`: the same lead and companion shapes, the same launch facts
// (only the backend differs), the same name-claim rules, and nothing of the
// kanban surface in the Codex child's argv. Unsupported shapes are refused
// rather than degraded to a plain launch.

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// kanbanFactKeys are the facts a kanban launch publishes to the launched
// session. MOAI_KANBAN_BACKEND is compared separately (it is the one fact
// that must differ) and MOAI_KANBAN_SETTINGS_INJECTED is Claude's --settings
// transport, which has no Codex counterpart.
var kanbanFactKeys = []string{
	config.EnvMoaiKanban,
	config.EnvMoaiKanbanID,
	config.EnvMoaiKanbanSpec,
	config.EnvMoaiKanbanLeadAddr,
	config.EnvMoaiKanbanLabel,
	config.EnvMoaiKanbanLeadName,
	config.EnvAutonomyTier,
	config.EnvClaudeCodeMaxConcurrentSubagents,
}

// kanbanLaunch is what one launch published: the facts, the backend, the
// session name it answers to, and the argv the child received.
type kanbanLaunch struct {
	facts   map[string]string
	backend string
	name    string
	argv    []string
}

func factsFromEnv(lookup func(string) (string, bool)) (map[string]string, string) {
	facts := map[string]string{}
	for _, k := range kanbanFactKeys {
		if v, ok := lookup(k); ok {
			facts[k] = v
		}
	}
	backend, _ := lookup(config.EnvMoaiKanbanBackend)
	return facts, backend
}

// ccKanbanLaunch runs `moai cc` with the given args under project root and
// records what the launch function saw.
func ccKanbanLaunch(t *testing.T, root string, args ...string) kanbanLaunch {
	t.Helper()
	t.Setenv(config.EnvClaudeProjectDir, root)
	var got kanbanLaunch
	c := &cobra.Command{Use: "cc"}
	c.SetOut(new(strings.Builder))
	c.SetErr(new(strings.Builder))
	err := runClaudeEntry(c, args, "cc", "claude", kanban.BackendClaude, func(_, _ string, argv []string) error {
		got.facts, got.backend = factsFromEnv(os.LookupEnv)
		got.argv = append([]string(nil), argv...)
		got.name, _ = parseNamedLabel(argv, func(string) bool { return true })
		return nil
	})
	if err != nil {
		t.Fatalf("cc %v: %v", args, err)
	}
	return got
}

// codexKanbanLaunch runs `moai codex` with the given args under project root
// through the capture harness and records the child's environment and argv.
func codexKanbanLaunch(t *testing.T, cap *codexLaunchCapture, root string, args ...string) kanbanLaunch {
	t.Helper()
	t.Setenv(config.EnvClaudeProjectDir, root)
	before := cap.count()
	if _, stderr, err := runCodexCmd(t, args...); err != nil {
		t.Fatalf("codex %v: %v (stderr %q)", args, err, stderr)
	}
	if cap.count() != before+1 {
		t.Fatalf("codex %v: launches %d -> %d, want one", args, before, cap.count())
	}
	rec := cap.records[len(cap.records)-1]
	var got kanbanLaunch
	got.facts, got.backend = factsFromEnv(func(k string) (string, bool) { return codexEnvLast(rec.Env, k) })
	got.argv = rec.Argv
	if label, ok := got.facts[config.EnvMoaiKanbanLabel]; ok {
		got.name = label
	} else {
		got.name = got.facts[config.EnvMoaiKanbanLeadName]
	}
	return got
}

// assertNoKanbanTokens checks the child argv carries none of the kanban
// surface the launcher consumed.
func assertNoKanbanTokens(t *testing.T, argv []string, extra ...string) {
	t.Helper()
	banned := append([]string{"-k", "--kanban", "--name", "-n"}, extra...)
	for _, tok := range argv[1:] {
		for _, b := range banned {
			if tok == b || strings.HasPrefix(tok, "--kanban=") || strings.HasPrefix(tok, "--name=") {
				t.Errorf("kanban token %q reached the codex child argv %#v", tok, argv)
			}
		}
	}
}

// compareLaunches is the parity judgement: identical facts, backends claude
// and codex respectively, identical session names.
func compareLaunches(t *testing.T, cc, cx kanbanLaunch) {
	t.Helper()
	if !reflect.DeepEqual(cc.facts, cx.facts) {
		t.Errorf("launch facts differ:\n cc    %v\n codex %v", cc.facts, cx.facts)
	}
	if cc.backend != kanban.BackendClaude || cx.backend != codexFactoryBackend {
		t.Errorf("backends = (%q, %q), want (%q, %q)", cc.backend, cx.backend, kanban.BackendClaude, codexFactoryBackend)
	}
	if cc.name == "" || cc.name != cx.name {
		t.Errorf("session names differ: cc %q, codex %q", cc.name, cx.name)
	}
}

// TestCodexKanbanEntryParity is AC-DHR-009.
func TestCodexKanbanEntryParity(t *testing.T) {
	// A fixed run id makes the published id comparable across the two paths;
	// both adopt an id already standing in the environment.
	t.Setenv(config.EnvMoaiKanbanID, "run-parity")
	for _, k := range append(append([]string{}, kanbanFactKeys...), config.EnvMoaiKanbanBackend) {
		if k == config.EnvMoaiKanbanID {
			continue
		}
		t.Setenv(k, "")
		_ = os.Unsetenv(k)
	}

	t.Run("lead", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		ccRoot, cxRoot := t.TempDir(), t.TempDir()
		withCodexProjectRoot(t, cxRoot)
		// Two launches each: the first claims `lead`, the second — while the
		// first claim's holder is alive — is bumped by the same rule.
		for i := 0; i < 2; i++ {
			cc := ccKanbanLaunch(t, ccRoot, "-k", "SPEC-PARITY-001")
			cx := codexKanbanLaunch(t, cap, cxRoot, "-k", "SPEC-PARITY-001")
			compareLaunches(t, cc, cx)
			assertNoKanbanTokens(t, cx.argv, "SPEC-PARITY-001", cc.name)
			if cx.facts[config.EnvMoaiKanban] != "1" || cx.facts[config.EnvMoaiKanbanSpec] != "SPEC-PARITY-001" {
				t.Errorf("codex lead facts %v lack the kanban signal or SPEC", cx.facts)
			}
		}
		if got := cap.records[1].Env; got == nil {
			t.Fatal("second codex launch recorded no environment")
		}
		if name, _ := codexEnvLast(cap.records[1].Env, config.EnvMoaiKanbanLeadName); name != kanban.LeadNumberLabel(1) {
			t.Errorf("second codex lead name = %q, want the bumped %q", name, kanban.LeadNumberLabel(1))
		}
	})

	t.Run("companion", func(t *testing.T) {
		cap := withCodexLaunchCapture(t)
		ccRoot, cxRoot := t.TempDir(), t.TempDir()
		withCodexProjectRoot(t, cxRoot)
		for i := 0; i < 2; i++ {
			cc := ccKanbanLaunch(t, ccRoot, "-k", "--name", "plan")
			cx := codexKanbanLaunch(t, cap, cxRoot, "-k", "--name", "plan")
			compareLaunches(t, cc, cx)
			assertNoKanbanTokens(t, cx.argv, "plan", cc.name)
		}
		if label, _ := codexEnvLast(cap.records[1].Env, config.EnvMoaiKanbanLabel); label != "plan-1" {
			t.Errorf("second codex companion label = %q, want the bumped plan-1", label)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for _, args := range [][]string{
			{"-k", "3"},                        // the factory count shape belongs to -f
			{"-k", "--name", "worker-1"},       // a factory worker name
			{"--kanban=SPEC-PARITY-001"},       // joined form is the count's only
			{"-k", "-f"},                       // kanban and factory together
			{"status", "-k"},                   // a readout has no session to enter
			{"--name", "plan"},                 // a companion name without -k
			{"-k", "--name", "plan", "--name"}, // dangling name flag
		} {
			t.Run(strings.Join(args, "_"), func(t *testing.T) {
				cap := withCodexLaunchCapture(t)
				withCodexProjectRoot(t, t.TempDir())
				t.Setenv(config.EnvClaudeProjectDir, t.TempDir())
				_, stderr, err := runCodexCmd(t, args...)
				if err == nil {
					t.Fatalf("codex %v was accepted", args)
				}
				if code, ok := ResolveExitCode(err); ok && code == 0 {
					t.Errorf("codex %v exited 0", args)
				}
				codexWantLaunches(t, cap, 0, 0, 0)
				if !strings.Contains(stderr+err.Error(), "usage") {
					t.Errorf("codex %v: diagnostic %q / %v carries no usage", args, stderr, err)
				}
			})
		}
	})
}
