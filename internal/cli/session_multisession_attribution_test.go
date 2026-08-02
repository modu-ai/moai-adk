package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/session"
)

// TestResolveArmSessionID_MultiSessionWarns reproduces the multi-session
// side-channel hazard. The per-project .moai/state/current-session-id.txt file
// is unconditionally overwritten on every SessionStart (internal/hook/
// session_start.go), so when >=2 Claude Code sessions share a project directory
// the file holds only the most-recently-started session's id -- which may belong
// to a FOREIGN session. Arming under a foreign id lands the goal under the wrong
// session's state file and silently breaks the arm<->eval keying contract (the
// authoritative id Claude Code passes to the stop-goal hook via stdin would
// differ).
//
// Reproduction: register two concurrent sessions (A and B) both rooted in the
// temp project dir, write the side-channel file with A's id, then call the arm
// resolver and assert it produces a non-empty unreliability warning. On the
// unfixed code the resolver trusts the side-channel unconditionally and returns
// an empty warning -> this test FAILS (the RED state).
//
// The test is serial (not t.Parallel): withTempRegistry mutates the process
// cwd, and the registry read path is cwd-relative.
func TestResolveArmSessionID_MultiSessionWarns(t *testing.T) {
	dir := withTempRegistry(t) // chdir to tempdir + mkdir .moai/state (serial).

	// Use the canonical (symlink-resolved on macOS) cwd as CLAUDE_PROJECT_DIR so
	// the entry.CWD field (populated by os.Getwd at register time) matches
	// resolveProjectDir() (sourced from CLAUDE_PROJECT_DIR). Without this, the
	// /var vs /private/var divergence makes the count mismatch spuriously.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", cwd)

	// Write the side-channel file with session id A -- simulating the
	// SessionStart hook's unconditional overwrite (the file holds only ONE id).
	sidecar := filepath.Join(dir, session.CurrentSideChannelFile)
	if err := os.WriteFile(sidecar, []byte("sess-A"), 0o600); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}

	// Register TWO concurrent sessions both rooted in this project dir. This is
	// the hazard: a single side-channel file cannot represent both.
	if err := session.RegisterSession("sess-A", "SPEC-X", "run"); err != nil {
		t.Fatalf("register A: %v", err)
	}
	if err := session.RegisterSession("sess-B", "SPEC-X", "run"); err != nil {
		t.Fatalf("register B: %v", err)
	}

	id, warn := resolveArmSessionID("")
	if id != "sess-A" {
		t.Errorf("id: got %q, want sess-A (side-channel still resolves to the file content)", id)
	}
	// Load-bearing assertion: under multi-session concurrency the resolver MUST
	// NOT silently trust the side-channel id.
	if warn == "" {
		t.Fatal("expected a non-empty multi-session unreliability warning when >=2 sessions " +
			"share the project dir; got empty (silent trust of a possibly-foreign side-channel id)")
	}
}

// TestResolveArmSessionID_SingleSessionNoWarn locks in the PRESERVED happy
// path: when <=1 registry entry matches the project dir, the arm resolver
// returns the side-channel id with NO warning (single-session behavior is
// unchanged by the fix).
func TestResolveArmSessionID_SingleSessionNoWarn(t *testing.T) {
	dir := withTempRegistry(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", cwd)

	sidecar := filepath.Join(dir, session.CurrentSideChannelFile)
	if err := os.WriteFile(sidecar, []byte("sess-only"), 0o600); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
	// Exactly one session registered in this project dir.
	if err := session.RegisterSession("sess-only", "SPEC-Y", "run"); err != nil {
		t.Fatalf("register: %v", err)
	}

	id, warn := resolveArmSessionID("")
	if id != "sess-only" {
		t.Errorf("id: got %q, want sess-only", id)
	}
	if warn != "" {
		t.Errorf("single-session must NOT warn; got warning: %s", warn)
	}
}
