// lane_resume.go — the resume emergency path's launcher-side pieces
// (SPEC-SESSION-CC-VERSION-001 §B.3): the pure child-argv assembler
// (REQ-SCV-008), the valueless-token validation (REQ-SCV-009), and the token
// detector the relaunch guard judges (REQ-SCV-010's dual-spelling clause).
//
// The emergency form is the bare lane join `moai cc -l -- --resume
// <session-id>`: the entry parse desugars -l into the injected --name pair,
// the pass-through tokens survive into the child argv, and the launcher adds
// its settings. The t1348 proposal's `moai cc -f lane-<n>` spelling is itself
// refused (factoryFlagUsageError), and an operator --name beside -l is
// refused at the entry parse (laneFlagNameError).

package cli

import (
	"errors"
	"strings"
)

// resumeFlag is the emergency pass-through token: on the one-shot lane join
// it reaches the child argv and resumes one interrupted conversation in that
// card's worktree. Under the relaunch policy it cannot mean what it says —
// every card session the loop starts would receive it — which is what the
// guard refuses. resumeFlagShort is Claude's short alias of the same option
// and is recognized identically (card-review round 1, P1).
const (
	resumeFlag      = "--resume"
	resumeFlagShort = "-r"
)

// argSeparator is the bare `--` token. In the args these functions scan it
// appears at most twice with two owners: the FIRST one is MoAI's
// pass-through separator (the launcher's parsers treat it so, and the
// emergency form's resume tokens live after it), the SECOND is Claude's own
// argument separator, after which every token is prompt text and is never
// judged (card-review round 1, P2). Round 3's leader ruling bounds both
// modes: a `--` token is NEVER consumed as an option's value — when an
// option is followed by `--`, the separator wins, the option goes
// valueless, and the tokens after the `--` are prompt text.
const argSeparator = "--"

// claudeValueTakingOptions is the set of CLAUDE's options whose value-ness
// is DEFINITIVE: their synopsis marks a REQUIRED value (<value>), and the
// measured parser consumes the next token as that value even when it is
// flag-shaped (reviewer-verified for --append-system-prompt, card-review
// round 2). Measured from `claude --help` (Claude Code 2.1.289,
// /Users/<u>/.local/bin/claude -> …/versions/2.1.289, 2026-10-04).
// `--resume`/`-r` are value-taking too but are special-cased before this
// table (REQ-SCV-009's own value check).
//
// @MX:DEBT: required-value table may lag claude's option surface — r5 found --remote-control-session-name-prefix (required in 2.1.289) missing, so a legit call whose prefix value is literally "--resume" is falsely refused (validator + guard)
// @MX:CEILING: table hand-synced to claude 2.1.289 --help as of 2026-10-05; unknown later options repeat this class
// @MX:UPGRADE: t1515 — replace the hand-rolled scan with a real argv parse or spawn-time resume detection
var claudeValueTakingOptions = map[string]bool{
	// Claude Code 2.1.289, required <value> synopses (measured):
	"--add-dir": true, "--agent": true, "--agents": true,
	"--allowedTools": true, "--allowed-tools": true,
	"--append-system-prompt": true, "--autocompact": true, "--betas": true,
	"--debug-file": true, "--disallowedTools": true, "--disallowed-tools": true,
	"--effort": true, "--environment": true, "--fallback-model": true,
	"--file": true, "--input-format": true, "--json-schema": true,
	"--max-budget-usd": true, "--mcp-config": true, "--model": true,
	"--name": true, "-n": true, "--output-format": true,
	"--permission-mode": true, "--permission-prompts": true,
	"--plugin-dir": true, "--plugin-url": true, "--session-id": true,
	"--setting-sources": true, "--settings": true, "--system-prompt": true,
	"--system-prompt-snapshot": true, "--tools": true,
}

// launcherValueTakingOptions is the LAUNCHER's own value-taking flags — they
// consume a value ONLY in the pre-separator segment (card-review round 4,
// P2): after MoAI's `--` the argv belongs to claude, where `-p` is the
// boolean --print, not the value-taking --profile, and none of these flags
// exist — letting them consume there manufactured a phantom value that
// shielded a valueless --resume. Measured from the launcher's own parsers
// (parseProfileFlag refuses a flag-shaped value, parseFactoryFlag likewise;
// both error before any launch).
var launcherValueTakingOptions = map[string]bool{
	"-p": true, "--profile": true, "--branch": true, "--factory-run": true,
	"--leader": true, "--clear-policy": true, "-m": true,
}

// ambiguousValueOptions is the OPTIONAL-value class ([value] in the same
// measured help): their parser refuses to consume a flag-shaped token as the
// value, so whether the next token is a value or a real option DEPENDS on
// its shape — the scanner cannot definitively interpret it. The two modes
// resolve the ambiguity in opposite directions (card-review round 3, leader
// ruling): the GUARD is fail-closed and judges the next token (an ambiguous
// option followed by a resume-shaped token fires — `-w --resume <id>` cannot
// be proven not to resume), while the VALIDATOR never refuses on ambiguity
// and passes the next token silently. `-w`/`--worktree` sit here: claude's
// own synopsis marks it [name], and launcher-side the worktree parsers
// refuse a flag-shaped value the same way.
var ambiguousValueOptions = map[string]bool{
	"--cloud": true, "-d": true, "--debug": true, "--from-pr": true,
	"--prompt-suggestions": true, "--remote-control": true, "--teleport": true,
	"-w": true, "--worktree": true,
}

// isShortClusterCarryingR reports whether a token is a short-option cluster
// (single leading dash, not a `--` long option) whose body carries an `r`
// anywhere we cannot prove is a plain letter. With a value-taking `-r`, such
// a cluster resumes exactly like `-r<uuid>` (`-pr<uuid>` = `-p` + `-r<uuid>`)
// — card-review round 3, P1. Over-matching `-root` is the SAFE side for the
// guard: a false fire costs a restatable launch, a false pass leaks a resume
// across every card the loop starts. The validator never refuses on a
// cluster: the r-segment's value is attached, or the shape is genuinely
// ambiguous.
func isShortClusterCarryingR(arg string) bool {
	return strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && strings.Contains(arg, "r")
}

// resumeRequiresValueError is REQ-SCV-009's refusal text; it names the
// required form verbatim.
const resumeRequiresValueError = "--resume requires a session id (--resume <session-id>); a bare --resume or --resume= launches nothing"

// relaunchResumeRefusal is REQ-SCV-010's refusal text: the loop refuses to
// start (nothing leased, nothing launched, the token neither propagated nor
// stripped) and names the safe one-shot lane-join form verbatim.
const relaunchResumeRefusal = "factory lane: --resume cannot run under --clear-policy relaunch — every card session the loop starts would resume the same conversation in a foreign worktree; the emergency form is the one-shot lane join: moai cc -l -- --resume <session-id>"

// @MX:NOTE: pure by construction — takes no environment, filesystem, or
// process parameter and returns a fresh []string (REQ-SCV-008). On the
// one-shot path the behavior is unchanged from the inline append it replaced:
// the launcher-side args already carry the desugared --name pair, and the
// settings flag is appended at the same position.
// @MX:SPEC: SPEC-SESSION-CC-VERSION-001
// laneJoinChildArgv assembles the lane-join child argv in one pure function
// of its three inputs: the launcher-side arguments (carrying the injected
// --name pair and the `--` pass-through marker), the lane's session name, and
// the injected settings flag. The returned argv carries all three — the
// session name (injected before the marker when the launcher args carry no
// name), the settings pair, and the pass-through tokens — in a deterministic
// order, and the inputs are never mutated.
func laneJoinChildArgv(launcherArgs []string, sessionName string, settingsFlag []string) []string {
	argv := append(make([]string, 0, len(launcherArgs)+len(settingsFlag)+2), launcherArgs...)
	if sessionName != "" && !operatorSuppliedName(argv) {
		argv = insertBeforePassthrough(argv, nameFlagLong, sessionName)
	}
	return append(argv, settingsFlag...)
}

// validateResumeArgs refuses a --resume (or -r) token that carries no
// session id, in either spelling: the space form with no following token,
// and the equals form with an empty value. A space-form token followed by
// another flag is refused too — a session id never begins with "-", so a
// flag there means the value is missing (and claude would otherwise consume
// that flag as the value). The walk consumes the values of DEFINITIVE
// value-taking options and passes AMBIGUOUS ones silently — ambiguity never
// refuses (round 3 leader ruling); the only refusal stays the definitive
// valueless resume. Interpretation splits at MoAI's separator (round 4): the
// launcher's value flags consume only BEFORE it — after it the argv belongs
// to claude, where they do not exist. A `--` is never consumed as an
// option's value: it counts as the separator instead, and the scan stops at
// Claude's separator. It returns nil when every resume token carries a value.
func validateResumeArgs(args []string) error {
	separators := 0
	for i := 0; i < len(args); i++ {
		if args[i] == argSeparator {
			separators++
			if separators >= 2 {
				return nil // Claude's separator: everything after is prompt text
			}
			continue
		}
		if args[i] == resumeFlag || args[i] == resumeFlagShort {
			if i+1 >= len(args) || !resumeValuePlausible(args[i+1]) {
				return errors.New(resumeRequiresValueError)
			}
			i++ // the value is consumed; a later resume token is judged on its own
			continue
		}
		if value, isEquals := resumeEqualsValue(args[i]); isEquals {
			if value == "" {
				return errors.New(resumeRequiresValueError)
			}
			continue // a non-empty equals form is self-contained
		}
		if isShortClusterCarryingR(args[i]) {
			// The cluster's r-segment carries its value attached, or the
			// shape is genuinely ambiguous — never a refusal here.
			continue
		}
		if takesValueInSegment(args[i], separators == 0) {
			// Definitive: the next token is the value. Ambiguous: it MIGHT be
			// the value, and ambiguity never refuses. Either way it is passed
			// silently — unless it is the separator, which is never a value.
			if i+1 < len(args) && args[i+1] != argSeparator {
				i++
			}
		}
	}
	return nil
}

// takesValueInSegment reports whether the token consumes the next token as
// its value in the current segment: interpretation splits at MoAI's
// separator (round 4) — claude's options take values everywhere they appear,
// the launcher's value flags only before the separator (after it the argv
// belongs to claude, where they do not exist).
func takesValueInSegment(arg string, preSeparator bool) bool {
	if claudeValueTakingOptions[arg] || ambiguousValueOptions[arg] {
		return true
	}
	return preSeparator && launcherValueTakingOptions[arg]
}

// resumeEqualsValue reports the value a `--resume=`/`-r=`-prefixed token
// carries. ok is false for tokens in no equals form.
func resumeEqualsValue(arg string) (value string, ok bool) {
	if v, isLong := strings.CutPrefix(arg, resumeFlag+"="); isLong {
		return v, true
	}
	if v, isShort := strings.CutPrefix(arg, resumeFlagShort+"="); isShort {
		return v, true
	}
	return "", false
}

// resumeValuePlausible reports whether the token following a resume token
// can be its value.
func resumeValuePlausible(token string) bool {
	return token != "" && !strings.HasPrefix(token, "-")
}

// carriesResumeToken reports whether args contain a resume carrier before
// Claude's argument separator. The guard is FAIL-CLOSED on everything the
// scanner cannot definitively interpret (card-review round 3, leader
// ruling): a short-option cluster whose body carries an `r` (the attached
// `-r<uuid>` and clusters like `-pr<uuid>` all resume), and the token after
// an AMBIGUOUS value option (`-w --resume <id>` cannot be proven not to
// resume, so it fires). What IS definitive skips: the value after a
// required-value option (a system prompt that MENTIONS --resume is prompt
// text, not a resume), and everything after Claude's own argument separator.
// A `--` is never consumed as an option's value — it counts as the
// separator.
func carriesResumeToken(args []string) bool {
	seen := false
	separators := 0
	for i := 0; i < len(args); i++ {
		if args[i] == argSeparator {
			separators++
			if separators >= 2 {
				return seen // Claude's separator stops the scan
			}
			continue
		}
		switch {
		case args[i] == resumeFlag || args[i] == resumeFlagShort:
			seen = true
			if i+1 < len(args) && args[i+1] != argSeparator {
				i++ // its value is consumed, not judged
			}
			continue
		case strings.HasPrefix(args[i], resumeFlag+"="):
			seen = true
			continue
		case isShortClusterCarryingR(args[i]):
			seen = true // fail-closed: the cluster may carry `-r` with an attached value
			continue
		}
		if ambiguousValueOptions[args[i]] {
			// Fail-closed: the next token stays JUDGED — this is the
			// validator's mirror (round 3 ruling).
			continue
		}
		if takesValueInSegment(args[i], separators == 0) && i+1 < len(args) && args[i+1] != argSeparator {
			i++ // definitive value: never judged, never a separator
		}
	}
	return seen
}
