package cli

// codex_launcher.go — SPEC-CODEX-LAUNCHER-001 M3+M4
// (REQ-CL-001/002/003/012, REQ-CL-014 help copy):
// the `moai codex` command surface. Cobra registration sits in the launch
// group next to cc/glm/cg; the verb routing accepts exactly {bare, status}
// for the readout forms and {cli, app} for the launch verbs (closed sets —
// unknown tokens are rejected, never routed to a launch); --spawn moves the
// launch to a new tmux window.
//
// The readout rows come from M2's codexReadiness VERBATIM (AC-CL-004 — the
// command surface never re-words a row), and the binary/auth values come from
// the shared probe, so no second classification path forks here (REQ-CL-007).
// The launcher never writes: no directory is created, no file mutated
// (REQ-CL-013). Launch is a child process whose exit code propagates — no
// process replacement, no OS build tags (AC-CL-014).

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/execerr"
	"github.com/spf13/cobra"
)

// Diagnostics — named constants compared with == by the tests (AC-CL-002's
// "usage constant", AC-CL-011's single-line install action).
const (
	// codexUsageDiag is the one-line usage diagnostic every unknown token
	// receives (AC-CL-002). Byte-identical for all six probe tokens so the
	// rejection cannot leak which token was seen.
	codexUsageDiag = "unknown verb - usage: moai codex [status] | moai codex cli [codex-args...] | moai codex app"

	// codexInstallHint is the single diagnostic line the launch verbs print
	// when the codex binary is unresolved (AC-CL-011: exactly one line, exact
	// match; the readout forms still succeed in the same state).
	codexInstallHint = "codex not found - install the Codex CLI first: https://developers.openai.com/codex/cli"

	// codexSpawnReadoutDiag rejects --spawn on the readout forms (AC-CL-003):
	// a readout is not something to open a new window for.
	codexSpawnReadoutDiag = "--spawn applies to the launch verbs only (moai codex cli --spawn / moai codex app --spawn)"
)

// codexVerb classifies a routed token: which tokens launch and which render
// the readout is a CLOSED SET (AC-CL-002) — an unknown token never falls
// through to a launch.
type codexVerb int

const (
	codexVerbReadout codexVerb = iota // "" (bare) and "status"
	codexVerbLaunchCli
	codexVerbLaunchApp
)

// codexVerbRouting is the routing table the tests read symbolically
// (AC-CL-002): the launch-verb set derived from it must equal {cli, app} and
// the readout-token set must equal {"", status}. Absence from the table IS
// the rejection — no default branch exists.
var codexVerbRouting = map[string]codexVerb{
	"":       codexVerbReadout,
	"status": codexVerbReadout,
	"cli":    codexVerbLaunchCli,
	"app":    codexVerbLaunchApp,
}

// launches reports whether the verb class starts a process (AC-CL-002's
// launch set derivation).
func (v codexVerb) launches() bool { return v == codexVerbLaunchCli || v == codexVerbLaunchApp }

// codexLaunchRequest is the immutable launch intent both paths consume: the
// resolved codex binary, the argv tail handed to it, and the working
// directory (the PROJECT ROOT, not the process cwd — AC-CL-002's cwd axis).
type codexLaunchRequest struct {
	Program string
	Args    []string // argv AFTER the program token: [verb, passthrough...]
	Dir     string
}

// codexDirectLaunchFn is the direct-launch seam: it receives the fully
// assembled *exec.Cmd (Stdin/Stdout/Stderr already assigned to the parent's
// own values — AC-CL-002's stdio identity) and runs it. The capture harness
// records the cmd's seven fields here.
var codexDirectLaunchFn = func(cmd *exec.Cmd) error { return cmd.Run() }

// codexSpawnLaunchFn is the spawn seam: it receives the NEW-WINDOW TARGET —
// (dir, program, args) of the codex child itself, NOT a tmux invocation
// (AC-CL-003: the capture compares these tokens with the direct path's).
// The default implementation opens the tmux window via the shared tmuxSpawnFn
// (spawn.go owns the only exec.Command("tmux") primitive — AC-CL-016's
// closed set of executables this SPEC's files launch).
var codexSpawnLaunchFn = defaultCodexSpawnLaunch

// defaultCodexSpawnLaunch opens a detached tmux window running codex
// directly. The command string is shell-quoted token-by-token so a tail
// containing spaces, quotes, or $ survives the round trip.
func defaultCodexSpawnLaunch(dir, program string, args []string) error {
	command := buildCodexSpawnCommand(program, args)
	paneID, err := tmuxSpawnFn(dir, command)
	if err != nil {
		return fmt.Errorf("spawn tmux window: %w", err)
	}
	_, _ = fmt.Fprintf(os.Stdout, "Spawned pane %s running `%s` in %s\n", paneID, command, dir)
	_, _ = fmt.Fprintln(os.Stdout, "Switch to it with: tmux select-window -t "+paneID)
	return nil
}

// buildCodexSpawnCommand renders the shell command string for the new tmux
// window: the codex binary followed by its argv tail, every token quoted.
func buildCodexSpawnCommand(program string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, shellQuote(program))
	for _, a := range args {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

// splitCodexDashDash splits head (verb position, before --) from tail
// (verbatim passthrough, from -- on). Tokens after -- are never inspected.
func splitCodexDashDash(args []string) (head, tail []string, hasTail bool) {
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:], true
		}
	}
	return args, nil, false
}

// codexCmd — the launcher-family sibling of cc/glm/cg (same group, same
// DisableFlagParsing discipline: --spawn and -- passthrough are handled by
// the launcher itself). SilenceErrors/SilenceUsage keep every diagnostic
// byte-identical to the named constants above (cobra would otherwise prefix
// "Error: " and append the usage block, breaking the exact-match cells).
var codexCmd = &cobra.Command{
	Use:   "codex [status | cli [codex-args...] | app]",
	Short: "Codex launcher: readiness readout and explicit cli/app launch",
	Long: "Readiness readout and explicit launch for Codex.\n" +
		"\n" +
		"The readout reports six rows and never launches anything on its own:\n" +
		"the codex binary, CODEX_HOME, the auth provider, the project wiring,\n" +
		"the generated agent TOMLs, and the harness entry. Launching requires\n" +
		"an explicit verb. An incomplete wiring row is informational, not an\n" +
		"error: moai init --agent codex generates the .codex wiring files.\n" +
		"\n" +
		"  moai codex [status]   print the readiness readout (no launch)\n" +
		"  moai codex cli        launch the Codex CLI at the project root\n" +
		"  moai codex app        launch the Codex desktop app (codex app)\n" +
		"  --spawn               open the launch in a new tmux window (cli/app only)\n" +
		"  -- <codex-args...>    arguments after -- pass to codex verbatim",
	Example: "  # Show what is installed, resolved, and wired (no launch)\n" +
		"  moai codex\n" +
		"\n" +
		"  # Launch the Codex CLI here, passing arguments through verbatim\n" +
		"  moai codex cli -- --model o3\n" +
		"\n" +
		"  # Launch the desktop app in a new tmux window\n" +
		"  moai codex app --spawn",
	GroupID:            "launch",
	DisableFlagParsing: true,
	SilenceErrors:      true,
	SilenceUsage:       true,
	RunE:               runCodex,
}

func init() {
	rootCmd.AddCommand(codexCmd)
	// Registered for --help visibility and the neutrality flag-usage scan
	// (AC-CL-013); DisableFlagParsing means the launcher itself strips the
	// flag via stripSpawnFlag before the verb lookup.
	codexCmd.Flags().Bool("spawn", false, "open the launch in a new tmux window (cli/app verbs only)")
}

// runCodex routes the invocation: --help first (DisableFlagParsing means
// cobra never intercepts it), then --spawn stripping, then the closed-set
// verb lookup, then the readout or launch path.
func runCodex(cmd *cobra.Command, args []string) error {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return cmd.Help()
		}
		if a == "--" {
			break
		}
	}

	args, spawn := stripSpawnFlag(args)
	head, tail, hasTail := splitCodexDashDash(args)
	if len(head) > 1 {
		return codexUsageFailure(cmd)
	}
	verb := ""
	if len(head) == 1 {
		verb = head[0]
	}
	kind, ok := codexVerbRouting[verb]
	if !ok {
		return codexUsageFailure(cmd)
	}

	if kind.launches() {
		return runCodexLaunch(cmd, verb, tail, spawn)
	}
	if spawn {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexSpawnReadoutDiag)
		return &exitCodeError{code: 1}
	}
	if hasTail {
		return codexUsageFailure(cmd)
	}
	return runCodexReadout(cmd)
}

// codexUsageFailure prints the usage constant to stderr and fails with rc 1.
func codexUsageFailure(cmd *cobra.Command) error {
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexUsageDiag)
	return &exitCodeError{code: 1}
}

// runCodexReadout renders M2's six rows VERBATIM to stdout and succeeds —
// whatever the probes found (AC-CL-004/006: informational, rc 0, stdout only).
func runCodexReadout(cmd *cobra.Command) error {
	r := probeCodexReadiness(cmd.Context())
	for _, row := range r.rows() {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), row)
	}
	return nil
}

// runCodexLaunch resolves the binary and hands off to the direct or spawn
// path. A missing binary is a single-line install hint, launch count 0
// (AC-CL-011).
func runCodexLaunch(cmd *cobra.Command, verb string, tail []string, spawn bool) error {
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexInstallHint)
		return &exitCodeError{code: 1}
	}

	// The launch cwd is the PROJECT ROOT, not the process cwd (AC-CL-002):
	// a call from a subdirectory still launches at the root. An unresolvable
	// root degrades to the process cwd rather than refusing to launch.
	dir := ""
	if root, rerr := findProjectRootFn(); rerr == nil && root != "" {
		dir = root
	} else if cwd, gerr := os.Getwd(); gerr == nil {
		dir = cwd
	}

	req := codexLaunchRequest{Program: binaryPath, Args: append([]string{verb}, tail...), Dir: dir}
	// SPEC-CODEX-INIT-001: the init-offer gate — the ONE call site both
	// launch verbs pass through right before launching. The gate takes no
	// spawn argument: both launch paths cross the same function (REQ-CI-002).
	if err := codexInitOfferGate(cmd, dir); err != nil {
		return err
	}
	if spawn {
		return codexSpawnLaunch(req)
	}
	return codexDirectLaunch(req)
}

// codexDirectLaunch assembles the child (stdio = the parent's OWN os.Stdin /
// os.Stdout / os.Stderr values — the interactive-tty precondition,
// AC-CL-002), runs it through the seam, and propagates the child's exit code
// verbatim (AC-CL-002 rc axis, AC-CL-016).
func codexDirectLaunch(req codexLaunchRequest) error {
	c := exec.Command(req.Program, req.Args...)
	c.Dir = req.Dir
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := codexDirectLaunchFn(c)
	if err == nil {
		return nil
	}
	return codexPropagateLaunchError(err)
}

// codexPropagateLaunchError maps a launch failure onto the exit-code
// discipline: a genuine subprocess exit becomes a DELIBERATE exitCodeError
// carrying the child's code (this is the launcher's contract — unlike every
// other wrap site, the child's code IS the answer), an ExitCoder the seam
// handed us propagates as-is, and anything else (start failure) is described
// via StatusDetail — never a raw %w chain (ResolveExitCode refuses those by
// design; the conversion here is what makes propagation deliberate).
func codexPropagateLaunchError(err error) error {
	var raw *exec.ExitError
	if errors.As(err, &raw) {
		return &exitCodeError{code: raw.ExitCode()}
	}
	if code, ok := ResolveExitCode(err); ok {
		return &exitCodeError{code: code}
	}
	return fmt.Errorf("codex: %s", execerr.StatusDetail(err))
}

// codexSpawnLaunch checks the shared --spawn preconditions (the SAME
// diagnostics moai cc --spawn emits, byte-identical — AC-CL-003) and opens
// the new window through the seam.
func codexSpawnLaunch(req codexLaunchRequest) error {
	if err := checkSpawnPrereqs(); err != nil {
		return err
	}
	if err := codexSpawnLaunchFn(req.Dir, req.Program, req.Args); err != nil {
		return fmt.Errorf("spawn tmux window: %w", err)
	}
	return nil
}
