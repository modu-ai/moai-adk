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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/execerr"
	"github.com/modu-ai/moai-adk/internal/gitenv"
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
	v, _ := resolveOwning(entry, "")
	return v
}

// resolveOwning is the one walk behind resolveLaunchView. With ownFlag set, the
// same walk also takes that launcher-owned flag where it stands (see
// walkLaunchArgs): the returned view is the view of the argv the backend
// receives — the flag already gone from FinalArgv when it was an override — and
// the returned indices say which tokens of entry were the overrides.
//
// The branch split and the worktree normalization do not leave the tokens alone:
// the split removes the --branch tokens, and the normalization rewrites
// `--worktree=<v>` / `-w=<v>` into two tokens, the second synthesized out of
// the value. Every token of LaunchArgs therefore carries its origin — the index
// of the typed token in entry it came from, -1 where the rewrite made it — so a
// token the user did not type as a standalone flag can never stand as the
// override, and a token is removed by its origin, not by counting equal ones.
func resolveOwning(entry []string, ownFlag string) (launchView, []int) {
	v := launchView{Entry: entry}
	v.StageA = readFirstWorktree(entry)

	rest, restOrigin, branch, has, err := splitWorktreeBranchFlagFrom(entry)
	v.Branch = branchAnnotation{Present: has, Branch: branch, Err: err}
	v.Stripped = entry
	strippedOrigin := []int(nil) // nil: v.Stripped is entry itself, origin is the identity
	if err == nil && has {
		v.Stripped, strippedOrigin = rest, restOrigin
	}
	v.StageB = readFirstWorktree(v.Stripped)
	var launchOrigin []int
	v.LaunchArgs, launchOrigin = normalizeWorktreeFlagFrom(v.Stripped)
	typed := make([]int, len(launchOrigin))
	for k, o := range launchOrigin {
		typed[k] = o
		if o >= 0 && strippedOrigin != nil {
			typed[k] = strippedOrigin[o]
		}
	}

	if err != nil {
		// A malformed --branch is refused when the launch materializes it, so there
		// is nothing to override: removing the flag must not turn an argv the launch
		// refuses into one it accepts.
		ownFlag = ""
	}
	var dropped []int
	v.Consumed, dropped, _ = walkLaunchArgs(v.LaunchArgs, typed, ownFlag)
	v.FinalArgv = v.Consumed.PassThrough
	v.Selection = selectBackendWorktree(v.FinalArgv)
	return v, dropped
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

// consumeLaunchArgs is runLaunchClaude's consumption loop: the launcher eats
// -b/--bypass, --permission-mode, -c/--continue, --model/-m (and the = forms)
// and drops the first `--` it does not eat as a value; everything else, and the
// whole tail after that separator, passes through verbatim. A second separator in
// the tail belongs to the backend and stays byte-for-byte intact.
func consumeLaunchArgs(extraArgs []string) launchConsumption {
	c, _, _ := walkLaunchArgs(extraArgs, nil, "")
	return c
}

// walkLaunchArgs is the consumption walk. consumeLaunchArgs runs it as is; the
// leader entry runs it with ownFlag, the launcher's own override flag
// (REQ-LA-005), so the override decision and the argv the backend receives are
// two outputs of ONE walk.
//
// The flag counts only where it stands: typed as a standalone token (typed[i] >= 0
// — a token the -w normalization synthesized out of an `=` form has -1 and never
// stands; nil typed reads every token as typed), in the launcher segment (before
// the first separator the walk did not eat as an option value) and not as the
// value of a backend option. The last two are read on what is left once the walk
// has dropped the launcher's own options — the sequence the backend reads — never
// on the typed argv, where a token's neighbour may be one the backend never sees.
// The flag is then dropped from PassThrough, and the second returned slice names
// the dropped tokens by their typed index (typed[i], or i when typed is nil).
//
// The third return maps PassThrough back onto the walk's input: origin[k] is the
// index in extraArgs of the raw token PassThrough[k] came from — launcher-segment
// remainder and separator tail alike. It lets a consumer rewrite the raw argv at
// a position the walk selected in PassThrough coordinates, and is aligned only
// when ownFlag is empty (an override drop rebuilds both slices together).
func walkLaunchArgs(extraArgs []string, typed []int, ownFlag string) (launchConsumption, []int, []int) {
	var c launchConsumption
	var origin []int // origin[k]: the index in extraArgs of the raw token PassThrough[k] came from
	segment := -1    // len(PassThrough) at the first separator the walk did not eat
	for i := 0; i < len(extraArgs); i++ {
		arg := extraArgs[i]
		if arg == "--" {
			segment = len(c.PassThrough)
			c.PassThrough = append(c.PassThrough, extraArgs[i+1:]...)
			for j := i + 1; j < len(extraArgs); j++ { // the verbatim tail keeps its raw positions too
				origin = append(origin, j)
			}
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
				origin = append(origin, i)
			}
		}
	}
	if segment < 0 {
		segment = len(c.PassThrough)
	}

	var dropped []int
	if ownFlag != "" {
		at := make([]int, segment) // at[k]: typed index of the launcher-segment token PassThrough[k]
		for k := range at {
			at[k] = origin[k]
			if typed != nil {
				at[k] = typed[origin[k]]
			}
		}
		if standing := standingOwnFlags(c.PassThrough[:segment], at, ownFlag); len(standing) > 0 {
			drop := make(map[int]bool, len(standing))
			for _, k := range standing {
				drop[k] = true
				dropped = append(dropped, at[k])
			}
			kept := make([]string, 0, len(c.PassThrough)-len(standing))
			keptOrigin := make([]int, 0, len(origin)-len(standing))
			for k, tok := range c.PassThrough {
				if !drop[k] {
					kept = append(kept, tok)
					keptOrigin = append(keptOrigin, origin[k])
				}
			}
			c.PassThrough = kept
			origin = keptOrigin
		}
	}
	c.ChainTarget = extractWorktreeLaunchTarget(c.PassThrough)
	return c, dropped, origin
}

// standingOwnFlags returns the positions in segment — the launcher segment as the
// backend reads it — where flag stands as an option of its own: typed as a
// standalone token (at[i] >= 0, the typed index) and not taken by the backend
// option model as another option's value.
func standingOwnFlags(segment []string, at []int, flag string) []int {
	var standing []int
	for i := 0; i < len(segment); i++ {
		tok := segment[i]
		if tok == flag && at[i] >= 0 {
			standing = append(standing, i)
			continue
		}
		if len(tok) < 2 || tok[0] != '-' {
			continue // a positional: no option to read
		}
		_, span := backendOptionStep(segment, i, worktreeSelection{})
		i += span
	}
	return standing
}

// consumeAllowCardAnchor is the launcher's consumption of its own override flag
// (REQ-LA-005): it removes --allow-card-anchor from args where it stands and
// reports whether it stood anywhere. The decision is made by the walk that
// produces the exec argv (walkLaunchArgs inside resolveOwning), so it follows
// what the backend will actually read: a flag after the launcher's separator is
// the backend's, and a flag that is a backend option's value — after the
// launcher has dropped its own options — is that option's value. Both stay in
// the argv.
func consumeAllowCardAnchor(args []string) (rest []string, allowed bool) {
	_, dropped := resolveOwning(args, allowCardAnchorFlag)
	drop := make(map[int]bool, len(dropped))
	for _, i := range dropped {
		drop[i] = true
	}
	rest = make([]string, 0, len(args))
	for i, a := range args {
		if !drop[i] {
			rest = append(rest, a)
		}
	}
	return rest, len(dropped) > 0
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
		var span int
		sel, span = backendOptionStep(argv, i, sel)
		i += span
	}
	return sel
}

// backendOptionStep reads the single option the flag-shaped token argv[i] starts,
// by the backend's option model: the worktree selection after it (cur unless the
// token is a worktree option) and how many following tokens it takes as its value.
// It is the one step both the worktree selection and the override-flag position
// walk, so a token is a value to one exactly when it is to the other.
func backendOptionStep(argv []string, i int, cur worktreeSelection) (worktreeSelection, int) {
	tok := argv[i]
	switch {
	case tok == spellingLong || tok == spellingShort:
		sel := worktreeSelection{Present: true, Spelling: tok, Index: i}
		if v, ok := optionalValueAfter(argv, i); ok {
			sel.Value, sel.HasValue = v, true
			return sel, 1
		}
		return sel, 0
	case strings.HasPrefix(tok, spellingLongEquals):
		v := strings.TrimPrefix(tok, spellingLongEquals)
		return worktreeSelection{Present: true, Spelling: spellingLongEquals, Value: v, HasValue: v != "", Index: i}, 0
	case strings.HasPrefix(tok, spellingShortEqual):
		v := strings.TrimPrefix(tok, spellingShortEqual)
		return worktreeSelection{Present: true, Spelling: spellingShortEqual, Value: v, HasValue: v != "", Index: i}, 0
	case strings.HasPrefix(tok, "--"):
		return cur, longOptionValueSpan(argv, i)
	default:
		return scanShortCluster(argv, i, cur)
	}
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
	normalized, _ := normalizeWorktreeFlagFrom(args)
	return normalized
}

// normalizeWorktreeFlagFrom is normalizeWorktreeFlag that also returns, for each
// token of the result, the index in args of the token it came from — -1 for the
// name an `=` form was split into, which nobody typed as a token of its own.
func normalizeWorktreeFlagFrom(args []string) (normalized []string, origin []int) {
	normalized = make([]string, 0, len(args)+1)
	origin = make([]int, 0, len(args)+1)
	emit := func(tok string, from int) {
		normalized = append(normalized, tok)
		origin = append(origin, from)
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			// Everything after -- is verbatim pass-through.
			for j := i; j < len(args); j++ {
				emit(args[j], j)
			}
			break
		}
		switch {
		case arg == "-w" || arg == "--worktree":
			emit("--worktree", i)
			// Optional value: consume the next token unless it is another flag
			// or the pass-through marker.
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				emit(args[i+1], i+1)
				i++
			}
		case strings.HasPrefix(arg, "--worktree="):
			name := strings.TrimPrefix(arg, "--worktree=")
			emit("--worktree", i)
			if name != "" {
				emit(name, -1)
			}
		case strings.HasPrefix(arg, "-w="):
			name := strings.TrimPrefix(arg, "-w=")
			emit("--worktree", i)
			if name != "" {
				emit(name, -1)
			}
		default:
			emit(arg, i)
		}
	}

	return normalized, origin
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
	rest, _, branch, has, err = splitWorktreeBranchFlagFrom(args)
	return rest, branch, has, err
}

// splitWorktreeBranchFlagFrom is splitWorktreeBranchFlag that also returns, for
// each token of rest, its index in args.
func splitWorktreeBranchFlagFrom(args []string) (rest []string, origin []int, branch string, has bool, err error) {
	rest = make([]string, 0, len(args))
	origin = make([]int, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			for j := i; j < len(args); j++ {
				rest = append(rest, args[j])
				origin = append(origin, j)
			}
			return rest, origin, branch, has, nil
		}
		switch {
		case arg == "--branch":
			has = true
			if i+1 >= len(args) || args[i+1] == "--" || strings.HasPrefix(args[i+1], "-") {
				return nil, nil, "", true, fmt.Errorf("--branch requires a branch name")
			}
			branch = args[i+1]
			i++
		case strings.HasPrefix(arg, "--branch="):
			has = true
			branch = strings.TrimPrefix(arg, "--branch=")
			if branch == "" {
				return nil, nil, "", true, fmt.Errorf("--branch requires a branch name")
			}
		default:
			rest = append(rest, arg)
			origin = append(origin, i)
		}
	}
	return rest, origin, branch, has, nil
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

// --- -w short-name .moai collision resolution (card t1791, GH #1813) ----------

// moaiShortNameState is the resolution of one short-name -w value against the
// project-local .moai/worktrees/<name> tree of the same name.
type moaiShortNameState int

const (
	// moaiShortNameAbsent: no same-named .moai tree exists. The backend
	// creates .claude/worktrees/<name> as before — current behavior stands.
	moaiShortNameAbsent moaiShortNameState = iota
	// moaiShortNameRegistered: the same-named .moai tree exists and git lists
	// it as a worktree of this repository. The launch re-enters it.
	moaiShortNameRegistered
	// moaiShortNameConflict: the same-named path exists but is not a
	// registered worktree — a plain directory, a file, or a repository whose
	// worktree list cannot be read. Refuse rather than guess.
	moaiShortNameConflict
)

// resolveMoaiWorktreeShortNameState is the seam the resolution reads its
// filesystem answer from. Tests override it to drive the three states without
// touching disk or git.
var resolveMoaiWorktreeShortNameState = resolveMoaiWorktreeShortNameStateReal

// resolveMoaiWorktreeShortNameStateReal answers the state of
// <projectRoot>/.moai/worktrees/<name>: absent when nothing is there,
// registered when git lists the path among the repository's worktrees, and a
// conflict for anything else. A worktree list that cannot be read is a
// conflict with the cause — the same fail-closed norm worktreeWriterRefusal
// applies to an unreadable lock state.
func resolveMoaiWorktreeShortNameStateReal(projectRoot, name string) (moaiShortNameState, error) {
	tree := filepath.Join(projectRoot, ".moai", "worktrees", name)
	// Lstat, not Stat, decides whether an entry stands: Stat follows a
	// symlink, so a DANGLING link at the tree path reads ENOENT and would
	// classify absent — letting the backend materialize the .claude twin
	// beside a same-named .moai path entry (sync-audit F2).
	if _, err := os.Lstat(tree); err != nil {
		if os.IsNotExist(err) {
			return moaiShortNameAbsent, nil
		}
		return moaiShortNameConflict, fmt.Errorf("lstat %s: %w", tree, err)
	}
	// An entry that stands but does not resolve to a directory — a dangling
	// link, a file — is a conflict, never absent.
	resolved, err := os.Stat(tree)
	if err != nil {
		return moaiShortNameConflict, nil
	}
	if !resolved.IsDir() {
		return moaiShortNameConflict, nil
	}
	cmd := exec.Command("git", "-C", projectRoot, "worktree", "list", "--porcelain")
	cmd.Env = gitenv.Env() // an inherited GIT_DIR would outrank -C
	out, err := cmd.Output()
	if err != nil {
		return moaiShortNameConflict, fmt.Errorf("git worktree list: %s", execerr.StatusDetail(err))
	}
	want := canonicalTreePath(tree)
	for _, line := range strings.Split(string(out), "\n") {
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			if canonicalTreePath(filepath.FromSlash(strings.TrimSpace(path))) == want {
				return moaiShortNameRegistered, nil
			}
		}
	}
	return moaiShortNameConflict, nil
}

// moaiWorktreeOccurrence is the worktree option the backend's option model
// honors LAST in the raw launch args — the one that decides the tree claude
// enters. Attached reports that the value rides the option token itself (the
// attached `-wcard` and cluster `-cwcard` spellings scanShortCluster reads);
// the first-match stage readers never see those forms, so an occurrence that
// would need rewriting cannot be rewritten safely and must be refused
// instead. ValueIndex is the argv position of the value when it is a token of
// its own (the two-token spellings), -1 when it rides the option token.
type moaiWorktreeOccurrence struct {
	Has        bool
	Value      string
	Attached   bool
	OptIndex   int
	ValueIndex int
}

// readBackendSelectedWorktree reads raw launch args the way the backend
// selects a worktree from the argv it receives. The worktree normalization
// runs FIRST — normalizeWorktreeFlagFrom rewrites every -w/--worktree spelling
// into the canonical two-token form, and a valueless `=` form followed by a
// positional token (`-w= card`) reads Has=false over the pre-normalization
// argv while the rewrite makes card the value the backend honors (sync-audit
// delta D2) — then the launcher consumption (walkLaunchArgs) removes the
// launcher-owned options (-b, -c, --model, --permission-mode, each value
// option with its next token) and the first "--", and the selection walk runs
// over PassThrough — the argv the backend reads — with selectBackendWorktree,
// the same walk resolveOwning's Selection takes over FinalArgv. The honored
// occurrence is mapped back through the walk's origin indices and the
// normalization's origins to the raw entry for the in-place rewrite; a value
// the normalization split out of an `=` form has origin -1, and the rewrite
// touches the `=`-form option token instead. Walking the pre-normalization
// argv instead diverges wherever a launcher-owned token sits next to a
// value-consuming option: the launcher REMOVES the token while an in-place
// skip only hides it, so the two adjacencies differ (sync-audit delta D1:
// `-w card --append-system-prompt -c -w fresh` and `-w -c card`). For a
// single -w the read agrees with worktreeFlagValue; when several stand, this
// — not the first match — is the value the collision resolution must judge
// (sync-audit F1).
func readBackendSelectedWorktree(args []string) moaiWorktreeOccurrence {
	normalized, normOrigin := normalizeWorktreeFlagFrom(args)
	c, _, passOrigin := walkLaunchArgs(normalized, nil, "")
	sel := selectBackendWorktree(c.PassThrough)
	if !sel.Present || sel.Index < 0 || sel.Index >= len(passOrigin) {
		return moaiWorktreeOccurrence{}
	}
	// rawIndex maps a PassThrough position to the raw argv through both origin
	// hops, -1 where the token is not of the raw argv at all.
	rawIndex := func(k int) int {
		at := passOrigin[k]
		if at < 0 || at >= len(normOrigin) {
			return -1
		}
		return normOrigin[at]
	}
	optIndex := rawIndex(sel.Index)
	if optIndex < 0 {
		// The selection landed on a token the normalization synthesized out of
		// an `=` form — a flag-shaped `=` value the backend's optional-value
		// rule skips past. The rewrite targets the `=`-form token that parent
		// belongs to: the nearest preceding raw token.
		for j := passOrigin[sel.Index] - 1; j >= 0; j-- {
			if j < len(normOrigin) && normOrigin[j] >= 0 {
				optIndex = normOrigin[j]
				break
			}
		}
		if optIndex < 0 {
			return moaiWorktreeOccurrence{}
		}
	}
	occ := moaiWorktreeOccurrence{
		Has:        sel.HasValue,
		Value:      sel.Value,
		Attached:   sel.Spelling == spellingAttached || sel.Spelling == spellingCluster,
		OptIndex:   optIndex,
		ValueIndex: -1,
	}
	// The two-token spellings (and a cluster whose w ends the token) take the
	// next PassThrough token as their value; the option step reports that span.
	// Its raw position comes from the same two-hop mapping — a value the
	// normalization synthesized out of an `=` form maps to -1, and the rewrite
	// touches the `=`-form option token instead.
	if _, span := backendOptionStep(c.PassThrough, sel.Index, worktreeSelection{}); span == 1 && sel.Index+1 < len(passOrigin) {
		if valueIndex := rawIndex(sel.Index + 1); valueIndex >= 0 {
			occ.ValueIndex = valueIndex
		}
	}
	return occ
}

// rewriteBackendSelectedWorktreeValue returns args with the honored
// occurrence's value replaced by value: the value token itself for the
// two-token spellings, the option token rewritten for the = forms. The token
// count never changes, so every downstream origin index stays aligned. The
// attached and cluster spellings are refused upstream and never reach this
// rewrite; a valueless = form cannot stand either (Has is false for one).
func rewriteBackendSelectedWorktreeValue(args []string, occ moaiWorktreeOccurrence, value string) []string {
	if occ.ValueIndex >= 0 {
		out := append([]string(nil), args...)
		out[occ.ValueIndex] = value
		return out
	}
	tok := args[occ.OptIndex]
	switch {
	case strings.HasPrefix(tok, spellingLongEquals):
		out := append([]string(nil), args...)
		out[occ.OptIndex] = spellingLongEquals + value
		return out
	case strings.HasPrefix(tok, spellingShortEqual):
		out := append([]string(nil), args...)
		out[occ.OptIndex] = spellingShortEqual + value
		return out
	}
	return args
}

// resolveMoaiWorktreeShortName rewrites a short-name -w/--worktree value that
// names an existing registered .moai/worktrees/<name> tree into that tree's
// absolute path, so the backend re-enters the MoAI tree instead of
// materializing a same-named .claude twin (the 2026-10-08 lane-12 incident).
// A same-named path that is not a registered worktree warns and aborts —
// silently creating the twin is the collision this resolution exists to
// refuse. Every other argv passes through unchanged: no -w, a bare -w, an
// absolute -w (already names an existing tree), a value with a path
// separator, and a name with no same-named .moai tree.
//
// The value judged is the one the BACKEND selects — the last honored worktree
// option (readBackendSelectedWorktree), not the first match the stage reads
// carry. For a single -w the two grammars agree; when several stand, the
// last-honored occurrence is the tree the backend would enter or twin, and
// that is what this resolution refuses or re-enters. An attached or cluster
// spelling that would need rewriting is refused as ambiguous launcher input:
// the first-match stage readers cannot judge those forms, so rewriting one
// would resolve a value the downstream readers never see.
//
// The entry points call this BEFORE the anchor guard and resolveLaunchView, so
// every reader downstream judges the tree the backend will actually enter.
func resolveMoaiWorktreeShortName(args []string, warn io.Writer) ([]string, error) {
	occ := readBackendSelectedWorktree(args)
	if !occ.Has || occ.Value == "" || filepath.IsAbs(occ.Value) {
		return args, nil
	}
	name := occ.Value
	if name == "." || name == ".." || strings.ContainsRune(name, filepath.Separator) || strings.ContainsRune(name, '/') {
		return args, nil // not a single-segment short name; not this resolution's input
	}
	projectRoot, err := findProjectRootFn()
	if err != nil || projectRoot == "" {
		return args, nil // no root to resolve against: keep current behavior (the launch needs a root anyway)
	}
	tree := filepath.Join(projectRoot, ".moai", "worktrees", name)
	state, err := resolveMoaiWorktreeShortNameState(projectRoot, name)
	if err != nil {
		return nil, fmt.Errorf("-w %s: %w", name, err)
	}
	switch state {
	case moaiShortNameAbsent:
		return args, nil
	case moaiShortNameRegistered:
		if occ.Attached {
			_, _ = fmt.Fprintf(warn, "moai: refusing to launch: -w %s rides the attached spelling %q in the argv, which this launcher's worktree readers cannot judge against the registered %s; spell it as `-w %s` so the same-named .moai tree resolution can see it.\n", name, args[occ.OptIndex], tree, name)
			return nil, fmt.Errorf("-w %s: attached spelling %q cannot be resolved against the registered .moai tree; refusing rather than guessing which tree the backend enters", name, args[occ.OptIndex])
		}
		return rewriteBackendSelectedWorktreeValue(args, occ, tree), nil
	default: // moaiShortNameConflict
		_, _ = fmt.Fprintf(warn, "moai: refusing to launch: -w %s names %s, which exists but is not a registered git worktree of this repository; entering it is unsafe, and creating .claude/worktrees/%s with the same name is the twin collision this launcher refuses. Remove or rename the path, or pass the tree's absolute path explicitly after verifying it.\n", name, tree, name)
		return nil, fmt.Errorf("-w %s: %s exists but is not a registered git worktree; refusing to create the same-named .claude twin", name, tree)
	}
}
