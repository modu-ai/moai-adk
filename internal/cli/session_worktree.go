package cli

// session_worktree.go — SPEC-SESSION-WORKTREE-001 M2 session-worktree auto-entry.
//
// enterSessionWorktree is the wrapper invoked at the top of moai init (M2),
// moai web (M3), and moai profile (M6) BEFORE any shared-state mutation. It
// materializes an isolated git worktree when the feature is active so parallel
// invocations of these subcommands do not collide on the shared primary
// checkout's `.moai/state/`, `.claude/settings.local.json`, and
// `.moai/config/sections/*.yaml`.
//
// Activation (config.SessionWorktreeEnabled, M1) is the single decision site;
// default OFF (REQ-SW-001 byte-identical baseline). Every failure path is a
// fail-back to the shared checkout with a stderr notice (REQ-SW-004: a
// materialization failure MUST NOT abort the invocation).

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/spf13/cobra"
)

// SessionWorktreeBranchPrefix is the literal branch-name prefix.
//
// REQ-SW-006 carries the literal `[WT]` intent, but square brackets are
// rejected by `git check-ref-format` (verified at M2 pre-flight: `git
// check-ref-format --branch '[WT]-x'` exits non-zero with `'[WT]-x' is not a
// valid branch name`). Q2 (plan.md §I) resolves to the bracket-free `WT-`
// prefix as the documented fallback — see progress.md §E.2 for the verbatim
// check-ref-format exit evidence.
const SessionWorktreeBranchPrefix = "WT-"

// sessionWorktreeSubdir is the worktree landing directory relative to the
// primary checkout's project root, matching the Claude-native L1 convention
// (`.claude/worktrees/<name>/`) from worktree-integration.md § Terminology
// Glossary.
const sessionWorktreeSubdir = ".claude" + string(filepath.Separator) + "worktrees"

// Function-variable seams for test injection. Each has a Real counterpart
// below; tests swap these via swapSessionWorktreeSeams and restore on cleanup.
var (
	// sessionWorktreeGitWorktreeAdd runs `git worktree add -b <branch> <dest>`
	// and returns the absolute worktree path on success.
	sessionWorktreeGitWorktreeAdd = gitWorktreeAddReal

	// sessionWorktreeInGitWorktree reports whether cwd is already inside a git
	// worktree (git-dir != git-common-dir) — REQ-SW-012.
	sessionWorktreeInGitWorktree = inGitWorktreeReal

	// sessionWorktreeResolveSessionShort resolves the <session-short> segment
	// (first 8 chars of the session UUID), falling back to 6-byte random hex
	// when no session id is available (REQ-SW-007 EC-4).
	sessionWorktreeResolveSessionShort = resolveSessionShortReal

	// sessionWorktreeGitCommonDir returns the absolute path of the common git
	// directory (the primary checkout's .git dir), or an error when cwd is
	// not inside a git repo — the BI-3 / REQ-SW-004 fail-back trigger.
	sessionWorktreeGitCommonDir = gitCommonDirReal

	// sessionWorktreeGitConfigSet runs `git -C <dir> config <key> <value>`.
	sessionWorktreeGitConfigSet = gitConfigSetReal
)

// enterSessionWorktree is the auto-entry wrapper shared by moai init (M2),
// moai web (M3), and moai profile (M6). It MUST be called BEFORE any
// shared-state mutation.
//
// Returns the materialized worktree's absolute path, or "" when:
//   - the feature is OFF (REQ-SW-001: byte-identical baseline; no side effect),
//   - already inside a worktree (REQ-SW-012: skip + info notice), or
//   - materialization failed (REQ-SW-004: fail-back + notice, never abort).
//
// All notices route to out (stderr in production). The function NEVER returns
// an error — every failure path is a fail-back.
//
// @MX:ANCHOR: [AUTO] session-worktree auto-entry wrapper (M2 init / M3 web / M6 profile)
// @MX:REASON: REQ-SW-001/002/004/012 — single entry-point shape consumed by three subcommands; default-off short-circuit MUST precede any side effect, and fail-back MUST NOT abort
func enterSessionWorktree(cfg *config.Config, subcommand string, out io.Writer) string {
	if !config.SessionWorktreeEnabled(cfg) {
		// REQ-SW-001: default-off is byte-identical to the baseline — no
		// notice, no git invocation, no observable side effect.
		return ""
	}
	if sessionWorktreeInGitWorktree() {
		// REQ-SW-012: do not nest. The user is already worktree-isolated.
		_, _ = fmt.Fprintf(out, "moai: already inside a git worktree; skipping session-worktree auto-entry for %q\n", subcommand)
		return ""
	}
	branch := sessionWorktreeBranchName(subcommand)
	wtPath, err := materializeSessionWorktree(branch)
	if err != nil {
		// REQ-SW-004: a materialization failure MUST NOT abort the invocation.
		// Fall back to the shared-checkout behavior and emit a non-blocking
		// notice naming the failure reason.
		_, _ = fmt.Fprintf(out, "moai: session-worktree materialization failed (%v); continuing in shared checkout for %q\n", err, subcommand)
		return ""
	}
	_, _ = fmt.Fprintf(out, "moai: entered session worktree %s (branch %s) for %q\n", wtPath, branch, subcommand)
	return wtPath
}

// sessionWorktreeBranchName builds the deterministic branch name
// WT-<session-short>-<subcommand> (REQ-SW-007). The Q2 bracket fallback
// (SessionWorktreeBranchPrefix = "WT-") is applied here.
func sessionWorktreeBranchName(subcommand string) string {
	return SessionWorktreeBranchPrefix + sessionWorktreeResolveSessionShort() + "-" + subcommand
}

// materializeSessionWorktree creates the worktree at the conventional path
// under the primary checkout's project root and applies the M2 init-specific
// git config.
//
// BI-1 decision (progress.md §E.2): no reusable `worktree.Add` Go helper
// exists in internal/cli/worktree/ or internal/workflow/ (only mock-interface
// Add methods in _test.go). M2 shells out to `git worktree add` directly
// rather than extracting a helper — the shell-out is simpler and the
// extraction would be premature given M3/M6 will reuse this same wrapper.
func materializeSessionWorktree(branch string) (string, error) {
	commonDir, err := sessionWorktreeGitCommonDir()
	if err != nil {
		return "", fmt.Errorf("resolve git common dir: %w", err)
	}
	// The primary checkout's project root is the parent of the common git dir
	// (the directory that contains `.git/`). For a bare repo this would be the
	// common dir itself, but moai init/profile/web operate inside a working
	// tree, so the parent is correct.
	projectRoot := filepath.Dir(commonDir)
	destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)
	wtPath, err := sessionWorktreeGitWorktreeAdd(destDir, branch)
	if err != nil {
		return "", fmt.Errorf("git worktree add: %w", err)
	}

	// REQ-SW-020 (M2 init-only config): set init.defaultBranch=main on the
	// worktree's local config so any new repository initialized inside the
	// worktree defaults to `main` (not the git built-in `master`). A failure
	// here is non-fatal — the worktree is usable; defaultBranch is a
	// convenience applied best-effort.
	//
	// M7 (REQ-SW-018/019/021): ApplyGitConfig call site — safe.directory
	// (REQ-SW-018), global-gitconfig identity (REQ-SW-019), and opt-in
	// options (REQ-SW-021) land here once M7 implements the helper.
	if cerr := sessionWorktreeGitConfigSet(wtPath, "init.defaultBranch", "main"); cerr != nil {
		// Swallow: best-effort. The worktree is returned usable.
		return wtPath, nil
	}
	return wtPath, nil
}

// --- real implementations (overridable in tests via the seams above) ---

// gitWorktreeAddReal runs `git worktree add -b <branch> <dest>` and returns
// the absolute destination path on success.
func gitWorktreeAddReal(destDir, branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "add", "-b", branch, destDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	abs, err := filepath.Abs(destDir)
	if err != nil {
		return destDir, nil
	}
	return abs, nil
}

// inGitWorktreeReal reports whether cwd is inside a git worktree by comparing
// --git-dir and --git-common-dir (REQ-SW-012 detection). Any git error degrades
// to false (the caller then attempts materialization and hits the fail-back).
func inGitWorktreeReal() bool {
	gitDir, err1 := exec.Command("git", "rev-parse", "--git-dir").Output()
	commonDir, err2 := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err1 != nil || err2 != nil {
		return false
	}
	return strings.TrimSpace(string(gitDir)) != strings.TrimSpace(string(commonDir))
}

// resolveSessionShortReal resolves the 8-char session-short segment from the
// side-channel session id, falling back to 6-byte random hex (REQ-SW-007 EC-4)
// when no session id is available.
func resolveSessionShortReal() string {
	id, _, ok := resolveCurrentSessionID()
	if ok {
		id = strings.TrimSpace(id)
		if len(id) >= 8 {
			return id[:8]
		}
		if id != "" {
			return id
		}
	}
	// Fallback: 6-byte random hex (12 hex chars) per REQ-SW-007 EC-4.
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read failure is exceptional; return a deterministic non-empty
		// segment so the branch name is still well-formed.
		return "000000000000"
	}
	return hex.EncodeToString(b[:])
}

// gitCommonDirReal returns the absolute path of the common git directory.
func gitCommonDirReal() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", err
	}
	p := strings.TrimSpace(string(out))
	if p == "" {
		return "", fmt.Errorf("empty git-common-dir")
	}
	if !filepath.IsAbs(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return p, nil
		}
		p = abs
	}
	return p, nil
}

// gitConfigSetReal runs `git -C <dir> config <key> <value>`.
func gitConfigSetReal(dir, key, value string) error {
	cmd := exec.Command("git", "-C", dir, "config", key, value)
	return cmd.Run()
}

// loadSessionWorktreeConfig loads the project config for the session-worktree
// activation decision. It degrades to nil (treated as OFF by
// config.SessionWorktreeEnabled) when the config cannot be loaded — the
// activation site is fail-safe to OFF, preserving the REQ-SW-001 byte-identical
// baseline on any config-read failure.
//
// cmd is accepted so future web/profile wiring can resolve the project root
// from the command's flags; init uses cwd (the default).
func loadSessionWorktreeConfig(cmd *cobra.Command) *config.Config {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	loader := config.NewLoader()
	cfg, err := loader.Load(filepath.Join(cwd, ".moai"))
	if err != nil || cfg == nil {
		return nil
	}
	return cfg
}
