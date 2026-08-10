package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// bashPostToolPayload is the VERBATIM shape Claude Code 2.1.226 delivers to a
// matcher-null PostToolUse hook for a Bash call, captured by the M0 probe
// (plan.md §F M0). It is reproduced here rather than invented so that a runtime
// wire-format change shows up as a failing test rather than as a silently empty
// telemetry file.
//
// Two properties of the real payload are load-bearing and easy to get wrong:
//   - it arrives FLAT snake_case, not nested camelCase;
//   - tool_response carries stdout/stderr but NO exit_code field, so pass/fail
//     detection falls through to the output-text heuristic.
func bashPostToolPayload(command, stdout string) string {
	payload := map[string]any{
		"session_id":      "sess-bash-1",
		"cwd":             "/tmp/probe",
		"permission_mode": "acceptEdits",
		"hook_event_name": "PostToolUse",
		"tool_name":       "Bash",
		"tool_input": map[string]any{
			"command":     command,
			"description": "run the test suite",
		},
		"tool_response": map[string]any{
			"stdout":           stdout,
			"stderr":           "",
			"interrupted":      false,
			"isImage":          false,
			"noOutputExpected": false,
		},
		"tool_use_id": "toolu_01probe",
		"duration_ms": 462,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// readTelemetryRecords returns every usage record written under root today.
func readTelemetryRecords(t *testing.T, root string) []telemetry.UsageRecord {
	t.Helper()
	day := time.Now().UTC().Format("2006-01-02")
	path := filepath.Join(root, ".moai", "evolution", "telemetry", "usage-"+day+".jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read telemetry: %v", err)
	}
	var out []telemetry.UsageRecord
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec telemetry.UsageRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("parse telemetry line %q: %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

// TestHarnessObserve_BashEvidenceRecorded is the AC-HLE-013 positive half
// (REQ-HLE-011): the Bash evidence-record path, measured as unreachable from the
// shipped hook wiring, now executes on the matcher-null observe channel.
//
// The handle-post-tool.sh wrapper is registered for Write|Edit|MultiEdit only, so
// its Bash branch never runs in production. The observe wrapper IS registered for
// every tool; this test drives that handler with the real Bash payload and
// asserts a telemetry record with is_test_pass actually lands.
func TestHarnessObserve_BashEvidenceRecorded(t *testing.T) {
	root := t.TempDir()
	writeHarnessYAML(t, root, "learning:\n  enabled: true\n")
	writeConfig(t, root, "system.yaml", "hook:\n  opt_in:\n    enabled: true\n")
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Chdir(root)

	cmd := &cobra.Command{}
	withStdin(t, bashPostToolPayload("go test ./...", "ok  \tgithub.com/x/y\t0.42s\n"), func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve returned error: %v", err)
		}
	})

	recs := readTelemetryRecords(t, root)
	if len(recs) != 1 {
		t.Fatalf("got %d telemetry records, want exactly 1 (the Bash evidence record)", len(recs))
	}
	if !recs[0].IsTestPass {
		t.Errorf("is_test_pass = false; the observed passing test result must set it (record: %+v)", recs[0])
	}
	if recs[0].IsTestFail {
		t.Errorf("is_test_fail must not be set on a passing run (record: %+v)", recs[0])
	}
	if recs[0].SessionID != "sess-bash-1" {
		t.Errorf("session_id = %q, want sess-bash-1 (session correlation is what the Stop seam reads)", recs[0].SessionID)
	}
}

// TestHarnessObserve_BashEvidenceTestFail pins the failing-run direction. The
// real payload carries no exit_code (M0 probe), so this exercises the
// output-text fallback rather than the structured exit-code path.
func TestHarnessObserve_BashEvidenceTestFail(t *testing.T) {
	root := t.TempDir()
	writeHarnessYAML(t, root, "learning:\n  enabled: true\n")
	writeConfig(t, root, "system.yaml", "hook:\n  opt_in:\n    enabled: true\n")
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Chdir(root)

	cmd := &cobra.Command{}
	withStdin(t, bashPostToolPayload("go test ./...", "--- FAIL: TestX (0.01s)\nFAIL\n"), func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve returned error: %v", err)
		}
	})

	recs := readTelemetryRecords(t, root)
	if len(recs) != 1 {
		t.Fatalf("got %d telemetry records, want exactly 1", len(recs))
	}
	if !recs[0].IsTestFail || recs[0].IsTestPass {
		t.Errorf("want is_test_fail=true is_test_pass=false, got %+v", recs[0])
	}
}

// TestHarnessObserve_NonTestBashWritesNoEvidence keeps the write-volume
// discipline visible: a non-test Bash command produces no evidence record, so
// the carve-out cannot quietly turn every shell call into a telemetry line.
func TestHarnessObserve_NonTestBashWritesNoEvidence(t *testing.T) {
	root := t.TempDir()
	writeHarnessYAML(t, root, "learning:\n  enabled: true\n")
	writeConfig(t, root, "system.yaml", "hook:\n  opt_in:\n    enabled: true\n")
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Chdir(root)

	cmd := &cobra.Command{}
	withStdin(t, bashPostToolPayload("ls -la", "total 0\n"), func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve returned error: %v", err)
		}
	})

	if recs := readTelemetryRecords(t, root); len(recs) != 0 {
		t.Fatalf("non-test Bash must write no evidence record, got %d: %+v", len(recs), recs)
	}
}

// TestHarnessObserve_NonBashToolWritesNoEvidence is the AC-HLE-013 no-double-write
// half as seen from the observe channel: Edit and Write are already owned by
// handle-post-tool.sh, so the observe handler must NOT also record evidence for
// them. Without this scoping every Edit would produce two telemetry records
// (plan.md §G AP-8).
func TestHarnessObserve_NonBashToolWritesNoEvidence(t *testing.T) {
	root := t.TempDir()
	writeHarnessYAML(t, root, "learning:\n  enabled: true\n")
	writeConfig(t, root, "system.yaml", "hook:\n  opt_in:\n    enabled: true\n")
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Chdir(root)

	payload := `{"session_id":"sess-edit","hook_event_name":"PostToolUse","tool_name":"Edit",` +
		`"tool_input":{"file_path":"/tmp/probe/main.go"},"tool_response":{"stdout":""}}`

	cmd := &cobra.Command{}
	withStdin(t, payload, func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve returned error: %v", err)
		}
	})

	if recs := readTelemetryRecords(t, root); len(recs) != 0 {
		t.Fatalf("observe channel must not record evidence for Edit (handle-post-tool.sh owns it), got %d: %+v", len(recs), recs)
	}
}

// TestHarnessObserve_BashEvidenceGated pins REQ-HLE-013 for the carve-out: with
// either observation gate closed, no telemetry file is written at all.
func TestHarnessObserve_BashEvidenceGated(t *testing.T) {
	cases := []struct {
		name       string
		systemYAML string
		harnessYML string
	}{
		{
			name:       "hook opt-in closed",
			systemYAML: "hook:\n  opt_in:\n    enabled: false\n",
			harnessYML: "learning:\n  enabled: true\n",
		},
		{
			name:       "hook opt-in absent (fail-closed default)",
			systemYAML: "other: 1\n",
			harnessYML: "learning:\n  enabled: true\n",
		},
		{
			name:       "learning closed",
			systemYAML: "hook:\n  opt_in:\n    enabled: true\n",
			harnessYML: "learning:\n  enabled: false\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeHarnessYAML(t, root, tc.harnessYML)
			writeConfig(t, root, "system.yaml", tc.systemYAML)
			t.Setenv(config.EnvClaudeProjectDir, root)
			t.Chdir(root)

			cmd := &cobra.Command{}
			withStdin(t, bashPostToolPayload("go test ./...", "ok  \tpkg\t0.1s\n"), func() {
				if err := runHarnessObserve(cmd, nil); err != nil {
					t.Fatalf("runHarnessObserve returned error: %v", err)
				}
			})

			if recs := readTelemetryRecords(t, root); len(recs) != 0 {
				t.Fatalf("a closed gate must write no evidence record, got %d: %+v", len(recs), recs)
			}
		})
	}
}
