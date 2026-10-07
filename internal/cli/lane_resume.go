// lane_resume.go — the resume emergency path's launcher-side pieces
// (SPEC-SESSION-CC-VERSION-001 §B.3): the pure child-argv assembler
// (REQ-SCV-008), the valueless-token validation (REQ-SCV-009), and the token
// detector the relaunch guard judges (REQ-SCV-010's dual-spelling clause).
//
// Since SPEC-SESSION-CC-VERSION-002 the two interpreters read the child
// argv as TWO SEGMENTS (REQ-SCV-011) through the claude option model of
// REQ-SCV-012 (lane_resume_model.go): the launcher segment before MoAI's
// `--`, the claude segment after it, and everything after Claude's own
// second `--` is prompt text. The predecessor's three hand-synced value
// tables are gone — the claude side is the derived model with its
// compile-time snapshot fallback, and the launcher side is the repo-owned
// launcherSegmentValueFlags surface.
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

// launcherSegmentValueFlags is the LAUNCHER's own value-taking surface —
// decided by the launcher's own parsers (parseProfileFlag refuses a
// flag-shaped value, parseFactoryFlag likewise; both error before any
// launch), so it is repo-owned and not drift-prone the way a hand-synced
// claude table was. Its flags consume a value ONLY in the pre-separator
// segment (card-review round 4, P2): after MoAI's `--` the argv belongs to
// claude, where `-p` is the boolean --print, not the value-taking --profile,
// and none of these flags exist — letting them consume there manufactured a
// phantom value that shielded a valueless --resume.
var launcherSegmentValueFlags = map[string]bool{
	"-p": true, "--profile": true, "--branch": true, "--factory-run": true,
	"--leader": true, "--clear-policy": true, "-m": true,
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
// that flag as the value). The validator's polarity (REQ-SCV-013): a
// required-value option's next token is consumed in both modes, an
// optional-value option's next token is consumed silently (ambiguity never
// refuses), and a boolean or unknown token consumes nothing — the only
// refusal stays the definitive valueless resume. Interpretation splits at
// MoAI's separator (round 4): the launcher's value flags consume only BEFORE
// it — after it the argv belongs to claude, where they do not exist. A `--`
// is never consumed as an option's value: it counts as the separator
// instead, and the scan stops at Claude's separator. It returns nil when
// every resume token carries a value.
func validateResumeArgs(args []string) error {
	if _, valueless := scanResumeArgs(args, false); valueless {
		return errors.New(resumeRequiresValueError)
	}
	return nil
}

// scanResumeArgs is the single two-segment walk behind both interpreters
// (REQ-SCV-011): one pass over the argv, in one of the two polarities. The
// guard polarity (guard=true) is fail-closed — only a model-known
// required-value option (and, pre-separator, a launcher-owned value flag)
// shields the token after it. The validator polarity consumes the optional-
// value option's next token silently and records the valueless resume
// instead of refusing inline, so one walk serves both refusals and
// detections. It reports whether a resume carrier was seen and whether a
// definitive valueless resume token stands before Claude's separator.
func scanResumeArgs(args []string, guard bool) (carrier, valuelessResume bool) {
	seen := false
	valueless := false
	separators := 0
	moaiSeparated := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == argSeparator {
			separators++
			if separators >= 2 {
				return seen, valueless // Claude's separator: everything after is prompt text
			}
			moaiSeparated = true
			continue
		}
		if arg == resumeFlag || arg == resumeFlagShort {
			seen = true
			if i+1 >= len(args) || !resumeValuePlausible(args[i+1]) {
				valueless = true // the definitive valueless resume; a later token is judged on its own
				continue
			}
			i++ // the value is consumed, not judged
			continue
		}
		if value, isEquals := resumeEqualsValue(arg); isEquals {
			seen = true
			if value == "" {
				valueless = true
			}
			continue // a non-empty equals form is self-contained
		}
		if isShortClusterCarryingR(arg) {
			// The cluster's r-segment carries its value attached, or the shape
			// is genuinely ambiguous: the guard is fail-closed and fires; the
			// validator never refuses on the shape.
			if guard {
				seen = true
			}
			continue
		}
		switch segmentClass(arg, !moaiSeparated) {
		case claudeOptionRequiredValue:
			// Definitive: the next token is the value, in both modes — unless
			// it is the separator, which is never a value.
			if i+1 < len(args) && args[i+1] != argSeparator {
				i++
			}
		case claudeOptionOptionalValue:
			if !guard && i+1 < len(args) && args[i+1] != argSeparator {
				i++ // the validator consumes silently — ambiguity never refuses
			}
			// The guard judges the token after it: fail-closed.
		case claudeOptionBoolean, claudeOptionUnknown:
			// Nothing is consumed — both modes judge the next token (the
			// guard fires on an unprovable carrier; the validator refuses
			// a bare resume).
		}
	}
	return seen, valueless
}

// segmentClass classifies a non-resume token in its segment. Before MoAI's
// separator the launcher-owned value flags take precedence — there they are
// the launcher's own parsers' surface, whatever claude would make of the
// spelling (`-p` is the profile flag launcher-side and the boolean --print
// claude-side). After it the launcher's flags do not exist and the token
// reads the claude option model alone.
func segmentClass(arg string, preSeparator bool) claudeOptionClass {
	if preSeparator && launcherSegmentValueFlags[arg] {
		return claudeOptionRequiredValue
	}
	return activeClaudeOptionModel[arg]
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
