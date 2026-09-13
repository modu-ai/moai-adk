package codexapp

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("CODEX_HOME") == "" || len(os.Args) < 2 || os.Args[len(os.Args)-1] != "codexapp-helper" {
		return
	}
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		var m Message
		if json.Unmarshal(scan.Bytes(), &m) != nil {
			continue
		}
		switch m.Method {
		case "initialize":
			fmt.Printf("{\"id\":%s,\"result\":{\"userAgent\":\"fake\"}}\n", m.ID)
		case "account/login/start":
			fmt.Printf("{\"id\":%s,\"result\":%s}\n", m.ID, m.Params)
		case "account/read", "model/list", "account/logout", "account/login/cancel":
			fmt.Printf("{\"id\":%s,\"result\":{}}\n", m.ID)

		case "echo":
			fmt.Printf("{\"method\":\"notice\",\"params\":{}}\n{\"id\":%s,\"result\":%s}\n", m.ID, m.Params)
		case "server":
			fmt.Printf("{\"id\":\"server-1\",\"method\":\"item/tool/call\",\"params\":{}}\n{\"id\":%s,\"result\":{}}\n", m.ID)
		case "error":
			fmt.Printf("{\"id\":%s,\"error\":{\"code\":-1,\"message\":\"secret-token\"}}\n", m.ID)
		case "malformed":
			fmt.Println("not-json")
		case "flood":
			for i := 0; i < 100; i++ {
				fmt.Println(`{"method":"notice"}`)
			}
		case "oversize":
			fmt.Println(strings.Repeat("x", 4096))
		case "eof":
			os.Exit(0)
		case "wait":
		case "lease":
			file, err := profileLease(filepath.Join(os.Getenv("CODEX_HOME"), ".moai-appserver.lock"))
			if err == nil {
				_ = file.Close() // probe answers the acquire flag; the lease is released by exit
			}
			fmt.Printf("{\"id\":%s,\"result\":{\"acquired\":%t}}\n", m.ID, err == nil)
		case "env":
			fmt.Printf("{\"id\":%s,\"result\":{\"key\":%q}}\n", m.ID, os.Getenv("OPENAI_API_KEY"))
		default:
			if len(m.ID) > 0 {
				fmt.Printf("{\"method\":\"responded\",\"params\":{}}\n")
			}
		}
	}
	os.Exit(0)
}
func helper(t *testing.T) *Client {
	t.Helper()
	home, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	c, err := startProcess(context.Background(), Config{Binary: exe, Home: home, MaxMessageBytes: 2048, QueueSize: 8}, []string{"-test.run=TestHelperProcess", "codexapp-helper"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})
	return c
}
func TestRoutingEventsAndResponse(t *testing.T) {
	c := helper(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var got map[string]int
	if err := c.Call(ctx, "echo", map[string]int{"n": 7}, &got); err != nil || got["n"] != 7 {
		t.Fatal(got, err)
	}
	if e := <-c.Events(); e.Method != "notice" {
		t.Fatal(e)
	}
	if err := c.Call(ctx, "server", nil, nil); err != nil {
		t.Fatal(err)
	}
	e := <-c.Events()
	if err := c.Respond(ctx, e.ID, map[string]bool{"success": true}); err != nil {
		t.Fatal(err)
	}
	if err := c.Respond(ctx, e.ID, nil); err == nil {
		t.Fatal("duplicate response accepted")
	}
}
func TestFailureAndSecretRedaction(t *testing.T) {
	for _, method := range []string{"error", "oversize", "eof"} {
		t.Run(method, func(t *testing.T) {
			c := helper(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := c.Call(ctx, method, nil, nil)
			if err == nil || strings.Contains(err.Error(), "secret-token") {
				t.Fatal(err)
			}
		})
	}
}
func TestCancellationAndEnvironment(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "secret-token")
	c := helper(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := c.Call(ctx, "wait", nil, nil); err == nil {
		t.Fatal("not canceled")
	}
	var got map[string]string
	ctx2, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := c.Call(ctx2, "env", nil, &got); err != nil || got["key"] != "" {
		t.Fatal(got, err)
	}
}

func TestManagedMethods(t *testing.T) {
	c := helper(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := c.Initialize(ctx, "moai-test", "1"); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []LoginMode{LoginBrowser, LoginDevice, LoginAPIKey} {
		key := ""
		if mode == LoginAPIKey {
			key = "synthetic-key"
		}
		r, err := c.Login(ctx, mode, key)
		if err != nil || r.Type != string(mode) {
			t.Fatal(r, err)
		}
	}
	if _, err := c.Login(ctx, LoginMode("chatgptAuthTokens"), "secret"); err == nil {
		t.Fatal("external token mode accepted")
	}
	if _, err := c.Account(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Models(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if err := c.Logout(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestProtocolIDsAndBounds(t *testing.T) {
	for _, raw := range []string{`null`, `true`, `1.5`, `{}`, `""`, `"x" extra`} {
		if _, err := idKey(json.RawMessage(raw)); err == nil {
			t.Fatal(raw)
		}
	}
	for _, method := range []string{"malformed", "flood"} {
		t.Run(method, func(t *testing.T) {
			c := helper(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := c.Call(ctx, method, nil, nil); err == nil {
				t.Fatal("missing error")
			}
		})
	}
	c := helper(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.Call(ctx, "echo", strings.Repeat("x", 3000), nil); err != ErrLimit {
		t.Fatal(err)
	}
}
func TestConcurrentResponses(t *testing.T) {
	c := helper(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	errs := make(chan error, 6)
	for i := 0; i < 6; i++ {
		go func(n int) {
			var got map[string]int
			err := c.Call(ctx, "echo", map[string]int{"n": n}, &got)
			if err == nil && got["n"] != n {
				err = fmt.Errorf("mismatched response")
			}
			errs <- err
		}(i)
	}
	for i := 0; i < 6; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
}
func TestRejectProfileConfigAndSymlink(t *testing.T) {
	home, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model_provider='foreign'"), 0600); err != nil {
		t.Fatal(err)
	}
	if c, err := Start(context.Background(), Config{Binary: exe, Home: home}); err == nil {
		if closeErr := c.Close(); closeErr != nil {
			t.Errorf("close client: %v", closeErr)
		}
		t.Fatal("foreign config accepted")
	}
}

func TestProfileLeaseAndRelease(t *testing.T) {
	c := helper(t)
	cfg := Config{Binary: c.cmd.Path, Home: c.cmd.Dir}
	if other, err := startProcess(context.Background(), cfg, []string{"-test.run=TestHelperProcess", "codexapp-helper"}); err == nil {
		if closeErr := other.Close(); closeErr != nil {
			t.Errorf("close client: %v", closeErr)
		}
		t.Fatal("parallel profile accepted")
	}
	if err := c.Close(); err != nil {
		t.Errorf("close client: %v", err)
	}
	other, err := startProcess(context.Background(), cfg, []string{"-test.run=TestHelperProcess", "codexapp-helper"})
	if err != nil {
		t.Fatal(err)
	}
	if err := other.Close(); err != nil {
		t.Errorf("close client: %v", err)
	}
}

func TestLeaseAcrossChildProcess(t *testing.T) {
	c := helper(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var got map[string]bool
	if err := c.Call(ctx, "lease", nil, &got); err != nil || got["acquired"] {
		t.Fatal(got, err)
	}
}
func TestValidationAndClosedConnection(t *testing.T) {
	c := helper(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := c.Initialize(ctx, "", ""); err == nil {
		t.Fatal("empty client")
	}
	for _, mode := range []LoginMode{LoginBrowser, LoginDevice} {
		if _, err := c.Login(ctx, mode, "secret"); err == nil {
			t.Fatal("unexpected key")
		}
	}
	if _, err := c.Login(ctx, LoginAPIKey, ""); err == nil {
		t.Fatal("empty key")
	}
	if err := c.CancelLogin(ctx, ""); err == nil {
		t.Fatal("empty login")
	}
	if err := c.CancelLogin(ctx, "synthetic"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Models(ctx, "next"); err != nil {
		t.Fatal(err)
	}
	if err := c.Call(ctx, "", nil, nil); err != ErrProtocol {
		t.Fatal(err)
	}
	if err := c.Call(ctx, "echo", make(chan int), nil); err != ErrProtocol {
		t.Fatal(err)
	}
	if err := c.Notify(ctx, "bad", make(chan int)); err != ErrProtocol {
		t.Fatal(err)
	}
	if err := c.Respond(ctx, json.RawMessage(`null`), nil); err != ErrProtocol {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("close client: %v", err)
	}
	if c.Err() == nil {
		t.Fatal("missing terminal error")
	}
	if err := c.Call(ctx, "echo", nil, nil); err == nil {
		t.Fatal("closed accepted")
	}
	if err := c.Notify(ctx, "x", nil); err == nil {
		t.Fatal("closed notification")
	}
}
func TestStartupValidation(t *testing.T) {
	home, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	for _, cfg := range []Config{{Binary: "relative", Home: home}, {Binary: exe, Home: "relative"}, {Binary: exe, Home: home, QueueSize: -1}, {Binary: exe, Home: home, MaxMessageBytes: 1}, {Binary: "/missing-codex-binary", Home: home}} {
		if c, err := Start(context.Background(), cfg); err == nil {
			if closeErr := c.Close(); closeErr != nil {
				t.Errorf("close client: %v", closeErr)
			}
			t.Fatal("accepted invalid startup")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Start(ctx, Config{Binary: exe, Home: home}); err == nil {
		t.Fatal("canceled startup")
	}
}

func TestInstalledAppServerNoAuthSmoke(t *testing.T) {
	binary := os.Getenv("CODEXAPP_LIVE_BINARY")
	if binary == "" {
		t.Skip("explicit installed binary required")
	}
	home, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	c, err := Start(ctx, Config{Binary: binary, Home: home})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	}()
	init, err := c.Initialize(ctx, "moai_as1_probe", "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	account, err := c.Account(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if account.Account != nil {
		t.Fatal("fresh profile unexpectedly authenticated")
	}
	t.Logf("initialize_user_agent=%q account=null requiresOpenaiAuth=%t", init.UserAgent, account.RequiresOpenAIAuth)
}

func TestAuthorizedConfigDigest(t *testing.T) {
	home, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	raw := []byte("features.shell_tool = false\n")
	if err := os.WriteFile(filepath.Join(home, "config.toml"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	cfg := Config{Binary: exe, Home: home, ExpectedConfigSHA256: hex.EncodeToString(sum[:])}
	c, err := startProcess(context.Background(), cfg, []string{"-test.run=TestHelperProcess", "codexapp-helper"})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("close client: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if other, err := startProcess(context.Background(), cfg, nil); err == nil {
		if closeErr := other.Close(); closeErr != nil {
			t.Errorf("close client: %v", closeErr)
		}
		t.Fatal("changed authorized config accepted")
	}
	if err := os.Remove(filepath.Join(home, "config.toml")); err != nil {
		t.Fatal(err)
	}
	if _, err := startProcess(context.Background(), cfg, nil); err == nil {
		t.Fatal("missing authorized config accepted")
	}
}
func TestLifetimeCancellation(t *testing.T) {
	home, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	ctx, cancel := context.WithCancel(context.Background())
	c, err := startProcess(ctx, Config{Binary: exe, Home: home}, []string{"-test.run=TestHelperProcess", "codexapp-helper"})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case <-c.done:
	case <-time.After(time.Second):
		t.Fatal("lifetime cancellation stuck")
	}
	if c.Err() != context.Canceled {
		t.Fatal(c.Err())
	}
	if err := c.Close(); err != nil {
		t.Errorf("close client: %v", err)
	}
}

func TestDiscardRequestPreservesOtherIDsAndRejectsReplay(t *testing.T) {
	c := helper(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Call(ctx, "server", nil, nil); err != nil {
		t.Fatal(err)
	}
	event := <-c.Events()
	if err := c.DiscardRequest(json.RawMessage(`"foreign"`)); err == nil {
		t.Fatal("unknown discard accepted")
	}
	if err := c.DiscardRequest(json.RawMessage(`null`)); err == nil {
		t.Fatal("invalid discard accepted")
	}
	if err := c.Respond(ctx, event.ID, map[string]bool{"success": true}); err != nil {
		t.Fatal("foreign discard consumed owned request", err)
	}
	if event := <-c.Events(); event.Method != "responded" {
		t.Fatal(event)
	}
	if err := c.Call(ctx, "server", nil, nil); err != nil {
		t.Fatal(err)
	}
	event = <-c.Events()
	if err := c.DiscardRequest(event.ID); err != nil {
		t.Fatal(err)
	}
	if err := c.DiscardRequest(event.ID); err == nil {
		t.Fatal("discard replay accepted")
	}
	if err := c.Respond(ctx, event.ID, nil); err == nil {
		t.Fatal("discarded request answered")
	}
}
