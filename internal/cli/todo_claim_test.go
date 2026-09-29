// todo_claim_test.go — SPEC-TODO-CLAIM-LEASE-001 M3/M4 acceptance tests:
// the `moai todo claim [--lane <label>] [--renew <id>]` verb surface
// (AC-TCL-001/003/006/009/010/012) and the MCP todo_claim mirror's refusal
// equality (REQ-TCL-013).
//
// Lane governance drives factoryLaneRefusal through the same internal/config
// env constants the launcher stamps — never inline strings — and every
// byte-identity assertion compares the engine artifact's SHA-256 around the
// refused invocation.
package cli

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// claimFileHash returns the sha256 of the fixture queue's engine artifact —
// the byte-identity surface AC-TCL-003's 사전/사후 hash comparison names.
func claimFileHash(t *testing.T, store *kanban.BacklogStore) string {
	t.Helper()
	raw, err := os.ReadFile(store.EnginePath())
	if err != nil {
		t.Fatalf("read engine artifact: %v", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// claimSeedRecord mutates the fixture record directly (the engine-bypass
// seed pattern), stamping an optional lease.
func claimSeedRecord(t *testing.T, store *kanban.BacklogStore, id, state, holder, expiresAt string) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != id {
				continue
			}
			rec.Items[i].State = kanban.BacklogState(state)
			if holder != "" {
				h, e := holder, expiresAt
				rec.Items[i].PickedBy = &h
				rec.Items[i].LeaseExpiresAt = &e
			}
			return nil
		}
		return errors.New("no item " + id)
	}); err != nil {
		t.Fatalf("seed %s: %v", id, err)
	}
}

// AC-TCL-001 — claim takes the oldest queued card, stamps the lease, and
// prints id + text prefix + expiry. --lane attributes the claim to the
// operator-supplied label.
func TestTodoClaim_Success(t *testing.T) {
	root, store := todoFixture(t)
	_ = root
	if _, _, err := storeAdd(t, store, "first card to claim"); err != nil {
		t.Fatalf("seed 1: %v", err)
	}
	if _, _, err := storeAdd(t, store, "second card to claim"); err != nil {
		t.Fatalf("seed 2: %v", err)
	}

	stdout, _, err := runTodo(t, "claim")
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if !strings.Contains(stdout, "claimed t1") || !strings.Contains(stdout, "first card to claim") {
		t.Errorf("claim stdout = %q, want the oldest card's id + text prefix", stdout)
	}
	if !strings.Contains(stdout, "lease_expires_at=") {
		t.Errorf("claim stdout = %q, want the lease expiry", stdout)
	}

	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if rec.Items[0].State != kanban.BacklogStatePicked {
		t.Fatalf("t1 state = %s, want picked", rec.Items[0].State)
	}
	if rec.Items[0].PickedBy == nil || *rec.Items[0].PickedBy != kanban.BacklogOperatorHolder {
		t.Errorf("picked_by = %v, want %q", rec.Items[0].PickedBy, kanban.BacklogOperatorHolder)
	}
	if rec.Items[0].LeaseExpiresAt == nil {
		t.Fatal("lease_expires_at = nil")
	}
	expiresAt, perr := time.Parse(time.RFC3339, *rec.Items[0].LeaseExpiresAt)
	if perr != nil {
		t.Fatalf("expiry %q not RFC 3339: %v", *rec.Items[0].LeaseExpiresAt, perr)
	}
	if d := time.Until(expiresAt); d < 14*time.Minute || d > 16*time.Minute {
		t.Errorf("lease duration = %v, want ≈ 15m", d)
	}
	if rec.Items[0].PickedAt == nil {
		t.Error("picked_at = nil, want stamped")
	}
	if rec.Items[1].State != kanban.BacklogStateQueued {
		t.Errorf("t2 state = %s, want still queued", rec.Items[1].State)
	}

	// --lane attributes the claim to the named lane.
	_, _, err = runTodo(t, "claim", "--lane", "lane-9")
	if err != nil {
		t.Fatalf("claim --lane: %v", err)
	}
	rec, err = store.LoadPure()
	if err != nil {
		t.Fatalf("reload after --lane: %v", err)
	}
	if rec.Items[1].PickedBy == nil || *rec.Items[1].PickedBy != "lane-9" {
		t.Errorf("t2 picked_by = %v, want lane-9", rec.Items[1].PickedBy)
	}
}

// AC-TCL-003 — a queue holding only non-queued states refuses with a
// distinct message and the queue file stays byte-identical; a future state
// inserted past the CHECK refuses by the same positive enumeration.
func TestTodoClaim_RefusesNonQueued(t *testing.T) {
	root, store := todoFixture(t)

	// picked (live lease) → the raced refusal.
	if _, _, err := storeAdd(t, store, "picked card"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	claimSeedRecord(t, store, "t1", "picked", "lane-1", time.Now().Add(15*time.Minute).UTC().Format(time.RFC3339))
	before := claimFileHash(t, store)
	_, _, err := runTodo(t, "claim")
	if err == nil {
		t.Fatal("claim over a live lease: err = nil, want refusal")
	}
	if !strings.Contains(err.Error(), "raced") {
		t.Errorf("err = %v, want the distinct raced indication", err)
	}
	if after := claimFileHash(t, store); after != before {
		t.Error("refused claim modified the queue file (picked arm)")
	}

	// hold → no eligible card: the dedicated non-error surface (exit code 3).
	claimSeedRecord(t, store, "t1", "hold", "", "")
	before = claimFileHash(t, store)
	_, _, err = runTodo(t, "claim")
	if err == nil {
		t.Fatal("claim over a held card: err = nil, want refusal")
	}
	var coded *exitCodeError
	if !errors.As(err, &coded) || coded.ExitCode() != factoryNextNoCardExit {
		t.Errorf("err = %v, want the no-card exit code %d", err, factoryNextNoCardExit)
	}
	if after := claimFileHash(t, store); after != before {
		t.Error("refused claim modified the queue file (hold arm)")
	}

	// dropped → no eligible card.
	claimSeedRecord(t, store, "t1", "dropped", "", "")
	before = claimFileHash(t, store)
	_, _, err = runTodo(t, "claim")
	if err == nil {
		t.Fatal("claim over a dropped card: err = nil, want refusal")
	}
	if after := claimFileHash(t, store); after != before {
		t.Error("refused claim modified the queue file (dropped arm)")
	}

	// A state no CHECK ever admitted — default-refuse by positive
	// enumeration (REQ-TCL-011), inserted past the constraint the way the
	// migration-test corruption fixture does.
	if err := insertFutureStateRow(t, store, "t-future"); err != nil {
		t.Fatalf("insert future state: %v", err)
	}
	before = claimFileHash(t, store)
	_, _, err = runTodo(t, "claim")
	if err == nil {
		t.Fatal("claim over a future state: err = nil, want refusal")
	}
	if after := claimFileHash(t, store); after != before {
		t.Error("refused claim modified the queue file (future-state arm)")
	}
	_ = root
}

// AC-TCL-006 — `claim --renew <id>` extends only the expiry; a foreign
// holder label is refused.
func TestTodoClaim_Renew(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := storeAdd(t, store, "renewable card"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, _, err := runTodo(t, "claim", "--lane", "lane-1"); err != nil {
		t.Fatalf("claim: %v", err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	oldExpiry := *rec.Items[0].LeaseExpiresAt
	pickedAt := rec.Items[0].PickedAt

	stdout, _, err := runTodo(t, "claim", "--renew", "t1", "--lane", "lane-1")
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	if !strings.Contains(stdout, "renewed t1") {
		t.Errorf("renew stdout = %q, want the renewed line", stdout)
	}
	rec, err = store.LoadPure()
	if err != nil {
		t.Fatalf("reload after renew: %v", err)
	}
	newExpiry, perr := time.Parse(time.RFC3339, *rec.Items[0].LeaseExpiresAt)
	if perr != nil {
		t.Fatalf("renewed expiry %q not RFC 3339: %v", *rec.Items[0].LeaseExpiresAt, perr)
	}
	old, _ := time.Parse(time.RFC3339, oldExpiry)
	if !newExpiry.After(old) {
		t.Errorf("renewal did not extend: %s → %s", oldExpiry, *rec.Items[0].LeaseExpiresAt)
	}
	if rec.Items[0].PickedBy == nil || *rec.Items[0].PickedBy != "lane-1" {
		t.Errorf("renew moved picked_by: %v", rec.Items[0].PickedBy)
	}
	if rec.Items[0].PickedAt == nil || *rec.Items[0].PickedAt != *pickedAt {
		t.Error("renew moved picked_at")
	}

	// A foreign holder label is refused with no change.
	if _, _, err := runTodo(t, "claim", "--renew", "t1", "--lane", "lane-2"); err == nil {
		t.Fatal("renew by a foreign holder: err = nil, want refusal")
	} else if !strings.Contains(err.Error(), "lease holder refused") {
		t.Errorf("err = %v, want the holder-mismatch refusal", err)
	}
}

// AC-TCL-009 — lane governance, three arms. The refusal text is the ONE
// wording source (todoLaneMutationRefusalText) and the queue file stays
// byte-identical on every refused arm.
func TestTodoClaim_LaneGovernance(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := storeAdd(t, store, "governed card"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	wantRefusal := todoLaneMutationRefusalText("claim")

	// Arm 1 — bare claim in a lane session: the existing REQ-SD-015 refusal.
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
	before := claimFileHash(t, store)
	_, _, err := runTodo(t, "claim")
	if err == nil {
		t.Fatal("bare claim in a lane session: err = nil, want refusal")
	}
	if err.Error() != wantRefusal {
		t.Errorf("arm 1 refusal = %q, want the shared text %q", err.Error(), wantRefusal)
	}
	if after := claimFileHash(t, store); after != before {
		t.Error("arm 1 refusal modified the queue file")
	}

	// Arm 3 — the --lane form in a lane session: the SAME refusal text (the
	// flag form is not a carve-out; the predicate assumes nothing about
	// caller identity).
	_, _, err = runTodo(t, "claim", "--lane", "lane-9")
	if err == nil {
		t.Fatal("--lane claim in a lane session: err = nil, want refusal")
	}
	if err.Error() != wantRefusal {
		t.Errorf("arm 3 refusal = %q, want the SAME shared text %q", err.Error(), wantRefusal)
	}
	if after := claimFileHash(t, store); after != before {
		t.Error("arm 3 refusal modified the queue file")
	}
	t.Setenv(config.EnvMoaiFactoryWorker, "")

	// Arm 2 — operator/lead session, --lane <label>: allowed, attributed.
	_, _, err = runTodo(t, "claim", "--lane", "lane-9")
	if err != nil {
		t.Fatalf("arm 2 claim --lane: %v", err)
	}
	rec, rerr := store.LoadPure()
	if rerr != nil {
		t.Fatalf("reload: %v", rerr)
	}
	if rec.Items[0].PickedBy == nil || *rec.Items[0].PickedBy != "lane-9" {
		t.Errorf("arm 2 picked_by = %v, want lane-9", rec.Items[0].PickedBy)
	}
}

// AC-TCL-010 — an empty queue exits with the dedicated no-card code and a
// non-error message.
func TestTodoClaim_NoCardExit(t *testing.T) {
	todoFixture(t)

	stdout, _, err := runTodo(t, "claim")
	if err == nil {
		t.Fatal("claim over an empty queue: err = nil, want the no-card exit")
	}
	var coded *exitCodeError
	if !errors.As(err, &coded) {
		t.Fatalf("err = %T(%v), want an exitCodeError", err, err)
	}
	if coded.ExitCode() != factoryNextNoCardExit {
		t.Errorf("exit code = %d, want %d (factoryNextNoCardExit)", coded.ExitCode(), factoryNextNoCardExit)
	}
	if !strings.Contains(stdout, "no queued card is available") {
		t.Errorf("stdout = %q, want the non-error message", stdout)
	}
}

// AC-TCL-012 — the human list/history surfaces expose picked_by and
// lease_expires_at on a claimed card, while the JSON surface stays excluded
// (the golden gate, REQ-TCL-014).
func TestTodoClaim_ListHistoryExposesLeaseColumns(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := storeAdd(t, store, "visible lease card"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, _, err := runTodo(t, "claim", "--lane", "lane-3"); err != nil {
		t.Fatalf("claim: %v", err)
	}

	listOut, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(listOut, "by=lane-3") || !strings.Contains(listOut, "lease=") {
		t.Errorf("human list = %q, want the by= and lease= cells on the claimed row", listOut)
	}

	historyOut, _, err := runTodo(t, "history", "t1")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if !strings.Contains(historyOut, "by=lane-3") || !strings.Contains(historyOut, "lease=") {
		t.Errorf("human history = %q, want the by= and lease= cells on the live line", historyOut)
	}

	// A row carrying no lease stays in its historical shape — no trailing
	// lease cells on the un-claimed row.
	listOut, _, err = runTodo(t, "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	if strings.Contains(listOut, `"picked_by"`) {
		t.Errorf("list --json exposed the lease columns: %q — REQ-TCL-014 forbids it", listOut)
	}
}

// insertFutureStateRow plants a row whose state no CHECK ever admitted,
// suspending the constraint exactly the way the migration corruption
// fixture does, so the default-refuse arm has a genuine future state. The
// refused claim writes nothing, so the planted row survives byte-identically.
func insertFutureStateRow(t *testing.T, store *kanban.BacklogStore, id string) error {
	t.Helper()
	db, err := sql.Open("sqlite", store.EnginePath())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	if _, err := db.ExecContext(context.Background(), `PRAGMA ignore_check_constraints = ON`); err != nil {
		return err
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (99, ?, 'a state from the future', '2026-09-29T00:00:00Z', NULL, 'emergent')`, id)
	return err
}

// storeAdd appends one card through the store and discards position.
func storeAdd(t *testing.T, store *kanban.BacklogStore, text string) (*kanban.BacklogItem, int, error) {
	t.Helper()
	return store.Add(text)
}

// REQ-TCL-013 — the MCP todo_claim mirror performs the SAME decision as the
// CLI verb and refuses with the SAME text: a lane session is refused in the
// bare and the --lane form alike, an allowed claim goes through the shared
// runTodoClaimRoot body anchored at the caller's project_root.
func TestTodoClaimMCP_Mirror(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := storeAdd(t, store, "mcp mirror card"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	wantRefusal := todoLaneMutationRefusalText("claim")

	// Refusal arm — lane session, --lane form: the identical text.
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
	before := claimFileHash(t, store)
	_, err := sdCallTool(t, handleTodoClaim, map[string]any{"project_root": root, "lane": "lane-9"})
	if err == nil {
		t.Fatal("mcp todo_claim in a lane session: err = nil, want refusal")
	}
	if !strings.Contains(err.Error(), wantRefusal) {
		t.Errorf("mcp refusal = %q, want it to carry the CLI-identical text %q (the todo_claim: prefix is the MCP surface convention, matching the todo_add parity guard)", err.Error(), wantRefusal)
	}
	if after := claimFileHash(t, store); after != before {
		t.Error("mcp refusal modified the queue file")
	}

	// Allowed arm — operator session: claims through the same body.
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	out, err := sdCallTool(t, handleTodoClaim, map[string]any{"project_root": root})
	if err != nil {
		t.Fatalf("mcp todo_claim: %v", err)
	}
	if !strings.Contains(out, "claimed t1") || !strings.Contains(out, "picked_by=operator") {
		t.Errorf("mcp output = %q, want the shared claim line", out)
	}
	rec, rerr := store.LoadPure()
	if rerr != nil {
		t.Fatalf("reload: %v", rerr)
	}
	if rec.Items[0].PickedBy == nil || *rec.Items[0].PickedBy != kanban.BacklogOperatorHolder {
		t.Errorf("picked_by = %v, want operator", rec.Items[0].PickedBy)
	}
}
