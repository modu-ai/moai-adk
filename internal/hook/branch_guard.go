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
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// execCommand is the package-level indirection over exec.Command. Tests inject
// a mock runner here (AC-WBG-005: simulates an older-git host rejecting
// --path-format=absolute so the dispatcher's fallback path is exercised).
// Restore via t.Cleanup in every test that swaps it.
var execCommand = exec.Command

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
		// Bare `git stash` (end of input) OR a mutating subcommand. Excludes
		// `git stash list` / `git stash show` (read-only). REQ-2 AC-REQ-2a/2b/2d.
		{`\bgit\s+stash(\s+(push|pop|apply|drop)\b|$)`, "git stash"},
		{`\bgit\s+rebase\b`, "git rebase"},
		// Trailing whitespace after `merge` excludes `git merge-base`. REQ-2 AC-REQ-2c/2e.
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
// Primary path (git 2.31+, March 2021): --path-format=absolute for both
// rev-parse invocations. Fallback (older git / Apple Git rejecting the flag):
// --absolute-git-dir + cwd-normalized --git-common-dir. The fallback decision
// lives INSIDE this dispatcher — callers do not invoke the fallback directly
// (direct invocation is a vacuous pass per AC-WBG-005).
func isPrimaryCheckout(projectDir string) (bool, error) {
	if projectDir == "" {
		return false, fmt.Errorf("branch_guard: empty projectDir")
	}

	// Primary path: a SINGLE rev-parse call emits both absolute paths, halving
	// the git-spawn cost versus two separate calls. This keeps the per-invocation
	// latency under the AC-WBG-010 ceiling on Windows, where each git.exe spawn
	// is expensive (issue #1225 — TestBranchGuard_Latency). `--path-format=absolute`
	// applies to every following path flag, so --git-dir and --git-common-dir
	// each print one absolute line, in argument order.
	out, err := runGitRevParse(projectDir, "--path-format=absolute", "--git-dir", "--git-common-dir")
	if err == nil {
		lines := strings.Split(out, "\n")
		if len(lines) >= 2 {
			return strings.TrimSpace(lines[0]) == strings.TrimSpace(lines[1]), nil
		}
		// Unexpected output shape; fall through to the fallback below.
	}

	// Fallback: --absolute-git-dir + cwd-normalized --git-common-dir. The bare
	// --git-common-dir form returns a repo-relative path (.git) on older git,
	// so it is normalized against projectDir before comparison.
	absGitDir, err := runGitRevParse(projectDir, "--absolute-git-dir")
	if err != nil {
		return false, fmt.Errorf("branch_guard: --absolute-git-dir failed: %w", err)
	}
	relCommon, err := runGitRevParse(projectDir, "--git-common-dir")
	if err != nil {
		return false, fmt.Errorf("branch_guard: --git-common-dir failed: %w", err)
	}
	absCommon := relCommon
	if !filepath.IsAbs(absCommon) {
		absCommon = filepath.Join(projectDir, relCommon)
	}
	return filepath.Clean(absGitDir) == filepath.Clean(absCommon), nil
}

// runGitRevParse runs `git -C projectDir rev-parse <args...>` via the
// package-level execCommand indirection and returns the trimmed stdout. Returns
// an error on non-zero exit (covers missing git binary, non-git directory, and
// unknown-flag rejection by older git).
func runGitRevParse(projectDir string, args ...string) (string, error) {
	full := append([]string{"-C", projectDir, "rev-parse"}, args...)
	cmd := execCommand("git", full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse %s: %w (%s)",
			strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// checkBranchState returns DecisionDeny + a "BRANCH_GUARD_VIOLATION: <suffix>
// in primary checkout (...)" reason when ALL THREE hold: (a) primary checkout,
// (b) command matches a branch-state pattern, (c) invoking agent is not exempt.
// Otherwise it returns ("", "") (allow fall-through). On git-context
// uncertainty it fails OPEN: returns ("", "") AND writes an advisory to stderr
// plus appends a structured entry to .moai/logs/branch-guard-audit.log
// (REQ-WBG-012).
//
// The deny fires ONLY on positive evidence; uncertainty never denies.
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
	isPrimary, err := isPrimaryCheckout(projectDir)
	if err != nil {
		// Fail OPEN with advisory (REQ-WBG-012). The deny requires positive
		// evidence of a primary checkout; an error is NOT evidence.
		appendBranchGuardAdvisory(input, projectDir, command, err)
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
// (REQ-WBG-012). Errors during logging are debug-level only — fail-open must
// never block the hook's allow decision.
func appendBranchGuardAdvisory(input *HookInput, projectDir, command string, cause error) {
	sessionID := ""
	if input != nil {
		sessionID = input.SessionID
	}
	msg := fmt.Sprintf("branch_guard: fail-open for command %q in %q: %v", command, projectDir, cause)
	fmt.Fprintln(os.Stderr, msg)

	entry := fmt.Sprintf("[%s] session=%s command=%q cause=%v\n",
		time.Now().UTC().Format(time.RFC3339), sessionID, command, cause)
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

