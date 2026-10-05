package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/chain"
	"github.com/modu-ai/moai-adk/internal/config"
)

// TestChainEventHook verifies AC-CHAIN-012: the chain-event hook appends a
// completion edge when a SubagentStop fires with a known session_id.
func TestChainEventHook(t *testing.T) {
	dir := t.TempDir()
	chainDir := filepath.Join(dir, ".moai", "state", "chain")
	_ = os.MkdirAll(chainDir, 0o755)

	// Populate ledger with a node that has a known session_id.
	storePath := filepath.Join(chainDir, "events.jsonl")
	store, _ := chain.NewStore(storePath)
	_ = store.Append(chain.ChainEvent{
		EventType:    chain.EventNodeEnter,
		NodeID:       "N1",
		WorktreePath: "/tmp/wt-1",
		ParentNodeID: "N0",
		Depth:        1,
		SessionID:    "sess-hook-test",
		Milestone:    "M2",
		ResumeTarget: "Continue M3",
	})

	// Simulate SubagentStop payload.
	payload := chainEventPayload{
		SessionID:  "sess-hook-test",
		ProjectDir: dir,
		CWD:        "/tmp/wt-1",
	}
	payloadJSON, _ := json.Marshal(payload)

	t.Setenv(config.EnvChainNodeID, "N1")

	err := RunChainEvent(strings.NewReader(string(payloadJSON)))
	if err != nil {
		t.Fatalf("RunChainEvent: %v", err)
	}

	// Verify the completion edge was appended.
	events, _ := store.ReadAll()
	var foundEdge bool
	for _, ev := range events {
		if ev.EventType == chain.EventCompletionEdge {
			foundEdge = true
			if ev.ParentNode != "N0" {
				t.Errorf("ParentNode = %q, want N0", ev.ParentNode)
			}
			if ev.ChildNode != "N1" {
				t.Errorf("ChildNode = %q, want N1", ev.ChildNode)
			}
			if ev.CompletedMilestone != "M2" {
				t.Errorf("CompletedMilestone = %q, want M2", ev.CompletedMilestone)
			}
			if ev.NextResumeTarget != "Continue M3" {
				t.Errorf("NextResumeTarget = %q, want Continue M3", ev.NextResumeTarget)
			}
		}
	}
	if !foundEdge {
		t.Error("no completion-edge event found in ledger")
	}
}

// TestChainEventHookNoChainContext verifies fail-open: when no chain node
// matches, the hook is a silent no-op.
func TestChainEventHookNoChainContext(t *testing.T) {
	dir := t.TempDir()

	payload := chainEventPayload{
		SessionID:  "nonexistent-session",
		ProjectDir: dir,
		CWD:        "/tmp/nonexistent",
	}
	payloadJSON, _ := json.Marshal(payload)

	_ = os.Unsetenv(config.EnvChainNodeID)

	err := RunChainEvent(strings.NewReader(string(payloadJSON)))
	if err != nil {
		t.Fatalf("RunChainEvent should return nil on no-chain-context, got: %v", err)
	}
}

// TestChainEventHookMalformedPayload verifies fail-open on malformed JSON.
func TestChainEventHookMalformedPayload(t *testing.T) {
	err := RunChainEvent(strings.NewReader("{this is not json"))
	if err != nil {
		t.Fatalf("RunChainEvent should return nil on malformed payload, got: %v", err)
	}
}

// TestChainEventHookEmptyInput verifies fail-open on empty stdin.
func TestChainEventHookEmptyInput(t *testing.T) {
	err := RunChainEvent(strings.NewReader(""))
	if err != nil {
		t.Fatalf("RunChainEvent should return nil on empty input, got: %v", err)
	}
}

// TestChainEventHookNoProjectDir verifies fail-open when no project dir is
// resolvable.
func TestChainEventHookNoProjectDir(t *testing.T) {
	_ = os.Unsetenv(config.EnvClaudeProjectDir)
	payload := chainEventPayload{
		SessionID: "sess",
	}
	payloadJSON, _ := json.Marshal(payload)

	err := RunChainEvent(strings.NewReader(string(payloadJSON)))
	if err != nil {
		t.Fatalf("RunChainEvent should return nil, got: %v", err)
	}
}

// TestChainEventHookUsesLastCompletedMilestone verifies the hook prefers
// LastCompletedMilestone over Milestone for the completed field.
func TestChainEventHookUsesLastCompletedMilestone(t *testing.T) {
	dir := t.TempDir()
	chainDir := filepath.Join(dir, ".moai", "state", "chain")
	_ = os.MkdirAll(chainDir, 0o755)

	storePath := filepath.Join(chainDir, "events.jsonl")
	store, _ := chain.NewStore(storePath)
	_ = store.Append(chain.ChainEvent{
		EventType:              chain.EventNodeEnter,
		NodeID:                 "N1",
		WorktreePath:           "/tmp/wt-ms",
		Depth:                  1,
		SessionID:              "sess-ms",
		Milestone:              "M3",
		LastCompletedMilestone: "M2-done",
	})

	payload := chainEventPayload{
		SessionID:  "sess-ms",
		ProjectDir: dir,
		CWD:        "/tmp/wt-ms",
	}
	payloadJSON, _ := json.Marshal(payload)
	t.Setenv(config.EnvChainNodeID, "N1")

	_ = RunChainEvent(strings.NewReader(string(payloadJSON)))

	events, _ := store.ReadAll()
	for _, ev := range events {
		if ev.EventType == chain.EventCompletionEdge {
			if ev.CompletedMilestone != "M2-done" {
				t.Errorf("CompletedMilestone = %q, want M2-done (prefers LastCompletedMilestone)", ev.CompletedMilestone)
			}
			return
		}
	}
	t.Error("no completion-edge event found")
}
