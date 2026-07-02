// Package cli — harness-observe wire-format + project-root regression matrix.
//
// Finding 1 (HIGH): the harness-observe family decoded stdin with a single
// hard-coded field convention per handler (a mix of camelCase and snake_case),
// so under the OTHER Claude Code wire format part of each struct silently
// decoded empty:
//   - native CC 2.1.x sends camelCase (lastAssistantMessage, agentId,
//     agentType, agentName, toolName) with nested session.id
//   - the flat legacy format sends all snake_case (last_assistant_message,
//     agent_id, tool_name) with a TOP-LEVEL session_id
//
// The handlers also resolved the project root as a bare os.Getwd() instead of
// the env-first CLAUDE_PROJECT_DIR → os.Getwd() priority mandated by
// internal/hook/CLAUDE.md §B7 (worktree hooks run with cwd inside the
// worktree, not the project root).
//
// These tests feed BOTH wire formats to each handler and assert the fields are
// preserved, and verify the env-first project-root resolution. They fail
// against the pre-fix handlers (field-loss / wrong log location) and pass once
// the handlers route through hook.Protocol.ReadInput (which applies
// normalizeHookInput) and resolve the project root env-first.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

// readObserveWireEntry reads the single JSONL entry from the harness usage log
// under root and returns it as a decoded map. It fails the test when the log is
// missing or does not contain exactly one line.
func readObserveWireEntry(t *testing.T, root string) map[string]any {
	t.Helper()
	logPath := filepath.Join(root, ".moai", "harness", "usage-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("usage-log.jsonl not created under %q: %v", root, err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("JSONL line count: got=%d, want=1\nraw=%q", len(lines), string(data))
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("JSONL parse failed: %v\nraw=%q", err, lines[0])
	}
	return entry
}

// TestRunHarnessObserveStop_NativeCamelWireFormat feeds the native Claude Code
// 2.1.x wire format (camelCase lastAssistantMessage + nested session.id) to the
// Stop handler. The pre-fix handler decoded the snake tag
// `last_assistant_message`, so lastAssistantMessage decoded empty and the
// message hash/len fields were silently dropped.
func TestRunHarnessObserveStop_NativeCamelWireFormat(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	const assistantMsg = "native camel assistant message"
	cmd := &cobra.Command{}
	withStdin(t, `{"lastAssistantMessage":"`+assistantMsg+`","session":{"id":"sess-native"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop returned error: %v", err)
		}
	})

	entry := readObserveWireEntry(t, dir)

	// session.id (nested) must map to session_id.
	if entry["session_id"] != "sess-native" {
		t.Errorf("session_id: got=%v, want=%q", entry["session_id"], "sess-native")
	}
	// The field-loss bug: native camel lastAssistantMessage did not decode →
	// hash/len omitted.
	if _, ok := entry["last_assistant_message_hash"]; !ok {
		t.Errorf("last_assistant_message_hash missing — native camel lastAssistantMessage was not decoded (field loss)")
	}
	if _, ok := entry["last_assistant_message_len"]; !ok {
		t.Errorf("last_assistant_message_len missing — native camel lastAssistantMessage was not decoded (field loss)")
	}
}

// TestRunHarnessObserveStop_FlatSnakeWireFormat feeds the flat legacy format
// (all snake_case with a TOP-LEVEL session_id). The pre-fix handler decoded the
// nested `session.id`, so a top-level session_id decoded empty and session_id
// was silently dropped.
func TestRunHarnessObserveStop_FlatSnakeWireFormat(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	const assistantMsg = "flat snake assistant message"
	cmd := &cobra.Command{}
	withStdin(t, `{"last_assistant_message":"`+assistantMsg+`","session_id":"sess-flat"}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop returned error: %v", err)
		}
	})

	entry := readObserveWireEntry(t, dir)

	// The field-loss bug: a top-level session_id did not decode (handler read
	// the nested session.id only).
	if entry["session_id"] != "sess-flat" {
		t.Errorf("session_id: got=%v, want=%q — top-level session_id was not decoded (field loss)", entry["session_id"], "sess-flat")
	}
	// last_assistant_message (snake) decodes under both formats.
	if _, ok := entry["last_assistant_message_hash"]; !ok {
		t.Errorf("last_assistant_message_hash missing — snake last_assistant_message should decode")
	}
}

// TestRunHarnessObserveSubagentStop_FlatSnakeWireFormat feeds the flat legacy
// format (agent_type / agent_name snake + top-level session_id). The pre-fix
// handler decoded camelCase agentType/agentName and the nested session.id, so
// all four fields decoded empty under this format.
func TestRunHarnessObserveSubagentStop_FlatSnakeWireFormat(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	cmd := &cobra.Command{}
	payload := `{"agent_type":"general-purpose","agent_name":"backend-worker","agent_id":"ag-flat","session_id":"sess-sub-flat"}`
	withStdin(t, payload, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveSubagentStop returned error: %v", err)
		}
	})

	entry := readObserveWireEntry(t, dir)

	// subject = agent_name; the pre-fix handler read camelCase agentName so this
	// decoded empty → subject fell back to "unknown".
	if entry["subject"] != "backend-worker" {
		t.Errorf("subject: got=%v, want=%q — snake agent_name was not decoded (field loss)", entry["subject"], "backend-worker")
	}
	want := map[string]string{
		"agent_name":        "backend-worker",
		"agent_type":        "general-purpose",
		"agent_id":          "ag-flat",
		"parent_session_id": "sess-sub-flat",
	}
	for field, wantVal := range want {
		if entry[field] != wantVal {
			t.Errorf("field %q: got=%v, want=%q (field loss under flat snake format)", field, entry[field], wantVal)
		}
	}
}

// TestRunHarnessObserveSubagentStop_NativeCamelWireFormat feeds the native
// format (camelCase agentId + nested session.id). The pre-fix handler decoded
// the snake agent_id and nested session.id, so agentId (camel) and the parent
// session decoded empty.
func TestRunHarnessObserveSubagentStop_NativeCamelWireFormat(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, dir, true)
	t.Chdir(dir)

	cmd := &cobra.Command{}
	payload := `{"agentType":"general-purpose","agentName":"frontend-worker","agentId":"ag-native","session":{"id":"sess-sub-native"}}`
	withStdin(t, payload, func() {
		if err := runHarnessObserveSubagentStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveSubagentStop returned error: %v", err)
		}
	})

	entry := readObserveWireEntry(t, dir)

	want := map[string]string{
		"subject":           "frontend-worker",
		"agent_name":        "frontend-worker",
		"agent_type":        "general-purpose",
		"agent_id":          "ag-native", // pre-fix: agentId (camel) not decoded
		"parent_session_id": "sess-sub-native",
	}
	for field, wantVal := range want {
		if entry[field] != wantVal {
			t.Errorf("field %q: got=%v, want=%q (field loss under native camel format)", field, entry[field], wantVal)
		}
	}
}

// TestRunHarnessObserve_FlatSnakeWireFormat feeds the flat legacy tool_name
// (snake). The pre-fix PostToolUse handler decoded camelCase toolName, so the
// subject fell back to "unknown".
func TestRunHarnessObserve_FlatSnakeWireFormat(t *testing.T) {
	dir := t.TempDir()
	writeHarnessYAML(t, dir, "learning:\n  enabled: true\n")
	t.Chdir(dir)

	cmd := &cobra.Command{}
	withStdin(t, `{"tool_name":"Bash"}`, func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve returned error: %v", err)
		}
	})

	entry := readObserveWireEntry(t, dir)
	if entry["subject"] != "Bash" {
		t.Errorf("subject: got=%v, want=%q — snake tool_name was not decoded (field loss)", entry["subject"], "Bash")
	}
}

// TestRunHarnessObserve_EnvFirstProjectRoot verifies that the PostToolUse
// observer resolves the project root env-first (CLAUDE_PROJECT_DIR) rather than
// via bare os.Getwd(). The handler runs with cwd set to a separate "worktree"
// directory while CLAUDE_PROJECT_DIR points at the real project root. The
// harness usage log MUST be written under CLAUDE_PROJECT_DIR, not under the
// worktree cwd.
func TestRunHarnessObserve_EnvFirstProjectRoot(t *testing.T) {
	projectDir := t.TempDir()
	worktreeDir := t.TempDir()

	// learning config lives under the project root only.
	writeHarnessYAML(t, projectDir, "learning:\n  enabled: true\n")

	t.Setenv(config.EnvClaudeProjectDir, projectDir)
	t.Chdir(worktreeDir)

	cmd := &cobra.Command{}
	withStdin(t, `{"toolName":"Edit"}`, func() {
		if err := runHarnessObserve(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserve returned error: %v", err)
		}
	})

	// The log must be written under CLAUDE_PROJECT_DIR (env-first), not the cwd.
	projectLog := filepath.Join(projectDir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(projectLog); err != nil {
		t.Errorf("usage-log.jsonl not written under CLAUDE_PROJECT_DIR (env-first project root not honored): %v", err)
	}
	worktreeLog := filepath.Join(worktreeDir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(worktreeLog); err == nil {
		t.Errorf("usage-log.jsonl was written under the cwd worktree — bare os.Getwd() used instead of CLAUDE_PROJECT_DIR")
	}
}

// TestRunHarnessObserveStop_EnvFirstProjectRoot verifies the Stop handler also
// resolves the project root env-first for its config gate + log location.
func TestRunHarnessObserveStop_EnvFirstProjectRoot(t *testing.T) {
	projectDir := t.TempDir()
	worktreeDir := t.TempDir()

	writeHarnessYAML(t, projectDir, "learning:\n  enabled: true\n")
	writeSystemYAMLHookOptIn(t, projectDir, true)

	t.Setenv(config.EnvClaudeProjectDir, projectDir)
	t.Chdir(worktreeDir)

	cmd := &cobra.Command{}
	withStdin(t, `{"lastAssistantMessage":"env-first msg","session":{"id":"sess-envfirst"}}`, func() {
		if err := runHarnessObserveStop(cmd, nil); err != nil {
			t.Fatalf("runHarnessObserveStop returned error: %v", err)
		}
	})

	projectLog := filepath.Join(projectDir, ".moai", "harness", "usage-log.jsonl")
	if _, err := os.Stat(projectLog); err != nil {
		t.Errorf("usage-log.jsonl not written under CLAUDE_PROJECT_DIR (env-first not honored): %v", err)
	}
}
