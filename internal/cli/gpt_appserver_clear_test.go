package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

func TestManagedGPTClearSessionIsIsolatedAndAttested(t *testing.T) {
	home, _ := filepath.EvalSymlinks(t.TempDir())
	old := "11111111-1111-4111-8111-111111111111"
	next := "22222222-2222-4222-8222-222222222222"
	c := gatewayPrivateConversation{FamilyID: old, SessionID: old, ReceiptDir: filepath.Join(home, old, "receipt"), CWD: "/project"}
	native := filepath.Join(home, old, "native")
	dir := filepath.Join(native, "projects", "project")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	headers := http.Header{"X-Claude-Code-Session-Id": []string{next}}
	identity, _ := json.Marshal(map[string]string{"session_id": next, "device_id": "device"})
	body, _ := json.Marshal(map[string]any{"metadata": map[string]string{"user_id": string(identity)}})
	if _, err := managedGPTConversationForRequest(context.Background(), c, headers, body); err == nil {
		t.Fatal("unattested UUID accepted")
	}
	command := `{"type":"user","uuid":"clear","sessionId":"` + next + `","cwd":"/project","isSidechain":false,"message":{"role":"user","content":"<command-name>/clear</command-name> <command-message>clear</command-message> <command-args></command-args>"}}`
	complete := `{"type":"system","subtype":"local_command","parentUuid":"clear","sessionId":"` + next + `","cwd":"/project","isSidechain":false,"content":"<local-command-stdout></local-command-stdout>"}`
	if err := os.WriteFile(filepath.Join(dir, next+".jsonl"), []byte(command+"\n"+complete+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := managedGPTConversationForRequest(context.Background(), c, headers, body)
	if err != nil || got != next {
		t.Fatalf("clear: got %q, err %v", got, err)
	}
	headers.Set("X-Claude-Code-Session-Id", old)
	if _, err := managedGPTConversationForRequest(context.Background(), c, headers, body); err == nil {
		t.Fatal("header/metadata mismatch accepted")
	}
	// Exercise the real preparation entry point, not only identity parsing.
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(context.Background(), entry)
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare(old, c.CWD, translate.Limits{PolicyProfile: translate.PolicyGPTNative}, managedGPTTestStore(t), c)
	plain := []byte(`{"max_tokens":10,"messages":[{"role":"user","content":"hello"}]}`)
	before, err := prepare(context.Background(), gateway.RoutedRequest{Entry: entry, Managed: grant, Body: plain, Headers: headers})
	if err != nil {
		t.Fatal(err)
	}
	headers.Set("X-Claude-Code-Session-Id", next)
	after, err := prepare(context.Background(), gateway.RoutedRequest{Entry: entry, Managed: grant, Body: plain, Headers: headers})
	if err != nil {
		t.Fatal(err)
	}
	if before.Owner.ConversationID != old || after.Owner.ConversationID != next {
		t.Fatalf("clear reused owner: old=%s new=%s", before.Owner.ConversationID, after.Owner.ConversationID)
	}
}
