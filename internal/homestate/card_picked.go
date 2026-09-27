package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	specIDPattern = regexp.MustCompile(`^SPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*$`)
	hex64Pattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// ContractRef is the optional pointer from a card record to a signed
// contract: the contract's SPEC identifier, its signed digest, the signing
// time, and the locator of the signing event in the moai-owned contract store
// (the SHA-256 of that store line; empty until the store exists). F1 checks
// the pointer's format only — it never opens the store or the contract file.
type ContractRef struct {
	SpecID   string `json:"spec_id"`
	SHA256   string `json:"sha256"`
	SignedAt string `json:"signed_at"`
	Event    string `json:"event"`
}

// ParseContractRef parses `<spec-id>,<sha256>,<signed-at>[,<event>]`.
func ParseContractRef(s string) (ContractRef, error) {
	parts := strings.Split(strings.TrimSpace(s), ",")
	if len(parts) != 3 && len(parts) != 4 {
		return ContractRef{}, fmt.Errorf("%w: contract ref wants <spec-id>,<sha256>,<signed-at>[,<event>], got %d fields", ErrInvalidCardInput, len(parts))
	}
	ref := ContractRef{SpecID: strings.TrimSpace(parts[0]), SHA256: strings.TrimSpace(parts[1]), SignedAt: strings.TrimSpace(parts[2])}
	if len(parts) == 4 {
		ref.Event = strings.TrimSpace(parts[3])
	}
	return ref, ref.Validate()
}

// Validate checks the pointer's format: a SPEC identifier, a 64-hex digest,
// an RFC 3339 signing time, and a 64-hex or empty event locator.
func (r ContractRef) Validate() error {
	switch {
	case !specIDPattern.MatchString(r.SpecID):
		return fmt.Errorf("%w: contract spec id %q is not a SPEC identifier", ErrInvalidCardInput, r.SpecID)
	case !hex64Pattern.MatchString(r.SHA256):
		return fmt.Errorf("%w: contract digest must be 64 lowercase hex characters", ErrInvalidCardInput)
	case r.Event != "" && !hex64Pattern.MatchString(r.Event):
		return fmt.Errorf("%w: contract event locator must be 64 lowercase hex characters or empty", ErrInvalidCardInput)
	}
	if _, err := time.Parse(time.RFC3339, r.SignedAt); err != nil {
		return fmt.Errorf("%w: contract signing time %q is not RFC 3339", ErrInvalidCardInput, r.SignedAt)
	}
	return nil
}

// CardFields are the values `assign` may set on a picked card. A nil field is
// left as it is; a pointer to "" clears it.
type CardFields struct {
	HintPrefer, HintAfter, SpecID, WorktreePath *string
	Contract                                    *ContractRef
}

func (f CardFields) empty() bool {
	return f.HintPrefer == nil && f.HintAfter == nil && f.SpecID == nil && f.WorktreePath == nil && f.Contract == nil
}

func (f CardFields) validate() error {
	if f.HintPrefer != nil && *f.HintPrefer != "" {
		key, value, ok := strings.Cut(*f.HintPrefer, "=")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: prefer hint %q is not key=value", ErrInvalidCardInput, *f.HintPrefer)
		}
	}
	if f.HintAfter != nil && *f.HintAfter != "" && !ValidCardID(*f.HintAfter) {
		return fmt.Errorf("%w: after hint %q is not a card id", ErrInvalidCardInput, *f.HintAfter)
	}
	if f.SpecID != nil && *f.SpecID != "" && !specIDPattern.MatchString(*f.SpecID) {
		return fmt.Errorf("%w: %q is not a SPEC identifier", ErrInvalidCardInput, *f.SpecID)
	}
	if f.WorktreePath != nil && *f.WorktreePath != "" && !filepath.IsAbs(*f.WorktreePath) {
		return fmt.Errorf("%w: worktree path %q is not absolute", ErrInvalidCardInput, *f.WorktreePath)
	}
	if f.Contract != nil {
		return f.Contract.Validate()
	}
	return nil
}

func (f CardFields) apply(c *Card) {
	set := func(dst *string, src *string) {
		if src != nil {
			*dst = strings.TrimSpace(*src)
		}
	}
	set(&c.HintPrefer, f.HintPrefer)
	set(&c.HintAfter, f.HintAfter)
	set(&c.SpecID, f.SpecID)
	set(&c.WorktreePath, f.WorktreePath)
	if f.Contract != nil {
		c.ContractSpecID, c.ContractSHA256, c.ContractSignedAt, c.ContractEvent = f.Contract.SpecID, f.Contract.SHA256, f.Contract.SignedAt, f.Contract.Event
	}
}

// RecordPicked is T1: it creates the card record at `picked`, or updates the
// fields of a card still in `picked`. Whether the queue item is in the queue
// state `picked` is the caller's precondition (REQ-FR-022) — this package
// cannot read the queue. A card past `picked` keeps its fields; asking to
// change them is refused, and asking for nothing returns the card unchanged.
func (f *FactoryDB) RecordPicked(ctx context.Context, runID, cardID string, fields CardFields, actor string, now time.Time) (Card, error) {
	if strings.TrimSpace(runID) == "" || !ValidCardID(cardID) {
		return Card{}, fmt.Errorf("%w: run id and a valid card id are required", ErrInvalidCardInput)
	}
	if err := fields.validate(); err != nil {
		return Card{}, err
	}
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	nowText := now.Format(time.RFC3339Nano)
	var result Card
	err := f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, runID, cardID)
		if errors.Is(err, ErrCardNotFound) {
			c := Card{RunID: runID, CardID: cardID, State: CardPicked, Version: 1, UpdatedAt: nowText}
			fields.apply(&c)
			// SQL: the concatenated fragment is a compile-time constant; every value goes through a ? placeholder.
			if _, err := tx.ExecContext(ctx, `INSERT INTO cards(`+cardSelectColumns+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				c.RunID, c.CardID, c.OwnerLabel, c.State, c.Version, c.EvidencePath, c.UpdatedAt,
				c.Stage, c.LeaseHolder, c.LeaseExpiresAt, c.HeartbeatAt, c.DecisionGate, c.DecisionQuestion, c.DecisionResume,
				c.Decider, c.DecidedAt, c.FailureReason, c.HintPrefer, c.HintAfter, c.SpecID, c.WorktreePath, c.EvidenceSHA,
				c.MergeSHA, c.MergeTree, c.RemeasurePath, c.ContractSpecID, c.ContractSHA256, c.ContractSignedAt, c.ContractEvent); err != nil {
				return nil, err
			}
			if err := appendEvent(ctx, tx, runID, "card.transition", map[string]any{"card_id": cardID, "from": "", "to": CardPicked, "version": 1, "actor": actor}, now); err != nil {
				return nil, err
			}
			result = c
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if fields.empty() {
			result = cur
			return nil, nil
		}
		if cur.State != CardPicked {
			return nil, fmt.Errorf("%w: card %s is %s; its hints and pointers change only while picked", ErrIllegalTransition, cur.CardID, cur.State)
		}
		next := cur
		fields.apply(&next)
		if next == cur {
			result = cur
			return nil, nil
		}
		next.Version = cur.Version + 1
		next.UpdatedAt = nowText
		if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
			return nil, err
		}
		if err := appendEvent(ctx, tx, runID, "card.fields", map[string]any{"card_id": cardID, "version": next.Version, "actor": actor}, now); err != nil {
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

// predecessorMerged is the T2 `after` guard (REQ-FR-016): the predecessor
// must have a factory record — in any run — that reached the local merge.
func predecessorMerged(ctx context.Context, tx *sql.Tx, after string) error {
	rows, err := tx.QueryContext(ctx, `SELECT state FROM cards WHERE card_id=?`, after)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	seen := false
	for rows.Next() {
		var state string
		if err := rows.Scan(&state); err != nil {
			return err
		}
		seen = true
		switch state {
		case CardMergedLocal, CardPushed, CardCIGreen, CardDone:
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !seen {
		return fmt.Errorf("%w: %s has no factory record (clear the hint with assign --after \"\")", ErrUnknownPredecessor, after)
	}
	return fmt.Errorf("%w: %s has not reached merged-local", ErrPredecessorUnmerged, after)
}
