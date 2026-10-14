// factory_messages_lock_test.go — SPEC-MCP-NEUTRALITY-001 M3 (card t1632,
// REQ-MN-011): the inbox COUNTING path is read-only. An empty inbox answers
// "empty-or-receipt-only" WITHOUT taking the broker's write lock — the old
// path paid CheckWritable's BEGIN IMMEDIATE probe (and its deadline) on every
// inspection, which is the 1,619-recorded "inbox context deadline exceeded"
// defect family. A held writer must therefore not disturb the empty answer.
package hook

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryHookEmptyInboxSkipsWriteLock pins REQ-MN-011: with the broker's
// write lock held by another connection for the whole inspection window, an
// empty inbox still answers "empty-or-receipt-only" — the counting path is a
// read and never touches CheckWritable. The old path degraded with the
// write-probe failure.
func TestFactoryHookEmptyInboxSkipsWriteLock(t *testing.T) {
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=active")
	owner, start := factoryHookOwnerIdentity(t)
	ctx := context.Background()

	s, err := factorymsg.Open(root, "runX")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeFactoryHookStore(s) })
	if _, err := s.RegisterPeer(ctx, factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: "runX", Backend: "claude", Role: "lane", Slot: "lane-3", SessionUUID: "s1", Generation: 1, PID: owner, ProcessStart: start}); err != nil {
		t.Fatal(err)
	}

	// Hold the write lock from a second connection across the inspection:
	// BEGIN IMMEDIATE stays open until cleanup, so any write probe this
	// inspection attempts blocks past its budget.
	brokerPath, err := factorymsg.BrokerPath(root, "runX")
	if err != nil {
		t.Fatal(err)
	}
	locker, err := sql.Open("sqlite", "file:"+filepath.ToSlash(brokerPath)+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = locker.Close() })
	if _, err := locker.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		t.Fatalf("hold the broker write lock: %v", err)
	}
	defer locker.ExecContext(ctx, `ROLLBACK`)

	msg, surface, state := factoryHookBatchForRun(ctx, &HookInput{SessionID: "s1", ProjectDir: root}, EventStop, "")
	_ = msg
	_ = surface
	if state != "empty-or-receipt-only" {
		t.Fatalf("empty inbox under a held write lock: state = %q, want %q (the counting path must not take the write lock)", state, "empty-or-receipt-only")
	}
}
