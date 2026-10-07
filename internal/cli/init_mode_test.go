package cli

// init_mode_test.go — the M2 init-mode criteria (SPEC-INIT-SHRINK-001
// acceptance.md AC-001 (a), AC-003, AC-004, AC-005, AC-007). Every run is a
// non-interactive (or wizard-injected) init into a scratch project; no test
// reaches a real tool (the default runner refuses under a test binary, and
// the confirmed arm injects a fake).

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
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

// The default deploy keeps project harness files and installs common core assets
// user-side; optional assets require an explicit bundle selection.
func TestDefaultDeploySetUsesUserCoreAssets(t *testing.T) {
	home := initModeUserHome(t)
	root, _, _ := runInitForMode(t, map[string]string{"non-interactive": "true"})
	assertDeployedPresent(t, root)
	assertUserCatalogEntries(t, root, home, nil)
	if got := readDeployModeForTest(t, root); got != "local" {
		t.Errorf("deployment_mode = %q, want local", got)
	}
}

func initModeUserHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDirFn = orig })
	return home
}

// Check actual entry kinds in both harness roots and the recorded selection;
// the same common asset must never also be deployed to the project.
func assertUserCatalogEntries(t *testing.T, project, home string, bundles []string) {
	t.Helper()
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	selected := map[string]bool{}
	for _, b := range bundles {
		selected[b] = true
	}
	entries := append([]template.Entry{}, cat.Catalog.Core.Skills...)
	entries = append(entries, cat.Catalog.Core.Agents...)
	present := map[string]bool{}
	for _, entry := range entries {
		present[entry.Path] = true
	}
	for name, pack := range cat.Catalog.OptionalPacks {
		packEntries := append(append([]template.Entry{}, pack.Skills...), pack.Agents...)
		entries = append(entries, packEntries...)
		if selected[name] {
			for _, entry := range packEntries {
				present[entry.Path] = true
			}
		}
	}
	for _, entry := range entries {
		var paths []string
		if strings.HasSuffix(entry.Path, "/") {
			paths = []string{filepath.Join(".claude", "skills", entry.Name, "SKILL.md"), filepath.Join(".agents", "skills", entry.Name, "SKILL.md")}
		} else if strings.HasSuffix(entry.Path, ".md") {
			paths = []string{filepath.Join(".claude", "agents", entry.Name+".md"), filepath.Join(".codex", "agents", entry.Name+".toml")}
		} else {
			t.Fatalf("unsupported catalog entry kind: %+v", entry)
		}
		for _, rel := range paths {
			_, err := os.Stat(filepath.Join(home, rel))
			if present[entry.Path] && err != nil {
				t.Errorf("user asset %s (%s): %v", rel, entry.Tier, err)
			}
			if !present[entry.Path] && !os.IsNotExist(err) {
				t.Errorf("unselected optional user asset exists %s: %v", rel, err)
			}
		}
		rel := strings.TrimPrefix(entry.Path, "templates/")
		if strings.HasSuffix(rel, "/") {
			rel += "SKILL.md"
		}
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Errorf("common asset duplicated in project %s: %v", rel, err)
		}
	}
	manifest, err := userassets.Load(userassets.ManifestPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Bundles) != len(selected) {
		t.Errorf("recorded bundles=%v, want %v", manifest.Bundles, bundles)
	}
	for _, b := range manifest.Bundles {
		if !selected[b] {
			t.Errorf("unexpected bundle opt-in %q", b)
		}
	}
	for key, file := range manifest.Files {
		if file.Bundle != "core" && !selected[file.Bundle] {
			t.Errorf("unselected optional asset installed: %s (%s)", key, file.Bundle)
		}
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

// Retired plugin probes cannot send users back to marketplace commands.
func TestInitGuidanceUsesLocalAssetsWithoutPluginRecourse(t *testing.T) {
	home := initModeUserHome(t)
	root, stdout, stderr := runInitInteractiveForMode(t, nil)
	for _, retired := range []string{"plugin install", "plugin marketplace add", "--no-plugin --force", "plugin install could not be demonstrated"} {
		if strings.Contains(stdout+stderr, retired) {
			t.Errorf("retired plugin guidance %q: %s", retired, stdout+stderr)
		}
	}
	if !strings.Contains(stdout+stderr, "Deploy mode: local") {
		t.Errorf("missing local deploy guidance: %s", stdout+stderr)
	}
	assertUserCatalogEntries(t, root, home, nil)
	if got := readDeployModeForTest(t, root); got != "local" {
		t.Errorf("deployment_mode = %q, want local", got)
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

// --all selects the project harness templates; it cannot silently opt a user
// into optional bundles. Explicit bundle selection installs every entry kind.
func TestAllFlagRespectsUserBundleSelection(t *testing.T) {
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for name := range cat.Catalog.OptionalPacks {
		all = append(all, name)
	}
	sort.Strings(all)
	if len(all) == 0 {
		t.Fatal("catalog has no optional bundles to exercise")
	}
	for _, bundles := range [][]string{nil, all} {
		name := "core-only"
		if len(bundles) != 0 {
			name = "all-explicit-bundles"
		}
		t.Run(name, func(t *testing.T) {
			home := initModeUserHome(t)
			root, _, _ := runInitForMode(t, map[string]string{"non-interactive": "true", "all": "true", "bundles": strings.Join(bundles, ",")})
			assertDeployedPresent(t, root)
			assertUserCatalogEntries(t, root, home, bundles)
			if got := readDeployModeForTest(t, root); got != "local" {
				t.Errorf("deployment_mode = %q, want local", got)
			}
		})
	}
}

// readDeployModeForTest reads the project's deployment_mode through the
// production reader.
func readDeployModeForTest(t *testing.T, root string) string {
	t.Helper()
	return config.ReadDeployMode(root)
}
