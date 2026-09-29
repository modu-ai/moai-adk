// todo_claim.go — SPEC-TODO-CLAIM-LEASE-001 M3: `moai todo claim
// [--lane <label>] [--renew <id>]`, the operator/leader-side atomic claim
// verb over the store's Mutate-callback claim family.
//
// One implementation per verb (design.md §3): runTodoClaimRoot is the body
// both the cobra RunE and the MCP todo_claim tool call, anchored at an
// explicit root. The verb is a queue MUTATION — it is deliberately absent
// from todoLaneReadOnlyVerbs, so a lane session is refused in both the bare
// and the --lane form by the unchanged REQ-SD-015 guard (C6 flag-form guard
// extension; the flag form grants nothing — see the note at
// todoRefuseLaneMutation).
//
// SUBAGENT BOUNDARY: nothing here prompts (headless-safe, like the rest of
// the todo surface).
package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// newTodoClaimCmd — `moai todo claim [--lane <label>] [--renew <id>]`.
// Bare: claim the oldest queued card under a fresh lease
// (DefaultFactoryLeaseDuration). --lane attributes the claim to an
// operator/leader-supplied lane label. --renew <id> extends the addressed
// card's lease instead of claiming a new one (holder label must match).
func newTodoClaimCmd() *cobra.Command {
	var lane, renew string
	cmd := &cobra.Command{
		Use:   "claim [--lane <label>] [--renew <id>]",
		Short: "Claim the oldest queued card under a lease (atomic CAS + lease)",
		Long: `Claim the oldest queued card as one atomic compare-and-set: the card
moves to picked with a lease (picked_by, lease_expires_at = now +
DefaultFactoryLeaseDuration, picked_at), and the confirmation carries the
card id, its text prefix, and the expiry instant.

Before selecting, every lapsed lease in the queue is returned to queued
(expiry-first): the returned cards are announced with their id and PREVIOUS
holder, then the claim takes the oldest — which may be the card just
returned. An unparseable expiry is judged expired.

` + "`--lane <label>`" + ` attributes the claim to an operator/leader-supplied lane
label; without it the claim is the operator's own (picked_by=operator).
The flag grants nothing to a lane session: a session for which the lane
boundary holds is refused with or without --lane.

` + "`--renew <id>`" + ` extends the addressed card's lease instead of claiming.
Only the current holder's label (the --lane value, or operator) may renew;
a foreign label is refused with no change, and renewing a lapsed lease
returns the card to queued (committed) and refuses — it never extends.

No eligible card exits with status 3 and a non-error message, so a
supervising launcher can distinguish it from failure.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTodoClaimRoot(resolveTodoQueueRoot(), cmd, lane, renew)
		},
	}
	cmd.Flags().StringVar(&lane, "lane", "",
		"Attribute the claim to this operator/leader-supplied lane label")
	cmd.Flags().StringVar(&renew, "renew", "",
		"Renew the addressed card's lease instead of claiming a new card")
	return cmd
}

// todoJSONProjection strips the lease fields from a record copy destined
// for the `todo list --json` surface. REQ-TCL-014/AC-TCL-008: the JSON face
// stays byte-identical to the frozen golden fixture and NEVER exposes the
// new columns — the human face may (REQ-TCL-010), which is why the fields
// stay on BacklogItem and the exclusion happens at this one render, not at
// the type. The copy is shallow on the record and fresh on the item slice:
// mutating the shared item structs would strip the fields the caller's
// still-open record holds.
func todoJSONProjection(rec *kanban.BacklogRecord) *kanban.BacklogRecord {
	out := *rec
	out.Items = make([]kanban.BacklogItem, len(rec.Items))
	for i, it := range rec.Items {
		it.PickedBy, it.LeaseExpiresAt = nil, nil
		out.Items[i] = it
	}
	if len(rec.Archived) > 0 {
		out.Archived = make([]kanban.BacklogArchiveEntry, len(rec.Archived))
		for i, entry := range rec.Archived {
			entry.Item.PickedBy, entry.Item.LeaseExpiresAt = nil, nil
			out.Archived[i] = entry
		}
	}
	return &out
}

// todoLeaseCell renders one lease field for the human surface: the stored
// value, or "-" when absent — the same absent-value convention the history
// stamp cells established.
func todoLeaseCell(s *string) string {
	if s == nil || *s == "" {
		return "-"
	}
	return *s
}

// claimStrOr renders a nullable claim field for an output line, falling
// back when absent.
func claimStrOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// todoLeaseCells renders the human-surface lease cells (REQ-TCL-010):
// "by=<holder>\tlease=<expiry>\t" on a card that carries lease fields, and
// "" on one that does not — so a card with no lease keeps its historical
// line shape byte-identically. The cells sit BEFORE the card text, which is
// always the last field (the extension point the history line contract
// names). The JSON surface is a different face entirely: the omitempty
// pointers keep it excluded (REQ-TCL-014).
func todoLeaseCells(it kanban.BacklogItem) string {
	if it.PickedBy == nil && it.LeaseExpiresAt == nil {
		return ""
	}
	return "by=" + todoLeaseCell(it.PickedBy) + "\tlease=" + todoLeaseCell(it.LeaseExpiresAt) + "\t"
}

// todoClaimHolder maps the --lane flag onto the lease holder label: the
// named lane, or the operator's own act when the flag is absent.
func todoClaimHolder(lane string) string {
	if l := strings.TrimSpace(lane); l != "" {
		return l
	}
	return kanban.BacklogOperatorHolder
}

// todoClaimReclaimLines renders the C5 audit surface: one human-readable
// line per card the operation's expiry-first pass returned to queued,
// naming the id and the previous holder. No events table (C5) — the output
// line IS the audit.
func todoClaimReclaimLines(out *strings.Builder, reclaimed []kanban.BacklogReclamation) {
	for _, r := range reclaimed {
		fmt.Fprintf(out, "reclaimed %s prev_holder=%s returned to queued\n", r.ItemID, dash(r.PrevHolder))
	}
}

// todoClaimRefusal maps the claim family's sentinels onto the CLI surface:
// no eligible card is NOT an error — the dedicated no-card exit code (3)
// and a non-error message let a supervising launcher branch on it
// (REQ-TCL-012); every other refusal surfaces its distinct message on
// stderr (REQ-TCL-003/006).
func todoClaimRefusal(cmd *cobra.Command, err error) error {
	if errors.Is(err, kanban.ErrClaimNoCard) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no queued card is available")
		return &exitCodeError{code: factoryNextNoCardExit, msg: "todo claim: no card is available"}
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
	return err
}

// runTodoClaimRoot is the claim body both surfaces share, anchored at root
// (design.md §3). The queue-layout disclosure rides the same entry the read
// verbs use — cross-checkout routing comes from the resolving store path,
// and the stale-store fact is the single detector's (REQ-TSS-001), reused
// from discloseQueueLayout rather than re-probed.
func runTodoClaimRoot(root string, cmd *cobra.Command, lane, renew string) error {
	store := todoStoreAt(root)
	_ = discloseQueueLayout(cmd, "claim")
	out := &strings.Builder{}
	holder := todoClaimHolder(lane)

	if renew != "" {
		id := normalizeTodoRef(renew)
		result, err := store.RenewLease(id, holder)
		if result != nil {
			// The reclamation lines render on BOTH outcomes: a successful
			// renewal's expiry-first pass, and the committed renew-after-
			// expiry return the refusal rides on.
			todoClaimReclaimLines(out, result.Reclaimed)
		}
		if err != nil {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), out.String())
			return todoClaimRefusal(cmd, err)
		}
		fmt.Fprintf(out, "renewed %s lease_expires_at=%s holder=%s\n",
			id, claimStrOr(result.Item.LeaseExpiresAt, "unknown"), holder)
		_, _ = fmt.Fprint(cmd.OutOrStdout(), out.String())
		return nil
	}

	result, err := store.Claim(holder)
	if err != nil {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), out.String())
		return todoClaimRefusal(cmd, err)
	}
	todoClaimReclaimLines(out, result.Reclaimed)
	fmt.Fprintf(out, "claimed %s %s lease_expires_at=%s picked_by=%s\n",
		result.Item.ID, todoTextPrefix(result.Item.Text),
		claimStrOr(result.Item.LeaseExpiresAt, "unknown"), claimStrOr(result.Item.PickedBy, "unknown"))
	_, _ = fmt.Fprint(cmd.OutOrStdout(), out.String())
	recordFactoryCardState(result.Item.ID, "", "picked", "card.assigned")
	return nil
}
