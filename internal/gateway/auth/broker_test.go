package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCodexBrokerProcess(t *testing.T) {
	if os.Getenv("CODEX_HOME") == "" || !strings.Contains(strings.Join(os.Args, " "), "-- app-server") {
		return
	}
	home := os.Getenv("CODEX_HOME")
	if os.Getenv("HOME") != home || os.Getenv("OPENAI_API_KEY") != "" || os.Getenv("ANTHROPIC_AUTH_TOKEN") != "" {
		os.Exit(33)
	}
	modeBytes, _ := os.ReadFile(filepath.Join(home, "mock-mode"))
	mode := string(modeBytes)
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		var message struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		if json.Unmarshal(scan.Bytes(), &message) != nil {
			os.Exit(34)
		}
		switch message.Method {
		case "initialize":
			fmt.Println(`{"method":"account/updated","params":{}}`)
			if mode == "invalid-json" {
				fmt.Println(`{invalid`)
				continue
			}
			if mode == "wrong-id" {
				fmt.Println(`{"id":99,"result":{}}`)
				continue
			}
			if mode == "empty-init" {
				fmt.Println(`{"id":1}`)
				continue
			}
			if mode == "init-error" {
				fmt.Println(`{"id":1,"error":{"message":"secret-error"}}`)
				continue
			}
			fmt.Println(`{"id":1,"result":{"userAgent":"fake"}}`)
		case "account/login/start":
			fmt.Println(`{"method":"account/updated","params":{}}`)
			if mode == "login-bad" {
				fmt.Println(`{"id":2,"result":{"type":"apiKey"}}`)
				continue
			}
			if mode == "url-bad" || mode == "invalid-json" || mode == "wrong-id" || mode == "empty-init" {
				fmt.Println(`{"id":2,"result":{"type":"chatgpt","authUrl":"https://evil.example/","loginId":"current"}}`)
				continue
			}
			fmt.Println(`{"id":2,"result":{"type":"chatgpt","authUrl":"https://auth.openai.com/authorize?state=fixture","loginId":"current"}}`)
			if mode == "wait" {
				continue
			}
			tokenFixture(t, home, "broker", time.Now().Add(time.Hour))
			switch mode {
			case "wrong":
				fmt.Println(`{"method":"account/login/completed","params":{"loginId":"other","success":true}}`)
			case "failed":
				fmt.Println(`{"method":"account/login/completed","params":{"loginId":"current","success":false,"error":"secret-error"}}`)
			default:
				fmt.Println(`{"method":"account/login/completed","params":{"loginId":"current","success":true}}`)
			}
		case "account/read":
			if mode == "refresh-wrong-id" {
				fmt.Println(`{"id":99,"result":{}}`)
				continue
			}
			if mode == "refresh-api" {
				fmt.Println(`{"id":2,"result":{"account":{"type":"apiKey"}}}`)
				continue
			}
			fmt.Println(`{"method":"account/updated","params":{}}`)
			fmt.Println(`{"id":2,"result":{"account":{"type":"chatgpt"},"requiresOpenaiAuth":true}}`)
		}
	}
	if mode == "exit-failed" {
		os.Exit(39)
	}
	_ = os.WriteFile(filepath.Join(home, "exited"), []byte("yes"), 0600)
	os.Exit(0)
}
func TestCodexBrokerIsolatedLoginCompletionAndTermination(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "secret-api")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "secret-claude")
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"normal", "wrong", "failed", "wait", "exit-failed", "init-error", "login-bad", "url-bad", "invalid-json", "wrong-id", "empty-init", "callback-error"} {
		t.Run(mode, func(t *testing.T) {
			home, e := privateBrokerTestHome(t)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(home, "mock-mode"), []byte(mode), 0600); e != nil {
				t.Fatal(e)
			}
			if err := os.Chmod(home, 0700); err != nil {
				t.Fatal(err)
			}
			prompts := 0
			b := CodexBroker{Executable: executable, Timeout: 150 * time.Millisecond, OnLogin: func(p LoginPrompt) error {
				prompts++
				if mode == "callback-error" {
					return fmt.Errorf("secret-callback")
				}
				if p.LoginID != "current" {
					t.Fatal("wrong prompt")
				}
				return nil
			}, testEnv: []string{"GORACE=atexit_sleep_ms=0"}, testArgs: []string{"-test.run=TestCodexBrokerProcess", "--"}}
			e = b.Run(context.Background(), home, false)
			if mode == "normal" {
				if e != nil {
					t.Fatal(e)
				}
				if _, e = os.Stat(filepath.Join(home, "exited")); e != nil {
					t.Fatal("returned before subprocess exit")
				}
			} else if e == nil || strings.Contains(e.Error(), "secret-error") {
				t.Fatalf("bad completion accepted: %v", e)
			}
			if mode == "init-error" || mode == "login-bad" || mode == "url-bad" || mode == "invalid-json" || mode == "wrong-id" || mode == "empty-init" {
				if prompts != 0 {
					t.Fatal("invalid prompt emitted")
				}
				return
			}
			if prompts != 1 {
				t.Fatalf("prompt count %d", prompts)
			}
		})
	}
}
func TestCodexBrokerRefreshDoesNotClaimTokenAcceptance(t *testing.T) {
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	home, e := privateBrokerTestHome(t)
	if e != nil {
		t.Fatal(e)
	}
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	b := CodexBroker{Executable: executable, Timeout: time.Second, testEnv: []string{"GORACE=atexit_sleep_ms=0"}, testArgs: []string{"-test.run=TestCodexBrokerProcess", "--"}}
	if e = b.Run(context.Background(), home, true); e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(filepath.Join(home, "auth.json")); !os.IsNotExist(e) {
		t.Fatal("wrapper fabricated token")
	}
}

func TestCodexBrokerRejectsSymlinkedEnvironmentDirectories(t *testing.T) {
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	home, e := privateBrokerTestHome(t)
	if e != nil {
		t.Fatal(e)
	}
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	if e = os.Symlink(t.TempDir(), filepath.Join(home, "config")); e != nil {
		t.Fatal(e)
	}
	b := CodexBroker{Executable: executable, Timeout: time.Second, OnLogin: func(LoginPrompt) error { return nil }, testEnv: []string{"GORACE=atexit_sleep_ms=0"}, testArgs: []string{"-test.run=TestCodexBrokerProcess", "--"}}
	if e = b.Run(context.Background(), home, false); e == nil {
		t.Fatal("symlink config escape accepted")
	}
}

func TestCodexBrokerRefreshRejectsWrongRPCAndAccountType(t *testing.T) {
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"refresh-wrong-id", "refresh-api"} {
		home, e := privateBrokerTestHome(t)
		if e != nil {
			t.Fatal(e)
		}
		if err := os.Chmod(home, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(home, "mock-mode"), []byte(mode), 0600); err != nil {
			t.Fatal(err)
		}
		b := CodexBroker{Executable: executable, Timeout: time.Second, testEnv: []string{"GORACE=atexit_sleep_ms=0"}, testArgs: []string{"-test.run=TestCodexBrokerProcess", "--"}}
		if e = b.Run(context.Background(), home, true); e == nil {
			t.Fatal("invalid refresh RPC accepted")
		}
	}
}
