package cli

// init_mode_test.go — the M2 init-mode criteria (SPEC-INIT-SHRINK-001
// acceptance.md AC-001 (a), AC-003, AC-004, AC-005, AC-007). Every run is a
// non-interactive (or wizard-injected) init into a scratch project; no test
// reaches a real tool (the default runner refuses under a test binary, and
// the confirmed arm injects a fake).

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// initModeTestCmd mirrors the production initCmd flag surface for the mode
// tests, including --no-plugin (which newInitTestCmd predates).
func initModeTestCmd() *cobra.Command {
	cmd := newInitTestCmd()
	cmd.Flags().Bool("no-plugin", false, "")
	return cmd
}

// runInitForMode runs init into projectDir with the given flag overrides and
// returns root, stdout, stderr.
func runInitForMode(t *testing.T, flags map[string]string) (string, string, string) {
	t.Helper()
	projectDir := filepath.Join(t.TempDir(), "mode-proj")
	cmd := initModeTestCmd()
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

// runInitInteractiveForMode is runInitForMode with the wizard injected (the
// interactive path — the only one where opts.MCPProvision is true).
func runInitInteractiveForMode(t *testing.T, flags map[string]string) (string, string, string) {
	t.Helper()
	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })
	return runInitForMode(t, flags)
}

// readProjectMcpJSON reads the project's .mcp.json into a generic map.
func readProjectMcpJSON(t *testing.T, root string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".mcp.json"))
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}
	return doc
}

// mcpServersOf extracts the mcpServers object.
func mcpServersOf(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	servers, ok := doc["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf(".mcp.json carries no mcpServers object: %v", doc)
	}
	return servers
}

// assertDeployedAbsent asserts none of the dropped component roots exists.
func assertDeployedAbsent(t *testing.T, root string) {
	t.Helper()
	for _, rel := range []string{".claude/skills", ".claude/commands"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Errorf("plugin deploy wrote dropped root %s (stat err: %v)", rel, err)
		}
	}
}

// assertDeployedPresent asserts the kept components a thin deploy still
// carries.
func assertDeployedPresent(t *testing.T, root string) {
	t.Helper()
	for _, rel := range []string{
		".claude/agents",
		".claude/rules",
		"CLAUDE.md",
		"AGENTS.md",
		".moai/config",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("plugin deploy lost kept component %s: %v", rel, err)
		}
	}
}

// TestDefaultDeploySetExcludesSkillsAndCommands is AC-001 (a): the default
// (plugin-path) init deploy carries no .claude/skills or .claude/commands
// file and keeps the instruction files, rules, agents, settings render, and
// .moai/config.
func TestDefaultDeploySetExcludesSkillsAndCommands(t *testing.T) {
	root, _, _ := runInitForMode(t, map[string]string{"non-interactive": "true"})
	assertDeployedAbsent(t, root)
	assertDeployedPresent(t, root)
	// The deploy-mode record follows the deploy path (REQ-009), never the
	// probe.
	if got := readDeployModeForTest(t, root); got != "plugin" {
		t.Errorf("deployment_mode = %q, want plugin", got)
	}
}

// TestNoPluginPathDeploysFullLocalPayload is AC-003: the opt-out deploy
// equals today's payload — skills, commands, the project .mcp.json moai
// entry (from the unstripped render), and the Codex mirror included.
func TestNoPluginPathDeploysFullLocalPayload(t *testing.T) {
	root, _, _ := runInitForMode(t, map[string]string{
		"non-interactive": "true",
		"no-plugin":       "true",
	})
	for _, rel := range []string{".claude/skills", ".claude/commands"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("--no-plugin deploy lost %s: %v", rel, err)
		}
	}
	servers := mcpServersOf(t, readProjectMcpJSON(t, root))
	if _, ok := servers["moai"]; !ok {
		t.Error("--no-plugin .mcp.json lost the moai entry")
	}
	if got := readDeployModeForTest(t, root); got != "local" {
		t.Errorf("deployment_mode = %q, want local", got)
	}
}

// TestShrinkInitGuidanceOnMissingPlugin is AC-004: on the default path with
// the probe reading not-demonstrated (the default runner's refusal under a
// test binary degrades every surface read), exactly one guidance block
// names both recourses and the init exit status stays 0.
func TestShrinkInitGuidanceOnMissingPlugin(t *testing.T) {
	root, _, stderr := runInitInteractiveForMode(t, nil)
	if got := strings.Count(stderr, "note: the moai plugin install could not be demonstrated"); got != 1 {
		t.Fatalf("guidance block count = %d, want 1\nstderr:\n%s", got, stderr)
	}
	if !strings.Contains(stderr, "--no-plugin") {
		t.Errorf("guidance does not name the --no-plugin recourse:\n%s", stderr)
	}
	// Card t1438 review finding 6: a plain re-run fails "project already
	// initialized" — the guidance must name the flags that actually work
	// (--no-plugin --force) and state what force re-initialization does with
	// the existing .moai/.
	if !strings.Contains(stderr, "--no-plugin --force") {
		t.Errorf("guidance does not name the working re-entry flags (--no-plugin --force):\n%s", stderr)
	}
	if !strings.Contains(stderr, ".moai-backups") {
		t.Errorf("guidance does not state the --force backup disposition of the existing .moai/:\n%s", stderr)
	}
	if !strings.Contains(stderr, "plugin marketplace add") {
		t.Errorf("guidance does not name the manual install commands:\n%s", stderr)
	}
	// The guidance names both recourses; the record still reads plugin (the
	// deploy path decides the record).
	if got := readDeployModeForTest(t, root); got != "plugin" {
		t.Errorf("deployment_mode = %q, want plugin", got)
	}
}

// TestDefaultPathMcpEntryPolicy is AC-005: context7 and staggeredStartup
// always present; the moai entry follows the probe — absent on confirmed,
// present on not-demonstrated and on the --no-plugin path.
func TestDefaultPathMcpEntryPolicy(t *testing.T) {
	t.Run("not-demonstrated-entry-present", func(t *testing.T) {
		root, _, _ := runInitInteractiveForMode(t, nil)
		doc := readProjectMcpJSON(t, root)
		servers := mcpServersOf(t, doc)
		if _, ok := servers["moai"]; !ok {
			t.Error("not-demonstrated fallback did not write the moai entry")
		}
		if _, ok := servers["context7"]; !ok {
			t.Error("context7 lost from .mcp.json")
		}
		if _, ok := doc["staggeredStartup"]; !ok {
			t.Error("staggeredStartup lost from .mcp.json")
		}
	})

	t.Run("confirmed-entry-absent", func(t *testing.T) {
		// The fake lists the ref ONLY on post-execution reads: the pre
		// snapshot is the first `claude plugin list`, the post read the
		// second. The step's own add+install calls answer success.
		fake := &fakePluginRunner{}
		claudeLists := 0
		fake.fn = func(_ context.Context, call pluginCall, _ int) ([]byte, error) {
			if filepath.Base(call.bin) == "claude-pinned" && len(call.args) >= 2 &&
				call.args[0] == "plugin" && call.args[1] == "list" {
				claudeLists++
				if claudeLists >= 2 {
					return []byte(`moai@moai-adk 1.0.0`), nil
				}
				return []byte("(no plugins installed)"), nil
			}
			return nil, nil
		}
		withPluginRunner(t, fake)
		setupPluginTools(t)

		root, _, _ := runInitInteractiveForMode(t, nil)
		servers := mcpServersOf(t, readProjectMcpJSON(t, root))
		if _, ok := servers["moai"]; ok {
			t.Error("confirmed install still wrote the project moai entry")
		}
		if _, ok := servers["context7"]; !ok {
			t.Error("context7 lost from .mcp.json on the confirmed path")
		}
	})

	t.Run("no-plugin-entry-present", func(t *testing.T) {
		root, _, _ := runInitForMode(t, map[string]string{
			"non-interactive": "true",
			"no-plugin":       "true",
		})
		servers := mcpServersOf(t, readProjectMcpJSON(t, root))
		if _, ok := servers["moai"]; !ok {
			t.Error("--no-plugin path lost the moai entry")
		}
	})
}

// TestAllFlagDeploysAllTiersLocally is AC-007 (OD-7 settled (a)): --all is
// the local full deploy — the --no-plugin payload plus the wider tier — and
// the record reads local.
func TestAllFlagDeploysAllTiersLocally(t *testing.T) {
	root, _, _ := runInitForMode(t, map[string]string{
		"non-interactive": "true",
		"all":             "true",
	})
	for _, rel := range []string{".claude/skills", ".claude/commands"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("--all deploy lost %s: %v", rel, err)
		}
	}
	// The wider tier: an optional-pack catalog entry the slim deploy hides.
	cat, catErr := template.LoadEmbeddedCatalog()
	if catErr != nil {
		t.Fatalf("load catalog: %v", catErr)
	}
	optionalPack := 0
	for _, entry := range cat.AllEntries() {
		if entry.Tier == "core" || entry.Tier == "harness-generated" {
			continue
		}
		optionalPack++
		skillFile := filepath.Join(root, ".claude", "skills", entry.Name, "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			t.Errorf("--all deploy lost optional-pack entry %s (%s): %v", entry.Name, entry.Tier, err)
		}
	}
	if optionalPack == 0 {
		t.Fatal("no optional-pack entries to probe (catalog changed?)")
	}
	if got := readDeployModeForTest(t, root); got != "local" {
		t.Errorf("deployment_mode = %q, want local", got)
	}
}

// readDeployModeForTest reads the project's deployment_mode through the
// production reader.
func readDeployModeForTest(t *testing.T, root string) string {
	t.Helper()
	return config.ReadDeployMode(root)
}
