package cli

// codex_launcher.go — SPEC-CODEX-LAUNCHER-001 M3+M4
// (REQ-CL-001/002/003/012, REQ-CL-014 help copy):
// the `moai codex` command surface. Cobra registration sits in the launch
// group next to cc/glm/cg; the verb routing accepts exactly {bare, cli, app}
// for the launch forms and {status} for the readout (closed sets — unknown
// tokens are rejected, never routed to a launch); --spawn moves the launch to
// a new tmux window.
//
// Two tables, not one. ROUTING answers "what did the operator ask for" and
// stays closed. ARGV TRANSLATION, downstream of it, answers "what does codex
// accept": a routed verb reaches the child ONLY where it names a real codex
// subcommand, so a verb moai synthesized never lands in the child's argv.
//
// -w/--worktree is consumed HERE and never forwarded: it points the child's
// working directory at an EXISTING worktree and never creates one. CODEX_HOME
// reaches the child as an explicit environment entry rather than by ambient
// inheritance, on both the direct and the new-window path.
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
	"path/filepath"
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
	codexUsageDiag = "unknown verb - usage: moai codex [cli] [-w <worktree>] [-- codex-args...] | moai codex status | moai codex app"

	// codexWorktreeValueDiag rejects a -w with no value. moai cc lets a bare
	// -w mean "auto-generate a name" because claude CREATES the worktree;
	// this launcher only RESOLVES one, so there is nothing to resolve.
	codexWorktreeValueDiag = "-w requires a worktree name or path - moai codex resolves an existing worktree and never creates one"

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
	codexVerbReadout   codexVerb = iota // "status"
	codexVerbLaunchCli                  // "" (bare) and "cli"
	codexVerbLaunchApp
)

// codexVerbRouting is the routing table the tests read symbolically: the
// launch-token set derived from it must equal {"", cli, app} and the
// readout-token set must equal {status}. Absence from the table IS the
// rejection — no default branch exists.
//
// The bare token launches. The risk that argued for the opposite default —
// an accidental invocation carrying the session away — does not apply to this
// launcher: the launch is an os/exec CHILD whose exit code propagates, so the
// shell is still there when codex exits.
var codexVerbRouting = map[string]codexVerb{
	"":       codexVerbLaunchCli,
	"cli":    codexVerbLaunchCli,
	"status": codexVerbReadout,
	"app":    codexVerbLaunchApp,
}

// codexChildSubcommand is the ARGV TRANSLATION table — strictly downstream of
// codexVerbRouting and a strict SUBSET of it. A class present here forwards
// its subcommand token to the child; a class absent here forwards nothing but
// the operator's own tail.
//
// app is the only entry, because app is the only routed verb naming a real
// codex subcommand. codex's usage is `codex [OPTIONS] [PROMPT]` with no cli
// subcommand, so forwarding a synthesized cli token would hand codex a PROMPT
// reading "cli"; the bare form's token is the empty string, which is worse.
var codexChildSubcommand = map[codexVerb]string{
	codexVerbLaunchApp: "app",
}

// codexChildArgs assembles the child's argv tail: the translated subcommand
// token where one exists, then the operator's tail verbatim.
func codexChildArgs(kind codexVerb, tail []string) []string {
	args := make([]string, 0, len(tail)+1)
	if sub, ok := codexChildSubcommand[kind]; ok {
		args = append(args, sub)
	}
	return append(args, tail...)
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
// window: the resolved CODEX_HOME as a command-scoped assignment, then the
// codex binary and its argv tail, every token quoted.
//
// The assignment is what makes the two launch paths agree. A tmux window
// inherits the tmux SERVER's environment, not this process's, so without it
// the new window could resolve a different CODEX_HOME than the direct path
// put on its child.
func buildCodexSpawnCommand(program string, args []string) string {
	parts := make([]string, 0, len(args)+2)
	// resolveCodexHomeDir's second result is the source label, not an error.
	if home, _ := resolveCodexHomeDir(); home != "" {
		parts = append(parts, codexHomeEnvVar+"="+shellQuote(home))
	}
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

// codexWorktreeArg is the -w/--worktree value as the operator supplied it,
// already removed from the token stream. present distinguishes "no flag" from
// "flag with an empty value" — the second is an error, the first is not.
type codexWorktreeArg struct {
	present bool
	value   string
}

// stripCodexWorktreeFlag removes -w/--worktree (and their =value forms) from
// the verb-position tokens and returns the value. Only the head is scanned:
// tokens after -- belong to codex and are never inspected. The four accepted
// token shapes mirror normalizeWorktreeFlag's, so the two surfaces accept the
// same spellings.
func stripCodexWorktreeFlag(head []string) ([]string, codexWorktreeArg) {
	rest := make([]string, 0, len(head))
	arg := codexWorktreeArg{}
	for i := 0; i < len(head); i++ {
		token := head[i]
		switch {
		case token == "-w" || token == "--worktree":
			arg.present = true
			// The value is the next token unless that token is itself a flag
			// (a bare -w, which this launcher rejects — see the diagnostic).
			if i+1 < len(head) && !strings.HasPrefix(head[i+1], "-") {
				arg.value = head[i+1]
				i++
			}
		case strings.HasPrefix(token, "--worktree="):
			arg.present = true
			arg.value = strings.TrimPrefix(token, "--worktree=")
		case strings.HasPrefix(token, "-w="):
			arg.present = true
			arg.value = strings.TrimPrefix(token, "-w=")
		default:
			rest = append(rest, token)
		}
	}
	return rest, arg
}

// resolveCodexWorktreeDir turns the operator's -w value into the directory the
// child starts in. It RESOLVES and never CREATES: a name that does not name an
// existing directory is a diagnostic, and the directory is still absent
// afterwards.
//
// The asymmetry with `moai cc` is deliberate. There the flag is FORWARDED to
// claude, which creates the worktree and enters it; codex's top-level help
// exposes no worktree flag, so a forwarded -w would be an unknown token — the
// same failure shape a forwarded synthesized verb has. Consuming the flag
// means moai would have to implement worktree CREATION to match cc, which
// belongs to the worktree tooling that already owns it.
//
// Absolute values are validated by the SAME rule cc applies
// (resolveWorktreeL2Path), so an out-of-prefix path fails with cc's own
// diagnostic rather than a second, divergent one.
func resolveCodexWorktreeDir(projectRoot, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s", codexWorktreeValueDiag)
	}

	path := value
	if filepath.IsAbs(path) {
		if err := resolveWorktreeL2Path([]string{"--worktree", value}); err != nil {
			return "", err
		}
	} else {
		if projectRoot == "" {
			return "", fmt.Errorf("cannot resolve worktree %q: the project root is unresolved", value)
		}
		path = filepath.Join(projectRoot, ".claude", "worktrees", value)
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf(
			"worktree %q does not exist at %s\n"+
				"  moai codex resolves an existing worktree and never creates one\n"+
				"  create it first, then re-run this command",
			value, path,
		)
	}
	return path, nil
}

// codexChildEnv is the environment handed to the codex child: the parent's own
// environment with the RESOLVED CODEX_HOME appended. Appending (rather than
// replacing) keeps the rest of the parent environment intact, and last-wins
// makes the appended entry the value the child reads.
//
// Explicit rather than ambient: where the parent has no CODEX_HOME at all,
// there is nothing to inherit, and the child would otherwise resolve its own
// default independently of the value the readout reports.
func codexChildEnv() []string {
	env := os.Environ()
	// resolveCodexHomeDir's second result is the source label, not an error.
	if home, _ := resolveCodexHomeDir(); home != "" {
		env = append(env, codexHomeEnvVar+"="+home)
	}
	return env
}

// codexCmd — the launcher-family sibling of cc/glm/cg (same group, same
// DisableFlagParsing discipline: --spawn and -- passthrough are handled by
// the launcher itself). SilenceErrors/SilenceUsage keep every diagnostic
// byte-identical to the named constants above (cobra would otherwise prefix
// "Error: " and append the usage block, breaking the exact-match cells).
var codexCmd = &cobra.Command{
	Use:   "codex [cli | status | app]",
	Short: "Codex launcher: launch the Codex CLI, or print the readiness readout",
	Long: "Launch Codex, or ask what is installed and wired.\n" +
		"\n" +
		"Called with no verb, this launches the Codex CLI at the project root.\n" +
		"The readiness readout moved to an explicit alias: moai codex status\n" +
		"reports six rows and starts nothing - the codex binary, CODEX_HOME,\n" +
		"the auth provider, the project wiring, the generated agent TOMLs, and\n" +
		"the harness entry. An incomplete wiring row is informational, not an\n" +
		"error: moai init --agent codex generates the .codex wiring files.\n" +
		"\n" +
		"  moai codex            launch the Codex CLI at the project root\n" +
		"  moai codex cli        the same launch, named explicitly\n" +
		"  moai codex status     print the readiness readout (starts nothing)\n" +
		"  moai codex app        launch the Codex desktop app (codex app)\n" +
		"  -w <worktree>         launch in an EXISTING worktree instead of the\n" +
		"                        project root; this never creates one\n" +
		"  --spawn               open the launch in a new tmux window\n" +
		"  -- <codex-args...>    arguments after -- pass to codex verbatim",
	Example: "  # Launch the Codex CLI here\n" +
		"  moai codex\n" +
		"\n" +
		"  # Show what is installed, resolved, and wired (starts nothing)\n" +
		"  moai codex status\n" +
		"\n" +
		"  # Launch, passing arguments through verbatim\n" +
		"  moai codex -- --model o3\n" +
		"\n" +
		"  # Launch inside an existing worktree\n" +
		"  moai codex -w feat-login\n" +
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
	// -w is consumed before the verb lookup so its tokens can never be
	// mistaken for a verb, and so the verb position keeps its one-token shape.
	head, worktree := stripCodexWorktreeFlag(head)
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
		return runCodexLaunch(cmd, kind, tail, spawn, worktree)
	}
	if worktree.present {
		// A readout starts no process, so it has no working directory to
		// point anywhere.
		return codexUsageFailure(cmd)
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
func runCodexLaunch(cmd *cobra.Command, kind codexVerb, tail []string, spawn bool, worktree codexWorktreeArg) error {
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexInstallHint)
		return &exitCodeError{code: 1}
	}

	// The launch cwd is the PROJECT ROOT, not the process cwd (AC-CL-002):
	// a call from a subdirectory still launches at the root. An unresolvable
	// root degrades to the process cwd rather than refusing to launch.
	projectRoot := ""
	if root, rerr := findProjectRootFn(); rerr == nil && root != "" {
		projectRoot = root
	} else if cwd, gerr := os.Getwd(); gerr == nil {
		projectRoot = cwd
	}

	// -w moves the CHILD's working directory only. The gate below keeps
	// reading the project root: the wiring it classifies is a property of the
	// project, and a linked worktree need not carry a copy of it.
	dir := projectRoot
	if worktree.present {
		resolved, werr := resolveCodexWorktreeDir(projectRoot, worktree.value)
		if werr != nil {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), werr.Error())
			return &exitCodeError{code: 1}
		}
		dir = resolved
	}

	req := codexLaunchRequest{Program: binaryPath, Args: codexChildArgs(kind, tail), Dir: dir}
	// SPEC-CODEX-INIT-001: the init-offer gate — the ONE call site every
	// launch form passes through right before launching, the bare form
	// included. The gate takes no spawn argument: both launch paths cross the
	// same function (REQ-CI-002).
	if err := codexInitOfferGate(cmd, projectRoot); err != nil {
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
	c.Env = codexChildEnv()
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
