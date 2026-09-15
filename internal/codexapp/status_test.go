package codexapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoginStatusHelper(t *testing.T) {
	if len(os.Args) < 2 || !strings.HasPrefix(os.Args[len(os.Args)-1], "status-helper-") {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "status-helper-chatgpt":
		if os.Getenv("OPENAI_API_KEY") != "" {
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "Logged in using ChatGPT")
		os.Exit(0)
	case "status-helper-none":
		fmt.Fprintln(os.Stderr, "Not logged in")
		os.Exit(1)
	default:
		fmt.Fprintln(os.Stderr, "unexpected secret-token")
		os.Exit(1)
	}
}

func TestLoginStatusWorksWhileProfileIsLeased(t *testing.T) {
	c := helper(t)
	t.Setenv("OPENAI_API_KEY", "not-forwarded")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	loggedIn, err := loginStatus(ctx, Config{Binary: c.cmd.Path, Home: c.cmd.Dir}, []string{"-test.run=TestLoginStatusHelper", "status-helper-chatgpt"})
	if err != nil || !loggedIn {
		t.Fatalf("status while gateway owns profile: authenticated=%t error=%v", loggedIn, err)
	}
	if _, err := c.Initialize(ctx, "still-active", "1"); err != nil {
		t.Fatalf("status disrupted active client: %v", err)
	}
}

func TestLoginStatusFailsClosedWithoutLeakingOutput(t *testing.T) {
	home, _ := filepath.EvalSymlinks(t.TempDir())
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	cfg := Config{Binary: exe, Home: home}
	for _, mode := range []string{"none", "error"} {
		got, err := loginStatus(context.Background(), cfg, []string{"-test.run=TestLoginStatusHelper", "status-helper-" + mode})
		if got || (mode == "none" && err != nil) || (mode == "error" && (err == nil || strings.Contains(err.Error(), "secret-token"))) {
			t.Fatalf("mode %s: %t %v", mode, got, err)
		}
	}
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("untrusted = true"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loginStatus(context.Background(), cfg, []string{"-test.run=TestLoginStatusHelper", "status-helper-chatgpt"}); err == nil {
		t.Fatal("unauthorized profile config accepted")
	}
}
