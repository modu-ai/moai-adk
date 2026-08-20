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
	"github.com/modu-ai/moai-adk/internal/execerr"
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

	// sessionWorktreeGitWorktreeRemove runs `git worktree remove <wtPath>`
	// (M4 session-exit disposal, REQ-SW-009).
	sessionWorktreeGitWorktreeRemove = gitWorktreeRemoveReal

	// --- M7 git-config seams (REQ-SW-018/019/021) ---
	//
	// These seams cover the FOUR config tiers applied at materialization by
	// applyWorktreeGitConfig. They are separate from the M2 configSet seam
	// because their failure semantics differ (safe.directory is a global
	// mutation; globalGet is read-only; gitVersion is a one-shot probe) and
	// because isolating them lets tests swap them without disturbing the M2
	// init.defaultBranch expectations.

	// sessionWorktreeGitSafeDirAdd registers a worktree path via
	// `git config --global --add safe.directory <path>` — the ONE permitted
	// global mutation (REQ-SW-018). Idempotent: `--add` on an existing entry
	// is a no-op (BI-7).
	sessionWorktreeGitSafeDirAdd = gitSafeDirAddReal

	// sessionWorktreeGitSafeDirUnset removes a worktree path from the global
	// safe.directory allowlist via `git config --global --unset safe.directory
	// <path>` — the R5 mitigation, invoked from M4 cleanup (and M8 PR-merge
	// cleanup) when a worktree is actually removed.
	sessionWorktreeGitSafeDirUnset = gitSafeDirUnsetReal

	// sessionWorktreeGitSafeDirGetAll reads all safe.directory entries via
	// `git config --global --get-all safe.directory` — used by tests to
	// verify idempotency (one path → one entry after N applies).
	sessionWorktreeGitSafeDirGetAll = gitSafeDirGetAllReal

	// sessionWorktreeGitGlobalGet reads a single value from the user's GLOBAL
	// git config via `git config --global --get <key>` (REQ-SW-019, read-only
	// source for user.name / user.email). NEVER mutates the global config.
	sessionWorktreeGitGlobalGet = gitGlobalGetReal

	// sessionWorktreeGitVersion probes the installed git version once
	// (BI-6 / EC-9). The helper uses `git -C <wt> config` (writes local repo
	// config in all git versions), so the `--worktree` flag — which needs git
	// ≥ 2.20 AND extensions.worktreeConfig — is never required; when git < 2.20
	// is detected the helper emits an informative notice (EC-9) but the
	// application proceeds via the universal local-config path.
	sessionWorktreeGitVersion = gitVersionReal

	// sessionWorktreeGitStatusPorcelain runs `git -C <wtPath> status
	// --porcelain` and returns its stdout (M4 dirty guard, REQ-SW-010; shared
	// with the M8 PR-merge path, REQ-SW-024).
	sessionWorktreeGitStatusPorcelain = gitStatusPorcelainReal
)

// SessionExitCleanupNoticePrefix is the literal prefix of the session-exit
// removal notice. It MUST NOT collide with the M8 PR-merge notice prefix
// ("removed by PR-merge cleanup:") so the two notices are distinguishable in
// combined output (REQ-SW-009 / AC-SW-009 / EC-13).
const SessionExitCleanupNoticePrefix = "removed by session-exit cleanup:"

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
	wtPath, err := materializeSessionWorktree(branch, out)
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
// git config plus the M7 worktree-scoped config tiers.
//
// BI-1 decision (progress.md §E.2): no reusable `worktree.Add` Go helper
// exists in internal/cli/worktree/ or internal/workflow/ (only mock-interface
// Add methods in _test.go). M2 shells out to `git worktree add` directly
// rather than extracting a helper — the shell-out is simpler and the
// extraction would be premature given M3/M6 will reuse this same wrapper.
func materializeSessionWorktree(branch string, out io.Writer) (string, error) {
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
	// convenience applied best-effort (the error is intentionally swallowed).
	_ = sessionWorktreeGitConfigSet(wtPath, "init.defaultBranch", "main")

	// M7 (REQ-SW-018/019/021): apply the four worktree-scoped config tiers.
	// The helper is ADDITIVE — it handles tiers 1 (safe.directory), 2
	// (identity), and 4 (options); tier 3 (init.defaultBranch) is owned by
	// the M2 direct call above. A helper result is best-effort and never
	// aborts materialization (the worktree is already on disk and usable).
	_ = applyWorktreeGitConfig(wtPath, out)
	return wtPath, nil
}

// --- real implementations (overridable in tests via the seams above) ---

// gitWorktreeAddReal runs `git worktree add -b <branch> <dest>` and returns
// the absolute destination path on success.
func gitWorktreeAddReal(destDir, branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "add", "-b", branch, destDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		// execerr.StatusDetail, not %w: a raw *exec.ExitError chain would be
		// mistaken for an intentional ExitCoder at the cmd/moai seam (t130).
		return "", fmt.Errorf("%s (%s)", strings.TrimSpace(string(out)), execerr.StatusDetail(err))
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

// gitSafeDirAddReal registers a worktree path via
// `git config --global --add safe.directory <path>` (REQ-SW-018). Idempotent:
// git's `--add` is a no-op when the entry already exists for the exact value.
func gitSafeDirAddReal(path string) error {
	cmd := exec.Command("git", "config", "--global", "--add", "safe.directory", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s (%s)", strings.TrimSpace(string(out)), execerr.StatusDetail(err))
	}
	return nil
}

// gitSafeDirUnsetReal removes a worktree path from the global safe.directory
// allowlist via `git config --global --unset safe.directory <path>` (R5
// mitigation). A non-zero exit when the entry is absent is treated as success
// (already gone — the desired state).
func gitSafeDirUnsetReal(path string) error {
	cmd := exec.Command("git", "config", "--global", "--unset", "safe.directory", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		// `--unset` on a non-matching entry exits 5 — already gone, which is
		// the desired state. Treat any exit code as success for cleanup paths
		// (the entry being absent is not an error for the unset intent).
		_ = strings.TrimSpace(string(out))
	}
	return nil
}

// gitSafeDirGetAllReal reads all safe.directory entries from the global config
// via `git config --global --get-all safe.directory`. Empty stdout (no entries)
// is valid.
func gitSafeDirGetAllReal() ([]string, error) {
	out, err := exec.Command("git", "config", "--global", "--get-all", "safe.directory").Output()
	if err != nil {
		// `--get-all` exits 1 when the key is absent — not an error for our
		// reader.
		return nil, nil
	}
	var entries []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			entries = append(entries, line)
		}
	}
	return entries, nil
}

// gitGlobalGetReal reads a single value from the user's GLOBAL git config via
// `git config --global --get <key>` (REQ-SW-019, read-only). An empty result
// (no value set) returns "" with a nil error — the caller treats empty as the
// no-identity / no-op signal.
func gitGlobalGetReal(key string) string {
	out, err := exec.Command("git", "config", "--global", "--get", key).Output()
	if err != nil {
		// `--get` exits 1 when the key is unset — that is the empty case, not
		// an error.
		return ""
	}
	return strings.TrimSpace(string(out))
}

// gitVersionInfo is the parsed `git --version` output. SupportsWorktreeConfig
// gates the `--worktree` extension flag (git ≥ 2.20, December 2018) — see BI-6.
type gitVersionInfo struct {
	Major, Minor, Patch int
}

// SupportsWorktreeConfig reports whether `git config --worktree` (the
// per-worktree config file managed via extensions.worktreeConfig) is available.
// git ≥ 2.20 supports it. Our helper does NOT depend on this flag — it applies
// worktree-scoped config via `git -C <wt> config` (local repo config) which
// works in all git versions with worktree support — but the predicate drives
// the EC-9 fallback notice.
func (v gitVersionInfo) SupportsWorktreeConfig() bool {
	if v.Major > 2 {
		return true
	}
	if v.Major == 2 && v.Minor >= 20 {
		return true
	}
	return false
}

// gitVersionReal probes the installed git version via `git --version`. Any
// probe failure degrades to a zero gitVersionInfo (major 0), which
// SupportsWorktreeConfig treats as unsupported — the conservative, fail-open
// choice (the helper still proceeds via local config; the notice fires).
//
// @MX:NOTE: [AUTO] git version parser — `git version 2.50.1` shape (Apple Git
// suffix tolerated); a parse failure degrades to zero, not an error, because
// the helper's local-config path is version-independent.
func gitVersionReal() gitVersionInfo {
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return gitVersionInfo{}
	}
	return parseGitVersion(string(out))
}

// parseGitVersion extracts the major.minor.patch triple from a `git --version`
// string. Tolerates trailing suffixes (e.g. "2.50.1 (Apple Git-155)"). Returns
// a zero gitVersionInfo when the prefix does not match the expected shape.
func parseGitVersion(s string) gitVersionInfo {
	// Expected shape: "git version 2.50.1" optionally followed by suffix.
	idx := strings.Index(s, "git version ")
	if idx < 0 {
		return gitVersionInfo{}
	}
	rest := strings.TrimSpace(s[idx+len("git version "):])
	// Take the first whitespace-delimited token (the version), drop suffix.
	if sp := strings.IndexAny(rest, " \n"); sp >= 0 {
		rest = rest[:sp]
	}
	parts := strings.Split(rest, ".")
	if len(parts) < 2 {
		return gitVersionInfo{}
	}
	v := gitVersionInfo{}
	v.Major = atoiOrZero(parts[0])
	v.Minor = atoiOrZero(parts[1])
	if len(parts) >= 3 {
		v.Patch = atoiOrZero(parts[2])
	}
	return v
}

// atoiOrZero parses a decimal integer, returning 0 on any non-numeric input.
func atoiOrZero(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// worktreeGitConfigResult records what applyWorktreeGitConfig applied. Each
// field is independently observable so tests can assert per-tier without
// coupling to notice text.
type worktreeGitConfigResult struct {
	// safeDirectoryApplied is true when `git config --global --add
	// safe.directory <wtPath>` succeeded (tier 1, REQ-SW-018).
	safeDirectoryApplied bool

	// identityName / identityEmail are the values read read-only from the
	// global gitconfig and applied worktree-scoped (tier 2, REQ-SW-019).
	// Empty when the global carries no value for that field.
	identityName  string
	identityEmail string

	// identityNoop is true when BOTH global user.name and user.email are empty
	// (EC-14 verified no-op).
	identityNoop bool

	// gitVersionFallback is true when git < 2.20 was detected (BI-6 / EC-9).
	// The helper's local-config application path proceeds regardless; this
	// flag drives the informational notice.
	gitVersionFallback bool

	// optionsApplied is true when REQ-SW-021 opt-in options were applied.
	// Currently always false: the profile carries NO signingkey/gpg opt-in
	// fields (verified — ProfilePreferences has no such schema), so tier 4 is
	// a documented no-op (tracked schema debt, E7 blocker).
	optionsApplied bool
}

// applyWorktreeGitConfig applies the M7 config tiers to a freshly-materialized
// session worktree. It runs AFTER `git worktree add` succeeds and BEFORE the
// subcommand proceeds with shared-state mutation. Every tier is best-effort:
// the worktree is returned usable regardless of any single tier's failure
// (fail-open, matching REQ-SW-004's spirit).
//
// Tiers applied (plan.md §F M7):
//  1. safe.directory (REQ-SW-018, MUST-PASS) — `git config --global --add
//     safe.directory <wtPath>`. The ONE permitted global mutation. Without it
//     git 2.35.2+ rejects the worktree ("detected dubious ownership").
//  2. Global-gitconfig identity (REQ-SW-019, MUST-PASS) — read user.name AND
//     user.email read-only from ~/.gitconfig; when BOTH non-empty apply both
//     worktree-scoped. Empty (no global identity) → verified no-op (EC-14).
//     Partial (one empty) → apply the non-empty one (EC-8). The profile is NOT
//     consulted (carries Name only, no Email — source invariant).
//  3. init.defaultBranch (REQ-SW-020) — OWNED by M2's direct call in
//     materializeSessionWorktree (additive design: the helper does NOT
//     duplicate it; consolidating would change M2's seam expectations and risk
//     test-machine git pollution).
//  4. Options (REQ-SW-021, SHALL/opt-in) — verified no-op for now
//     (ProfilePreferences lacks signingkey/gpg schema). core.hooksPath is
//     REMOVED at v0.2.1 and is NEVER set.
//
// A stderr notice (out) names the applied identity (or the no-op) per REQ-SW-019
// / R7, and names the git-version fallback per EC-9.
//
// @MX:ANCHOR: [AUTO] worktree-scoped git config helper (M7)
// @MX:REASON: REQ-SW-018/019/021 — four config tiers at materialization; safe.directory is the sole permitted global mutation and MUST be idempotent; identity is read-only from global gitconfig (profile is NOT the source)
func applyWorktreeGitConfig(wtPath string, out io.Writer) worktreeGitConfigResult {
	res := worktreeGitConfigResult{}

	// --- Tier 1: safe.directory (REQ-SW-018) ---
	if err := sessionWorktreeGitSafeDirAdd(wtPath); err == nil {
		res.safeDirectoryApplied = true
	}
	// A safe.directory failure is non-fatal (fail-open): the worktree is
	// returned usable; the user may see git's dubious-ownership warning and
	// can register the entry manually.

	// --- Tier 2: global-gitconfig identity (REQ-SW-019) ---
	name := sessionWorktreeGitGlobalGet("user.name")
	email := sessionWorktreeGitGlobalGet("user.email")
	switch {
	case name == "" && email == "":
		// EC-14: verified no-op. git inherits unchanged.
		res.identityNoop = true
	case name != "" && email != "":
		// Both present — apply both worktree-scoped.
		_ = sessionWorktreeGitConfigSet(wtPath, "user.name", name)
		_ = sessionWorktreeGitConfigSet(wtPath, "user.email", email)
		res.identityName, res.identityEmail = name, email
	default:
		// EC-8: partial — apply whichever field is non-empty, leave the
		// empty one to git's default resolution.
		if name != "" {
			_ = sessionWorktreeGitConfigSet(wtPath, "user.name", name)
			res.identityName = name
		}
		if email != "" {
			_ = sessionWorktreeGitConfigSet(wtPath, "user.email", email)
			res.identityEmail = email
		}
	}

	// --- Tier 3: init.defaultBranch (REQ-SW-020) ---
	// Owned by M2's direct sessionWorktreeGitConfigSet call inside
	// materializeSessionWorktree — intentionally NOT duplicated here.

	// --- Tier 4: options (REQ-SW-021) ---
	// Verified no-op for now: ProfilePreferences carries NO signingkey/gpg
	// opt-in fields. core.hooksPath is REMOVED at v0.2.1. When the profile
	// gains the opt-in schema, this tier applies commit.gpgsign +
	// user.signingkey (worktree-scoped) and, for new repos created by moai
	// init, core.autocrlf=input / fetch.prune=true / push.default=current.
	// Tracked as schema debt (E7 blocker).

	// --- BI-6 / EC-9: git version detection ---
	res.gitVersionFallback = !sessionWorktreeGitVersion().SupportsWorktreeConfig()

	// --- stderr notice (REQ-SW-019 / R7 + EC-9) ---
	emitWorktreeGitConfigNotice(out, res)
	return res
}

// emitWorktreeGitConfigNotice writes the user-facing identity + fallback
// notice. The notice is non-blocking (errors to out are swallowed) — it
// informs the user what identity the worktree will commit under, or that no
// global identity was found (so the user can configure one).
func emitWorktreeGitConfigNotice(out io.Writer, res worktreeGitConfigResult) {
	switch {
	case res.identityNoop:
		_, _ = fmt.Fprintln(out, "moai: no global git identity found (user.name/user.email unset in ~/.gitconfig); the worktree inherits git's default identity resolution")
	case res.identityName != "" && res.identityEmail != "":
		_, _ = fmt.Fprintf(out, "moai: worktree commits under global git identity %s <%s> (read-only from ~/.gitconfig)\n", res.identityName, res.identityEmail)
	case res.identityName != "":
		_, _ = fmt.Fprintf(out, "moai: worktree commits under global user.name %s (no global user.email — git's default email resolution applies)\n", res.identityName)
	case res.identityEmail != "":
		_, _ = fmt.Fprintf(out, "moai: worktree commits under global user.email %s (no global user.name — git's default name resolution applies)\n", res.identityEmail)
	}
	if res.gitVersionFallback {
		_, _ = fmt.Fprintln(out, "moai: git < 2.20 detected — worktree-scoped config applied via local repo config (the --worktree extension flag is unavailable; REQ-SW-019/021 use the universal local-config path)")
	}
}

// cleanupSessionWorktree performs session-exit disposal of a materialized
// session worktree (M4, REQ-SW-008/009/010).
//
// Three cases:
//   - Default-manual (REQ-SW-008): auto_cleanup == false (the distributed
//     default) -> worktree persists. No side effect, no notice.
//   - Opt-in auto-cleanup, clean exit (REQ-SW-009): git worktree remove +
//     stderr notice carrying SessionExitCleanupNoticePrefix.
//   - Opt-in auto-cleanup, non-clean exit: preserve for post-mortem + notice.
//
// The dirty guard (REQ-SW-010) runs immediately before any removal: uncommitted
// changes are NEVER deleted. wtPath == "" is a no-op. Every failure path is a
// non-blocking notice (REQ-SW-004 fail-open spirit) — the function NEVER returns
// an error and NEVER aborts the caller.
//
// @MX:ANCHOR: [AUTO] session-worktree exit disposal (M4 session-exit cleanup)
// @MX:REASON: REQ-SW-008/009/010 — auto_cleanup toggle gates removal; clean-exit-only + dirty-guard are non-negotiable safety invariants; notice prefix MUST stay distinct from the M8 PR-merge path
func cleanupSessionWorktree(cfg *config.Config, wtPath string, cleanExit bool, out io.Writer) {
	if wtPath == "" {
		// No worktree was materialized (feature OFF / fail-back / already-in-
		// worktree). Nothing to dispose.
		return
	}
	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
		// REQ-SW-008: default-manual — auto_cleanup is OFF (the distributed
		// default per CLAUDE.local.md §22.8). The worktree PERSISTS after exit;
		// the user disposes explicitly via `moai worktree done` / `remove`. No
		// notice (byte-identical to the baseline: nothing happens at exit).
		return
	}
	if !cleanExit {
		// REQ-SW-009 clean-exit-only: removal happens ONLY on clean exit (exit
		// 0). A non-zero exit PRESERVES the worktree for post-mortem so the
		// user can inspect the failure state.
		_, _ = fmt.Fprintf(out, "moai: session-exit cleanup skipped (non-zero exit): worktree %s preserved for post-mortem\n", wtPath)
		return
	}
	// REQ-SW-010 dirty guard: uncommitted changes are NEVER deleted. The guard
	// runs immediately before any `git worktree remove`. A status-check error
	// is fail-open (preserve) — never risk deleting a worktree whose dirtiness
	// could not be verified.
	dirty, derr := worktreeIsDirty(wtPath)
	if derr != nil {
		_, _ = fmt.Fprintf(out, "moai: session-exit cleanup skipped (dirty-check failed: %v): worktree %s preserved\n", derr, wtPath)
		return
	}
	if dirty {
		_, _ = fmt.Fprintf(out, "moai: session-exit cleanup skipped (uncommitted changes): worktree %s preserved (dispose manually via 'moai worktree remove' or 'git worktree remove')\n", wtPath)
		return
	}
	// Clean worktree + clean exit -> remove. A removal failure is non-blocking
	// (REQ-SW-004 fail-open spirit): the worktree is left on disk and a notice
	// names the failure.
	if rerr := sessionWorktreeGitWorktreeRemove(wtPath); rerr != nil {
		_, _ = fmt.Fprintf(out, "moai: session-exit cleanup failed (%v): worktree %s left on disk\n", rerr, wtPath)
		return
	}
	// R5 mitigation (BI-7 reverse): now that the worktree is physically gone,
	// unset its safe.directory entry from the global allowlist so the list
	// does not accumulate orphans. An unset of an already-absent entry is a
	// no-op (EC-10: orphaned entries are harmless, but we clean up when we can
	// because the removal path already proved the worktree is gone). This
	// unset is best-effort — a failure never re-creates the worktree and is
	// swallowed silently.
	_ = sessionWorktreeGitSafeDirUnset(wtPath)
	_, _ = fmt.Fprintf(out, "moai: %s worktree %s\n", SessionExitCleanupNoticePrefix, wtPath)
}

// worktreeIsDirty reports whether the worktree at wtPath has uncommitted
// changes (git status --porcelain non-empty). It is the SHARED guard used by
// both the session-exit cleanup (REQ-SW-010) and the M8 PR-merge cleanup
// (REQ-SW-024) — one helper, two call sites. An error reading status is
// returned so the caller can fail-open (preserve rather than risk deletion).
func worktreeIsDirty(wtPath string) (bool, error) {
	porcelain, err := sessionWorktreeGitStatusPorcelain(wtPath)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(porcelain) != "", nil
}

// gitWorktreeRemoveReal runs `git worktree remove <wtPath>`.
func gitWorktreeRemoveReal(wtPath string) error {
	cmd := exec.Command("git", "worktree", "remove", wtPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s (%s)", strings.TrimSpace(string(out)), execerr.StatusDetail(err))
	}
	return nil
}

// gitStatusPorcelainReal runs `git -C <wtPath> status --porcelain` and returns
// its stdout. Empty stdout means a clean worktree; non-empty means dirty
// (uncommitted changes present).
func gitStatusPorcelainReal(wtPath string) (string, error) {
	out, err := exec.Command("git", "-C", wtPath, "status", "--porcelain").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
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
