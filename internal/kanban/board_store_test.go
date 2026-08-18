// board_store_test.go — the sole-writer guard, the atomic write, and the
// minimal WIP-bounded transition beneath the board-wide lock
// (SPEC-KANBAN-BOARD-001 REQ-KB-017/018/009, M1).
//
// The sole-writer rule is ENFORCED, not documented (AP-17): a session whose
// declared role is not `lead` is refused a board write with the board
// byte-unchanged, and a session with no readable declaration is refused too —
// a refusal with nothing to read does not refuse, it admits every write. Both
// directions are required: a guard that refuses unconditionally satisfies the
// refusal alone.
package kanban

import (
	"os"
	"sync"
	"testing"
	"time"
)

// seedLead declares a lead session beneath root.
func seedLead(t *testing.T, root, sessionID string) {
	t.Helper()
	if err := DeclareRole(root, sessionID, RoleLead, "planner-host"); err != nil {
		t.Fatalf("DeclareRole(lead %s): %v", sessionID, err)
	}
}

// TestWriteBoardState_NonLeadRefused_BoardUnchanged — AC-KB-017 runtime half,
// refusal direction: a session declaring a non-lead role cannot write the
// board, and the refusal leaves the board byte-unchanged.
func TestWriteBoardState_NonLeadRefused_BoardUnchanged(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")
	if err := DeclareRole(root, "worker-sess", "runner", "runner-alpha"); err != nil {
		t.Fatalf("DeclareRole(worker): %v", err)
	}
	// Prior board content, so byte-unchanged is observable.
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-KB-0010", Column: "backlog", LastMovedAt: "2026-08-14T00:00:00Z"},
	}})
	before := readBoardBytes(t, root)

	err := WriteBoardState(root, "worker-sess", func(st *BoardState) error {
		st.Cards = append(st.Cards, Card{SpecID: "SPEC-KB-0099", Column: "run"})
		return nil
	})
	if err == nil {
		t.Fatal("WriteBoardState(non-lead) err = nil, want refusal")
	}
	if !IsNotSoleWriter(err) {
		t.Fatalf("err = %v, want ErrNotSoleWriter", err)
	}
	after := readBoardBytes(t, root)
	if string(before) != string(after) {
		t.Fatalf("board changed on refusal:\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestWriteBoardState_LeadSucceeds — AC-KB-017 runtime half, success
// direction: a session declaring `lead` writes through the same entry point
// and the write lands.
func TestWriteBoardState_LeadSucceeds(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")

	err := WriteBoardState(root, "leader-sess", func(st *BoardState) error {
		st.Cards = append(st.Cards, Card{SpecID: "SPEC-KB-0011", Column: "backlog"})
		return nil
	})
	if err != nil {
		t.Fatalf("WriteBoardState(lead) error = %v", err)
	}
	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard after lead write: %v", err)
	}
	if len(st.Cards) != 1 || st.Cards[0].SpecID != "SPEC-KB-0011" {
		t.Fatalf("write did not land: %+v", st.Cards)
	}
}

// TestWriteBoardState_UndeclaredSessionRefused — a session with no declaration
// at all is refused: the guard fails CLOSED, because treating an unreadable
// role as an admission is the failure mode the runtime half exists to close.
func TestWriteBoardState_UndeclaredSessionRefused(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	err := WriteBoardState(root, "ghost-sess", func(st *BoardState) error { return nil })
	if err == nil {
		t.Fatal("WriteBoardState(undeclared) err = nil, want fail-closed refusal")
	}
	if !IsNotSoleWriter(err) {
		t.Fatalf("err = %v, want ErrNotSoleWriter", err)
	}
	if _, statErr := os.Stat(BoardPath(root)); statErr == nil {
		t.Fatal("board state file appeared despite refused write")
	}
}

// TestWriteBoardState_CreatesStateDirectory — REQ-KB-021's directory half:
// the board directory is gitignored, so it cannot ship with the repository;
// the sole writer creates it on the write path.
func TestWriteBoardState_CreatesStateDirectory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")

	if err := WriteBoardState(root, "leader-sess", func(st *BoardState) error { return nil }); err != nil {
		t.Fatalf("WriteBoardState: %v", err)
	}
	if info, err := os.Stat(BoardDir(root)); err != nil || !info.IsDir() {
		t.Fatalf("board directory not created by the sole writer: %v", err)
	}
}

// TestWriteBoardState_NoLeftoverTempFiles — REQ-KB-018's hygiene tail: after a
// write, the board directory holds the state file and nothing else (the temp
// file is renamed away, not leaked). The same-directory property itself is
// checked statically (AC-KB-018's static half) — the temp must sit in the
// target's own directory, never the system temp dir.
func TestWriteBoardState_NoLeftoverTempFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")

	if err := WriteBoardState(root, "leader-sess", func(st *BoardState) error {
		st.Cards = append(st.Cards, Card{SpecID: "SPEC-KB-0012", Column: "plan"})
		return nil
	}); err != nil {
		t.Fatalf("WriteBoardState: %v", err)
	}
	entries, err := os.ReadDir(BoardDir(root))
	if err != nil {
		t.Fatalf("read board dir: %v", err)
	}
	// board.json, board.lock, and the roles/ carrier are legitimate residents;
	// anything else — a .board-*.tmp — is a leaked temp file.
	allowed := map[string]bool{"board.json": true, "board.lock": true, "roles": true}
	leaked := []string{}
	for _, e := range entries {
		if !allowed[e.Name()] {
			leaked = append(leaked, e.Name())
		}
	}
	if len(leaked) > 0 {
		t.Fatalf("board dir holds leaked files %v — the temp file must be renamed away, not left behind", leaked)
	}
}

// TestWriteBoardState_ConcurrentReaderSeesWholeBoards — AC-KB-018's dynamic
// half: while writes proceed, a reader repeatedly reading the board observes
// only WHOLE boards — every read parses and yields a complete state, never a
// prefix. The reader is the vantage point because the writer cannot observe
// its own torn write.
func TestWriteBoardState_ConcurrentReaderSeesWholeBoards(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")

	const writes = 60
	var wg sync.WaitGroup
	wg.Add(2)

	writeErr := make(chan error, 1)
	go func() {
		defer wg.Done()
		for i := 0; i < writes; i++ {
			spec := "SPEC-KB-00" + string(rune('A'+i%26)) + "-" + time.Now().Format("150405.000000000")
			err := WriteBoardState(root, "leader-sess", func(st *BoardState) error {
				st.Cards = append(st.Cards, Card{SpecID: spec, Column: "plan"})
				return nil
			})
			if err != nil {
				writeErr <- err
				return
			}
		}
		writeErr <- nil
	}()

	readFailures := make(chan error, 1)
	go func() {
		defer wg.Done()
		previous := -1
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			st, err := LoadBoard(root)
			if err != nil {
				if IsBoardUnknown(err) {
					readFailures <- err
					return
				}
				readFailures <- err
				return
			}
			if len(st.Cards) < previous {
				readFailures <- os.ErrInvalid
				return
			}
			previous = len(st.Cards)
			select {
			case werr := <-writeErr:
				// Writer done: one final read, then stop.
				if werr != nil {
					readFailures <- werr
					return
				}
				readFailures <- nil
				return
			default:
			}
		}
		readFailures <- nil
	}()

	wg.Wait()
	if err := <-readFailures; err != nil {
		t.Fatalf("concurrent reader observed a non-whole board or the writer failed: %v", err)
	}
}

// TestTransitionIntoRun_WipBound — REQ-KB-009's bound, single-threaded, ahead
// of the M2 admission it becomes: at most two cards occupy `run`, the third
// transition is refused with a NAMED error and the board is byte-unchanged,
// and the positive control proves the refusal is conditional (bound 2, not
// 0): with one card in run, the same transition succeeds.
func TestTransitionIntoRun_WipBound(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")

	// One card already in run (positive control baseline).
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-KB-0020", Column: "run", Holder: "runner-sess-1", LastMovedAt: "2026-08-14T00:00:00Z"},
	}})

	// Positive control: with one card in run, the second transition succeeds.
	if err := TransitionIntoRun(root, "leader-sess", "SPEC-KB-0021"); err != nil {
		t.Fatalf("second transition into run refused: %v — the bound is 2, not 0", err)
	}

	before := readBoardBytes(t, root)

	// The third is refused with the named WIP error, board unchanged.
	err := TransitionIntoRun(root, "leader-sess", "SPEC-KB-0022")
	if err == nil {
		t.Fatal("third transition into run err = nil, want ErrWipLimitExceeded")
	}
	if !IsWipLimitExceeded(err) {
		t.Fatalf("err = %v, want ErrWipLimitExceeded (a named refusal)", err)
	}
	if string(before) != string(readBoardBytes(t, root)) {
		t.Fatal("board changed on WIP refusal — it must be left unchanged")
	}

	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	runCount := 0
	for _, c := range st.Cards {
		if c.Column == "run" {
			runCount++
		}
	}
	if runCount != 2 {
		t.Fatalf("cards in run = %d, want 2 — never three", runCount)
	}
}

// readBoardBytes reads the raw board state file bytes for byte-equality checks.
func readBoardBytes(t *testing.T, root string) []byte {
	t.Helper()
	raw, err := os.ReadFile(BoardPath(root))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read board bytes: %v", err)
	}
	return raw
}
