package cli

import (
	"bytes"
	"context"
	"errors"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCGRetiredCommandHasNoRootRegistration(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "cg" {
			t.Fatal("retired cg remains registered")
		}
	}
}
func TestCGRetiredEntryAndModeHaveZeroEffects(t *testing.T) {
	root := cgProject(t)
	oldLaunch, oldRoot, oldInject := unifiedLaunchFunc, findProjectRootFn, injectTmuxSessionEnvFn
	defer func() { unifiedLaunchFunc, findProjectRootFn, injectTmuxSessionEnvFn = oldLaunch, oldRoot, oldInject }()
	launch, reads, injections := 0, 0, 0
	unifiedLaunchFunc = func(string, string, []string) error { launch++; return nil }
	findProjectRootFn = func() (string, error) { reads++; return root, nil }
	injectTmuxSessionEnvFn = func(*GLMConfigFromYAML, string) error { injections++; return nil }
	for _, args := range [][]string{nil, {"--help"}, {"-p"}, {"--spawn"}, {"-k"}, {"-f", "3"}, {"--unknown"}} {
		if err := runCG(cgCmd, args); !errors.Is(err, errCGRetired) {
			t.Fatalf("retired %v: %v", args, err)
		}
	}
	for _, mode := range []string{"cg", "claude_glm"} {
		if err := unifiedLaunchWithGateway("", mode, nil, nil); !errors.Is(err, errCGRetired) {
			t.Fatalf("mode %s: %v", mode, err)
		}
	}
	if err := applyCGMode(root, ""); !errors.Is(err, errCGRetired) {
		t.Fatalf("old apply entry: %v", err)
	}
	if launch != 0 || reads != 0 || injections != 0 {
		t.Fatalf("retired effects: launch=%d root-read=%d injection=%d", launch, reads, injections)
	}
}

func TestCGRetirementHistoricalRecordRemainsReadable(t *testing.T) {
	root := t.TempDir()
	path := kanban.RecordPath(root, "historical-session")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	original := []byte(`{"session_id":"historical-session","backend":"cg","spec_id":"SPEC-HISTORICAL","future_field":"keep"}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	rec, err := kanban.Read(root, "historical-session")
	if err != nil || rec.Backend != "cg" {
		t.Fatalf("historical backend %+v %v", rec, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(original) {
		t.Fatal("reading history rewrote record")
	}
	if template.IsGLMBackend(config.LLMConfig{TeamMode: rec.Backend}) {
		t.Fatal("historical record reactivated live GLM backend")
	}
}

func TestCGRetirementCompleteEntryShapesAndCounters(t *testing.T) {
	root := cgProject(t)
	t.Chdir(root)
	t.Setenv(config.EnvClaudeProjectDir, root)
	oldRoot, oldLaunch, oldTmux, oldSpawn, oldLook, oldWorktree, oldHome := findProjectRootFn, unifiedLaunchFunc, inTmuxFn, tmuxSpawnFn, spawnLookPath, launcherWorktreeMaterialize, userHomeDirFn
	defer func() {
		findProjectRootFn, unifiedLaunchFunc, inTmuxFn, tmuxSpawnFn, spawnLookPath, launcherWorktreeMaterialize, userHomeDirFn = oldRoot, oldLaunch, oldTmux, oldSpawn, oldLook, oldWorktree, oldHome
	}()
	findProjectRootFn = func() (string, error) { return root, nil }
	launches, spawns, worktrees, credentialHomes := 0, 0, 0, 0
	launch := func(string, string, []string) error { launches++; return nil }
	unifiedLaunchFunc = launch
	inTmuxFn = func() bool { return true }
	spawnLookPath = func(string) (string, error) { return "/mock/moai", nil }
	tmuxSpawnFn = func(string, string) (string, error) { spawns++; return "%1", nil }
	launcherWorktreeMaterialize = func(string, string, string, io.Writer) error { worktrees++; return nil }
	userHomeDirFn = func() (string, error) { credentialHomes++; return root, nil }
	shapes := [][]string{nil, {"--model", "opus"}, {"--continue"}, {"--resume", "prior-session"}, {"--spawn"}, {"-w", "owned-feature", "--branch", "existing"}, {"-p", "profile"}, {"-k", "2"}, {"-f", "2"}}
	for _, name := range []string{"cc", "glm", "gpt"} {
		for _, args := range shapes {
			cmd := newGPTCommand(gptCommandServices{Launch: launch})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			var err error
			switch name {
			case "cc":
				err = runCC(cmd, args)
			case "glm":
				err = runGLM(cmd, args)
			default:
				err = cmd.RunE(cmd, args)
			}
			if !errors.Is(err, config.ErrLegacyCG) {
				t.Fatalf("%s %v: %v", name, args, err)
			}
		}
	}
	if launches+spawns+worktrees+credentialHomes != 0 {
		t.Fatalf("guard counters launch=%d spawn=%d worktree=%d credential-home=%d", launches, spawns, worktrees, credentialHomes)
	}
	source := filepath.Join(root, ".moai/config/sections/llm.yaml")
	raw, _ := os.ReadFile(source)
	if string(raw) != migrationCGInput {
		t.Fatal("guard changed source")
	}
	migrated, err := config.PlanCGMigration(raw, "claude-only")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, migrated.Bytes, 0600); err != nil {
		t.Fatal(err)
	}
	// Valid, explicitly migrated controls traverse the same full entry shapes.
	for _, name := range []string{"cc", "glm", "gpt"} {
		for _, args := range shapes {
			beforeLaunch, beforeSpawn := launches, spawns
			cmd := newGPTCommand(gptCommandServices{Launch: launch})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			var err error
			switch name {
			case "cc":
				err = runCC(cmd, args)
			case "glm":
				err = runGLM(cmd, args)
			default:
				err = cmd.RunE(cmd, args)
			}
			if err != nil {
				t.Fatalf("control %s %v: %v", name, args, err)
			}
			if launches+spawns != beforeLaunch+beforeSpawn+1 {
				t.Fatalf("control did not reach prepared boundary %s %v", name, args)
			}
		}
	}
	if worktrees != 3 {
		t.Fatalf("worktree counter not live: %d", worktrees)
	}
}
func TestCGRetirementHelpDoesNotOfferCG(t *testing.T) {
	for _, group := range rootHelpGroups() {
		for _, row := range group.rows {
			if row[0] == "moai cg" {
				t.Fatal("help offers retired command")
			}
		}
	}
}

func TestCGRetirementLiveHelpHasNoLaunchRecommendation(t *testing.T) {
	for name, help := range map[string]string{"root": rootCmd.Long, "glm": glmCmd.Long} {
		if strings.Contains(help, "moai cg") {
			t.Fatalf("%s help recommends retired launcher", name)
		}
	}
}

func TestCGRetiredSpawnBoundaryDoesNotOpenWindow(t *testing.T) {
	cap := withSpawnStubs(t, true, "%7", nil)
	if err := spawnLaunch(io.Discard, "cg", []string{"-w", "feature"}); !errors.Is(err, errCGRetired) {
		t.Fatalf("retired spawn: %v", err)
	}
	if cap.calls != 0 {
		t.Fatalf("retired spawn opened %d windows", cap.calls)
	}
}

func TestCGRetiredTopLevelDiagnostic(t *testing.T) {
	original := os.Args
	defer func() { os.Args = original }()
	for _, args := range [][]string{{"moai", "cg"}, {"moai", "cg", "--help"}, {"moai", "cg", "--spawn", "-f", "2"}} {
		os.Args = args
		if err := Execute(); !errors.Is(err, errCGRetired) {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

// TestCGRetiredExecutableDiagnostic observes the same Execute/exit boundary as
// cmd/moai/main.go, in a separately reaped process with an isolated home.
func TestCGRetiredExecutableDiagnostic(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"cg"}, {"cg", "--help"}, {"cg", "--spawn", "-f", "2"}} {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, executable, append([]string{"-test.run=^TestCGRetiredExecutableHelper$", "--"}, args...)...)
		home := t.TempDir()
		cmd.Dir = home
		cmd.Env = append(os.Environ(), "MOAI_CG_EXEC_HELPER=1", "HOME="+home, "MOAI_HOME="+home, "NO_COLOR=1")
		var out, errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut
		err = cmd.Run()
		cancel()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			t.Fatalf("%v exit: %v", args, err)
		}
		if out.Len() != 0 || strings.Count(errOut.String(), "moai migrate cg") != 1 {
			t.Fatalf("%v stdout=%q stderr=%q", args, out.String(), errOut.String())
		}
	}
}

func TestCGRetiredExecutableHelper(t *testing.T) {
	if os.Getenv("MOAI_CG_EXEC_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"moai"}, os.Args[i+1:]...)
			break
		}
	}
	if err := Execute(); err != nil {
		if code, ok := ResolveExitCode(err); ok {
			os.Exit(code)
		}
		os.Exit(1)
	}
	os.Exit(0)
}
