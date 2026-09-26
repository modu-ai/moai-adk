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

// insideAnySpan reports whether offset falls within one of the [start, end)
// spans, which arrive sorted and non-overlapping from FindAllStringIndex.
func insideAnySpan(offset int, spans [][]int) bool {
	for _, s := range spans {
		if offset >= s[0] && offset < s[1] {
			return true
		}
	}
	return false
}

// heredocOpenerPattern matches a heredoc redirection operator and captures its
// delimiter word in whichever of the three spellings the shell accepts:
// `<<EOF`, `<<'EOF'` and `<<"EOF"` (with `<<-` and surrounding spaces allowed).
var heredocOpenerPattern = regexp.MustCompile(`<<-?\s*(?:'([^']*)'|"([^"]*)"|([A-Za-z_][A-Za-z0-9_]*))`)

// substituteHeredocBodies replaces the BODY lines of every heredoc with the
// same placeholder quoted spans collapse to, leaving the command line that
// opens the heredoc — and every line after the terminator — intact.
//
// A heredoc body is data written to the command's stdin; it never executes.
// Without this, the pattern scan read that data as if it were the command:
// `moai handoff save --stdin … <<EOF … EOF` was denied with `git merge in
// primary checkout` because the resume body it was saving named
// `git merge --no-ff <sha>` (measured twice on 2026-09-10). The lane then
// skipped the save, which closed the handoff-record path — the guard's false
// positive cost a record, while the command it refused was never a git command
// at all. substituteQuotedArguments does not cover this: a heredoc body carries
// no quotes.
//
// The collapse is bounded to the body so the guard is not blinded: a real
// branch-state command sharing the line with a heredoc, or following its
// terminator, still matches.
func substituteHeredocBodies(command string) string {
	if !strings.Contains(command, "<<") {
		return command
	}
	lines := strings.Split(command, "\n")
	out := make([]string, 0, len(lines))
	var pending []string // delimiters opened on the current line, in order
	for _, line := range lines {
		if len(pending) > 0 {
			// Inside a body: the terminator is the delimiter alone on its line
			// (leading whitespace allowed — `<<-` strips indentation).
			if strings.TrimSpace(line) == pending[0] {
				pending = pending[1:]
				out = append(out, line)
				continue
			}
			out = append(out, quotedArgumentPlaceholder)
			continue
		}
		quoted := quotedArgumentPattern.FindAllStringIndex(line, -1)
		for _, m := range heredocOpenerPattern.FindAllStringSubmatchIndex(line, -1) {
			// A `<<EOF` inside a quoted argument is text the shell never reads
			// as a redirection. Honouring it would let any command blind the
			// guard for every following line simply by quoting the token.
			if insideAnySpan(m[0], quoted) {
				continue
			}
			for g := 1; g <= 3; g++ {
				if m[2*g] >= 0 {
					pending = append(pending, line[m[2*g]:m[2*g+1]])
					break
				}
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// shellCommentStart returns the byte offset of the `#` that opens a shell
// comment on the line, or -1 when the line opens none. POSIX: `#` begins a
// comment only at the start of a word — at line start, or immediately
// preceded by whitespace or one of the command separators `;`, `&`, `|`, `(`.
// A `#` inside a word is a literal hash (`git switch feat#123`,
// `v=bar/#x`); admitting any of those characters to the word-start set would
// blind the guard on exactly the operands branch names carry. `)` and `}` are
// deliberately NOT admitted either — the shell does open a comment after
// them, so omitting them over-matches (the guard scans text the shell
// discards) instead of blinding, the safe direction for this set
// (SPEC-GUARD-COMMENT-SCAN-001 REQ-GCS-002; residuals in that spec's §F).
func shellCommentStart(line string) int {
	for i := 0; i < len(line); i++ {
		if line[i] != '#' {
			continue
		}
		if i == 0 {
			return i
		}
		switch line[i-1] {
		case ' ', '\t', ';', '&', '|', '(':
			return i
		}
	}
	return -1
}

// substituteShellComments elides shell comments from the command before the
// pattern scan, so a branch-state pattern matches the command being RUN
// rather than prose carried in a comment.
//
// A comment is text the shell strips before execution — it can never be the
// command being run. Without this step the pattern scan read that prose as if
// it were the command: `# align with git merge --ff-only develop` matched
// `git merge` (measured 2026-09-21, SPEC-GUARD-COMMENT-SCAN-001), the same
// "data is not a command" defect class the quoted-argument and heredoc
// collapses already close.
//
// The elision is bounded to the physical line the comment opens on and no
// further (a real command on the next line is fully scannable), starts at the
// FIRST word-start `#` on the line (a second `#` later in the same run does
// not restart it), and leaves the text preceding that `#` intact — a
// branch-state command carrying a trailing comment still matches. The comment
// run is elided rather than replaced by the operand placeholder: a quoted
// span IS an operand the shell passes to the command, but a comment is
// removed by the shell, so modelling it as an operand would misstate the
// shell — `git checkout -b # x` must present no operand after `-b`.
//
// Ordering is load-bearing: this step runs LAST — outermost, on the already
// quote- and heredoc-collapsed string — so a `#` inside a quoted argument or
// a heredoc body is already part of the placeholder and opens no comment
// (`echo "text # more" ; git switch main` keeps matching).
//
// Residuals (accepted, SPEC-GUARD-COMMENT-SCAN-001 spec.md §F): a manufactured
// word-start `#` after a quoted span (`foo"bar"#baz` arrives here as
// `foo X #baz`) elides the rest of the line — blinding; the sound fix belongs
// to substituteQuotedArguments, a different step outside that SPEC's scope. A
// backslash-newline continuation leaves the next physical line's `#` mid-word
// to the shell when no whitespace precedes the backslash, but line-start here.
// `)` and `}` are not in the word-start set, so a `#` after them is scanned
// though the shell discards it (over-match, the safe side).
func substituteShellComments(command string) string {
	if !strings.Contains(command, "#") {
		return command
	}
	lines := strings.Split(command, "\n")
	for i, line := range lines {
		if idx := shellCommentStart(line); idx >= 0 {
			lines[i] = line[:idx]
		}
	}
	return strings.Join(lines, "\n")
}

// matchBranchStateCommand returns the deny-reason suffix of the first
// branch-state pattern matching command, and a bool indicating whether any
// pattern matched. Heredoc bodies and quoted arguments collapse to a
// placeholder first (substituteHeredocBodies, substituteQuotedArguments) and
// shell comments are elided last (substituteShellComments) so a match
// reflects the command being invoked, not its data. The PowerShell form
// expansions of SPEC-HOOK-GUARD-POWERSHELL-FORMS-001 run inside the same
// scan: the .exe-suffixed git spelling normalizes onto the plain one, a call
// operator's quoted call target is unquoted (it is command, not data), and
// command-position backtick escapes are de-escaped. When the outer text does
// not match, a pwsh -Command-family payload is scanned with the same
// pipeline — the payload is executed code. Used by checkBranchState (M2) and
// by M1 pattern-set tests.
func matchBranchStateCommand(command string) (string, bool) {
	if suffix, matched := matchNormalizedBranchState(command); matched {
		return suffix, true
	}
	// REQ-HGF-004: a -Command payload is executed code, not data. The full
	// pipeline — quoted-span collapse and comment elision included — is
	// computed WITHIN the payload, so a query payload passes and a nested
	// quoted span inside the payload stays data
	// (pwsh -Command "Write-Output 'git switch'" keeps its allow).
	if payload := extractPowerShellCommandPayload(command); payload != "" {
		if suffix, matched := matchNormalizedBranchState(payload); matched {
			return suffix, true
		}
	}
	return "", false
}

// matchNormalizedBranchState runs the collapse pipeline and the pattern set
// over one text (the outer command, or a -Command payload).
func matchNormalizedBranchState(command string) (string, bool) {
	scanned := substituteShellComments(normalizeGitExeSuffix(
		substituteCommandBackticks(substituteQuotedArguments(
			substituteCallOperatorTargets(substituteHeredocBodies(command))))))
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

// --- PowerShell form expansions (SPEC-HOOK-GUARD-POWERSHELL-FORMS-001) ---

// gitExeSuffixPattern matches the .exe-suffixed spelling of the git
// executable, optionally behind a path (REQ-HGF-001). The \b bounds keep a
// longer word like `gitexe` or `xgit.exe` from matching.
var gitExeSuffixPattern = regexp.MustCompile(`(?i)\bgit\.exe\b`)

// normalizeGitExeSuffix rewrites the .exe-suffixed git spelling onto the
// plain spelling so the branch-state patterns reach it. Runs on the
// quote-collapsed text — a `git.exe` inside quoted prose is already part of
// the placeholder and is never normalized.
func normalizeGitExeSuffix(command string) string {
	return gitExeSuffixPattern.ReplaceAllString(command, "git")
}

// callOperatorTargetPattern locates a PowerShell call operator (&) in command
// position — at the start of the command, after a segment separator or an
// opening parenthesis, or after whitespace following one — immediately
// followed by a quoted span. That span names the executable the call operator
// invokes: it is the ONE site where a quoted span is command, not data
// (REQ-HGF-002).
var callOperatorTargetPattern = regexp.MustCompile(`(?:^|[(;&|][ \t]*|[ \t]+)&[ \t]*('[^']*'|"[^"]*")`)

// isGitExecutableSpelling reports whether content names the git executable —
// plain, .exe-suffixed, or behind a path. Only a git-naming span gains
// command meaning; any other call target stays data (collapsed by
// substituteQuotedArguments like every quoted span), so the expansion never
// widens what counts as a git invocation.
func isGitExecutableSpelling(content string) bool {
	base := content[strings.LastIndexAny(content, `/\`)+1:]
	return strings.EqualFold(base, "git") || strings.EqualFold(base, "git.exe")
}

// substituteCallOperatorTargets unquotes ONLY the quoted span immediately
// following a PowerShell call operator in command position, and only when the
// span names git. PowerShell resolves that span as the executable name, so
// `& 'git' switch probe` must reach the same decision as `git switch probe`.
//
// The general quoted-argument collapse is untouched: a quoted `git switch`
// carried as data inside another command's argument is still collapsed to the
// placeholder by substituteQuotedArguments, which runs right after — the
// measured quoted-prose false positive (`moai todo add "… git switch …"`)
// stays protected. An & that merely sits inside quoted prose is skipped too:
// the & position is checked against the quoted spans before any rewrite.
func substituteCallOperatorTargets(command string) string {
	matches := callOperatorTargetPattern.FindAllStringSubmatchIndex(command, -1)
	if len(matches) == 0 {
		return command
	}
	quoted := quotedArgumentPattern.FindAllStringIndex(command, -1)
	var b strings.Builder
	last := 0
	for _, m := range matches {
		spanStart, spanEnd := m[2], m[3]
		if spanStart < 0 {
			continue
		}
		ampIdx := strings.LastIndexByte(command[m[0]:spanStart], '&')
		if ampIdx < 0 {
			continue
		}
		if insideAnySpan(m[0]+ampIdx, quoted) {
			continue // the & itself sits in quoted prose — the collapse will blank it
		}
		content := command[spanStart+1 : spanEnd-1]
		if !isGitExecutableSpelling(content) {
			continue // not git: the span stays data
		}
		b.WriteString(command[last:spanStart])
		b.WriteString(content)
		last = spanEnd
	}
	if last == 0 {
		return command
	}
	b.WriteString(command[last:])
	return b.String()
}

// substituteCommandBackticks removes PowerShell backtick escape characters
// from the scanned text (REQ-HGF-003). PowerShell itself de-escapes a
// backtick in command position — “git swi`tch probe“ runs as
// `git switch probe` — so the guard reaches the decision the de-escaped
// command reaches. Single-quoted spans are already collapsed to the
// placeholder at this point in the pipeline, so a literal backtick inside one
// is gone with its span; that is the bound the REQ states.
func substituteCommandBackticks(command string) string {
	if !strings.ContainsRune(command, '`') {
		return command
	}
	return strings.ReplaceAll(command, "`", "")
}

// splitPSCommandSegments splits a command on segment separators that sit
// OUTSIDE quoted spans (`;`, `|`, `&&`, newline). A bare `&` is the call
// operator, not a separator. Quoted tracking is single-level, matching the
// quote model substituteQuotedArguments uses.
func splitPSCommandSegments(command string) []string {
	var segs []string
	start := 0
	quote := byte(0)
	for i := 0; i < len(command); i++ {
		c := command[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
		case ';', '|', '\n':
			segs = append(segs, command[start:i])
			start = i + 1
		case '&':
			if i+1 < len(command) && command[i+1] == '&' {
				segs = append(segs, command[start:i])
				start = i + 2
				i++
			}
		}
	}
	return append(segs, command[start:])
}

// splitPSTokens splits one command segment into whitespace-separated tokens,
// keeping quoted spans as single tokens so a -Command payload token survives
// whole (`"git switch probe"` is one token, not five).
func splitPSTokens(segment string) []string {
	var toks []string
	var cur strings.Builder
	quote := byte(0)
	for i := 0; i < len(segment); i++ {
		c := segment[i]
		if quote != 0 {
			cur.WriteByte(c)
			if c == quote {
				quote = 0
			}
			continue
		}
		switch {
		case c == '\'' || c == '"':
			quote = c
			cur.WriteByte(c)
		case c == ' ' || c == '\t':
			if cur.Len() > 0 {
				toks = append(toks, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		toks = append(toks, cur.String())
	}
	return toks
}

// stripOuterQuotes removes ONE layer of quoting — the layer the shell or
// wrapper stripped on its way to pwsh, which is not part of the executed
// script text. Inner quoting survives and is handled by the payload's own
// pipeline (the nested-quote mutant must stay allowed).
func stripOuterQuotes(tok string) string {
	if len(tok) >= 2 {
		if (tok[0] == '"' && tok[len(tok)-1] == '"') || (tok[0] == '\'' && tok[len(tok)-1] == '\'') {
			return tok[1 : len(tok)-1]
		}
	}
	return tok
}

// commandPayloadParameter reports whether a PowerShell parameter name selects
// the -Command family — the parameter whose payload is executed code
// (REQ-HGF-004). Covers -Command, -c, -CommandWithArgs and their prefix
// abbreviations.
func commandPayloadParameter(name string) bool {
	if name == "c" || name == "cwa" {
		return true
	}
	return strings.HasPrefix("command", name) || strings.HasPrefix(name, "commandwith")
}

// extractPowerShellCommandPayload returns the script text pwsh executes for a
// -Command-family parameter, or "" when the command carries none. The payload
// is every token after the parameter, with the outer quoting of each token
// stripped — that quoting belongs to the wrapper, not the script. -File and
// other script parameters end the parameter scan: their remaining tokens
// belong to the script FILE, not to an inline payload. Each pwsh /
// powershell segment is considered independently, so an iex segment sharing
// the line does not blind the extraction.
func extractPowerShellCommandPayload(command string) string {
	for _, seg := range splitPSCommandSegments(command) {
		tokens := splitPSTokens(seg)
		for i, tok := range tokens {
			base := tok[strings.LastIndexAny(tok, `/\`)+1:]
			base = strings.TrimSuffix(base, ".exe")
			if low := strings.ToLower(base); low != "pwsh" && low != "powershell" {
				continue
			}
			for j := i + 1; j < len(tokens); j++ {
				name, ok := powerShellParameterName(tokens[j])
				if !ok {
					continue
				}
				if commandPayloadParameter(name) {
					payload := make([]string, 0, len(tokens)-j-1)
					for _, p := range tokens[j+1:] {
						payload = append(payload, stripOuterQuotes(p))
					}
					return strings.Join(payload, " ")
				}
				if isScriptParameter(name) {
					break // -File: the remaining tokens belong to the script
				}
			}
		}
	}
	return ""
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
		// D2 (SPEC-HOOK-MATCHER-POWERSHELL-001 REQ-HMP-010): a PowerShell
		// indirection the pattern scan cannot see is allowed, and recorded.
		if isPowerShellTool(input.ToolName) {
			if construct := powerShellIndirection(command); construct != "" {
				recordBranchGuardUnclassified(input, projectDir, command, construct)
			}
		}
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

// recordBranchGuardUnclassified appends one unclassifiable-command line for a
// PowerShell call the guard could not classify — but only where a classified
// command would have been judged at all: the agent is not exempt and the
// command's cwd is the primary checkout. An uncertain git context takes the
// existing fail-open advisory instead. The call is allowed on every path.
func recordBranchGuardUnclassified(input *HookInput, projectDir, command, construct string) {
	if isExemptAgent(input) {
		return
	}
	gitContextCwd := resolveProjectRootFromInputOrEnv(input, "branch_guard")
	isPrimary, err := isPrimaryCheckout(gitContextCwd)
	if err != nil {
		appendBranchGuardAdvisory(input, projectDir, command, err, gitContextCwd)
		return
	}
	if !isPrimary {
		return
	}
	appendUnclassifiedAudit(projectDir, branchGuardAuditRelPath, input, construct, command, gitContextCwd)
}

// extractBranchStateCommand parses the command string from shell tool input
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
