package sessionmsg

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/config"
)

// DefaultStateRoot is the canonical project-relative path of the session
// messaging broker's file store. It is covered by the .moai/state/ blanket
// gitignore rule. Tests pass a t.TempDir() root via NewStore so they never
// touch the real state directory (plan.md B8/B10).
const DefaultStateRoot = ".moai/state/session-msg"

// LockTimeout is the default maximum wait for advisory lock acquisition,
// mirroring the internal/session registry precedent (design.md §5.2).
// Beyond this, broker operations return ErrLockTimeout.
const LockTimeout = 2 * time.Second

// ErrLockTimeout is returned when mailbox/registration lock acquisition
// exceeds the configured timeout.
var ErrLockTimeout = errors.New("sessionmsg: lock acquisition timed out")

// Clock is the time abstraction so tests can drive TTL and heartbeat
// behavior deterministically with a FakeClock (AP-MSC-004 precedent).
type Clock interface {
	Now() time.Time
}

// realClock returns time.Now() in UTC.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// FakeClock is a deterministic clock for tests. Now() returns Current until
// the test reassigns it. A FakeClock shared across goroutines must not be
// mutated while those goroutines call Now() (data race).
type FakeClock struct {
	Current time.Time
}

// Now returns the current FakeClock value.
func (f *FakeClock) Now() time.Time { return f.Current }

// Store is the broker's file-store handle bound to a state root. All state
// lives under root: agents/<agentId>.json, mailbox/<agentId>/{pending,claimed}/,
// and locks/. The zero mutation surface (root+clock+lockTimeout only) makes
// a Store safe to share across goroutines.
type Store struct {
	root        string
	clock       Clock
	lockTimeout time.Duration
}

// NewStore constructs a Store bound to the given state root. The root may be
// relative (resolved against CWD) or absolute; pass DefaultStateRoot for the
// production location or a t.TempDir() in tests. A nil clock selects the
// real UTC clock.
func NewStore(root string, clock Clock) *Store {
	if clock == nil {
		clock = realClock{}
	}
	return &Store{root: root, clock: clock, lockTimeout: LockTimeout}
}

// WithLockTimeout returns a copy of s with the advisory-lock acquisition
// timeout overridden (tests with high contention may extend it).
func (s *Store) WithLockTimeout(d time.Duration) *Store {
	clone := *s
	clone.lockTimeout = d
	return &clone
}

// PollResult is the outcome of one Poll call (REQ-CSM-006): the claimed
// batch, the number of messages still pending after the call, the number of
// messages the lazy sweep deleted as expired, and the number of ack_ids
// actually deleted.
type PollResult struct {
	Messages     []Envelope
	Remaining    int
	ExpiredCount int
	AckedCount   int
}

// State-layout path helpers (design.md §2). agentId ("<kind>-<hex8>") and
// broker-minted messageIds are filename-safe by construction.
func (s *Store) agentsDir() string           { return filepath.Join(s.root, "agents") }
func (s *Store) agentPath(id string) string  { return filepath.Join(s.root, "agents", id+".json") }
func (s *Store) mailboxDir(id string) string { return filepath.Join(s.root, "mailbox", id) }
func (s *Store) pendingDir(id string) string {
	return filepath.Join(s.mailboxDir(id), "pending")
}
func (s *Store) claimedDir(id string) string {
	return filepath.Join(s.mailboxDir(id), "claimed")
}
func (s *Store) lockPath(name string) string {
	return filepath.Join(s.root, "locks", name+".lock")
}

// writeJSONAtomic marshals v to indented JSON and writes it via
// temp-file + atomicfile.Replace (REQ-CSM-009: every write atomic — a
// concurrent reader never observes a partial file).
func writeJSONAtomic(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("sessionmsg: marshal %s: %w", path, err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("sessionmsg: mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*.tmp")
	if err != nil {
		return fmt.Errorf("sessionmsg: create temp in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("sessionmsg: write temp %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("sessionmsg: close temp %s: %w", tmpPath, err)
	}
	if err := atomicfile.Replace(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("sessionmsg: rename temp -> %s: %w", path, err)
	}
	return nil
}

// UnknownAgentError is the structured error returned when a Send or Poll
// names an agent that is not registered (REQ-CSM-005): the Known list lets
// the caller (and the M2 MCP handler) surface the registered agents instead
// of a bare failure.
type UnknownAgentError struct {
	AgentID string      `json:"agentId"`
	Role    string      `json:"role"` // "sender" | "receiver" | "agent"
	Known   []AgentInfo `json:"knownAgents"`
}

// Error implements error with a compact known-agents rendering.
func (e *UnknownAgentError) Error() string {
	names := make([]string, 0, len(e.Known))
	for _, k := range e.Known {
		names = append(names, fmt.Sprintf("%s/%s(%s)", k.Kind, k.Name, k.AgentID))
	}
	role := e.Role
	if role == "" {
		role = "agent"
	}
	return fmt.Sprintf("sessionmsg: unknown %s agent %q; known agents: %s", role, e.AgentID, strings.Join(names, ", "))
}

// newMessageID mints a broker message id ("msg-<hex16>"), filename-safe by
// construction.
func newMessageID() (string, error) {
	b := make([]byte, 8) // 8 bytes → 16 hex chars
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("sessionmsg: message id entropy: %w", err)
	}
	return "msg-" + hex.EncodeToString(b), nil
}

// unknownAgentError builds the structured UnknownAgentError, best-effort
// loading the known-agent list (REQ-CSM-005).
func (s *Store) unknownAgentError(agentID, role string) error {
	known, err := s.ListAgents()
	if err != nil {
		known = nil
	}
	return &UnknownAgentError{AgentID: agentID, Role: role, Known: known}
}

// Send validates and persists a message from one registered agent to another
// (REQ-CSM-005): both counterparties must be registered, the built message
// must pass envelope validation (text and part-count ceilings), and the
// envelope is written atomically to the recipient's
// mailbox/<agentId>/pending/<messageId>.json under the recipient's advisory
// lock. The sender's heartbeat is refreshed after the mailbox scope ends
// (REQ-CSM-004) — locks are never nested.
//
// @MX:ANCHOR: [AUTO] broker send — validated atomic enqueue into recipient pending
// @MX:REASON: fan_in >= 3 (M2 session_msg_send handler, e2e both directions, concurrency test). Envelope layout or validation changes here alter every receiver's on-disk contract.
func (s *Store) Send(fromAgentID, toAgentID, text string, data json.RawMessage, contextID, taskID string) (string, error) {
	sender, err := s.readAgent(fromAgentID)
	if err != nil {
		if os.IsNotExist(err) {
			return "", s.unknownAgentError(fromAgentID, "sender")
		}
		return "", err
	}
	if _, err := s.readAgent(toAgentID); err != nil {
		if os.IsNotExist(err) {
			return "", s.unknownAgentError(toAgentID, "receiver")
		}
		return "", err
	}

	now := s.clock.Now().UTC()
	msgID, err := newMessageID()
	if err != nil {
		return "", err
	}

	parts := make([]Part, 0, 2)
	if text != "" {
		parts = append(parts, Part{Kind: PartKindText, Text: text})
	}
	if len(data) > 0 {
		parts = append(parts, Part{Kind: PartKindData, Data: append(json.RawMessage(nil), data...)})
	}
	env := Envelope{
		Message: Message{
			MessageID: msgID,
			ContextID: contextID,
			TaskID:    taskID,
			Role:      RoleAgent,
			Parts:     parts,
		},
		Delivery: Delivery{
			SenderID:   sender.AgentID,
			SenderKind: sender.Kind,
			SentAt:     now,
			ExpiresAt:  now.Add(config.DefaultSessionMsgMessageTTL),
		},
	}
	if err := env.Validate(); err != nil {
		return "", err
	}

	err = s.withAgentLock(toAgentID, func() error {
		return writeJSONAtomic(filepath.Join(s.pendingDir(toAgentID), msgID+".json"), env)
	})
	if err != nil {
		return "", err
	}

	// Sender-side heartbeat refresh, sequenced after the mailbox scope so no
	// lock ever nests inside another (design.md §5.2 deadlock avoidance).
	if err := s.heartbeat(fromAgentID); err != nil {
		return "", err
	}
	return msgID, nil
}

// Poll claims a batch of pending messages for one agent (REQ-CSM-006),
// applying the lazy sweep first (REQ-CSM-007 claim-TTL redemption,
// REQ-CSM-008 message-TTL deletion), then ack_ids deletion, then the
// pending→claimed move up to config.DefaultSessionMsgPollBatch — all under
// the agent's mailbox advisory lock. The agent's heartbeat is refreshed
// after the mailbox scope ends (REQ-CSM-004).
//
// @MX:ANCHOR: [AUTO] broker poll — lazy sweep + atomic batch claim + ack deletion
// @MX:REASON: fan_in >= 3 (M2 session_msg_poll handler, e2e receive path on both session kinds, concurrency test). Claim/sweep semantics define the broker's at-least-once delivery contract.
func (s *Store) Poll(agentID string, ackIDs []string) (PollResult, error) {
	if _, err := s.readAgent(agentID); err != nil {
		if os.IsNotExist(err) {
			return PollResult{}, s.unknownAgentError(agentID, "agent")
		}
		return PollResult{}, err
	}

	var result PollResult
	err := s.withAgentLock(agentID, func() error {
		now := s.clock.Now().UTC()
		claimedDir := s.claimedDir(agentID)
		pendingDir := s.pendingDir(agentID)

		// Sweep 1 — claimed envelopes (REQ-CSM-007 + REQ-CSM-008): expired
		// ones are deleted; claim-TTL-exceeded ones return to pending so a
		// dead receiver's unacked messages are re-delivered (at-least-once).
		claimed, err := listEnvelopes(claimedDir)
		if err != nil {
			return err
		}
		for _, env := range claimed {
			id := env.Message.MessageID
			if now.After(env.Delivery.ExpiresAt) {
				if err := removeIfExists(filepath.Join(claimedDir, id+".json")); err != nil {
					return err
				}
				result.ExpiredCount++
				continue
			}
			var claimedAt time.Time
			if env.Delivery.ClaimedAt != nil {
				claimedAt = *env.Delivery.ClaimedAt
			}
			if now.Sub(claimedAt) > config.DefaultSessionMsgClaimTTL {
				env.Delivery.ClaimedAt = nil
				if err := writeJSONAtomic(filepath.Join(pendingDir, id+".json"), env); err != nil {
					return err
				}
				if err := removeIfExists(filepath.Join(claimedDir, id+".json")); err != nil {
					return err
				}
			}
		}

		// Sweep 2 — pending envelopes (REQ-CSM-008): expired ones are deleted.
		pending, err := listEnvelopes(pendingDir)
		if err != nil {
			return err
		}
		for _, env := range pending {
			if now.After(env.Delivery.ExpiresAt) {
				if err := removeIfExists(filepath.Join(pendingDir, env.Message.MessageID+".json")); err != nil {
					return err
				}
				result.ExpiredCount++
			}
		}

		// Ack (REQ-CSM-006 ack_ids): delete each acked id from claimed, then
		// pending. Deleting before the claim step also drops a redeemed
		// message the receiver already processed, so it is not re-delivered.
		for _, id := range ackIDs {
			removed, err := removeFirstOf(
				filepath.Join(claimedDir, id+".json"),
				filepath.Join(pendingDir, id+".json"),
			)
			if err != nil {
				return err
			}
			if removed {
				result.AckedCount++
			}
		}

		// Claim: move up to the batch ceiling from pending to claimed.
		live, err := listEnvelopes(pendingDir)
		if err != nil {
			return err
		}
		batch := config.DefaultSessionMsgPollBatch
		if batch <= 0 {
			batch = 1
		}
		claimCount := len(live)
		if claimCount > batch {
			claimCount = batch
		}
		for i := 0; i < claimCount; i++ {
			env := live[i]
			claimedAt := now
			env.Delivery.ClaimedAt = &claimedAt
			if err := writeJSONAtomic(filepath.Join(claimedDir, env.Message.MessageID+".json"), env); err != nil {
				return err
			}
			if err := removeIfExists(filepath.Join(pendingDir, env.Message.MessageID+".json")); err != nil {
				return err
			}
			result.Messages = append(result.Messages, env)
		}
		result.Remaining = len(live) - claimCount
		return nil
	})
	if err != nil {
		return PollResult{}, err
	}

	// Heartbeat refresh, sequenced after the mailbox scope (no nested locks).
	if err := s.heartbeat(agentID); err != nil {
		return PollResult{}, err
	}
	return result, nil
}

// listEnvelopes loads every *.json envelope in dir (deterministic directory
// order; a missing directory is an empty mailbox).
func listEnvelopes(dir string) ([]Envelope, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("sessionmsg: read %s: %w", dir, err)
	}
	var envs []Envelope
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := atomicfile.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("sessionmsg: read envelope %s: %w", e.Name(), err)
		}
		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return nil, fmt.Errorf("sessionmsg: parse envelope %s: %w", e.Name(), err)
		}
		envs = append(envs, env)
	}
	return envs, nil
}

// removeIfExists removes path, tolerating an already-absent file.
func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("sessionmsg: remove %s: %w", path, err)
	}
	return nil
}

// removeFirstOf removes the first existing path among the candidates and
// reports whether any was removed.
func removeFirstOf(paths ...string) (bool, error) {
	for _, p := range paths {
		if err := os.Remove(p); err == nil {
			return true, nil
		} else if !os.IsNotExist(err) {
			return false, fmt.Errorf("sessionmsg: remove %s: %w", p, err)
		}
	}
	return false, nil
}
