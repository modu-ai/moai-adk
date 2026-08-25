package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This file carries card t257's pre-push preservation tests, born as the
// measurement that reproduced the defect (both UserModified tests RED against
// the pre-t257 blind-overwrite installer) and retained as the regression guard
// for the fix: a marker-bearing hook the user edited must be backed up before
// replacement (REQ-PCP-003 equivalent) and the replacement must be disclosed on
// the warning writer (REQ-PCP-004 equivalent) — never silently lost. The
// unmodified-reinstall case must stay QUIET: no backup, no notice.

// newPrePushTestRepo creates a git repository in a temp dir and returns its
// root. git is required; its absence fails rather than skips, so an unrun
// criterion cannot read as a pass (mirrors newPreCommitTestRepo).
func newPrePushTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", dir, err, out)
	}
	return dir
}

func prepushHookPath(root string) string {
	return filepath.Join(root, ".git", "hooks", "pre-push")
}

// findPrePushBackups returns every pre-push.bak.* artifact in root's hooks
// dir — the pre-push naming equivalent of the pre-commit backup pattern
// (preCommitBackupPrefix, REQ-PCP-003).
func findPrePushBackups(t *testing.T, root string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, ".git", "hooks", "pre-push.bak.*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	return matches
}

// installPrePushAndPatch installs the hook once (baseline), then simulates a
// user editing the installed body: a customization is appended while the MoAI
// marker lines (first 3 lines) are untouched, so the marker gate still
// recognises the hook as MoAI-installed. Returns the patched bytes.
func installPrePushAndPatch(t *testing.T, root string) string {
	t.Helper()
	if err := NewPrePushInstaller(root).InstallPrePushHook(false); err != nil {
		t.Fatalf("baseline InstallPrePushHook: %v", err)
	}
	if !strings.Contains(prePushHookContent, moaiPrePushMarker) {
		t.Fatalf("fixture construction: prePushHookContent missing the marker %q", moaiPrePushMarker)
	}
	patched := prePushHookContent + "\n# USER PATCH: notify #ci on failure (added by the user)\n"
	if patched == prePushHookContent {
		t.Fatalf("fixture construction: user patch is vacuous")
	}
	if err := os.WriteFile(prepushHookPath(root), []byte(patched), 0o755); err != nil {
		t.Fatalf("write user-patched hook: %v", err)
	}
	return patched
}

// TestInstallPrePushHook_UserModified_ReinstallBacksUp — t257 reproduction,
// part one. Reinstalling over a marker-bearing, user-modified pre-push hook
// must not silently lose the user's edit: the t230 standard (REQ-PCP-003
// equivalent) requires a backup holding the pre-run bytes, taken before the
// replacement.
func TestInstallPrePushHook_UserModified_ReinstallBacksUp(t *testing.T) {
	root := newPrePushTestRepo(t)
	patched := installPrePushAndPatch(t, root)

	// Reinstall over the user-modified hook.
	if err := NewPrePushInstaller(root).InstallPrePushHook(false); err != nil {
		t.Fatalf("reinstall InstallPrePushHook: %v", err)
	}

	// Confirm the replacement actually happened, so a missing backup reads as
	// "silent loss" rather than "install failed".
	if got, err := os.ReadFile(prepushHookPath(root)); err != nil {
		t.Fatalf("read hook after reinstall: %v", err)
	} else if string(got) != prePushHookContent {
		t.Errorf("expected the canonical content after reinstall, got %d bytes", len(got))
	}

	// The t230 standard: the user's pre-run bytes must survive in a backup.
	backups := findPrePushBackups(t, root)
	if len(backups) == 0 {
		t.Errorf("SILENT LOSS: reinstall of a marker-bearing user-modified pre-push hook took no backup — the user patch is unrecoverable")
		return
	}
	if len(backups) != 1 {
		t.Errorf("expected exactly one backup, found %d: %v", len(backups), backups)
	}
	got, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read backup %s: %v", backups[0], err)
	}
	if string(got) != patched {
		t.Errorf("backup must hold the pre-run user-modified bytes: got %d bytes, want the %d-byte patched content", len(got), len(patched))
	}
}

// TestInstallPrePushHook_UserModified_ReplacementDisclosed — t257, part two.
// The replacement of a user-modified hook must be disclosed (REQ-PCP-004
// equivalent): the warning writer names the backup, and the progress writer
// must NOT carry the notice — `moai update` binds the progress writer to
// stdout, where a redirected run would swallow a data-loss notice whole.
func TestInstallPrePushHook_UserModified_ReplacementDisclosed(t *testing.T) {
	root := newPrePushTestRepo(t)
	installPrePushAndPatch(t, root)

	var out, warn strings.Builder
	installPrePushHookOptional(root, false, &out, &warn)

	if !strings.Contains(warn.String(), "backed up") {
		t.Errorf("SILENT REPLACEMENT: warning output %q does not disclose the replacement of the user-modified hook (standard: notice naming the backup)", warn.String())
	}
	if strings.Contains(out.String(), "backed up") {
		t.Errorf("disclosure leaked to the progress writer %q — the notice belongs on the warning writer only", out.String())
	}
}

// TestInstallPrePushHook_UnchangedReinstall_Quiet verifies the mirror of the
// backup path: a reinstall over an UNMODIFIED hook (provenance record matches)
// must be quiet — no backup file, no warning. Guards the noisy direction from
// both ends so the classifier's unmodified verdict keeps routine reinstalls
// silent.
func TestInstallPrePushHook_UnchangedReinstall_Quiet(t *testing.T) {
	root := newPrePushTestRepo(t)
	// Baseline install writes the hook plus its provenance record.
	if err := NewPrePushInstaller(root).InstallPrePushHook(false); err != nil {
		t.Fatalf("baseline InstallPrePushHook: %v", err)
	}

	var out, warn strings.Builder
	installPrePushHookOptional(root, false, &out, &warn)

	if len(findPrePushBackups(t, root)) != 0 {
		t.Errorf("unchanged reinstall must not take a backup, found: %v", findPrePushBackups(t, root))
	}
	if warn.String() != "" {
		t.Errorf("unchanged reinstall must produce no warning output, got %q", warn.String())
	}
	if !strings.Contains(out.String(), "Pre-push hook installed") {
		t.Errorf("unchanged reinstall should still report the install on the progress writer, got %q", out.String())
	}
}
