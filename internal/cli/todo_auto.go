package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/session"
)

// `/moai:todo --auto` — the serial card-processing cycle (SPEC-MANAGER-TODO-001).
//
// The cycle is the factory foreman skill's contract driven from the CLI surface:
// pick one card → emit the dispatch directive for ONE isolated in-session
// worker → judge completion only by reading the worker's disk evidence →
// record `done` on that evidence → emit the /clear guidance → accept the next
// card. Cards are processed strictly one at a time; a worker that dies or
// leaves no readable evidence is unpicked back to `queued` with a labelled
// non-finding, never silently done. The invocation itself is the operator's
// batch approval: it authorizes serial consumption of the queue and nothing
// else. The cycle carries one auto-scoped ranking exception — it may rank the
// queued candidates it is about to accept (todo_auto_rank.go), which changes
// its selection order only — and it never admits, drops, or edits cards; the
// queue itself is unchanged.

// autoEvidenceRelPath is the evidence file the dispatch directive names, per
// the foreman convention. Completion is judged by reading THIS file, never by
// the worker's claims.
func autoEvidencePath(root, cardID string) string {
	return filepath.Join(root, ".moai", "reports", cardID, "evidence.md")
}

// autoRegistryEntry is the slice of a session-registry row the liveness
// registry channel needs.
type autoRegistryEntry struct {
	Cwd string
	PID int
}

// autoLiveness carries the two owner-measurement channels (design D-4). Both
// are seams so tests inject fixture state; the production wiring shells out
// to the session registry and lsof, exactly like the worktree-move guard.
//
// Decision rule: EITHER channel showing life means the owner is alive — a
// false "alive" costs a skipped card, a false "dead" steals a living
// session's work. Measurement is re-taken at every pickup decision, never
// cached across cards (AC-MT-011 non-cache arm).
type autoLiveness struct {
	registryEntries func() ([]autoRegistryEntry, error)
	processCWDs     func() ([]string, error)
	pidAlive        func(pid int) bool
}

// newAutoLiveness returns the production measurement channels.
func newAutoLiveness() autoLiveness {
	return autoLiveness{
		registryEntries: func() ([]autoRegistryEntry, error) {
			entries, err := session.QueryActiveWork("")
			if err != nil {
				return nil, err
			}
			out := make([]autoRegistryEntry, 0, len(entries))
			for _, e := range entries {
				out = append(out, autoRegistryEntry{Cwd: e.CWD, PID: e.PID})
			}
			return out, nil
		},
		processCWDs: activeProcessCWDs,
		pidAlive: func(pid int) bool {
			if pid <= 0 {
				return false
			}
			return exec.Command("kill", "-0", fmt.Sprint(pid)).Run() == nil
		},
	}
}

// autoOwnerWorktrees names the candidate owner trees for a card: the tree
// keeps the card id by convention, under either worktree root.
func autoOwnerWorktrees(root, cardID string) []string {
	return []string{
		filepath.Join(root, ".claude", "worktrees", cardID),
		filepath.Join(root, ".moai", "worktrees", cardID),
	}
}

func autoInsideTree(dir, tree string) bool {
	rel, err := filepath.Rel(tree, dir)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && rel != "")
}

// ownerAlive judges one picked card's owner on both channels. The second
// return reports a degraded measurement (lsof unavailable): the cycle then
// decides on the registry channel alone and labels the notice. An error from
// BOTH channels is measurement failure — the conservative default holds and
// the owner is reported alive (no takeover on an unmeasurable board).
func (lv autoLiveness) ownerAlive(root, cardID string) (alive bool, degraded bool, err error) {
	trees := autoOwnerWorktrees(root, cardID)
	registryLive := false
	entries, regErr := lv.registryEntries()
	if regErr == nil {
		for _, e := range entries {
			for _, tree := range trees {
				if autoInsideTree(e.Cwd, tree) && lv.pidAlive(e.PID) {
					registryLive = true
				}
			}
		}
	}
	cwds, procErr := lv.processCWDs()
	degraded = procErr != nil
	processLive := false
	if procErr == nil {
		for _, dir := range cwds {
			for _, tree := range trees {
				if autoInsideTree(dir, tree) {
					processLive = true
				}
			}
		}
	}
	if regErr != nil && procErr != nil {
		// Nothing was measurable — refuse the takeover.
		return true, degraded, regErr
	}
	return registryLive || processLive, degraded, nil
}

// autoPickTargets selects the cycle's pickup targets (REQ-MT-008/010):
// first the unfinished already-picked cards whose owner measures dead, then
// the unpicked cards in queue order. The predicate is written in POSITIVE
// state vocabulary — `state == queued` plus the dead-owner `picked`
// carve-out — so a future card state (a `hold`, say) is excluded from pickup
// automatically. Liveness is re-measured per decision; the notes slice
// carries the labelled non-findings (degraded measurement, live-owner skip).
func autoPickTargets(rec *factory.BacklogRecord, lv autoLiveness, root string) (targets []factory.BacklogItem, notes []string, err error) {
	for _, it := range rec.Items {
		// REQ-THS-012: machine selectors enumerate the states they accept
		// positively — this scan admits exactly `picked` (dead-owner rescue).
		if it.State == factory.BacklogStatePicked {
			alive, degraded, liveErr := lv.ownerAlive(root, it.ID)
			if liveErr != nil {
				notes = append(notes, fmt.Sprintf("non-finding: %s owner liveness unmeasurable (%v) — card skipped, no takeover", it.ID, liveErr))
				continue
			}
			if alive {
				notes = append(notes, fmt.Sprintf("non-finding: %s owner measured alive — card untouchable", it.ID))
				continue
			}
			if degraded {
				notes = append(notes, fmt.Sprintf("non-finding: %s owner judged from the session registry alone (lsof unavailable — degraded measurement)", it.ID))
			}
			targets = append(targets, it)
		}
	}
	for _, it := range rec.Items {
		if it.State == factory.BacklogStateQueued {
			// SPEC-RELATION-PICKUP-FILTER-001 (REQ-RPF-001/002): a queued
			// card on the blocked side of a live sequencing finding waits
			// for its predecessor — it is not a pickup candidate while the
			// finding exists in the live record (the predecessor's done
			// archives the finding, resolving the block with no relation
			// bookkeeping). The dead-owner rescue arm above is deliberately
			// NOT gated on relations (REQ-RPF-006).
			if blockers := rec.FindingsBlocking(it.ID); len(blockers) > 0 {
				// REQ-RPF-004: one labelled non-finding per skipped card,
				// naming the card id, the relation, and the blocking
				// predecessor id. FindingsBlocking returns only findings
				// whose blocked side (the waits-on waiter) IS this card, so
				// the edge's target is always the predecessor.
				for _, f := range blockers {
					_, predecessor, _ := factory.WaitsOnOf(f)
					notes = append(notes, fmt.Sprintf(
						"non-finding: %s skipped (relation-blocked: %s %s %s) — waiting for predecessor %s",
						it.ID, f.SubjectID, f.Relation, f.RelatedID, predecessor))
				}
				continue
			}
			targets = append(targets, it)
		}
	}
	return targets, notes, nil
}

// autoOptions carries the cycle's seams and knobs.
type autoOptions struct {
	wait      time.Duration            // evidence deadline per card
	liveness  autoLiveness             // owner-measurement channels
	sessionID string                   // the invoking (operator) session, named in the guidance
	sleep     func(time.Duration)      // poll-tick seam (tests drive evidence arrival here)
	now       func() time.Time         // clock seam for the deadline
	jev       func(root string) string // display-only consultation seam (tests stub it)

	// The two inputs of the ranking stage. A nil value is INERT: a nil landed
	// lookup leaves the landed signal unmeasured, a nil Jev ranker is Jev
	// unavailable (`jev-disabled`) — neither runs a subprocess or sends a
	// request. Production wiring (todo.go) sets both live seams.
	landed  autoLandedLookup // landed-state lookup over the whole record
	jevRank autoJevRanker    // one bounded Jev request over the candidates

	// quota is the quota-pressure steering seam (SPEC-QUOTA-AWARE-SCHEDULING-001
	// REQ-QAS-019): it returns the one line printed immediately before each
	// accept line, or "" when there is nothing to say. A nil value is INERT — no
	// line, no read of any quota record or registry. Printing only: the line
	// never reaches a queue write, a lease, or a dispatch.
	quota autoQuotaLine
}

// autoQuotaLine returns the quota-pressure steering line for the project root,
// or "" while pressure is off. It is evaluated afresh before every accept line.
type autoQuotaLine func(root string) string

// runAutoCycle executes the serial cycle against the store, writing the
// narrated output (accept → directive → evidence → done/unpick → guidance)
// to out. Exactly one card is in flight at any time.
func runAutoCycle(out io.Writer, store *factory.BacklogStore, root string, opts autoOptions) error {
	if opts.sleep == nil {
		opts.sleep = time.Sleep
	}
	if opts.now == nil {
		opts.now = time.Now
	}
	if opts.sessionID == "" {
		opts.sessionID = os.Getenv(config.EnvClaudeCodeSessionID)
		if opts.sessionID == "" {
			opts.sessionID = "operator session (id unavailable)"
		}
	}
	if opts.jev == nil {
		opts.jev = consultJev
	}

	// The script consultation is display-only (REQ-MT-014/015): its signal is
	// rendered verbatim as a labelled line and consumed by NO decision —
	// never a queue mutation, a completion verdict, a merge approval, or an
	// operator gate. Absent scripts or an absent key degrade to a labelled
	// non-finding and the cycle proceeds on lead judgment alone, exit 0.
	// This line is not the ranking input: the one place a Jev answer informs
	// the cycle is the ranking stage below (opts.jevRank, the Go capability),
	// and it sets the selection order only.
	_, _ = fmt.Fprintln(out, opts.jev(root))

	rec, err := store.LoadPure()
	if err != nil {
		return err
	}
	targets, notes, err := autoPickTargets(rec, opts.liveness, root)
	if err != nil {
		return err
	}
	for _, n := range notes {
		_, _ = fmt.Fprintln(out, n)
	}
	// Ranking stage (SPEC-TODO-AUTO-PRIORITY-001): orders the queued suffix of
	// the targets and prints the selection record; the dead-owner rescue
	// targets stay first. It writes nothing to the queue.
	targets = autoRankTargets(out, rec, targets, opts)
	if len(targets) == 0 {
		_, _ = fmt.Fprintln(out, "no eligible card: queue is empty or every card is untouched-by-authority (nothing to do)")
		return nil
	}

	for _, card := range targets {
		// Quota steering (SPEC-QUOTA-AWARE-SCHEDULING-001 REQ-QAS-019): one
		// line immediately before the accept line, evaluated afresh per card.
		// Printing only — no queue write, no lease, no dispatch — and absent
		// entirely while pressure is off or the seam is nil.
		if opts.quota != nil {
			if line := opts.quota(root); line != "" {
				_, _ = fmt.Fprintln(out, line)
			}
		}
		_, _ = fmt.Fprintf(out, "accept %s %s\n", card.ID, todoTextPrefix(card.Text))
		// Claim the card before dispatch: a queued card becomes picked (the
		// cycle's own pick); a dead-owner picked card is already claimed. The
		// card this cycle picked is the only one it may later close.
		if card.State == factory.BacklogStateQueued {
			if err := store.Mutate(func(r *factory.BacklogRecord) error {
				for i := range r.Items {
					if r.Items[i].ID == card.ID {
						// REQ-THS-012: positive enumeration — the pick
						// admits exactly `queued`, refuses everything else
						// by name.
						if r.Items[i].State == factory.BacklogStateQueued {
							r.Items[i].State = factory.BacklogStatePicked
							return nil
						}
						return fmt.Errorf("auto: card %s is %s, not queued — refusing the pick", card.ID, r.Items[i].State)
					}
				}
				return fmt.Errorf("auto: card %s vanished", card.ID)
			}); err != nil {
				_, _ = fmt.Fprintf(out, "non-finding: %s (%v)\n", card.ID, err)
				continue
			}
		}
		evidence := autoEvidencePath(root, card.ID)
		writeAutoDirective(out, card, evidence)

		// One worker in flight: wait for the evidence the directive named,
		// polling until the per-card deadline. Completion is judged by
		// reading the file, never by any worker claim.
		collected := false
		deadline := opts.now().Add(opts.wait)
		for opts.now().Before(deadline) {
			if body, readErr := os.ReadFile(evidence); readErr == nil && len(strings.TrimSpace(string(body))) > 0 {
				collected = true
				break
			}
			opts.sleep(5 * time.Second)
		}

		if collected {
			_, _ = fmt.Fprintf(out, "evidence collected: %s\n", evidence)
			err := store.Mutate(func(r *factory.BacklogRecord) error {
				for i := range r.Items {
					if r.Items[i].ID == card.ID {
						// REQ-THS-012: positive enumeration — the done
						// admits exactly `picked`, refuses everything else
						// by name.
						if r.Items[i].State == factory.BacklogStatePicked {
							// The leader-approval gate (SPEC-FACTORY-COMPLETION-RECOVERY-001
							// REQ-FCR-002a): the third completion surface
							// verifies the receipt at the archive moment,
							// serialized exactly like the manual done path.
							gate, gateErr := holdDoneApprovalGate(context.Background(), root, card.ID)
							if gateErr != nil {
								return gateErr
							}
							if err := gate.verify(context.Background(), todoCardUUID(&r.Items[i])); err != nil {
								gate.refuse()
								return err
							}
							if err := r.ArchiveCard(card.ID); err != nil {
								gate.refuse()
								return err
							}
							if err := gate.commit(); err != nil {
								return err
							}
							return nil
						}
						return fmt.Errorf("auto: card %s is %s, not picked — changed hands mid-flight", card.ID, r.Items[i].State)
					}
				}
				return fmt.Errorf("auto: card %s vanished", card.ID)
			})
			if err != nil {
				// The card changed hands mid-flight: this cycle's claim is
				// gone — report the non-finding and move on, never archive.
				_, _ = fmt.Fprintf(out, "non-finding: %s (%v)\n", card.ID, err)
				continue
			}
			_, _ = fmt.Fprintf(out, "done %s\n", card.ID)
			writeAutoClearGuidance(out, card.ID, opts.sessionID)
			continue
		}

		// Failure path: no readable evidence at the deadline — unpick with a
		// labelled non-finding. Never silently done, never left picked here.
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == card.ID {
					if r.Items[i].State == factory.BacklogStatePicked {
						r.Items[i].State = factory.BacklogStateQueued
						r.Items[i].SpecID = nil
					}
					return nil
				}
			}
			return nil
		}); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "unpick %s non-finding: worker evidence absent at deadline (%s) — card returned to queued, never done\n", card.ID, evidence)
	}
	return nil
}

// consultJev asks the local Jev scripts for a dispatch-order/priority signal.
// The return value is a labelled display line — the ONLY thing the cycle does
// with it is print it. No decision reads it: Jev output is never the basis of
// a queue mutation, a completion verdict, a merge approval, or any
// operator-gate decision (REQ-MT-015). An absent script, key, or network
// degrades to a labelled non-finding; degradation is never an error exit.
// The ranking stage's Jev consumer is a separate seam (autoJevRanker) and does
// not read this line.
func consultJev(root string) string {
	script := filepath.Join(root, "scripts", "jev", "route.sh")
	if _, err := os.Stat(script); err != nil {
		return "jev: unavailable (no local scripts) — labelled non-finding; proceeding on the operator session's own judgment"
	}
	out, err := exec.Command(script).Output()
	if err != nil {
		return fmt.Sprintf("jev: consultation failed (%v) — labelled non-finding; proceeding on the operator session's own judgment", err)
	}
	return "jev signal (display-only): " + strings.TrimSpace(string(out))
}

// writeAutoDirective emits the fixed-field dispatch address block — a
// pointer, not a copy, ten lines at most. The directive names ONE isolated
// in-session Agent() worker (isolation: worktree); it creates no factory
// lease and claims no slot.
func writeAutoDirective(out io.Writer, card factory.BacklogItem, evidence string) {
	_, _ = fmt.Fprintln(out, "dispatch (one isolated in-session Agent() worker, isolation: worktree):")
	_, _ = fmt.Fprintf(out, "card: %s\n", card.ID)
	if card.SpecID != nil && *card.SpecID != "" {
		_, _ = fmt.Fprintf(out, "spec: %s\n", *card.SpecID)
	}
	_, _ = fmt.Fprintf(out, "evidence: %s\n", evidence)
	_, _ = fmt.Fprintln(out, "worker orders: implement the card lane-locally; write the evidence file above (decisions, verbatim output tails, gaps, residual risk); commit by explicit pathspec; never push, never merge.")
}

// writeAutoClearGuidance emits the per-card /clear guidance (REQ-MT-011):
// unconditional for every completed card, naming the completed card, the
// next step, and the invoking (operator) session — this session, the one
// hosting the cycle and the next card's dispatch, never the worker's.
func writeAutoClearGuidance(out io.Writer, cardID, sessionID string) {
	_, _ = fmt.Fprintf(out, "--- /clear guidance ---\n")
	_, _ = fmt.Fprintf(out, "card %s is complete. Clear this session (/clear) before the next card.\n", cardID)
	_, _ = fmt.Fprintf(out, "next step: re-run `moai todo --auto` to continue the queue; next pickup follows the same order.\n")
	_, _ = fmt.Fprintf(out, "operator session: %s\n", sessionID)
	_, _ = fmt.Fprintf(out, "----------------------\n")
}
