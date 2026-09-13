//go:build !windows

package receipt

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestReceiptLockHelper(t *testing.T) {
	dir := os.Getenv("MOAI_RECEIPT_LOCK_TEST")
	if dir == "" {
		return
	}
	f, e := os.OpenFile(filepath.Join(dir, ".lock"), os.O_RDWR, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	if e = lock(context.Background(), f); e != nil {
		t.Fatal(e)
	}
	defer unlock(f)
	fmt.Println("LOCKED")
	<-time.After(30 * time.Second) // Parent command context and test cleanup bound lifetime.
}
func TestProcessLockCancellationAndCrashRelease(t *testing.T) {
	ctx := context.Background()
	dir := privateDir(t)
	s, e := OpenStore(ctx, dir, testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(childCtx, os.Args[0], "-test.run=^TestReceiptLockHelper$")
	cmd.Env = append(os.Environ(), "MOAI_RECEIPT_LOCK_TEST="+dir)
	out, e := cmd.StdoutPipe()
	if e != nil {
		t.Fatal(e)
	}
	if e = cmd.Start(); e != nil {
		t.Fatal(e)
	}
	waited := false
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		if !waited {
			_ = cmd.Wait()
		}
	})
	ready := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(out)
		if scanner.Scan() {
			ready <- scanner.Text()
		} else {
			ready <- ""
		}
	}()
	select {
	case line := <-ready:
		if line != "LOCKED" {
			t.Fatal(line)
		}
	case <-childCtx.Done():
		t.Fatal("lock helper readiness timeout")
	}
	deadline, stop := context.WithTimeout(ctx, 40*time.Millisecond)
	defer stop()
	if _, e = s.Snapshot(deadline); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("lock cancellation", e)
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	waited = true
	if _, e = s.Snapshot(ctx); e != nil {
		t.Fatal("crash lock retained", e)
	}
}
func TestStoreStateFailureBoundaries(t *testing.T) {
	ctx := context.Background()
	if _, e := OpenStore(ctx, "relative", testUUID, true); e == nil {
		t.Fatal("relative root")
	}
	dir := privateDir(t)
	s, e := OpenStore(ctx, dir, testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	c := candidate("p", "parent", "opaque")
	if e = s.Publish(ctx, c); e != nil {
		t.Fatal(e)
	}
	if e = s.Publish(ctx, c); e != nil {
		t.Fatal(e)
	}
	c.Complete = false
	if e = s.Publish(ctx, c); e == nil {
		t.Fatal("incomplete publication")
	}
	manifest := filepath.Join(dir, "manifest.json")
	if e = os.WriteFile(manifest, []byte("truncated"), 0600); e != nil {
		t.Fatal(e)
	}
	if e = s.Publish(ctx, c); e == nil {
		t.Fatal("corrupt store overwritten")
	}
	_ = s.Close()
	if _, e = s.Snapshot(ctx); e == nil {
		t.Fatal("closed store")
	}
	if e = s.Close(); e != nil {
		t.Fatal(e)
	}
	dir = privateDir(t)
	s, e = OpenStore(ctx, dir, testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	target := filepath.Join(privateDir(t), "hardlink")
	if e = os.Link(filepath.Join(dir, "manifest.json"), target); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Snapshot(ctx); e == nil {
		t.Fatal("hardlinked receipt")
	}
	_ = os.Remove(target)
	_ = os.Chmod(filepath.Join(dir, ".lock"), 0644)
	if _, e = s.Snapshot(ctx); e == nil {
		t.Fatal("public lock")
	}
}

func TestFailedPublicationPreservesPreviousSnapshot(t *testing.T) {
	dir := privateDir(t)
	ctx := context.Background()
	s, e := OpenStore(ctx, dir, testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	// A privileged runner bypasses directory permissions; exercise a closed
	// filesystem handle instead, still requiring the previous disk state intact.
	if os.Geteuid() == 0 {
		_ = s.root.Close()
		m, _ := New(testUUID)
		_ = m.Publish(candidate("new", "parent", "opaque"))
		if s.write(ctx, m, s.rootValid, nil) == nil {
			t.Fatal("closed filesystem reported publication success")
		}
		raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		previous, err := Parse(raw, testUUID)
		if err != nil || len(previous.Candidates()) != 0 {
			t.Fatal("failed publication changed manifest", err)
		}
		return
	}
	if e = os.Chmod(dir, 0500); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0700) })
	if e = s.Publish(ctx, candidate("new", "parent", "opaque")); e == nil {
		t.Fatal("failed storage reported publication success")
	}
	if e = os.Chmod(dir, 0700); e != nil {
		t.Fatal(e)
	}
	snapshot, e := s.Snapshot(ctx)
	if e != nil || len(snapshot.Candidates()) != 0 {
		t.Fatal("failed publish changed manifest", e)
	}
}
