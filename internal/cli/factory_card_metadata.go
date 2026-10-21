// factory_card_metadata.go — SPEC-CARD-METADATA-VERBS-001: the
// card-metadata editing verbs `moai factory set-spec` (M2) and `moai
// factory reclassify` (M3), gated by workflow.card_metadata_verbs.enabled
// (default OFF, REQ-CMV-001/002) and refused in lane sessions (REQ-CMV-015).
//
// The write sequence every verb follows is the honest boundary of WithLock
// (a lock, not a transaction — backlog_store.go): factory-record write
// first, queue write second, audit append only after both stores commit,
// all inside the queue lock (REQ-CMV-004). No cross-database
// transactionality is claimed or built; the failure matrix — compensating
// restore, surfaced divergence, audit-loss sentinel, re-run heal, and the
// direction-6 converged-state catch-up — is the rule set for every
// direction in which the stores or the audit surface can diverge.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	// metadataFlagKey names the capability gate in every flag-off refusal
	// (REQ-CMV-001: the refusal names the flag key).
	metadataFlagKey = "workflow.card_metadata_verbs.enabled"
	// metadataDivergenceSentinel marks the record/queue divergence the
	// failure matrix surfaces for operator repair (REQ-CMV-004 directions
	// 2 and 3) — this SPEC builds no cross-database transactionality.
	metadataDivergenceSentinel = "card metadata divergence"
	// metadataAuditLossSentinel marks a mutation whose store writes both
	// committed but whose audit row could not be appended after one retry
	// (REQ-CMV-004 direction 4, REQ-CMV-010) — a surfaced missing row,
	// never a silent success.
	metadataAuditLossSentinel = "card metadata audit row lost"
	// metadataReasonMaxRunes is the reason bound (REQ-CMV-012), counted in
	// runes, not bytes.
	metadataReasonMaxRunes = 500

	// The audit row's kind and field vocabulary (REQ-CMV-010).
	metadataKindSetSpec     = "set-spec"
	metadataKindReclassify  = "reclassify"
	metadataFieldSpecID     = "spec_id"
	metadataFieldClassific  = "classification"
	metadataDeciderIdentity = "human"
)

// metadataWriteFault is the write-sequence test seam (the shape of homestate's
// cardTransitionFault): when set, it is called at the named points of the
// metadata write sequence and a non-nil return aborts that step. Stages:
// "record-set" (before the factory-record write), "queue-write" (before the
// queue write), "record-restore" (before the compensating restore), and
// "audit-append" (before each append attempt — the single retry calls it
// twice).
var metadataWriteFault func(stage string) error

func metadataFault(stage string) error {
	if metadataWriteFault == nil {
		return nil
	}
	return metadataWriteFault(stage)
}

// cardMetadataVerbsEnabled reads the capability gate (REQ-CMV-001). A load
// failure reads as disabled — the verbs are opt-in, never fail-open.
func cardMetadataVerbsEnabled(root string) bool {
	cfg, err := config.NewLoader().Load(filepath.Join(root, ".moai"))
	if err != nil {
		return false
	}
	return cfg.Workflow.CardMetadataVerbs.Enabled
}

// metadataActor is the audit row's capture-layer actor resolution
// (REQ-CMV-011): the lane label from the lane marker environment variable
// when set, otherwise operator — the capture pattern of `factory decide`
// (factory_audit_decide.go). The session UUID rides the existing
// session-identity capture; an empty result is carried as empty (the row
// records that no identity was resolvable, never a placeholder).
func metadataActor() (actor, session string) {
	if lane := strings.TrimSpace(os.Getenv(config.EnvMoaiFactoryWorker)); lane != "" {
		return lane, integrationSessionID("")
	}
	return "operator", integrationSessionID("")
}

// validateMetadataReason is the REQ-CMV-012 gate: the caller-supplied reason
// is required, may not be whitespace-only, and is bounded at
// metadataReasonMaxRunes runes.
func validateMetadataReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("--reason is required: the audit row records why the metadata changed (1-%d runes)", metadataReasonMaxRunes)
	}
	if n := utf8.RuneCountInString(reason); n > metadataReasonMaxRunes {
		return fmt.Errorf("--reason is %d runes; the bound is %d", n, metadataReasonMaxRunes)
	}
	return nil
}

// newFactorySetSpecCommand builds `moai factory set-spec <card> --spec <id>
// --reason <text>` — the post-pick spec-id setter the unpick/re-pick round
// trip used to be the only route to (spec.md §A).
func newFactorySetSpecCommand() *cobra.Command {
	var spec, reason, run string
	cmd := &cobra.Command{
		Use:   "set-spec <card>",
		Short: "Set a picked card's SPEC identifier after pick (audit-trailed; gated by " + metadataFlagKey + ")",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runFactorySetSpec(cmd, args[0], strings.TrimSpace(spec), reason, run); err != nil {
				return fmt.Errorf("factory set-spec: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&spec, "spec", "",
		"SPEC identifier in the strict form SPEC-<DOMAIN>-<NNN> (letter-leading uppercase middle segments, three-digit tail)")
	cmd.Flags().StringVar(&reason, "reason", "",
		"why the spec id is being set (required, 1-500 runes; recorded verbatim in the audit row)")
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

func runFactorySetSpec(cmd *cobra.Command, cardID, spec, reason, run string) error {
	if !homestate.ValidCardID(cardID) {
		return fmt.Errorf("%q is not a card id", cardID)
	}
	// REQ-CMV-015 (decision-index Q2, Default applied): the metadata verbs
	// are leader/operator acts. The lane boundary is checked ahead of the
	// gate order, per the refusal-family conventions (plan.md §F M4).
	if factoryLaneRefusal() {
		return fmt.Errorf("refused — %s: the metadata verbs are the leader's or the operator's act (%s=%s marks a lane)",
			factoryLaneBoundarySentinel, config.EnvFactoryRole, config.FactoryRoleLane)
	}
	root := factoryCardRoot()
	// REQ-CMV-001: the flag gate. Writing neither store and no audit row.
	if !cardMetadataVerbsEnabled(root) {
		return fmt.Errorf("refused — this verb is gated by %s (default false); set it in .moai/config/sections/workflow.yaml to admit card metadata edits", metadataFlagKey)
	}
	// REQ-CMV-012: the reason gate.
	if err := validateMetadataReason(reason); err != nil {
		return err
	}
	// REQ-CMV-003 gate (c): the strict grammar — the value is trimmed before
	// every comparison (acceptance.md §C), and the refusal quotes the
	// rejected value.
	if !homestate.ValidSetSpecID(spec) {
		return fmt.Errorf("%q is not a SPEC identifier of the required form SPEC-<DOMAIN>-<NNN> (letter-leading uppercase middle segments, three-digit tail)", spec)
	}
	// REQ-CMV-003 gate (a): the queue item is picked (the requireQueuePicked
	// family — POSITIVE enumeration, REQ-FR-022).
	if err := requireQueuePicked(root, cardID); err != nil {
		return err
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	runID, err := resolveFactoryCardRun(ctx, root, run)
	if err != nil {
		return err
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	// REQ-CMV-003 gate (b): the factory record must exist. The record layer
	// creates no row and invents no run identity for a never-assigned card —
	// the operator assigns first, then sets the spec.
	card, err := db.LoadCard(ctx, runID, cardID)
	if errors.Is(err, homestate.ErrCardNotFound) {
		return fmt.Errorf("card %s has no factory record — assign it first (the cards table is keyed (run_id, card_id)); set-spec creates no record", cardID)
	}
	if err != nil {
		return err
	}
	actor, session := metadataActor()
	now := factoryCardNow()
	out := cmd.OutOrStdout()
	// Everything below runs inside the queue lock: the gates re-read the
	// queue authoritatively, and the write sequence — record first, queue
	// second, audit append after both commit — is serialized against every
	// other queue writer.
	return newTodoStore().WithLock(func(l *factory.LockedBacklog) error {
		return setSpecLocked(ctx, l, db, runID, cardID, card, spec, reason, actor, session, now, out)
	})
}

// setSpecLocked is runFactorySetSpec's body under the held queue lock.
func setSpecLocked(ctx context.Context, l *factory.LockedBacklog, db *homestate.FactoryDB, runID, cardID string, card homestate.Card, spec, reason, actor, session string, now time.Time, out io.Writer) error {
	rec, err := l.LoadPure()
	if err != nil {
		return fmt.Errorf("read queue: %w", err)
	}
	var current string
	found := false
	for _, it := range rec.Items {
		if it.ID == cardID {
			found = true
			if it.SpecID != nil {
				current = strings.TrimSpace(*it.SpecID)
			}
			if it.State != factory.BacklogStatePicked {
				return fmt.Errorf("queue item %s is %s, not picked", cardID, it.State)
			}
			break
		}
	}
	if !found {
		return fmt.Errorf("queue item %s is not in the queue", cardID)
	}
	latest, audited, err := db.LatestCardMetadataNewValue(ctx, cardID, metadataKindSetSpec, metadataFieldSpecID)
	if err != nil {
		return err
	}
	// REQ-CMV-003 gate (d): the no-change gate reads the QUEUE item — the
	// store cardToSpecID reads — never the factory record alone, so a
	// direction-5 divergence is admitted for convergence rather than refused.
	// A converged-AND-audited duplicate refuses; a converged-but-unaudited
	// item is the REQ-CMV-004 direction-6 converged-state audit completion.
	if current == spec {
		if audited && latest == spec {
			return fmt.Errorf("card %s already carries spec id %q and the audit surface's latest row already records it (converged-and-audited duplicate)", cardID, spec)
		}
		old := ""
		if audited {
			old = latest
		}
		if err := db.RecordCardMetadataCatchUp(ctx, homestate.CardMetadataEvent{
			CardID: cardID, Kind: metadataKindSetSpec, Actor: actor, Session: session,
			Field: metadataFieldSpecID, Old: old, New: spec, Reason: reason,
		}, now); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "%s spec %s (attested; both stores already converged, no store re-mutated)\n", cardID, spec)
		return nil
	}
	// Write sequence (REQ-CMV-004): the factory-record write first.
	if err := metadataFault("record-set"); err != nil {
		return err
	}
	written, err := db.SetCardSpec(ctx, runID, cardID, spec, card.Version, now)
	if err != nil {
		return err
	}
	// The queue write second. A failure here leaves the record ahead of the
	// queue: the compensating restore runs through the named M1 entry point,
	// and the divergence is surfaced either way (directions 2 and 3).
	queueErr := metadataFault("queue-write")
	if queueErr == nil {
		queueErr = writeQueueSpec(l, cardID, spec)
	}
	if queueErr != nil {
		if metadataFault("record-restore") == nil {
			if _, rerr := db.RestoreCardSpec(ctx, runID, cardID, card.SpecID, written.Version, now); rerr != nil {
				// Direction 3: the compensating write itself failed — the
				// divergence stands, surfaced with card, prior, attempted.
				return fmt.Errorf("%s on %s: the queue write failed (%v) and the compensating restore failed (%v); the record carries %q, the queue keeps %q, attempted %q",
					metadataDivergenceSentinel, cardID, queueErr, rerr, spec, dash(card.SpecID), spec)
			}
		}
		// Direction 2: restored, still non-zero with the divergence sentinel.
		return fmt.Errorf("%s on %s: the queue write failed after the record commit; the record was restored to %q (attempted %q)",
			metadataDivergenceSentinel, cardID, card.SpecID, spec)
	}
	// The audit append only after both stores commit — one retry, then the
	// audit-loss sentinel with the completed store writes left in place.
	for attempt := 0; attempt < 2; attempt++ {
		if aerr := metadataFault("audit-append"); aerr == nil {
			if aerr = db.RecordCardMetadataEvent(ctx, homestate.CardMetadataEvent{
				CardID: cardID, Kind: metadataKindSetSpec, Actor: actor, Session: session,
				Field: metadataFieldSpecID, Old: current, New: spec, Reason: reason,
			}, now); aerr == nil {
				_, _ = fmt.Fprintf(out, "%s spec %s (was %s)\n", cardID, spec, dash(current))
				return nil
			}
		}
	}
	return fmt.Errorf("%s on %s (set-spec → %q): both stores committed but the audit append failed after one retry; the missing row is surfaced — re-run set-spec to land the catch-up attestation",
		metadataAuditLossSentinel, cardID, spec)
}

// writeQueueSpec lands the spec id on the queue item inside the caller's
// locked Mutate — never a direct engine write (plan.md §D constraint 7).
func writeQueueSpec(l *factory.LockedBacklog, cardID, spec string) error {
	return l.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != cardID {
				continue
			}
			if rec.Items[i].State != factory.BacklogStatePicked {
				return fmt.Errorf("queue item %s is %s, not picked", cardID, rec.Items[i].State)
			}
			value := spec
			rec.Items[i].SpecID = &value
			return nil
		}
		return fmt.Errorf("no backlog item %s", cardID)
	})
}
