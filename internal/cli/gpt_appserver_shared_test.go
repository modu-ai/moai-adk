package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"golang.org/x/net/websocket"
)

func TestManagedGPTAppServerConfigBindsProfileConfiguration(t *testing.T) {
	home := sharedGPTFixtureHome(t)
	home, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}
	config, err := managedGPTAppServerConfig("/absolute/codex", home)
	if err != nil || config.ExpectedConfigSHA256 != "" {
		t.Fatalf("empty profile config: digest_present=%v err=%v", config.ExpectedConfigSHA256 != "", err)
	}
	raw := []byte("model = \"managed-fixture\"\n")
	path := filepath.Join(home, "config.toml")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256(raw)
	config, err = managedGPTAppServerConfig("/absolute/codex", home)
	if err != nil || config.ExpectedConfigSHA256 != hex.EncodeToString(want[:]) {
		t.Fatalf("profile config binding: digest_match=%v err=%v", config.ExpectedConfigSHA256 == hex.EncodeToString(want[:]), err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := managedGPTAppServerConfig("/absolute/codex", home); err == nil {
		t.Fatal("weak profile configuration accepted")
	}
}

func TestManagedGPTAppServerConfigRejectsChangedProfileConfiguration(t *testing.T) {
	root, err := filepath.EvalSymlinks(sharedGPTFixtureHome(t))
	if err != nil {
		t.Fatal(err)
	}
	startSharedGPTWiringFixture(t, root, filepath.Join(root, "protocol.jsonl"))
	home := filepath.Join(root, "gpt-appserver")
	path := filepath.Join(home, "config.toml")
	if err := os.WriteFile(path, []byte("model = \"before\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	config, err := managedGPTAppServerConfig("/absolute/codex", home)
	if err != nil {
		t.Fatal(err)
	}
	client, err := codexapp.ConnectShared(context.Background(), config)
	if err != nil {
		t.Fatal("unchanged profile configuration was rejected")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("model = \"after\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := codexapp.ConnectShared(context.Background(), config); err == nil {
		t.Fatal("stale profile configuration digest accepted")
	}
}

func TestGPTGatewayFactoryConnectsSharedOwnerWithManagedConfiguration(t *testing.T) {
	root, err := filepath.EvalSymlinks(sharedGPTFixtureHome(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", root)
	protocolLog := filepath.Join(root, "protocol.jsonl")
	startSharedGPTWiringFixture(t, root, protocolLog)
	profile := filepath.Join(root, "gpt-appserver")
	if err := os.WriteFile(filepath.Join(profile, "config.toml"), []byte("model = \"managed-fixture\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: "private-fixture", ModelIDs: []string{"gpt-5.6-sol"}, ContextTokens: gatewayContextWindow})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := newGPTAppServerGatewayHandler(payload)
	if err != nil {
		t.Fatal("stage=factory")
	}
	closer, ok := handler.(io.Closer)
	if !ok {
		t.Fatal("factory handler lifecycle unavailable")
	}
	if err := closer.Close(); err != nil {
		t.Fatal("stage=factory_close")
	}
}

func sharedGPTFixtureHome(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return t.TempDir()
	}
	home, err := os.MkdirTemp("/tmp", "moai-gpt-fixture-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(home) })
	return home
}

func startSharedGPTWiringFixture(t *testing.T, home, protocolLog string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	} // Product retains exclusive stdio on Windows.
	profile := filepath.Join(home, "gpt-appserver")
	if err := os.MkdirAll(profile, 0700); err != nil {
		t.Fatal(err)
	}
	path := codexapp.SharedSocketPath(profile)
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = listener.Close()
		t.Fatal(err)
	}
	var mu sync.Mutex
	connections := map[*websocket.Conn]bool{}
	server := &http.Server{ReadHeaderTimeout: time.Second, Handler: websocket.Handler(func(ws *websocket.Conn) {
		mu.Lock()
		connections[ws] = true
		mu.Unlock()
		defer func() { _ = ws.Close(); mu.Lock(); delete(connections, ws); mu.Unlock() }()
		for {
			var raw []byte
			if err := websocket.Message.Receive(ws, &raw); err != nil {
				return
			}
			f, err := os.OpenFile(protocolLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
			if err != nil {
				return
			}
			_, err = f.Write(append(raw, '\n'))
			_ = f.Close()
			if err != nil {
				return
			}
			var m codexapp.Message
			if json.Unmarshal(raw, &m) != nil {
				return
			}
			var result any
			switch m.Method {
			case "initialize":
				result = map[string]string{"userAgent": "shared-production-wiring-fake"}
			case "initialized":
				continue
			case "account/read":
				result = map[string]any{"account": map[string]string{"type": "chatgpt", "planType": "test"}}
			case "model/list":
				result = map[string]any{"data": []any{}, "nextCursor": nil}
			default:
				return
			}
			if err := websocket.JSON.Send(ws, map[string]any{"id": m.ID, "result": result}); err != nil {
				return
			}
		}
	})}
	done := make(chan struct{})
	go func() { defer close(done); _ = server.Serve(listener) }()
	t.Cleanup(func() {
		_ = server.Close()
		mu.Lock()
		for ws := range connections {
			_ = ws.Close()
		}
		mu.Unlock()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("shared fixture did not stop")
		}
	})
}

// TestSharedGPTStartupStageProbe is an opt-in, read-only startup diagnostic.
// It deliberately reports only stable stage codes: underlying App Server
// errors may contain operator-local paths or other private state.
func TestSharedGPTStartupStageProbe(t *testing.T) {
	if os.Getenv("MOAI_GPT_STARTUP_PROBE") != "1" {
		t.Skip("set MOAI_GPT_STARTUP_PROBE=1 to inspect the existing shared owner")
	}
	profile := os.Getenv("MOAI_GPT_PROBE_HOME")
	if profile == "" {
		var err error
		profile, err = managedGPTProfile()
		if err != nil {
			t.Fatal("stage=profile")
		}
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal("stage=codex_lookup")
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal("stage=codex_path")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	config, err := managedGPTAppServerConfig(binary, profile)
	if err != nil {
		t.Fatal("stage=config_binding")
	}
	client, err := codexapp.ConnectShared(ctx, config)
	if err != nil {
		t.Fatal("stage=shared_connect")
	}
	defer func() { _ = client.Close() }()
	if _, err = client.Initialize(ctx, "moai-startup-probe", "1"); err != nil {
		t.Fatal("stage=initialize")
	}
	account, err := client.Account(ctx)
	if err != nil {
		t.Fatal("stage=account_read")
	}
	if account.Account == nil || account.Account.Type != "chatgpt" {
		t.Fatal("stage=account_scope")
	}
}

func TestSharedGPTSupervisorTerminatesOfficialOwner(t *testing.T) {
	binary := os.Getenv("MOAI_GPT_TEST_BINARY")
	if binary == "" {
		t.Skip("built CLI lifecycle integration requires MOAI_GPT_TEST_BINARY")
	}
	codex, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	codex, _ = filepath.Abs(codex)
	home, err := os.MkdirTemp("/tmp", "moai-supervisor-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(home) })
	home, _ = filepath.EvalSymlinks(home)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, binary, "internal-gpt-appserver", "--home", home, "--codex", codex)
	cmd.Dir = home
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Kill() // A supervisor that already exited needs no further signal.
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("supervisor cleanup timed out")
		}
	})
	cfg := codexapp.Config{Binary: codex, Home: home}
	var first *codexapp.Client
	for {
		first, err = codexapp.ConnectShared(ctx, cfg)
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
	t.Cleanup(func() { _ = first.Close() })
	if _, err := first.Initialize(ctx, "supervisor-test", "1"); err != nil {
		t.Fatal(err)
	}
	second, err := codexapp.ConnectShared(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })
	if _, err := second.Initialize(ctx, "supervisor-test-2", "1"); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Account(ctx); err != nil {
		t.Fatalf("peer close broke owner: %v", err)
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		done <- err
	case <-ctx.Done():
		t.Fatal("SIGTERM did not reap official owner")
	}
	if _, err := os.Lstat(codexapp.SharedSocketPath(home)); !os.IsNotExist(err) {
		t.Fatalf("supervisor left socket: %v", err)
	}
	fresh, err := codexapp.Start(ctx, cfg)
	if err != nil {
		t.Fatalf("supervisor left auth lease: %v", err)
	}
	if err := fresh.Close(); err != nil {
		t.Fatal(err)
	}
}
