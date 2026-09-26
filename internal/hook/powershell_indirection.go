package hook

// powershell_indirection.go — the unclassifiable-command policy for PowerShell
// calls (SPEC-HOOK-MATCHER-POWERSHELL-001 REQ-HMP-010, decision D2).
//
// The branch guard and the integration lock read command text with a
// POSIX-flavoured scan. Some PowerShell forms run git in a way that scan cannot
// see: the git invocation sits inside a string that Invoke-Expression or
// Start-Process executes, or inside a base64 payload passed to
// -EncodedCommand. The guards are fail-open and deny only on positive
// evidence, so such a call is allowed — and, so the gap is visible rather than
// silent, one line is appended to the guard's audit log whose reason field
// carries the word "unclassifiable".
//
// No PowerShell parser is added and an encoded payload is never decoded: an
// encoded command is unclassifiable whatever it contains. Over-detection costs
// one audit line and never denies anything, so the detector errs wide.

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// integrationLockAuditRelPath is the integration lock's audit log, relative to
// the handler's project root. It holds only unclassifiable-command lines; the
// guard's other fail-open paths keep writing their stderr advisories.
const integrationLockAuditRelPath = ".moai/logs/integration-lock-audit.log"

// Construct names recorded in an unclassifiable audit line.
const (
	constructEncodedCommand    = "encoded-command"
	constructInvokeExpression  = "invoke-expression"
	constructStartProcess      = "start-process"
	constructDynamicResolution = "dynamic-resolution"
)

// unclassifiedEvent tags the audit line and unclassifiedReason is its reason
// field; the literal word "unclassifiable" is part of the contract (AC-HMP-009).
const (
	unclassifiedEvent  = "powershell-unclassified"
	unclassifiedReason = "unclassifiable PowerShell indirection; allowed without classification (fail-open)"
)

// gitWordPattern finds git named as a word anywhere in raw command text.
var gitWordPattern = regexp.MustCompile(`(?i)\bgit\b`)

// commandSegmentSeparators end one command segment of a compound line. A bare
// "&" is PowerShell's call operator, not a separator, so it is kept.
var commandSegmentSeparators = regexp.MustCompile(`&&|\|\||[;|\n]`)

// callOperatorSubexpression matches a PowerShell call operator in command
// position followed directly by a parenthesized subexpression — the
// executable is resolved at RUNTIME (REQ-HGF-005, the measured form is
// `& (Get-Command git) switch probe`). Matched on the quote-collapsed text,
// so a `& (` carried inside quoted prose never fires it.
var callOperatorSubexpression = regexp.MustCompile(`(?:^|[(;&|][ \t]*|\n[ \t]*)&[ \t]*\(`)

// powerShellIndirection returns the name of an indirection construct in a
// PowerShell command that runs git out of the guards' sight, or "" when there
// is none. Constructs are matched on the quote-collapsed text, so a construct
// named inside a string argument (Write-Output "pwsh -enc …") does not count;
// the git operand of Invoke-Expression / Start-Process is matched on the raw
// text, because it usually sits inside the quoted string being executed.
//
// The dynamic-resolution construct (REQ-HGF-005) covers a call operator whose
// target is a parenthesized subexpression: the resolved executable is a
// runtime fact, so the call is allowed with one audit line rather than
// judged — the same treatment encoded commands get.
func powerShellIndirection(command string) string {
	scanned := strings.ToLower(substituteQuotedArguments(command))
	for _, segment := range commandSegmentSeparators.Split(scanned, -1) {
		if encodedCommandSegment(strings.Fields(segment)) {
			return constructEncodedCommand
		}
	}
	if !gitWordPattern.MatchString(command) {
		return ""
	}
	for _, tok := range strings.Fields(commandSegmentSeparators.ReplaceAllString(scanned, " ")) {
		switch strings.TrimLeft(tok, "({&") {
		case "iex", "invoke-expression":
			return constructInvokeExpression
		// saps and start are documented Start-Process aliases (REQ-HGF-006);
		// the git-word gate above still bounds their over-match — no git,
		// no line.
		case "start-process", "saps", "start":
			return constructStartProcess
		}
	}
	if callOperatorSubexpression.MatchString(scanned) {
		return constructDynamicResolution
	}
	return ""
}

// encodedCommandSegment reports whether a lower-cased token list invokes pwsh
// or powershell with an argument that PowerShell resolves to -EncodedCommand.
// The scan stops at -Command / -File, after which the remaining arguments
// belong to the script, not to pwsh.
func encodedCommandSegment(tokens []string) bool {
	for i, tok := range tokens {
		if !isPowerShellExecutable(tok) {
			continue
		}
		for _, arg := range tokens[i+1:] {
			name, ok := powerShellParameterName(arg)
			if !ok {
				continue
			}
			if isEncodedCommandParameter(name) {
				return true
			}
			if isScriptParameter(name) {
				break
			}
		}
	}
	return false
}

// isPowerShellExecutable matches pwsh / powershell, with or without a path or
// an .exe suffix.
func isPowerShellExecutable(tok string) bool {
	base := tok[strings.LastIndexAny(tok, `/\`)+1:]
	base = strings.TrimSuffix(base, ".exe")
	return base == "pwsh" || base == "powershell"
}

// powerShellParameterName strips a parameter prefix (-, --, or /) and any
// attached ":value", returning the lower-cased name.
func powerShellParameterName(arg string) (string, bool) {
	var name string
	switch {
	case strings.HasPrefix(arg, "--"):
		name = arg[2:]
	case strings.HasPrefix(arg, "-"), strings.HasPrefix(arg, "/"):
		name = arg[1:]
	default:
		return "", false
	}
	name, _, _ = strings.Cut(name, ":")
	return strings.ToLower(name), name != ""
}

// isEncodedCommandParameter matches the documented aliases -e and -ec and any
// prefix of EncodedCommand from "en" on. Measured acceptance by pwsh itself is
// recorded in the SPEC's run-phase evidence; over-matching costs one audit
// line and never a deny.
func isEncodedCommandParameter(name string) bool {
	if name == "e" || name == "ec" {
		return true
	}
	return len(name) >= 2 && strings.HasPrefix(name, "en") && strings.HasPrefix("encodedcommand", name)
}

// isScriptParameter matches -Command, -CommandWithArgs, and -File (with their
// short forms), after which the remaining arguments are the script's.
func isScriptParameter(name string) bool {
	switch name {
	case "c", "f", "cwa":
		return true
	}
	return (len(name) >= 2 && strings.HasPrefix("command", name)) ||
		strings.HasPrefix(name, "commandwith") ||
		(len(name) >= 2 && strings.HasPrefix("file", name))
}

// appendUnclassifiedAudit appends one unclassifiable-command line to
// <projectDir>/<relPath>. Logging failures are debug-level only: the call is
// allowed whether or not the line lands.
func appendUnclassifiedAudit(projectDir, relPath string, input *HookInput, construct, command, cwd string) {
	if projectDir == "" {
		return
	}
	sessionID := ""
	if input != nil {
		sessionID = input.SessionID
	}
	entry := fmt.Sprintf("[%s] event=%s session=%s construct=%s command=%q cwd=%q reason=%q\n",
		time.Now().UTC().Format(time.RFC3339), unclassifiedEvent, sessionID, construct, command, cwd, unclassifiedReason)
	logPath := filepath.Join(projectDir, relPath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		slog.Debug("unclassified audit: could not create log dir", "path", logPath, "error", err)
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Debug("unclassified audit: could not open log", "path", logPath, "error", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(entry); err != nil {
		slog.Debug("unclassified audit: could not write log entry", "path", logPath, "error", err)
	}
}
