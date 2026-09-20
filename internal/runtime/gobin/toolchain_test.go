package gobin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/runtime/gobin"
)

// TestDetect_DoesNotDownloadToolchain is the characterization test for
// SPEC-GOBIN-GOTOOLCHAIN-001 (AC-GGT-003).
//
// Detect spawns "go env GOBIN" and "go env GOPATH". Where the PATH go is older
// than the module's go directive, that child go resolves a toolchain itself and
// downloads a full toolchain into whatever HOME points at. The module cache it
// writes is read-only, which also defeats t.TempDir cleanup.
//
// The firing condition is the invocation form, not the environment: under
// "go test" the runtime prepends the resolved toolchain's bin to the test
// process's PATH, so the child go has nothing to resolve and nothing downloads.
// Under a precompiled test binary run directly, that PATH injection is absent.
// A green "go test" here is therefore NOT evidence the assertion is wrong; see
// .moai/specs/SPEC-GOBIN-GOTOOLCHAIN-001/spec.md section 1.
func TestDetect_DoesNotDownloadToolchain(t *testing.T) {
	tempHome := t.TempDir()

	// Bind HOME to the temp dir and leave GOBIN/GOPATH unset so the child go
	// derives its module cache from HOME rather than from the real GOPATH.
	t.Setenv("HOME", tempHome)
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "")

	_ = gobin.Detect(tempHome)

	pattern := filepath.Join(tempHome, "go", "pkg", "mod", "golang.org", "toolchain@*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob %s: %v", pattern, err)
	}

	// Empty, not absent: no match at all is a pass, and a matched directory
	// holding no entries is also a pass.
	for _, match := range matches {
		entries, err := os.ReadDir(match)
		if err != nil {
			t.Fatalf("read toolchain dir %s: %v", match, err)
		}
		if len(entries) > 0 {
			t.Errorf(
				"gobin.Detect downloaded a Go toolchain into the test HOME: %s holds %d entries (glob %s); the go env subprocesses must run with GOTOOLCHAIN=local",
				match, len(entries), pattern,
			)
		}
	}
}
