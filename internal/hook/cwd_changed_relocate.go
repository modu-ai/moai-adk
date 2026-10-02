package hook

// cwd_changed_relocate.go — t74 registry-CWD relocation on CwdChanged.
//
// A session that enters a worktree mid-session (Claude Code's EnterWorktree)
// keeps its registry entry's CWD at the launch-time value forever, which
// makes it invisible to anchor detection (session.LiveAnchoredSessions).
// When the runtime reports a working-directory change, the entry's CWD moves
// to the new directory — in whichever registry the entry actually lives.

import (
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/session"
)

// relocationGitContext resolves the git worktree context for a target
// directory: the tree root (git rev-parse --show-toplevel) and the worktree
// porcelain listing. It is a package variable so tests can express the
// context without a real git repository — the same test seam shape the
// session package uses for sessionProcessLiveness.
var relocationGitContext = func(dir string) (treeRoot, porcelain string, ok bool) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", "", false
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", "", false
	}
	porcelainOut, err := exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return root, "", false
	}
	return root, string(porcelainOut), true
}

// anchorRelocationGuardEnabled reads the opt-in anchor relocation guard flag
// (workflow.anchor_relocation_guard.enabled, REQ-SAA-005). Nil-safe: a nil
// provider reads as disabled — the distributed default keeps the guard inert.
func anchorRelocationGuardEnabled(cfg ConfigProvider) bool {
	if cfg == nil || cfg.Get() == nil {
		return false
	}
	return cfg.Get().Workflow.AnchorRelocationGuard.Enabled
}

// relocationTargetContext resolves the target tree root and its git worktree
// lock in one git-context read. The lock is nil when the context is
// unresolvable or the tree carries no lock — the ownership evaluation then
// reads OwnerNone (no lock observable), never a guessed one.
func relocationTargetContext(newCwd string) (string, *session.LockInfo) {
	treeRoot, porcelain, ok := relocationGitContext(newCwd)
	if !ok {
		return "", nil
	}
	locks := session.ParseWorktreeLocks(porcelain)
	if lock, found := locks[treeRoot]; found {
		return treeRoot, &lock
	}
	if lock, found := locks[resolveSymlinks(treeRoot)]; found {
		return treeRoot, &lock
	}
	return treeRoot, nil
}

// relocateSessionCwd moves the session's registry entry CWD to newCwd.
// Candidate registries are found by walking UP from the old and new working
// directories (and the payload CWD) to the first directory holding a
// .moai/state/active-sessions.json that contains the session — the entry
// was written by the launch checkout's SessionStart, which is an ancestor
// of the old directory in the enter case and of the new one in the exit
// case. When none of those walks carries the session, the repository's
// PRIMARY registry is consulted (see relocateRegistryCandidates). Fail-open:
// no registry found, unreadable registry, or a relocate error leaves
// everything untouched and never fails the hook.
//
// SPEC-SESSION-ANCHOR-ATTR-001 W2: trigger names the hook event driving the
// relocation (recorded on the audit row), and refuseFlagged is the opt-in
// anchor relocation guard — when true, an ownership-flagged relocation is
// refused (ErrRelocationRefused path) instead of proceeding advisory.
func relocateSessionCwd(input *HookInput, newCwd, trigger string, refuseFlagged bool) {
	if input == nil || input.SessionID == "" || newCwd == "" {
		return
	}
	candidates := []string{input.OldCwd, input.NewCwd, input.CWD}
	for _, regPath := range relocateRegistryCandidates(candidates) {
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

		// W2: resolve the target-tree ownership context. The git lookups are
		// fail-open — an unresolvable context degrades the audit row to
		// OwnerNone and never blocks the relocation.
		treeRoot, lock := relocationTargetContext(newCwd)
		anchoredOthers := 0
		if treeRoot != "" {
			anchoredOthers = session.CountAnchoredOthers(entries, input.SessionID, treeRoot)
		}

		_, err = reg.RelocateSessionWithOptions(input.SessionID, newCwd, session.RelocationOptions{
			Trigger:        trigger,
			TargetLock:     lock,
			AnchoredOthers: anchoredOthers,
			RefuseFlagged:  refuseFlagged,
		})
		if err != nil {
			if errors.Is(err, session.ErrRelocationRefused) {
				// An intended opt-in refusal (REQ-SAA-005) — the audit row
				// records it; the hook logs and leaves the registry as-is.
				slog.Warn("cwd-changed: relocation refused by anchor relocation guard",
					"session_id", input.SessionID,
					"new_cwd", newCwd,
				)
				return
			}
			slog.Warn("cwd-changed: registry relocate failed (non-blocking)",
				"error", err.Error(),
				"session_id", input.SessionID,
			)
		}
		return
	}
}

// relocateRegistryCandidates returns the registry files to consult for a
// relocation, in order and without repeats.
//
// Two passes, and the order is load-bearing. Pass 1 is the pre-existing
// upward walk over every candidate working directory, so a relocation the
// old implementation already resolved resolves identically. Pass 2 adds the
// repository's PRIMARY registry for each candidate, resolved through
// session.RegistryPathFor — the same anchor seam the write path uses.
//
// Pass 2 exists because the upward walk stops at the FIRST
// active-sessions.json above a candidate: when that file does not carry the
// session, the loop advanced to the next candidate WORKING DIRECTORY, not to
// the parent directory. Where every candidate sits inside one linked worktree
// holding its own orphan registry, that left no path to the primary registry
// at all and the relocation silently did nothing
// (SPEC-SESSION-REGISTRY-READ-ANCHOR-001 REQ-RAR-002, REQ-RAR-003).
//
// The home boundary findRegistryUpwardFrom enforces applies to pass 2 as
// well: ~/.moai/state is global state, never a project registry, so a home
// directory that happens to be a git repository does not become an anchor.
func relocateRegistryCandidates(dirs []string) []string {
	home, _ := paths.Home()
	return relocateRegistryCandidatesFrom(dirs, home)
}

// relocateRegistryCandidatesFrom is relocateRegistryCandidates with the home
// boundary supplied rather than read from the environment, so the ordering
// can be exercised over a directory tree the test fully owns.
func relocateRegistryCandidatesFrom(dirs []string, homeDir string) []string {
	var out []string
	seenDir := make(map[string]bool)
	seenPath := make(map[string]bool)
	add := func(path string) {
		if path == "" || seenPath[path] {
			return
		}
		seenPath[path] = true
		out = append(out, path)
	}

	unique := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if dir == "" || seenDir[dir] {
			continue
		}
		seenDir[dir] = true
		unique = append(unique, dir)
	}

	// Pass 1 — the unchanged upward walk.
	for _, dir := range unique {
		if path, ok := findRegistryUpwardFrom(dir, homeDir); ok {
			add(path)
		}
	}
	// Pass 2 — the repository's primary registry.
	for _, dir := range unique {
		add(primaryRegistryFor(dir, homeDir))
	}
	return out
}

// primaryRegistryFor resolves the repository's primary registry for dir
// through session.RegistryPathFor, or "" when the result would cross the
// home boundary. RegistryPathFor falls back to dir joined with
// DefaultRegistryPath outside a git repository; that fallback is kept (it is
// a path pass 1 has already covered when the file exists) except at the home
// directory itself, which the walk deliberately refuses.
func primaryRegistryFor(dir, homeDir string) string {
	if dir == "" {
		return ""
	}
	path := session.RegistryPathFor(dir)
	if homeDir == "" {
		return path
	}
	// DefaultRegistryPath has three components, so three Dir calls recover
	// the root RegistryPathFor joined it to.
	root := filepath.Dir(filepath.Dir(filepath.Dir(path)))
	if resolveSymlinks(root) == resolveSymlinks(homeDir) {
		return ""
	}
	return path
}

// findRegistryUpwardFrom walks from dir toward the filesystem root and
// reports the first <dir>/.moai/state/active-sessions.json that exists,
// stopping at the supplied home directory.
//
// The home boundary is the one project.FindProjectRoot already enforces:
// ~/.moai is global state, not a project. Without it a session working
// outside any checkout climbs past $HOME and relocates its entry into the
// global registry — a write to shared state on behalf of a project that was
// never found.
//
// The boundary is a parameter rather than an environment read so the walk can
// be exercised over a directory tree the test fully owns. Both paths are
// symlink-resolved before comparison (macOS /private/var, Windows 8.3 short
// paths), matching the normalization project.FindProjectRoot performs.
func findRegistryUpwardFrom(dir, homeDir string) (string, bool) {
	if dir == "" {
		return "", false
	}
	dir = resolveSymlinks(dir)
	homeDir = resolveSymlinks(homeDir)

	for {
		// Stop at the home directory — ~/.moai/state is global state, never a
		// project registry.
		if homeDir != "" && dir == homeDir {
			return "", false
		}
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

// resolveSymlinks returns path with symlinks resolved, or path unchanged when
// it cannot be resolved (a directory that does not exist, for instance).
func resolveSymlinks(path string) string {
	if path == "" {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return path
}
