package cli

// Card t1139 follow-up (sync-audit F2): the clean-reinstall half of the fix —
// identity names in the Step 5 render context, and the template snapshot
// written right after the Step 5 deploy rather than after the Step 5.5 restore
// — driven through the REAL runCleanReinstall with the real embedded-template
// deployer. The helper-level snapshot tests never reach that call site, so a
// revert of either half passed them all.
//
// prepareSafeInitHome sets environment variables and the update helper chdirs,
// so this test must not run in parallel.

import (
	"bytes"
	"context"
	"os"
	"testing"
)

// TestCleanReinstall_SnapshotIsTheDeployedRenderAndIdentitySurvives proves on
// the real clean-reinstall path that (i) the snapshot records the deployed
// render, not the restored user values, and (ii) project.name, user.name and
// the wizard-patched lsp.enabled survive the reinstall and the next update.
func TestCleanReinstall_SnapshotIsTheDeployedRenderAndIdentitySurvives(t *testing.T) {
	root := initIdentityProject(t)
	assertIdentityKeys(t, root, "after init")
	if t.Failed() {
		t.Fatalf("init did not produce the expected identity keys; the reinstall assertions would be vacuous")
	}

	// Turn the fresh v3 project into one the clean-reinstall accepts: a v2
	// version reading (Signal 1) plus a deprecated path (Signal 3). The v3
	// version init wrote would otherwise veto the reinstall.
	writeTestFile(t, root, ".moai/config/sections/system.yaml", "moai:\n    version: v2.16.1\n")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired\n")

	var out, errOut bytes.Buffer
	migrate := &stubMigrateRunner{}
	result, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              &out,
		ErrOut:           &errOut,
		RunMigrateAgency: migrate.Run,
	})
	if err != nil {
		t.Fatalf("runCleanReinstall: %v\nout: %s\nerr: %s", err, out.String(), errOut.String())
	}
	// Positive control: the reinstall body ran (not the early not-v2 return).
	if !result.Detected.IsV2 {
		t.Fatalf("fixture was not detected as v2; the reinstall body never ran (details: %v)", result.Detected.SignalDetails)
	}

	// (i) The snapshot is the Step 5 render: the template ships lsp.enabled
	// false, and the render carries the names the project already had.
	if got := sectionValue(t, snapshotFile(root, "lsp.yaml"), "lsp", "enabled"); got != false {
		t.Errorf("after reinstall: snapshot lsp.enabled = %v, want false (the deployed render, not the restored user value)", got)
	}
	if got := sectionValue(t, snapshotFile(root, "user.yaml"), "user", "name"); got != identityUserName {
		t.Errorf("after reinstall: snapshot user.name = %v, want %q (the render carries the existing name)", got, identityUserName)
	}
	if got := sectionValue(t, snapshotFile(root, "project.yaml"), "project", "name"); got != identityProjectName {
		t.Errorf("after reinstall: snapshot project.name = %v, want %q (the render carries the existing name)", got, identityProjectName)
	}

	// (ii) The live tree keeps the user's values.
	assertIdentityKeys(t, root, "after reinstall")

	// The snapshot the reinstall left is the next update's BASE; with a
	// post-restore snapshot the next update resets lsp.enabled.
	runForcedTemplateSyncAt(t, root)
	assertIdentityKeys(t, root, "after reinstall + update")

	if _, statErr := os.Stat(snapshotFile(root, "lsp.yaml")); statErr != nil {
		t.Errorf("snapshot lsp.yaml missing after the follow-up update: %v", statErr)
	}
}

// TestCleanReinstall_SnapshotWarningGoesToErrOut pins sync-audit F7: a failed
// section-snapshot write is a warning, and the clean-reinstall routes warnings
// to ErrOut (its Out carries stdout in production) — the same stream the
// template-sync path uses for the identical warning.
func TestCleanReinstall_SnapshotWarningGoesToErrOut(t *testing.T) {
	sentinel, _ := homeSeamSpy(t)
	snapAssertSandboxHome(t, sentinel)

	root := makeScenarioA(t)
	// A regular file where the snapshot's sections/ directory belongs makes
	// WriteSnapshot fail at MkdirAll.
	writeTestFile(t, root, ".moai/cache/template-snapshot/sections", "a file where a directory belongs")

	out, errOut := snapRunCleanReinstall(t, root, &stubDeployer{})
	const prefix = "Warning: template snapshot write failed:"
	if n := snapCountPrefixed(errOut, prefix); n != 1 {
		t.Errorf("snapshot warnings on ErrOut = %d, want 1\nerrOut:\n%s", n, errOut)
	}
	if n := snapCountPrefixed(out, prefix); n != 0 {
		t.Errorf("snapshot warnings on Out = %d, want 0 (warnings belong on the error stream)\nout:\n%s", n, out)
	}
}
