package cli

// verify_run.go — `moai verify run`: run a command, or reuse a passing result
// recorded for the same working-tree state (SPEC-VERIFY-RUN-REUSE-001). It is
// one verb over the existing snapshot store: key, receipt comparison and
// RecordCheck are internal/verify's; this file only orchestrates them.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/verify"
)

const (
	// verifyRunDefaultTimeout bounds the executed command (REQ-VRR-003); it
	// covers a 30-minute suite without being an unbounded wait.
	verifyRunDefaultTimeout = 60 * time.Minute
	// verifyRunToolVersionTimeout bounds the tool-identity command (REQ-VRR-006).
	verifyRunToolVersionTimeout = 30 * time.Second
	// verifyRunUnversioned is the tool identity when --tool-version-cmd is absent.
	verifyRunUnversioned = "unversioned"
	// verifyRunWaitDelay bounds how long Wait lingers on output pipes a
	// surviving descendant still holds after the command is killed.
	verifyRunWaitDelay = 2 * time.Second

	exitUsage    = 2
	exitTimeout  = 124
	exitNotFound = 127
)

func init() {
	verifyExtraCommands = append(verifyExtraCommands, newVerifyRunCmd)
}

func newVerifyRunCmd(projectRoot *string) *cobra.Command {
	var (
		checkID        string
		ttl            time.Duration
		envNames       []string
		toolVersionCmd []string
		toolVersionTO  time.Duration
		timeout        time.Duration
	)
	cmd := &cobra.Command{
		Use:   "run [flags] -- <command...>",
		Short: "Run a command, or reuse a passing result recorded for the same tree state",
		Long: `Run <command...> in the project root, unless a passing result for the same
command is already recorded under the current working-tree key — then print a
reuse notice on stderr and exit 0 without running it.

A recording is reused only when ALL hold: the tree key is unchanged, the
command matches byte for byte (argv elements joined by a space; an element with
whitespace or a quote is quoted), the recorded exit code is 0, the bound
environment digest and tool identity are unchanged, and the recording is within
--ttl (default 10m). Otherwise the command runs with inherited stdio, bounded by
--timeout (default 60m), and exits with its own exit code (124 on timeout, 127
if it cannot start). A run during which the tree key did not move is recorded.

The command is exec'd directly, with no shell: pass shell syntax as one
element, e.g. -- sh -c 'go vet ./... && go test ./pkg'.

The key does NOT cover what you do not bind. List the environment variables the
result depends on with --env NAME,... (unset and empty differ), and the toolchain
with a repeatable --tool-version-cmd, e.g. --tool-version-cmd go --tool-version-cmd
version; without it the identity is "unversioned". If the tool-version command
fails, times out or prints nothing, nothing is reused or recorded.

A reuse is a prior observation, not one made in this run: cite its key and
recorded_at and list "output not re-observed" under Gaps. To bypass: --ttl 1ns.`,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if c.ArgsLenAtDash() != 0 || len(args) == 0 {
				return &exitCodeError{code: exitUsage, msg: "verify run: usage: moai verify run [flags] -- <command...>"}
			}
			root, err := verifyResolveRoot(*projectRoot)
			if err != nil {
				return &exitCodeError{code: exitUsage, msg: err.Error()}
			}
			return verifyRun(c, verifyRunParams{
				root: root, argv: args, checkID: checkID, ttl: ttl, envNames: envNames,
				toolVersionCmd: toolVersionCmd, toolVersionTimeout: toolVersionTO, timeout: timeout,
			})
		},
	}
	cmd.Flags().StringVar(&checkID, "check-id", "run", "check identifier recorded with the result (filter for `verify check --check`)")
	cmd.Flags().DurationVar(&ttl, "ttl", 0, "reuse TTL, measured from the recorded command's completion (default 10m)")
	cmd.Flags().StringSliceVar(&envNames, "env", nil, "environment variable names the result depends on (comma-separated or repeated)")
	cmd.Flags().StringArrayVar(&toolVersionCmd, "tool-version-cmd", nil, "one argv element of the tool-identity command; repeat per element")
	cmd.Flags().DurationVar(&toolVersionTO, "tool-version-timeout", verifyRunToolVersionTimeout, "bound for the tool-identity command")
	cmd.Flags().DurationVar(&timeout, "timeout", verifyRunDefaultTimeout, "bound for the command; on expiry it is terminated and the verb exits 124")
	return cmd
}

type verifyRunParams struct {
	root               string
	argv               []string
	checkID            string
	ttl                time.Duration
	envNames           []string
	toolVersionCmd     []string
	toolVersionTimeout time.Duration
	timeout            time.Duration
}

// verifyRun is the whole verb. Anything that stops reuse or recording from
// being trustworthy (store, key, tool identity) is reported on stderr and
// degrades to a plain run: the failure mode is a re-execution, never a
// fabricated hit (REQ-VRR-007).
func verifyRun(c *cobra.Command, p verifyRunParams) error {
	ctx := verifyReceiptContext(c)
	stderr := c.ErrOrStderr()
	notef := func(format string, a ...any) { _, _ = fmt.Fprintf(stderr, "verify run: "+format+"\n", a...) }

	state := verify.ReceiptState{
		Command:      verify.CanonicalCommand(p.argv),
		ConfigDigest: verify.EnvDigest(p.envNames, os.LookupEnv),
		ToolVersion:  verifyRunUnversioned,
	}
	bound := true
	if len(p.toolVersionCmd) > 0 {
		ident, err := verifyRunToolIdentity(ctx, p.root, p.toolVersionCmd, p.toolVersionTimeout)
		if err != nil {
			notef("tool version unbound (%v); running without reuse or record", err)
			bound = false
		}
		state.ToolVersion = ident
	}
	store, err := verifyStoreRoot(p.root)
	if err != nil {
		notef("snapshot store unavailable (%v); running without reuse or record", err)
		bound = false
	}
	var keyBefore string
	if bound {
		keyBefore, err = verifyRunKey(ctx, p.root)
		if err != nil {
			notef("%v; running without reuse or record", err)
			bound = false
		}
	}
	if bound {
		state.Head, state.TreeDigest, _ = strings.Cut(keyBefore, ":")
		snap, loadErr := verify.Load(store, keyBefore)
		if loadErr != nil {
			notef("snapshot load failed (%v); running without reuse or record", loadErr)
			bound = false
		} else if hit, entry, reason := verify.DecideReuse(snap, state, time.Now(), p.ttl); hit {
			notef("reuse key=%s recorded_at=%s duration_ms=%d", keyBefore, entry.RecordedAt.Format(time.RFC3339), entry.DurationMS)
			return nil
		} else {
			notef("miss (%s)", reason)
		}
	}

	start := time.Now()
	code, completed := verifyRunExec(ctx, c, p)
	end := time.Now()
	if bound && completed {
		verifyRunRecord(ctx, notef, p, state, store, keyBefore, code, start, end)
	}
	if code == 0 {
		return nil
	}
	return &exitCodeError{code: code, msg: fmt.Sprintf("verify run: command exited with code %d", code)}
}

// verifyRunExec runs the command with the caller's stdio in the project root.
// completed is false when the command timed out or could not start: those
// outcomes are never recorded (REQ-VRR-003).
func verifyRunExec(ctx context.Context, c *cobra.Command, p verifyRunParams) (code int, completed bool) {
	runCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, p.argv[0], p.argv[1:]...)
	cmd.Dir = p.root
	cmd.Stdin, cmd.Stdout, cmd.Stderr = c.InOrStdin(), c.OutOrStdout(), c.ErrOrStderr()
	cmd.WaitDelay = verifyRunWaitDelay
	verifyRunPrepare(cmd)
	if err := cmd.Start(); err != nil {
		_, _ = fmt.Fprintf(c.ErrOrStderr(), "verify run: cannot start command: %v\n", err)
		return exitNotFound, false
	}
	err := cmd.Wait()
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		_, _ = fmt.Fprintf(c.ErrOrStderr(), "verify run: command exceeded --timeout %s and was terminated\n", p.timeout)
		return exitTimeout, false
	}
	var exitErr *exec.ExitError
	switch {
	case err == nil, errors.Is(err, exec.ErrWaitDelay):
		return 0, true
	case errors.As(err, &exitErr):
		if code = exitErr.ExitCode(); code > 0 {
			return code, true
		}
		return 1, true // terminated by a signal: no exit code to pass through
	default:
		_, _ = fmt.Fprintf(c.ErrOrStderr(), "verify run: waiting for the command failed: %v\n", err)
		return 1, false
	}
}

// verifyRunRecord records the finished run under the key computed before it
// started, but only if the tree key did not move while it ran: a result that
// straddles a tree change describes no single tree (REQ-VRR-004).
func verifyRunRecord(ctx context.Context, notef func(string, ...any), p verifyRunParams, state verify.ReceiptState, store, keyBefore string, code int, start, end time.Time) {
	keyAfter, err := verifyRunKey(ctx, p.root)
	if err != nil {
		notef("%v; not recording", err)
		return
	}
	if keyAfter != keyBefore {
		notef("working tree changed during the run (key %s -> %s); not recording", keyBefore, keyAfter)
		return
	}
	verdict := "pass"
	if code != 0 {
		verdict = "fail"
	}
	entry := verify.CheckEntry{
		CheckID:      p.checkID,
		Command:      state.Command,
		ExitCode:     code,
		RecordedAt:   end,
		DurationMS:   end.Sub(start).Milliseconds(),
		ConfigDigest: state.ConfigDigest,
		ToolVersion:  state.ToolVersion,
		Verdict:      verdict,
	}
	if _, err := verify.RecordCheck(store, keyBefore, entry); err != nil {
		notef("snapshot record failed (%v); the result is not reusable", err)
	}
}

// verifyRunKey computes the working-tree key under the CLI's key timeout.
func verifyRunKey(ctx context.Context, root string) (string, error) {
	kctx, cancel := context.WithTimeout(ctx, verifyKeyTimeout)
	defer cancel()
	key, err := verify.Key(kctx, root)
	if err != nil {
		return "", fmt.Errorf("key computation failed: %v", err)
	}
	return key, nil
}

// verifyRunToolIdentity runs the tool-identity command once and returns its
// trimmed stdout; stderr is discarded. Every failure leaves the identity
// unbound (REQ-VRR-006).
func verifyRunToolIdentity(ctx context.Context, root string, argv []string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = verifyRunToolVersionTimeout
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(tctx, argv[0], argv[1:]...)
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.WaitDelay = verifyRunWaitDelay
	err := cmd.Run()
	if errors.Is(tctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("timed out after %s", timeout)
	}
	if err != nil {
		return "", err
	}
	ident := strings.TrimSpace(out.String())
	if ident == "" {
		return "", errors.New("the command printed nothing on stdout")
	}
	return ident, nil
}
