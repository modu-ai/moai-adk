package codexapp

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestSharedOwnerCannotBypassLegacyLease(t *testing.T) {
	c := helper(t)
	err := ServeShared(context.Background(), Config{Binary: c.cmd.Path, Home: c.cmd.Dir})
	if !errors.Is(err, ErrProfileBusy) {
		t.Fatalf("legacy lease bypassed: %v", err)
	}
}

func TestSharedClientRejectsNonSocket(t *testing.T) {
	c := helper(t)
	cfg := Config{Binary: c.cmd.Path, Home: c.cmd.Dir}
	if err := os.WriteFile(SharedSocketPath(cfg.Home), []byte("not socket"), 0600); err != nil {
		t.Fatal(err)
	}
	if other, err := ConnectShared(context.Background(), cfg); err == nil {
		if closeErr := other.Close(); closeErr != nil {
			t.Errorf("close unexpected client: %v", closeErr)
		}
		t.Fatal("regular file accepted")
	}
}

func TestSharedOwnerPreservesLiveOrphanSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shared Unix supervisor is not enabled on Windows")
	}
	home, err := os.MkdirTemp("/tmp", "moai-orphan-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(home); err != nil {
			t.Errorf("remove test home: %v", err)
		}
	})
	home, _ = filepath.EvalSymlinks(home)
	socket := SharedSocketPath(home)
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Errorf("close test listener: %v", err)
		}
	})
	if err := os.Chmod(socket, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := ServeShared(ctx, Config{Binary: filepath.Join(home, "must-not-start"), Home: home}); !errors.Is(err, ErrProfileBusy) {
		t.Fatalf("live orphan not protected: %v", err)
	}
	if _, err := os.Lstat(socket); err != nil {
		t.Fatalf("live socket removed: %v", err)
	}
}

// This opt-in check launches only an unauthenticated official local server;
// model requests and the operator's profile are never used.
func TestSharedOfficialTransport(t *testing.T) {
	if os.Getenv("MOAI_CODEX_TRANSPORT_LIVE") != "1" {
		t.Skip("official Codex local transport check requires MOAI_CODEX_TRANSPORT_LIVE=1")
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, _ = filepath.Abs(binary)
	shortHome, err := os.MkdirTemp("/tmp", "moai-shared-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(shortHome); err != nil {
			t.Errorf("remove test home: %v", err)
		}
	})
	home, _ := filepath.EvalSymlinks(shortHome)
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Binary: binary, Home: home}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	done := make(chan error, 1)
	go func() { done <- ServeShared(ctx, cfg) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(4 * time.Second):
			t.Error("shared owner cleanup timed out")
		}
	})
	for {
		if _, err := os.Lstat(SharedSocketPath(home)); err == nil {
			break
		}
		select {
		case err := <-done:
			done <- err
			t.Fatalf("owner startup: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
	first, err := ConnectShared(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := first.Close(); err != nil {
			t.Errorf("close first client: %v", err)
		}
	})
	second, err := ConnectShared(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := second.Close(); err != nil {
			t.Errorf("close second client: %v", err)
		}
	})
	for _, c := range []*Client{first, second} {
		if _, err := c.Initialize(ctx, "shared-test", "1"); err != nil {
			t.Fatal(err)
		}
		if account, err := c.Account(ctx); err != nil || account.Account != nil {
			t.Fatalf("fresh account: %+v %v", account, err)
		}
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Account(ctx); err != nil {
		t.Fatalf("closing peer disrupted owner: %v", err)
	}
}

func TestSharedOfficialLeaseInheritance(t *testing.T) {
	if os.Getenv("MOAI_CODEX_TRANSPORT_LIVE") != "1" {
		t.Skip("official Codex inherited lease check requires MOAI_CODEX_TRANSPORT_LIVE=1")
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.MkdirTemp("/tmp", "moai-inherit-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(home); err != nil {
			t.Errorf("remove test home: %v", err)
		}
	})
	home, _ = filepath.EvalSymlinks(home)
	lock := filepath.Join(home, ".moai-appserver.lock")
	lease, err := profileLease(lock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := lease.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Errorf("close test lease: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, binary, appServerArgs("unix://"+SharedSocketPath(home))...)
	configureProcess(cmd, home)
	inheritSharedLease(cmd, lease)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			t.Errorf("stop test server: %v", err)
		}
		_ = cmd.Wait() // A deliberately killed test server exits unsuccessfully.
	})
	for {
		if _, err := os.Lstat(SharedSocketPath(home)); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	if other, err := profileLease(lock); err == nil {
		if closeErr := other.Close(); closeErr != nil {
			t.Errorf("close unexpected lease: %v", closeErr)
		}
		t.Fatal("official auth owner discarded inherited lease")
	}
}
