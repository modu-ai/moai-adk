package cli

// t583 reproduction probes (plan-phase evidence, NOT a landed test). They run
// the real init code paths with the home directory redirected through the
// package seams (userHomeDirFn, profile.BaseDirOverride, MOAI_HOME) instead of
// HOME, and cross-check the operator's real ~/.claude state before and after
// so any leak past the seams is recorded rather than assumed absent.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// realHomeFingerprint records the operator's real ~/.claude/settings.json hash
// and whether ~/.claude/hooks/moai exists. Read-only.
func realHomeFingerprint(t *testing.T) (settingsHash string, hooksDirPresent bool) {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("resolve real home: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json")); err == nil {
		sum := sha256.Sum256(data)
		settingsHash = hex.EncodeToString(sum[:])
	} else {
		settingsHash = "absent"
	}
	_, statErr := os.Stat(filepath.Join(home, ".claude", "hooks", "moai"))
	return settingsHash, statErr == nil
}

// runInitT583 runs runInit in a temp project with every home seam redirected.
// A nil wiz takes the --non-interactive path; a non-nil wiz swaps the wizard
// seams so the interactive branch runs with that result.
func runInitT583(t *testing.T, wiz *wizard.WizardResult, flags map[string]string) (projectDir, stdout, stderr string) {
	t.Helper()

	beforeHash, beforeHooks := realHomeFingerprint(t)
	t.Cleanup(func() {
		afterHash, afterHooks := realHomeFingerprint(t)
		t.Logf("real-home guard: settings.json sha256 before=%s after=%s; hooks/moai before=%t after=%t",
			beforeHash, afterHash, beforeHooks, afterHooks)
		if afterHash != beforeHash || afterHooks != beforeHooks {
			t.Errorf("real home state changed during the probe (seam leak)")
		}
	})

	fakeHome := t.TempDir()
	origHome := userHomeDirFn
	userHomeDirFn = func() (string, error) { return fakeHome, nil }
	t.Cleanup(func() { userHomeDirFn = origHome })

	origBase := profile.BaseDirOverride
	profile.BaseDirOverride = filepath.Join(fakeHome, ".moai", "claude-profiles")
	t.Cleanup(func() { profile.BaseDirOverride = origBase })

	t.Setenv("MOAI_HOME", filepath.Join(fakeHome, ".moai"))
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	// Pre-run self-guard: refuse to call runInit unless every redirected home
	// resolves away from the real home. A post-run hash comparison alone is too
	// late to prevent a write.
	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("resolve real home: %v", err)
	}
	seamHome, err := userHomeDirFn()
	if err != nil {
		t.Fatalf("seam home: %v", err)
	}
	for label, p := range map[string]string{
		"userHomeDirFn":           seamHome,
		"profile.BaseDirOverride": profile.GetBaseDir(),
		"MOAI_HOME":               os.Getenv("MOAI_HOME"),
	} {
		if p == realHome || strings.HasPrefix(p, realHome+string(filepath.Separator)) {
			t.Fatalf("seam guard: %s resolves inside the real home (%s); refusing to run init", label, p)
		}
	}

	if wiz != nil {
		origInteractive := isInteractiveStdin
		isInteractiveStdin = func() bool { return true }
		t.Cleanup(func() { isInteractiveStdin = origInteractive })

		origDeps := deps
		deps = nil
		t.Cleanup(func() { deps = origDeps })

		origWizard := runWizardFn
		runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) { return wiz, nil }
		t.Cleanup(func() { runWizardFn = origWizard })
	}

	projectDir = filepath.Join(t.TempDir(), "t583-proj")
	cmd := newInitTestCmd()
	for name, val := range flags {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
	}
	return projectDir, out.String(), errBuf.String()
}

func readWorkflowYAML(t *testing.T, projectDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		t.Fatalf("read workflow.yaml: %v", err)
	}
	return string(data)
}

// TestT583Repro_F1_Narrow: the wizard's "yes" to worktree_auto_create through
// the real opts mapping and the real tracker-gated writer. No home access.
func TestT583Repro_F1_Narrow(t *testing.T) {
	cmd := newInitTestCmd()
	opts := project.InitOptions{}
	applyWizardPage3ToOpts(cmd, &wizard.WizardResult{WorktreeAutoCreate: true}, &opts)
	t.Logf("after applyWizardPage3ToOpts: WorktreeAutoCreate=%t WorktreeAutoCreateSet=%t",
		opts.WorktreeAutoCreate, opts.WorktreeAutoCreateSet)

	sectionsDir := t.TempDir()
	path := filepath.Join(sectionsDir, "workflow.yaml")
	if err := os.WriteFile(path, embeddedTemplateWorkflow(t), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := project.WriteWorkflowTogglesYAML(sectionsDir, opts, &project.InitResult{}); err != nil {
		t.Fatalf("WriteWorkflowTogglesYAML: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "auto_create: true") {
		t.Errorf("F1 reproduced (narrow): wizard answered yes, workflow.yaml auto_create is not true")
	}
}

// TestT583Repro_F1_RunInit: the same "yes" through the full interactive runInit.
func TestT583Repro_F1_RunInit(t *testing.T) {
	projectDir, _, _ := runInitT583(t, &wizard.WizardResult{WorktreeAutoCreate: true, MCPProvision: true}, nil)
	wf := readWorkflowYAML(t, projectDir)
	for _, line := range strings.Split(wf, "\n") {
		if strings.Contains(line, "auto_create:") {
			t.Logf("deployed workflow.yaml line: %q", line)
		}
	}
	if !strings.Contains(wf, "auto_create: true") {
		t.Errorf("F1 reproduced (runInit): wizard answered yes, deployed auto_create is not true")
	}
}

// TestT583Repro_F4_NonInteractive measures what a --non-interactive init leaves
// in .mcp.json and whether the ensure-entry call announced itself.
func TestT583Repro_F4_NonInteractive(t *testing.T) {
	projectDir, stdout, _ := runInitT583(t, nil, map[string]string{"non-interactive": "true"})

	announced := strings.Contains(stdout, "Provisioned the moai MCP server entry in .mcp.json")
	t.Logf("ensure-entry announcement on stdout: %t", announced)

	data, err := os.ReadFile(filepath.Join(projectDir, ".mcp.json"))
	if err != nil {
		t.Logf(".mcp.json read error: %v", err)
		t.Errorf("F4 reproduced: no .mcp.json after --non-interactive init")
		return
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}
	servers, _ := root["mcpServers"].(map[string]any)
	names := make([]string, 0, len(servers))
	for k := range servers {
		names = append(names, k)
	}
	t.Logf(".mcp.json mcpServers keys: %v", names)
	if _, ok := servers["moai"]; !ok {
		t.Errorf("F4 reproduced: --non-interactive init left no moai entry in .mcp.json")
	}
}
