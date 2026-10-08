package factory

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

type GTDOperationState string

const (
	GTDOperationPrepared   GTDOperationState = "prepared"
	GTDOperationInvoking   GTDOperationState = "invoking"
	GTDOperationReconciled GTDOperationState = "reconciled"
	GTDOperationBlocked    GTDOperationState = "blocked"
)

// GTDActionDispatch is the dispatch action's vocabulary. The engine keys its
// factory-binding reconciliation on it (review round-20 P1,
// SPEC-FACTORY-COMPLETION-RECOVERY-001): the owner's readback proves its own
// effect, and the dispatch binding is the factory record's half of the same
// success.
const GTDActionDispatch = "dispatch"

type GTDOperation struct {
	OperationID  string            `json:"operation_id"`
	MissionID    string            `json:"mission_id"`
	Action       string            `json:"action"`
	Target       string            `json:"target"`
	State        GTDOperationState `json:"state"`
	ReceiptJSON  []byte            `json:"receipt"`
	SnapshotHash string            `json:"snapshot_hash"`
	UpdatedAt    string            `json:"updated_at"`
}

type GTDOperationOwner interface {
	Readback(context.Context, GTDOperation) (bool, error)
	Apply(context.Context, GTDOperation) error
}

func validGTDOperation(op GTDOperation) bool {
	return strings.TrimSpace(op.OperationID) != "" && strings.TrimSpace(op.MissionID) != "" && strings.TrimSpace(op.Action) != "" && strings.TrimSpace(op.Target) != "" && strings.TrimSpace(op.SnapshotHash) != "" && len(op.ReceiptJSON) > 0
}

// DispatchIdentifiers is an optional GTDOperationOwner capability: the
// assigned card and its run a dispatch operation's factory reconciliation
// keys on (review round-21, card t1538). The op's own Target/MissionID are
// the supervisor lineage's identifiers and may name another vocabulary —
// the goal mission's gtd item ref and session id — so the owner that
// actually performed the assignment carries the factory vocabulary to the
// engine. An owner without the capability reconciles against the op's own
// identifiers.
type DispatchIdentifiers interface {
	DispatchReconcileIdentifiers() (cardID, runID string)
}

// gtdDispatchReconcileIdentifiers resolves the reconcile identifiers for a
// dispatch operation: the owner's revealed assignment when it carries one,
// else the operation's own Target and MissionID.
func gtdDispatchReconcileIdentifiers(ctx context.Context, owner GTDOperationOwner, op GTDOperation) (string, string) {
	if reveal, ok := owner.(DispatchIdentifiers); ok {
		if cardID, runID := reveal.DispatchReconcileIdentifiers(); cardID != "" && runID != "" {
			return cardID, runID
		}
	}
	return op.Target, op.MissionID
}

func scanGTDOperation(row interface{ Scan(...any) error }) (GTDOperation, error) {
	var op GTDOperation
	var state string
	err := row.Scan(&op.OperationID, &op.MissionID, &op.Action, &op.Target, &state, &op.ReceiptJSON, &op.SnapshotHash, &op.UpdatedAt)
	op.State = GTDOperationState(state)
	return op, err
}

func LoadGTDOperation(ctx context.Context, store *BacklogStore, operationID string) (GTDOperation, error) {
	db, err := openGTDDB(store)
	if err != nil {
		return GTDOperation{}, err
	}
	defer func() { _ = db.Close() }()
	op, err := scanGTDOperation(db.QueryRowContext(ctx, `SELECT operation_id,mission_id,action,target,state,receipt_json,snapshot_hash,updated_at FROM gtd_operations WHERE operation_id=?`, operationID))
	if err != nil {
		return GTDOperation{}, mapBacklogEngineError("load gtd operation", err)
	}
	return op, nil
}

func PrepareGTDOperation(ctx context.Context, store *BacklogStore, candidate GTDOperation) (GTDOperation, bool, error) {
	if !validGTDOperation(candidate) {
		return GTDOperation{}, false, errors.New("gtd operation: invalid")
	}
	db, err := openGTDDB(store)
	if err != nil {
		return GTDOperation{}, false, err
	}
	defer func() { _ = db.Close() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO gtd_operations(operation_id,mission_id,action,target,state,receipt_json,snapshot_hash,updated_at) VALUES(?,?,?,?,?,?,?,?)`, candidate.OperationID, candidate.MissionID, candidate.Action, candidate.Target, string(GTDOperationPrepared), candidate.ReceiptJSON, candidate.SnapshotHash, now)
	if err != nil {
		return GTDOperation{}, false, mapBacklogEngineError("prepare gtd operation", err)
	}
	rows, _ := res.RowsAffected()
	stored, err := scanGTDOperation(db.QueryRowContext(ctx, `SELECT operation_id,mission_id,action,target,state,receipt_json,snapshot_hash,updated_at FROM gtd_operations WHERE operation_id=?`, candidate.OperationID))
	if err != nil {
		return GTDOperation{}, false, mapBacklogEngineError("read prepared gtd operation", err)
	}
	if stored.MissionID != candidate.MissionID || stored.Action != candidate.Action || stored.Target != candidate.Target || stored.SnapshotHash != candidate.SnapshotHash || !bytes.Equal(stored.ReceiptJSON, candidate.ReceiptJSON) {
		return GTDOperation{}, false, errors.New("gtd operation: identity_collision")
	}
	return stored, rows == 1, nil
}

func markGTDOperationState(ctx context.Context, store *BacklogStore, operationID string, from, to GTDOperationState) error {
	db, err := openGTDDB(store)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	res, err := db.ExecContext(ctx, `UPDATE gtd_operations SET state=?,updated_at=? WHERE operation_id=? AND state=?`, string(to), time.Now().UTC().Format(time.RFC3339Nano), operationID, string(from))
	if err != nil {
		return mapBacklogEngineError("transition gtd operation", err)
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("gtd operation: transition_conflict")
	}
	return nil
}

// ExecuteGTDOperation makes the SQLite receipt the effect gate. An uncertain
// invoking record is never retried from a timeout: only authoritative owner
// readback may reconcile it, preventing a second external effect after crash.
func ExecuteGTDOperation(ctx context.Context, store *BacklogStore, candidate GTDOperation, owner GTDOperationOwner) (GTDOperation, error) {
	if owner == nil {
		return GTDOperation{}, errors.New("gtd operation: nil owner")
	}
	stored, _, err := PrepareGTDOperation(ctx, store, candidate)
	if err != nil {
		return GTDOperation{}, err
	}
	applied, err := owner.Readback(ctx, stored)
	if err != nil {
		return stored, err
	}
	// The owner's readback proves its own effect; a dispatch reconciles
	// only when the factory record's half agrees too (review round-20 P1,
	// SPEC-FACTORY-COMPLETION-RECOVERY-001). A partial failure — the
	// assignment save landed, the binding write did not — is repaired HERE
	// rather than declared reconciled over: re-running the owner's apply
	// would repeat the save that already committed, and the binding write
	// alone is what went missing.
	// A COMPLETED (reconciled) operation must never move the binding on
	// replay: its assignment row persists in the queue record, so the
	// replay's readback proves history, not a current engagement — and
	// re-running the reconcile would regress the binding to the replayed
	// operation's run, arming the old run's approval against the new run's
	// work (review round-21 P1). Only an operation still in flight
	// (prepared or invoking) reconciles the factory record's half.
	if applied && stored.Action == GTDActionDispatch && stored.State != GTDOperationReconciled {
		cardID, runID := gtdDispatchReconcileIdentifiers(ctx, owner, stored)
		if rerr := reconcileGTDDispatchBinding(ctx, store, cardID, runID); rerr != nil {
			return stored, rerr
		}
	}
	if applied {
		if stored.State == GTDOperationReconciled {
			return stored, nil
		}
		if err := markGTDOperationState(ctx, store, stored.OperationID, stored.State, GTDOperationReconciled); err != nil {
			return LoadGTDOperation(ctx, store, stored.OperationID)
		}
		return LoadGTDOperation(ctx, store, stored.OperationID)
	}
	if stored.State != GTDOperationPrepared {
		return stored, nil
	}
	if err := markGTDOperationState(ctx, store, stored.OperationID, GTDOperationPrepared, GTDOperationInvoking); err != nil {
		return LoadGTDOperation(ctx, store, stored.OperationID)
	}
	stored.State = GTDOperationInvoking
	if err := owner.Apply(ctx, stored); err != nil {
		return stored, err
	}
	applied, err = owner.Readback(ctx, stored)
	if err != nil {
		return stored, err
	}
	if !applied {
		return stored, errors.New("gtd operation: authoritative_readback_missing")
	}
	if stored.Action == GTDActionDispatch {
		cardID, runID := gtdDispatchReconcileIdentifiers(ctx, owner, stored)
		if rerr := reconcileGTDDispatchBinding(ctx, store, cardID, runID); rerr != nil {
			return stored, rerr
		}
	}
	if err := markGTDOperationState(ctx, store, stored.OperationID, GTDOperationInvoking, GTDOperationReconciled); err != nil {
		return stored, err
	}
	return LoadGTDOperation(ctx, store, stored.OperationID)
}

// reconcileGTDDispatchBinding verifies — and repairs — the factory record's
// half of a dispatch success (review round-20 P1): the owner's readback
// proves the assignment, and the dispatch binding must name the operation's
// run before the operation may reconcile. The repair runs
// repairGTDDispatchRecord, because the missing writes are the factory
// record's alone — the assignment they pair with already committed — and
// re-running the owner's apply would repeat that committed save.
func reconcileGTDDispatchBinding(ctx context.Context, store *BacklogStore, cardID, runID string) error {
	current, err := gtdDispatchBindingCurrent(ctx, store, cardID, runID)
	if err != nil {
		return err
	}
	if current {
		return nil
	}
	path := gtdFactoryDBForStore(store)
	if path == "" {
		return fmt.Errorf("gtd operation: dispatch target %s does not name run %s and no factory database is reachable for the binding repair", cardID, runID)
	}
	if err := repairGTDDispatchRecord(ctx, store, path, cardID, runID); err != nil {
		return err
	}
	current, err = gtdDispatchBindingCurrent(ctx, store, cardID, runID)
	if err != nil {
		return err
	}
	if !current {
		return fmt.Errorf("gtd operation: dispatch target %s does not name run %s after the binding repair", cardID, runID)
	}
	return nil
}

// repairGTDDispatchRecord completes the factory-record half of a dispatch
// whose apply landed the assignment save but failed the factory writes
// (review round-20 P1, T2 half added by round-21): the card row —
// RecordPicked when the run holds none, the mirror's T1 — moved to assigned
// under the authoritative assignment's owner when it sits at picked (the
// mirror's T2), and the dispatch binding. The mirror's shape, driven from
// the operation engine: the same queue-lock discipline, a picked queue item
// as the only precondition, and the runtime assignment as the owner of
// record. The completion gate reads the repaired triple as one binding: the
// row resolves the bound run, the owner names the engaged lane, and the
// binding re-targets the gate away from the superseded approval.
func repairGTDDispatchRecord(ctx context.Context, store *BacklogStore, factoryDBPath, cardID, runID string) error {
	return store.WithLock(func(l *LockedBacklog) error {
		record, err := l.LoadPure()
		if err != nil {
			return fmt.Errorf("read queue: %w", err)
		}
		picked := false
		owner := ""
		for _, item := range record.Items {
			if item.ID == cardID {
				picked = item.State == BacklogStatePicked
			}
		}
		for _, a := range record.Runtime.Assignments {
			if a.CardID == cardID && a.RunID == runID {
				owner = a.OwnerLabel
				break
			}
		}
		if !picked {
			return fmt.Errorf("queue item %s is not picked", cardID)
		}
		if owner == "" {
			return fmt.Errorf("queue item %s carries no runtime assignment for run %s — the authoritative owner is unknown", cardID, runID)
		}
		db, err := homestate.OpenFactoryPath(factoryDBPath)
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
		now := time.Now().UTC()
		card, err := db.LoadCard(ctx, runID, cardID)
		if errors.Is(err, homestate.ErrCardNotFound) {
			if card, err = db.RecordPicked(ctx, runID, cardID, homestate.CardFields{}, "dispatch", now); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		switch {
		case card.State == homestate.CardPicked:
			// The mirror's T2: the authoritative assignment owns the row.
			if _, err := db.Transition(ctx, homestate.TransitionRequest{RunID: runID, CardID: cardID, To: homestate.CardAssigned, ExpectedVersion: card.Version, Actor: "dispatch", Owner: owner, Now: now}); err != nil {
				return err
			}
		case card.State == homestate.CardAssigned && card.OwnerLabel == owner:
			// Already the authoritative assignment.
		default:
			return fmt.Errorf("factory record for %s is %s (owner %q), not assignable to %s", cardID, card.State, card.OwnerLabel, owner)
		}
		return db.RecordDispatchBinding(ctx, cardID, runID, now)
	})
}

// RecordDispatchBindingIfEngaged records cardID -> runID as the card's
// current factory engagement when the card has factory rows — REQ-FCR-002's
// scope sentence at the storage layer (review round-18 P2, tightened by the
// round-19 edge): a card with NO factory row in ANY run is an ordinary card,
// out of scope for the binding; the write is skipped silently and the card
// keeps its existing completion behavior. A card WITH rows (in this or any
// other run) is bound to the targeted run: the later done gate then refuses
// the stale approval (run mismatch, or run-unresolvable when the targeted
// run's row is yet to be created by the dispatch) — never success on the old
// approval. A missing factory database is a no-op: the first real dispatch
// records the binding through the mirror path instead.
func RecordDispatchBindingIfEngaged(root, cardID, runID string) error {
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil
		}
		return statErr
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	var existing int
	if err := db.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM cards WHERE card_id=?`, cardID).Scan(&existing); err != nil {
		return err
	}
	if existing == 0 {
		return nil
	}
	return db.RecordDispatchBinding(context.Background(), cardID, runID, time.Now())
}

// gtdDispatchBindingCurrent reports whether the card's recorded dispatch
// binding names the operation's run AND the card row that run resolves
// against — the factory record's whole half of a dispatch reconciliation
// (review round-20 P1, completed by round-21). Scope follows REQ-FCR-002's
// sentence: a missing factory database, a missing card_dispatch table, or a
// card with no factory row leaves the axis vacuous (the completion gate does
// not apply); a binding naming another run, or factory rows orphaned from
// any binding, reads stale — repairGTDDispatchRecord recovers both.
//
// The factory database is located by gtdFactoryDBForStore — homestate's
// canonical resolution when a project root sits above the queue, and the
// queue directory's state-directory sibling for a home-resolved queue.
func gtdDispatchBindingCurrent(ctx context.Context, store *BacklogStore, cardID, runID string) (bool, error) {
	path := gtdFactoryDBForStore(store)
	if path == "" {
		return true, nil
	}
	db, err := homestate.OpenFactoryReadonly(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = db.Close() }()
	hasDispatch, err := db.FactoryTablePresent(ctx, "card_dispatch")
	if err != nil {
		return false, err
	}
	if !hasDispatch {
		// An older-schema store: no dispatch bindings exist anywhere, so
		// none can be stale (the completion gate reads the same store as
		// unverified, never as a migration trigger).
		return true, nil
	}
	var bound string
	err = db.DB.QueryRowContext(ctx, `SELECT run_id FROM card_dispatch WHERE card_id=?`, cardID).Scan(&bound)
	if errors.Is(err, sql.ErrNoRows) {
		var rows int
		if rerr := db.DB.QueryRowContext(ctx, `SELECT count(*) FROM cards WHERE card_id=?`, cardID).Scan(&rows); rerr != nil {
			return false, rerr
		}
		// No factory row at all: out of scope, the axis passes. Rows
		// orphaned from their binding: stale — apply's binding write
		// recovers exactly this shape.
		return rows == 0, nil
	}
	if err != nil {
		return false, err
	}
	if bound != runID {
		return false, nil
	}
	// The binding names this run, but the mirror half may still be missing:
	// reconciliation requires the card row the bound run resolves against
	// (review round-20 P1) — without it the completion gate reads the
	// binding as run-unresolvable and every close refuses.
	var row int
	if err := db.DB.QueryRowContext(ctx, `SELECT count(*) FROM cards WHERE run_id=? AND card_id=?`, runID, cardID).Scan(&row); err != nil {
		return false, err
	}
	return row > 0, nil
}

// gtdFactoryDBForStore locates the project factory database for a queue
// store. The project layout resolves through homestate's canonical
// FactoryDBPath from the project root found above the queue (see
// gtdProjectRootForStore) — the same resolution every dispatch writer and
// the completion gate use. A queue resolved to a home state directory has no
// project .moai above it, and its factory database is the queue directory's
// state-directory sibling (<state>/factory/factory.db — the layout
// <home>/db/<key>/{todo,factory} produces). "" means no factory database is
// reachable and the dispatch-binding axis stays vacuous for the engine
// check.
func gtdFactoryDBForStore(store *BacklogStore) string {
	if root := gtdProjectRootForStore(store); root != "" {
		if path, err := homestate.FactoryDBPath(root); err == nil {
			if _, statErr := os.Stat(path); statErr == nil {
				return path
			}
		}
	}
	sibling := filepath.Join(filepath.Dir(filepath.Dir(store.path)), "factory", "factory.db")
	if _, err := os.Stat(sibling); err == nil {
		return sibling
	}
	return ""
}

// gtdProjectRootForStore walks up from the queue store to the project root —
// the directory whose .moai carries the queue — so the factory database
// resolves through homestate's canonical FactoryDBPath, the same resolution
// every dispatch writer and the completion gate use. The walk is bounded; a
// queue with no project .moai ancestor (the home-queue fallback) yields ""
// and the caller falls back to the state-directory sibling.
func gtdProjectRootForStore(store *BacklogStore) string {
	dir := filepath.Dir(store.path)
	for i := 0; i < 6; i++ {
		if fi, err := os.Stat(filepath.Join(dir, ".moai")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
	return ""
}
