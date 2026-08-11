package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEpicStatusCmd_JSONOutput exercises the CLI end-to-end via a fixture:
// `moai epic status NAVIGATOR-SYNC --json` produces valid JSON with the
// orphan_mx = ["M2","M3","M5"] ground truth.
func TestEpicStatusCmd_JSONOutput(t *testing.T) {
	root := t.TempDir()
	writeEpicFixture(t, root, "SPEC-NAVIGATOR-SYNC-001", "Navigator Sync (BAS M0) — graph", "completed")
	writeEpicFixture(t, root, "SPEC-NAVIGATOR-SYNC-002", "Navigator Sync (BAS M4) — 4-tier", "completed")
	writeEpicFixture(t, root, "SPEC-NAVIGATOR-SYNC-003", "Navigator Sync (BAS M1) — detect", "in-progress")
	writeEpicFixtureRaw(t, root, ".moai/reports/navigator-redesign-bas-20260805.html", basReportFixtureHTML())
	writeEpicFixtureRaw(t, root, ".moai/specs/SPEC-NAVIGATOR-SYNC-001/progress.md",
		"## §E.4 Sync-phase Audit-Ready Signal\nsync_commit_sha: \"abc123\"\n")
	writeEpicFixtureRaw(t, root, ".moai/specs/SPEC-NAVIGATOR-SYNC-002/progress.md",
		"## §E.4 Sync-phase Audit-Ready Signal\nsync_commit_sha: \"def456\"\n")

	cmd := newEpicStatusCmd()
	cmd.SetArgs([]string{"NAVIGATOR-SYNC", "--json", "--base-dir", root})
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("cmd.Execute: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	if parsed["epic"] != "NAVIGATOR-SYNC" {
		t.Errorf("epic = %v, want NAVIGATOR-SYNC", parsed["epic"])
	}
	orphan, ok := parsed["orphan_mx"].([]any)
	if !ok {
		t.Fatalf("orphan_mx not an array: %T", parsed["orphan_mx"])
	}
	if len(orphan) != 3 {
		t.Fatalf("orphan_mx len = %d, want 3: %v", len(orphan), orphan)
	}
	want := map[string]bool{"M2": true, "M3": true, "M5": true}
	for _, o := range orphan {
		if !want[o.(string)] {
			t.Errorf("unexpected orphan %v", o)
		}
	}
}

// TestEpicStatusCmd_HumanOutput exercises the non-JSON path.
func TestEpicStatusCmd_HumanOutput(t *testing.T) {
	root := t.TempDir()
	writeEpicFixture(t, root, "SPEC-X-001", "(EPICX M0) foo", "completed")
	cmd := newEpicStatusCmd()
	cmd.SetArgs([]string{"X", "--base-dir", root})
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("cmd.Execute: %v", err)
	}
	if !strings.Contains(out.String(), "🎯") {
		t.Errorf("human output lacks 🎯:\n%s", out.String())
	}
}

// TestEpicCmd_HelpListsStatus verifies AC-ES-010: `moai epic --help` lists
// `status` as a subcommand.
func TestEpicCmd_HelpListsStatus(t *testing.T) {
	cmd := newEpicCmd()
	cmd.SetArgs([]string{"--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	// --help returns an error from cobra (pflag.ErrHelp); ignore it.
	_ = cmd.Execute()
	if !strings.Contains(out.String(), "status") {
		t.Errorf("epic --help does not list status:\n%s", out.String())
	}
}

// TestEpicCmd_StatusHelpFlags verifies AC-ES-010: `moai epic status --help`
// lists the --json / --design-report / --marker flags.
func TestEpicCmd_StatusHelpFlags(t *testing.T) {
	cmd := newEpicStatusCmd()
	cmd.SetArgs([]string{"--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	_ = cmd.Execute()
	s := out.String()
	for _, flag := range []string{"--json", "--design-report", "--marker"} {
		if !strings.Contains(s, flag) {
			t.Errorf("status --help lacks %s:\n%s", flag, s)
		}
	}
}

// TestNewEpicCmd_NoAskUserQuestion is the C-HRA-008 static guard for the epic
// CLI surface — mirrors worktree/new_test.go's canonical pattern. The CLI is
// non-interactive (REQ-ES-010).
func TestNewEpicCmd_NoAskUserQuestion(t *testing.T) {
	data, err := os.ReadFile("epic.go")
	if err != nil {
		t.Fatalf("read epic.go: %v", err)
	}
	body := string(data)
	for _, forbidden := range []string{"AskUserQuestion", "mcp__askuser", "fmt.Scanln", "bufio.NewReader"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("epic.go contains forbidden interactive primitive %q (C-HRA-008)", forbidden)
		}
	}
}

// TestEpicStatusCmd_NoPersistedStore verifies AC-ES-013: the producer creates
// ZERO new files under .moai/. We snapshot the file count before and after.
func TestEpicStatusCmd_NoPersistedStore(t *testing.T) {
	root := t.TempDir()
	writeEpicFixture(t, root, "SPEC-X-001", "(T M0) foo", "completed")
	before := countAllFiles(t, root)

	cmd := newEpicStatusCmd()
	cmd.SetArgs([]string{"X", "--json", "--base-dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("cmd.Execute: %v", err)
	}
	after := countAllFiles(t, root)
	if after != before {
		t.Errorf("producer created files: before=%d after=%d (REQ-ES-013 violated)", before, after)
	}
}

// --- test helpers ---

func writeEpicFixture(t *testing.T, root, id, title, status string) {
	t.Helper()
	content := "---\n" +
		"id: " + id + "\n" +
		"title: \"" + title + "\"\n" +
		"version: \"0.1.0\"\n" +
		"status: " + status + "\n" +
		"created: 2026-08-11\n" +
		"updated: 2026-08-11\n" +
		"author: test\n" +
		"priority: P1\n" +
		"phase: \"v3.2.0\"\n" +
		"module: internal/epic\n" +
		"lifecycle: spec-anchored\n" +
		"tags: test\n" +
		"---\n\n# " + id + "\n"
	writeEpicFixtureRaw(t, root, ".moai/specs/"+id+"/spec.md", content)
}

func writeEpicFixtureRaw(t *testing.T, root, relPath, content string) {
	t.Helper()
	full := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", full, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
}

// countAllFiles walks dir and returns the total file count (used by the
// no-persisted-store guard). Named countAllFiles to avoid colliding with the
// package's existing countFiles(dir, ext) helper in status.go.
func countAllFiles(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return n
}

func basReportFixtureHTML() string {
	return "<html><body><h2>7. slice</h2><table><tbody>" +
		"<tr><td>M0 graph</td><td>x</td><td>—</td><td>L</td></tr>" +
		"<tr><td>M1 detect</td><td>x</td><td>M0</td><td>M</td></tr>" +
		"<tr><td>M2 route</td><td>x</td><td>M0</td><td>M</td></tr>" +
		"<tr><td>M3 fix</td><td>x</td><td>M1+M2</td><td>L</td></tr>" +
		"<tr><td>M4 4-tier</td><td>x</td><td>M0</td><td>L</td></tr>" +
		"<tr><td>M5 brownfield</td><td>x</td><td>M4</td><td>M</td></tr>" +
		"</tbody></table></body></html>"
}
