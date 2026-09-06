package cli

// integration_settings_drift_report_test.go — the human-facing report surface.
//
// The lead reads this text, so its three states are pinned as text. Each
// assertion is on rendered content; none is on a process exit code.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestSettingsDriftReportNamesTheTreeInEveryState(t *testing.T) {
	t.Parallel()

	clean := settingsDriftReportText(kanban.SettingsDriftResult{
		Status: kanban.SettingsDriftClean, MatchCount: 0, Worktree: "/tmp/lane-tree",
	})
	if !strings.Contains(clean, "/tmp/lane-tree") {
		t.Errorf("clean report does not name the measured tree: %q", clean)
	}
	if !strings.Contains(clean, "clean") {
		t.Errorf("clean report does not state the verdict: %q", clean)
	}

	// `undetermined` must not be able to read as a pass to someone skimming.
	undet := settingsDriftReportText(kanban.SettingsDriftResult{
		Status:     kanban.SettingsDriftUndetermined,
		MatchCount: kanban.SettingsDriftMatchCountUnmeasured,
		Worktree:   "/tmp/lane-tree",
		Err:        errors.New("not a git repository"),
	})
	if !strings.Contains(undet, "UNDETERMINED") {
		t.Errorf("undetermined report does not name the state: %q", undet)
	}
	if !strings.Contains(undet, "not a pass") {
		t.Errorf("undetermined report does not say it is not a pass: %q", undet)
	}
	if !strings.Contains(undet, "not a git repository") {
		t.Errorf("undetermined report drops the reason: %q", undet)
	}
	// A match count under undetermined would be read as a verdict.
	if strings.Contains(undet, "0 matches") {
		t.Errorf("undetermined report reports a match count: %q", undet)
	}

	drift := settingsDriftReportText(kanban.SettingsDriftResult{
		Status:        kanban.SettingsDriftDetected,
		MatchCount:    1,
		Worktree:      "/tmp/lane-tree",
		Path:          "/tmp/lane-tree/.claude/settings.json",
		SHA256:        "abc123",
		SizeBytes:     42,
		PreservedPath: "/primary/.moai/state/settings-drift/settings.json.t488.x.abc123",
		Bypassed:      true,
		PreserveErr:   errors.New("disk full"),
	})
	for _, want := range []string{
		"DRIFT", "/tmp/lane-tree/.claude/settings.json", "abc123",
		"settings.json.t488.x.abc123", "bypassed", "preserve failed", "disk full",
		"report this to the lead", "Nothing was restored",
	} {
		if !strings.Contains(drift, want) {
			t.Errorf("drift report is missing %q:\n%s", want, drift)
		}
	}
	// The file's contents must never reach the report — it can hold secrets.
	if strings.Contains(drift, "{") {
		t.Errorf("drift report appears to carry file contents:\n%s", drift)
	}
}

// TestWriteSettingsDriftReportSurfacesAWriteFailure pins that the single write
// is checked rather than dropped.
func TestWriteSettingsDriftReportSurfacesAWriteFailure(t *testing.T) {
	t.Parallel()
	err := writeSettingsDriftReport(failingWriter{}, kanban.SettingsDriftResult{
		Status: kanban.SettingsDriftClean, Worktree: "/tmp/lane-tree",
	})
	if err == nil {
		t.Fatal("a failed report write was swallowed")
	}
	if !strings.Contains(err.Error(), "settings-drift report") {
		t.Errorf("error does not name what failed: %v", err)
	}
}

// An empty result renders nothing and writes nothing, so a zero-valued struct
// cannot produce a blank line that reads as a verdict.
func TestWriteSettingsDriftReportIsSilentForAnUnsetResult(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := writeSettingsDriftReport(&buf, kanban.SettingsDriftResult{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("an unset result rendered %q", buf.String())
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed pipe") }

// TestPreflightHumanReportRendersOnTheNonJSONPath covers the default surface a
// human actually invokes.
func TestPreflightHumanReportRendersOnTheNonJSONPath(t *testing.T) {
	clean := newDriftWorktree(t)
	root := t.TempDir()

	out, err := runIntegration(t, root, "preflight", clean)
	if err != nil {
		t.Fatalf("preflight on a clean tree errored: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "settings drift: clean") {
		t.Errorf("clean report missing from stdout: %q", out)
	}

	dirty := dirtyDriftWorktree(t)
	root2 := t.TempDir()
	out2, err2 := runIntegration(t, root2, "preflight", "--card", "t488", dirty)
	if err2 == nil {
		t.Errorf("preflight returned no signal on a hit; the exit code is a convenience, but it should be non-zero")
	}
	if !strings.Contains(out2, "settings drift: DRIFT") {
		t.Errorf("drift report missing from stdout: %q", out2)
	}
	if !strings.Contains(out2, "preserved:") {
		t.Errorf("drift report does not name the preserved copy: %q", out2)
	}
}

// TestPreflightDefaultsToTheCurrentTree covers the no-argument path, which is
// how a lane standing in its own worktree would call it.
func TestPreflightDefaultsToTheCurrentTree(t *testing.T) {
	worktree := dirtyDriftWorktree(t)
	root := t.TempDir()

	out, err := runIntegrationIn(t, worktree, root, "preflight", "--json")
	if err == nil {
		t.Errorf("preflight on a dirty tree returned no signal\noutput: %s", out)
	}
	if !strings.Contains(out, `"status":"drift"`) {
		t.Errorf("preflight did not measure the current tree: %q", out)
	}
	if !strings.Contains(out, worktree) {
		t.Errorf("report does not name the tree it measured (%s): %q", worktree, out)
	}
}
