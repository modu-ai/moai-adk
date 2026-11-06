package hook

// windows_scope_t2063_test.go — card t2063 (PR #1969 Windows CI): two families
// of internal/hook tests carry an assumption the Windows build does not meet by
// design, so they run where the assumption holds and say why they do not on
// Windows.
//
//  1. Branch-guard command-text tests (t2013 / t2049). The guard reads a Bash
//     tool command as a POSIX shell does: an explicit target is an absolute
//     `/…` path (explicitGitTargetPattern), and a backslash or `~` in a path is
//     shell expansion the scan refuses to trust (shellExpansionChars,
//     nonLiteralShellChars). The fixtures splice the OS-native temp path into the
//     command text; on Windows that is `C:\Users\RUNNER~1\…`, which is neither
//     absolute in the guard's sense nor free of expansion characters, so the
//     scan leaves the named target unjudged and the hook-cwd judgment decides.
//  2. Cross-process state-file lock tests (M5). internal/lockfile documents the
//     Windows lock as process-local ("do NOT silently upgrade this to
//     LockFileEx"), so a lock held by a second OS process is invisible there.
//
// Neither skip hides a behaviour the Windows build is meant to have; both name
// the documented limit. What the Windows PowerShell form of the guard should do
// with a native `git -C C:\primary commit` is a separate product question and
// is not decided here.

import (
	"runtime"
	"strings"
	"testing"
)

func skipWindowsPOSIXCommandText(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("t2063: the branch guard reads Bash command text as a POSIX shell; an OS-native Windows path (C:\\…, backslash, 8.3 ~1) is never adopted as a target (explicitGitTargetPattern needs a leading /; shellExpansionChars holds \\ and ~)")
	}
}

func skipWindowsCrossProcessLock(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("t2063: the Windows state-file lock is process-local by contract (internal/lockfile/lockfile_windows.go), so a lock held by another OS process is not observable")
	}
}

// TestWindowsNativePathIsOutsideTheGuardTextScan_T2063 pins the cause on every
// OS: the two functions that decide whether a command's named target is
// judged see a Windows-native path as neither a target nor an agreeing path.
func TestWindowsNativePathIsOutsideTheGuardTextScan_T2063(t *testing.T) {
	const win = `C:\Users\RUNNER~1\AppData\Local\Temp\repo`
	if explicitGitTargetPattern.MatchString("git -C " + win + " commit -m x") {
		t.Errorf("explicitGitTargetPattern adopted a drive-letter path; the POSIX-shaped scan must not")
	}
	if targetPathAgrees(win) {
		t.Errorf("targetPathAgrees(%q) = true, want false (backslash and ~ are expansion characters)", win)
	}
	if !strings.ContainsAny(win, shellExpansionChars) {
		t.Errorf("shellExpansionChars no longer covers the Windows-native path %q", win)
	}
	// The POSIX form of the same repository is what the scan is built for.
	if !targetPathAgrees("/tmp/repo") || !explicitGitTargetPattern.MatchString("git -C /tmp/repo commit -m x") {
		t.Errorf("a plain POSIX path must stay adoptable")
	}
}
