package cli

// fetch_web.go — SPEC-FETCH-WEB-001 M3/M4. `moai fetch web <url>
// "<question>"`: the Go port of the jevfetch pipeline (internal/fetchweb).
// The command layer stays thin (plan B3): argument parsing, the
// workflow.jev.enabled gate read, pipeline invocation, output shaping, and
// the exit-code mapping. No scoring or fetching logic lives here.
//
// Off-gate contract (REQ-FW-008, plan B1): while workflow.jev.enabled is
// false (the shipped default) the command makes ZERO network requests —
// neither the page fetch nor any Jev call — because the gate is read BEFORE
// any transport is constructed, prints a WebFetch fallback notice on stderr,
// and exits 2.
//
// User interaction prompts are prohibited here (orchestrator-only HARD,
// C-HRA-008 family).

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/fetchweb"
	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/jevcred"
)

// fetchWebJSON backs the --json flag.
var fetchWebJSON bool

// fetchWebPageClientFn is the page-fetch client seam: the gate is read
// BEFORE this is consulted, so a gated-off run constructs no transport at
// all (the zero-network property is a counted observation in tests, not an
// absence). Replaced in tests.
var fetchWebPageClientFn = func() *http.Client {
	return &http.Client{Timeout: fetchweb.FetchTimeout}
}

// fetchWebAskerFn is the Jev seam: a fully wired production client, or a
// test stub. Called only after the gate has passed. The disk cache
// (REQ-FW-011) rides jev.DefaultCacheDir (~/.moai/cache/jev); an
// unresolvable home leaves it off (fail-open).
var fetchWebAskerFn = func() fetchweb.Answerer {
	c := jev.New(true)
	c.LoadCredential = jevcred.Load
	if dir, err := jev.DefaultCacheDir(); err == nil {
		c.CacheDir = dir
	}
	return c
}

// fetchWebOffGateNotice is the WebFetch fallback notice (REQ-FW-008): the
// capability is off, so spend-adjacent I/O is refused and the operator is
// pointed at the always-available fallback.
const fetchWebOffGateNotice = `moai fetch web: workflow.jev.enabled is false — no page fetch and no Jev call was made.

Use WebFetch (the built-in summarizing fetch) for this page, or opt in to the
verbatim pipeline by setting workflow.jev.enabled: true in
.moai/config/sections/workflow.yaml (a TypeSafe credential at ~/.moai/.env.typesafe
is then also required).`

func init() {
	rootCmd.AddCommand(newFetchCmd())
}

// newFetchCmd builds the `fetch` parent. The naming is operator-fixed
// (spec.md §D5): `moai fetch web`, never a `web` sub-verb (collision with
// the web UI command).
func newFetchCmd() *cobra.Command {
	fetch := &cobra.Command{
		Use:     "fetch",
		Short:   "Fetch-and-extract commands",
		GroupID: "tools",
	}
	web := &cobra.Command{
		Use:   `web <url> "<question>"`,
		Short: "Fetch a page and return verbatim excerpts that answer a question (Jev-scored)",
		Long: `Fetch a web page, split it into blocks (paragraphs, list items, table
rows, code blocks), score the blocks for relevance against the question in
two Jev stages, and print the kept blocks VERBATIM — no paraphrase, no
truncation — each preceded by its heading path.

Output: a first line beginning "#" carrying one of the statuses
filtered / passthrough / no_match / empty / fetch_error, then per kept block
a "##" heading-path line followed by the block's verbatim text.

Exit codes: 0 for filtered and passthrough, 1 for no_match and empty,
2 for fetch_error (and for the gated-off state: while workflow.jev.enabled
is false the command makes no network request at all and points at WebFetch).

Gated behind workflow.jev.enabled (shipped default false); a TypeSafe
credential at ~/.moai/.env.typesafe is required once enabled.

Flags:
  --json   Emit the same information as a machine-readable JSON object`,
		Args:          cobra.ExactArgs(2),
		RunE:          runFetchWeb,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	web.Flags().BoolVar(&fetchWebJSON, "json", false, "emit the result as a machine-readable JSON object")
	fetch.AddCommand(web)
	return fetch
}

// runFetchWeb executes the command. The gate read is FIRST and
// short-circuits (the zero-network off-gate depends on this ordering), then
// the bounded page fetch, then the shared pipeline entry point — the same
// one the MCP fetch_web tool calls (REQ-FW-009, AC-FW-006).
//
// @MX:WARN: [AUTO] gate-read path — the workflow.jev.enabled read must stay first and short-circuiting
// @MX:REASON: [AUTO] a reordered gate would let a spend-adjacent page fetch (or a Jev call) fire from a project that never opted in; the zero-network off-gate contract (REQ-FW-008) is enforced by this ordering, not by the transport.
func runFetchWeb(cmd *cobra.Command, args []string) error {
	url, question := args[0], args[1]

	root, rootErr := findProjectRootFn()
	enabled := false
	if rootErr == nil {
		enabled, _ = jevEnabled(root) // read failure reads as disabled (fail-safe)
	}
	if !enabled {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), fetchWebOffGateNotice)
		return &exitCodeError{code: 2, msg: "moai fetch web: gated off (workflow.jev.enabled is false)"}
	}

	res := RunFetchPipeline(cmd.Context(), url, question, fetchWebPageClientFn(), fetchWebAskerFn())

	var out string
	if fetchWebJSON {
		raw, err := fetchweb.RenderJSON(res, url, question)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai fetch web: rendering failed: %v\n", err)
			return &exitCodeError{code: 2, msg: "moai fetch web: rendering failed"}
		}
		out = string(raw)
	} else {
		out = fetchweb.RenderMarkdown(res)
	}
	_, _ = fmt.Fprint(cmd.OutOrStdout(), out)
	if res.Condition != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai fetch web: %s\n", res.Condition)
	}
	if code := res.Status.ExitCode(); code != 0 {
		return &exitCodeError{code: code, msg: "moai fetch web: " + string(res.Status)}
	}
	return nil
}

// RunFetchPipeline is the ONE shared pipeline entry point the CLI command
// and the MCP fetch_web tool both call (REQ-FW-009, AC-FW-006): fetch the
// page, split it, run the two-stage scoring, return the result. A page-fetch
// failure maps to the fetch_error status with its condition.
//
// @MX:ANCHOR: [AUTO] fetch-web pipeline SSOT — CLI and MCP dispatch share this single entry point
// @MX:REASON: [AUTO] a second pipeline copy in the MCP dispatch would drift from the CLI's status/exit semantics; AC-FW-006 asserts both surfaces call this same function.
func RunFetchPipeline(ctx context.Context, url, question string, pageClient *http.Client, asker fetchweb.Answerer) fetchweb.Result {
	pageHTML, err := fetchweb.FetchPage(ctx, url, pageClient)
	if err != nil {
		return fetchweb.Result{Status: fetchweb.StatusFetchError, Condition: err.Error()}
	}
	blocks := fetchweb.SplitBlocks(pageHTML)
	return fetchweb.Process(ctx, blocks, question, asker)
}
