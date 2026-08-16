package hook

// cwd_changed_relocate.go — t74 registry-CWD relocation on CwdChanged.
//
// A session that enters a worktree mid-session (Claude Code's EnterWorktree)
// keeps its registry entry's CWD at the launch-time value forever, which
// makes it invisible to anchor detection (session.LiveAnchoredSessions).
// When the runtime reports a working-directory change, the entry's CWD moves
// to the new directory — in whichever registry the entry actually lives.

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/session"
)

// relocateSessionCwd moves the session's registry entry CWD to newCwd.
// Candidate registries are found by walking UP from the old and new working
// directories (and the payload CWD) to the first directory holding a
// .moai/state/active-sessions.json that contains the session — the entry
// was written by the launch checkout's SessionStart, which is an ancestor
// of the old directory in the enter case and of the new one in the exit
// case. Fail-open: no registry found, unreadable registry, or a relocate
// error leaves everything untouched and never fails the hook.
func relocateSessionCwd(input *HookInput, newCwd string) {
	if input == nil || input.SessionID == "" || newCwd == "" {
		return
	}
	candidates := []string{input.OldCwd, input.NewCwd, input.CWD}
	tried := make(map[string]bool)
	for _, root := range candidates {
		if root == "" || tried[root] {
			continue
		}
		tried[root] = true
		regPath, ok := findRegistryUpward(root)
		if !ok {
			continue
		}
		reg := session.NewRegistry(regPath, nil)
		entries, err := reg.Query("")
		if err != nil {
			continue
		}
		found := false
		for _, e := range entries {
			if e.SessionID == input.SessionID {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		if err := reg.RelocateSession(input.SessionID, newCwd); err != nil {
			slog.Warn("cwd-changed: registry relocate failed (non-blocking)",
				"error", err.Error(),
				"session_id", input.SessionID,
			)
		}
		return
	}
}

// findRegistryUpward walks from dir toward the filesystem root and reports
// the first <dir>/.moai/state/active-sessions.json that exists.
func findRegistryUpward(dir string) (string, bool) {
	for {
		candidate := filepath.Join(dir, session.DefaultRegistryPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
