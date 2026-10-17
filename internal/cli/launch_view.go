package cli

// launch_view.go — SPEC-LAUNCHER-ARGV-001 M1: the one canonical launch-argv
// resolver.
//
// Before this file, the launcher interpreted the same argv with several
// hand-written readers — the consumption loop in runLaunchClaude, the -w value
// scan, the --branch split, the chain-target extractor — and each of the six
// audit rounds of card t1604 found one more spelling on which two readers
// disagreed. This file is the single place that turns an argv into facts:
//
//   - the launcher consumption model (what runLaunchClaude eats and what it
//     passes through),
//   - the `--` separator meaning (the first one is the launcher's, the tail is
//     verbatim),
//   - the backend (claude) option model for the worktree selection (the LAST
//     option wins; the four spellings; the short-option cluster).
//
// The resolver and its walk helpers (worktreeFlagValue, normalizeWorktreeFlag,
// splitWorktreeBranchFlag, extractWorktreeLaunchTarget, selectBackendWorktree)
// are defined here and nowhere else. No other file on the launch path derives a
// worktree value from raw arguments: every reader receives a value this file
// produced. TestLaunchViewSingleParse scans for that.
//
// The view keeps TWO families of worktree facts apart, on purpose:
//
//   - the BACKEND selection (Selection): the last worktree option of the final
//     argv, decomposed by the backend's option model. Only the leader anchor
//     guard reads it — it must judge the tree claude will actually enter.
//   - the CONSUMER-scoped stage reads (StageA, StageB, Consumed.ChainTarget):
//     what the five non-guard readers read TODAY. They do not read one value —
//     they read the pipeline at three stages, each with its own first-match
//     rule (measured on the unrefactored tree: launcher_worktree_stage_cases_test.go).
//     SPEC §2 requires the non-leader behaviour to stay literally the same, so
//     each stage keeps its own field, computed by the same reader function it
//     always used, on the same stage input.

import (
	"fmt"
	"io"
	"strings"
)

// worktreeRead is one first-match -w / --worktree read: the value and whether a
// value was supplied (the result shape of worktreeFlagValue).
type worktreeRead struct {
	Value string
	Has   bool
}

// branchAnnotation is the launcher's own --branch consumption (card t295): the
// value the materialize step would check the new tree out at. Err carries a
// --branch parse failure, which the launch surfaces at the same point it always
// did (after the L2 prefix check).
type branchAnnotation struct {
	Present bool
	Branch  string
	Err     error
}

// launchConsumption is what runLaunchClaude's consumption loop makes of an argv.
// PassThrough is the backend-facing argv (the launcher-consumed options removed,
// the first separator dropped); the *Set fields say which launcher-owned options
// were present, with last-wins values.
type launchConsumption struct {
	PassThrough       []string
	PermissionMode    string
	PermissionModeSet bool
	Continue          bool
	Model             string
	ModelSet          bool
	// ChainTarget is the stage-C read: the worktree the chain-node populator
	// resolves from PassThrough.
	ChainTarget string
}

// worktreeSelection is the worktree the BACKEND would select from an argv.
type worktreeSelection struct {
	// Present reports that at least one worktree option stands in the argv.
	Present bool
	// Spelling names the token form that carried the selected (last) option.
	Spelling string
	// Value is its value; HasValue is false for a bare option (claude names the
	// tree itself) and for an empty equals form.
	Value    string
	HasValue bool
	// Index is the argv position of the selected option token.
	Index int
}

// Spellings recorded by selectBackendWorktree.
const (
	spellingLong       = "--worktree"
	spellingLongEquals = "--worktree="
	spellingShort      = "-w"
	spellingShortEqual = "-w="
	spellingAttached   = "-w<value>"
	spellingCluster    = "-<cluster>w<value>"
)

// launchView is the canonical resolution of one launch argv.
type launchView struct {
	// Entry is the argv the resolver was given.
	Entry []string

	// StageA is the first-match worktree read of Entry (the raw argv).
	StageA worktreeRead
	// Branch is the --branch consumption; Stripped is Entry without it.
	Branch   branchAnnotation
	Stripped []string
	// StageB is the first-match worktree read of Stripped.
	StageB worktreeRead
	// LaunchArgs is Stripped after normalizeWorktreeFlag: the argv handed to the
	// launch seam.
	LaunchArgs []string

	// Consumed is the consumption loop applied to LaunchArgs; FinalArgv is its
	// PassThrough — the backend argv without the launcher-injected
	// --permission-mode / --model / --continue flags.
	Consumed  launchConsumption
	FinalArgv []string

	// Selection is the backend-model worktree selection over FinalArgv.
	Selection worktreeSelection
}

// resolveLaunchView resolves entry once. It is pure: no filesystem, environment
// or process state is read.
//
// @MX:ANCHOR: [AUTO] the single launch-argv resolver; the cc and glm entries each call it twice (the anchor guard, then the launch readers)
// @MX:REASON: [AUTO] fan_in 4 (cc.go x2, glm.go x2); the guard and every worktree reader must judge the same view, so a change to the view shape or its walk rules moves all four sites together
// @MX:SPEC: SPEC-LAUNCHER-ARGV-001
func resolveLaunchView(entry []string) launchView {
	v := launchView{Entry: entry}
	v.StageA = readFirstWorktree(entry)

	rest, branch, has, err := splitWorktreeBranchFlag(entry)
	v.Branch = branchAnnotation{Present: has, Branch: branch, Err: err}
	v.Stripped = entry
	if err == nil && has {
		v.Stripped = rest
	}
	v.StageB = readFirstWorktree(v.Stripped)
	v.LaunchArgs = normalizeWorktreeFlag(v.Stripped)

	v.Consumed = consumeLaunchArgs(v.LaunchArgs)
	v.FinalArgv = v.Consumed.PassThrough
	v.Selection = selectBackendWorktree(v.FinalArgv)
	return v
}

// materializeBranch runs the card-t295 --branch materialization the view
// annotates: nothing without --branch, the parse error when --branch was
// malformed, otherwise the worktree is created at the existing branch.
func (v launchView) materializeBranch(warn io.Writer) error {
	if v.Branch.Err != nil {
		return v.Branch.Err
	}
	if !v.Branch.Present {
		return nil
	}
	return materializeWorktreeBranch(v.StageB, v.Branch.Branch, warn)
}

// readFirstWorktree wraps worktreeFlagValue into the stage-read shape.
func readFirstWorktree(args []string) worktreeRead {
	value, ok := worktreeFlagValue(args)
	return worktreeRead{Value: value, Has: ok}
}

// consumeLaunchArgs is runLaunchClaude's consumption loop, extracted unchanged:
// the launcher eats -b/--bypass, --permission-mode, -c/--continue, --model/-m
// (and the = forms) and drops the first `--`; everything else, and the whole
// tail after that separator, passes through verbatim. A second separator in the
// tail belongs to the backend and stays byte-for-byte intact.
func consumeLaunchArgs(extraArgs []string) launchConsumption {
	var c launchConsumption
	for i := 0; i < len(extraArgs); i++ {
		arg := extraArgs[i]
		if arg == "--" {
			c.PassThrough = append(c.PassThrough, extraArgs[i+1:]...)
			break
		}
		switch arg {
		case "-b", "--bypass":
			c.PermissionMode, c.PermissionModeSet = "bypassPermissions", true
		case "--permission-mode":
			if i+1 < len(extraArgs) {
				c.PermissionMode, c.PermissionModeSet = extraArgs[i+1], true
				i++
			}
		case "-c", "--continue":
			c.Continue = true
		case "--model", "-m":
			if i+1 < len(extraArgs) {
				c.Model, c.ModelSet = extraArgs[i+1], true
				i++
			}
		default:
			switch {
			case strings.HasPrefix(arg, "--model="):
				c.Model, c.ModelSet = strings.TrimPrefix(arg, "--model="), true
			case strings.HasPrefix(arg, "--permission-mode="):
				c.PermissionMode, c.PermissionModeSet = strings.TrimPrefix(arg, "--permission-mode="), true
			default:
				c.PassThrough = append(c.PassThrough, arg)
			}
		}
	}
	c.ChainTarget = extractWorktreeLaunchTarget(c.PassThrough)
	return c
}

// consumeAllowCardAnchor is the launcher's consumption of its own override
// flag (REQ-LA-005): it removes --allow-card-anchor from the launcher segment of
// args (before the first `--`; tokens after it are the backend's own) and
// reports whether it stood there. The flag counts only where it stands: the
// token after a required-value option is that option's VALUE — never the
// override — and stays in the argv the backend receives. Which options take the
// next token as a value is the repo's one option knowledge: the launcher-owned
// value flags plus the measured claude option model.
func consumeAllowCardAnchor(args []string) (rest []string, allowed bool) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == argSeparator {
			return append(rest, args[i:]...), allowed
		}
		if arg == allowCardAnchorFlag {
			allowed = true
			continue
		}
		rest = append(rest, arg)
		if optionValueSpan(args, i) > 0 {
			rest = append(rest, args[i+1])
			i++
		}
	}
	return rest, allowed
}

// optionValueSpan reports how many following tokens args[i] consumes as its
// value in the launcher segment: 1 when the token is an option whose next token
// is its value even when that token is flag-shaped, else 0. A whole-token
// option is looked up in the launcher-owned value flags and the claude option
// model; a bundled short cluster (`-cn <name>`: `c` boolean, `n` takes the next
// token) is decomposed by scanShortCluster — the same walk the backend
// selection uses — so one argv has one interpretation.
//
// A following `--` is decided by the launcher's own consumption, not by the
// backend model: consumeLaunchArgs eats it as the VALUE of --model /
// --permission-mode / -m, and for every other option it is the launcher's
// separator. Asking consumeLaunchArgs keeps that knowledge in one place.
func optionValueSpan(args []string, i int) int {
	arg := args[i]
	if i+1 < len(args) && args[i+1] == argSeparator {
		if c := consumeLaunchArgs([]string{arg, argSeparator}); c.ModelSet || c.PermissionModeSet {
			return 1
		}
		return 0
	}
	if launcherSegmentValueFlags[arg] || activeClaudeOptionModel[arg] == claudeOptionRequiredValue {
		if i+1 < len(args) {
			return 1
		}
		return 0
	}
	if len(arg) >= 2 && arg[0] == '-' && arg[1] != '-' {
		_, consumed := scanShortCluster(args, i, worktreeSelection{})
		return consumed
	}
	return 0
}

// --- backend option model: the worktree selection ----------------------------

// selectBackendWorktree returns the worktree the backend selects from argv: the
// LAST worktree option wins. It reads the four spellings (`--worktree v`,
// `--worktree=v`, `-w v`, attached `-wv`) and a short-option cluster in which a
// boolean precedes the value option (`-pwv`: `p` is a boolean, `w` takes the
// rest as its value). An optional-value option's value is the next token unless
// that token starts with `-` or is `--`; a `--` ends option parsing (everything
// after it is positional) unless it is the VALUE of a required-value option —
// a required-value option takes the next token even when it is flag-shaped or
// `--` (measured on claude 2.1.302: `claude --append-system-prompt -- --version`
// prints the version), and parsing continues after it. A token that is the
// value of another option is not an option. The knowledge of which options take
// values is activeClaudeOptionModel, the repo's one measured claude option
// surface. What follows an optional-value worktree option's `--` is unmeasured
// and unchanged.
func selectBackendWorktree(argv []string) worktreeSelection {
	var sel worktreeSelection
	for i := 0; i < len(argv); i++ {
		tok := argv[i]
		if tok == argSeparator {
			break // the backend's option terminator
		}
		if len(tok) < 2 || tok[0] != '-' {
			continue // a positional (a prompt, or a value already consumed)
		}
		switch {
		case tok == spellingLong || tok == spellingShort:
			sel = worktreeSelection{Present: true, Spelling: tok, Index: i}
			if v, ok := optionalValueAfter(argv, i); ok {
				sel.Value, sel.HasValue = v, true
				i++
			}
		case strings.HasPrefix(tok, spellingLongEquals):
			v := strings.TrimPrefix(tok, spellingLongEquals)
			sel = worktreeSelection{Present: true, Spelling: spellingLongEquals, Value: v, HasValue: v != "", Index: i}
		case strings.HasPrefix(tok, spellingShortEqual):
			v := strings.TrimPrefix(tok, spellingShortEqual)
			sel = worktreeSelection{Present: true, Spelling: spellingShortEqual, Value: v, HasValue: v != "", Index: i}
		case strings.HasPrefix(tok, "--"):
			i += longOptionValueSpan(argv, i)
		default:
			var consumed int
			sel, consumed = scanShortCluster(argv, i, sel)
			i += consumed
		}
	}
	return sel
}

// optionalValueAfter returns the token after argv[i] when it can be an
// optional-value option's value: present, not `--`, not flag-shaped.
func optionalValueAfter(argv []string, i int) (string, bool) {
	if i+1 < len(argv) && argv[i+1] != argSeparator && !strings.HasPrefix(argv[i+1], "-") {
		return argv[i+1], true
	}
	return "", false
}

// longOptionValueSpan reports how many following tokens a long option
// (no `=`) consumes as its value, by the claude option model.
func longOptionValueSpan(argv []string, i int) int {
	if strings.Contains(argv[i], "=") {
		return 0
	}
	switch activeClaudeOptionModel[argv[i]] {
	case claudeOptionRequiredValue:
		// A required-value option takes the next token as its value even when it
		// is `--` (measured on claude 2.1.302), and parsing continues after it.
		if i+1 < len(argv) {
			return 1
		}
	case claudeOptionOptionalValue:
		if _, ok := optionalValueAfter(argv, i); ok {
			return 1
		}
	case claudeOptionBoolean, claudeOptionUnknown:
	}
	return 0
}

// scanShortCluster decomposes the single-dash token argv[i] the way the
// backend's parser does: letters are read left to right; a boolean is skipped;
// the worktree letter `w` takes the rest of the token as its value (or the next
// token when nothing is attached); any other value-taking or unknown letter
// ends the walk, because the rest of the token is that option's value. It
// returns the (possibly updated) selection and how many following tokens the
// token consumed as a value.
func scanShortCluster(argv []string, i int, cur worktreeSelection) (worktreeSelection, int) {
	tok := argv[i]
	for j := 1; j < len(tok); j++ {
		letter := tok[j]
		attached := tok[j+1:]
		if letter == 'w' {
			spelling := spellingAttached
			if j > 1 {
				spelling = spellingCluster
			}
			if attached != "" {
				attached = strings.TrimPrefix(attached, "=")
				return worktreeSelection{Present: true, Spelling: spelling, Value: attached, HasValue: attached != "", Index: i}, 0
			}
			sel := worktreeSelection{Present: true, Spelling: spellingShort, Index: i}
			if j > 1 {
				sel.Spelling = spellingCluster
			}
			if v, ok := optionalValueAfter(argv, i); ok {
				sel.Value, sel.HasValue = v, true
				return sel, 1
			}
			return sel, 0
		}
		switch activeClaudeOptionModel["-"+string(letter)] {
		case claudeOptionBoolean:
			continue
		case claudeOptionRequiredValue:
			if attached == "" && i+1 < len(argv) {
				return cur, 1 // the next token, `--` included, is the value
			}
			return cur, 0
		case claudeOptionOptionalValue:
			if attached == "" {
				if _, ok := optionalValueAfter(argv, i); ok {
					return cur, 1
				}
			}
			return cur, 0
		case claudeOptionUnknown:
			return cur, 0
		}
	}
	return cur, 0
}

// --- first-match readers and token rewriters (moved here unchanged) -----------

// normalizeWorktreeFlag rewrites -w/--worktree[=name] forms (before any "--"
// pass-through marker) into the canonical two-token "--worktree [name]" form
// that Claude Code accepts, preserving argument order. Tokens after "--" are
// left untouched — they are already verbatim pass-through to claude.
// The name is optional: a bare -w lets claude auto-generate a worktree name.
func normalizeWorktreeFlag(args []string) []string {
	normalized := make([]string, 0, len(args)+1)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			// Everything after -- is verbatim pass-through.
			normalized = append(normalized, args[i:]...)
			break
		}
		switch {
		case arg == "-w" || arg == "--worktree":
			normalized = append(normalized, "--worktree")
			// Optional value: consume the next token unless it is another flag
			// or the pass-through marker.
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				normalized = append(normalized, args[i+1])
				i++
			}
		case strings.HasPrefix(arg, "--worktree="):
			name := strings.TrimPrefix(arg, "--worktree=")
			normalized = append(normalized, "--worktree")
			if name != "" {
				normalized = append(normalized, name)
			}
		case strings.HasPrefix(arg, "-w="):
			name := strings.TrimPrefix(arg, "-w=")
			normalized = append(normalized, "--worktree")
			if name != "" {
				normalized = append(normalized, name)
			}
		default:
			normalized = append(normalized, arg)
		}
	}

	return normalized
}

// worktreeFlagValue scans args (stopping at the "--" pass-through marker) for
// the value of the -w / --worktree flag. Returns the value and whether a
// value was supplied. Mirrors the token-shape handling of normalizeWorktreeFlag
// without mutating args. The FIRST match decides.
func worktreeFlagValue(args []string) (string, bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			// Everything after -- is verbatim pass-through — do not scan.
			return "", false
		}
		switch {
		case arg == "-w" || arg == "--worktree":
			// Optional value: the next token is the value unless it is another
			// flag or the pass-through marker.
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1], true
			}
			// Bare -w (claude auto-generates a worktree name): no value.
			return "", false
		case strings.HasPrefix(arg, "--worktree="):
			return strings.TrimPrefix(arg, "--worktree="), true
		case strings.HasPrefix(arg, "-w="):
			return strings.TrimPrefix(arg, "-w="), true
		}
	}
	return "", false
}

// splitWorktreeBranchFlag finds and removes the --branch flag from args,
// scanning only tokens before the "--" pass-through marker (tokens after it
// are verbatim backend input and are never touched). Accepts both the
// two-token `--branch <value>` and the joined `--branch=<value>` forms. A
// missing or option-shaped value is a parse error at this boundary — an
// unconsumed option-shaped token must never leak into the backend argv.
func splitWorktreeBranchFlag(args []string) (rest []string, branch string, has bool, err error) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			rest = append(rest, args[i:]...)
			return rest, branch, has, nil
		}
		switch {
		case arg == "--branch":
			has = true
			if i+1 >= len(args) || args[i+1] == "--" || strings.HasPrefix(args[i+1], "-") {
				return nil, "", true, fmt.Errorf("--branch requires a branch name")
			}
			branch = args[i+1]
			i++
		case strings.HasPrefix(arg, "--branch="):
			has = true
			branch = strings.TrimPrefix(arg, "--branch=")
			if branch == "" {
				return nil, "", true, fmt.Errorf("--branch requires a branch name")
			}
		default:
			rest = append(rest, arg)
		}
	}
	return rest, branch, has, nil
}

// extractWorktreeLaunchTarget returns the -w/--worktree value from normalized
// launch args. It understands the canonical two-token "--worktree <name>" form
// (the output of normalizeWorktreeFlag) and the "--worktree=<name>" spelling.
// A bare --worktree (claude auto-names the worktree) and flags after the "--"
// pass-through marker return "".
func extractWorktreeLaunchTarget(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			return ""
		}
		switch {
		case args[i] == "--worktree":
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1]
			}
			return ""
		case strings.HasPrefix(args[i], "--worktree="):
			return strings.TrimPrefix(args[i], "--worktree=")
		}
	}
	return ""
}

// --- entry points for callers outside the cc/glm launch path ------------------
//
// The codex launcher and the existing unit tests hand an argv straight to a
// reader. These wrappers derive the stage value HERE, then call the reader's
// value-taking core, so the derivation stays in this file.

// resolveWorktreeL2Path validates the first -w/--worktree value of args against
// the accepted worktree prefixes (validateWorktreeL2Value).
func resolveWorktreeL2Path(args []string, warn io.Writer) error {
	return validateWorktreeL2Value(readFirstWorktree(args), warn)
}

// resolveWorktreeExistingBranch scans args for a --branch flag accompanying a
// short-name -w, materializes the worktree at the existing branch, and returns
// the args with the --branch tokens removed. When no --branch flag is present
// the args are returned unchanged and nothing runs.
func resolveWorktreeExistingBranch(args []string, warn io.Writer) ([]string, error) {
	rest, branch, has, err := splitWorktreeBranchFlag(args)
	if err != nil {
		return nil, err
	}
	if !has {
		return args, nil
	}
	if err := materializeWorktreeBranch(readFirstWorktree(rest), branch, warn); err != nil {
		return nil, err
	}
	return rest, nil
}

// seedAdmittedWorktreeHooks seeds the tree the first -w value of args names.
func seedAdmittedWorktreeHooks(args []string, warn io.Writer) {
	seedAdmittedWorktreeHooksFor(readFirstWorktree(args), warn)
}

// injectChainNodeForLaunch records the chain node for the worktree the
// pass-through args name (injectChainNodeForTarget).
func injectChainNodeForLaunch(extraArgs []string, env []string, warn io.Writer) []string {
	return injectChainNodeForTarget(extractWorktreeLaunchTarget(extraArgs), env, warn)
}
