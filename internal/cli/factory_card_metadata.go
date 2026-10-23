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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
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

	// The audit row's kind and field vocabulary (REQ-CMV-010). The decider
	// identity a reclassify records is factory.DeciderIdentityHuman — the
	// closed-set constant, never a re-spelled copy.
	metadataKindSetSpec    = "set-spec"
	metadataKindReclassify = "reclassify"
	metadataFieldSpecID    = "spec_id"
	metadataFieldClassific = "classification"
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
	fdbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		return err
	}
	// REQ-CMV-003 gate (b): the factory record must exist. The record layer
	// creates no row and invents no run identity for a never-assigned card —
	// the operator assigns first, then sets the spec. The existence probe
	// precedes the store open (sync-audit-1 D-B, the accessor convention the
	// lease gate set at SPEC-FACTORY-LEASE-RENEW-001: a refused open leaves
	// no database behind): OpenFactory migrates on open, so refusing after it
	// would materialize an empty store — the database file plus its meta row
	// — as a side effect of a read that found nothing. When the store file is
	// absent no record can exist, so the verb refuses there; a present store
	// is opened as before and the row-level LoadCard gate decides.
	// Only a truly absent store carries the missing-record refusal (sync-audit-2
	// D-G): a stat error with any other cause — an untraversable path component
	// around an existing store included — is not evidence about the record, so
	// the probe falls through and the open's own error surfaces the real cause.
	// The probe also precedes the RUN resolution (card-review CR1): the shared
	// resolver's omitted form opens the store to list active runs — an open
	// that CREATES it — and, on several active runs, reconciles (a runs-table
	// write), so a verb that resolved first could mutate the store on a
	// refused path.
	if _, serr := os.Stat(fdbPath); errors.Is(serr, fs.ErrNotExist) {
		return fmt.Errorf("card %s has no factory record — assign it first (the cards table is keyed (run_id, card_id)); set-spec creates no record", cardID)
	}
	runID, err := resolveSetSpecRun(ctx, fdbPath, run)
	if err != nil {
		return err
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
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

// resolveSetSpecRun is the metadata verb's run resolution (card-review CR1).
// An explicit --run is used as given — the verb's own gates (queue-picked,
// the record row keyed (run_id, card_id)) are its validation, matching the
// shared resolver's explicit form. The omitted form resolves through the
// read-only active-run listing (factorymsg.ActiveRunIDsAt), which never
// creates the store and never retires runs — a validation-phase lookup must
// not mutate the store it validates against — and surviving ambiguity is
// real ambiguity: fail closed, pass --run.
func resolveSetSpecRun(ctx context.Context, fdbPath, explicit string) (string, error) {
	if run := strings.TrimSpace(explicit); run != "" {
		return run, nil
	}
	ids, err := factorymsg.ActiveRunIDsAt(ctx, fdbPath)
	if err != nil {
		return "", err
	}
	switch len(ids) {
	case 1:
		return ids[0], nil
	case 0:
		return "", fmt.Errorf("no --run given and no active factory run")
	default:
		return "", fmt.Errorf("no --run given and %d active factory runs; pass --run", len(ids))
	}
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
			switch it.State {
			case factory.BacklogStatePicked:
			default:
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
	// item is the REQ-CMV-004 direction-6 converged-state audit completion —
	// whose given is BOTH stores carrying the commanded value, so the record
	// is re-read under the lock before the branch is taken (sync-audit-1 B1):
	// a queue-converged/record-stale state (the assign-without-spec blank
	// this SPEC §A exists to close) is not direction 6 — it falls through to
	// the write path below, which converges the record and lands the
	// transition row anchored on the record's true prior.
	auditOld := current
	if current == spec {
		recordNow, err := db.LoadCard(ctx, runID, cardID)
		if errors.Is(err, homestate.ErrCardNotFound) {
			return fmt.Errorf("card %s has no factory record — assign it first (the cards table is keyed (run_id, card_id)); set-spec creates no record", cardID)
		}
		if err != nil {
			return err
		}
		if recordPrior := strings.TrimSpace(recordNow.SpecID); recordPrior == spec {
			if audited && latest == spec {
				return fmt.Errorf("card %s already carries spec id %q and the audit surface's latest row already records it (converged-and-audited duplicate)", cardID, spec)
			}
			old := ""
			if audited {
				old = latest
			}
			// The catch-up append carries the write path's append discipline
			// (sync-audit-1 D-A): one retry, then the audit-loss sentinel —
			// no store was mutated either way, so a re-run heals by re-running.
			for attempt := 0; attempt < 2; attempt++ {
				if aerr := metadataFault("audit-append"); aerr == nil {
					if aerr = db.RecordCardMetadataCatchUp(ctx, homestate.CardMetadataEvent{
						CardID: cardID, Kind: metadataKindSetSpec, Actor: actor, Session: session,
						Field: metadataFieldSpecID, Old: old, New: spec, Reason: reason,
					}, now); aerr == nil {
						_, _ = fmt.Fprintf(out, "%s spec %s (attested; both stores already converged, no store re-mutated)\n", cardID, spec)
						return nil
					}
				}
			}
			return fmt.Errorf("%s on %s (set-spec catch-up → %q): both stores already carried the value and neither was re-mutated, but the audit append failed after one retry; the missing attestation is surfaced — re-run set-spec to land it",
				metadataAuditLossSentinel, cardID, spec)
		}
		// The record is stale: the write path below runs on the re-read state
		// — the version fence and the compensating restore anchor on what the
		// lock actually observed — and the transition row's old value is the
		// record's prior, the store that actually moves.
		card = recordNow
		auditOld = strings.TrimSpace(recordNow.SpecID)
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
				Field: metadataFieldSpecID, Old: auditOld, New: spec, Reason: reason,
			}, now); aerr == nil {
				_, _ = fmt.Fprintf(out, "%s spec %s (was %s)\n", cardID, spec, dash(auditOld))
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
			switch rec.Items[i].State {
			case factory.BacklogStatePicked:
			default:
				return fmt.Errorf("queue item %s is %s, not picked", cardID, rec.Items[i].State)
			}
			value := spec
			rec.Items[i].SpecID = &value
			return nil
		}
		return fmt.Errorf("no backlog item %s", cardID)
	})
}

// metadataAxesJSON renders the classification's VALUE axes — blocked, mode,
// priority — as the canonical JSON the audit row's old/new carry and the
// LATEST-row duplicate test compares. The decider, stamp, and reason are
// payload fields, never part of the compared value: a re-run's stamp differs
// by definition, and the REQ-CMV-010 duplicate test is about the value the
// queue item carries.
func metadataAxesJSON(c factory.CardClassification) string {
	data, err := json.Marshal(struct {
		Blocked  bool   `json:"blocked"`
		Mode     string `json:"mode"`
		Priority string `json:"priority"`
	}{Blocked: c.Blocked, Mode: c.Mode, Priority: c.Priority})
	if err != nil {
		return ""
	}
	return string(data)
}

// newFactoryReclassifyCommand builds `moai factory reclassify <card>
// [--mode <m>] [--priority <p>] --reason <text>` — the post-issuance
// re-designation verb. Its write target is the queue item's classification
// record (REQ-CMV-009): no Card-row write and no run identity — a
// queued-only card takes the queue-only path, and only the audit append
// touches homestate.
func newFactoryReclassifyCommand() *cobra.Command {
	var mode, priority, reason string
	cmd := &cobra.Command{
		Use:   "reclassify <card>",
		Short: "Re-designate a queued or picked card's mode and/or priority (audit-trailed; gated by " + metadataFlagKey + ")",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runFactoryReclassify(cmd, args[0], mode, priority, reason); err != nil {
				return fmt.Errorf("factory reclassify: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "",
		"execution mode: "+factory.ClassModeSerial+" or "+factory.ClassModeParallelizable)
	cmd.Flags().StringVar(&priority, "priority", "",
		"priority: "+factory.ClassPriorityHigh+", "+factory.ClassPriorityNormal+", or "+factory.ClassPriorityLow)
	cmd.Flags().StringVar(&reason, "reason", "",
		"why the classification is changing (required, 1-500 runes; recorded verbatim in the audit row)")
	return cmd
}

func runFactoryReclassify(cmd *cobra.Command, cardID, mode, priority, reason string) error {
	if !homestate.ValidCardID(cardID) {
		return fmt.Errorf("%q is not a card id", cardID)
	}
	// REQ-CMV-015: leader/operator acts; the lane boundary ahead of the gate
	// order, as in set-spec.
	if factoryLaneRefusal() {
		return fmt.Errorf("refused — %s: the metadata verbs are the leader's or the operator's act (%s=%s marks a lane)",
			factoryLaneBoundarySentinel, config.EnvFactoryRole, config.FactoryRoleLane)
	}
	root := factoryCardRoot()
	// REQ-CMV-001: the flag gate.
	if !cardMetadataVerbsEnabled(root) {
		return fmt.Errorf("refused — this verb is gated by %s (default false); set it in .moai/config/sections/workflow.yaml to admit card metadata edits", metadataFlagKey)
	}
	// REQ-CMV-012: the reason gate.
	if err := validateMetadataReason(reason); err != nil {
		return err
	}
	// REQ-CMV-007 gate (c): at least one axis.
	if !cmd.Flags().Changed("mode") && !cmd.Flags().Changed("priority") {
		return fmt.Errorf("supply at least one of --mode or --priority")
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	// The homestate database is opened for the audit surface only — the
	// classification write itself is a queue-layer write (REQ-CMV-009) — and
	// it opens INSIDE the lock, after the queue-state and classification
	// validation has admitted the input (card-review CR2-2, the CR1 ordering
	// mirrored on this verb): an open CREATES the store, so a refusable
	// input must never reach it.
	actor, session := metadataActor()
	now := factoryCardNow()
	out := cmd.OutOrStdout()
	return newTodoStore().WithLock(func(l *factory.LockedBacklog) error {
		return reclassifyLocked(ctx, l, root, cardID, mode, priority, reason, cmd, actor, session, now, out)
	})
}

// reclassifyLocked is runFactoryReclassify's body under the held queue lock.
func reclassifyLocked(ctx context.Context, l *factory.LockedBacklog, root, cardID, mode, priority, reason string, cmd *cobra.Command, actor, session string, now time.Time, out io.Writer) error {
	rec, err := l.LoadPure()
	if err != nil {
		return fmt.Errorf("read queue: %w", err)
	}
	var item *factory.BacklogItem
	for i := range rec.Items {
		if rec.Items[i].ID == cardID {
			item = &rec.Items[i]
			break
		}
	}
	if item == nil {
		return fmt.Errorf("queue item %s is not in the queue", cardID)
	}
	// REQ-CMV-007 gate (a): POSITIVE enumeration — the states the verb
	// admits are named; every other state (a state added later included)
	// refuses.
	switch item.State {
	case factory.BacklogStateQueued, factory.BacklogStatePicked:
		// the admitted states
	default:
		return fmt.Errorf("queue item %s is %s, not queued or picked", cardID, item.State)
	}
	// The commanded classification: the read-side default seam derives any
	// absent axis, the supplied axes override, and the blocked axis is never
	// altered (REQ-CMV-008). Decider is human — a reclassify is a human
	// re-judgment (REQ-CMV-009).
	base := factory.EffectiveCardClassification(*item)
	oldAxes := metadataAxesJSON(base)
	target := base
	if cmd.Flags().Changed("mode") {
		target.Mode = mode
	}
	if cmd.Flags().Changed("priority") {
		target.Priority = priority
	}
	target.Decider = factory.DeciderIdentityHuman
	// REQ-CMV-007 gate (b): the closed sets through the EXISTING validator —
	// no re-spelled copy (classification.go's own header rule).
	if err := factory.ValidateCardClassification(target); err != nil {
		return err
	}
	commanded := metadataAxesJSON(target)
	// The store opens only for an input the validation admitted (card-review
	// CR2-2): every refusal above is queue-side or validator-side, so a
	// project without a factory database is not materialized by a refusal.
	// The catch-up and append paths below need the audit surface; creating
	// the store there is the verb doing its write, never a side effect of a
	// read that found nothing.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	// REQ-CMV-010: the LATEST-row duplicate test ahead of the write. The
	// queue item already carrying the commanded axes splits two ways: a
	// converged-AND-audited duplicate refuses, a converged-but-unaudited
	// item is the converged-state audit completion — exactly the row the
	// crashed run owed, no store re-mutation, exit zero.
	if item.Classification != nil && metadataAxesJSON(*item.Classification) == commanded {
		latest, audited, err := db.LatestCardMetadataNewValue(ctx, cardID, metadataKindReclassify, metadataFieldClassific)
		if err != nil {
			return err
		}
		if audited && latest == commanded {
			return fmt.Errorf("card %s already carries the commanded classification and the audit surface's latest row already records it (converged-and-audited duplicate)", cardID)
		}
		old := ""
		if audited {
			old = latest
		}
		// The catch-up append carries the write path's append discipline
		// (sync-audit-1 D-A): one retry, then the audit-loss sentinel — the
		// queue write stands exactly as the crashed run committed it either
		// way, and a re-run heals by re-running.
		for attempt := 0; attempt < 2; attempt++ {
			if aerr := metadataFault("audit-append"); aerr == nil {
				if aerr = db.RecordCardMetadataCatchUp(ctx, homestate.CardMetadataEvent{
					CardID: cardID, Kind: metadataKindReclassify, Actor: actor, Session: session,
					Field: metadataFieldClassific, Old: old, New: commanded, Reason: reason,
				}, now); aerr == nil {
					_, _ = fmt.Fprintf(out, "%s classification %s (attested; the queue item already carried it, no store re-mutated)\n", cardID, commanded)
					return nil
				}
			}
		}
		return fmt.Errorf("%s on %s (reclassify catch-up → %q): the queue item already carried the axes and nothing was re-mutated, but the audit append failed after one retry; the missing attestation is surfaced — re-run reclassify to land it",
			metadataAuditLossSentinel, cardID, commanded)
	}
	// The write: fresh classified-at stamp, the caller-supplied reason,
	// decider human — inside the caller's locked Mutate. Absence marshals
	// byte-identically: a card classified for the first time here ACQUIRES
	// the field, never a placeholder (REQ-CMV-009's additive discipline).
	stamp := now.UTC().Format(time.RFC3339Nano)
	target.ClassifiedAt = stamp
	target.Reason = reason
	if err := l.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != cardID {
				continue
			}
			switch rec.Items[i].State {
			case factory.BacklogStateQueued, factory.BacklogStatePicked:
				// re-checked under the lock — a pick racing the read lands
				// the classification on the post-pick item (acceptance.md §C)
			default:
				return fmt.Errorf("queue item %s is %s, not queued or picked", cardID, rec.Items[i].State)
			}
			value := target
			rec.Items[i].Classification = &value
			// The reclassify is a classification write that may change the
			// rank — the same locked write re-establishes the queue's
			// classification order (REQ-TCD-005), as every add does
			// (backlog_store.go's append). Stable sort: a rank-preserving
			// write re-sorts to the order it found. Selection reads stored
			// order as priority order, so an unsorted queue makes every
			// position-reading consumer lie (card-review CR2-1).
			rec.SortByClassification()
			return nil
		}
		return fmt.Errorf("no backlog item %s", cardID)
	}); err != nil {
		return err
	}
	// The audit append only after the queue write commits — one retry, then
	// the audit-loss sentinel with the completed queue write left in place
	// (the single-store mirror of REQ-CMV-004 direction 4).
	for attempt := 0; attempt < 2; attempt++ {
		if aerr := metadataFault("audit-append"); aerr == nil {
			if aerr = db.RecordCardMetadataEvent(ctx, homestate.CardMetadataEvent{
				CardID: cardID, Kind: metadataKindReclassify, Actor: actor, Session: session,
				Field: metadataFieldClassific, Old: oldAxes, New: commanded, Reason: reason, ClassifiedAt: stamp,
			}, now); aerr == nil {
				_, _ = fmt.Fprintf(out, "%s classification %s (was %s)\n", cardID, commanded, oldAxes)
				return nil
			}
		}
	}
	return fmt.Errorf("%s on %s (reclassify → %q): the queue write committed but the audit append failed after one retry; the missing row is surfaced — re-run reclassify to land the catch-up attestation",
		metadataAuditLossSentinel, cardID, commanded)
}
