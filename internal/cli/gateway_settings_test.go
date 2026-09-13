package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway"
)

func TestGatewaySettingsCleanupPreservesBackupAndPermissions(t *testing.T) {
	for _, backup := range []string{"", "preserved-oauth"} {
		t.Run(backup, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "settings.local.json")
			env := map[string]any{"KEEP": "value", "ANTHROPIC_AUTH_TOKEN": "polluted", "MOAI_BACKUP_AUTH_TOKEN": backup, "CLAUDE_CODE_MAX_CONTEXT_TOKENS": "polluted", "ANTHROPIC_BASE_URL": "polluted"}
			data, _ := json.Marshal(map[string]any{"env": env, "permissions": map[string]any{"allow": []string{"Read"}}, "teammateMode": "tmux"})
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			if err := cleanupGatewaySettings(path); err != nil {
				t.Fatal(err)
			}
			data, _ = os.ReadFile(path)
			var got map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			after := got["env"].(map[string]any)
			if after["KEEP"] != "value" || got["permissions"] == nil || after["MOAI_BACKUP_AUTH_TOKEN"] != nil || after["ANTHROPIC_BASE_URL"] != nil || after["CLAUDE_CODE_MAX_CONTEXT_TOKENS"] != nil {
				t.Fatal(string(data))
			}
			if backup != "" && after["ANTHROPIC_AUTH_TOKEN"] != backup {
				t.Fatal("backup not restored")
			}
			if backup == "" && after["ANTHROPIC_AUTH_TOKEN"] != nil {
				t.Fatal("empty backup survived")
			}
		})
	}
}

func TestGatewayOverlaySeparatesSettingsEnvironment(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.json")
	if err := os.WriteFile(path, []byte(`{"env":{"SECRET_FIXTURE":"private","KEEP":"local"},"permissions":{"allow":["Read"]},"modelPicker":[{"model":"user-model"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	overlay, env, args, err := prepareGatewayOverlay([]string{"--settings", path, "--", "--settings=prompt"}, []string{"KEEP=parent"}, map[string]any{"teammateMode": "in-process"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(overlay, &got); err != nil {
		t.Fatal(err)
	}
	if got["env"] != nil || got["permissions"] == nil || got["modelPicker"] == nil || got["teammateMode"] != "in-process" {
		t.Fatalf("overlay %s", overlay)
	}
	if strings.Contains(string(overlay), "private") || !reflect.DeepEqual(args, []string{"--", "--settings=prompt"}) {
		t.Fatalf("overlay or args %s %v", overlay, args)
	}
	values := map[string]string{}
	for _, item := range env {
		k, v, _ := strings.Cut(item, "=")
		values[k] = v
	}
	if values["KEEP"] != "local" || values["SECRET_FIXTURE"] != "private" {
		t.Fatal("settings environment not extracted")
	}
}

func TestGatewayOverlayRejectsMalformedSettingsEnvironment(t *testing.T) {
	for _, raw := range []string{`{"env":{"KEY":1}}`, `{"env":null}`, `{"env":{"KEY":"a","KEY":"b"}}`, `{"permissions":true,"permissions":false}`} {
		if _, _, _, err := prepareGatewayOverlay([]string{"--settings", raw}, nil, nil); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestGatewayOverlayRejectsUnreadableRepeatedAndInvalidEntries(t *testing.T) {
	nullEnv, _ := json.Marshal(map[string]any{"env": map[string]string{"KEY": string([]byte{0})}})
	for _, args := range [][]string{{"--settings"}, {"--settings", "/missing/gateway-fixture"}, {"--settings", "{}", "--settings={}"}, {"--settings", `{"env":{"":"value"}}`}, {"--settings", string(nullEnv)}, {"--settings", `{"env":{"K=Y":"value"}}`}} {
		if _, _, _, err := prepareGatewayOverlay(args, nil, nil); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if _, _, _, err := prepareGatewayOverlay(nil, nil, map[string]any{"env": map[string]string{"SECRET": "value"}}); err == nil {
		t.Fatal("materialized overlay env")
	}
	if _, _, _, err := prepareGatewayOverlay(nil, nil, map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("unsupported settings encoded")
	}
	path := filepath.Join(t.TempDir(), "oversized.json")
	if err := os.WriteFile(path, bytes.Repeat([]byte(" "), gateway.MaxChildConfigBytes+1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := prepareGatewayOverlay([]string{"--settings", path}, nil, nil); err == nil {
		t.Fatal("oversized settings accepted")
	}
	if _, err := replaceGatewaySettingsArgs([]string{"--settings"}, "new"); err == nil {
		t.Fatal("missing settings value accepted")
	}
}

func TestGatewayOverlayPreservesFallbackTextAndArgumentBoundary(t *testing.T) {
	for _, args := range [][]string{
		{"--", "--fallback-model", "claude-sonnet-5"},
		{"--", "--fallback-model=claude-sonnet-5"},
		{"--settings", `{"description":"--fallback-model=claude-sonnet-5"}`, "explain --fallback-model"},
		{"--settings={\"description\":\"--fallback-model\"}", "--fallback-modeling"},
	} {
		data, _, remaining, err := prepareGatewayOverlay(args, nil, nil)
		if err != nil {
			t.Fatalf("args=%v: %v", args, err)
		}
		want := args
		if args[0] == "--settings" {
			want = args[2:]
		} else if strings.HasPrefix(args[0], "--settings=") {
			want = args[1:]
		}
		if !reflect.DeepEqual(remaining, want) {
			t.Fatalf("remaining=%v want=%v", remaining, want)
		}
		if args[0] != "--" && !strings.Contains(string(data), "--fallback-model") {
			t.Fatal("settings text lost")
		}
		got, err := replaceGatewaySettingsArgs(args, "replacement")
		if err != nil || !strings.Contains(strings.Join(got, " "), strings.Join(want, " ")) {
			t.Fatalf("replacement lost text: %v %v", got, err)
		}
	}
}
