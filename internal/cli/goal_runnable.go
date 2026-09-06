package cli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// goalCmdProbeTimeout bounds the arm-time command-resolution probe. The probe
// is advisory: exceeding the bound allows the arm rather than refusing it.
const goalCmdProbeTimeout = 3 * time.Second

// shellKeywords are the compound-command openers whose first word names a shell
// construct rather than a command. `command -v if` resolves to nothing, so
// probing them would refuse a perfectly runnable condition.
var shellKeywords = map[string]bool{
	"if": true, "for": true, "while": true, "until": true,
	"case": true, "function": true,
}

// declaredMechanical reports whether the raw condition text carries an explicit
// `cmd:` declaration prefix.
//
// An explicit declaration is the author saying "this IS a command" in so many
// words, and it exempts the condition from the arm-time runnability gate below.
// The gate exists to catch prose that reached the mechanical tier by ACCIDENT —
// it has nothing to correct once the tier was chosen deliberately. Two reasons
// the exemption is not merely permissive:
//
//   - `command -v` probes the ARMING environment. A condition naming a tool that
//     exists at eval time but not at arm time — a different PATH, a container, a
//     binary the goal itself builds — is a legitimate goal the gate would refuse
//     with no way for the author to say it is wrong.
//   - The refusal message points at `cmd:` as the remedy for exactly that case.
//     Without this exemption the message names a door that does not open.
//
// A deliberate `cmd:` on genuine prose is still caught — by the eval-time exit-127
// backstop, one turn in rather than thirty.
func declaredMechanical(raw string) bool {
	m := conditionDeclarationPrefix.FindStringSubmatch(strings.TrimSpace(raw))
	return m != nil && strings.EqualFold(m[1], "cmd")
}

// unrunnableCommandToken reports the first word of a mechanical condition when
// that word positively resolves to nothing — neither an executable on PATH, nor
// a shell builtin, nor a function or alias. It is the arm-time evidence that a
// condition can never exit 0 and would block every turn-end to the ceiling.
//
// The predicate is deliberately asymmetric, because the two errors are not
// equally bad. A missed catch costs the eval-time backstop one turn; a false
// refusal blocks a legitimate goal at the door, with the user having no way to
// tell the gate it is wrong. So it refuses ONLY on positive evidence, and skips
// every shape a naive first-token extraction cannot judge:
//
//   - an empty command, or one with no fields;
//   - `FOO=bar cmd` — the first token is an assignment, not the command;
//   - a token containing '/' — a path form whose target may legitimately not
//     exist yet at arm time (a script the goal itself will produce);
//   - a token opening a shell construct: '(', '{', '!', '$', quote, backtick,
//     '#';
//   - a shell keyword (`if`, `for`, ...).
//
// It resolves via `command -v` rather than exec.LookPath because LookPath sees
// only PATH executables: `cd`, `true`, and `:` are builtins that LookPath may
// miss and that `sh -c` runs perfectly well. The token is passed as an argv
// element, never interpolated into the script, so a condition string cannot
// inject shell syntax through the probe.
func unrunnableCommandToken(ctx context.Context, cmd string) (string, bool) {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return "", false
	}
	tok := fields[0]
	if strings.Contains(tok, "=") || strings.Contains(tok, "/") {
		return "", false
	}
	if shellKeywords[strings.ToLower(tok)] {
		return "", false
	}
	if strings.IndexByte("({!$\"'`#", tok[0]) >= 0 {
		return "", false
	}
	if commandTokenResolves(ctx, tok) {
		return "", false
	}
	return tok, true
}

// commandTokenResolves probes whether `sh` can resolve tok to something it can
// run. It fails OPEN: any error that is not the probe's own non-zero exit — a
// missing shell, a deadline, a spawn failure — reports true (resolvable), so a
// broken probe never refuses an arm.
func commandTokenResolves(ctx context.Context, tok string) bool {
	pctx, cancel := context.WithTimeout(ctx, goalCmdProbeTimeout)
	defer cancel()
	// `sh -c '<script>' sh <tok>` puts tok in $1 — argv, not script text.
	probe := exec.CommandContext(pctx, "sh", "-c", `command -v -- "$1" >/dev/null 2>&1`, "sh", tok)
	err := probe.Run()
	if err == nil {
		return true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && pctx.Err() == nil {
		return false // positive evidence: sh resolved nothing
	}
	return true // probe itself failed → fail open
}

// unrunnableConditionError renders the arm-time refusal. It names the offending
// first word, says plainly that the string was routed to the shell, and points
// at the `model:` prefix — the remedy, since the classifier's substring
// fallback cannot recognize a claim written in any other language.
func unrunnableConditionError(verb, token, cmd string) error {
	return fmt.Errorf(
		"%s: %q was classified as a mechanical condition and would be run as a "+
			"shell command, but its first word %q resolves to no command — the "+
			"condition can never exit 0, so the goal would block every turn-end "+
			"until the ceiling. If this is a claim about the conversation rather "+
			"than a command, declare it explicitly: %s \"model: %s\". If it really "+
			"is a command that this environment cannot resolve yet, declare that "+
			"with the cmd: prefix, which skips this check",
		verb, truncateCondition(cmd), token, verb, truncateCondition(cmd))
}

// truncateCondition bounds a condition string quoted back in an error message;
// a multi-line ac_converge paragraph would otherwise swamp the remedy.
func truncateCondition(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = strings.TrimSpace(s[:i]) + " ..."
	}
	const maxLen = 80
	if len(s) > maxLen {
		// Trim on a rune boundary so multi-byte text is not cut mid-character.
		r := []rune(s)
		if len(r) > maxLen {
			r = r[:maxLen]
		}
		s = string(r) + " ..."
	}
	return s
}
