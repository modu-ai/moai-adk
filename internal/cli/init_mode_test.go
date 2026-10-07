package cli

// init_mode_test.go — the M2 init-mode criteria (SPEC-INIT-SHRINK-001
// acceptance.md AC-001 (a), AC-003, AC-004, AC-005, AC-007). Every run is a
// non-interactive (or wizard-injected) init into a scratch project; no test
// reaches a real tool (the default runner refuses under a test binary, and
// the confirmed arm injects a fake).

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
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
	expected := initModeExpectedManifest(t, bundles)
	if err := checkInitModeManifest(home, bundles, expected); err != nil {
		t.Error(err)
	}
}

func checkInitModeManifest(home string, bundles []string, expected map[string]map[string]bool) error {
	if _, err := os.Stat(userassets.ManifestPath(home)); err != nil {
		return fmt.Errorf("installed manifest missing: %w", err)
	}
	manifest, err := userassets.Load(userassets.ManifestPath(home))
	if err != nil {
		return err
	}
	selected := map[string]bool{}
	for _, bundle := range bundles {
		selected[bundle] = true
	}
	if len(manifest.Bundles) != len(selected) {
		return fmt.Errorf("recorded bundles=%v, want %v", manifest.Bundles, bundles)
	}
	for _, bundle := range manifest.Bundles {
		if !selected[bundle] {
			return fmt.Errorf("unexpected bundle opt-in %q", bundle)
		}
	}
	if len(expected) == 0 || len(manifest.Files) != len(expected) {
		return fmt.Errorf("manifest files=%d, want independent catalog keys=%d", len(manifest.Files), len(expected))
	}
	roots := map[string]string{
		"claude-skills": filepath.Join(home, ".claude", "skills"),
		"agents-skills": filepath.Join(home, ".agents", "skills"),
		"claude-agents": filepath.Join(home, ".claude", "agents"),
		"codex-agents":  filepath.Join(home, ".codex", "agents"),
	}
	for key, owners := range expected {
		file, ok := manifest.Files[key]
		if !ok {
			return fmt.Errorf("manifest missing expected asset %s", key)
		}
		if !owners[file.Bundle] {
			return fmt.Errorf("asset %s owner=%q, want catalog membership %v", key, file.Bundle, owners)
		}
		slug, rel, ok := strings.Cut(key, "/")
		if !ok || roots[slug] == "" {
			return fmt.Errorf("invalid expected manifest key %s", key)
		}
		data, err := os.ReadFile(filepath.Join(roots[slug], filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("read installed %s: %w", key, err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(data)); file.SHA256 != got {
			return fmt.Errorf("asset %s recorded hash=%s, actual=%s", key, file.SHA256, got)
		}
	}
	return nil
}

// Derive required files from declared catalog membership and shipped trees,
// independently of Installer targets and the manifest under examination.
// A shared core/optional entry accepts only its actual selected memberships.
func initModeExpectedManifest(t *testing.T, bundles []string) map[string]map[string]bool {
	t.Helper()
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	source, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]map[string]bool{}
	add := func(key, owner string) {
		if expected[key] == nil {
			expected[key] = map[string]bool{}
		}
		expected[key][owner] = true
	}
	groups := map[string][]template.Entry{"core": append(append([]template.Entry{}, cat.Catalog.Core.Skills...), cat.Catalog.Core.Agents...)}
	for _, bundle := range bundles {
		pack, ok := cat.Catalog.OptionalPacks[bundle]
		if !ok {
			t.Fatalf("unknown fixture bundle %q", bundle)
		}
		groups[bundle] = append(append([]template.Entry{}, pack.Skills...), pack.Agents...)
	}
	for owner, entries := range groups {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Path, "/") {
				dir := strings.TrimSuffix(strings.TrimPrefix(entry.Path, "templates/"), "/")
				err := fs.WalkDir(source, dir, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					rel := entry.Name + "/" + strings.TrimPrefix(path, dir+"/")
					add("claude-skills/"+rel, owner)
					add("agents-skills/"+rel, owner)
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			} else if strings.HasSuffix(entry.Path, ".md") {
				add("claude-agents/"+entry.Name+".md", owner)
				add("codex-agents/"+entry.Name+".toml", owner)
			} else {
				t.Fatalf("unsupported catalog entry %+v", entry)
			}
		}
	}
	return expected
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

// Corrupting bookkeeping while installed files remain intact must be observable.
func TestInitModeManifestRejectsMissingOrFalseProvenance(t *testing.T) {
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for name := range cat.Catalog.OptionalPacks {
		all = append(all, name)
	}
	sort.Strings(all)
	for _, bundles := range [][]string{nil, all} {
		name := "core-only"
		if len(bundles) != 0 {
			name = "all-explicit-bundles"
		}
		t.Run(name, func(t *testing.T) {
			home := initModeUserHome(t)
			runInitForMode(t, map[string]string{"non-interactive": "true", "bundles": strings.Join(bundles, ",")})
			manifestPath := userassets.ManifestPath(home)
			original, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			manifest, err := userassets.Load(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			if len(manifest.Files) == 0 {
				t.Fatal("positive control has no installed records")
			}
			expected := initModeExpectedManifest(t, bundles)
			var keys []string
			sharedLabels := map[string]int{}
			for key, owners := range expected {
				keys = append(keys, key)
				if len(owners) > 1 {
					sharedLabels[manifest.Files[key].Bundle]++
				}
			}
			sort.Strings(keys)
			key := keys[0]
			t.Logf("independent keys=%d actual records=%d shared record labels=%v", len(expected), len(manifest.Files), sharedLabels)
			if err := checkInitModeManifest(home, bundles, expected); err != nil {
				t.Fatalf("valid manifest rejected: %v", err)
			}
			mutations := []string{"missing-manifest", "empty-files", "missing-asset-record", "wrong-hash", "wrong-owner"}
			sharedKey := ""
			for _, candidate := range keys {
				if len(expected[candidate]) > 1 {
					sharedKey = candidate
					break
				}
			}
			if len(bundles) > 0 {
				if sharedKey == "" {
					t.Fatal("all-bundles fixture lost shared-asset coverage")
				}
				// Both real memberships are valid; an unrelated selected bundle is not.
				for owner := range expected[sharedKey] {
					file := manifest.Files[sharedKey]
					file.Bundle = owner
					manifest.Files[sharedKey] = file
					if err := manifest.Save(manifestPath); err != nil {
						t.Fatal(err)
					}
					if err := checkInitModeManifest(home, bundles, expected); err != nil {
						t.Fatalf("genuine shared owner %s rejected: %v", owner, err)
					}
				}
				mutations = append(mutations, "wrong-shared-owner")
			}
			for _, mutation := range mutations {
				t.Run(mutation, func(t *testing.T) {
					key := key
					if mutation == "wrong-shared-owner" {
						key = sharedKey
					}
					if err := os.WriteFile(manifestPath, original, 0600); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() {
						if err := os.WriteFile(manifestPath, original, 0600); err != nil {
							t.Error(err)
						}
					})
					m, err := userassets.Load(manifestPath)
					if err != nil {
						t.Fatal(err)
					}
					switch mutation {
					case "missing-manifest":
						err = os.Remove(manifestPath)
					case "empty-files":
						m.Files = map[string]userassets.FileEntry{}
					case "missing-asset-record":
						m.Files["claude-skills/unexpected/SKILL.md"] = m.Files[key]
						delete(m.Files, key)
					case "wrong-hash":
						file := m.Files[key]
						file.SHA256 = strings.Repeat("0", 64)
						m.Files[key] = file
					case "wrong-owner", "wrong-shared-owner":
						file := m.Files[key]
						file.Bundle = "unselected"
						for _, bundle := range bundles {
							if !expected[key][bundle] {
								file.Bundle = bundle
								break
							}
						}
						if len(bundles) > 0 && file.Bundle == "unselected" {
							t.Fatal("no unrelated selected bundle for wrong-owner control")
						}
						m.Files[key] = file
					}
					if err != nil {
						t.Fatal(err)
					}
					if mutation != "missing-manifest" {
						if err := m.Save(manifestPath); err != nil {
							t.Fatal(err)
						}
					}
					if err := checkInitModeManifest(home, bundles, expected); err == nil {
						t.Errorf("accepted %s despite installed files remaining intact", mutation)
					}
				})
			}
		})
	}
}
