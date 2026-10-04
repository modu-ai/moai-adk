package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// saveCCVersionSeams captures the version-read seams and restores them at
// cleanup, so a failing test cannot leak an injected fixture into a later
// test in the package run.
func saveCCVersionSeams(t *testing.T) {
	t.Helper()
	prevMapping := readProcessMapping
	prevAlive := pidIsAlive
	prevLookPath := ccInstalledLookPath
	t.Cleanup(func() {
		readProcessMapping = prevMapping
		pidIsAlive = prevAlive
		ccInstalledLookPath = prevLookPath
	})
}

// TestRunningVersionFromInjectedMapping (AC-SCV-001, REQ-SCV-001) — a fixture
// pid→mapping table in which pid 4242 maps to a text mapping carrying
// …/claude/versions/2.1.281 resolves to running version 2.1.281. The fixture
// is lsof-shaped on purpose: `-d txt` also lists mapped frameworks and dyld,
// so the parse must anchor on the line naming the claude binary itself — a
// version-shaped segment on a library mapping must not satisfy the read.
func TestRunningVersionFromInjectedMapping(t *testing.T) {
	saveCCVersionSeams(t)
	pidIsAlive = func(int) bool { return true }
	readProcessMapping = func(pid int) (string, bool) {
		if pid != 4242 {
			return "", false
		}
		// The satisfying line is the NATIVE INSTALLER'S real shape, measured
		// live (plan §F.4): the shipped binary file is named by its version,
		// so the mapping line's path ends at …/claude/versions/2.1.281 with no
		// trailing "/claude". The library lines around it are the anchor
		// hazard: CoreFoundation's Versions/9 segment is version-shaped (its
		// capital V is what a case-sensitive match refuses).
		mapping := strings.Join([]string{
			"2.1.281  4242 dev  txt  REG  1,4  123  456 /System/Library/Frameworks/CoreFoundation.framework/Versions/9/CoreFoundation",
			"2.1.281  4242 dev  txt  REG  1,4  234  567 /Users/dev/.local/share/claude/versions/2.1.281",
			"2.1.281  4242 dev  txt  REG  1,4  345  678 /usr/lib/dyld",
		}, "\n")
		return mapping, true
	}

	view := ResolveCCVersions(4242)
	if view.Running != "2.1.281" {
		t.Fatalf("ResolveCCVersions(4242).Running = %q, want 2.1.281 (a library mapping's version shape must not satisfy the read)", view.Running)
	}

	t.Run("binary-named-file shape also satisfies", func(t *testing.T) {
		readProcessMapping = func(int) (string, bool) {
			return "claude  4242 dev  txt  REG  1,4  234  567 /opt/installs/claude/versions/2.0.9/bin/claude", true
		}
		if got := (ResolveCCVersions(4242)).Running; got != "2.0.9" {
			t.Fatalf("…/versions/2.0.9/bin/claude resolved %q, want 2.0.9", got)
		}
	})
	t.Run("npm layout satisfies", func(t *testing.T) {
		readProcessMapping = func(int) (string, bool) {
			return "node  4242 dev  txt  REG  1,4  234  567 /opt/node/lib/node_modules/@anthropic-ai/claude-code/2.1.284/cli", true
		}
		if got := (ResolveCCVersions(4242)).Running; got != "2.1.284" {
			t.Fatalf("…/claude-code/2.1.284/cli resolved %q, want 2.1.284", got)
		}
	})
}

// TestRunningVersionFromDeletedBinary (card-review round 1, P2) — Linux
// names a deleted executable's /proc/<pid>/exe value "<path> (deleted)":
// the binary was replaced on disk while the process still runs it — exactly
// the staleness case this SPEC exists to surface. The read boundary strips
// the suffix before the anchor and the version parse, in both install
// shapes, so the version is still recovered.
func TestRunningVersionFromDeletedBinary(t *testing.T) {
	saveCCVersionSeams(t)
	pidIsAlive = func(int) bool { return true }
	readProcessMapping = func(pid int) (string, bool) {
		if pid != 4242 {
			return "", false
		}
		return "/Users/dev/.local/share/claude/versions/2.1.287 (deleted)", true
	}
	if got := (ResolveCCVersions(4242)).Running; got != "2.1.287" {
		t.Fatalf("deleted-binary path resolved %q, want 2.1.287", got)
	}

	t.Run("binary-named tail with the deleted suffix", func(t *testing.T) {
		readProcessMapping = func(int) (string, bool) {
			return "/opt/installs/claude/versions/2.1.281/claude (deleted)", true
		}
		if got := (ResolveCCVersions(4242)).Running; got != "2.1.281" {
			t.Fatalf("deleted binary-named path resolved %q, want 2.1.281", got)
		}
	})
}

// TestInstalledVersionFromResolvedPath (AC-SCV-002, REQ-SCV-002) — a fixture
// PATH directory whose claude entry is a symlink resolving to a versioned
// install path reports the path's version segment, for both house shapes
// (versions/<X.Y.Z> and claude-code/<X.Y.Z>). The symlink resolution is real
// (files under t.TempDir); only the PATH lookup itself is seamed. The
// version-segment parser is asserted directly alongside, including its
// negatives.
func TestInstalledVersionFromResolvedPath(t *testing.T) {
	saveCCVersionSeams(t)

	t.Run("versions shape", func(t *testing.T) {
		got := installedVersionViaSymlink(t, "versions/2.1.288")
		if got != "2.1.288" {
			t.Fatalf("installed version = %q, want 2.1.288", got)
		}
	})
	t.Run("claude-code shape", func(t *testing.T) {
		got := installedVersionViaSymlink(t, "claude-code/2.1.284")
		if got != "2.1.284" {
			t.Fatalf("installed version = %q, want 2.1.284", got)
		}
	})

	// Direct parser assertions: both shapes, and the negatives the read
	// depends on (an unversioned path renders no segment; a prefix that merely
	// contains the word "versions" must not match).
	direct := map[string]string{
		"/Users/dev/.local/share/claude/versions/2.1.281":                  "2.1.281",
		"/Users/dev/.local/share/claude/versions/2.1.281/claude":           "2.1.281",
		"/opt/node/lib/node_modules/@anthropic-ai/claude-code/2.1.284/cli": "2.1.284",
		"/usr/local/bin/claude":                                            "",
		"/opt/homebrew/Caskroom/conversions/2.1/x":                         "",
	}
	for path, want := range direct {
		if got := versionSegmentFromPath(path); got != want {
			t.Errorf("versionSegmentFromPath(%q) = %q, want %q", path, got, want)
		}
	}
}

// installedVersionViaSymlink builds <dir>/bin/claude → <dir>/<shape>/claude
// (a real symlink under t.TempDir), points the installed read's PATH seam at
// it, and returns the installed version the read reports.
func installedVersionViaSymlink(t *testing.T, shape string) string {
	t.Helper()
	dir := t.TempDir()
	real := filepath.Join(dir, shape)
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", real, err)
	}
	binary := filepath.Join(real, "claude")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write fixture binary: %v", err)
	}
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	link := filepath.Join(binDir, "claude")
	if err := os.Symlink(binary, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	ccInstalledLookPath = func() (string, bool) { return link, true }
	return installedCCVersion()
}

// TestVersionDegradationRendersUnknown (AC-SCV-003, REQ-SCV-003) — each of
// the four degradations renders `unknown`, never an error and never an
// inferred value: a dead pid, an injected probe error, a resolved path with
// no version segment, and the unsupported-platform read.
func TestVersionDegradationRendersUnknown(t *testing.T) {
	t.Run("dead pid", func(t *testing.T) {
		saveCCVersionSeams(t)
		// Liveness says not-alive; the mapping seam would return a version if
		// it were ever consulted, proving the liveness gate short-circuits
		// before the platform read.
		pidIsAlive = func(int) bool { return false }
		readProcessMapping = func(int) (string, bool) {
			return "/x/claude/versions/9.9.9/claude", true
		}
		if got := (ResolveCCVersions(4242)).Running; got != UnknownCCVersion {
			t.Fatalf("dead pid running = %q, want %q", got, UnknownCCVersion)
		}
	})
	t.Run("probe error", func(t *testing.T) {
		saveCCVersionSeams(t)
		pidIsAlive = func(int) bool { return true }
		readProcessMapping = func(int) (string, bool) { return "", false }
		if got := (ResolveCCVersions(4242)).Running; got != UnknownCCVersion {
			t.Fatalf("probe error running = %q, want %q", got, UnknownCCVersion)
		}
	})
	t.Run("resolved path with no version segment", func(t *testing.T) {
		saveCCVersionSeams(t)
		pidIsAlive = func(int) bool { return true }
		readProcessMapping = func(int) (string, bool) {
			return "/usr/local/bin/claude", true
		}
		view := ResolveCCVersions(4242)
		if view.Running != UnknownCCVersion {
			t.Fatalf("unversioned running = %q, want %q", view.Running, UnknownCCVersion)
		}
		ccInstalledLookPath = func() (string, bool) { return "/usr/local/bin/claude", true }
		if got := installedCCVersion(); got != UnknownCCVersion {
			t.Fatalf("unversioned installed = %q, want %q", got, UnknownCCVersion)
		}
	})
	t.Run("unsupported platform read", func(t *testing.T) {
		saveCCVersionSeams(t)
		// Stands in for the ccversion_other.go build: that reader always
		// reports not-ok (the GOOS=windows cross-build compiles it; its
		// runtime path cannot execute on this machine). The contract under
		// test is the shared view's answer to an unsupported reader.
		pidIsAlive = func(int) bool { return true }
		readProcessMapping = func(int) (string, bool) { return "", false }
		if got := (ResolveCCVersions(4242)).Running; got != UnknownCCVersion {
			t.Fatalf("unsupported platform running = %q, want %q", got, UnknownCCVersion)
		}
	})
}
