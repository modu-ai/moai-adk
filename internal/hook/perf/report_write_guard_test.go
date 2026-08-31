package perf

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	// guardChildEnv marks a `go test` process spawned by the guard. The guard
	// skips itself when it sees the sentinel, so the package-glob child cannot
	// recurse into another guard run.
	guardChildEnv = "MOAI_HOOK_PERF_GUARD_CHILD"
	// perfUpdateEnv is the opt-in write gate consumed by updatePerfReports.
	perfUpdateEnv = "MOAI_HOOK_PERF_UPDATE"
	// perfSkipEnv is the package's pre-existing cost escape hatch.
	perfSkipEnv = "MOAI_HOOK_PERF_SKIP"
)

// guardFixtures are the two tracked report files, relative to the project root.
var guardFixtures = []string{
	filepath.Join(".moai", "specs", "SPEC-HOOK-PRETOOL-PERF-001", "baseline.md"),
	filepath.Join(".moai", "specs", "SPEC-HOOK-PRETOOL-PERF-001", "postchange.md"),
}

// childEnvDenied names the variables the guard REMOVES from the inherited
// environment when building a child environment. Everything else passes through.
//
// This is deliberately a denylist, not an allowlist. An earlier revision
// enumerated the ~17 variables to copy, which named only POSIX ones and so
// could not hand a Windows child LOCALAPPDATA, TMP, TEMP, USERPROFILE or
// SystemRoot — with GOCACHE unset, os.UserCacheDir reads LocalAppData and the
// toolchain fails with "GOCACHE is not defined". An allowlist breaks silently
// whenever a platform or toolchain needs a variable nobody enumerated, and
// Windows was merely the first instance. A denylist targets the actual hazard
// by name instead of hoping the enumeration never omits something.
//
// Do NOT "tighten" this back into an allowlist.
//
// What is removed and why:
//   - MOAI_HOOK_PERF_UPDATE: the reason REQ-PFW-005 exists. A parent-set gate
//     leaking into the NEGATIVE child makes it write, turning the guard red for
//     a reason that is not a regression.
//   - MOAI_HOOK_PERF_SKIP: belt-and-braces. Unreachable today, because the
//     guard self-skips when it sees this variable and so never reaches child
//     construction with it set; removed anyway so the child stays correct if
//     that ordering ever changes.
//   - GOFLAGS: can carry -short or -count and would silently reshape the child
//     invocation out from under the assertions.
var childEnvDenied = []string{
	perfUpdateEnv,
	perfSkipEnv,
	"GOFLAGS",
}

// TestPerfReportWriteGuard is the regression guard for SPEC-PERF-FIXTURE-WRITE-001.
//
// It spawns `go test` as a child process twice and decides on a SHA-256 content
// hash of each tracked fixture, taken before and after each child:
//
//   - negative: the CI invocation shape (package glob, no -short, gate off) must
//     leave both fixtures byte-identical.
//   - positive: the same package with MOAI_HOOK_PERF_UPDATE=1 must move both.
//
// Content, not modification time, is the key — git measures content, so a file
// whose bytes are unchanged cannot be swept into a commit. A no-op rewrite that
// only touches mtime is deliberately NOT detected; it is not the defect.
//
// The guard restores the captured original bytes on every NORMAL exit path,
// including the path where its own assertion fails, so a RED guard never becomes
// a fresh instance of the defect it exists to detect. Two residuals are known and
// accepted:
//
//   - t.Cleanup does not run on a kill (SIGINT/SIGKILL) or on a `go test -timeout`
//     panic. On those paths the fixtures stay modified.
//   - While the positive leg runs (~10s) the two tracked fixtures are genuinely
//     modified. A concurrent lane reading the tree in that window sees them as `M`.
func TestPerfReportWriteGuard(t *testing.T) {
	// Deliberately NOT t.Parallel(): the restore contract below assumes this
	// test's t.Cleanup fires before any later test in the package begins.
	if os.Getenv(guardChildEnv) != "" {
		t.Skip("perf-guard: guard-child sentinel set, skipping to break recursion")
	}
	if testing.Short() {
		t.Skip("perf-guard: skipping in short mode")
	}
	if os.Getenv(perfSkipEnv) != "" {
		t.Skip("perf-guard: " + perfSkipEnv + " set")
	}

	root := projectRoot(t)

	// Capture the originals. A missing fixture is a failure, never a skip: a
	// regression that deletes the file must not read as green.
	originals := make(map[string][]byte, len(guardFixtures))
	for _, rel := range guardFixtures {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("perf-guard: read fixture %s: %v", rel, err)
		}
		originals[rel] = b
	}
	t.Cleanup(func() {
		for _, rel := range guardFixtures {
			path := filepath.Join(root, rel)
			cur, err := os.ReadFile(path)
			if err == nil && string(cur) == string(originals[rel]) {
				continue
			}
			if werr := os.WriteFile(path, originals[rel], 0o644); werr != nil {
				t.Errorf("perf-guard: restore %s: %v", rel, werr)
			}
		}
	})

	before := hashGuardFixtures(t, root)

	// Negative direction: the shape CI actually runs today.
	negOut, negRC := runGuardChild(t, root, false)
	t.Logf("perf-guard: negative child rc=%d", negRC)
	t.Logf("perf-guard: negative child combined output:\n%s", negOut)
	if negRC != 0 {
		t.Fatalf("perf-guard: negative child did not complete (rc=%d) — "+
			"the hash comparison would be meaningless, see the combined output above", negRC)
	}
	afterNegative := hashGuardFixtures(t, root)
	var moved []string
	for _, rel := range guardFixtures {
		t.Logf("perf-guard: %s before=%s after-negative=%s", rel, before[rel], afterNegative[rel])
		if afterNegative[rel] != before[rel] {
			moved = append(moved, rel)
		}
	}
	if len(moved) > 0 {
		t.Errorf("perf-guard: fixture content changed after a gate-off run: %s", strings.Join(moved, ", "))
		return
	}

	// Positive direction: without it, a "never writes at all" regression passes
	// on the negative assertion alone.
	posOut, posRC := runGuardChild(t, root, true)
	t.Logf("perf-guard: positive child rc=%d", posRC)
	t.Logf("perf-guard: positive child combined output:\n%s", posOut)
	if posRC != 0 {
		t.Fatalf("perf-guard: positive child did not complete (rc=%d) — "+
			"see the combined output above", posRC)
	}
	afterPositive := hashGuardFixtures(t, root)
	var stuck []string
	for _, rel := range guardFixtures {
		t.Logf("perf-guard: %s before=%s after-positive=%s", rel, before[rel], afterPositive[rel])
		if afterPositive[rel] == before[rel] {
			stuck = append(stuck, rel)
		}
	}
	if len(stuck) > 0 {
		t.Errorf("perf-guard: fixture content did not move under %s=1: %s",
			perfUpdateEnv, strings.Join(stuck, ", "))
	}
}

// runGuardChild runs `go test` over this package in a child process and returns
// its combined output and exit code. gateOn selects the positive leg.
func runGuardChild(t *testing.T, root string, gateOn bool) (string, int) {
	t.Helper()
	args := []string{"test", "./internal/hook/perf/...", "-count=1"}
	if gateOn {
		// The positive leg only has to observe that a write happens, so it is
		// narrowed to save cost. The negative leg keeps the package glob,
		// because that is the shape under which CI rewrites the fixtures.
		args = append(args, "-run", "^TestPreToolProfiling")
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Env = guardChildEnvironment(gateOn)
	out, err := cmd.CombinedOutput()
	return string(out), guardChildExitCode(t, err)
}

// guardChildEnvironment builds a child environment by inheriting the parent's
// and removing every variable in childEnvDenied, then appending the sentinel.
func guardChildEnvironment(gateOn bool) []string {
	denied := make(map[string]struct{}, len(childEnvDenied))
	for _, key := range childEnvDenied {
		denied[key] = struct{}{}
	}
	parent := os.Environ()
	env := make([]string, 0, len(parent)+2)
	for _, entry := range parent {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if _, blocked := denied[key]; blocked {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, guardChildEnv+"=1")
	if gateOn {
		env = append(env, perfUpdateEnv+"=1")
	}
	return env
}

// guardChildExitCode extracts the child's exit status. A failure to start the
// process at all is distinct from a non-zero exit and aborts the guard.
func guardChildExitCode(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	t.Fatalf("perf-guard: child process could not be run: %v", err)
	return -1
}

// hashGuardFixtures returns the SHA-256 hex digest of each fixture's content.
func hashGuardFixtures(t *testing.T, root string) map[string]string {
	t.Helper()
	sums := make(map[string]string, len(guardFixtures))
	for _, rel := range guardFixtures {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("perf-guard: read fixture %s: %v", rel, err)
		}
		sum := sha256.Sum256(b)
		sums[rel] = hex.EncodeToString(sum[:])
	}
	return sums
}
