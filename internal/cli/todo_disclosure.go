// todo_disclosure.go — SPEC-BACKLOG-JSON-DISCLOSURE-001 (card t395): a
// `backlog.json` at the canonical queue path is not the queue, and the read
// surface says so.
//
// The file answers a direct read silently. A human who cats it, or an agent
// told by a stale instruction that queue state lives there, gets a
// confident wrong answer with no signal that the database beside it is what
// `moai todo` actually operates on. The writer that produced the one in
// this repository was never identified, so a cleanup cannot prevent
// recurrence — making the READING side able to tell is the defence that
// holds whoever the writer turns out to be (spec.md §A.3).
//
// The fact rides the EXISTING store-identity surface,
// kanban.InspectBacklogArchiveVouch, which already measured it and threw it
// away (REQ-BJD-006). There is no second inspector and no second probe.
//
// stderr only (REQ-BJD-004): stdout is a machine surface for these verbs —
// `moai todo list --json` is read by the foreman loop — and must stay
// byte-identical to its no-JSON form.
package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// discloseNonAuthoritativeBacklogJSON writes one line naming the store that
// answered and naming the backlog.json beside it as not authoritative, and
// writes nothing at all when there is nothing to disclose (REQ-BJD-003).
// It touches no file and takes no lock.
func discloseNonAuthoritativeBacklogJSON(w io.Writer, verb string, vouch kanban.BacklogArchiveVouch) error {
	if !vouch.NonAuthoritativeJSON {
		return nil
	}
	_, err := fmt.Fprintf(w,
		"%s: answered by %s; the backlog.json beside it is NOT the queue — an export or a legacy leftover, whose contents can be arbitrarily stale\n",
		verb, vouch.Store)
	return err
}

// discloseStaleLocalStores — SPEC-TODO-STALE-STORE-001 (card t1307)
// REQ-TSS-001/005: when a read verb is answered by the home database while
// a stale project-local queue store diverges from it, one stderr line per
// divergent store names the store's path and BOTH last_seq values. A tree
// with no legacy store, or one whose last_seq matches the home database,
// discloses no divergence line (REQ-TSS-005). stderr only (REQ-TSS-002) —
// the line is separate from the backlog.json disclosure above, never
// instead of it.
//
// SPEC-TODO-SURFACE-POLISH-001 REQ-TSP-041: the same entry point carries
// the once-per-class ghost artifact notice — every read and write verb
// passes through here, so the ghost discovery cannot be silent on one
// surface and loud on another, and the once-only marker cannot drift
// between verbs.
func discloseStaleLocalStores(w io.Writer, verb string, fact kanban.StaleStoreFact) error {
	if fact.Divergent {
		for _, st := range fact.Stores {
			if !st.Readable || st.LastSeq == fact.HomeLastSeq {
				continue
			}
			if _, err := fmt.Fprintf(w,
				"%s: a stale project-local queue store exists at %s (last_seq %d) while the home database answered (last_seq %d) — the store is NOT the queue; a rollback snapshot whose contents can be arbitrarily stale\n",
				verb, st.Path, st.LastSeq, fact.HomeLastSeq); err != nil {
				return err
			}
		}
	}
	return discloseGhostStoresOnce(w, verb, fact)
}

// todoQueueRootForDisclosure is the queue root every read-verb disclosure
// resolves stale-store facts against — the same root newTodoReadStore
// resolves its answering store through, so a disclosure can never name a
// store the verb's own read never considered.
func todoQueueRootForDisclosure() string {
	return kanban.ResolveTodoQueueRoot(resolveProjectDir())
}

// @MX:ANCHOR fan_in=5 - SPEC-BACKLOG-JSON-DISCLOSURE-001 REQ-BJD-002 sole
// disclosure entry point for the read verbs that do not already hold a vouch
// (bare/list, show, why, pr, triage); history calls
// discloseNonAuthoritativeBacklogJSON directly with the vouch it already
// holds.
// discloseQueueLayout probes the queue layout and discloses, for the read
// verbs that do not already hold a vouch.
//
// The probe runs BEFORE the verb's own read: BacklogStore.Load adopts —
// on a State D layout carrying an interrupted-migration marker it completes
// the quarantine, which renames the very file this line reports.
//
// SPEC-TODO-STALE-STORE-001: the same entry point carries the stale
// project-local store disclosure (REQ-TSS-001) — the divergence fact comes
// from the single kanban detector the doctor check also uses
// (REQ-TSS-004), so the two surfaces cannot disagree.
func discloseQueueLayout(cmd *cobra.Command, verb string) error {
	if err := discloseNonAuthoritativeBacklogJSON(cmd.ErrOrStderr(), verb,
		kanban.InspectBacklogArchiveVouch(newTodoReadStore().Path())); err != nil {
		return err
	}
	return discloseStaleLocalStores(cmd.ErrOrStderr(), verb,
		kanban.InspectStaleLocalStores(todoQueueRootForDisclosure()))
}
