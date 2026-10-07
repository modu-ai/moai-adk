package cli

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

// The executable is immutable within a package run; each caller still owns
// its project, environment, and child processes. TestMain removes it after
// every test has finished, rather than the first caller's t.TempDir cleanup.
var cliBinaryFixture struct {
	once sync.Once
	dir  string
	path string
	err  error
}

func buildMoaiBinary(t *testing.T) string {
	t.Helper()
	cliBinaryFixture.once.Do(func() {
		cliBinaryFixture.dir, cliBinaryFixture.err = os.MkdirTemp("", "moai-cli-test-")
		if cliBinaryFixture.err != nil {
			return
		}
		cliBinaryFixture.path = filepath.Join(cliBinaryFixture.dir, "moai")
		if runtime.GOOS == "windows" {
			cliBinaryFixture.path += ".exe"
		}
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
		defer cancel()
		// The import path also works from the internal/cli package directory.
		cmd := exec.CommandContext(ctx, "go", "build", "-o", cliBinaryFixture.path,
			"github.com/modu-ai/moai-adk/cmd/moai")
		cmd.WaitDelay = time.Second
		if out, err := cmd.CombinedOutput(); err != nil {
			cliBinaryFixture.err = fmt.Errorf("build moai: %w\n%s", err, out)
		}
	})
	if cliBinaryFixture.err != nil {
		t.Fatal(cliBinaryFixture.err)
	}
	return cliBinaryFixture.path
}

func cleanupCLIBinaryFixture() error {
	if cliBinaryFixture.dir == "" {
		return nil
	}
	return os.RemoveAll(cliBinaryFixture.dir)
}

func TestCLIBinaryFixtureSurvivesCallingTest(t *testing.T) {
	var first string
	t.Run("first consumer", func(t *testing.T) {
		first = buildMoaiBinary(t)
	})
	// A fixture owned by the first subtest's TempDir would be gone here.
	t.Run("later consumer", func(t *testing.T) {
		bin := buildMoaiBinary(t)
		if bin != first {
			t.Fatalf("binary rebuilt between consumers: %s != %s", bin, first)
		}
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()
		if out, err := exec.CommandContext(ctx, bin, "version").CombinedOutput(); err != nil {
			t.Fatalf("shared executable after first consumer cleanup: %v\n%s", err, out)
		}
	})
}
