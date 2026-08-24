package sessionmsg

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Edge-behavior tests for the broker's error surfaces: structured error
// rendering, idempotent heartbeat, ack of never-claimed messages, defensive
// sweep handling, corrupt store files, and lock acquisition bounds.

func TestUnknownAgentErrorMessage(t *testing.T) {
	e := &UnknownAgentError{
		AgentID: "codex-00000000",
		Role:    "receiver",
		Known: []AgentInfo{
			{AgentID: "claude-11111111", Kind: KindClaude, Name: "lead"},
			{AgentID: "codex-22222222", Kind: KindCodex, Name: "worker"},
		},
	}
	msg := e.Error()
	for _, want := range []string{"codex-00000000", "receiver", "claude/lead(claude-11111111)", "codex/worker(codex-22222222)"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message %q missing %q", msg, want)
		}
	}
	// Empty role falls back to the generic label.
	e2 := &UnknownAgentError{AgentID: "x"}
	if !strings.Contains(e2.Error(), "unknown agent") {
		t.Errorf("empty-role fallback missing: %q", e2.Error())
	}
}

func TestHeartbeatIdempotentOnMissing(t *testing.T) {
	s := NewStore(t.TempDir(), &FakeClock{Current: time.Now().UTC()})
	// Heartbeating an agent that never registered is a no-op (registry
	// precedent — no error, no file created).
	if err := s.heartbeat("claude-00000000"); err != nil {
		t.Fatalf("heartbeat on missing agent: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.root, "agents", "claude-00000000.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("heartbeat on missing agent created a record (err=%v)", err)
	}
}

func TestAckPendingMessage(t *testing.T) {
	root := t.TempDir()
	clk := &FakeClock{Current: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)
	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}
	id1, err := s.Send(sender.AgentID, receiver.AgentID, "one", nil, "", "")
	if err != nil {
		t.Fatalf("send 1: %v", err)
	}
	id2, err := s.Send(sender.AgentID, receiver.AgentID, "two", nil, "", "")
	if err != nil {
		t.Fatalf("send 2: %v", err)
	}

	// Ack id1 while it is still pending (never claimed): the ack deletes it
	// from pending, and the same poll delivers only id2.
	res, err := s.Poll(receiver.AgentID, []string{id1})
	if err != nil {
		t.Fatalf("poll with pending ack: %v", err)
	}
	if res.AckedCount != 1 {
		t.Errorf("ack count = %d, want 1", res.AckedCount)
	}
	if len(res.Messages) != 1 || res.Messages[0].Message.MessageID != id2 {
		t.Errorf("poll after pending ack delivered wrong batch: %+v", res.Messages)
	}

	// Acking an unknown id is tolerated (nothing to delete). The id must
	// still be WELL-FORMED — a malformed one is rejected outright (ids.go);
	// this fixture is a valid msg-<hex16> that was simply never minted.
	res2, err := s.Poll(receiver.AgentID, []string{"msg-0123456789abcdef"})
	if err != nil {
		t.Fatalf("poll with unknown ack: %v", err)
	}
	if res2.AckedCount != 0 {
		t.Errorf("unknown ack counted as deleted: %d", res2.AckedCount)
	}
}

func TestSweepRedeemsClaimedWithoutTimestamp(t *testing.T) {
	root := t.TempDir()
	clk := &FakeClock{Current: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)
	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}
	id, err := s.Send(sender.AgentID, receiver.AgentID, "handmade", nil, "", "")
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	// Hand-move the envelope into claimed/ with no claimedAt timestamp (the
	// defensive sweep path: a nil ClaimedAt reads as infinitely old, so the
	// claim TTL is exceeded and the message returns to pending).
	pendingPath := filepath.Join(root, "mailbox", receiver.AgentID, "pending", id+".json")
	claimedPath := filepath.Join(root, "mailbox", receiver.AgentID, "claimed", id+".json")
	if err := os.MkdirAll(filepath.Dir(claimedPath), 0o755); err != nil {
		t.Fatalf("mkdir claimed: %v", err)
	}
	if err := os.Rename(pendingPath, claimedPath); err != nil {
		t.Fatalf("hand-move to claimed: %v", err)
	}

	res, err := s.Poll(receiver.AgentID, nil)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(res.Messages) != 1 || res.Messages[0].Message.MessageID != id {
		t.Fatalf("timestamp-less claimed envelope not redeemed and re-delivered: %+v", res.Messages)
	}
	if res.Messages[0].Delivery.ClaimedAt == nil {
		t.Errorf("re-claim did not stamp claimedAt")
	}
}

func TestCorruptStoreFiles(t *testing.T) {
	root := t.TempDir()
	s := NewStore(root, &FakeClock{Current: time.Now().UTC()})

	// A non-JSON file in agents/ is skipped; a corrupt .json makes listing
	// fail loudly rather than silently dropping agents.
	agentsDir := filepath.Join(root, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "notes.txt"), []byte("not json"), 0o644); err != nil {
		t.Fatalf("write notes.txt: %v", err)
	}
	if _, err := s.ListAgents(); err != nil {
		t.Fatalf("listing with non-json bystander failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(agentsDir, "bad.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write bad.json: %v", err)
	}
	if _, err := s.ListAgents(); err == nil {
		t.Error("listing with corrupt agent record did not fail")
	}

	// Register hits the same corrupt-file failure inside its critical
	// section.
	if _, err := s.Register(KindClaude, "x", ""); err == nil {
		t.Error("register with corrupt agents dir did not fail")
	}
}

func TestSendWithCorruptCounterpartyRecord(t *testing.T) {
	root := t.TempDir()
	s := NewStore(root, &FakeClock{Current: time.Now().UTC()})
	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}

	// Corrupt the sender record: Send must surface the raw read error, not
	// classify it as unknown-agent.
	senderPath := filepath.Join(root, "agents", sender.AgentID+".json")
	if err := os.WriteFile(senderPath, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("corrupt sender record: %v", err)
	}
	_, err = s.Send(sender.AgentID, receiver.AgentID, "x", nil, "", "")
	if err == nil {
		t.Fatal("send with corrupt sender record accepted")
	}
	var unknown *UnknownAgentError
	if errors.As(err, &unknown) {
		t.Errorf("corrupt-record error misclassified as UnknownAgentError: %v", err)
	}

	// Restore the sender record, corrupt the receiver: same classification rule.
	if err := writeJSONAtomic(senderPath, sender); err != nil {
		t.Fatalf("restore sender record: %v", err)
	}
	receiverPath := filepath.Join(root, "agents", receiver.AgentID+".json")
	if err := os.WriteFile(receiverPath, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("corrupt receiver record: %v", err)
	}
	_, err = s.Send(sender.AgentID, receiver.AgentID, "x", nil, "", "")
	if err == nil {
		t.Fatal("send with corrupt receiver record accepted")
	}
	if errors.As(err, &unknown) {
		t.Errorf("corrupt-receiver error misclassified as UnknownAgentError: %v", err)
	}
}

func TestUnknownAgentErrorWhenListingFails(t *testing.T) {
	root := t.TempDir()
	s := NewStore(root, &FakeClock{Current: time.Now().UTC()})
	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	// Corrupt a bystander record so the known-agents listing inside
	// unknownAgentError fails: the structured error must still be returned
	// (with an empty known list), never a bare listing failure.
	if err := os.WriteFile(filepath.Join(root, "agents", "bad.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatalf("corrupt bystander: %v", err)
	}
	_, err = s.Send(sender.AgentID, "codex-00000000", "x", nil, "", "")
	if err == nil {
		t.Fatal("send to unknown receiver accepted")
	}
	var unknown *UnknownAgentError
	if !errors.As(err, &unknown) {
		t.Errorf("error is not UnknownAgentError when listing fails: %v", err)
	}
	if len(unknown.Known) != 0 {
		t.Errorf("known list should be empty on listing failure: %+v", unknown.Known)
	}
}

func TestWithAgentLockTimeout(t *testing.T) {
	root := t.TempDir()
	s := NewStore(root, nil).WithLockTimeout(30 * time.Millisecond)
	// Make the lock path un-acquirable (a directory cannot be flock-opened
	// O_RDWR): acquisition retries, then times out with ErrLockTimeout.
	lockPath := s.lockPath(lockNameRegister)
	if err := os.MkdirAll(lockPath, 0o755); err != nil {
		t.Fatalf("mkdir lock path: %v", err)
	}
	err := s.withAgentLock(lockNameRegister, func() error {
		return errors.New("must not run")
	})
	if !errors.Is(err, ErrLockTimeout) {
		t.Fatalf("expected ErrLockTimeout, got %v", err)
	}
}

func TestLockRetryDelayBounds(t *testing.T) {
	if d := lockRetryDelay(0, time.Minute); d <= 0 || d > 5*time.Millisecond {
		t.Errorf("attempt 0 delay %v outside (0, 5ms]", d)
	}
	if d := lockRetryDelay(50, time.Minute); d <= 0 || d > 50*time.Millisecond {
		t.Errorf("saturated attempt delay %v outside (0, 50ms]", d)
	}
	if d := lockRetryDelay(0, 0); d != 0 {
		t.Errorf("no remaining budget must yield 0, got %v", d)
	}
	if d := lockRetryDelay(0, -time.Second); d != 0 {
		t.Errorf("negative remaining budget must yield 0, got %v", d)
	}
	if d := lockRetryDelay(0, 1*time.Millisecond); d > 1*time.Millisecond {
		t.Errorf("delay %v exceeds remaining 1ms (clamp failed)", d)
	}
}

func TestAgentLockAcquireRelease(t *testing.T) {
	root := t.TempDir()
	lockPath := filepath.Join(root, "test.lock")

	l := newAgentLock()
	// Release before acquire is idempotent.
	if err := l.release(); err != nil {
		t.Fatalf("release before acquire: %v", err)
	}
	if err := l.acquire(lockPath); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := l.release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	// Double release stays idempotent.
	if err := l.release(); err != nil {
		t.Fatalf("double release: %v", err)
	}
	// Acquire on a directory path fails (un-acquirable path).
	if err := os.MkdirAll(filepath.Join(root, "dir.lock"), 0o755); err != nil {
		t.Fatalf("mkdir dir lock: %v", err)
	}
	if err := newAgentLock().acquire(filepath.Join(root, "dir.lock")); err == nil {
		t.Error("acquire on directory path unexpectedly succeeded")
	}
}

func TestWriteJSONAtomicFailures(t *testing.T) {
	root := t.TempDir()
	// Marshal failure: a channel cannot be serialized.
	if err := writeJSONAtomic(filepath.Join(root, "ok.json"), map[string]any{"ch": make(chan int)}); err == nil {
		t.Error("marshal failure not surfaced")
	}
	// Mkdir failure: parent path is a regular file.
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	if err := writeJSONAtomic(filepath.Join(blocker, "sub", "f.json"), map[string]string{}); err == nil {
		t.Error("mkdir failure not surfaced")
	}
	// No temp files are left behind on success.
	okPath := filepath.Join(root, "ok2.json")
	if err := writeJSONAtomic(okPath, map[string]string{"a": "b"}); err != nil {
		t.Fatalf("write ok2: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read root: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
	// Round-trip: written file is valid JSON with a trailing newline.
	data, err := os.ReadFile(okPath)
	if err != nil {
		t.Fatalf("read ok2: %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Errorf("written file lacks trailing newline: %q", data)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("written file is not valid JSON: %v", err)
	}
}
