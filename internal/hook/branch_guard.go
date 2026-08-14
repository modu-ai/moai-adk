// Package hook : branch_guard.go — Main-Checkout Branch-State Guard
// (SPEC-WORKTREE-BRANCH-GUARD-001).
//
// Denies branch-state-changing git commands when ALL THREE hold:
// (a) the invocation occurs in the primary checkout (git-dir == git-common-dir),
// (b) the command matches a branch-state pattern, and
// (c) the invoking agent is not exempt.
// Every other path — including any git-context uncertainty — fails OPEN
// (allow + stderr advisory + audit-log append) per REQ-WBG-012.
package hook

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
)

// branchGuardExemptEnv is the sentinel env var the orchestrator sets when
// spawning Agent(manager-git, ...) for Late-Branch closure work. Closes the
// gap that HookInput.AgentType is populated only for main-thread `claude
// --agent` invocations (REQ-WBG-011b).
const branchGuardExemptEnv = "MOAI_BRANCH_GUARD_EXEMPT"

// branchGuardAuditRelPath is the fail-open advisory log path, relative to the
// handler's projectDir. Appended on every fail-open event (REQ-WBG-012).
const branchGuardAuditRelPath = ".moai/logs/branch-guard-audit.log"

// branchGuardViolationPrefix is the sentinel reason prefix the orchestrator
// pattern-matches without parsing the full reason (REQ-WBG-001).
const branchGuardViolationPrefix = "BRANCH_GUARD_VIOLATION"

// branchStatePattern pairs a compiled branch-state regex with a human-readable
// deny-reason suffix naming the matched command class.
type branchStatePattern struct {
	re     *regexp.Regexp
	suffix string
}

// branchStatePatterns is the named, blankable pattern set. M6's deny-origin
// test (AC-WBG-010) swaps this to nil to prove a deny comes from
// checkBranchState rather than checkBashCommand.
//
// The "non-flag token after subcommand" rule ([^\s-]) keeps list-only and
// path-restore invocations from matching:
//   - `git checkout -- <path>`  → token `--` starts with `-` → no match (path restore)
//   - `git branch -v` / `-a`    → token `-v`/`-a` starts with `-` → no match (list-only)
//
// This honors SPEC §E edge case E-5 and plan.md §F "Permitted". The delegation
// prompt's table lists `\S` for two rows; `\S` would match `git checkout --
// file` and `git branch -v` (false positives), contradicting the same prompt's
// "non-flag token after subcommand" rule. `[^\s-]` is the implementation of
// that stated rule. Residual: `git checkout <file>` (single-file restore, bare
// name) still matches because it is lexically indistinguishable from
// `git checkout <branch>`; operators use the explicit `git checkout -- <file>`
// form to restore files in the primary checkout (documented limitation,
// §E Residual-risk).
//
// SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 REQ-2 read-only refinement:
//   - `git merge` now anchors on trailing whitespace `\bgit\s+merge\s` so the
//     read-only `git merge-base ...` (whitespace-free `-base` suffix) does NOT
//     match. Actual `git merge <branch>` still matches. A bare `git merge`
//     (no operand) is treated as mutating and still denies — operators rarely
//     invoke it read-only.
//   - `git stash` now requires EITHER bare end-of-input OR one of the mutating
//     subcommands (push/pop/apply/drop) as the trailing token. The read-only
//     `git stash list` / `git stash show` forms are excluded because their
//     trailing token is not in the mutating set and the bare-prefix branch is
//     anchored to end-of-string. AC-REQ-2a/2b/2d.
//
// Patterns are case-insensitive (compiled with the (?i) prefix, matching the
// existing compilePatterns convention in pre_tool.go).
//
// @MX:ANCHOR: [AUTO] branchStatePatterns — branch-state regex SSOT; M2 deny-origin test swaps this var.
// @MX:REASON: fan_in >= 3 (matchBranchStateCommand, M6 TestBranchGuard_CheckBranchStateOrigin, future callers)
var branchStatePatterns = func() []branchStatePattern {
	specs := []struct {
		pattern string
		suffix  string
	}{
		{`\bgit\s+switch\b`, "git switch"},
		{`\bgit\s+checkout\s+(-b\s+)?[^\s-]`, "git checkout <branch/-b>"},
		{`\bgit\s+branch\s+(-[dDmM]\s+)?[^\s-]`, "git branch"},
		{`\bgit\s+reset\s+--hard\b`, "git reset --hard"},
		// `git stash` followed by EITHER a mutating subcommand (push/pop/apply/
		// drop), end-of-input, OR a command separator/operator boundary ([;&|]).
		// The separator branch catches bare `git stash` embedded in a compound
		// command (`git stash && git status`, `git stash; ...`, `git stash | ...`)
		// — bare stash defaults to `git stash push` (mutating), so it MUST deny
		// even when chained (sync-audit F1). Excludes the read-only forms
		// `git stash list` / `git stash show` because their trailing token is
		// neither a mutating subcommand nor a separator nor end-of-input.
		// REQ-2 AC-REQ-2a/2b/2d.
		{`\bgit\s+stash(\s+(push|pop|apply|drop)\b|\s*[;&|]|$)`, "git stash"},
		{`\bgit\s+rebase\b`, "git rebase"},
		// Trailing whitespace after `merge` excludes `git merge-base`
		// (read-only). REQ-2 AC-REQ-2c/2e. Note: a bare `git merge` with no
		// operand is intentionally NOT matched here — `git merge` with no branch
		// argument is a no-op error from git itself, and the common dangerous
		// form always carries a branch argument (`git merge feature/x`).
		{`\bgit\s+merge\s`, "git merge"},
	}
	out := make([]branchStatePattern, 0, len(specs))
	for _, s := range specs {
		re, err := regexp.Compile("(?i)" + s.pattern)
		if err != nil {
			slog.Warn("branch_guard: failed to compile pattern", "pattern", s.pattern, "error", err)
			continue
		}
		out = append(out, branchStatePattern{re: re, suffix: s.suffix})
	}
	return out
}()

// matchBranchStateCommand returns the deny-reason suffix of the first
// branch-state pattern matching command, and a bool indicating whether any
// pattern matched. Used by checkBranchState (M2) and by M1 pattern-set tests.
func matchBranchStateCommand(command string) (string, bool) {
	for _, p := range branchStatePatterns {
		if p.re.MatchString(command) {
			return p.suffix, true
		}
	}
	return "", false
}

// isExemptAgent returns true when the invoking agent identity is the trusted
// manager-git agent OR the sentinel MOAI_BRANCH_GUARD_EXEMPT=1 env var is set
// (REQ-WBG-011). Identity-based and unconditional within the identity — NOT
// scoped to any internal execution phase. A nil input is treated as
// non-exempt unless the env var is set.
func isExemptAgent(input *HookInput) bool {
	if os.Getenv(branchGuardExemptEnv) == "1" {
		return true
	}
	if input == nil {
		return false
	}
	// Identity check against the trusted manager-git agent. The literal is
	// retained (not extracted to a const) so the AC-WBG-011 grep for
	// 'AgentType == "manager-git"' mechanically confirms the path is present.
	return input.AgentType == "manager-git"
}

// isPrimaryCheckout returns true when projectDir is the primary git checkout
// (absolute git-dir == absolute git-common-dir). Returns (false, error) on any
// uncertainty (non-git dir, missing git binary, rev-parse non-zero with no
// fallback resolution) so the caller fails OPEN per REQ-WBG-012.
//
// The resolution itself lives in internal/core/git (ResolveGitDirs, extracted
// per SPEC-KANBAN-BOARD-001 REQ-KB-005 taking the REQ-KW-018 extraction
// disposition) — this caller keeps its boolean contract and delegates. The
// fallback decision lives INSIDE that dispatcher; callers do not invoke the
// fallback directly (direct invocation is a vacuous pass per AC-WBG-005).
func isPrimaryCheckout(projectDir string) (bool, error) {
	return gitcore.IsPrimaryCheckout(projectDir)
}

// checkBranchState returns DecisionDeny + a "BRANCH_GUARD_VIOLATION: <suffix>
// in primary checkout (...)" reason when ALL THREE hold: (a) primary checkout
// AT THE COMMAND'S ACTUAL CWD, (b) command matches a branch-state pattern,
// (c) invoking agent is not exempt. Otherwise it returns ("", "") (allow
// fall-through). On git-context uncertainty it fails OPEN: returns ("", "")
// AND writes an advisory to stderr plus appends a structured entry to
// .moai/logs/branch-guard-audit.log (REQ-WBG-012).
//
// The deny fires ONLY on positive evidence; uncertainty never denies.
//
// The projectDir argument is the AUDIT-LOG project directory — resolved by the
// caller (pre_tool.go) via $CLAUDE_PROJECT_DIR → os.Getwd() and pinned to the
// primary checkout for central logging (REQ-WBG-D-004). It is NOT the
// git-context directory. The git-context cwd — the directory the Bash command
// will actually execute in — is resolved HERE from input.CWD via
// resolveProjectRootFromInputOrEnv (SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001
// REQ-WBG-D-001, Seam A). The two were a single variable before this SPEC and
// MUST stay separate: querying the primary checkout about itself always
// answered "primary", misclassifying a worktree-resident agent's command as a
// primary-checkout violation.
func checkBranchState(input *HookInput, projectDir string) (decision string, reason string) {
	if input == nil || len(input.ToolInput) == 0 {
		return "", ""
	}
	command := extractBranchStateCommand(input.ToolInput)
	if command == "" {
		return "", ""
	}
	suffix, matched := matchBranchStateCommand(command)
	if !matched {
		return "", ""
	}
	if isExemptAgent(input) {
		return "", ""
	}
	// Seam A (SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 REQ-WBG-D-001): query the
	// git context at the command's actual cwd (input.CWD, falling back through
	// the CLAUDE_PROJECT_DIR → os.Getwd() chain), NOT the audit-log project
	// dir. This is the one-line correction that fixes the worktree
	// misclassification — isPrimaryCheckout stays a pure function of its
	// argument; only which directory the caller asks it to query changes.
	gitContextCwd := resolveProjectRootFromInputOrEnv(input, "branch_guard")
	isPrimary, err := isPrimaryCheckout(gitContextCwd)
	if err != nil {
		// Fail OPEN with advisory (REQ-WBG-012). The deny requires positive
		// evidence of a primary checkout; an error is NOT evidence. The
		// resolved cwd is recorded in the advisory so a silent
		// $CLAUDE_PROJECT_DIR fallback cannot re-introduce the bug (AP-D-003).
		appendBranchGuardAdvisory(input, projectDir, command, err, gitContextCwd)
		return "", ""
	}
	if !isPrimary {
		return "", ""
	}
	reason = fmt.Sprintf("%s: %s in primary checkout (use a worktree or invoke via manager-git)",
		branchGuardViolationPrefix, suffix)
	return DecisionDeny, reason
}

// extractBranchStateCommand parses the command string from Bash tool input
// JSON. Returns "" when the input is not parseable or lacks a command field.
func extractBranchStateCommand(toolInput json.RawMessage) string {
	var parsed map[string]any
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return ""
	}
	c, _ := parsed["command"].(string)
	return c
}

// appendBranchGuardAdvisory writes a fail-open advisory to stderr and appends a
// structured entry to <projectDir>/.moai/logs/branch-guard-audit.log
// (REQ-WBG-012). projectDir is the audit-log project directory (the primary
// checkout, per REQ-WBG-D-004 — central logging MUST stay on the primary even
// when the command cwd is a worktree). resolvedCwd is the git-context directory
// the discriminant queried (input.CWD-resolved); it is recorded in the entry
// (AP-D-003) so a silent $CLAUDE_PROJECT_DIR fallback that re-introduced the
// discriminant bug would be observable in the audit trail. Errors during
// logging are debug-level only — fail-open must never block the hook's allow
// decision.
func appendBranchGuardAdvisory(input *HookInput, projectDir, command string, cause error, resolvedCwd string) {
	sessionID := ""
	if input != nil {
		sessionID = input.SessionID
	}
	msg := fmt.Sprintf("branch_guard: fail-open for command %q at cwd %q (audit-log dir %q): %v", command, resolvedCwd, projectDir, cause)
	fmt.Fprintln(os.Stderr, msg)

	entry := fmt.Sprintf("[%s] session=%s command=%q cwd=%q cause=%v\n",
		time.Now().UTC().Format(time.RFC3339), sessionID, command, resolvedCwd, cause)
	logPath := filepath.Join(projectDir, branchGuardAuditRelPath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		slog.Debug("branch_guard: could not create audit log dir", "path", logPath, "error", err)
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Debug("branch_guard: could not open audit log", "path", logPath, "error", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(entry); err != nil {
		slog.Debug("branch_guard: could not write audit log entry", "path", logPath, "error", err)
	}
}

