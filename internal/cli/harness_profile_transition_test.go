package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

const (
	transitionUserRole   = ".codex/agents/user-reviewer.toml"
	transitionUserRoleV  = "name = \"user-reviewer\"\ndescription = \"mine\"\n"
	transitionUserConfig = "model = \"gpt-5\"\n\n[mcp_servers.other]\ncommand = \"other\"\n"
	transitionBadHooks   = "{ \"hooks\": [ this is not json\n"
	transitionRetired    = ".codex/agents/moai/retired-role.toml"
)

// runUpdatePathAt drives the update path's two stages that touch .codex/ —
// the template sync (backup, managed-path clean, deploy, restore) and the
// Codex wiring refresh — in root, the way runUpdate calls them. It returns
// stdout and stderr together.
func runUpdatePathAt(t *testing.T, root string) string {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", true, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("no-hooks", true, "")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(context.Background())
	if err := runTemplateSyncWithReporter(cmd, nil, true); err != nil {
		t.Fatalf("template sync: %v\n%s", err, buf.String())
	}
	refreshCodexWiringBestEffort(&buf, &buf)
	return buf.String()
}

// codexSnapshot hashes every regular file under .codex/.
func codexSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	base := filepath.Join(root, ".codex")
	err := filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = wiringSHA(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// simulateOlderDeployment makes the project look deployed by an older binary:
// wiring files without part records (the sidecar is the only trace), and a
// .codex/ template file the newer templates no longer ship.
func simulateOlderDeployment(t *testing.T, root string) {
	t.Helper()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatal(err)
	}
	mf := mgr.Manifest()
	delete(mf.Files, codexwiring.HooksRelPath)
	delete(mf.Files, codexwiring.ConfigRelPath)
	body := "name = \"retired-role\"\n"
	writeTestFile(t, root, transitionRetired, body)
	mf.Files[transitionRetired] = manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: wiringSHA([]byte(body)), DeployedHash: wiringSHA([]byte(body)), CurrentHash: wiringSHA([]byte(body))}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, codexwiring.SidecarPath, "{}\n")
}

// TestHarnessProfileTransitionPreservesUserData covers AC-DHR-005
// (REQ-DHR-007). For each harness profile transition applied through the
// update path: user files and parts under .codex/ stay byte-identical, a
// corrupt hooks.json is left alone, wiring the target profile no longer uses
// is reported with `moai tool disable codex` and left byte-identical, and a
// .codex/ template path the target no longer deploys is reported and left in
// place. The .claude/ managed roots are not asserted (spec §F).
func TestHarnessProfileTransitionPreservesUserData(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	rows := []struct {
		name, from, to string
		older          bool
		blocked        string
	}{
		{"claude_to_both", "claude", "both", false, ""},
		{"gpt_to_both", "gpt", "both", false, ""},
		{"both_to_claude", "both", "claude", false, ""},
		{"both_to_gpt", "both", "gpt", false, ""},
		{"gpt_to_claude", "gpt", "claude", false, ""},
		{"older_binary_to_newer", "both", "both", true, ""},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			if row.blocked != "" {
				t.Skip(row.blocked)
			}
			root := filepath.Join(t.TempDir(), "proj")
			writeTestFile(t, root, transitionUserRole, transitionUserRoleV)
			writeTestFile(t, root, codexwiring.ConfigRelPath, transitionUserConfig)
			writeTestFile(t, root, codexwiring.HooksRelPath, transitionBadHooks)
			// The deployment the transition starts from: `moai init` is the
			// path that records the deployed templates in the manifest.
			initProfileAt(t, root, row.from)
			if row.older {
				simulateOlderDeployment(t, root)
			}
			if err := template.ApplyHarness(root, row.to); err != nil {
				t.Fatal(err)
			}
			before := codexSnapshot(t, root)
			// SPEC-USER-ASSET-INSTALL-001 (REQ-005): the project tree no
			// longer carries .codex/agents/moai/ templates, so the
			// undeployed-templates half of the fixture has no material in
			// the post-M4 shape — the preservation invariant this row once
			// exercised there now rides the older_binary arm's
			// manifest-seeded retired entry, and the claude-target arms
			// keep their orphaned-wiring assertions below.
			undeployed := []string{}
			if row.older {
				undeployed = []string{transitionRetired}
			}
			cfgBefore := string(readTestFile(t, root, codexwiring.ConfigRelPath))

			out := runUpdatePathAt(t, root)
			after := codexSnapshot(t, root)

			if got := string(readTestFile(t, root, transitionUserRole)); got != transitionUserRoleV {
				t.Errorf("user role file changed: %q", got)
			}
			if got := string(readTestFile(t, root, codexwiring.HooksRelPath)); got != transitionBadHooks {
				t.Errorf("corrupt hooks.json was touched: %q", got)
			}
			cfgAfter := string(readTestFile(t, root, codexwiring.ConfigRelPath))
			if !strings.HasPrefix(cfgAfter, transitionUserConfig) {
				t.Errorf("user config parts changed:\n%q", cfgAfter)
			}
			for _, rel := range undeployed {
				if before[rel] == "" || after[rel] != before[rel] {
					t.Errorf("%s not preserved byte-identical (before=%q after=%q)", rel, before[rel], after[rel])
				}
				if !strings.Contains(out, rel+" is no longer deployed") {
					t.Errorf("undeployed %s not reported", rel)
				}
			}
			if row.to == "claude" {
				if cfgAfter != cfgBefore {
					t.Errorf("orphaned wiring rewritten:\nbefore=%q\nafter =%q", cfgBefore, cfgAfter)
				}
				// SPEC-USER-ASSET-INSTALL-001 (REQ-005): the .codex/agents/
				// moai/ TOML placement no longer deploys, so deployments
				// carry ONLY GeneratedManaged wiring entries — the
				// orphaned-wiring report (keyed on non-GeneratedManaged
				// .codex/ entries) legitimately stays silent in every arm
				// of this table now; the older_binary arm's retired entry
				// is still covered by the undeployed loop's preservation
				// assertions above. Wiring files stay byte-identical.
			} else if strings.Contains(out, codexwiring.DisableCommand) {
				t.Errorf("wiring the %s profile uses was reported orphaned:\n%s", row.to, lastLines(out, 20))
			}
		})
	}
}

// initProfileAt runs a non-interactive `moai init` of root under harness.
func initProfileAt(t *testing.T, root, harness string) {
	t.Helper()
	cmd := newInitTestCmd()
	for name, val := range map[string]string{"llm": harness, "non-interactive": "true", "name": "transition", "language": "go", "mode": "tdd"} {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if err := runInit(cmd, []string{root}); err != nil {
		t.Fatalf("init --llm %s: %v (stderr: %s)", harness, err, errBuf.String())
	}
}

func readTestFile(t *testing.T, root, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
