package cli

// m7_gate51_test.go — gate round 51:
//
//   - 51-2 [P2]: the checkpoint OUTLIVES the asset install — a transient
//     failure in a LATER post-step must leave the next run a resume route
//     (the resume-pending marker), and the marker retires on completion.
//   - 51-3 [P2]: the deployed LAYOUT is the interrupted --llm record — a
//     resume over a deployed GPT layout restores the gpt wiring instead of
//     recording the claude default.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitResumeCheckpointRetiresOnCompletion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte("{ corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	seamsForInit(t)
	projectDir := filepath.Join(t.TempDir(), "marker-proj")

	cmd1 := newInitTestCmd()
	var out1, err1b strings.Builder
	cmd1.SetOut(&out1)
	cmd1.SetErr(&err1b)
	if err := runInit(cmd1, []string{projectDir}); err == nil {
		t.Skip("the ensure-failure injection did not bite")
	}

	// The marker alone routes the half-done resume even after the original
	// evidence (the corrupt manifest) is quarantined away — a LATER-step
	// failure must not wedge behind "already initialized".
	if !initResumeCheckpoint(home, projectDir) {
		t.Fatal("the interruption checkpoint vanished with the evidence it was derived from (gate 51-2 precondition)")
	}

	cmd2 := newInitTestCmd()
	var out2, err2b strings.Builder
	cmd2.SetOut(&out2)
	cmd2.SetErr(&err2b)
	if err := runInit(cmd2, []string{projectDir}); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectDir, ".moai", "resume-pending")); !os.IsNotExist(err) {
		t.Fatalf("the resume-pending marker survived a COMPLETED resume (stat err = %v) — every later run would re-enter the checkpoint (gate 51-2)", err)
	}
}

func TestInitResumeRestoresGPTSelection(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	if err := os.MkdirAll(filepath.Join(home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".moai", "user-assets.json"), []byte("{ corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	seamsForInit(t)
	projectDir := filepath.Join(t.TempDir(), "gpt-resume-proj")

	// Run 1 with --llm gpt: the CODEX-only deployer lays out the project
	// (AGENTS.md, no CLAUDE.md) and dies at the user-asset ensure.
	cmd1 := newInitTestCmd()
	if err := cmd1.Flags().Set("llm", "gpt"); err != nil {
		t.Fatal(err)
	}
	var out1, err1b strings.Builder
	cmd1.SetOut(&out1)
	cmd1.SetErr(&err1b)
	if err := runInit(cmd1, []string{projectDir}); err == nil {
		t.Skip("the ensure-failure injection did not bite")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "AGENTS.md")); err != nil {
		t.Skipf("run 1 did not deploy the gpt layout (AGENTS.md missing): %v", err)
	}

	// Run 2 WITHOUT options: the deployed layout restores the gpt wiring —
	// the recorded harness must be gpt, not the claude default.
	cmd2 := newInitTestCmd()
	var out2, err2b strings.Builder
	cmd2.SetOut(&out2)
	cmd2.SetErr(&err2b)
	if err := runInit(cmd2, []string{projectDir}); err != nil {
		t.Fatalf("run 2 (the resume) failed: %v", err)
	}
	llm, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "llm.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(llm), `harness: "gpt"`) && !strings.Contains(string(llm), "harness: gpt") {
		t.Fatalf("the resume recorded the claude default over a deployed GPT layout (gate 51-3) — llm.yaml:\n%s", llm)
	}
}
