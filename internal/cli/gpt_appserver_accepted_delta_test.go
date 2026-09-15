package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

// A completed tool reply followed by an acknowledged local cancellation may
// leave no assistant message between the accepted reply and the next input.
func TestManagedGPTAcceptedToolReplyIsNotResubmittedAfterCancel(t *testing.T) {
	for _, merged := range []bool{false, true} {
		for _, tampered := range []bool{false, true} {
			t.Run(fmt.Sprintf("merged=%v/tampered=%v", merged, tampered), func(t *testing.T) {
				ctx := context.Background()
				authority, _ := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
				entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
				grant, _ := authority.Authorize(ctx, entry)
				dir := t.TempDir()
				if err := os.Chmod(dir, 0700); err != nil {
					t.Fatal(err)
				}
				store, err := codexbridge.OpenStore(dir)
				if err != nil {
					t.Fatal(err)
				}
				cwd := t.TempDir()
				prepare := newManagedGPTPrepare("family", cwd, translate.Limits{PolicyProfile: translate.PolicyGPTNative}, store)
				result := map[string]any{"type": "tool_result", "tool_use_id": "accepted-call", "content": "accepted result"}
				messages := []any{map[string]any{"role": "user", "content": "work"}, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": "accepted-call", "name": "Skill", "input": map[string]any{}}}}, map[string]any{"role": "user", "content": []any{result, map[string]any{"type": "text", "text": "accepted skill context"}}}}
				body := map[string]any{"max_tokens": 1024, "messages": messages}
				raw, _ := json.Marshal(body)
				request := gateway.RoutedRequest{Entry: entry, Managed: grant, Body: raw, Headers: http.Header{"X-Claude-Code-Session-Id": {"family"}}}
				accepted, err := prepare(ctx, request)
				if err != nil {
					t.Fatal(err)
				}
				// Seed the durable idle barrier produced only after native cancel acknowledgement.
				key := sha256.Sum256([]byte(grant.Scope() + "\x00family"))
				barrier, _ := json.Marshal(map[string]any{"Schema": 2, "Owner": map[string]string{"ConversationID": "family", "AccountScope": grant.Scope(), "ThreadID": "native-thread"}, "Phase": "idle", "Prefix": accepted.PrefixDigest, "Model": entry.UpstreamID, "CWD": cwd})
				if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%x.json", key)), barrier, 0600); err != nil {
					t.Fatal(err)
				}
				if tampered {
					result["content"] = "altered result"
				}
				fresh := []any{map[string]any{"type": "text", "text": "new question"}, map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": "aW1hZ2U="}}}
				if merged {
					messages[2].(map[string]any)["content"] = append(messages[2].(map[string]any)["content"].([]any), fresh...)
				} else {
					messages = append(messages, map[string]any{"role": "user", "content": fresh})
				}
				body["messages"] = messages
				request.Body, _ = json.Marshal(body)
				q, err := prepare(ctx, request)
				if err != nil {
					t.Fatal(err)
				}
				if tampered {
					if len(q.Results) == 0 {
						t.Fatal("unproven tool result discarded")
					}
					return
				}
				out, _ := json.Marshal(q.Input)
				if len(q.Results) != 0 || len(q.Input) != 2 || !strings.Contains(string(out), "new question") || strings.Contains(string(out), "accepted skill") {
					t.Fatalf("accepted reply replayed or new input lost: input=%s results=%+v", out, q.Results)
				}
			})
		}
	}
}
