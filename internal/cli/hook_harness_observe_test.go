// Package cli — hook harness-observe gate tests.
//
// REQ-HRN-FND-009 (SPEC-V3R4-HARNESS-001): While learning.enabled in
// harness.yaml resolves to false, the PostToolUse observer hook must be a
// complete no-op — no read, write, or append to usage-log.jsonl.
//
// REQ-HRN-FND-010: When learning.enabled is true, the PostToolUse observer
// must append one JSONL entry per tool invocation containing at minimum
// ISO-8601 timestamp, event_type, subject, and context hash.
package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

// writeHarnessYAML writes a harness.yaml document under dir/.moai/config/sections/
// containing the supplied YAML body. The directory tree is created if missing.
func writeHarnessYAML(t *testing.T, dir, body string) {
	t.Helper()
	configDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	path := filepath.Join(configDir, "harness.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write harness.yaml: %v", err)
	}
}

// withStdin temporarily replaces os.Stdin with the given reader for the
// duration of fn, restoring afterwards. Used to simulate hook stdin JSON.
func withStdin(t *testing.T, payload string, fn func()) {
	t.Helper()
	orig := os.Stdin
	t.Cleanup(func() { os.Stdin = orig })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdin = r
	go func() {
		_, _ = io.WriteString(w, payload)
		_ = w.Close()
	}()
	fn()
}

// ─────────────────────────────────────────────
// TestIsHarnessLearningEnabled — table-driven gate tests
// ─────────────────────────────────────────────

func TestIsHarnessLearningEnabled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		setupYAML string // empty = do not create harness.yaml
		want      bool
		reason    string
	}{
		{
			name:      "missing_harness_yaml_is_fail_open",
			setupYAML: "",
			want:      true,
			reason:    "fail-open: missing config preserves baseline observation per REQ-HRN-FND-009 semantics",
		},
		{
			name:      "empty_harness_yaml_is_fail_open",
			setupYAML: "\n",
			want:      true,
			reason:    "empty doc has no learning block; default to enabled",
		},
		{
			name:      "no_learning_block_is_enabled",
			setupYAML: "harness:\n  default_profile: \"default\"\n",
			want:      true,
			reason:    "harness section without learning key defaults to enabled",
		},
		{
			name:      "learning_block_without_enabled_key_is_enabled",
			setupYAML: "learning:\n  tier_thresholds: [1, 3, 5, 10]\n",
			want:      true,
			reason:    "learning block missing enabled key defaults to enabled",
		},
		{
			name:      "learning_enabled_true",
			setupYAML: "learning:\n  enabled: true\n",
			want:      true,
			reason:    "explicit true",
		},
		{
			name:      "learning_enabled_false_blocks_observation",
			setupYAML: "learning:\n  enabled: false\n",
			want:      false,
			reason:    "REQ-HRN-FND-009 — explicit false makes observer a no-op",
		},
		{
			name:      "invalid_yaml_is_fail_open",
			setupYAML: "learning:\n  enabled: : :: not yaml\n",
			want:      true,
			reason:    "parse error preserves baseline observation (fail-open)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if tc.setupYAML != "" {
				writeHarnessYAML(t, dir, tc.setupYAML)
			}

			got := isHarnessLearningEnabled(dir)
			if got != tc.want {
				t.Errorf("isHarnessLearningEnabled() = %v, want %v (%s)", got, tc.want, tc.reason)
			}
		})
	}
}

// ─────────────────────────────────────────────
// TestRunHarnessObserve_NoOpWhenLearningDisabled (REQ-HRN-FND-009)
// ─────────────────────────────────────────────

// Verifies that when learning.enabled is false, runHarnessObserve neither
// creates nor appends to .moai/harness/usage-log.jsonl and exits with no
// error. This is the AC-HRN-FND-007 happy path.
func TestRunHarnessObserve_NoOpWhenLearningDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")

	// Change cwd so runHarnessObserve sees this project root.
	t.Chdir(dir)

	// Build a cobra command for stderr capture.
	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderr)

	// Provide a representative PostToolUse payload — must NOT be read in no-op path.
	withStdin(t, `{"toolName":"Edit"}`, func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve returned error: %v", err)
		}
	})

	// Verify usage-log.jsonl was NOT created.
	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("usage-log.jsonl must not exist when learning.enabled is false; stat err=%v", err)
	}

	// Stderr should be silent — no error path on no-op.
	if stderr.Len() != 0 {
		t.Errorf("stderr must be empty on no-op; got: %q", stderr.String())
	}
}

// ─────────────────────────────────────────────
// TestRunHarnessObserve_PreservesExistingLogWhenDisabled (REQ-HRN-FND-009)
// ─────────────────────────────────────────────

// Verifies that pre-existing usage-log.jsonl entries are NOT deleted or
// truncated when learning.enabled is false. The no-op semantics protect
// the historical record.
func TestRunHarnessObserve_PreservesExistingLogWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: false\n")

	// Seed an existing usage-log.jsonl with two entries.
	logDir := filepath.Join(dir, ".moai", "harness")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("mkdir harness dir: %v", err)
	}
	logPath := filepath.Join(logDir, "usage-log.jsonl")
	seed := `{"timestamp":"2026-05-14T10:00:00Z","event_type":"agent_invocation","subject":"Edit","context_hash":""}
{"timestamp":"2026-05-14T10:05:00Z","event_type":"agent_invocation","subject":"Bash","context_hash":""}
`
	if err := os.WriteFile(logPath, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed usage-log.jsonl: %v", err)
	}

	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"toolName":"Write"}`, func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve error: %v", err)
		}
	})

	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read usage-log.jsonl: %v", err)
	}
	if string(got) != seed {
		t.Errorf("seed entries were modified; want unchanged.\nwant=%q\n got=%q", seed, string(got))
	}
}

// ─────────────────────────────────────────────
// TestRunHarnessObserve_RecordsWhenEnabled (REQ-HRN-FND-010)
// ─────────────────────────────────────────────

// Verifies that when learning.enabled is true, runHarnessObserve appends one
// JSONL entry to usage-log.jsonl containing the four required schema fields:
// timestamp (ISO-8601), event_type, subject, context_hash. REQ-HRN-FND-010.
//
// Together with TestIsHarnessLearningEnabled this constitutes the T-C3 schema
// verification in tasks.md: every observation entry has the four required keys.
func TestRunHarnessObserve_RecordsWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")

	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"toolName":"Edit"}`, func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve error: %v", err)
		}
	})

	logPath := filepath.Join(dir, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl was not created when learning.enabled=true: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly 1 JSONL entry, got %d:\n%s", len(lines), string(data))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("entry is not valid JSON: %v\nraw=%q", err, lines[0])
	}

	// REQ-HRN-FND-010 schema fields. Field names follow the observer's
	// canonical JSONL contract; if these names diverge in a future refactor,
	// update both the observer and this assertion together.
	for _, required := range []string{"timestamp", "event_type", "subject"} {
		if _, ok := entry[required]; !ok {
			t.Errorf("entry missing required field %q\nraw=%q", required, lines[0])
		}
	}
	// context_hash key may be present with empty value (per existing observer behavior).
	if _, hasCtx := entry["context_hash"]; !hasCtx {
		// Accept absence only if observer emits an alternate name; otherwise fail.
		// Current implementation (internal/harness/observer.go) emits "context_hash".
		t.Errorf("entry missing context_hash field; SPEC REQ-HRN-FND-010 requires it.\nraw=%q", lines[0])
	}

	// Subject must reflect the tool name from stdin.
	if entry["subject"] != "Edit" {
		t.Errorf("subject = %v, want \"Edit\"", entry["subject"])
	}
}
