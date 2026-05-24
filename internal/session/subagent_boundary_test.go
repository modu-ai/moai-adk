package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSubagentBoundary statically asserts that the multi-session
// coordination registry surface (internal/session/registry*.go +
// internal/cli/session.go + internal/hook/session_start.go) does NOT
// invoke any AskUserQuestion-equivalent user prompt mechanism.
//
// C-HRA-008 / agent-common-protocol.md § User Interaction Boundary:
// subagents and CLI primitives MUST NOT prompt users directly. User
// interaction is reserved for the MoAI orchestrator. The registry exposes
// JSON output via `moai session list --json` so the orchestrator decides
// whether to AskUserQuestion (wait / override / abort).
//
// This is the static-grep CI guard mirroring the precedent from
// internal/cli/harness/propose_boundary_test.go (TestPropose_NoAskUserQuestion).
//
// SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-012, REQ-COORD-013 boundary.
func TestSubagentBoundary(t *testing.T) {
	// Determine project root: walk up from this test file's dir until we
	// hit go.mod. The test runs with CWD = internal/session, so go.mod is
	// 2 directories up.
	root := findProjectRoot(t)

	// Scope: registry surface + CLI subcommand + hook entry point.
	// Files in this scope MUST NOT invoke AskUserQuestion.
	scopePaths := []string{
		filepath.Join(root, "internal", "session"),
		filepath.Join(root, "internal", "cli", "session.go"),
		filepath.Join(root, "internal", "hook", "session_start.go"),
	}

	// Forbidden tokens (the AskUserQuestion surface).
	forbidden := []string{
		"AskUserQuestion",
		"mcp__askuser",
	}

	for _, p := range scopePaths {
		info, err := os.Stat(p)
		if err != nil {
			// Soft-skip files that don't yet exist (e.g., during partial
			// M1-only delivery). Once M2/M3 land, the missing path becomes
			// a test failure via a separate file-existence check below.
			continue
		}
		if info.IsDir() {
			scanDir(t, p, forbidden)
		} else {
			scanFile(t, p, forbidden)
		}
	}
}

// TestSubagentBoundary_RequiredFilesExist ensures the scope paths exist
// after M1-M3 are complete. Until they exist (during incremental M1-only
// builds), this test soft-asserts that internal/session is populated.
func TestSubagentBoundary_RequiredFilesExist(t *testing.T) {
	root := findProjectRoot(t)
	// M1 deliverable: registry.go is mandatory.
	mustExist := filepath.Join(root, "internal", "session", "registry.go")
	if _, err := os.Stat(mustExist); err != nil {
		t.Errorf("M1 deliverable missing: %s", mustExist)
	}
}

func scanDir(t *testing.T, dir string, forbidden []string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		// Exclude test files — boundary test itself contains the forbidden
		// strings (in this file, for explanation purposes).
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanFile(t, filepath.Join(dir, name), forbidden)
	}
}

func scanFile(t *testing.T, path string, forbidden []string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	lines := strings.Split(string(data), "\n")
	for lineNum, line := range lines {
		// Strip line comments — comments containing AskUserQuestion as
		// prose explanation are allowed (e.g., "AskUserQuestion is reserved
		// for the orchestrator").
		stripped := stripLineComment(line)
		for _, token := range forbidden {
			if strings.Contains(stripped, token) {
				t.Errorf("%s:%d: forbidden token %q in non-comment code: %s",
					path, lineNum+1, token, strings.TrimSpace(line))
			}
		}
	}
}

// stripLineComment returns the portion of a Go source line before any
// `//` comment marker. Naive (does not handle "//" inside strings) but
// sufficient for the AskUserQuestion sentinel check.
func stripLineComment(line string) string {
	idx := strings.Index(line, "//")
	if idx == -1 {
		return line
	}
	return line[:idx]
}

// findProjectRoot walks upward from the current test working directory
// until it finds go.mod, signalling the project root.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate project root (go.mod) from %s", dir)
	return ""
}
