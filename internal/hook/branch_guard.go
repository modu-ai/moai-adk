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
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// execCommand is the package-level indirection over exec.Command. Tests inject
// a mock runner here (AC-WBG-005: simulates an older-git host rejecting
// --path-format=absolute so the dispatcher's fallback path is exercised).
// Restore via t.Cleanup in every test that swaps it.
var execCommand = exec.Command

// branchGuardExemptAgent is the trusted agent identity that may perform
// branch-state changes in the primary checkout (manager-git Phase D,
// Late-Branch closure). Identity-based and unconditional within the identity —
// NOT scoped to any phase (REQ-WBG-011).
const branchGuardExemptAgent = "manager-git"

// branchGuardExemptEnv is the sentinel env var the orchestrator sets when
// spawning Agent(manager-git, ...) for Phase D work. Closes the gap that
// HookInput.AgentType is populated only for main-thread `claude --agent`
// invocations (REQ-WBG-011b).
const branchGuardExemptEnv = "MOAI_BRANCH_GUARD_EXEMPT"

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
		{`\bgit\s+stash(\s+(push|pop|apply|drop)\b)?`, "git stash"},
		{`\bgit\s+rebase\b`, "git rebase"},
		{`\bgit\s+merge\b`, "git merge"},
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
// scoped to any phase. A nil input is treated as non-exempt unless the env var
// is set.
func isExemptAgent(input *HookInput) bool {
	if os.Getenv(branchGuardExemptEnv) == "1" {
		return true
	}
	if input == nil {
		return false
	}
	return input.AgentType == branchGuardExemptAgent
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

	// Primary path: --path-format=absolute for both rev-parse calls.
	gitDir, err := runGitRevParse(projectDir, "--path-format=absolute", "--git-dir")
	if err == nil {
		gitCommonDir, errCommon := runGitRevParse(projectDir, "--path-format=absolute", "--git-common-dir")
		if errCommon == nil {
			return gitDir == gitCommonDir, nil
		}
		// Second primary call failed; fall through to the fallback below.
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
