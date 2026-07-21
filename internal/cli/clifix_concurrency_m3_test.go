package cli

// SPEC-CLIFIX-CONCURRENCY-001 M3 — atomic-writer consolidation characterization tests.
//
// These tests pin the OBSERVABLE BEHAVIOR of each former atomic-writer site's
// output (file content, file mode, JSON trailing-newline convention) so that the
// M3 consolidation (6 sites → 1 shared writeFileAtomic helper) can prove behavior
// preservation: the tests pass against the pre-consolidation code (2e4513f90)
// AND against the post-consolidation commit.
//
// DDD PRESERVE discipline inside a tdd loop: characterize first, consolidate,
// prove the characterization tests still PASS (GREEN = behavior preserved).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ── writeFileAtomic (the consolidated helper) ─────────────────────────────────

// TestCharacterize_WriteFileAtomic_Perm0600 pins the 0600 perm used by
// credential-bearing files (settings.local.json, ~/.claude.json, merge-history
// cache). The consolidated helper MUST preserve tmp.Chmod(perm).
func TestCharacterize_WriteFileAtomic_Perm0600(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.json")
	data := []byte(`{"key":"value"}`)
	if err := writeFileAtomic(path, data, 0o600); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o600 {
		t.Errorf("perm = %04o, want 0600", perm)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("content mismatch: got %q, want %q", got, data)
	}
}

// TestCharacterize_WriteFileAtomic_Perm0644 pins the 0644 perm used by
// non-credential config files (workflow.yaml). The harness_mute inline site
// previously used os.WriteFile(tmp, out, 0o644); the consolidated helper MUST
// produce the same 0644 mode.
func TestCharacterize_WriteFileAtomic_Perm0644(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte("key: value\n")
	if err := writeFileAtomic(path, data, 0o644); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o644 {
		t.Errorf("perm = %04o, want 0644", perm)
	}
}

// TestCharacterize_WriteFileAtomic_NoPartialContent verifies the atomic-rename
// guarantee: the target path is never observed in a half-written state. After
// the call, the file exists with the full content; there is no leftover temp.
func TestCharacterize_WriteFileAtomic_NoPartialContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out.bin")
	data := []byte("complete payload")
	if err := writeFileAtomic(path, data, 0o600); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("partial content: got %q, want %q", got, data)
	}
	// No leftover temp files in the directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected exactly 1 file (the target), got %d: %v", len(entries), entries)
	}
}

// TestCharacterize_WriteFileAtomic_CreatesParentDir verifies the safe-superset
// MkdirAll behavior added during consolidation (writeClaudeJSONBytes and
// preference/atomicWrite already had it; writeFileAtomic gains it as the union).
func TestCharacterize_WriteFileAtomic_CreatesParentDir(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	nested := filepath.Join(base, "a", "b", "c", "file.json")
	data := []byte(`{}`)
	if err := writeFileAtomic(nested, data, 0o600); err != nil {
		t.Fatalf("writeFileAtomic nested: %v", err)
	}
	got, err := os.ReadFile(nested)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != `{}` {
		t.Errorf("content: got %q, want {}", got)
	}
}

// ── saveMergeHistoryLedger (former atomicWriteJSON caller) ────────────────────

// TestCharacterize_SaveMergeHistoryLedger_JSONTrailingNewline pins the
// json.Encoder.Encode trailing-newline behavior. atomicWriteJSON uses
// json.NewEncoder(tmp).Encode(value) which appends '\n' after the JSON value.
// The consolidated path (marshal via Encoder into a buffer + writeFileAtomic)
// MUST preserve this trailing newline.
func TestCharacterize_SaveMergeHistoryLedger_JSONTrailingNewline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ledger := map[string]mergeHistoryEntry{
		"file.yaml":     {FallbackCount: 1, LastFailedAt: "2026-07-11T00:00:00Z"},
		"other/file.go": {FallbackCount: 3, LastFailedAt: "2026-07-10T12:00:00Z"},
	}
	if err := saveMergeHistoryLedger(dir, ledger); err != nil {
		t.Fatalf("saveMergeHistoryLedger: %v", err)
	}
	path := filepath.Join(dir, ".moai", "cache", "merge-history.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read merge-history.json: %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Errorf("merge-history.json missing trailing newline (Encoder.Encode behavior)")
	}
	var got map[string]mergeHistoryEntry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
	// File mode is 0600 (CreateTemp default, credential-bearing cache).
	info, _ := os.Stat(path)
	if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o600 {
		t.Errorf("perm = %04o, want 0600", perm)
	}
}

// ── saveWorkflowMuteConfig (former harness_mute.go inline caller) ─────────────

// TestCharacterize_SaveWorkflowMuteConfig_Perm0644 pins the 0644 file mode of
// workflow.yaml. The harness_mute inline previously used os.WriteFile(tmp, out,
// 0o644); the consolidated path MUST produce the same 0644 mode.
func TestCharacterize_SaveWorkflowMuteConfig_Perm0644(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	// Seed an existing workflow.yaml so saveWorkflowMuteConfig can parse it.
	seed := []byte("harness:\n  proposal:\n    mute:\n      categories: []\n")
	if err := os.WriteFile(path, seed, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	cfg := workflowMuteConfig{}
	cfg.Harness.Proposal.Mute.Categories = []string{"w1", "w2"}
	if err := saveWorkflowMuteConfig(path, cfg); err != nil {
		t.Fatalf("saveWorkflowMuteConfig: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o644 {
		t.Errorf("perm = %04o, want 0644", perm)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), "w1") || !strings.Contains(string(data), "w2") {
		t.Errorf("mute categories not persisted: %s", data)
	}
}

// ── saveLLMSection (former glm.go inline caller) ──────────────────────────────

// TestCharacterize_SaveLLMSection_WritesLLMYaml pins the on-disk format of
// llm.yaml written by saveLLMSection. The inline previously used CreateTemp +
// write + rename (0600 default); the consolidated path MUST produce valid YAML
// with the llm root key.
func TestCharacterize_SaveLLMSection_WritesLLMYaml(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	llm := config.NewDefaultLLMConfig()
	if err := saveLLMSection(dir, llm); err != nil {
		t.Fatalf("saveLLMSection: %v", err)
	}
	path := filepath.Join(dir, "llm.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read llm.yaml: %v", err)
	}
	if !strings.Contains(string(data), "llm:") {
		t.Errorf("llm.yaml missing root 'llm:' key: %s", data)
	}
	// File mode is 0600 (CreateTemp default, potentially credential-bearing).
	info, _ := os.Stat(path)
	if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o600 {
		t.Errorf("perm = %04o, want 0600", perm)
	}
}

// ── writeClaudeJSONBytes (M2 seam, delegates to consolidated helper) ──────────

// TestCharacterize_WriteClaudeJSONBytes_NoTrailingNewline pins the
// json.MarshalIndent behavior (NO trailing newline, unlike Encoder.Encode).
// writeClaudeJSONAtomic used MarshalIndent; the consolidated path MUST preserve
// the absence of a trailing newline for ~/.claude.json writes.
func TestCharacterize_WriteClaudeJSONBytes_NoTrailingNewline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")
	// MarshalIndent does NOT add a trailing newline (unlike Encoder.Encode).
	root := map[string]any{"mcpServers": map[string]any{"foo": map[string]any{"command": "bar"}}}
	jsonBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := writeClaudeJSONBytes(path, jsonBytes); err != nil {
		t.Fatalf("writeClaudeJSONBytes: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(data) == 0 || data[len(data)-1] == '\n' {
		t.Errorf("claude.json unexpectedly has trailing newline (MarshalIndent behavior)")
	}
	// File mode is 0600 (credential-bearing).
	info, _ := os.Stat(path)
	if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o600 {
		t.Errorf("perm = %04o, want 0600", perm)
	}
}
