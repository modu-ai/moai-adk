// Package cli — update_repair_test.go
//
// Regression guards for four `moai update` defects (card t587):
//
//   - F2: the recovery hint printed after an update names a flag the update
//     command does not define, so the advertised recovery path always fails.
//   - F5: the update render contexts never carry the project's git mode, so a
//     team or personal project is rendered as manual whenever a file is
//     regenerated.
//   - F6: the managed cleanup removes .claude/skills/moai* before the legacy
//     skills under that glob are archived, so the archive finds nothing.
//   - F7: the `update -c` model-policy write discards system.yaml read, parse,
//     and write errors, and a parse failure overwrites the unreadable file.
//
// userHomeDirFn and the working directory are process-global, so no test in
// this file calls t.Parallel().

package cli

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// recoverHintPattern captures the long flag a recovery hint tells the user to run.
var recoverHintPattern = regexp.MustCompile(`Recover: moai update --([a-z][a-z-]*)`)

// isolateHome redirects userHomeDirFn into a temporary directory and stops the
// test before anything runs if that directory is the operator's real home. The
// update paths below edit ~/.claude through the seam, so a seam that resolved
// to the real home would write the operator's global settings.
func isolateHome(t *testing.T) {
	t.Helper()
	sentinel, _ := homeSeamSpy(t)
	seamHome, err := userHomeDirFn()
	if err != nil {
		t.Fatalf("seam home: %v", err)
	}
	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("real home: %v", err)
	}
	if seamHome != sentinel || filepath.Clean(seamHome) == filepath.Clean(realHome) {
		t.Fatalf("home seam is not isolated: seam %q, sentinel %q, real home %q", seamHome, sentinel, realHome)
	}
}

// wantRecoverHintSites is the number of recovery-hint literals in the update
// sources: one in the TUX outcome renderer and three in report.RenderOutcome.
const wantRecoverHintSites = 4

// TestUpdateRepair_RecoverHintFlagExists pins every recovery hint to a flag the
// update command actually defines, both in rendered output and in the source.
func TestUpdateRepair_RecoverHintFlagExists(t *testing.T) {
	isolateHome(t)
	const backupPath = "/backups/20260910"

	assertFlag := func(t *testing.T, where, text string) {
		t.Helper()
		m := recoverHintPattern.FindStringSubmatch(text)
		if m == nil {
			t.Fatalf("%s: no recovery hint rendered (reachability) in:\n%s", where, text)
		}
		if updateCmd.Flags().Lookup(m[1]) == nil {
			t.Errorf("%s: recovery hint names --%s, which `moai update` does not define", where, m[1])
		}
	}

	t.Run("rendered_report_outcomes", func(t *testing.T) {
		for _, kind := range []report.OutcomeKind{report.OutcomeAlreadyUpToDate, report.OutcomeUpdatedFiles, report.OutcomeDryRun} {
			assertFlag(t, "report.RenderOutcome("+kind.String()+")", report.RenderOutcome(kind, 2, backupPath))
		}
	})

	t.Run("rendered_tux_outcome", func(t *testing.T) {
		var buf bytes.Buffer
		renderUpdateOutcome(&buf, 2, updateOutcomeDetail{}, backupPath, tui.LightTheme())
		assertFlag(t, "renderUpdateOutcome", stripSGR(buf.String()))
	})

	// Source sweep: the rendered checks above only reach the renderers named
	// there. Counting the literals in the package tree catches a hint added
	// elsewhere, and the exact count keeps an empty sweep from passing.
	t.Run("source_literals", func(t *testing.T) {
		type hit struct{ file, flag string }
		var hits []hit
		err := filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, m := range recoverHintPattern.FindAllStringSubmatch(string(data), -1) {
				hits = append(hits, hit{file: path, flag: m[1]})
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk package sources: %v", err)
		}
		if len(hits) != wantRecoverHintSites {
			t.Fatalf("found %d recovery-hint literals, want %d: %v", len(hits), wantRecoverHintSites, hits)
		}
		for _, h := range hits {
			if updateCmd.Flags().Lookup(h.flag) == nil {
				t.Errorf("%s: recovery hint names --%s, which `moai update` does not define", h.file, h.flag)
			}
		}
	})
}

// syncFixture builds a minimal project the template sync can run against, with
// the working directory and home seam redirected into temporary directories.
func syncFixture(t *testing.T) (root string, cmd *cobra.Command, out *bytes.Buffer) {
	t.Helper()
	isolateHome(t)

	root = t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml", "system:\n  template_version: \"0.0.0\"\n")
	writeTestFile(t, root, ".moai/manifest.json", "{}\n")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if chErr := os.Chdir(root); chErr != nil {
		t.Fatalf("chdir to fixture: %v", chErr)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	cmd = &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("config", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	_ = cmd.Flags().Set("yes", "true")
	_ = cmd.Flags().Set("no-hooks", "true")
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	return root, cmd, out
}

// TestUpdateRepair_RenderUsesProjectGitMode asserts that a team project is
// rendered as team on both update render paths.
func TestUpdateRepair_RenderUsesProjectGitMode(t *testing.T) {
	const teamStrategy = "git_strategy:\n  mode: team\n"

	// The template sync removes .moai/config before it deploys, so the mode has
	// to be read before that removal. settings.json is absent from the fixture,
	// which leaves the rendered file unmerged: what lands is the render itself.
	t.Run("template_sync", func(t *testing.T) {
		root, cmd, out := syncFixture(t)
		writeTestFile(t, root, ".moai/config/sections/git-strategy.yaml", teamStrategy)

		if err := runTemplateSyncWithReporter(cmd, nil, true); err != nil {
			t.Fatalf("runTemplateSyncWithReporter: %v\noutput: %s", err, out.String())
		}

		data, err := os.ReadFile(filepath.Join(root, defs.ClaudeDir, defs.SettingsJSON))
		if err != nil {
			t.Fatalf("read rendered settings.json: %v", err)
		}
		settings := string(data)
		// Reachability: this line is rendered for every git mode, so its absence
		// means the file or its shape changed, not the mode.
		if !strings.Contains(settings, `"Bash(git add:*)"`) {
			t.Fatalf("rendered settings.json lacks the mode-independent git permission; fixture did not render as expected")
		}
		for _, perm := range []string{`"Bash(git commit:*)"`, `"Bash(git push:*)"`} {
			if !strings.Contains(settings, perm) {
				t.Errorf("rendered settings.json lacks %s: the team project was rendered with git mode manual", perm)
			}
		}
	})

	t.Run("clean_reinstall", func(t *testing.T) {
		isolateHome(t)
		root := makeScenarioA(t)
		writeTestFile(t, root, ".moai/config/sections/git-strategy.yaml", teamStrategy)
		deployer := &stubDeployer{}

		if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
			Out:              io.Discard,
			Deployer:         deployer,
			RunMigrateAgency: (&stubMigrateRunner{}).Run,
		}); err != nil {
			t.Fatalf("runCleanReinstall: %v", err)
		}
		if deployer.lastTmplCtx == nil {
			t.Fatalf("stub deployer captured no TemplateContext; the deploy step did not run")
		}
		if got := deployer.lastTmplCtx.GitMode; got != "team" {
			t.Errorf("TemplateContext.GitMode = %q, want %q from git-strategy.yaml", got, "team")
		}
	})
}

// TestUpdateRepair_SyncArchivesLegacySkillsBeforeCleanup asserts that a legacy
// skill present before the template sync is archived by the time the sync —
// whose cleanup removes .claude/skills/moai* — has finished.
func TestUpdateRepair_SyncArchivesLegacySkillsBeforeCleanup(t *testing.T) {
	const skillID = "moai-domain-mobile"
	const userCopy = "user-customized legacy skill\n"
	if !slices.Contains(legacySkillIDs, skillID) {
		t.Fatalf("%s is no longer a legacy skill id; the fixture exercises nothing", skillID)
	}

	root, cmd, out := syncFixture(t)
	writeTestFile(t, root, ".claude/skills/"+skillID+"/SKILL.md", userCopy)

	if err := runTemplateSyncWithReporter(cmd, nil, true); err != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v\noutput: %s", err, out.String())
	}

	archived := filepath.Join(root, ".moai", "archive", "skills", archiveVersion, skillID, "SKILL.md")
	data, err := os.ReadFile(archived)
	if err != nil {
		t.Fatalf("legacy skill was not archived before the cleanup removed it: %v\noutput: %s", err, out.String())
	}
	if string(data) != userCopy {
		t.Errorf("archived SKILL.md = %q, want the user's copy %q", data, userCopy)
	}
}

// TestUpdateRepair_WizardModelPolicySurfacesSystemYAMLErrors asserts that the
// model-policy persistence step reports a system.yaml it cannot use instead of
// returning success.
func TestUpdateRepair_WizardModelPolicySurfacesSystemYAMLErrors(t *testing.T) {
	isolateHome(t)
	result := func() *wizard.WizardResult {
		return &wizard.WizardResult{ModelPolicy: string(template.ModelPolicyHigh)}
	}

	t.Run("unparseable_file_is_reported_and_kept", func(t *testing.T) {
		root := setupSectionsDir(t)
		systemPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML)
		const broken = "moai: [unclosed\n"
		if err := os.WriteFile(systemPath, []byte(broken), defs.FilePerm); err != nil {
			t.Fatalf("write fixture: %v", err)
		}

		if err := applyWizardConfig(root, result()); err == nil {
			t.Errorf("applyWizardConfig returned nil for an unparseable system.yaml")
		}
		if data, _ := os.ReadFile(systemPath); string(data) != broken {
			t.Errorf("system.yaml was rewritten after a parse failure: %q", data)
		}
	})

	t.Run("unwritable_path_is_reported", func(t *testing.T) {
		root := setupSectionsDir(t)
		systemPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML)
		// A directory at the file's path fails both the read and the write on
		// every platform.
		if err := os.MkdirAll(systemPath, defs.DirPerm); err != nil {
			t.Fatalf("create fixture dir: %v", err)
		}

		if err := applyWizardConfig(root, result()); err == nil {
			t.Errorf("applyWizardConfig returned nil when system.yaml could not be read or written")
		}
	})
}
