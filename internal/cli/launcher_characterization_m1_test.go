package cli

// launcher_characterization_m1_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M1
// (card t1399): the characterization tests of AC-009 (the leader entry, the
// retired `cg`, the removed `gpt`, and the bare `moai` are unchanged by the
// entry-flag work) and the cli half of AC-017 (pre-existing kanban artifacts
// do not break a reader).
//
// AC-017 fixture note: the surviving-session environment is written with the
// marker names as string literals, not through the internal/config constants —
// M5b deletes those constants, and this test must outlive them (plan-audit
// finding D-A7).

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// TestLeaderEntryUnchanged characterizes today's leader entry and the removed
// or retired verbs (AC-009): `moai cc -f` and `moai glm -f` (and the long
// `--factory`) start a leader that records its run with the derived capacity,
// `moai codex -f` has no leader entry, `moai gpt` is an unknown command.
func TestLeaderEntryUnchanged(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend string
		entry   func([]string) error
		args    []string
	}{
		{"cc -f", factory.BackendClaude, netCC, []string{"-f"}},
		{"cc --factory", factory.BackendClaude, netCC, []string{"--factory"}},
		{"glm -f", factory.BackendGLM, netGLM, []string{"-f"}},
		{"glm --factory", factory.BackendGLM, netGLM, []string{"--factory"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := netLeaderFixture(t)
			launch := netDriveLaunch(t, root, tc.entry, tc.args)
			if launch.err != nil || !launch.launched {
				t.Fatalf("%v: launched=%v err=%v", tc.args, launch.launched, launch.err)
			}
			runID := launch.env[config.EnvFactoryRunID]
			if runID == "" {
				t.Fatalf("the leader carries no run id at launch")
			}
			if launch.env[config.EnvMoaiFactoryWorkers] != "1" {
				t.Errorf("leader markers: %s = %q, want 1", config.EnvMoaiFactoryWorkers, launch.env[config.EnvMoaiFactoryWorkers])
			}
			if launch.env[config.EnvFactoryLeadAddr] != factory.FactoryLeaderSocketPath(runID) {
				t.Errorf("leader markers: %s = %q, want the run's leader socket", config.EnvFactoryLeadAddr, launch.env[config.EnvFactoryLeadAddr])
			}
			netRunRecorded(t, root, runID, tc.backend)
		})
	}

	t.Run("codex -f has no leader entry", func(t *testing.T) {
		root := pinCodexRefusalRoot(t)
		netScrubLaneEnv(t)
		launches := 0
		prevDirect, prevSpawn, prevLook := codexDirectLaunchFn, codexSpawnLaunchFn, codexLookPath
		codexDirectLaunchFn = func(*exec.Cmd) error { launches++; return nil }
		codexSpawnLaunchFn = func(string, string, []string, []string) error { launches++; return nil }
		codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
		t.Cleanup(func() { codexDirectLaunchFn, codexSpawnLaunchFn, codexLookPath = prevDirect, prevSpawn, prevLook })

		for _, args := range [][]string{{"-f"}, {"--factory"}} {
			_, stderr, err := runCodexCmd(t, args...)
			if code, ok := ResolveExitCode(err); !ok || code != 1 {
				t.Errorf("codex %v: exit code = (%d, %v), want (1, true); err=%v", args, code, ok, err)
			}
			if strings.TrimSpace(stderr) == "" {
				t.Errorf("codex %v: no refusal line on stderr", args)
			}
		}
		if launches != 0 {
			t.Errorf("a refused codex leader entry launched %d child(ren)", launches)
		}
		if reg := loadFactoryRegistry(factoryRegistryPath(root)); len(reg) != 0 {
			t.Errorf("a refused codex leader entry wrote lane claims: %v", reg)
		}
	})

	t.Run("gpt is an unknown command", func(t *testing.T) {
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "gpt" {
				t.Fatal("the removed gpt verb is registered on the root command")
			}
		}
		out := new(bytes.Buffer)
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)
		rootCmd.SetArgs([]string{"gpt"})
		t.Cleanup(func() { rootCmd.SetArgs(nil); rootCmd.SetOut(nil); rootCmd.SetErr(nil) })
		err := rootCmd.Execute()
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), `unknown command "gpt"`) {
			t.Errorf("moai gpt: err = %v, want an unknown-command error naming gpt", err)
		}
	})
}

// TestBareMoaiPrintsBannerAndHelp characterizes the bare `moai` (AC-009): it
// prints the banner on stdout, then the help text, and succeeds (exit 0, no
// error).
func TestBareMoaiPrintsBannerAndHelp(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	help := new(bytes.Buffer)
	rootCmd.SetOut(help)
	rootCmd.SetErr(help)
	rootCmd.SetArgs([]string{})
	t.Cleanup(func() { rootCmd.SetArgs(nil); rootCmd.SetOut(nil); rootCmd.SetErr(nil) })

	// The banner goes through the stdout data channel, which resolves os.Stdout
	// at call time.
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	execErr := rootCmd.Execute()
	_ = w.Close()
	os.Stdout = old
	var banner bytes.Buffer
	_, _ = banner.ReadFrom(r)
	_ = r.Close()

	if execErr != nil {
		t.Fatalf("bare moai returned an error (the exit code would be non-zero): %v", execErr)
	}
	if !strings.Contains(banner.String(), "Agentic Development Kit") {
		t.Errorf("the banner is missing from stdout: %q", banner.String())
	}
	if !strings.Contains(help.String(), "MoAI-ADK") || !strings.Contains(help.String(), "moai doctor") {
		t.Errorf("the help text is missing from the command output: %q", help.String())
	}
}

// TestPreexistingKanbanArtifactsTolerated is the cli half of AC-017: a project
// holding session records with the retired chain roles, a kanban-board
// directory with an unreadable role declaration, and a surviving session
// environment carrying the two kanban markers does not make the doctor's
// readers fail; an unreadable artifact degrades to absence.
func TestPreexistingKanbanArtifactsTolerated(t *testing.T) {
	root := t.TempDir()
	for _, role := range []string{"plan", "run", "sync"} {
		path := factory.RecordPath(root, "old-"+role)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		raw := `{"session_id":"old-` + role + `","spec_id":"","role":"` + role + `","backend":"claude","entered_at":"2026-09-01T00:00:00Z","deepscan_dir":"","verify_reentries":0}` + "\n"
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// A record the reader cannot parse degrades to absence beside the good ones.
	garbled := factory.RecordPath(root, "old-garbled")
	if err := os.WriteFile(garbled, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	roles := filepath.Join(root, ".moai", "state", "kanban-board", "roles")
	if err := os.MkdirAll(roles, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(roles, "plan.json"), []byte("\x00\x01 not a role declaration"), 0o600); err != nil {
		t.Fatal(err)
	}
	// The marker names are literals on purpose (see the file comment).
	t.Setenv("MOAI_KANBAN", "1")
	t.Setenv("MOAI_KANBAN_LABEL", "plan")

	records, err := factory.ReadAll(root)
	if err != nil {
		t.Fatalf("factory.ReadAll over the pre-existing records: %v", err)
	}
	if len(records) < 3 {
		t.Errorf("factory.ReadAll returned %d record(s), want the three role records readable (positive control)", len(records))
	}

	for _, check := range []DiagnosticCheck{
		checkFactoryRun(root, true),
		checkOwnerLabelDrift(root, true),
	} {
		if check.Status == uikit.CheckFail {
			t.Errorf("doctor check %q failed on pre-existing kanban artifacts: %+v", check.Name, check)
		}
	}
}
