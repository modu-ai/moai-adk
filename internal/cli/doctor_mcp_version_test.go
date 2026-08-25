package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// TestCheckMCPServerVersion_NoRunningServer asserts the check reports OK, not
// WARN, when nothing is running — absence of a server is not a defect.
func TestCheckMCPServerVersion_NoRunningServer(t *testing.T) {
	const binCommit = "aaaaaaaaaaaa"
	check := checkMCPServerVersionAgainst(t.TempDir(), binCommit, false)

	if check.Name != mcpServerVersionCheckName {
		t.Errorf("check.Name = %q, want %q", check.Name, mcpServerVersionCheckName)
	}
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %v, want CheckOK", check.Status)
	}
	if !strings.Contains(check.Message, "no running") {
		t.Errorf("Message = %q, want it to state that no server is recorded", check.Message)
	}
}

// TestCheckMCPServerVersion_StaleServerWarns is the core regression: a live
// server stamped with a different commit than the installed binary MUST warn
// and MUST tell the operator to reconnect.
func TestCheckMCPServerVersion_StaleServerWarns(t *testing.T) {
	const binCommit = "bbbbbbbbbbbb"
	projectDir := t.TempDir()
	writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: os.Getpid(), Commit: "aaaaaaaaaaaa"})

	check := checkMCPServerVersionAgainst(projectDir, binCommit, false)

	if check.Status != uikit.CheckWarn {
		t.Fatalf("Status = %v, want CheckWarn (a live server on a different commit is stale)", check.Status)
	}
	if !strings.Contains(check.Message, "aaaaaaaaa") || !strings.Contains(check.Message, "bbbbbbbbb") {
		t.Errorf("Message = %q, want both the server and binary commits named", check.Message)
	}
	if !strings.Contains(check.Detail, "Reconnect") {
		t.Errorf("Detail = %q, want reconnect guidance", check.Detail)
	}
}

// TestCheckMCPServerVersion_MatchingServerOK asserts no warning when the live
// server was built from the installed commit.
func TestCheckMCPServerVersion_MatchingServerOK(t *testing.T) {
	const binCommit = "cccccccccccc"
	projectDir := t.TempDir()
	writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: os.Getpid(), Commit: "cccccccccccc"})

	check := checkMCPServerVersionAgainst(projectDir, binCommit, false)

	if check.Status != uikit.CheckOK {
		t.Fatalf("Status = %v, want CheckOK", check.Status)
	}
	if !strings.Contains(check.Message, "match") {
		t.Errorf("Message = %q, want it to state the server matches", check.Message)
	}
}

// TestCheckMCPServerVersion_AbbreviatedCommitMatches asserts a short-hash
// stamp is not reported stale against the same commit's long hash.
func TestCheckMCPServerVersion_AbbreviatedCommitMatches(t *testing.T) {
	const binCommit = "abcdef1234567890"
	projectDir := t.TempDir()
	writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: os.Getpid(), Commit: "abcdef123"})

	if check := checkMCPServerVersionAgainst(projectDir, binCommit, false); check.Status != uikit.CheckOK {
		t.Fatalf("Status = %v, want CheckOK for an abbreviated form of the same commit", check.Status)
	}
}

// TestCheckMCPServerVersion_DevBuildSkips asserts a binary with no commit
// metadata reports OK rather than warning on every server — an unattributable
// mismatch is a gap, not a defect.
func TestCheckMCPServerVersion_DevBuildSkips(t *testing.T) {
	const binCommit = "none"
	projectDir := t.TempDir()
	writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: os.Getpid(), Commit: "aaaaaaaaaaaa"})

	check := checkMCPServerVersionAgainst(projectDir, binCommit, false)
	if check.Status != uikit.CheckOK {
		t.Fatalf("Status = %v, want CheckOK for a dev build", check.Status)
	}
	if !strings.Contains(check.Message, "development build") {
		t.Errorf("Message = %q, want it to name the dev-build skip reason", check.Message)
	}
}

// TestCheckMCPServerVersion_DevBuildServerNotCounted asserts a server that was
// itself built without commit metadata is not counted as a mismatch.
func TestCheckMCPServerVersion_DevBuildServerNotCounted(t *testing.T) {
	const binCommit = "dddddddddddd"
	projectDir := t.TempDir()
	writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: os.Getpid(), Commit: "none"})

	if check := checkMCPServerVersionAgainst(projectDir, binCommit, false); check.Status != uikit.CheckOK {
		t.Fatalf("Status = %v, want CheckOK when the server carries no commit metadata", check.Status)
	}
}

// TestCheckMCPServerVersion_PrunesDeadRecords asserts a stamp left behind by a
// hard-killed server is removed in passing and never produces a warning.
func TestCheckMCPServerVersion_PrunesDeadRecords(t *testing.T) {
	const binCommit = "eeeeeeeeeeee"
	projectDir := t.TempDir()
	deadPath := writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: 4194303, Commit: "ffffffffffff"})

	check := checkMCPServerVersionAgainst(projectDir, binCommit, false)

	if check.Status != uikit.CheckOK {
		t.Fatalf("Status = %v, want CheckOK (a dead server is not a skew)", check.Status)
	}
	if _, err := os.Stat(deadPath); !os.IsNotExist(err) {
		t.Errorf("dead record still present at %s: %v", deadPath, err)
	}
}

// TestCheckMCPServerVersion_RegisteredInDoctor asserts the check runs as part
// of `moai doctor` and is addressable by --check, so the diagnostic is
// reachable rather than merely defined.
func TestCheckMCPServerVersion_RegisteredInDoctor(t *testing.T) {
	checks := runDiagnosticChecks(false, mcpServerVersionCheckName)
	if len(checks) != 1 {
		t.Fatalf("filtering by %q returned %d checks, want 1", mcpServerVersionCheckName, len(checks))
	}
	if checks[0].Name != mcpServerVersionCheckName {
		t.Errorf("check.Name = %q, want %q", checks[0].Name, mcpServerVersionCheckName)
	}
}

// TestMCPServerRuntimeDir_ProjectScoped asserts the record path is anchored
// under the project's .moai/state tree (not $HOME, not a relative path).
func TestMCPServerRuntimeDir_ProjectScoped(t *testing.T) {
	projectDir := t.TempDir()
	want := filepath.Join(projectDir, ".moai", "state", "mcp-server")
	if got := mcpServerRuntimeDir(projectDir); got != want {
		t.Errorf("mcpServerRuntimeDir = %q, want %q", got, want)
	}
	if got := mcpServerRuntimeDir(""); got != "" {
		t.Errorf("mcpServerRuntimeDir(\"\") = %q, want empty", got)
	}
}
