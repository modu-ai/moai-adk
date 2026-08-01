// Package cli — update_home_seam_test.go
//
// Guard for REQ-UGE-004 / REQ-UGE-005: every HOME resolution in the update
// subsystem routes through the userHomeDirFn seam, so an injected home reaches
// the template rendering inputs instead of the operator's real home directory.
//
// Why this matters (the debt this closes): the three call sites
// (update_clean_install.go deploy-context construction, update_template_sync.go
// Validate Templates and Deploy Templates) feed homeDir into BOTH
// template.WithHomeDir(homeDir) AND detectGoBinPathForUpdate(homeDir) →
// gobin.Detect, which falls back to filepath.Join(homeDir, "go", "bin") when
// GOBIN/GOPATH are empty. While those sites read the process $HOME directly, a
// test that injects userHomeDirFn still renders settings.json PATH and
// status_line.sh fallbacks against the operator's real home — the execution
// machine, not the fixture, decides the test's result.
//
// NFR-UGE-001: userHomeDirFn is a package-level variable. This test reassigns
// it, so neither it nor its subtests may call t.Parallel().

package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// homeSeamSpy replaces userHomeDirFn with a counting stub returning a sentinel
// home. It restores the original via t.Cleanup and returns the sentinel plus a
// call-count accessor.
//
// The count is the REACH evidence (was the seam actually consulted?); the
// sentinel is what the rendering-output assertion compares against, which is
// what binds REQ-UGE-005 rather than merely proving the seam was called.
func homeSeamSpy(t *testing.T) (sentinel string, calls func() int) {
	t.Helper()

	sentinel = t.TempDir()
	n := 0

	orig := userHomeDirFn
	t.Cleanup(func() { userHomeDirFn = orig })
	userHomeDirFn = func() (string, error) {
		n++
		return sentinel, nil
	}

	return sentinel, func() int { return n }
}

// TestUpdateSubsystem_HomeSeamReach asserts that the injected home reaches the
// update subsystem's rendering inputs at every call site REQ-UGE-004 names.
func TestUpdateSubsystem_HomeSeamReach(t *testing.T) {
	// Site 1 — update_clean_install.go deploy-context construction. The stub
	// deployer captures the constructed TemplateContext, so this subtest can
	// assert the rendering INPUT (not just that the seam was called).
	t.Run("clean_reinstall_deploy_context", func(t *testing.T) {
		sentinel, calls := homeSeamSpy(t)

		root := makeScenarioA(t)
		deployer := &stubDeployer{}
		migrate := &stubMigrateRunner{}

		if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
			Out:              io.Discard,
			Deployer:         deployer,
			RunMigrateAgency: migrate.Run,
		}); err != nil {
			t.Fatalf("runCleanReinstall: %v", err)
		}

		// Reach evidence — the seam was consulted on this path.
		if got := calls(); got < 1 {
			t.Errorf("userHomeDirFn calls = %d; want >= 1 (the deploy-context site does not route through the seam)", got)
		}

		// REQ-UGE-005 binding — the rendering input follows the INJECTED home,
		// not the process $HOME. TemplateContext.HomeDir is a plain field
		// assignment in WithHomeDir (no branch, no env var), so this comparison
		// is machine-independent.
		if deployer.lastTmplCtx == nil {
			t.Fatalf("stub deployer captured no TemplateContext; the deploy step did not run")
		}
		if deployer.lastTmplCtx.HomeDir != sentinel {
			t.Errorf("TemplateContext.HomeDir = %q; want the injected sentinel %q — rendering followed the process $HOME instead of the seam",
				deployer.lastTmplCtx.HomeDir, sentinel)
		}
	})

	// Sites 2 and 3 — update_template_sync.go "Validate Templates" and "Deploy
	// Templates". This driver builds its own deployer, so there is no captured
	// TemplateContext here; the reach evidence is the seam call count, and the
	// rendering-output binding is carried by site 1 above.
	t.Run("template_sync_validate_and_deploy", func(t *testing.T) {
		sentinel, calls := homeSeamSpy(t)
		_ = sentinel

		root := t.TempDir()
		writeTestFile(t, root, ".moai/config/sections/system.yaml",
			"system:\n  template_version: \"0.0.0\"\n")
		writeTestFile(t, root, ".moai/manifest.json", "{}\n")

		origDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}
		if chErr := os.Chdir(root); chErr != nil {
			t.Fatalf("chdir to fixture: %v", chErr)
		}
		t.Cleanup(func() { _ = os.Chdir(origDir) })

		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("force", false, "")
		cmd.Flags().Bool("yes", true, "")
		cmd.Flags().Bool("config", false, "")
		_ = cmd.Flags().Set("yes", "true")
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetContext(context.Background())

		if syncErr := runTemplateSyncWithReporter(cmd, nil, true); syncErr != nil {
			t.Fatalf("runTemplateSyncWithReporter: %v\noutput: %s", syncErr, buf.String())
		}

		// Two sites on this path — Validate Templates and Deploy Templates.
		if got := calls(); got < 2 {
			t.Errorf("userHomeDirFn calls = %d; want >= 2 (Validate Templates + Deploy Templates must both route through the seam)", got)
		}

		// NOT asserted here: that the rendered settings.json is free of the
		// operator's real home. REQ-UGE-004 scopes exactly three call sites, and
		// the rendered PATH does not come from any of them —
		// template.BuildSmartPATH (internal/template/settings.go) calls
		// os.UserHomeDir() itself and is a fourth, independent reader that this
		// SPEC does not touch. An assertion over the rendered PATH would fail
		// for a reason the seam cannot control, so it would test the wrong
		// thing. The seam's reach is proven by the call count above and by the
		// TemplateContext.HomeDir assertion in the sibling subtest.
	})
}
