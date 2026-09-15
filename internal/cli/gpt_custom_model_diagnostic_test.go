package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Installed Claude Code, synthetic local HTTP only. Custom picker registration
// is not the same contract as recognizing an Anthropic model's identity.
func TestClaudeCustomModelDiagnosticLocal(t *testing.T) {
	if os.Getenv("MOAI_CLAUDE_INTEGRATION") != "1" {
		t.Skip("requires opt-in installed Claude Code local fixture")
	}
	claude, err := exec.LookPath("claude")
	if err != nil {
		t.Fatal(err)
	}
	for _, registered := range []bool{false, true} {
		t.Run(fmt.Sprintf("registered=%v", registered), func(t *testing.T) {
			var mu sync.Mutex
			var models []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/messages" {
					http.NotFound(w, r)
					return
				}
				var body struct {
					Model  string `json:"model"`
					Stream bool   `json:"stream"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					http.Error(w, "invalid fixture request", 400)
					return
				}
				mu.Lock()
				models = append(models, body.Model)
				mu.Unlock()
				if !body.Stream {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{"id": "fixture", "type": "message", "role": "assistant", "model": body.Model, "content": []any{map[string]string{"type": "text", "text": "CUSTOM_MODEL_OK"}}, "stop_reason": "end_turn", "usage": map[string]int{"input_tokens": 10, "output_tokens": 4}})
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = fmt.Fprintf(w, "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"fixture\",\"type\":\"message\",\"role\":\"assistant\",\"model\":%q,\"content\":[],\"usage\":{\"input_tokens\":10,\"output_tokens\":0}}}\n\nevent: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\nevent: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"CUSTOM_MODEL_OK\"}}\n\nevent: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":4}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n", body.Model)
			}))
			defer server.Close()
			home := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, claude, "--print", "--model", "gpt-5.6-sol", "--tools", "", "--max-turns", "1", "Reply CUSTOM_MODEL_OK")
			cmd.Dir = home
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "CLAUDE_CONFIG_DIR=" + filepath.Join(home, "native"), "ANTHROPIC_API_KEY=synthetic-local-only", "ANTHROPIC_BASE_URL=" + server.URL, "ANTHROPIC_DEFAULT_HAIKU_MODEL=gpt-5.6-luna", "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1", "DISABLE_AUTOUPDATER=1"}
			if registered {
				cmd.Env = append(cmd.Env, "ANTHROPIC_CUSTOM_MODEL_OPTION=gpt-5.6-sol", "ANTHROPIC_CUSTOM_MODEL_OPTION_NAME=GPT 5.6 Sol", "ANTHROPIC_CUSTOM_MODEL_OPTION_DESCRIPTION=GPT through local gateway")
			}
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("local Claude: %v stderr=%q", err, stderr.String())
			}
			if !strings.Contains(stdout.String(), "CUSTOM_MODEL_OK") {
				t.Fatalf("response missing: %q", stdout.String())
			}
			combined := stdout.String() + stderr.String()
			if !strings.Contains(combined, "[claude-code:unrecognized_model]") {
				t.Fatal("installed Claude diagnostic contract changed; reassess custom model recognition")
			}
			mu.Lock()
			defer mu.Unlock()
			if len(models) == 0 {
				t.Fatal("no local model request")
			}
			for _, model := range models {
				if model != "gpt-5.6-sol" && model != "gpt-5.6-luna" {
					t.Errorf("unexpected model: %s", model)
				}
			}
			t.Logf("registered=%v response=CUSTOM_MODEL_OK diagnostic=true request_models=%v", registered, models)
		})
	}
}
