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
	"strings"
	"time"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
)

// branchGuardExemptEnv is the sentinel env var that exempts a session from the
// guard, complementing the AgentType identity axis (REQ-WBG-011b).
//
// Reachability — both axes are read from what arrives at THIS process, and a
// tool-spawned subagent can supply neither:
//
//   - AgentType arrives only in the hook payload, and Claude Code populates
//     agent_type for a main-thread `claude --agent <name>` launch. A subagent
//     spawned through the Agent tool sends no agent_type on PreToolUse, so the
//     identity axis cannot fire for it.
//   - This env var is read from the hook process's own environment. The hook
//     runs as a separate process spawned BEFORE the guarded command executes,
//     so an `export` inside that command cannot reach it. The variable must be
//     present in the environment Claude Code itself was launched with.
//
// Exporting the sentinel inside the command being guarded is therefore a no-op,
// and was mistaken for a broken exemption. Neither axis is defective; both are
// simply unreachable from inside a guarded command.
const branchGuardExemptEnv = "MOAI_BRANCH_GUARD_EXEMPT"

// branchGuardAuditRelPath is the fail-open advisory log path, relative to the
// handler's projectDir. Appended on every fail-open event (REQ-WBG-012).
const branchGuardAuditRelPath = ".moai/logs/branch-guard-audit.log"

// branchGuardViolationPrefix is the sentinel reason prefix the orchestrator
// pattern-matches without parsing the full reason (REQ-WBG-001).
const branchGuardViolationPrefix = "BRANCH_GUARD_VIOLATION"

// branchStatePattern pairs a branch-state matcher — a compiled regex OR a
// predicate function — with a human-readable deny-reason suffix naming the
// matched command class. The predicate form (match) carries the token-level
// `git branch` flag-class classifier (SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001);
// every other entry uses the regex form.
type branchStatePattern struct {
	re     *regexp.Regexp
	match  func(string) bool
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
// Branch-form completion (kanban card t42, 2026-08-15 measurement; superseded
// for `git branch` by SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001, card t467):
// the card's measured incident — `git branch --list develop -v` denied in the
// primary checkout — does NOT reproduce against any committed state of this
// file (pickaxe across history: no revision ever carried an undiscriminating
// `git branch` pattern). t42 added the copy flags `-c`/`-C` to the then-regex
// flag class; t467's M1 matrix then measured that the single-char class +
// "non-flag token" rule under-matched every other mutation form
// (`-f`/`--force`, `-u` family, `-t` family, `--delete`/`--move`/`--copy`,
// `--edit-description`, combined clusters `-df`/`-vD`/`-vt`/`-vux`,
// option-prefixed creation `-q qbranch`/`--no-force nfbranch`), closing the
// residual t42 had accepted for combined short flags — the entry is now the
// token-level classifier `matchGitBranchMutation` below.
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
		// NOTE: the `git branch` entry is NO LONGER a regex — it is the
		// token-level flag-class classifier appended after this specs list
		// (matchGitBranchMutation, SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001).
		// The old `\bgit\s+branch\s+(-[dDmMcC]\s+)?[^\s-]` single-char class
		// under-matched mutation forms (-f/--force, the -u upstream family,
		// -t track, --delete/--move/--copy/--edit-description, combined
		// clusters like -df/-vD, and option-prefixed creation like
		// `-q qbranch`), which the M1 matrix measured as the defect.
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
	out := make([]branchStatePattern, 0, len(specs)+1)
	for _, s := range specs {
		re, err := regexp.Compile("(?i)" + s.pattern)
		if err != nil {
			slog.Warn("branch_guard: failed to compile pattern", "pattern", s.pattern, "error", err)
			continue
		}
		out = append(out, branchStatePattern{re: re, suffix: s.suffix})
	}
	// The `git branch` entry: predicate-matcher form (see the specs-list NOTE
	// above). Kept INSIDE the set so blanking branchStatePatterns (M6
	// deny-origin tests) disables it together with every regex entry.
	out = append(out, branchStatePattern{match: matchGitBranchMutation, suffix: "git branch"})
	return out
}()

// quotedArgumentPattern matches a single- or double-quoted span in a shell
// command. Leftmost-first alternation handles the nested case correctly: in
// `echo "it's fine"` the double-quoted span starts first and swallows the
// apostrophe, so the trailing text is not mistaken for an open single quote.
var quotedArgumentPattern = regexp.MustCompile(`'[^']*'|"[^"]*"`)

// quotedArgumentPlaceholder is what a quoted span collapses to. It is a single
// non-flag word surrounded by spaces, chosen so the substitution preserves the
// SHAPE of the command while discarding the span's contents:
//
//   - non-flag, so `git checkout -b "feat/x"` still presents an operand after
//     `-b` and keeps matching. Blanking the span outright would have silently
//     un-guarded every branch-state command whose branch name was quoted;
//   - surrounded by spaces, so the tokens on either side cannot fuse into a new
//     word that matches by accident.
const quotedArgumentPlaceholder = " X "

// substituteQuotedArguments replaces every quoted span with a placeholder word
// so branch-state patterns match the command being RUN rather than text carried
// as data inside an argument.
//
// Without this, the pattern scan matched anywhere in the command string. A
// `moai todo add "… git switch …"` call was denied because the guarded text sat
// inside a quoted argument that would never execute — the command actually
// being run was `moai todo add`. Any command whose arguments quoted git prose
// was refused the same way.
//
// Residual (accepted): a command that passes git through a shell wrapper —
// `bash -c "git switch main"` — no longer matches, because the git invocation
// is inside the discarded span. The guard is opt-in, advisory in spirit, and
// fails open on every uncertainty, so under-matching a deliberately obfuscated
// form is the correct direction to err. The Claude Code worktree isolation
// guard refuses that same shape independently.
func substituteQuotedArguments(command string) string {
	return quotedArgumentPattern.ReplaceAllString(command, quotedArgumentPlaceholder)
}

// matchBranchStateCommand returns the deny-reason suffix of the first
// branch-state pattern matching command, and a bool indicating whether any
// pattern matched. Quoted arguments collapse to a placeholder first
// (substituteQuotedArguments) so a match reflects the command being invoked,
// not its data. Used by checkBranchState (M2) and by M1 pattern-set tests.
func matchBranchStateCommand(command string) (string, bool) {
	scanned := substituteQuotedArguments(command)
	for _, p := range branchStatePatterns {
		if p.match != nil {
			if p.match(scanned) {
				return p.suffix, true
			}
			continue
		}
		if p.re != nil && p.re.MatchString(scanned) {
			return p.suffix, true
		}
	}
	return "", false
}

// --- git branch flag-class discrimination (SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001) ---
//
// The corrected discrimination rule (plan.md §G): a `git branch` command is
// MUTATING (deny) iff either holds for the token stream after `git branch`:
//
//  1. Positional creation — a non-flag operand appears with NO list/filter
//     action selected and not consumed as the value of a preceding
//     value-taking flag. Covers bare creation, creation at a start point,
//     option-prefixed creation (`-- cr2`, `-q qbranch`, `--no-force nfbranch`,
//     `--color colprobe`), and creation modifiers (`--create-reflog <name>`).
//  2. Mutating flag anywhere — a short-flag cluster containing any of
//     d/D/m/M/c/C/f/t/u (leading OR mid-cluster), or a long flag whose name —
//     split at `=` first — exactly matches a member of the mutation set.
//
// Everything else allows. List/filter mode is selected by `--list`,
// `--show-current`, `-l` anywhere in a cluster, OR any filter selector
// (`--contains`/`--no-contains`/`--merged`/`--no-merged`/`--points-at`), and
// is VARIADIC: all remaining positionals are filter patterns. Long flags with
// a REQUIRED space-separated value (`--contains`, `--no-contains`, `--merged`,
// `--no-merged`, `--points-at`, `--format`, `--sort`) consume the following
// token; ATTACHED-ONLY optional-value flags (`--color`, `--abbrev`,
// `--column`) consume an attached `=<value>` only — a space-separated token
// after them is a positional (creation operand). Unknown flags fail OPEN
// (E-5): git prefix-abbreviations (`--dele` for `--delete`) and unknown
// short flags cannot be classified by full-token matching → allow, the
// documented under-match direction. Classification is case-insensitive
// (input is lowercased; a case-fold of a known mutation flag still denies).
//
// Residuals (named, fail-open): `git -C <path> branch …` breaks `git branch`
// adjacency and escapes classification; shell-wrapped forms are excluded by
// quoted-span collapse before this classifier runs.
//
// @MX:NOTE: [AUTO] git branch flag-class classifier — M1 matrix (branch_guard_flagclass_test.go) is the classification authority
// @MX:SPEC: SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001

// gitBranchCmdRe locates `git branch` occurrences (case-insensitive via the
// pre-lowered input) in the quoted-collapsed command.
var gitBranchCmdRe = regexp.MustCompile(`\bgit\s+branch\b`)

// gitBranchSeparatorRe matches command separators/operators that end the
// `git branch` segment of a compound command (E-6): the segment classifies
// alone, then any sibling patterns (e.g. `git switch`) match independently.
var gitBranchSeparatorRe = regexp.MustCompile(`&&|\|\||[;|&\n]`)

// gitBranchMutationLongFlags — whole-token long-flag mutation set (plan §G
// rule 2). Keys are the flag names without the `--` prefix and without any
// attached `=value`.
var gitBranchMutationLongFlags = map[string]bool{
	"force":            true,
	"delete":           true,
	"move":             true,
	"copy":             true,
	"set-upstream":     true,
	"set-upstream-to":  true,
	"unset-upstream":   true,
	"track":            true,
	"no-track":         true,
	"edit-description": true,
}

// gitBranchSpaceValueLongFlags — long flags with a REQUIRED space-separated
// value: they consume the following token so it is never read as a creation
// operand (Q-13/Q-14/Q-16 arity pins).
var gitBranchSpaceValueLongFlags = map[string]bool{
	"contains":    true,
	"no-contains": true,
	"merged":      true,
	"no-merged":   true,
	"points-at":   true,
	"format":      true,
	"sort":        true,
}

// gitBranchFilterSelectors — the subset of space-value flags that ALSO select
// variadic list/filter mode: every positional after the consumed value is a
// filter pattern, not a creation operand.
var gitBranchFilterSelectors = map[string]bool{
	"contains":    true,
	"no-contains": true,
	"merged":      true,
	"no-merged":   true,
	"points-at":   true,
}

// gitBranchListSelectors — list-action selectors that take NO value.
var gitBranchListSelectors = map[string]bool{
	"list":         true,
	"show-current": true,
}

// gitBranchKnownLongFlags — the full known long-flag universe (the sets above
// plus the attached-only optional-value flags). Used ONLY by the
// prefix-abbreviation check: a strict prefix of a known flag is a git
// prefix-abbreviation (`--dele` → `--delete`) that full-token matching cannot
// classify → the command allows (E-5 fail-open).
var gitBranchKnownLongFlags = func() map[string]bool {
	m := map[string]bool{"color": true, "abbrev": true, "column": true}
	for k := range gitBranchMutationLongFlags {
		m[k] = true
	}
	for k := range gitBranchSpaceValueLongFlags {
		m[k] = true
	}
	for k := range gitBranchListSelectors {
		m[k] = true
	}
	return m
}()

const (
	// gitBranchMutationShortChars — a short-flag cluster containing any of
	// these letters is mutating (input is lowercased, so this covers
	// d/D/m/M/c/C/f/F/t/T/u/U). No query short flag contains any of them.
	gitBranchMutationShortChars = "dmcftu"
	// gitBranchKnownShortChars — every real git branch short flag known to
	// this classifier (query: a i l q r v; mutation: c d f m t u). A cluster
	// containing a letter outside this set is not a real git branch flag —
	// git rejects the invocation — and the command is unclassifiable → allow
	// (E-5 fail-open).
	gitBranchKnownShortChars = "acdfilmqrtuv"
)

// matchGitBranchMutation reports whether the quoted-collapsed command
// contains a `git branch` invocation classified as mutating. It is the
// predicate-matcher entry of branchStatePatterns (suffix "git branch").
func matchGitBranchMutation(scanned string) bool {
	lower := strings.ToLower(scanned)
	for _, loc := range gitBranchCmdRe.FindAllStringIndex(lower, -1) {
		if classifyGitBranchTail(lower[loc[1]:]) {
			return true
		}
	}
	return false
}

// classifyGitBranchTail classifies the token stream following one `git
// branch` occurrence (already lowercased). Returns true when the invocation
// presents a mutating flag or a positional creation operand.
func classifyGitBranchTail(tail string) bool {
	if m := gitBranchSeparatorRe.FindStringIndex(tail); m != nil {
		tail = tail[:m[0]] // E-6: classify the git branch segment alone
	}
	listMode := false    // a list/filter action consumes positionals as patterns
	consumeNext := false // next token is a space-separated flag value
	for _, tok := range strings.Fields(tail) {
		if consumeNext {
			consumeNext = false
			continue
		}
		switch {
		case tok == "--":
			// End-of-options marker: later positionals are creation operands
			// (M-26). Nothing to consume; the positional arm below decides.
		case strings.HasPrefix(tok, "--"):
			name := strings.TrimPrefix(tok, "--")
			attached := false
			if i := strings.IndexByte(name, '='); i >= 0 {
				name, attached = name[:i], true
			}
			if gitBranchMutationLongFlags[name] {
				return true // rule 2: whole-token mutation membership
			}
			if gitBranchSpaceValueLongFlags[name] {
				if gitBranchFilterSelectors[name] {
					listMode = true
				}
				if !attached {
					consumeNext = true
				}
				continue
			}
			if gitBranchListSelectors[name] {
				listMode = true
				continue
			}
			// Attached-only optional-value flags (color/abbrev/column) fall
			// through: an attached =value is consumed by the split above; a
			// space-separated token after them is a positional (M-31/M-32).
			if isGitBranchFlagAbbreviation(name) {
				return false // E-5: prefix-abbreviation → unclassifiable → allow
			}
			// A genuinely unknown long flag is neutral at the flag level;
			// positional analysis still applies (M-28/M-30 measured creation).
		case len(tok) > 1 && tok[0] == '-':
			cluster := tok[1:]
			if strings.ContainsAny(cluster, gitBranchMutationShortChars) {
				return true // rule 2: cluster scan (leading or mid-cluster)
			}
			for i := 0; i < len(cluster); i++ {
				if !strings.ContainsRune(gitBranchKnownShortChars, rune(cluster[i])) {
					return false // E-5: unknown short flag → unclassifiable → allow
				}
			}
			if strings.ContainsRune(cluster, 'l') {
				listMode = true // -l anywhere in a cluster selects list mode
			}
		default:
			if !listMode {
				return true // rule 1: positional creation operand
			}
			// list/filter mode: the positional is a pattern — continue.
		}
	}
	return false
}

// isGitBranchFlagAbbreviation reports whether name is a STRICT prefix of a
// known long flag — a git prefix-abbreviation (`--dele` → `--delete`) that
// full-token matching cannot classify.
func isGitBranchFlagAbbreviation(name string) bool {
	for k := range gitBranchKnownLongFlags {
		if len(k) > len(name) && strings.HasPrefix(k, name) {
			return true
		}
	}
	return false
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
// The deny reason's remediation directs the caller to a worktree and
// deliberately does NOT suggest delegating to a manager-git subagent: both
// exemption axes are unreachable from tool-spawned subagents (see the
// branchGuardExemptEnv reachability note above), so such a delegation
// reproduces the same deny. Kanban card t43: the old "(use a worktree or
// invoke via manager-git)" wording sent orchestrator sessions down that dead
// end — one wasted turn per session, observed in two sessions.
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
	reason = fmt.Sprintf("%s: %s in primary checkout (use a worktree; the manager-git identity and %s exemptions fire only for main-thread launches, not for tool-spawned subagents)",
		branchGuardViolationPrefix, suffix, branchGuardExemptEnv)
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
