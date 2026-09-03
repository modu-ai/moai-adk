// todo_pr.go — `moai todo pr [<id>]` (SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-2, M3).
//
// The verb answers, for each card in the queue: is somebody already carrying
// this? It reads, and it writes NOTHING — not a field, not a findings entry,
// not a timestamp, not a cache, not a lock. That is the ruling in spec.md §B,
// and it is a property of this file rather than a convention its callers keep.
//
// Why a DEDICATED VERB rather than a column on `moai todo list`: one
// `gh pr list` costs 0.878s, essentially all round-trip. `todo list` is the
// queue's cheapest read — rendered on every operator glance and by the foreman
// loop — and a default-on column would make every one of those callers pay
// ~0.9s of network to serve the one caller who wanted the link. The cheap path
// stays cheap; the network cost is an explicit operator act.
//
// SUBAGENT BOUNDARY: nothing here prompts.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

// todoPROpenPRLimit bounds the open-PR page the single query fetches. One
// query, one page — a paging loop would spawn one `gh` process per page, which
// is exactly the per-invocation cost the dedicated-verb ruling exists to avoid
// and which the subprocess census forbids.
//
// The limit is therefore a real ceiling rather than a formality: a card whose
// pull request sits beyond it would resolve as `no-link` or `landed`, silently
// and wrongly. `fetchOpenPRs` reports saturation instead of hiding it — see
// there.
const todoPROpenPRLimit = 100

// todoPRSubprocessTimeout bounds every subprocess this verb spawns. `gh` talks
// to the network and `git log` walks history; neither is guaranteed to return,
// and a read-only status verb that hangs is worse than one that degrades — the
// fail-open path already renders a useful queue without either answer.
const todoPRSubprocessTimeout = 30 * time.Second

// todoRunCommand is the process seam every subprocess in the todo surface
// goes through. It exists so the subprocess-census tests can COUNT
// invocations: AC-009 asserts `todo list` spawns zero, and AC-014 asserts
// `todo pr` spawns exactly one `gh` regardless of queue length.
//
// Package-level rather than plumbed, because the assertion is about the whole
// command surface — a second, unrouted exec call is precisely the regression
// the census is there to catch, and it would be invisible to a seam only the
// routed path knows about.
var todoRunCommand kanban.CommandRunner = func(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), todoPRSubprocessTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if ctx.Err() != nil {
		return "", fmt.Errorf("%s timed out after %s: %w", name, todoPRSubprocessTimeout, ctx.Err())
	}
	return string(out), err
}

// newTodoPRCmd — `moai todo pr [<id>]`: report each card's pull-request and
// landed state. Read-only, fail-open, one `gh` query.
func newTodoPRCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "pr [<id>]",
		Short: "Report each card's open pull request or landed state (read-only)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var only string
			if len(args) == 1 {
				only = normalizeTodoRef(args[0])
			}
			return runTodoPR(cmd, only, jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false,
		"Emit the link outcomes as JSON on stdout")
	withResolvedLandedRef(cmd, func(landedRef string) {
		cmd.Long = todoPRLong(landedRef)
	})
	return cmd
}

// todoPRLong renders `todo pr`'s help body against the ref the landing
// question will actually be asked about.
//
// The ref is RESOLVED, not constant: a project that integrates on a branch
// other than the default asks the question about its own branch, and the help
// text names the ref the check will actually use rather than a default that
// may not apply here. It is resolved lazily — see withResolvedLandedRef.
func todoPRLong(landedRef string) string {
	return `Report, for every queued card, whether an open pull request already
delivers it or whether its work has already landed on ` + landedRef + `.

The verb writes NOTHING: no card field, no finding, no cache, no lock. It
computes every outcome live and prints it.

Five outcomes, distinguishable by kind alone:

  linked     one open pull request carries the card id
             confidence exact    — read off the PR title
             confidence inferred — read off a single PR body
  ambiguous  several open PR bodies carry it; every candidate is listed and
             none is chosen
  landed     no open PR carries it, and ` + landedRef + ` history names it.
             It means SOMETHING naming the card landed on that ref — NOT that
             the card's last step landed
  no-link    nobody has started this
  unknown    the landing question could not be asked (no such ref, no git, a
             failed query). This is NOT evidence of not-landed

The landed check is local git and keeps working when gh does not. When gh is
absent, unauthenticated, or offline the link column renders empty, the
degradation is noted on stderr, and the exit code stays 0.`
}

// runTodoPR renders the link view. Every exit path is exit 0 unless the queue
// itself is unreadable — a degraded link lookup is a note, not an error.
func runTodoPR(cmd *cobra.Command, only string, jsonOutput bool) error {
	// REQ-BJD-002 — probed before the read (todo_disclosure.go).
	_ = discloseQueueLayout(cmd, "pr")
	rec, err := newTodoStore().Load()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	prs, saturated, ghErr := fetchOpenPRs()
	if saturated {
		// The page filled exactly. A pull request past the ceiling is invisible
		// to the resolver, and its card would report `no-link` or `landed` —
		// wrong, and silent. Saying so is the whole mitigation: paging would
		// spawn a second `gh` process, which the one-query bound forbids.
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"note: the open pull-request query returned its %d-record ceiling; a card whose pull request sits beyond it is reported as if it had none\n",
			todoPROpenPRLimit)
	}
	if ghErr != nil {
		// Fail-open (REQ-2.3): the note names what degraded, so an empty link
		// column is never mistaken for "no card has a pull request".
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"note: open pull requests unavailable (%v); link column left empty, landed check still ran\n", ghErr)
		prs = nil
	}

	landedRef := todoLandedRef()
	landed := kanban.GitLandedQuerier{Run: todoRunCommand, Ref: landedRef}
	outcomes := make([]kanban.PRLinkOutcome, 0, len(rec.Items))
	var degraded []string
	for _, it := range rec.Items {
		if only != "" && it.ID != only {
			continue
		}
		out, err := kanban.ResolveCardPRLink(it.ID, prs, landed)
		if err != nil {
			degraded = append(degraded, it.ID)
		}
		outcomes = append(outcomes, out)
	}
	if len(degraded) > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"note: the landed check against %s could not answer for %s; those cards report unknown rather than no-link, because an unanswerable query is not evidence of not-landed\n",
			landedRef, strings.Join(degraded, " "))
	}

	out := cmd.OutOrStdout()
	if jsonOutput {
		data, err := json.Marshal(outcomes)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(out, string(data))
		return nil
	}
	if len(outcomes) == 0 {
		_, _ = fmt.Fprintln(out, "queue is empty")
		return nil
	}
	text := map[string]string{}
	state := map[string]kanban.BacklogState{}
	for _, it := range rec.Items {
		text[it.ID] = it.Text
		state[it.ID] = it.State
	}
	for _, o := range outcomes {
		// SIX columns, always present. The link column is BLANK rather than
		// omitted when there is nothing to show, so a degraded run and a
		// genuinely unlinked card render the same shape and differ only in
		// the stderr note (AC-005).
		//
		// The queue STATE sits between Confidence and the free-text tail: a
		// `picked` card with no commits and a `queued`, never-started one both
		// resolve to `no-link`, and without the state they rendered as the
		// same row. It is a column-count change on a machine-readable
		// surface — a consumer doing `cut -f5` now gets the state where it
		// used to get the text — but a consumer reading the LAST field still
		// reads the card text.
		//
		// No new subprocess and no new query: the value is already in hand
		// from the record this render already loaded.
		_, _ = fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\t%s\n",
			o.CardID, o.Kind, formatPRLinks(o.PRs), o.Confidence, state[o.CardID], text[o.CardID])
	}
	return nil
}

// formatPRLinks renders the candidate list, empty string for none.
func formatPRLinks(prs []int) string {
	if len(prs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(prs))
	for _, n := range prs {
		parts = append(parts, fmt.Sprintf("#%d", n))
	}
	return strings.Join(parts, ",")
}

// fetchOpenPRs issues THE single network query — one `gh pr list` for the
// whole invocation, never one per card (NFR-1). Any failure is returned for
// the caller to note; no error path here aborts the render.
//
// The second return value reports SATURATION: the record count came back equal
// to the requested ceiling, so there may be open pull requests the resolver
// never saw. It is reported rather than paged around, because a second page
// means a second `gh` process.
func fetchOpenPRs() (prs []kanban.PRRecord, saturated bool, err error) {
	out, err := todoRunCommand("gh", "pr", "list",
		"--state", "open",
		"--limit", strconv.Itoa(todoPROpenPRLimit),
		"--json", "number,title,body,state")
	if err != nil {
		return nil, false, err
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &prs); err != nil {
		return nil, false, fmt.Errorf("parsing gh pr list output: %w", err)
	}
	return prs, len(prs) >= todoPROpenPRLimit, nil
}
