package sessionmsg

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestPollRejectsTraversalAckIDs is the path-safety boundary for ack_ids: a
// caller-supplied messageId is joined into a delete target, so an id carrying
// traversal segments must be rejected BEFORE any filesystem work — and must
// delete nothing outside the state root.
//
// Pre-fix behaviour (observed on f33cd0564): Poll returned err=nil,
// AckedCount=1, and the victim file was gone.
//
// Mutant this test catches that a shallower one would not: an implementation
// that validates ack_ids but only AFTER the sweep/delete loop (or one that
// silently skips the malformed entry and returns nil) still deletes the
// victim — so asserting the error alone is insufficient. Both the error AND
// the victim's survival are asserted.
func TestPollRejectsTraversalAckIDs(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "state")
	victim := filepath.Join(base, "victim.json")
	if err := os.WriteFile(victim, []byte(`{"secret":true}`), 0o600); err != nil {
		t.Fatalf("seed victim: %v", err)
	}

	clk := &FakeClock{Current: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)}
	s := NewStore(root, clk)
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}

	// claimed dir is root/mailbox/<agentId>/claimed — four levels below base.
	traversal := filepath.Join("..", "..", "..", "..", "victim")
	res, err := s.Poll(receiver.AgentID, []string{traversal})
	if err == nil {
		t.Fatalf("Poll accepted traversal ack id %q (result %+v)", traversal, res)
	}
	var invalid *InvalidIDError
	if !errors.As(err, &invalid) {
		t.Errorf("error %v is not an *InvalidIDError", err)
	} else if invalid.Field != "ack_ids" {
		t.Errorf("InvalidIDError.Field = %q, want %q", invalid.Field, "ack_ids")
	}
	if res.AckedCount != 0 {
		t.Errorf("rejected Poll reported AckedCount=%d, want 0", res.AckedCount)
	}
	if _, statErr := os.Stat(victim); statErr != nil {
		t.Fatalf("file OUTSIDE the state root was deleted by a traversal ack id: %v", statErr)
	}
}

// TestPollRejectsMalformedAgentID pins that a caller-supplied agentId is
// shape-checked before it names a mailbox directory.
//
// Mutant this test catches: validating only ack_ids (the more obvious of the
// two path inputs) while leaving agentID free — the traversal agentId would
// then reach agentPath/mailboxDir path construction.
//
// This is why the assertion is errors.As(*InvalidIDError) and NOT merely
// "err != nil": on the pre-fix code EVERY id below already produced an error
// — UnknownAgentError, because no agent record exists at the traversed path.
// A test asserting only non-nil passes on the broken code (verified against
// f33cd0564). Only the error's TYPE distinguishes "rejected before touching
// the filesystem" from "the traversal happened and found nothing".
func TestPollRejectsMalformedAgentID(t *testing.T) {
	s := NewStore(t.TempDir(), &FakeClock{Current: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)})

	for _, id := range []string{
		"../../etc/passwd",
		"codex-..",
		"/absolute/codex-abcd1234",
		"codex-ABCD1234", // uppercase hex is not the minted shape
		"codex-abcd123",  // too short
		"gemini-abcd1234",
		"",
	} {
		_, err := s.Poll(id, nil)
		if err == nil {
			t.Errorf("Poll accepted malformed agentId %q", id)
			continue
		}
		var invalid *InvalidIDError
		if !errors.As(err, &invalid) {
			t.Errorf("Poll(%q) rejected with %T (%v), want *InvalidIDError — the id reached path construction", id, err, err)
		}
	}
}

// TestSendRejectsMalformedAgentIDs pins the same enforcement on the send
// path, where BOTH counterparty ids are caller-supplied.
//
// Mutant this test catches: validating only from_agent_id (which is read
// first) while to_agent_id — the id that actually names the written mailbox
// path — stays unchecked. As in the poll case above, the assertion is on the
// error TYPE: the pre-fix code already returned UnknownAgentError for both,
// so "err != nil" alone would pass on the broken implementation.
func TestSendRejectsMalformedAgentIDs(t *testing.T) {
	s := NewStore(t.TempDir(), &FakeClock{Current: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)})
	sender, err := s.Register(KindClaude, "sender", "")
	if err != nil {
		t.Fatalf("register sender: %v", err)
	}
	receiver, err := s.Register(KindCodex, "receiver", "")
	if err != nil {
		t.Fatalf("register receiver: %v", err)
	}

	cases := []struct {
		name      string
		from, to  string
		wantField string
	}{
		{"traversal from_agent_id", "../../evil", receiver.AgentID, "from_agent_id"},
		{"traversal to_agent_id", sender.AgentID, "../../evil", "to_agent_id"},
	}
	for _, tc := range cases {
		_, err := s.Send(tc.from, tc.to, "x", nil, "", "")
		if err == nil {
			t.Errorf("%s: Send accepted it", tc.name)
			continue
		}
		var invalid *InvalidIDError
		if !errors.As(err, &invalid) {
			t.Errorf("%s: rejected with %T (%v), want *InvalidIDError", tc.name, err, err)
			continue
		}
		if invalid.Field != tc.wantField {
			t.Errorf("%s: InvalidIDError.Field = %q, want %q", tc.name, invalid.Field, tc.wantField)
		}
	}
}

// TestIDShapeHelpers pins the accepted shapes directly, so a regression in
// the patterns surfaces here rather than only through a behavioural test.
func TestIDShapeHelpers(t *testing.T) {
	validAgents := []string{"claude-0123abcd", "codex-ffffffff"}
	for _, id := range validAgents {
		if !validAgentID(id) {
			t.Errorf("validAgentID(%q) = false, want true", id)
		}
	}
	for _, id := range []string{"claude-0123abc", "claude-0123abcde", "claude_0123abcd", "claude-0123abcd\n", "codex-0123ABCD"} {
		if validAgentID(id) {
			t.Errorf("validAgentID(%q) = true, want false", id)
		}
	}

	if !validMessageID("msg-0123456789abcdef") {
		t.Error("validMessageID rejected a well-formed id")
	}
	for _, id := range []string{"msg-0123456789abcde", "msg-0123456789abcdef0", "msg-../../x", "msg-0123456789ABCDEF", "0123456789abcdef"} {
		if validMessageID(id) {
			t.Errorf("validMessageID(%q) = true, want false", id)
		}
	}
}
