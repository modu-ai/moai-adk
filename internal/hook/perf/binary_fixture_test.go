package perf

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

const perfBinaryEnv = "MOAI_HOOK_PERF_TEST_BINARY"

var perfBinaryFixture struct {
	once sync.Once
	dir  string
	path string
	err  error
}

func TestMain(m *testing.M) {
	code := m.Run()
	if perfBinaryFixture.dir != "" {
		if err := os.RemoveAll(perfBinaryFixture.dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			code = 1
		}
	}
	os.Exit(code)
}

func buildMoaiBinary(t *testing.T) string {
	t.Helper()
	if os.Getenv(guardChildEnv) != "" {
		if bin := os.Getenv(perfBinaryEnv); bin != "" {
			if _, err := os.Stat(bin); err != nil {
				t.Fatal(err)
			}
			return bin
		}
	}
	perfBinaryFixture.once.Do(func() {
		perfBinaryFixture.dir, perfBinaryFixture.err = os.MkdirTemp("", "moai-perf-test-")
		if perfBinaryFixture.err != nil {
			return
		}
		perfBinaryFixture.path = filepath.Join(perfBinaryFixture.dir, "moai")
		if runtime.GOOS == "windows" {
			perfBinaryFixture.path += ".exe"
		}
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", perfBinaryFixture.path, "./cmd/moai")
		cmd.Dir = projectRoot(t)
		cmd.WaitDelay = time.Second
		if out, err := cmd.CombinedOutput(); err != nil {
			perfBinaryFixture.err = fmt.Errorf("build moai: %w\n%s", err, out)
		}
	})
	if perfBinaryFixture.err != nil {
		t.Fatal(perfBinaryFixture.err)
	}
	return perfBinaryFixture.path
}

func TestProfilingWorkerFailureReturns(t *testing.T) {
	// A missing executable used to call Fatalf in the worker and strand the
	// collector forever. This reaches the real exec failure without a build.
	_, err := runProfilingBatches(t, filepath.Join(t.TempDir(), "missing-moai"), t.TempDir(), "", 2, 1)
	if err == nil {
		t.Fatal("missing executable must fail the batch")
	}
}

func TestProfilingHookCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "hook")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexec sleep 30\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if _, _, err := runSingleHook(ctx, bin, dir, ""); err == nil || ctx.Err() == nil {
		t.Fatalf("silent hook must be cancelled: err=%v context=%v", err, ctx.Err())
	}
}
