package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/session"
)

// scrubSessionIDEnv removes CLAUDE_CODE_SESSION_ID for the duration of a test.
//
// The variable is stamped by the Claude Code runtime into every subprocess it
// spawns, so a `go test` run started from inside a Claude Code session inherits
// a real session UUID. Tests that exercise the degraded side-channel path must
// scrub it, or they measure the developer's session instead of the staged
// fixture — and pass or fail depending on where they were run from.
func scrubSessionIDEnv(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvClaudeCodeSessionID, "")
}

// TestResolveCurrentSessionID_EnvNamesTheAskingSession is the t221 regression.
//
// The defect: .moai/state/current-session-id.txt is ONE slot per project,
// overwritten on every SessionStart, so with two sessions in one checkout every
// session reads back whichever id was written last. The answer is not a
// function of the asking session — it is right sometimes, and nothing on the
// read side can tell which time.
//
// The test simulates two sessions sharing one project directory (one sidecar
// file, holding a third, foreign id) and asserts each resolves ITS OWN id. It
// asserts on the resolver's return value, not on any string in the source.
func TestResolveCurrentSessionID_EnvNamesTheAskingSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	// One shared sidecar slot, holding neither session's id — the last writer
	// in a multi-session checkout.
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	sidecar := filepath.Join(dir, session.CurrentSideChannelFile)
	if err := os.WriteFile(sidecar, []byte("foreign-session-id"), 0o600); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}

	for _, want := range []string{"session-A", "session-B"} {
		t.Run(want, func(t *testing.T) {
			t.Setenv(config.EnvClaudeCodeSessionID, want)

			got, source, ok := resolveCurrentSessionID()
			if !ok {
				t.Fatalf("resolveCurrentSessionID: ok=false, want true (env var is set to %q)", want)
			}
			if got != want {
				t.Errorf("session id = %q, want %q — the resolver answered with the shared "+
					"sidecar slot rather than this process's own session", got, want)
			}
			if !sessionIDSourceIsAuthoritative(source) {
				t.Errorf("source = %q, want the per-process env var; only that source is a "+
					"function of the asking session", source)
			}
		})
	}
}

// TestResolveCurrentSessionID_SidecarStaysAsDegradedPath locks in backward
// compatibility: with no env var (a runtime that does not stamp it), the
// side-channel file is still read. The fix reorders precedence; it removes no
// path.
func TestResolveCurrentSessionID_SidecarStaysAsDegradedPath(t *testing.T) {
	scrubSessionIDEnv(t)
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, session.CurrentSideChannelFile), []byte("sidecar-id"), 0o600); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}

	got, source, ok := resolveCurrentSessionID()
	if !ok || got != "sidecar-id" {
		t.Fatalf("resolveCurrentSessionID = (%q, %q, %v), want (\"sidecar-id\", side-channel, true)", got, source, ok)
	}
	if sessionIDSourceIsAuthoritative(source) {
		t.Errorf("source = %q, want the side-channel source", source)
	}
}

// TestResolveCurrentSessionID_NoSourceFallsBack keeps the canonical fallback
// reachable when neither source answers.
func TestResolveCurrentSessionID_NoSourceFallsBack(t *testing.T) {
	scrubSessionIDEnv(t)
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())

	got, source, ok := resolveCurrentSessionID()
	if ok {
		t.Fatalf("ok=true with no env var and no sidecar; got id %q", got)
	}
	if got != CanonicalFallbackSessionID || source != "fallback" {
		t.Errorf("= (%q, %q), want the canonical fallback", got, source)
	}
}

// TestResolveArmSessionID_EnvSkipsMultiSessionWarning: the multi-session
// warning exists because a sidecar id may be foreign. An env-resolved id
// cannot be, so concurrency no longer degrades the arm path — the goal is
// armed under this session's own id, with no warning to work around.
func TestResolveArmSessionID_EnvSkipsMultiSessionWarning(t *testing.T) {
	dir := withTempRegistry(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", cwd)
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-mine")

	// A foreign id in the shared slot, plus two concurrent registry entries —
	// exactly the state that made the unfixed resolver warn (or, worse, arm
	// under the foreign id).
	if err := os.WriteFile(filepath.Join(dir, session.CurrentSideChannelFile), []byte("sess-foreign"), 0o600); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
	if err := session.RegisterSession("sess-mine", "SPEC-X", "run"); err != nil {
		t.Fatalf("register mine: %v", err)
	}
	if err := session.RegisterSession("sess-foreign", "SPEC-X", "run"); err != nil {
		t.Fatalf("register foreign: %v", err)
	}

	id, warn := resolveArmSessionID("")
	if id != "sess-mine" {
		t.Errorf("arm id = %q, want sess-mine (this process's own session)", id)
	}
	if warn != "" {
		t.Errorf("warning = %q, want none: an env-resolved id cannot be foreign", warn)
	}
}

// TestIntegrationSessionID_EnvBeatsSidecar: the release-integration lock is
// held by a session id. Taking it under the shared slot's id means the holder
// recorded is not the lane holding it.
func TestIntegrationSessionID_EnvBeatsSidecar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, session.CurrentSideChannelFile), []byte("sess-foreign"), 0o600); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}

	t.Setenv(config.EnvClaudeCodeSessionID, "sess-mine")
	if got := integrationSessionID(""); got != "sess-mine" {
		t.Errorf("integrationSessionID = %q, want sess-mine", got)
	}
	// The explicit flag still wins over both.
	if got := integrationSessionID("sess-explicit"); got != "sess-explicit" {
		t.Errorf("integrationSessionID(explicit) = %q, want sess-explicit", got)
	}
}
