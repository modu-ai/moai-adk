package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// session_start_archive_guard_test.go — REQ-SSP-011 [HARD] / AC-SSP-011.
//
// The archive capability relocates SPEC directories. It must NEVER be reachable
// from the session-start critical path: a Claude Code session launch would then be
// blocked behind a filesystem mutation whose cost grows with the corpus — exactly
// the class of defect SPEC-SESSIONSTART-PERF-001 exists to remove. Archive runs
// on-demand (`moai spec archive`) and at /moai sync SPEC-close ONLY.
//
// This is a static reachability guard, not a runtime assertion: it fails the build
// the moment an archive symbol is referenced anywhere in the hook package, which is
// the only place the session-start handler's call graph is rooted.

// archiveSymbols are the exported entry points of the archive capability
// (internal/spec/archive.go). None may appear in a hook-package source file.
var archiveSymbols = []string{
	"PlanArchive",
	"ExecuteArchive",
	"ArchiveOptions",
	"ArchivePlan",
	"ArchiveCandidate",
	"IsArchiveTerminalStatus",
}

// TestSessionStart_NoArchiveInvocation asserts the session-start handler carries no
// archive call site.
func TestSessionStart_NoArchiveInvocation(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile("session_start.go")
	if err != nil {
		t.Fatalf("read session_start.go: %v", err)
	}

	assertNoArchiveSymbols(t, "session_start.go", string(src))
}

// TestHookPackage_NoArchiveInvocation widens the guard to the whole hook package.
// session_start.go delegates to helpers across the package, so a clean
// session_start.go alone would not prove unreachability — an archive call hidden in
// a helper the handler invokes would still sit on the critical path.
func TestHookPackage_NoArchiveInvocation(t *testing.T) {
	t.Parallel()

	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob hook sources: %v", err)
	}
	if len(sources) == 0 {
		t.Fatal("no hook sources found — the guard would pass vacuously")
	}

	scanned := 0
	for _, path := range sources {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		assertNoArchiveSymbols(t, path, string(src))
		scanned++
	}

	if scanned == 0 {
		t.Fatal("scanned 0 non-test hook sources — the guard would pass vacuously")
	}
	t.Logf("archive-reachability guard: scanned %d non-test hook sources, 0 archive call sites", scanned)
}

// assertNoArchiveSymbols reports any archive symbol reference outside comments.
func assertNoArchiveSymbols(t *testing.T, path, src string) {
	t.Helper()

	for lineNo, line := range strings.Split(src, "\n") {
		code, _, _ := strings.Cut(line, "//")
		if strings.TrimSpace(code) == "" {
			continue
		}

		for _, sym := range archiveSymbols {
			if strings.Contains(code, sym) {
				t.Errorf("%s:%d references archive symbol %q — archive must never sit on the session-start critical path (REQ-SSP-011)\n\t%s",
					path, lineNo+1, sym, strings.TrimSpace(line))
			}
		}
	}
}
