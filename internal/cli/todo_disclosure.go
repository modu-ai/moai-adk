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

// discloseQueueLayout probes the queue layout and discloses, for the read
// verbs that do not already hold a vouch.
//
// The probe runs BEFORE the verb's own read: BacklogStore.Load adopts —
// on a State D layout carrying an interrupted-migration marker it completes
// the quarantine, which renames the very file this line reports.
func discloseQueueLayout(cmd *cobra.Command, verb string) error {
	return discloseNonAuthoritativeBacklogJSON(cmd.ErrOrStderr(), verb,
		kanban.InspectBacklogArchiveVouch(newTodoStore().Path()))
}
