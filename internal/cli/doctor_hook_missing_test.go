package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// writeHookMissingLog writes the given lines to the project's
// .moai/logs/hook-missing.log so tests never touch the real project log.
func writeHookMissingLog(t *testing.T, projectRoot string, lines ...string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(dir, "hook-missing.log"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckHookMissingLog_AbsentIsOK(t *testing.T) {
	root := t.TempDir()

	check := checkHookMissingLog(root, false)

	if check.Name != hookMissingLogCheckName {
		t.Errorf("Name = %q, want %q", check.Name, hookMissingLogCheckName)
	}
	if check.Status != uikit.CheckOK {
		t.Errorf("Status = %q, want ok; msg=%s", check.Status, check.Message)
	}
}

func TestCheckHookMissingLog_PresentWithTimestampsWarns(t *testing.T) {
	root := t.TempDir()
	writeHookMissingLog(t, root,
		"2026-09-08T06:06:01Z hook missing: /x/.claude/hooks/moai/handle-config-change.sh",
		"2026-09-08T06:06:02Z hook missing: /x/.claude/hooks/moai/handle-config-change.sh",
	)

	check := checkHookMissingLog(root, false)

	if check.Status != uikit.CheckWarn {
		t.Fatalf("Status = %q, want warn; msg=%s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "2") {
		t.Errorf("Message = %q, want the entry count 2", check.Message)
	}
	if !strings.Contains(check.Message, "handle-config-change.sh") {
		t.Errorf("Message = %q, want the latest entry's script name", check.Message)
	}
}

func TestCheckHookMissingLog_TimestamplessEntriesCounted(t *testing.T) {
	root := t.TempDir()
	writeHookMissingLog(t, root,
		" hook missing: /x/.claude/hooks/moai/handle-config-change.sh",
	)

	check := checkHookMissingLog(root, false)

	if check.Status != uikit.CheckWarn {
		t.Fatalf("Status = %q, want warn; msg=%s", check.Status, check.Message)
	}
	if !strings.Contains(check.Message, "1") {
		t.Errorf("Message = %q, want the entry count 1", check.Message)
	}
	if !strings.Contains(check.Detail, "1") || !strings.Contains(check.Detail, "timestamp") {
		t.Errorf("Detail = %q, want the timestampless-entry count and the word timestamp", check.Detail)
	}
}

func TestCheckHookMissingLog_Registered(t *testing.T) {
	results := runDiagnosticChecks(false, hookMissingLogCheckName)
	if len(results) != 1 || results[0].Name != hookMissingLogCheckName {
		t.Fatalf("expected one %q result, got %+v", hookMissingLogCheckName, results)
	}
}
