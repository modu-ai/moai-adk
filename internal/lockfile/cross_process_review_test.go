package lockfile

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewTryLockExcludesAnotherProcess(t *testing.T) {
	if path := os.Getenv("MOAI_LOCK_REVIEW_CHILD"); path != "" {
		f, err := os.OpenFile(path, os.O_RDWR, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		if err := TryLock(f); err != nil {
			fmt.Println("REVIEW_LOCK_HELD")
			return
		}
		if err := Unlock(f); err != nil {
			t.Fatal(err)
		}
		fmt.Println("REVIEW_LOCK_ACQUIRED")
		return
	}
	path := filepath.Join(t.TempDir(), "migration.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := TryLock(f); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	probe := func(want string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, exe, "-test.run=^TestReviewTryLockExcludesAnotherProcess$", "-test.count=1")
		cmd.Env = append(os.Environ(), "MOAI_LOCK_REVIEW_CHILD="+path)
		out, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(out), want) {
			t.Fatalf("child lock result want %s: %v\n%s", want, err, out)
		}
	}
	probe("REVIEW_LOCK_HELD")
	if err := Unlock(f); err != nil {
		t.Fatal(err)
	}
	probe("REVIEW_LOCK_ACQUIRED")
}

func TestReviewTryLockReleasesWhenDescriptorCloses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "close-releases.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := TryLock(f); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	next, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = next.Close() }()
	if err := TryLock(next); err != nil {
		t.Fatalf("descriptor close retained the previous migration lock: %v", err)
	}
	if err := Unlock(next); err != nil {
		t.Fatal(err)
	}
}

func TestReviewLockPreservesOrdinaryFileIO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("before"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := Lock(f); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := Unlock(f); err != nil {
			t.Error(err)
		}
	}()
	if body, err := os.ReadFile(path); err != nil || string(body) != "before" {
		t.Fatalf("advisory lock blocked ordinary reading through another handle: %q, %v", body, err)
	}
	if err := os.WriteFile(path, []byte("after"), 0600); err != nil {
		t.Fatalf("advisory lock blocked ordinary writing through another handle: %v", err)
	}
}
