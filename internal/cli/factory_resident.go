package cli

// factory_resident.go — `moai factory resident`: runs the Agentic Factory resident monitor loop
// (Issue #1948).
//
// What it does: reads the project's agentic-factory.yaml, opens the claim store the file names,
// wires internal/agenticfactory/resident, and runs its loop until the process is interrupted. With
// webhook routes configured it also serves them over HTTP on --listen.
//
// What it deliberately does NOT do, and says so at start:
//
//   - It enables no write path. The gh runner it injects refuses everything outside gh's read
//     allowlist and AllowWrite stays false, so no Issue is created. Comment writes (the context
//     ledger) are not wired: they need the author policy's identity and that path is undecided.
//   - It never calls git on its default path. The read-only git runner is built (same bounds as
//     gh, plus GIT_ALLOW_PROTOCOL so a remote URL cannot select a dangerous transport) and is
//     kept for the first caller that needs `ls-remote`; the loop does not call it.
//   - It decides nothing that OD-17 leaves open. The tick period has no default (--tick is
//     required unless --once), and the webhook signing secret has no source: while routes are
//     configured and no secret source exists, it refuses to start and says why.
//   - It does not create the claim store. A missing directory is an error, as in the Web Console.
//
// Running it against a real GitHub repository is an operation the user approves; nothing here
// makes that decision.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ghexec"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/gitexec"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/notify/ghpr"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/resident"
	"github.com/modu-ai/moai-adk/internal/config"
)

// The bounds on the external command runners. The leader decided the values (OD-20); whether they
// become settings keys is still open, so they are named constants here, not literals at the call.
const (
	residentExecMaxOutput = 8 << 20 // bytes, standard output and standard error each
	residentExecWaitDelay = time.Second
	// residentShutdownGrace is how long the HTTP server may take to stop after the loop is cancelled.
	residentShutdownGrace = 2 * time.Second
	// residentGitAllowProtocol restricts the transports git may use, so a remote URL such as ext::
	// cannot run a program (a name check cannot catch that; the caller-side limit does).
	residentGitAllowProtocol = "GIT_ALLOW_PROTOCOL=https:ssh"
)

// residentExecs are the two command runners the resident gets, built once from the project root.
type residentExecs struct {
	GH  ghexec.Exec
	Git gitexec.Exec
}

func newResidentExecs(projectRoot, repo string) residentExecs {
	return residentExecs{
		// Repo pins the one write gh can ever run (`issue create`, refused anyway while AllowWrite is false)
		// to the configured repository; the reads do not depend on it.
		GH: ghexec.Exec{Dir: projectRoot, Repo: repo, MaxOutput: residentExecMaxOutput, WaitDelay: residentExecWaitDelay},
		Git: gitexec.Exec{
			Dir: projectRoot, MaxOutput: residentExecMaxOutput, WaitDelay: residentExecWaitDelay,
			Env: []string{residentGitAllowProtocol},
		},
	}
}

// residentOptions is everything runFactoryResident needs; the command fills it from flags and the
// tests from fakes.
type residentOptions struct {
	ProjectRoot string
	Tick        time.Duration // required unless Once: its source is undecided (OD-17)
	Listen      string        // "host:port"; required exactly when webhook routes are configured
	Once        bool          // one pass, then exit; no HTTP server
	// Secret returns the signing secret of one route. Nil in the command: no source is decided.
	Secret func(repo string) []byte
	// GHRead overrides the gh runner (tests); nil means the read-only ghexec runner.
	GHRead ghpr.Runner
	// Transport overrides the gh-backed host lookup transport (tests); nil means ghexec.Transport.
	Transport forge.Transport
	Now       func() time.Time
	Out       io.Writer
	// Ready is called with the address the HTTP server listens on (tests that pass ":0").
	Ready func(addr string)
}

func newFactoryResidentCommand() *cobra.Command {
	var o residentOptions
	cmd := &cobra.Command{
		Use:   "resident",
		Short: "Run the Agentic Factory resident monitor loop (read-only; no writes are wired)",
		Long: "Run the Agentic Factory resident monitor loop for this project.\n\n" +
			"Reads agentic-factory.yaml and the claim store it names, runs one monitor pass at once and\n" +
			"then one per --tick, and serves the configured webhook routes on --listen. It enables no\n" +
			"write path and never creates the claim store. --tick has no default and is required (the\n" +
			"loop period is an open decision); --once runs a single pass and exits.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			o.Out = cmd.OutOrStdout()
			if o.ProjectRoot == "" {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				o.ProjectRoot = wd
			}
			ctx, stop := codexSignalContext(cmd.Context())
			defer stop()
			return runFactoryResident(ctx, o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectRoot, "project-root", "", "project directory (default: the current directory)")
	cmd.Flags().DurationVar(&o.Tick, "tick", 0, "time between monitor passes (required unless --once; no default)")
	cmd.Flags().StringVar(&o.Listen, "listen", "", "host:port to serve webhook routes on (required when routes are configured)")
	cmd.Flags().BoolVar(&o.Once, "once", false, "run one pass and exit")
	return cmd
}

// runFactoryResident validates, wires and runs. It returns nil when the project has not set Agentic
// Factory up (a notice is printed) and when ctx ends the loop; every refusal is an error that says why.
func runFactoryResident(ctx context.Context, o residentOptions) error {
	out := o.Out
	if out == nil {
		out = io.Discard
	}
	res := config.AgenticFactoryConfigured(o.ProjectRoot)
	if !res.Configured {
		if res.Err != nil {
			return fmt.Errorf("agentic factory settings are broken: %s", res.Reason)
		}
		fmt.Fprintf(out, "Agentic Factory is not configured for this project: %s\n", res.Reason)
		return nil
	}
	s, err := config.LoadAgenticFactorySettings(o.ProjectRoot)
	if err != nil {
		return err
	}
	if !o.Once && o.Tick <= 0 {
		return errors.New("--tick is required (or --once): the loop period has no default because its source is an open decision (OD-17)")
	}
	routes := len(s.WebhookRoutes)
	switch {
	case routes > 0 && !o.Once && o.Listen == "":
		return fmt.Errorf("%d webhook route(s) are configured: --listen host:port is required to serve them", routes)
	case routes > 0 && !o.Once && o.Secret == nil:
		return errors.New("webhook routes are configured but no signing-secret source exists (OD-17 is open): refusing to start an unsigned channel")
	case routes == 0 && o.Listen != "":
		return errors.New("--listen was given but the settings name no webhook route")
	}

	dir := s.ClaimStoreDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(o.ProjectRoot, dir)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("claim store directory %q (workflow.agentic_factory.claim_store.dir) is not an existing directory; the resident does not create it", dir)
	}
	st, err := claim.Open(dir)
	if err != nil {
		return fmt.Errorf("claim store directory %q cannot be opened: %w", dir, err)
	}
	execs := newResidentExecs(o.ProjectRoot, s.PRSourceRepository)
	ghRead := o.GHRead
	if ghRead == nil {
		ghRead = execs.GH.Run
	}
	// The Runner's host Facts read through gh (GET only). Nothing in the loop calls them yet: they are used
	// by the authority checks of a claim, which has no production caller; wiring them does not add a network
	// call to a pass.
	transport := o.Transport
	if transport == nil {
		transport = ghexec.Transport{Exec: execs.GH}
	}
	now := o.Now
	if now == nil {
		now = time.Now
	}
	secret := o.Secret
	if routes > 0 && o.Once {
		secret = func(string) []byte { return nil } // Build needs a source; --once serves no route
	}
	r, err := resident.Build(s, resident.Deps{Store: st, GHRead: ghRead, Now: now, Secret: secret, Transport: transport})
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "agentic factory resident: claim store %s, %d webhook route(s)\n", dir, routes)
	fmt.Fprintln(out, "  local writes: claim records (waits, repair marks) and inbox/cursor documents in the claim store")
	fmt.Fprintln(out, "  external writes: none wired (Issue creation off; context-ledger comment writes not wired, path undecided)")
	fmt.Fprintln(out, "  git: not called on the default path")
	fmt.Fprintln(out, "  running against a real repository is an operation the user approves")

	ctx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()
	var srv *http.Server
	serveDone := make(chan error, 1)
	if routes > 0 && !o.Once {
		ln, err := net.Listen("tcp", o.Listen)
		if err != nil {
			return fmt.Errorf("--listen %q: %w", o.Listen, err)
		}
		mux := http.NewServeMux()
		mux.Handle("/hook/", r.Webhook)
		srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
		go func() {
			err := srv.Serve(ln)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				// A server that stopped on its own must not leave the loop running unobserved.
				fmt.Fprintf(out, "webhook server stopped: %v\n", err)
				cancelRun()
			}
			serveDone <- err
		}()
		fmt.Fprintf(out, "  webhook: listening on %s\n", ln.Addr())
		if o.Ready != nil {
			o.Ready(ln.Addr().String())
		}
	}

	var ticks <-chan time.Time
	if o.Once {
		closed := make(chan time.Time)
		close(closed)
		ticks = closed
	} else {
		t := time.NewTicker(o.Tick)
		defer t.Stop()
		ticks = t.C
	}
	runErr := r.Run(ctx, ticks, func(p resident.Pass) { writeResidentPass(out, p) })

	if srv != nil {
		sctx, cancel := context.WithTimeout(context.Background(), residentShutdownGrace)
		defer cancel()
		if err := srv.Shutdown(sctx); err != nil {
			_ = srv.Close()
		}
		if err := <-serveDone; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	if errors.Is(runErr, context.Canceled) {
		return nil
	}
	return runErr
}

// writeResidentPass prints one pass: the action count by kind, and every action that carries an error
// with its Issue. A silent failure is the defect this exists for (audit F3): a stuck record must show
// on screen, not only in the claim store.
func writeResidentPass(out io.Writer, p resident.Pass) {
	at := p.At.UTC().Format(time.RFC3339)
	if p.Err != nil {
		fmt.Fprintf(out, "pass %s: error: %v\n", at, p.Err)
		return
	}
	counts := map[string]int{}
	for _, a := range p.Actions {
		counts[a.What]++
	}
	kinds := make([]string, 0, len(counts))
	for k := range counts {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	parts := make([]string, 0, len(kinds))
	for _, k := range kinds {
		parts = append(parts, fmt.Sprintf("%s=%d", k, counts[k]))
	}
	fmt.Fprintf(out, "pass %s: %d action(s) %s\n", at, len(p.Actions), strings.Join(parts, " "))
	for _, a := range p.Actions {
		if a.Err != nil {
			fmt.Fprintf(out, "  %s issue #%d: error: %v\n", a.What, a.Issue, a.Err)
		}
	}
}
