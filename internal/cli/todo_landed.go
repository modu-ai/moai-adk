// todo_landed.go — `moai todo landed <id> [--sha <sha>] [--ref <ref>] | --clear`
// (SPEC-TODO-LANDING-EVIDENCE-001 REQ-TLE-004, REQ-TLE-007..010, REQ-TLE-012,
// REQ-TLE-020, M3).
//
// The verb records what the operator asserts about a card's landing. It is the
// FIRST writer of the evidence column, and three of its rules exist because
// every plausible slip on this axis is silent:
//
//  1. EVIDENCE NEVER TRANSITIONS A CARD. Recording that a card's work landed
//     is not the same act as closing it, and a verb that helpfully did both
//     would move cards nobody asked it to move. One row's one column; the
//     state, the position, the text, and the spec id are untouched.
//
//  2. A STORED DELIVERING SHA IS OPERATOR-ASSERTED OR ABSENT. The landed grep
//     finds commits that MENTION a card, and a mention is not a delivery
//     (SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-1.10). This file consults that
//     predicate nowhere; the only path a SHA takes into the store is --sha.
//
//  3. A WRITE THAT CANNOT VALIDATE REFUSES. `todo pr` degrades on an
//     unanswerable git because refusing would block every machine that cannot
//     answer a READ. The opposite holds here, deliberately: the alternative to
//     refusing is storing an unvalidated SHA permanently.
//
// SUBAGENT BOUNDARY: nothing here prompts.
package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

// The three ways a --sha may be refused. They are DISTINCT tokens rather than
// one "invalid sha" message because the operator's next act differs for each:
// a failed existence check means the id is wrong, a failed reachability check
// means the commit is real but on the wrong ref, and an unrunnable check means
// nothing was learned about the SHA at all.
const (
	landedCheckExistence    = "existence"
	landedCheckReachability = "reachability"
	landedCheckUnrunnable   = "unrunnable"
)

// newTodoLandedCmd — `moai todo landed <id>`: record, replace, or clear one
// card's landing evidence.
func newTodoLandedCmd() *cobra.Command {
	var sha string
	var ref string
	var clear bool
	landedRef := todoLandedRef()
	cmd := &cobra.Command{
		Use:   "landed <n>",
		Short: "Record the operator's landing evidence for a card (or clear it)",
		Long: `Record what YOU assert about a card's landing, as one locked write.

The record carries six facts: the ref the observation was made against, that
ref's head at the observation instant, the instant itself, the delivering
commit if you name one, its provenance, and the card's SPEC status read at
record time. Absent --ref the ref is ` + landedRef + `, the same one
` + "`moai todo pr`" + ` asks about, so the two surfaces cannot name different refs.

The record is EVIDENCE, not a transition. It moves no card, closes no card,
and reorders nothing: ` + "`moai todo done`" + ` remains the only way a card leaves the
queue.

--sha names the delivering commit ON YOUR AUTHORITY. It is the only path by
which a delivering SHA is ever stored — nothing here derives one from history,
because the landed check finds commits that MENTION a card and a mention is
not a delivery. A supplied SHA is validated before it is stored: it must
resolve to a commit, and that commit must be reachable from the record's ref.
The full resolved SHA is what gets stored, never the abbreviated input.

A check that fails, or that cannot be run at all (no git, a ref that resolves
to nothing), refuses the write and names which check it was. That is the
opposite of ` + "`moai todo pr`" + `, which degrades on the same condition — a read that
cannot answer stays permissive, a write that cannot validate refuses.

--clear removes the record, returning the card to the same state as one that
never carried evidence. It exists because a mistyped --sha would otherwise be
permanent. It takes no --sha and no --ref.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// A usage error is exit 2, distinct from the exit 1 a refused
			// validation returns: one says the invocation was malformed, the
			// other says it was well-formed and the facts refused it.
			if clear && (sha != "" || ref != "") {
				return &exitCodeError{code: 2,
					msg: "todo landed: --clear takes neither --sha nor --ref; clearing removes the record rather than recording a different one"}
			}
			return runTodoLanded(cmd, normalizeTodoRef(args[0]), sha, ref, clear)
		},
	}
	cmd.Flags().StringVar(&sha, "sha", "",
		"The delivering commit, on your authority (validated for existence and reachability)")
	cmd.Flags().StringVar(&ref, "ref", "",
		"Ask about this ref instead of "+landedRef)
	cmd.Flags().BoolVar(&clear, "clear", false,
		"Remove the card's landing record")
	return cmd
}

// runTodoLanded performs the single locked write.
//
// Everything that can refuse runs BEFORE the mutation: the card lookup is the
// only refusal inside it, and Mutate writes nothing when its callback returns
// an error. A validation performed inside the lock would hold the queue for
// the duration of two git subprocesses.
func runTodoLanded(cmd *cobra.Command, id, sha, ref string, clear bool) error {
	store := newTodoStore()

	var record *kanban.LandingEvidence
	if !clear {
		built, err := buildLandingEvidence(store, id, sha, ref)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			return err
		}
		record = built
	}

	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == id {
				// ONE field. The state, the position, the text, and the spec
				// id are not read here and not written; a card is exactly
				// where it was before this line ran.
				rec.Items[i].Landing = record
				return nil
			}
		}
		return fmt.Errorf("no backlog item %s", id)
	}); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		return err
	}

	if clear {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "landed %s cleared\n", id)
		return nil
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "landed %s ref=%s marker=%s\n", id, record.Ref, record.Marker())
	return nil
}

// buildLandingEvidence assembles the record, refusing rather than storing any
// fact it could not establish.
func buildLandingEvidence(store *kanban.BacklogStore, id, sha, ref string) (*kanban.LandingEvidence, error) {
	root := resolveTodoQueueRoot()
	if strings.TrimSpace(ref) == "" {
		ref = todoLandedRef()
	}

	// The ref head is a MANDATORY fact (REQ-TLE-005), so an unanswerable git
	// is already fatal to forming a record — which is why the refusing
	// posture of the --sha checks introduces no new policy. This is also the
	// single place case (d) originates for BOTH of its sub-conditions: git
	// absent, and a ref that resolves to nothing.
	refHead, err := todoGitOutput("rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if err != nil || refHead == "" {
		return nil, &exitCodeError{code: 1, msg: fmt.Sprintf(
			"todo landed: the %s check could not run: ref %q did not resolve to a commit (%v); "+
				"a write that cannot validate refuses rather than storing an unvalidated record",
			landedCheckUnrunnable, ref, gitReason(err))}
	}

	resolvedSHA, err := validateSuppliedSHA(sha, ref)
	if err != nil {
		return nil, err
	}

	evidence := &kanban.LandingEvidence{
		Ref:        ref,
		RefHead:    refHead,
		ObservedAt: time.Now().UTC().Format(time.RFC3339),
		SpecStatus: readCardSpecStatus(store, root, id),
	}
	if resolvedSHA != "" {
		evidence.SHA = resolvedSHA
		// The provenance is set HERE and nowhere else, and its only value is
		// `operator`. The encoder refuses a SHA without it, so a future path
		// that filled SHA from anywhere else would have to state a provenance
		// the SPEC does not define — and be refused for it.
		evidence.SHASource = kanban.LandingSHASourceOperator
	}
	if _, err := kanban.EncodeLandingEvidence(*evidence); err != nil {
		return nil, &exitCodeError{code: 1, msg: "todo landed: " + err.Error()}
	}
	return evidence, nil
}

// validateSuppliedSHA resolves and range-checks an operator-supplied SHA,
// returning the FULL resolved form.
//
// Neither command is given the card id. That is the structural reason this is
// a referential-integrity check and not the attribution REQ-1.10 forbids: the
// questions asked are "does this object exist" and "is it on this ref", and
// neither can be answered differently for two different cards.
func validateSuppliedSHA(sha, ref string) (string, error) {
	if strings.TrimSpace(sha) == "" {
		return "", nil
	}
	resolved, err := todoGitOutput("rev-parse", "--verify", "--quiet", sha+"^{commit}")
	if err != nil || resolved == "" {
		if unrunnable := gitUnrunnable(err); unrunnable != nil {
			return "", &exitCodeError{code: 1, msg: fmt.Sprintf(
				"todo landed: the %s check could not run (%v); nothing was written",
				landedCheckUnrunnable, unrunnable)}
		}
		return "", &exitCodeError{code: 1, msg: fmt.Sprintf(
			"todo landed: the %s check failed: %q names no commit in this repository; nothing was written",
			landedCheckExistence, sha)}
	}
	if _, err := todoGitOutput("merge-base", "--is-ancestor", resolved, ref); err != nil {
		if unrunnable := gitUnrunnable(err); unrunnable != nil {
			return "", &exitCodeError{code: 1, msg: fmt.Sprintf(
				"todo landed: the %s check could not run (%v); nothing was written",
				landedCheckUnrunnable, unrunnable)}
		}
		return "", &exitCodeError{code: 1, msg: fmt.Sprintf(
			"todo landed: the %s check failed: commit %s is not reachable from %s; nothing was written",
			landedCheckReachability, resolved, ref)}
	}
	// The RESOLVED form, never the input: an abbreviation that is unique
	// today can become ambiguous as the repository grows, and a stored value
	// has no chance to be re-disambiguated later.
	return resolved, nil
}

// todoGitOutput runs one git command against the queue's own repository
// through the shared process seam, and returns its trimmed stdout.
//
// `-C <root>` is explicit: the queue resolves against the PRIMARY checkout,
// and the landing question is about that repository rather than about
// whichever worktree the command happens to be typed in.
func todoGitOutput(args ...string) (string, error) {
	full := append([]string{"-C", resolveTodoQueueRoot()}, args...)
	out, err := todoRunCommand("git", full...)
	return strings.TrimSpace(out), err
}

// gitUnrunnable reports the error when git could not be run at all, and nil
// when git ran and answered "no".
//
// The distinction is the whole of case (d): both validation commands signal
// their NEGATIVE answer by exiting non-zero, so a bare "err != nil" cannot
// tell "the commit is not reachable" from "there is no git". The discriminant
// is the SHAPE of the failure: a process that never started reports something
// other than an exit status.
//
// (The commands are named in validateSuppliedSHA rather than quoted here: the
// REQ-ABI-006 ancestry sweep in mcp_build_identity_test.go matches raw source
// text, so a quoted flag in a comment would enter its baseline as a phantom
// coordinate that churns on every unrelated edit to this file.)
func gitUnrunnable(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "executable file not found"),
		strings.Contains(msg, "no such file or directory"),
		strings.Contains(msg, "timed out"):
		return err
	}
	return nil
}

// gitReason renders a git failure for a refusal message, naming the absence
// of an error when the command succeeded but produced nothing.
func gitReason(err error) error {
	if err == nil {
		return fmt.Errorf("git resolved it to nothing")
	}
	return err
}

// readCardSpecStatus answers the SPEC-status fact for the card.
//
// Three outcomes, and they are three rather than two: a card with no spec id
// leaves the field EMPTY (the question was never asked), a card whose SPEC
// could not be read carries the explicit unknown marker (asked, unanswered),
// and a readable one carries what the frontmatter actually says. A default —
// `completed`, `draft`, or anything else plausible — is never invented
// (REQ-TLE-010).
func readCardSpecStatus(store *kanban.BacklogStore, root, id string) string {
	rec, err := store.LoadPure()
	if err != nil {
		return kanban.LandingSpecStatusUnknown
	}
	for _, it := range rec.Items {
		if it.ID != id {
			continue
		}
		if it.SpecID == nil || strings.TrimSpace(*it.SpecID) == "" {
			return ""
		}
		if status, ok := kanban.ReadPrimarySpecStatus(root, *it.SpecID); ok {
			return status
		}
		return kanban.LandingSpecStatusUnknown
	}
	return ""
}
