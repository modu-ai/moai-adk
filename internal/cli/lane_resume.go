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
// guard refuses.
const resumeFlag = "--resume"

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

// validateResumeArgs refuses a --resume token that carries no session id, in
// either spelling: the space form with no following token, and the equals
// form with an empty value. A space-form token followed by another flag is
// refused too — a session id never begins with "-", so a flag there means
// the value is missing (and claude would otherwise consume that flag as the
// value). It returns nil when every --resume token carries a value.
func validateResumeArgs(args []string) error {
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == resumeFlag:
			if i+1 >= len(args) || !resumeValuePlausible(args[i+1]) {
				return errors.New(resumeRequiresValueError)
			}
			i++ // the value is consumed; a later --resume is judged on its own
		case strings.HasPrefix(arg, resumeFlag+"="):
			if arg[len(resumeFlag)+1:] == "" {
				return errors.New(resumeRequiresValueError)
			}
		}
	}
	return nil
}

// resumeValuePlausible reports whether the token following --resume can be
// its value.
func resumeValuePlausible(token string) bool {
	return token != "" && !strings.HasPrefix(token, "-")
}

// carriesResumeToken reports whether args contain a --resume token in either
// its --resume <value> or --resume=<value> spelling — REQ-SCV-010's dual
// recognition, so the equals form cannot evade the relaunch guard.
func carriesResumeToken(args []string) bool {
	for _, arg := range args {
		if arg == resumeFlag || strings.HasPrefix(arg, resumeFlag+"=") {
			return true
		}
	}
	return false
}
