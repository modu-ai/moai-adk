// m0_surface_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0, cli surface
// family: AC-013 (emission uncompared → CheckFail, born-green guard),
// AC-014 (skills disable resolves the user-installed skill), AC-017 (cancel
// keeps project assets — born-green guard), AC-018 (init resume after a
// user-asset ensure failure).
//
// M0 discipline: observation only — no production change. A RED on a
// born-green guard (AC-013/017) is a regression finding for the leader, not
// a fix target.
package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/merge"
)

// TestDoctorAgentEmissionUncomparedNotOK — AC-013 (ledger 5-잔여,
// REQ-SRF-002, born-green regression guard). Given a committed emission set
// whose extracted counterpart is missing one artifact, the check must report
// CheckFail (the cardinality gate at doctor_agentemit_embed.go:150-165). RED
// at HEAD is a regression finding.
func TestDoctorAgentEmissionUncomparedNotOK(t *testing.T) {
	root := t.TempDir()
	committedDir := filepath.Join(root, "internal", "template", "templates", ".codex", "agents", "moai")
	if err := os.MkdirAll(committedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const nameA = "moai-a.toml"
	const nameB = "moai-b.toml"
	for _, name := range []string{nameA, nameB} {
		if err := os.WriteFile(filepath.Join(committedDir, name), []byte("name = \""+strings.TrimSuffix(name, ".toml")+"\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(root, "bin", "moai")
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	// The fake extractor deploys only nameA — nameB has no counterpart.
	extract := func(string) (string, func(), error) {
		dir := t.TempDir()
		out := filepath.Join(dir, ".codex", "agents", "moai")
		if err := os.MkdirAll(out, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, nameA), []byte("name = \"moai-a\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir, func() {}, nil
	}

	check := checkAgentEmitEmbedAgainst(root, bin, extract, false)
	if check.Status != uikit.CheckFail {
		t.Fatalf("RED-guard regression: an uncompared emission artifact produced %s %q — the compared-shortfall must be CheckFail, never OK", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "compared 1/2") {
		t.Errorf("the shortfall message lost its cardinality: %q", check.Message)
	}
}

// TestSkillsDisableResolvesUserInstalledSkill — AC-014 (ledger 8d+5-증상,
// REQ-SRF-003). Given NO project mirror and a USER-installed skill only, the
// disable verb must resolve the user hierarchy face — not answer
// mirror-absent/nothing-to-disable. RED-now reason:
// resolveCodexSkillMirrorPath returns codexSkillMirrorAbsent whenever the
// project mirror is missing (codex_skills_disable.go:145-150); the
// user-installed skill (~/.claude/skills) never participates in resolution.
func TestSkillsDisableResolvesUserInstalledSkill(t *testing.T) {
	project := t.TempDir() // no .agents/skills mirror at all
	home := t.TempDir()
	userSkill := filepath.Join(home, ".claude", "skills", "moai-user-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(userSkill), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userSkill, []byte("---\nname: moai-user-skill\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := resolveCodexSkillMirrorPath(project, home, "moai-user-skill")
	if res.Outcome != codexSkillResolved {
		t.Fatalf("RED (intended): a user-installed skill with no project mirror resolved to outcome %d (%s) — the user hierarchy face is unreachable, the verb answers nothing-to-disable instead", res.Outcome, res.Reason)
	}
	if res.Path == "" {
		t.Fatalf("the resolution carried no path")
	}
}

// TestUpdateCancelKeepsProjectAssetsIntact — AC-017 (ledger 11a,
// REQ-SRF-006, born-green regression guard). Given an update run cancelled
// at the confirmation prompt, the project's managed assets must survive
// byte-intact (the migration's removal arm runs after the gate —
// update_template_sync.go:1185). RED at HEAD is a regression of the r5
// repair and a separate leader finding.
func TestUpdateCancelKeepsProjectAssetsIntact(t *testing.T) {
	root := buildMigrationFixture(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cmd := newUpdateTestCmd()
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetContext(context.Background())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	prevConfirm := confirmViaPreviewFn
	confirmViaPreviewFn = func(merge.MergeAnalysis, string) (bool, error) { return false, nil }
	t.Cleanup(func() { confirmViaPreviewFn = prevConfirm })

	skipped, err := runTemplateSyncWithProgress(cmd, true)
	if err != nil {
		t.Fatalf("runTemplateSyncWithProgress: %v", err)
	}
	if !skipped {
		t.Errorf("a cancelled run must report skipped, got false")
	}
	if !strings.Contains(out.String(), "Merge cancelled by user") {
		t.Errorf("cancellation banner missing:\n%s", out.String())
	}
	// The project's managed asset survives byte-for-byte.
	assertFilePresent(t, root, migIdenticalSkill)
}

// TestInitResumeAfterUserAssetEnsureFailure — AC-018 (ledger 11b,
// REQ-SRF-007). Given a project whose first init failed at the user-asset
// ensure, a re-run must offer RESUME — not refuse with "already initialized"
// (init.go:905) — and the shortfall plus the never-run post-steps
// (ApplyHarness, MCP entry, Codex wiring) must complete. RED-now reason:
// the ensure failure returns at :947 with the post-steps unexecuted, and the
// re-run's executor refuses at :905 because the project is initialized.
func TestInitResumeAfterUserAssetEnsureFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")

	// The ensure-failure injection: a corrupt USER manifest makes
	// InstallPreserveSelection fail after the executor finished deploying.
	moai := filepath.Join(home, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moai, "user-assets.json"), []byte("{ corrupt\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	seamsForInit(t)
	projectDir := filepath.Join(t.TempDir(), "resume-proj")

	// Run 1: fails at the ensure, AFTER the project was initialized.
	cmd1 := newInitTestCmd()
	var out1, err1b strings.Builder
	cmd1.SetOut(&out1)
	cmd1.SetErr(&err1b)
	err1 := runInit(cmd1, []string{projectDir})
	if err1 == nil {
		t.Skipf("the ensure-failure injection did not bite (run 1 succeeded) — the corrupt-manifest injection needs re-aiming; recorded as a probe finding")
	}
	if !strings.Contains(err1.Error(), "user-asset install failed") {
		t.Fatalf("run 1 failed for the wrong reason (want the user-asset ensure): %v", err1)
	}
	// The failure left the project initialized — the state the AC names.
	if _, err := os.Stat(filepath.Join(projectDir, ".moai", "manifest.json")); err != nil {
		t.Fatalf("run 1 did not initialize the project before failing: %v", err)
	}

	// Run 2: the resume attempt.
	cmd2 := newInitTestCmd()
	var out2, err2b strings.Builder
	cmd2.SetOut(&out2)
	cmd2.SetErr(&err2b)
	err2 := runInit(cmd2, []string{projectDir})
	if err2 != nil && strings.Contains(err2.Error(), "already initialized") {
		t.Fatalf("RED (intended): the re-run after an ensure failure is refused with %q — no resume path exists (init.go:905); stderr: %s", firstLineStr(err2.Error()), err2b.String())
	}
	// GREEN-path shape (post-M6): the shortfall completed and the post-step
	// outputs exist. At HEAD this arm is unreachable past the refusal above.
	if err2 != nil {
		t.Fatalf("run 2 failed (post-refusal shape): %v", err2)
	}
	for _, rel := range []string{
		filepath.Join(".moai", "harness"),
		filepath.Join(".mcp.json"),
	} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
			t.Errorf("post-step output %s missing after resume: %v", rel, err)
		}
	}
}

// seamsForInit swaps the wizard seams the same way the autonomy-wiring tests
// do, so runInit runs non-interactively against the injected result.
func seamsForInit(t *testing.T) {
	t.Helper()
	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{AutonomyTier: config.AutonomyTierAutomatic}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })
}

// newUpdateTestCmd mirrors the update flag surface runTemplateSyncWithProgress reads.
func newUpdateTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("no-plugin", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("check", "", "")
	return cmd
}

func firstLineStr(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}
