package cli

// Card t708 (SPEC-GATEWAY-ENVELOPE-REPAIR-001 M3): launcher wiring tests for
// the explicit --repair-envelope verb. The flag must run the repair before
// the resume flow, be stripped from the child args, refuse without --resume,
// and surface the frozen classified guidance on refusal — never run
// automatically.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
	"github.com/modu-ai/moai-adk/internal/gateway/opaque"
)

func repairWiringFixture(t *testing.T, stripped bool) (m *conversation.Manager, desc conversation.Descriptor, path, root string) {
	t.Helper()
	var err error
	root, err = filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, err = conversation.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	desc, err = m.New(context.Background(), conversation.NewRequest{CWD: root, Project: "proj"})
	if err != nil {
		t.Fatal(err)
	}
	env, err := opaque.Encode([]opaque.Item{{OutputIndex: 0, Raw: []byte(`{"type":"reasoning","id":"rs_1","summary":[],"encrypted_content":"cipher"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	marker, err := opaque.BindToolID("call_1", env)
	if err != nil {
		t.Fatal(err)
	}
	envelopeBlock := func() map[string]any {
		return map[string]any{"type": "redacted_thinking", "data": env.Data()}
	}
	rows := []map[string]any{
		{"type": "user", "sessionId": desc.UUID, "cwd": root, "message": map[string]any{"role": "user", "content": "hello"}},
	}
	if stripped {
		// The verbatim carrier survives in a separate issuance row; the
		// boundary row below was persisted without it (split-record shape).
		rows = append(rows, map[string]any{"type": "assistant", "sessionId": desc.UUID, "cwd": root,
			"message": map[string]any{"role": "assistant", "model": "gpt-5.6-sol", "content": []any{envelopeBlock()}}})
	}
	boundary := []any{map[string]any{"type": "tool_use", "id": marker, "name": "read", "input": map[string]any{}}}
	if !stripped {
		boundary = append([]any{envelopeBlock()}, boundary...)
	}
	rows = append(rows, map[string]any{"type": "assistant", "sessionId": desc.UUID, "cwd": root,
		"message": map[string]any{"role": "assistant", "model": "gpt-5.6-sol", "stop_reason": "end_turn", "content": boundary}})
	var lines []string
	for _, row := range rows {
		raw, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, string(raw))
	}
	path = filepath.Join(desc.ConfigDir, "projects", desc.UUID+".jsonl")
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Complete(context.Background(), desc.UUID, path, uint64(info.ModTime().UnixNano())); err != nil {
		t.Fatal(err)
	}
	return m, desc, path, root
}

func TestRepairEnvelopeFlagRunsBeforeResume(t *testing.T) {
	m, desc, _, root := repairWiringFixture(t, true)
	d, err := prepareGatewayConversation(gatewayLaunchRequest{
		CWD: root, Project: "proj",
		Args: []string{"--resume", desc.UUID, repairGatewayEnvelopeFlag},
	}, m)
	if err != nil {
		t.Fatalf("repaired resume rejected: %v", err)
	}
	if d.UUID != desc.UUID {
		t.Fatalf("descriptor uuid = %s, want %s", d.UUID, desc.UUID)
	}
	statePath := filepath.Join(root, "families", d.FamilyID, "repair", desc.UUID+".json")
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("repair did not run before resume: %v", err)
	}
}

func TestRepairEnvelopeRefusalSurfacesClassifiedGuidance(t *testing.T) {
	m, desc, _, root := repairWiringFixture(t, false)
	_, err := prepareGatewayConversation(gatewayLaunchRequest{
		CWD: root, Project: "proj",
		Args: []string{"--resume", desc.UUID, repairGatewayEnvelopeFlag},
	}, m)
	if err == nil {
		t.Fatal("intact history accepted the repair verb")
	}
	want := "conversation history changed, lacks reasoning, or belongs to another model family/account; start a new conversation"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("refusal lost the frozen classified guidance: %v", err)
	}
}

func TestRepairEnvelopeFlagRequiresResumeUUID(t *testing.T) {
	m, _, _, root := repairWiringFixture(t, true)
	if _, err := prepareGatewayConversation(gatewayLaunchRequest{
		CWD: root, Project: "proj",
		Args: []string{repairGatewayEnvelopeFlag},
	}, m); err == nil || !strings.Contains(err.Error(), "--resume") {
		t.Fatalf("bare repair verb must demand --resume, got %v", err)
	}
}

func TestGatewayConversationPassthroughStripsRepairFlag(t *testing.T) {
	out := gatewayConversationPassthrough([]string{"--model", "gpt-5.6-sol", "--resume", "u", repairGatewayEnvelopeFlag, "--fork-session"})
	joined := strings.Join(out, " ")
	if strings.Contains(joined, repairGatewayEnvelopeFlag) {
		t.Fatalf("repair flag leaked to the child: %v", out)
	}
	for _, kept := range []string{"--model", "gpt-5.6-sol"} {
		if !strings.Contains(joined, kept) {
			t.Fatalf("passthrough dropped %s: %v", kept, out)
		}
	}
}
