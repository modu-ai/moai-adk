package cli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/goal"
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

// proseWordFloor and proseBareRatio bound the prose-shape predicate below.
//
// A shell command is a verb plus operands, and its operands are flags, paths,
// assignments, or expressions — tokens carrying shell syntax. Prose is almost
// entirely bare words. The predicate keys on that ratio rather than on a word
// list, because a word list is an allowlist and an allowlist fails silently on
// what it omits — the exact way the "transcript"/"conversation" referent list
// let issue #1660's whole defect class through in every other language.
//
// Both bounds are tuned, not derived, and both exist to keep the predicate from
// firing on a real command: the floor exempts every short invocation
// (`go test ./...`, `make build`, `true`), and the ratio exempts anything
// carrying a meaningful amount of shell syntax.
const (
	proseWordFloor = 5
	proseBareRatio = 0.8
)

// shellSyntaxChars are the characters whose presence in a token is evidence the
// token is shell syntax rather than an ordinary word: redirects, pipes, globs,
// subshells, expansions, quoting, paths, and assignments.
const shellSyntaxChars = `/=|&;<>$(){}[]*?~\"'` + "`"

// proseShapedCommand reports whether a condition string that reached the
// mechanical tier reads as a sentence rather than as a command.
//
// It exists because the unrunnableCommandToken gate above is blind to the shape
// that actually reaches production: prose whose FIRST word happens to be a real
// command name. `make sure every AC row is marked PASS` resolves `make`, passes
// that gate, is handed to `sh -c`, exits non-zero, and blocks every turn-end to
// the ceiling — issue #1660's failure, surviving the first two repairs.
//
// Like its neighbour it refuses only on positive evidence and fails open: a
// short command, or one carrying shell syntax, is never flagged. A sentence that
// happens to embed a path stays catchable (the ratio tolerates a minority of
// shell-ish tokens), while a genuine long bare-word command is the known false
// positive — recoverable, because the refusal names the `cmd:` prefix that
// skips this check.
func proseShapedCommand(s string) bool {
	fields := strings.Fields(s)
	if len(fields) < proseWordFloor {
		return false
	}
	bare := 0
	for _, tok := range fields {
		if !shellSyntaxToken(tok) {
			bare++
		}
	}
	return float64(bare) >= float64(len(fields))*proseBareRatio
}

// shellSyntaxToken reports whether one token carries shell syntax. A trailing
// '.' or ',' is stripped first: those are sentence punctuation on the prose side
// ("... marked PASS.") and never meaningful on the shell side, where a dot that
// matters ("./...", "file.go") is never the token's last character.
func shellSyntaxToken(tok string) bool {
	tok = strings.TrimRight(tok, ".,")
	if tok == "" {
		return false
	}
	if strings.HasPrefix(tok, "-") {
		return true
	}
	return strings.ContainsAny(tok, shellSyntaxChars) || strings.Contains(tok, ".")
}

// proseConditionError renders the prose-shape refusal. It names BOTH remedies,
// because at this point the author's intent is genuinely unknown: the string
// reads as a sentence, but its first word is a real command.
func proseConditionError(verb, cmd string) error {
	return fmt.Errorf(
		"%s: %q was classified as a mechanical condition and would be run as a "+
			"shell command, but it reads as a sentence rather than a command — a "+
			"sentence handed to the shell exits non-zero, so the goal would block "+
			"every turn-end until the ceiling. Declare which one you meant: "+
			"%s \"model: %s\" for a claim the transcript must demonstrate, or "+
			"%s \"cmd: %s\" if it really is a command, which skips this check",
		verb, truncateCondition(cmd), verb, truncateCondition(cmd), verb, truncateCondition(cmd))
}

// armTimeConditionGate is the single arm-time gate both arming surfaces call —
// the CLI `goal arm` path and the goal_arm MCP wrapper. It is one function
// rather than two copies because the two surfaces disagreeing about which
// conditions are armable IS the defect class this gate exists to close: issue
// #1660 was reported against the MCP wrapper alone.
//
// It returns nil for every condition that should be armed, and a refusal error
// — with no state written — for one that can only block.
func armTimeConditionGate(ctx context.Context, verb, raw string, cond goal.Condition) error {
	if cond.Type != goal.ConditionMechanical || declaredMechanical(raw) {
		return nil
	}
	if tok, bad := unrunnableCommandToken(ctx, cond.Cmd); bad {
		return unrunnableConditionError(verb, tok, cond.Cmd)
	}
	if proseShapedCommand(cond.Cmd) {
		return proseConditionError(verb, cond.Cmd)
	}
	return nil
}
