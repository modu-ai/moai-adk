package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

func TestFactoryLiveClaudeCommandRoutesThroughBuiltMoaiGLM(t *testing.T) {
	root := t.TempDir()
	f := &factoryLiveFixture{root: root, moai: filepath.Join(t.TempDir(), "moai")}
	cmd, err := f.modelCommand(context.Background(), factorymsg.Peer{Backend: "claude"}, "live prompt")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		f.moai, "glm", "--", "-p", "--output-format", "json",
		"--mcp-config", filepath.Join(root, ".mcp.json"), "--strict-mcp-config",
		"--permission-mode", "bypassPermissions", "live prompt",
	}
	if cmd.Path != f.moai || !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("command path/args = %q %q, want %q %q", cmd.Path, cmd.Args, f.moai, want)
	}
	if filepath.Base(cmd.Path) == "claude" {
		t.Fatal("Claude live path selected claude directly")
	}
}

func TestFactoryLiveOperatorGLMKeyRestoresIsolatedHome(t *testing.T) {
	realHome, isolatedHome := t.TempDir(), t.TempDir()
	credentialDir := filepath.Join(realHome, ".moai")
	if err := os.MkdirAll(credentialDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(credentialDir, ".env.glm"), []byte("GLM_API_KEY=test-live-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	previousHomeFn := factoryLiveOperatorHomeFn
	factoryLiveOperatorHomeFn = func() (string, error) { return realHome, nil }
	t.Cleanup(func() { factoryLiveOperatorHomeFn = previousHomeFn })
	t.Setenv(config.EnvHome, isolatedHome)
	t.Setenv(config.EnvTestGLMKey, "")
	if got := factoryLiveOperatorGLMKey(); got != "test-live-key" {
		t.Fatalf("credential bridge returned unexpected value")
	}
	if got, present := os.LookupEnv(config.EnvHome); !present || got != isolatedHome {
		t.Fatalf("%s was not restored exactly: value=%q present=%v", config.EnvHome, got, present)
	}
	if _, err := os.Stat(filepath.Join(isolatedHome, ".env.glm")); !os.IsNotExist(err) {
		t.Fatalf("credential was copied into isolated MOAI_HOME: %v", err)
	}
}

func TestFactoryLiveGLMLauncherEnvDefersAttributionToMCPConfig(t *testing.T) {
	root := t.TempDir()
	f := &factoryLiveFixture{root: root, runID: "live-run", moai: "/fixture/moai"}
	p := factorymsg.Peer{Backend: "claude", Role: "worker", Slot: "agent-1", SessionUUID: "fixture-session", PID: 4242}
	if err := f.writePeerMCPConfig(p); err != nil {
		t.Fatal(err)
	}
	launcherEnv := factoryLiveWithoutAttribution([]string{
		"PATH=/bin", config.EnvMoaiKanbanID + "=stale", config.EnvMoaiSessionPID + "=999",
		config.EnvMoaiFactoryWorker + "=stale-worker", config.EnvClaudeCodeSessionID + "=stale-session",
	})
	for _, key := range []string{config.EnvMoaiKanbanID, config.EnvMoaiSessionPID, config.EnvMoaiFactoryWorker, config.EnvClaudeCodeSessionID} {
		if got := launchEnvValue(launcherEnv, key); got != "" {
			t.Fatalf("launcher env retained %s=%q", key, got)
		}
	}
	var doc struct {
		MCPServers map[string]struct {
			Env map[string]string `json:"env"`
		} `json:"mcpServers"`
	}
	body, err := os.ReadFile(filepath.Join(root, ".mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	env := doc.MCPServers["moai"].Env
	if env[config.EnvMoaiKanbanID] != f.runID || env[config.EnvMoaiSessionPID] != "4242" || env[config.EnvMoaiFactoryWorker] != "agent-1" || env[config.EnvClaudeCodeSessionID] != p.SessionUUID {
		t.Fatalf("MCP attribution env=%v", env)
	}
	if _, leaked := env[config.EnvTestGLMKey]; leaked {
		t.Fatal("GLM credential was written to MCP config")
	}
}
