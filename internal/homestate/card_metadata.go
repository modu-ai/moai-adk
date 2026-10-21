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
	"fmt"
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
