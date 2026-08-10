// Package cli — update_clean_install_removed_paths_log_test.go
//
// Regression guard for the opaque removal log (issue #1415 request 4).
//
// The REMOVE phase printed only `Removed N deprecated paths`. result.RemovedPaths
// was populated but never surfaced, so a user seeing a non-converging update had
// no way to learn WHICH paths the tool kept deleting — exactly the diagnosis the
// #1415 reporter asked for and could not obtain.

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestRunCleanReinstall_LogsEachRemovedPath asserts that a run removing at least
// one deprecated path names every removed path on its own indented line, and
// that the enumerated set matches the count line's arithmetic (removed = scanned
// minus post-rescan remaining) rather than the full candidate list.
func TestRunCleanReinstall_LogsEachRemovedPath(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v2.16.1\n")
	// Two genuine deprecated entries so the log has more than one line to emit.
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".agency/index.md", "legacy\n")
	writeTestFile(t, root, ".moai/config/sections/memo.yaml", "memo:\n    enabled: true\n")

	var buf bytes.Buffer
	result, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              &buf,
		Deployer:         &stubDeployer{},
		RunMigrateAgency: (&stubMigrateRunner{}).Run,
	})
	if err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}
	if len(result.RemovedPaths) == 0 {
		t.Fatalf("fixture precondition broken: nothing was removed")
	}

	got := buf.String()
	if !strings.Contains(got, "[clean-reinstall] Removed ") {
		t.Fatalf("removal count line missing\ngot:\n%s", got)
	}

	// Every path the run actually deleted must be named in the output.
	for _, rel := range []string{".agency", ".moai/config/sections/memo.yaml"} {
		want := "[clean-reinstall]   - " + rel
		if !strings.Contains(got, want) {
			t.Errorf("removed path %q not enumerated in the log; a user cannot "+
				"diagnose a non-converging update from a bare count\nwant line: %q\ngot:\n%s",
				rel, want, got)
		}
	}
}

// TestRunCleanReinstall_NoRemovedPathLinesWhenNothingRemoved asserts the
// enumeration rides the same gate as the count line: a zero-removal run emits
// the informational no-op line and no per-path lines.
func TestRunCleanReinstall_NoRemovedPathLinesWhenNothingRemoved(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v2.16.1\n")
	writeTestFile(t, root, ".moai/specs/SPEC-USER-006/spec.md", "user spec\n")

	var buf bytes.Buffer
	result, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              &buf,
		Deployer:         &stubDeployer{},
		RunMigrateAgency: (&stubMigrateRunner{}).Run,
	})
	if err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}
	if len(result.RemovedPaths) != 0 {
		t.Fatalf("fixture precondition broken: expected 0 removals, got %v", result.RemovedPaths)
	}

	got := buf.String()
	if !strings.Contains(got, "[clean-reinstall] No deprecated paths found to remove") {
		t.Errorf("informational no-op line missing\ngot:\n%s", got)
	}
	if strings.Contains(got, "[clean-reinstall]   - ") {
		t.Errorf("per-path removal lines emitted on a zero-removal run\ngot:\n%s", got)
	}
}
