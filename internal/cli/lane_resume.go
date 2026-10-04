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
// judged (card-review round 1, P2).
const argSeparator = "--"

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
// that flag as the value). The scan stops at Claude's argument separator:
// tokens after the second `--` are prompt text, not options. It returns nil
// when every resume token carries a value.
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
		if value, isEquals := resumeEqualsValue(args[i]); isEquals {
			if value == "" {
				return errors.New(resumeRequiresValueError)
			}
			continue
		}
		if args[i] == resumeFlag || args[i] == resumeFlagShort {
			if i+1 >= len(args) || !resumeValuePlausible(args[i+1]) {
				return errors.New(resumeRequiresValueError)
			}
			i++ // the value is consumed; a later resume token is judged on its own
		}
	}
	return nil
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

// carriesResumeToken reports whether args contain a resume token — `--resume`
// or its `-r` alias, in space or equals spelling — before Claude's argument
// separator (tokens after it are prompt text, not options, and cannot resume
// anything). REQ-SCV-010's dual recognition extends to the alias, so no
// spelling can evade the relaunch guard.
func carriesResumeToken(args []string) bool {
	seen := false
	separators := 0
	for _, arg := range args {
		if arg == argSeparator {
			separators++
			if separators >= 2 {
				return seen // Claude's separator stops the scan
			}
			continue
		}
		if arg == resumeFlag || arg == resumeFlagShort ||
			strings.HasPrefix(arg, resumeFlag+"=") || strings.HasPrefix(arg, resumeFlagShort+"=") {
			seen = true
		}
	}
	return seen
}
