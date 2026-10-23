// card_metadata.go — SPEC-CARD-METADATA-VERBS-001 M1: the card-keyed
// metadata audit surface and the set-spec record-layer mutation function.
//
// The audit surface is a NEW card_metadata_events table (decision-index Q3,
// Default applied 2026-10-17) reusing the shape of the run-keyed events
// table — autoincrement seq, kind, JSON payload, created-at — keyed by the
// card id the data actually has, added by the additive CREATE TABLE IF NOT
// EXISTS in factoryDDL (REQ-CMV-014: row-preserving, no existing row
// rewritten, migrated, or deleted).
//
// SetCardSpec is the record-layer mutation function REQ-CMV-005 defines: it
// lands spec_id at every card state from picked onward (every state
// CardStates() enumerates), superseding RecordPicked's while-picked-only
// field rule for this mutation function alone. The append timing rule is
// load-bearing and lives in the verb, not here: an audit row is appended
// only AFTER both stores have committed (REQ-CMV-004), so nothing in this
// file appends inside the mutation function or inside its homestate
// transaction — the card.fields event path in card_picked.go is the
// precedent for the versioned-write MECHANICS (withCardTx + updateCardRow +
// version fencing) only, never for append timing.
package homestate

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// cardMetadataSpecPattern is the strict SPEC-id grammar set-spec enforces
// (plan.md §D constraint 4): literal SPEC prefix, one or more letter-leading
// uppercase-alphanumeric middle segments, a digit-only three-digit tail. It
// sits BESIDE specIDPattern and never mutates it — pick-time acceptance
// keeps the looser historical pattern (spec.md §F).
var cardMetadataSpecPattern = regexp.MustCompile(`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`)

// ValidSetSpecID reports whether s matches the strict set-spec grammar.
func ValidSetSpecID(s string) bool { return cardMetadataSpecPattern.MatchString(s) }

// CardMetadataEvent is one audit row of the card_metadata_events surface:
// who changed what, from what, to what, and why (REQ-CMV-010).
type CardMetadataEvent struct {
	// CardID is the card the mutation landed on; rows are keyed by the card
	// id the data actually has, never by a run id (a queued card has none).
	CardID string
	// Kind names the verb: "set-spec" or "reclassify".
	Kind string
	// Actor is the lane label or "operator" (REQ-CMV-011); Session carries
	// the session UUID when one is resolvable.
	Actor   string
	Session string
	// Field names the mutated field: "spec_id" or "classification".
	Field string
	// Old and New carry the value transition. Old is empty where the prior
	// value is unknown (a value that arrived at pick time and was never
	// audited is honestly recorded as unknown-old, never fabricated).
	Old string
	New string
	// Reason is the caller-supplied reason, verbatim.
	Reason string
	// ClassifiedAt is the classification stamp a reclassify write used
	// (empty on set-spec rows and on catch-up attestations).
	ClassifiedAt string
}

// cardMetadataPayload is the JSON payload column's shape.
type cardMetadataPayload struct {
	Actor        string `json:"actor"`
	Session      string `json:"session,omitempty"`
	Field        string `json:"field"`
	Old          string `json:"old"`
	New          string `json:"new"`
	Reason       string `json:"reason"`
	ClassifiedAt string `json:"classified_at,omitempty"`
}

// RecordCardMetadataEvent appends exactly one audit row after both stores
// have committed (REQ-CMV-004 / REQ-CMV-010). It is the after-both-commits
// append entry point; the verb invokes it outside the mutation function.
func (f *FactoryDB) RecordCardMetadataEvent(ctx context.Context, ev CardMetadataEvent, now time.Time) error {
	return f.appendCardMetadataRow(ctx, ev, now)
}

// RecordCardMetadataCatchUp appends the direction-6 converged-state
// attestation row: the row a crashed run owed, appended by the admitted
// re-run. It mutates no store — the admission ruling and the store
// row-unchanged guarantee live in the verb; this entry point is the append,
// kept separate so the two append surfaces stay nameable apart (plan.md §F
// M1 GREEN).
func (f *FactoryDB) RecordCardMetadataCatchUp(ctx context.Context, ev CardMetadataEvent, now time.Time) error {
	return f.appendCardMetadataRow(ctx, ev, now)
}

func (f *FactoryDB) appendCardMetadataRow(ctx context.Context, ev CardMetadataEvent, now time.Time) error {
	if strings.TrimSpace(ev.CardID) == "" || strings.TrimSpace(ev.Kind) == "" || strings.TrimSpace(ev.Field) == "" {
		return fmt.Errorf("%w: card id, kind, and field are required for a metadata audit row", ErrInvalidCardInput)
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	payload, err := json.Marshal(cardMetadataPayload{
		Actor: ev.Actor, Session: ev.Session, Field: ev.Field,
		Old: ev.Old, New: ev.New, Reason: ev.Reason, ClassifiedAt: ev.ClassifiedAt,
	})
	if err != nil {
		return err
	}
	return retryFactoryBusy(ctx, func() error {
		// SQL: every value goes through a ? placeholder.
		_, err := f.DB.ExecContext(ctx, `INSERT INTO card_metadata_events(card_id,kind,payload_json,created_at) VALUES(?,?,?,?)`,
			ev.CardID, ev.Kind, string(payload), now.Format(time.RFC3339Nano))
		return err
	})
}

// LatestCardMetadataNewValue returns the new value the field's LATEST audit
// row records for the card and kind — the value the REQ-CMV-003 gate (d)
// duplicate test compares against, and the true prior a catch-up row
// records. The existence of an OLDER row whose new value equals a commanded
// value does not make a duplicate: only the LATEST row decides (spec.md
// REQ-CMV-003 gate (d)). found is false, with an empty value, when no row
// exists for the card, kind, and field.
func (f *FactoryDB) LatestCardMetadataNewValue(ctx context.Context, cardID, kind, field string) (string, bool, error) {
	// SQL: every value goes through a ? placeholder.
	rows, err := f.DB.QueryContext(ctx, `SELECT payload_json FROM card_metadata_events WHERE card_id=? AND kind=? ORDER BY seq DESC`, cardID, kind)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return "", false, err
		}
		var payload cardMetadataPayload
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return "", false, fmt.Errorf("decode metadata payload %q: %w", truncateCardMetadata(raw), err)
		}
		if payload.Field != field {
			continue
		}
		return payload.New, true, nil
	}
	return "", false, rows.Err()
}

// truncateCardMetadata bounds a corrupt payload quoted in an error.
func truncateCardMetadata(s string) string {
	if len(s) <= 64 {
		return s
	}
	return s[:64] + "..."
}

// SetCardSpec is the record-layer spec-id mutation REQ-CMV-005 defines. It
// lands specID on the card row at every state from picked onward — every
// state CardStates() enumerates, terminal states included (reachable at this
// layer only; the verb's admission stays queue-picked-only) — bumping the
// row's version by exactly one and refusing a stale expectedVersion with
// ErrStaleVersion. It writes the card row ONLY: the audit append happens
// after both stores commit, through the separate entry points above, never
// inside this function or its transaction. A missing row is ErrCardNotFound;
// nothing is created and no run identity is invented (REQ-CMV-003 gate (b)).
//
// @MX:ANCHOR: [AUTO] SetCardSpec — the sole post-pick spec-id record-layer mutation (REQ-CMV-005)
// @MX:REASON: the set-spec verb and its direction-2 compensating restore both route through this one version-fenced body; a second mutation path would fork the fencing contract and re-open the one-directional gap the failure matrix exists to close
// @MX:SPEC: SPEC-CARD-METADATA-VERBS-001
func (f *FactoryDB) SetCardSpec(ctx context.Context, runID, cardID, specID string, expectedVersion int64, now time.Time) (Card, error) {
	return f.mutateCardSpec(ctx, runID, cardID, specID, expectedVersion, now, false)
}

// RestoreCardSpec is the compensating-restore entry point REQ-CMV-004
// direction 2 requires: the same version-fenced mutation, restoring the
// prior spec_id after a queue-side failure left the record ahead of the
// queue. It is a named function rather than a second inline write so the
// failure-matrix tests inject against the same seam the verb calls.
func (f *FactoryDB) RestoreCardSpec(ctx context.Context, runID, cardID, priorSpecID string, expectedVersion int64, now time.Time) (Card, error) {
	return f.mutateCardSpec(ctx, runID, cardID, priorSpecID, expectedVersion, now, true)
}

// mutateCardSpec is the shared version-fenced body of SetCardSpec and
// RestoreCardSpec. allowEmpty admits the empty prior value the compensating
// restore legitimately lands (a card whose spec id was blank before the
// failed write — the blank-spec incident shape this SPEC exists to close);
// the forward SetCardSpec path keeps refusing it.
func (f *FactoryDB) mutateCardSpec(ctx context.Context, runID, cardID, specID string, expectedVersion int64, now time.Time, allowEmpty bool) (Card, error) {
	if strings.TrimSpace(runID) == "" || !ValidCardID(cardID) {
		return Card{}, fmt.Errorf("%w: run id and a valid card id are required", ErrInvalidCardInput)
	}
	value := strings.TrimSpace(specID)
	if value == "" && !allowEmpty {
		return Card{}, fmt.Errorf("%w: spec id is required", ErrInvalidCardInput)
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	var result Card
	err := f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, runID, cardID)
		if err != nil {
			return nil, err
		}
		if cur.Legacy() {
			return nil, fmt.Errorf("%w: card %s holds pre-F1 state %q; its fields change through the F1 transitions", ErrLegacyState, cur.CardID, cur.State)
		}
		// The version compare comes first, as in every fenced write: a
		// caller whose read is stale has no standing to be told anything else.
		if cur.Version != expectedVersion {
			return nil, staleErr(cur, expectedVersion)
		}
		next := cur
		next.SpecID = value
		next.Version = cur.Version + 1
		next.UpdatedAt = now.Format(time.RFC3339Nano)
		if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
			return nil, err
		}
		result = next
		return nil, nil
	})
	if err != nil {
		return Card{}, err
	}
	return result, nil
}

// mutateCardColumns is the shared version-fenced body of the t1930 metadata
// mutations (card t1930): the same withCardTx + updateCardRow + stale-version
// discipline mutateCardSpec carries, with the field changes themselves passed
// as an apply function so each verb keeps a named record-layer entry point.
// The audit append timing rule is unchanged: nothing here appends; the verb
// appends only after the store commit.
func (f *FactoryDB) mutateCardColumns(ctx context.Context, runID, cardID string, expectedVersion int64, now time.Time, apply func(*Card)) (Card, error) {
	if strings.TrimSpace(runID) == "" || !ValidCardID(cardID) {
		return Card{}, fmt.Errorf("%w: run id and a valid card id are required", ErrInvalidCardInput)
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	var result Card
	err := f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, runID, cardID)
		if err != nil {
			return nil, err
		}
		if cur.Legacy() {
			return nil, fmt.Errorf("%w: card %s holds pre-F1 state %q; its fields change through the F1 transitions", ErrLegacyState, cur.CardID, cur.State)
		}
		// The version compare comes first, as in every fenced write: a
		// caller whose read is stale has no standing to be told anything else.
		if cur.Version != expectedVersion {
			return nil, staleErr(cur, expectedVersion)
		}
		next := cur
		apply(&next)
		next.Version = cur.Version + 1
		next.UpdatedAt = now.Format(time.RFC3339Nano)
		if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
			return nil, err
		}
		result = next
		return nil, nil
	})
	if err != nil {
		return Card{}, err
	}
	return result, nil
}

// PreviousCardRow returns the card's most recent factory row recorded under a
// run other than excludeRunID — the re-lease carry-over's read (card t1930
// defect ③: a card leased under a NEW run used to be born with empty stage
// and evidence_sha, orphaning the prior run's progress). ok is false when no
// other run ever recorded the card. The ordering matches
// PreviousCardWorktree: newest run first, then newest row.
func (f *FactoryDB) PreviousCardRow(ctx context.Context, cardID, excludeRunID string) (Card, bool, error) {
	// SQL: one constant statement; cardID and excludeRunID ride ? placeholders,
	// never the statement text.
	const query = `SELECT c.run_id,c.card_id,c.owner_label,c.state,c.version,c.evidence_path,c.updated_at,
c.stage,c.lease_holder,c.lease_expires_at,c.heartbeat_at,c.decision_gate,c.decision_question,c.decision_resume,
c.decider,c.decided_at,c.failure_reason,c.hint_prefer,c.hint_after,c.spec_id,c.worktree_path,c.evidence_sha,
c.merge_sha,c.merge_tree,c.remeasure_path,c.bundle_id,c.bundle_order,c.contract_spec_id,c.contract_sha256,c.contract_signed_at,c.contract_event
FROM cards c LEFT JOIN runs r ON r.run_id = c.run_id
WHERE c.card_id = ? AND c.run_id != ?
ORDER BY r.created_at DESC, c.updated_at DESC
LIMIT 1`
	var c Card
	err := f.DB.QueryRowContext(ctx, query, cardID, excludeRunID).Scan(&c.RunID, &c.CardID, &c.OwnerLabel, &c.State, &c.Version, &c.EvidencePath, &c.UpdatedAt,
		&c.Stage, &c.LeaseHolder, &c.LeaseExpiresAt, &c.HeartbeatAt, &c.DecisionGate, &c.DecisionQuestion, &c.DecisionResume,
		&c.Decider, &c.DecidedAt, &c.FailureReason, &c.HintPrefer, &c.HintAfter, &c.SpecID, &c.WorktreePath, &c.EvidenceSHA,
		&c.MergeSHA, &c.MergeTree, &c.RemeasurePath, &c.BundleID, &c.BundleOrder, &c.ContractSpecID, &c.ContractSHA256, &c.ContractSignedAt, &c.ContractEvent)
	if errors.Is(err, sql.ErrNoRows) {
		return Card{}, false, nil
	}
	if err != nil {
		return Card{}, false, err
	}
	return c, true, nil
}

// SetCardStageEvidence is the re-lease carry-over write (card t1930 defect ③):
// a version-fenced mutation that lands the prior row's stage and evidence_sha
// on the freshly recorded row of the new run. It is a NEW row's carry-over,
// never a rewrite of the prior run's row, and the fence refuses a caller
// whose read of the fresh row is stale.
func (f *FactoryDB) SetCardStageEvidence(ctx context.Context, runID, cardID, stage, evidenceSHA string, expectedVersion int64, now time.Time) (Card, error) {
	return f.mutateCardColumns(ctx, runID, cardID, expectedVersion, now, func(c *Card) {
		c.Stage = strings.TrimSpace(stage)
		c.EvidenceSHA = strings.TrimSpace(evidenceSHA)
	})
}

// SetCardEvidenceSHA is the rebind-evidence record layer (card t1930 defect ④):
// a version-fenced mutation of evidence_sha alone. The VALUE this lands must
// come from ResolveRebindEvidence — the newest verdict file's audited_sha —
// never from the caller's own input; the record layer cannot enforce that
// provenance, so the verb does.
func (f *FactoryDB) SetCardEvidenceSHA(ctx context.Context, runID, cardID, sha string, expectedVersion int64, now time.Time) (Card, error) {
	if strings.TrimSpace(sha) == "" {
		return Card{}, fmt.Errorf("%w: evidence sha is required", ErrInvalidCardInput)
	}
	return f.mutateCardColumns(ctx, runID, cardID, expectedVersion, now, func(c *Card) {
		c.EvidenceSHA = strings.TrimSpace(sha)
	})
}

// ResolveRebindEvidence reads the card's newest `<phase>*.md` verdict file
// under its worktree's .moai/reports/<card-id>/ (the same newest-selection
// rule readAuditVerdict uses: highest -iter number, then latest mtime) and
// returns the FULL object name of the commit that verdict's audited_sha
// names, verified by the T10 reader's own conditions — the commit resolves in
// the card worktree and is an ancestor of, or equal to, its HEAD. It is the
// only value the rebind-evidence verb may write, so a lane cannot forge an
// audit claim by passing an arbitrary SHA.
func ResolveRebindEvidence(ctx context.Context, card Card, phase string) (string, error) {
	if !ValidCardID(card.CardID) {
		return "", evidenceErr("card id %q is not a safe path segment", card.CardID)
	}
	if strings.TrimSpace(card.WorktreePath) == "" {
		return "", evidenceErr("card has no worktree path")
	}
	if strings.TrimSpace(phase) == "" {
		return "", evidenceErr("no audit phase was named")
	}
	v, err := readNewestVerdict(card.WorktreePath, card.CardID, phase)
	if err != nil {
		return "", err
	}
	full, err := verifyCommitAtHead(ctx, card.WorktreePath, v.AuditedSHA)
	if err != nil {
		return "", err
	}
	return full, nil
}

// readNewestVerdict locates the newest `<phase>*.md` verdict file and applies
// the E-VERDICT single-value line rules (verdictLinePattern /
// auditedSHAPattern in card_evidence_readers.go) WITHOUT the recorded-commit
// comparison parseAuditVerdict adds — the rebind reader is the one place the
// recorded value is precisely what is being recomputed.
func readNewestVerdict(dir, cardID, phase string) (auditVerdict, error) {
	pattern := filepath.Join(dir, ".moai", "reports", cardID, phase+"*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return auditVerdict{}, evidenceErr("no %s verdict file under %s", phase, filepath.Dir(pattern))
	}
	newest, err := newestVerdictFile(matches)
	if err != nil {
		return auditVerdict{}, err
	}
	raw, err := readBoundedFile(newest)
	if err != nil {
		return auditVerdict{}, evidenceErr("verdict file %s: %v", newest, err)
	}
	verdicts, shas := map[string]bool{}, map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, " \t\r")
		if m := verdictLinePattern.FindStringSubmatch(line); m != nil {
			verdicts[m[1]] = true
		}
		if m := auditedSHAPattern.FindStringSubmatch(line); m != nil {
			shas[m[1]] = true
		}
	}
	out := auditVerdict{Path: newest}
	switch len(verdicts) {
	case 1:
		for v := range verdicts {
			out.Verdict = v
		}
	default:
		return out, evidenceErr("verdict file %s carries no single `verdict:` line", newest)
	}
	switch len(shas) {
	case 1:
		for s := range shas {
			out.AuditedSHA = s
		}
	default:
		return out, evidenceErr("verdict file %s carries no single `audited_sha:` line", newest)
	}
	return out, nil
}
