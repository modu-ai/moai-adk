package hook

// config_change_payload_test.go — card t1499 M1.
//
// The audit log held thousands of `rejected ... read file: open : no such
// file` rows with an empty path and source: the handler read field names that
// Claude Code's ConfigChange payload does not carry. The fixture below is a
// payload captured from a live Claude Code 2.1.289 session (a headless
// `claude -p` run that edited .claude/settings.json); it names the fields
// `source` and `file_path`.

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/testutil"
)

// liveConfigChangePayload loads the captured payload and re-points its
// machine-specific cwd and file_path at projectDir, keeping every key name
// exactly as captured. It returns the rewritten bytes and the config path.
func liveConfigChangePayload(t *testing.T, projectDir string) ([]byte, string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "config_change_payload_live.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	cfgPath := filepath.Join(projectDir, ".claude", "settings.json")
	m["cwd"] = projectDir
	m["file_path"] = cfgPath
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("re-encode fixture: %v", err)
	}
	return out, cfgPath
}

// TestConfigChange_LiveClaudeCodePayload feeds the captured real payload
// through the protocol decoder and the handler and asserts the audit row names
// the changed file and its source and records a valid config as reloaded.
func TestConfigChange_LiveClaudeCodePayload(t *testing.T) {
	projectDir := newMoaiProjectRoot(t)
	t.Setenv(config.EnvClaudeProjectDir, projectDir)

	payload, cfgPath := liveConfigChangePayload(t, projectDir)
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// settings.json is JSON, which is valid YAML.
	if err := os.WriteFile(cfgPath, []byte(`{"env":{"T1499":"1"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	input, err := NewProtocol().ReadInput(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("ReadInput: %v", err)
	}

	h := NewConfigChangeHandler().(*configChangeHandler)
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	testutil.WaitForAsync(t, h.waitGroup(), 5*time.Second)

	lines := readConfigChangeAudit(t, projectDir)
	if len(lines) != 1 {
		t.Fatalf("audit rows = %d, want 1: %v", len(lines), lines)
	}
	row := lines[0]
	for _, want := range []string{"path=" + cfgPath, "source=project_settings", "result=reloaded"} {
		if !strings.Contains(row, want) {
			t.Errorf("audit row missing %q: %s", want, row)
		}
	}
	if strings.Contains(row, "result=rejected") {
		t.Errorf("valid config recorded as rejected: %s", row)
	}
}

// TestConfigChange_LegacyFieldNamesStillResolve keeps the pre-existing
// MoAI-era names working: config_file_path / config_source and the older
// configuration_source.
func TestConfigChange_LegacyFieldNamesStillResolve(t *testing.T) {
	for name, payload := range map[string]string{
		"config_source":        `{"config_file_path":"/p/a.yaml","config_source":"user_settings"}`,
		"configuration_source": `{"config_file_path":"/p/a.yaml","configuration_source":"user_settings"}`,
		"official":             `{"file_path":"/p/a.yaml","source":"user_settings"}`,
	} {
		t.Run(name, func(t *testing.T) {
			var in HookInput
			if err := json.Unmarshal([]byte(payload), &in); err != nil {
				t.Fatal(err)
			}
			if got := configChangePath(&in); got != "/p/a.yaml" {
				t.Errorf("path = %q, want /p/a.yaml", got)
			}
			if got := configChangeSource(&in); got != "user_settings" {
				t.Errorf("source = %q, want user_settings", got)
			}
		})
	}
}
