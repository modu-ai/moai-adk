package hook

// Additional coverage tests for compact.go uncovered branches.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestReadWorktrees_MissingFile returns empty string.
func TestReadWorktrees_MissingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := readWorktrees(dir)
	if result != "" {
		t.Errorf("expected empty for missing file, got %q", result)
	}
}

// TestReadWorktrees_ValidJSON returns indented JSON.
func TestReadWorktrees_ValidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	_ = os.MkdirAll(stateDir, 0o755)

	content := `{"worktrees":[{"name":"worker-SPEC-001","path":"/tmp/wt1"}]}`
	_ = os.WriteFile(filepath.Join(stateDir, "worktrees.json"), []byte(content), 0o644)

	result := readWorktrees(dir)
	if result == "" {
		t.Error("expected non-empty result for valid JSON")
	}
	// Result should be indented JSON.
	var parsed any
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("result is not valid JSON: %v\nresult: %q", err, result)
	}
}

// TestReadWorktrees_MalformedJSON returns raw data as string.
func TestReadWorktrees_MalformedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	_ = os.MkdirAll(stateDir, 0o755)

	raw := "not json at all"
	_ = os.WriteFile(filepath.Join(stateDir, "worktrees.json"), []byte(raw), 0o644)

	result := readWorktrees(dir)
	if result != raw {
		t.Errorf("expected raw data for malformed JSON, got %q", result)
	}
}

// TestReadPersistentMode_MissingFile returns empty string.
func TestReadPersistentMode_MissingFile(t *testing.T) {
	t.Parallel()

	result := readPersistentMode(t.TempDir())
	if result != "" {
		t.Errorf("expected empty for missing file, got %q", result)
	}
}

// TestReadPersistentMode_ValidJSON returns formatted JSON.
func TestReadPersistentMode_ValidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	_ = os.MkdirAll(stateDir, 0o755)

	content := `{"mode":"solo","session_id":"abc123"}`
	_ = os.WriteFile(filepath.Join(stateDir, "persistent-mode.json"), []byte(content), 0o644)

	result := readPersistentMode(dir)
	if result == "" {
		t.Error("expected non-empty result for valid JSON")
	}
}

// TestReadPersistentMode_MalformedJSON returns raw data.
func TestReadPersistentMode_MalformedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	_ = os.MkdirAll(stateDir, 0o755)

	raw := "malformed"
	_ = os.WriteFile(filepath.Join(stateDir, "persistent-mode.json"), []byte(raw), 0o644)

	result := readPersistentMode(dir)
	if result != raw {
		t.Errorf("expected raw data for malformed JSON, got %q", result)
	}
}
