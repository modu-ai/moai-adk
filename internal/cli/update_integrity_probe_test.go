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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/merge"
)

// integrityProbeRelPaths are the representative project paths the probe checks.
// They are written as literals on purpose: the tests assert on the rendered rows.
var integrityProbeRelPaths = []string{
	".claude/settings.json",
	".moai/config/sections/system.yaml",
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

// TestIntegrityProbeEntry_DamageReasons pins every classification one
// representative entry can return. The end-to-end tests reach only some of them:
// a version match needs a non-empty system.yaml that carries the stamp, so the
// empty case is observable only through the entry check itself.
func TestIntegrityProbeEntry_DamageReasons(t *testing.T) {
	settings := integrityProbeEntry{rel: ".claude/settings.json", check: probeJSON}
	system := integrityProbeEntry{rel: ".moai/config/sections/system.yaml", check: probeNonEmpty}

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
		{"empty_system_yaml", system, func(t *testing.T, root string) {
			writeProbeFile(t, root, ".moai/config/sections/system.yaml", "")
		}, "empty"},
		{"intact_system_yaml", system, func(t *testing.T, root string) {
			writeProbeFile(t, root, ".moai/config/sections/system.yaml", "moai:\n  template_version: v1\n")
		}, ""},
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
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.setup(t, root)
			if got := tc.entry.damageReason(root); got != tc.want {
				t.Fatalf("damageReason(%s) = %q, want %q", tc.entry.rel, got, tc.want)
			}
		})
	}
}
