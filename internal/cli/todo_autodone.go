// todo_autodone.go — `moai todo auto-done` (SPEC-TODO-LAND-AUTO-DONE-001):
// the evidence-gated landing scan the LEAD runs immediately after its
// post-push remote-landing confirmation (`git fetch origin develop` +
// `git rev-parse origin/develop`).
//
// The scan closes landed cards at the moment landing is confirmed, with
// three misfire guards and full reversibility:
//
//  1. EVIDENCE IS A CLOSED SET (REQ-AD-004). A card closes on exactly one of
//     two forms — its recorded delivering SHA (operator-asserted via
//     `todo landed --sha`) reachable from the landed ref, or a subject on
//     the landed ref that `subjectAttribution` attributes to the card.
//     Everything else skips; nothing else is evidence.
//
//  2. THE SCAN IS THE ONLY SURFACE THAT CLOSES WITHOUT A PER-CARD `done`
//     (REQ-AD-002). No other verb, hook, or daemon reaches the archive
//     transition as a side effect of its own operation — a scope test pins
//     the two-verb allowlist.
//
//  3. AN INCONCLUSIVE EVALUATION NEVER CLOSES (REQ-AD-005). An unanswerable
//     git question yields `skip <id> reason=query-inconclusive`, never a
//     close: LandingUnknown is not evidence of landed, and the asymmetry
//     holds at the scan layer. Skip outcomes exit 0; exit 1 is reserved for
//     the scan itself being unable to run (the queue store unreadable).
//
// The three misfire guards live in the decision function the scan wires
// (internal/kanban/autodone_scan.go): M1 the reissued-id collision gate
// (`ambiguous-id`), M2 the sync gate (`spec-not-completed`), M3 the
// non-landing declaration exclusion (in the shared subject predicate).
//
// The scan does NOT write the Landing evidence column (REQ-AD-009): machine
// evidence lives only in this scan's own execution log, and the column's
// closed operator provenance set is untouched.
//
// SUBAGENT BOUNDARY (C-HRA-008 / REQ-TODO-014): nothing here prompts. Every
// path is structured stdout lines (or `--json`), human-readable errors on
// stderr.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// autoDoneLogFileName is the scan's append-only execution log, sibling to
// the queue's runtime state (REQ-AD-011). Machine-derived close evidence
// lives HERE and nowhere else — the Landing column's provenance set is
// frozen (REQ-AD-009).
//
// @MX:NOTE: [AUTO] `moai todo undone` reads this file to emit its reversal
// row (appendAutoDoneReversal); renaming the file or moving it out of
// RuntimeStateDirForRoot silently severs that pairing.
const autoDoneLogFileName = "auto-done-log.jsonl"

// autoDoneSource names the scan on every close line and log row, so a reader
// can tell an automated close from a manual `done`.
const autoDoneSource = "auto-land"

// autoDoneLogRow is one JSONL row of the execution log. Closed rows carry
// the answering evidence; skipped rows carry their reason; reversed rows
// name the closure row they invert.
type autoDoneLogRow struct {
	At          string `json:"at"`
	CardID      string `json:"card_id"`
	Outcome     string `json:"outcome"` // closed | skipped | reversed
	Form        string `json:"form,omitempty"`
	Subject     string `json:"subject,omitempty"`
	CommitSHA   string `json:"commit_sha,omitempty"`
	RecordedSHA string `json:"recorded_sha,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Ref         string `json:"ref,omitempty"`
	RefHead     string `json:"ref_head,omitempty"`
	OriginalAt  string `json:"original_at,omitempty"`
	Source      string `json:"source,omitempty"`
}

// autoDoneOutcome is one card's resolved scan result, in queue order.
type autoDoneOutcome struct {
	id          string
	specID      string
	closed      bool
	form        string
	reason      string
	subject     string
	commitSHA   string
	recordedSHA string
}

// newTodoAutoDoneCmd — `moai todo auto-done`.
func newTodoAutoDoneCmd() *cobra.Command {
	var fetch bool
	var dryRun bool
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "auto-done",
		Short: "Close queued/picked cards whose landing is evidenced on the landed ref",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTodoAutoDone(cmd, fetch, dryRun, jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&fetch, "fetch", false, "")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "")
	withResolvedLandedRef(cmd, func(landedRef string) {
		cmd.Long = todoAutoDoneLong(landedRef)
		cmd.Flags().Lookup("fetch").Usage = "Run git fetch for the landed ref's remote before evaluating (default: local remote-tracking refs only)"
	})
	return cmd
}

// todoAutoDoneLong renders the scan's help body against the ref it will ask
// about. It documents the two evidence forms, the CANONICAL four-token skip
// vocabulary, the exit-code policy, and the dry-run contract — the closed
// facts every reader of the scan's output keys on.
func todoAutoDoneLong(landedRef string) string {
	return `Close queued and picked cards whose landing is EVIDENCED on ` + landedRef + ` —
the step the LEAD runs immediately after the post-push remote-landing
confirmation (git fetch origin develop + git rev-parse origin/develop).
Lanes never run this scan: lanes never push, and a local merge is not a
landing.

Evidence is a CLOSED set of exactly two forms, both judged against the
landed ref:

  1. sha-recorded — the card's recorded delivering commit (the SHA an
     operator asserted via ` + "`moai todo landed --sha`" + `) resolves and is
     reachable from the landed ref at scan time.
  2. subject-attribution — the landed ref's subject stream carries a subject
     that attributes the card (the six positional shapes, plus the
     comma-form trailing parenthetical).

Everything else SKIPS — the scan is fail-closed. An id carried by more than
one distinct card text (a reissued id) skips on subject evidence alone,
because a subject mentioning the token cannot tell the predecessor's landed
work from this card's pending work; a recorded delivering SHA is the one
form that names THIS card, so it still closes through a reissued id. A card
whose SPEC frontmatter status is anything other than completed (an
unreadable status included) skips: a run commit landing does not license
the close while sync is unfinished. A commit subject carrying an explicit
non-landing declaration ("not merged" / "not landed") attributes nothing.

The skip-reason vocabulary is CLOSED at exactly four tokens:

  ambiguous-id        — guard M1: reissued id, subject evidence alone
  spec-not-completed  — guard M2: SPEC status is not completed (unknown included)
  not-landed          — no landing evidence, including a negated subject
  query-inconclusive  — the question could not be asked; never a close

Exit codes: 0 for every skip outcome (an inconclusive CARD is a skip, not a
command failure); 1 only when the scan itself cannot run (the queue store
is unreadable).

` + "`--fetch`" + ` runs git fetch for the landed ref's remote before evaluating.
Absent the flag the scan performs zero network fetches and reads the local
remote-tracking refs only.

` + "`--dry-run`" + ` evaluates every card and prints every close and skip line it
would produce, and writes NOTHING: the queue record stays byte-identical
and no log row is appended.

The scan writes no Landing evidence (a stored delivering commit stays
operator-asserted or absent); its machine evidence lives only in the
append-only execution log under the queue's runtime state directory. Every
close is reversible with ` + "`moai todo undone <n>`" + `, which appends a reversal
row naming the closure it inverts.`
}

// runTodoAutoDone is the scan.
//
// Facts are gathered OUTSIDE the queue lock (git subprocesses must not hold
// it — the same discipline todo landed records), the decision is the pure
// function, and the closes are applied inside ONE locked Mutate whose
// callback re-checks each planned close so a queue changed between snapshot
// and lock downgrades that card instead of refusing the whole scan.
func runTodoAutoDone(cmd *cobra.Command, fetch, dryRun, jsonOut bool) error {
	root := resolveTodoQueueRoot()
	store := newTodoStore()
	ref, _ := todoLandedRefResolved()

	snapshot, err := store.LoadPure()
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: todo auto-done: the queue could not be read: %v\n", err)
		return fmt.Errorf("todo auto-done: the queue could not be read: %w", err)
	}

	// REQ-AD-003: the fetch is OPT-IN. Absent the flag, zero network
	// fetches — the evaluation reads the local remote-tracking ref as it
	// stands. A failed fetch degrades to the same posture with a note.
	if fetch {
		if remote, branch, ok := splitLandedRefForFetch(ref); ok {
			if _, ferr := todoGitOutput("fetch", "--quiet", remote, branch); ferr != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
					"note: todo auto-done: git fetch %s %s failed (%v); evaluating the local remote-tracking ref as it stands\n",
					remote, branch, ferr)
			}
		}
	}

	refHead, _ := todoGitOutput("rev-parse", "--verify", "--quiet", gitEndOfOptions, ref+"^{commit}")

	commits, subjErr := kanban.ScanLandedSubjects(todoRunCommand, ref)
	subjectKnown := subjErr == nil
	var attributions map[string]kanban.LandedCommit
	if subjectKnown {
		attributions = kanban.LandedAttributions(commits, kanban.LandedBranchFromRef(ref))
	}

	outcomes := planAutoDone(snapshot, root, ref, subjectKnown, attributions)

	// Apply the closes in one locked write. Guards ran BEFORE the mutation
	// (on the snapshot) and the callback re-checks each card, so every
	// refusal inherits Mutate's byte-identity contract (C3).
	var applied []autoDoneOutcome
	if !dryRun {
		applied, err = applyAutoDoneCloses(store, outcomes)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: todo auto-done: %v\n", err)
			return err
		}
	} else {
		applied = effectiveOutcomes(outcomes)
	}

	stdout := cmd.OutOrStdout()
	if jsonOut {
		if err := writeAutoDoneJSON(stdout, ref, applied); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: todo auto-done: %v\n", err)
			return err
		}
	} else {
		writeAutoDoneLines(stdout, ref, applied)
	}

	if dryRun {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"note: todo auto-done: dry-run — the queue record was not modified and no log row was appended\n")
		return nil
	}

	// Post-mutation acts, in the done verb's order: factory state first, log
	// rows last. recordFactoryCardState is the done path's own propagation,
	// preserved so a scan close is indistinguishable downstream (AC-AD-009).
	for _, o := range applied {
		if o.closed {
			recordFactoryCardState(o.id, o.specID, "completed", "card.completed")
		}
	}
	appendAutoDoneRows(autoDoneLogPathFor(root), autoDoneLogRowsFor(ref, refHead, applied))
	return nil
}

// splitLandedRefForFetch splits ref into the remote and branch a single
// `git fetch <remote> <branch>` names. A ref without a remote part names no
// remote and fetches nothing (the caller then evaluates the local ref).
func splitLandedRefForFetch(ref string) (remote, branch string, ok bool) {
	i := strings.Index(ref, "/")
	if i <= 0 || i == len(ref)-1 {
		return "", "", false
	}
	return ref[:i], ref[i+1:], true
}

// planAutoDone decides every live queued/picked card. Live means exactly
// BacklogStateQueued or BacklogStatePicked (REQ-AD-001) — dropped cards are
// never evaluated, already-archived ids are not re-evaluated (NFR-4's
// idempotence), and the record's shape makes both true by construction.
func planAutoDone(snapshot *kanban.BacklogRecord, root, ref string, subjectKnown bool, attributions map[string]kanban.LandedCommit) []autoDoneOutcome {
	outcomes := make([]autoDoneOutcome, 0, len(snapshot.Items))
	for i := range snapshot.Items {
		it := snapshot.Items[i]
		if it.State != kanban.BacklogStateQueued && it.State != kanban.BacklogStatePicked {
			continue
		}
		o := autoDoneOutcome{id: it.ID}
		if it.SpecID != nil {
			o.specID = *it.SpecID
		}

		// Guard M2's input: the SPEC frontmatter status read at scan time.
		// No spec id means the gate does not apply (the Class A/B shape);
		// anything other than a read `completed` is not a pass (unknown
		// included).
		gate := kanban.AutoDoneYes
		if strings.TrimSpace(o.specID) != "" {
			if status, ok := kanban.ReadPrimarySpecStatus(root, o.specID); !ok || status != "completed" {
				gate = kanban.AutoDoneNo
			}
		}

		facts := kanban.AutoDoneFacts{
			SubjectKnown:  subjectKnown,
			DistinctTexts: kanban.AutoDoneDistinctTexts(snapshot, it.ID),
			SpecSyncGate:  gate,
		}
		if it.Landing != nil {
			o.recordedSHA = strings.TrimSpace(it.Landing.SHA)
		}
		if o.recordedSHA != "" {
			facts.RecordedSHA = o.recordedSHA
			facts.SHAReachable = todoAutoDoneSHAReachable(o.recordedSHA, ref)
		}
		if subjectKnown {
			if hit, ok := attributions[it.ID]; ok && kanban.AutoDoneSubjectFresh(hit, it.AddedAt) {
				o.subject = hit.Subject
				o.commitSHA = hit.SHA
				facts.SubjectHit = &hit
			}
		}

		decision := kanban.AutoDoneDecide(facts)
		if decision.Close {
			o.closed = true
			o.form = decision.Form
		} else {
			o.reason = decision.Reason
		}
		outcomes = append(outcomes, o)
	}
	return outcomes
}

// todoAutoDoneSHAReachable answers whether sha is reachable from ref — the
// same ancestry question the recorded-SHA validation asks at record time,
// re-asked at scan time against the ref as it NOW stands. A git exit code of
// exactly 1 is the NEGATIVE answer (not an ancestor); any other failure is
// an unanswerable question, never a negative.
func todoAutoDoneSHAReachable(sha, ref string) kanban.AutoDoneTri {
	_, err := todoGitOutput("merge-base", "--is-ancestor", gitEndOfOptions, sha, ref)
	if err == nil {
		return kanban.AutoDoneYes
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return kanban.AutoDoneNo
	}
	if gitUnrunnable(err) != nil {
		return kanban.AutoDoneUnknown
	}
	return kanban.AutoDoneUnknown
}

// applyAutoDoneCloses archives the planned closes in one locked Mutate. A
// planned close whose card moved or vanished between snapshot and lock is
// downgraded to a skip (query-inconclusive) rather than refusing the whole
// scan; a store-level failure (lock, unreadable engine) refuses everything
// and is the scan-cannot-run case.
func applyAutoDoneCloses(store *kanban.BacklogStore, outcomes []autoDoneOutcome) ([]autoDoneOutcome, error) {
	hasCloses := false
	for _, o := range outcomes {
		if o.closed {
			hasCloses = true
			break
		}
	}
	if !hasCloses {
		return outcomes, nil
	}
	err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for k := range outcomes {
			if !outcomes[k].closed {
				continue
			}
			at := -1
			for i := range rec.Items {
				if rec.Items[i].ID == outcomes[k].id {
					at = i
					break
				}
			}
			if at < 0 {
				// The queue changed between snapshot and lock; this scan's
				// facts for the card are stale.
				outcomes[k].downgrade(kanban.AutoDoneSkipQueryInconclusive)
				continue
			}
			if err := rec.ArchiveCard(outcomes[k].id); err != nil {
				outcomes[k].downgrade(kanban.AutoDoneSkipQueryInconclusive)
				continue
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("todo auto-done: %w", err)
	}
	return outcomes, nil
}

// effectiveOutcomes is the dry-run pass-through: the planned outcomes ARE
// the effective ones, because a dry run mutates nothing that could invalidate
// a plan.
func effectiveOutcomes(outcomes []autoDoneOutcome) []autoDoneOutcome {
	return outcomes
}

// downgrade turns a planned close into a skip.
func (o *autoDoneOutcome) downgrade(reason string) {
	o.closed = false
	o.form = ""
	o.reason = reason
	o.subject = ""
	o.commitSHA = ""
}

// writeAutoDoneLines renders the stdout contract (REQ-AD-010): one close
// line per closed card on the canonical `done <id> landing=landed` prefix
// every existing reader keys off (C4), one skip line per skipped card, and
// the summary line last.
func writeAutoDoneLines(w io.Writer, ref string, outcomes []autoDoneOutcome) {
	closed, skipped := 0, 0
	for _, o := range outcomes {
		if o.closed {
			closed++
			_, _ = fmt.Fprintf(w, "done %s landing=landed source=%s ref=%s form=%s\n",
				o.id, autoDoneSource, ref, o.form)
			continue
		}
		skipped++
		_, _ = fmt.Fprintf(w, "skip %s reason=%s\n", o.id, o.reason)
	}
	_, _ = fmt.Fprintf(w, "auto-done scanned=%d closed=%d skipped=%d ref=%s\n",
		len(outcomes), closed, skipped, ref)
}

// autoDoneJSONRow / autoDoneJSONReport — the machine-readable surface.
type autoDoneJSONRow struct {
	Card   string `json:"card"`
	Form   string `json:"form,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type autoDoneJSONReport struct {
	Ref     string            `json:"ref"`
	Closed  []autoDoneJSONRow `json:"closed"`
	Skipped []autoDoneJSONRow `json:"skipped"`
}

func writeAutoDoneJSON(w io.Writer, ref string, outcomes []autoDoneOutcome) error {
	report := autoDoneJSONReport{Ref: ref, Closed: []autoDoneJSONRow{}, Skipped: []autoDoneJSONRow{}}
	for _, o := range outcomes {
		if o.closed {
			report.Closed = append(report.Closed, autoDoneJSONRow{Card: o.id, Form: o.form})
			continue
		}
		report.Skipped = append(report.Skipped, autoDoneJSONRow{Card: o.id, Reason: o.reason})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// autoDoneLogPathFor resolves the execution log under the queue's runtime
// state directory (REQ-AD-011) — machine-local, matching the queue's
// locality.
func autoDoneLogPathFor(root string) string {
	return filepath.Join(kanban.RuntimeStateDirForRoot(root), autoDoneLogFileName)
}

// autoDoneLogPath resolves the log for the current queue (the undone verb's
// reversal row uses this).
func autoDoneLogPath() string {
	return autoDoneLogPathFor(resolveTodoQueueRoot())
}

// autoDoneLogRowsFor renders one log row per outcome (REQ-AD-011): closed
// rows carry the answering evidence and the landed ref + head at scan time;
// skipped rows carry their closed-vocabulary reason.
func autoDoneLogRowsFor(ref, refHead string, outcomes []autoDoneOutcome) []autoDoneLogRow {
	now := time.Now().UTC().Format(time.RFC3339)
	rows := make([]autoDoneLogRow, 0, len(outcomes))
	for _, o := range outcomes {
		row := autoDoneLogRow{
			At:          now,
			CardID:      o.id,
			Ref:         ref,
			RefHead:     refHead,
			Form:        o.form,
			Subject:     o.subject,
			CommitSHA:   o.commitSHA,
			RecordedSHA: o.recordedSHA,
			Reason:      o.reason,
			Source:      autoDoneSource,
		}
		if o.closed {
			row.Outcome = "closed"
			row.Reason = ""
		} else {
			row.Outcome = "skipped"
			row.Form = ""
			row.Subject = ""
			row.CommitSHA = ""
			row.RecordedSHA = ""
		}
		rows = append(rows, row)
	}
	return rows
}

// appendAutoDoneRows appends rows to the execution log. A log failure is
// reported on the command's stderr and does not undo the closes: the
// mutation has landed, and a swallowed write failure would make the audit
// trail lie by absence — so it is surfaced, never silent.
func appendAutoDoneRows(path string, rows []autoDoneLogRow) {
	if len(rows) == 0 {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		moaiStderrNote("todo auto-done: creating log dir: %v", err)
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		moaiStderrNote("todo auto-done: opening log: %v", err)
		return
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			moaiStderrNote("todo auto-done: encoding log row: %v", err)
			return
		}
	}
}

// appendAutoDoneReversal appends the reversal row REQ-AD-012 names: when
// `undone` restores a card the scan closed, the log gains a row naming the
// original closure. A manually-closed card has no closure row, and gains no
// reversal row — the log is the scan's audit trail, not every close's.
// Fail-open: a reversal that cannot be logged never un-restores the card.
func appendAutoDoneReversal(id string) {
	path := autoDoneLogPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var original *autoDoneLogRow
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row autoDoneLogRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		if row.CardID == id && row.Outcome == "closed" {
			copied := row
			original = &copied
		}
	}
	if original == nil {
		return
	}
	appendAutoDoneRows(path, []autoDoneLogRow{{
		At:         time.Now().UTC().Format(time.RFC3339),
		CardID:     id,
		Outcome:    "reversed",
		Form:       original.Form,
		OriginalAt: original.At,
		Source:     autoDoneSource,
	}})
}

// moaiStderrNote writes one note to the process stderr.
func moaiStderrNote(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "note: "+format+"\n", args...)
}
