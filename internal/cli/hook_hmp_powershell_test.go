package cli

// SPEC-HOOK-MATCHER-POWERSHELL-001 (card t1224) — CLI-level PowerShell parity.
//
//   - Site 10: the matcher-null harness-observe channel routes a PowerShell
//     test run to the evidence path exactly as it routes a Bash one.
//   - AC-HMP-008 (second half): an unparseable pre-tool stdin keeps the
//     SPEC-HOOK-STDIN-FAILCLOSED-001 deny whichever tool name its visible
//     prefix carries — the dispatcher rejects stdin before reading a tool name.
//
// Kept in its own file: a concurrent card edits hook_test.go.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

// hmpIsolateHome points MOAI_HOME at a per-test directory (REQ-HMP-014).
func hmpIsolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
}

// hmpShellPostToolPayload is bashPostToolPayload with the tool name swapped.
func hmpShellPostToolPayload(t *testing.T, tool, command, stdout string) string {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(bashPostToolPayload(command, stdout)), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	payload["tool_name"] = tool
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}
	return string(data)
}

// TestHMPHarnessObservePowerShellEvidence pins site 10 against its Bash control.
func TestHMPHarnessObservePowerShellEvidence(t *testing.T) {
	hmpIsolateHome(t)
	for _, tool := range []string{"Bash", "PowerShell"} {
		t.Run(tool, func(t *testing.T) {
			root := t.TempDir()
			writeHarnessYAML(t, root, "learning:\n  enabled: true\n")
			writeConfig(t, root, "system.yaml", "hook:\n  opt_in:\n    enabled: true\n")
			t.Setenv(config.EnvClaudeProjectDir, root)
			t.Chdir(root)

			withStdin(t, hmpShellPostToolPayload(t, tool, "go test ./...", "ok  \tgithub.com/x/y\t0.42s\n"), func() {
				if err := runHarnessObserve(&cobra.Command{}, nil); err != nil {
					t.Fatalf("runHarnessObserve: %v", err)
				}
			})
			recs := readTelemetryRecords(t, root)
			if len(recs) != 1 || !recs[0].IsTestPass {
				t.Fatalf("%s: got %d records (%+v), want 1 passing evidence record", tool, len(recs), recs)
			}
		})
	}
}

// TestHMPTruncatedStdinFailClosed pins the dispatcher half of AC-HMP-008.
func TestHMPTruncatedStdinFailClosed(t *testing.T) {
	hmpIsolateHome(t)
	stdout := map[string]string{}
	for _, tool := range []string{"Bash", "PowerShell"} {
		truncated := []byte(`{"hook_event_name":"PreToolUse","tool_name":"` + tool + `","tool_input":{"command":"git status",`)
		r := runHookWithStdin(t, "pre-tool", nil, "", truncated)
		reason, ok := denyReason("PreToolUse", r.stdout)
		if !ok || !strings.Contains(reason, "fail-closed") {
			t.Fatalf("%s: truncated stdin did not fail closed: stdout %q", tool, r.stdout)
		}
		stdout[tool] = strings.TrimSpace(r.stdout)
	}
	if stdout["PowerShell"] != stdout["Bash"] {
		t.Errorf("fail-closed output differs by tool name: PowerShell %q, Bash %q", stdout["PowerShell"], stdout["Bash"])
	}
}
