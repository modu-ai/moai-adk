package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/hygiene"
)

// hygieneFixture builds a temp project carrying one deletion-eligible
// context-usage candidate: the registry entry's pid is a reaped dead
// process, its heartbeat is a month old, the transcript under the fixture
// CLAUDE_CONFIG_DIR is a month old, and the candidate body records a
// captured_at a month old — DEAD + content-datable + aged.
func hygieneFixture(t *testing.T) (project, candidate string, key string) {
	t.Helper()
	project = t.TempDir()
	moai := filepath.Join(project, ".moai")
	for _, dir := range []string{
		filepath.Join(moai, "state"),
		filepath.Join(moai, "state", "context-usage"),
		filepath.Join(moai, "config", "sections"),
		filepath.Join(project, "cfg", "projects"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	// A dead pid: start a trivial child and reap it.
	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn probe child: %v", err)
	}
	deadPID := cmd.Process.Pid
	_, _ = cmd.Process.Wait()

	key = "22222222-aaaa-bbbb-cccc-000000000001"
	old := time.Now().AddDate(0, 0, -30).UTC()
	registry := []map[string]any{{
		"session_id":     key,
		"pid":            deadPID,
		"last_heartbeat": old.Format(time.RFC3339),
		"started_at":     old.Format(time.RFC3339),
	}}
	blob, err := json.Marshal(registry)
	if err != nil {
		t.Fatalf("marshal registry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moai, "state", "active-sessions.json"), blob, 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	body, err := json.Marshal(map[string]any{
		"captured_at": old.Format(time.RFC3339),
		"raw_pct":     0.42,
	})
	if err != nil {
		t.Fatalf("marshal candidate: %v", err)
	}
	candidate = filepath.Join(moai, "state", "context-usage", key+".json")
	if err := os.WriteFile(candidate, body, 0o644); err != nil {
		t.Fatalf("write candidate: %v", err)
	}

	transcript := filepath.Join(project, "cfg", "projects", key+".jsonl")
	if err := os.WriteFile(transcript, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}
	stale := time.Now().AddDate(0, 0, -30)
	if err := os.Chtimes(transcript, stale, stale); err != nil {
		t.Fatalf("chtimes transcript: %v", err)
	}

	// Vouch the fixture roots for the REQ-HYG-015 runtime guard: under a
	// test binary the units' entry points refuse any root no test
	// registered.
	hygiene.RegisterTestRoot(moai)
	hygiene.RegisterTestRoot(filepath.Join(moai, "logs"))

	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(project, "cfg"))
	return project, candidate, key
}

// writeHygieneConfig writes a workflow.yaml hygiene block into the fixture.
func writeHygieneConfig(t *testing.T, project, block string) {
	t.Helper()
	path := filepath.Join(project, ".moai", "config", "sections", "workflow.yaml")
	if err := os.WriteFile(path, []byte("workflow:\n  hygiene:\n"+block), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// runCleanHygieneQuiet invokes the hygiene scope with the working directory
// inside the fixture project and returns the printer output.
func runCleanHygieneQuiet(t *testing.T, scope string, apply bool) string {
	t.Helper()
	var out strings.Builder
	p := printer.New(printer.WithWriters(&out, &out))
	var err error
	if scope == "audit-logs" {
		err = runCleanHygiene(p, hygieneScopeAuditLogs, apply)
	} else {
		err = runCleanHygiene(p, hygieneScopeSessionState, apply)
	}
	if err != nil {
		t.Fatalf("runCleanHygiene(%s, apply=%v): %v", scope, apply, err)
	}
	return out.String()
}

// TestCleanHygieneFlags — AC-HYG-014 (L-014): dry-run default; --apply
// mutates that invocation only; the config mode alone never mutates the
// CLI; the D30 config-validation arms refuse mutation.
func TestCleanHygieneFlags(t *testing.T) {
	t.Run("dry-run default keeps the residue and prints decisions", func(t *testing.T) {
		project, candidate, _ := hygieneFixture(t)
		t.Chdir(project)
		out := runCleanHygieneQuiet(t, "session-state", false)
		if !strings.Contains(out, "[dry-run]") {
			t.Fatalf("dry-run marker missing: %s", out)
		}
		if _, err := os.Stat(candidate); err != nil {
			t.Fatalf("dry-run deleted the residue: %v", err)
		}
	})

	t.Run("config mode apply without --apply never mutates the CLI", func(t *testing.T) {
		project, candidate, _ := hygieneFixture(t)
		writeHygieneConfig(t, project, "    mode: apply\n")
		t.Chdir(project)
		out := runCleanHygieneQuiet(t, "session-state", false)
		if !strings.Contains(out, "[dry-run]") {
			t.Fatalf("config apply leaked into the CLI: %s", out)
		}
		if _, err := os.Stat(candidate); err != nil {
			t.Fatalf("config mode alone mutated: %v", err)
		}
	})

	t.Run("apply mutates this invocation", func(t *testing.T) {
		project, candidate, _ := hygieneFixture(t)
		t.Chdir(project)
		out := runCleanHygieneQuiet(t, "session-state", true)
		if strings.Contains(out, "[dry-run]") {
			t.Fatalf("apply run printed the dry-run marker: %s", out)
		}
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			t.Fatalf("eligible residue survived --apply: %v", err)
		}
	})

	t.Run("kept_rotations other than 1 refuses the run (D30)", func(t *testing.T) {
		project, candidate, _ := hygieneFixture(t)
		writeHygieneConfig(t, project, "    mode: apply\n    audit_log_kept_rotations: 2\n")
		t.Chdir(project)
		var out strings.Builder
		p := printer.New(printer.WithWriters(&out, &out))
		err := runCleanHygiene(p, hygieneScopeSessionState, true)
		if err == nil || !strings.Contains(err.Error(), "config-invalid") {
			t.Fatalf("err = %v, want config-invalid", err)
		}
		if _, statErr := os.Stat(candidate); statErr != nil {
			t.Fatalf("config-invalid run mutated anyway: %v", statErr)
		}
	})

	t.Run("non-positive floors refuse the run (D30)", func(t *testing.T) {
		project, candidate, _ := hygieneFixture(t)
		writeHygieneConfig(t, project, "    mode: apply\n    min_age_days: 0\n")
		t.Chdir(project)
		var out strings.Builder
		p := printer.New(printer.WithWriters(&out, &out))
		err := runCleanHygiene(p, hygieneScopeSessionState, true)
		if err == nil || !strings.Contains(err.Error(), "config-invalid") {
			t.Fatalf("err = %v, want config-invalid", err)
		}
		if _, statErr := os.Stat(candidate); statErr != nil {
			t.Fatalf("config-invalid run mutated anyway: %v", statErr)
		}
	})

	t.Run("unknown mode falls back to report (D30)", func(t *testing.T) {
		project, candidate, _ := hygieneFixture(t)
		writeHygieneConfig(t, project, "    mode: banana\n")
		t.Chdir(project)
		out := runCleanHygieneQuiet(t, "session-state", false)
		if !strings.Contains(out, "[dry-run]") {
			t.Fatalf("unknown mode did not fall back to report: %s", out)
		}
		if _, err := os.Stat(candidate); err != nil {
			t.Fatalf("unknown-mode run mutated: %v", err)
		}
	})

	t.Run("audit-logs scope runs the rotator dry by default", func(t *testing.T) {
		project, _, _ := hygieneFixture(t)
		logDir := filepath.Join(project, ".moai", "logs")
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			t.Fatalf("mkdir logs: %v", err)
		}
		sink := filepath.Join(logDir, "rule-load-audit.jsonl")
		oversized := strings.Repeat("2026-10-05T00:00:00Z audit row\n", 400000) // > 10 MiB
		if err := os.WriteFile(sink, []byte(oversized), 0o644); err != nil {
			t.Fatalf("seed sink: %v", err)
		}
		t.Chdir(project)
		out := runCleanHygieneQuiet(t, "audit-logs", false)
		if !strings.Contains(out, "[dry-run]") || !strings.Contains(out, "skipped-report-mode") {
			t.Fatalf("rotator dry-run output unexpected: %s", out)
		}
		info, err := os.Stat(sink)
		if err != nil || info.Size() == 0 {
			t.Fatalf("sink vanished in dry-run: %v", err)
		}
		if info.Size() < int64(10*1024*1024) {
			t.Fatalf("sink mutated in dry-run: %d bytes", info.Size())
		}
	})
}
