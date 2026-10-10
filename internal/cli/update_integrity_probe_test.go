// SPEC-UPDATE-MIGRATION-FIX-001 M3 (REQ-UMF-001..003; AC-UMF-001, AC-UMF-002):
// tests for the managed-surface integrity probe on the version-match skip path.
//
// The tests observe the probe through the real `moai update` command, not
// through its implementation, so they compile and fail on assertions before the
// probe exists.
package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// integrityProbeRelPaths are the representative project paths the probe checks.
// They are written as literals on purpose: the tests assert on the rendered rows.
var integrityProbeRelPaths = []string{
	".claude/settings.json",
	".moai/manifest.json",
}

// integrityRows returns the output lines that are integrity rows of the probe:
// lines that carry the Integrity label and name one of the representative paths.
// The clean-reinstall path prints its own "Integrity check" lines, which name no
// representative path and are therefore not counted.
func integrityRows(out string) []string {
	var rows []string
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "Integrity") {
			continue
		}
		for _, rel := range integrityProbeRelPaths {
			if strings.Contains(line, rel) {
				rows = append(rows, line)
				break
			}
		}
	}
	return rows
}

// runUpdateFixtureErr runs the real `moai update` command in root and returns the
// combined output together with the command's error. yes selects the unattended
// flag; check and force stay off. The flags are restored at cleanup, because
// updateCmd is a package-level command shared by every update test.
func runUpdateFixtureErr(t *testing.T, root string, yes bool) (string, error) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir fixture: %v", err)
	}

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = &Dependencies{UpdateChecker: &mockUpdateChecker{}}

	want := map[string]string{"check": "false", "force": "false", "yes": fmt.Sprintf("%t", yes)}
	for name, value := range want {
		name, value := name, value
		prev := updateCmd.Flags().Lookup(name).Value.String()
		t.Cleanup(func() { _ = updateCmd.Flags().Set(name, prev) })
		if err := updateCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set --%s: %v", name, err)
		}
	}

	var buf bytes.Buffer
	updateCmd.SetOut(&buf)
	updateCmd.SetErr(&buf)
	updateCmd.SetContext(context.Background())
	runErr := updateCmd.RunE(updateCmd, []string{})
	return buf.String(), runErr
}

// setupSyncedV3Fixture lays down a v3 project and runs one full sync, so the
// project carries its own managed surface and the stamped template version. A
// second update then matches that version and takes the skip path.
func setupSyncedV3Fixture(t *testing.T, root string) {
	t.Helper()
	writeV3ProjectFixture(t, root)
	if out, err := runUpdateFixtureErr(t, root, true); err != nil {
		t.Logf("first (full sync) update returned: %v\n%s", err, out)
	}
	for _, rel := range integrityProbeRelPaths {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("precondition: the full sync must produce %s: %v", rel, err)
		}
	}
}

// TestRunUpdate_VersionMatch_RunsIntegrityProbe pins REQ-UMF-001 and REQ-UMF-002
// on the version-match entry: an intact project prints no integrity row (no false
// positive), and a deleted representative file is named by one row while the
// update still exits with no error.
func TestRunUpdate_VersionMatch_RunsIntegrityProbe(t *testing.T) {
	root := t.TempDir()
	setupSyncedV3Fixture(t, root)

	out, err := runUpdateFixtureErr(t, root, true)
	if err != nil {
		t.Fatalf("an intact version-matched update must succeed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Up to date") {
		t.Fatalf("precondition: the version-match skip must run on the second update:\n%s", out)
	}
	if rows := integrityRows(out); len(rows) != 0 {
		t.Fatalf("false positive: an intact project printed integrity rows %q\n%s", rows, out)
	}

	if err := os.Remove(filepath.Join(root, ".claude", "settings.json")); err != nil {
		t.Fatalf("damage fixture: %v", err)
	}
	out, err = runUpdateFixtureErr(t, root, true)
	if err != nil {
		t.Fatalf("REQ-UMF-002: a damaged representative path must not fail the update: %v\n%s", err, out)
	}
	rows := integrityRows(out)
	if len(rows) != 1 || !strings.Contains(rows[0], ".claude/settings.json") {
		t.Fatalf("REQ-UMF-001: the version-match run must print one integrity row naming .claude/settings.json; rows=%q\n%s",
			rows, out)
	}
}

// TestRunUpdate_UserCancelled_SkipsIntegrityProbe pins REQ-UMF-003. A user-cancelled
// merge returns the same skipped=true as a version match, but it is not a version
// match, so it must print no integrity row even though the fixture has a missing
// representative path that a probe would name.
func TestRunUpdate_UserCancelled_SkipsIntegrityProbe(t *testing.T) {
	root := t.TempDir()
	// No template_version stamp: the sync is not version-matched, so the merge
	// confirmation is reached, and .claude/settings.json is absent.
	writeV3ProjectFixture(t, root)

	orig := confirmViaPreviewFn
	confirmViaPreviewFn = func(_ merge.MergeAnalysis, _ string) (bool, error) { return false, nil }
	t.Cleanup(func() { confirmViaPreviewFn = orig })

	out, err := runUpdateFixtureErr(t, root, false)
	if err != nil {
		t.Fatalf("a user-cancelled merge must not fail the update: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Merge cancelled by user") {
		t.Fatalf("precondition: the merge must be cancelled through the confirmation seam:\n%s", out)
	}
	if rows := integrityRows(out); len(rows) != 0 {
		t.Fatalf("REQ-UMF-003: a user-cancelled merge printed integrity rows %q\n%s", rows, out)
	}
}

// TestIntegrityProbe_FailOpen pins REQ-UMF-002 on an unreadable representative
// path: a directory where the manifest file belongs yields exactly one integrity
// row for that path, and the update returns no error.
func TestIntegrityProbe_FailOpen(t *testing.T) {
	root := t.TempDir()
	setupSyncedV3Fixture(t, root)

	manifest := filepath.Join(root, ".moai", "manifest.json")
	if err := os.Remove(manifest); err != nil {
		t.Fatalf("damage fixture: %v", err)
	}
	if err := os.Mkdir(manifest, 0o755); err != nil {
		t.Fatalf("plant a directory where the manifest file belongs: %v", err)
	}

	out, err := runUpdateFixtureErr(t, root, true)
	if err != nil {
		t.Fatalf("REQ-UMF-002: a probe fault must never return an error: %v\n%s", err, out)
	}
	count := 0
	for _, row := range integrityRows(out) {
		if strings.Contains(row, ".moai/manifest.json") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("REQ-UMF-002: want exactly one integrity row for the unreadable manifest, got %d\n%s", count, out)
	}
}

// writeProbeFile writes content to root/rel, creating parent directories.
func writeProbeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// probeReturnBound is how long one damageReason call may take before the test
// treats it as blocked. A classification is one stat plus at most one bounded
// read, so a healthy call returns in microseconds; the bound only has to tell a
// block from a slow machine.
const probeReturnBound = 5 * time.Second

// damageReasonWithin runs e.damageReason(root) on its own goroutine and waits at
// most bound for the classification. returned is false when the call is still
// blocked when the bound expires.
func damageReasonWithin(e integrityProbeEntry, root string, bound time.Duration) (reason string, returned bool) {
	done := make(chan string, 1)
	go func() { done <- e.damageReason(root) }()
	select {
	case reason = <-done:
		return reason, true
	case <-time.After(bound):
		return "", false
	}
}

// unblockPipe releases a goroutine stuck in the open-for-read of a named pipe at
// abs: opening the pipe for writing completes the reader's open, and closing the
// writer ends its stream. It does nothing for any other file type, so a bounded
// test can call it unconditionally after a timeout.
func unblockPipe(abs string) {
	info, err := os.Lstat(abs)
	if err != nil || info.Mode()&fs.ModeNamedPipe == 0 {
		return
	}
	if w, err := os.OpenFile(abs, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
		_ = w.Close()
	}
}

// TestIntegrityProbeEntry_DamageReasons pins every classification one
// representative entry can return. The end-to-end tests reach only some of them
// (the damage a version-matched run can observe); the rest are pinned here,
// through the entry check itself.
func TestIntegrityProbeEntry_DamageReasons(t *testing.T) {
	settings := integrityProbeEntry{rel: ".claude/settings.json", check: probeJSON}
	manifest := integrityProbeEntry{rel: ".moai/manifest.json", check: probeJSON}

	cases := []struct {
		name  string
		entry integrityProbeEntry
		setup func(t *testing.T, root string)
		want  string
	}{
		{"missing", settings, func(*testing.T, string) {}, "missing"},
		{"directory_in_place_of_file", settings, func(t *testing.T, root string) {
			if err := os.MkdirAll(filepath.Join(root, ".claude", "settings.json"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, "not a file"},
		{"intact_json", settings, func(t *testing.T, root string) {
			writeProbeFile(t, root, ".claude/settings.json", "{}")
		}, ""},
		{"unparseable_json", settings, func(t *testing.T, root string) {
			writeProbeFile(t, root, ".claude/settings.json", "{not json")
		}, "unparseable"},
		{"stat_fails_under_a_file", settings, func(t *testing.T, root string) {
			if runtime.GOOS == "windows" {
				t.Skip("a path through a file reports not-exist on windows, not an unreadable stat")
			}
			writeProbeFile(t, root, ".claude", "a file where the directory belongs")
		}, "unreadable"},
		{"read_denied", settings, func(t *testing.T, root string) {
			if runtime.GOOS == "windows" || os.Geteuid() == 0 {
				t.Skip("file permissions do not deny reads on this platform or for root")
			}
			writeProbeFile(t, root, ".claude/settings.json", "{}")
			abs := filepath.Join(root, ".claude", "settings.json")
			if err := os.Chmod(abs, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(abs, 0o644) })
		}, "unreadable"},
		{"fifo_in_place_of_json", manifest, func(t *testing.T, root string) {
			if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := makeCodexFIFOFixture(filepath.Join(root, ".moai", "manifest.json")); err != nil {
				if errors.Is(err, errCodexFixtureUnsupported) {
					t.Skip("named pipes cannot be created on this platform")
				}
				t.Fatalf("plant a named pipe where the manifest belongs: %v", err)
			}
		}, "not a file"},
		{"oversized_json", manifest, func(t *testing.T, root string) {
			// Valid JSON one byte past the read bound: the probe must refuse to read it.
			writeProbeFile(t, root, ".moai/manifest.json", `{"pad":"`+strings.Repeat("x", plan.MaxConfigSize)+`"}`)
		}, "unreadable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.setup(t, root)
			got, returned := damageReasonWithin(tc.entry, root, probeReturnBound)
			if !returned {
				unblockPipe(filepath.Join(root, filepath.FromSlash(tc.entry.rel)))
				t.Fatalf("damageReason(%s) did not return within %s: the probe is blocked on a read it must never make",
					tc.entry.rel, probeReturnBound)
			}
			if got != tc.want {
				t.Fatalf("damageReason(%s) = %q, want %q", tc.entry.rel, got, tc.want)
			}
		})
	}
}

// TestReadProbeFile_HandleRecheckRefusesNonRegular pins the checked-handle read
// (F1): the handle opened for a member is re-checked with Stat, so a named pipe is
// refused as "not a file" even when it is reached past the pre-open stat, and the
// refusal returns without a blocking read.
func TestReadProbeFile_HandleRecheckRefusesNonRegular(t *testing.T) {
	pipe := filepath.Join(t.TempDir(), "member.fifo")
	if err := makeCodexFIFOFixture(pipe); err != nil {
		if errors.Is(err, errCodexFixtureUnsupported) {
			t.Skip("named pipes cannot be created on this platform")
		}
		t.Fatalf("create a named pipe: %v", err)
	}
	done := make(chan string, 1)
	go func() {
		_, reason := readProbeFile(pipe)
		done <- reason
	}()
	select {
	case reason := <-done:
		if reason != "not a file" {
			t.Fatalf("readProbeFile(named pipe) = %q, want %q", reason, "not a file")
		}
	case <-time.After(probeReturnBound):
		unblockPipe(pipe)
		t.Fatalf("readProbeFile did not return within %s on a named pipe", probeReturnBound)
	}
}

// TestIntegrityProbeSet_EveryMemberObservableOnVersionMatchedPath pins F3. The
// probe runs only on the version-matched path, and that path is taken only while
// the template-version stamp in system.yaml matches the package version. So every
// member of the set must stay observable with the stamp intact: damaging a member
// must not flip the skip predicate, or the probe never runs and never names it.
func TestIntegrityProbeSet_EveryMemberObservableOnVersionMatchedPath(t *testing.T) {
	for _, e := range managedSurfaceProbeSet {
		t.Run(e.rel, func(t *testing.T) {
			root := t.TempDir()
			writeVersionStampedProject(t, root)
			if !versionMatchSkips(t, root) {
				t.Fatal("precondition: an intact stamped project must take the version-match skip")
			}
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(e.rel))); err != nil {
				t.Fatalf("damage fixture: %v", err)
			}
			if !versionMatchSkips(t, root) {
				t.Fatalf("F3: removing %s flips the version-match predicate, so the probe never runs on this project and cannot name the member", e.rel)
			}
			var out bytes.Buffer
			runManagedSurfaceIntegrityProbe(&out, root)
			if !strings.Contains(out.String(), e.rel) {
				t.Fatalf("F3: the version-matched probe did not name the removed member %s:\n%s", e.rel, out.String())
			}
		})
	}
}

// writeVersionStampedProject lays down an intact project whose system.yaml carries
// the package's template-version stamp, so the version-match predicate holds.
func writeVersionStampedProject(t *testing.T, root string) {
	t.Helper()
	writeProbeFile(t, root, ".moai/config/sections/system.yaml",
		fmt.Sprintf("moai:\n  template_version: %s\n", version.GetVersion()))
	writeProbeFile(t, root, ".claude/settings.json", "{}")
	writeProbeFile(t, root, ".moai/manifest.json", "{}")
}

// versionMatchSkips evaluates the real version-match predicate with the update
// command's force flag pinned off, so state a sibling test left on the shared
// command cannot change the answer.
func versionMatchSkips(t *testing.T, root string) bool {
	t.Helper()
	prev := updateCmd.Flags().Lookup("force").Value.String()
	t.Cleanup(func() { _ = updateCmd.Flags().Set("force", prev) })
	if err := updateCmd.Flags().Set("force", "false"); err != nil {
		t.Fatalf("set --force: %v", err)
	}
	return updateSkippedOnVersionMatch(updateCmd, root)
}
