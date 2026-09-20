package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readRuleLoadAudit returns the parsed rows of the audit log under root, and
// whether the file exists at all. A missing file is the "dropped" state the
// t1032 change removes, so the two outcomes are kept distinguishable.
func readRuleLoadAudit(t *testing.T, root string) ([]RuleLoadAuditRecord, bool) {
	t.Helper()
	path := filepath.Join(root, ".moai", "logs", ruleLoadAuditFileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false
	}
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	var recs []RuleLoadAuditRecord
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var rec RuleLoadAuditRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("unmarshal audit row %q: %v", line, err)
		}
		recs = append(recs, rec)
	}
	return recs, true
}

// TestInstructionsLoaded_AuditCarriesHostFields is the positive case: the three
// host-supplied fields the handler used to drop (load_reason, globs,
// trigger_file_path) reach the persistent audit row.
func TestInstructionsLoaded_AuditCarriesHostFields(t *testing.T) {
	root := t.TempDir()
	h := NewInstructionsLoadedHandler()

	if _, err := h.Handle(context.Background(), &HookInput{
		SessionID:       "sess-1",
		HookEventName:   "InstructionsLoaded",
		CWD:             root,
		FilePath:        filepath.Join(root, ".claude", "rules", "x.md"),
		LoadReason:      "path_glob_match",
		Globs:           []string{"internal/hook/**", "internal/cli/**"},
		TriggerFilePath: "internal/hook/instructions_loaded.go",
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	recs, exists := readRuleLoadAudit(t, root)
	if !exists {
		t.Fatal("audit log absent — the three fields are still dropped")
	}
	if len(recs) != 1 {
		t.Fatalf("rows = %d, want 1", len(recs))
	}
	got := recs[0]
	if got.LoadReason != "path_glob_match" {
		t.Errorf("LoadReason = %q, want %q", got.LoadReason, "path_glob_match")
	}
	if len(got.Globs) != 2 || got.Globs[0] != "internal/hook/**" || got.Globs[1] != "internal/cli/**" {
		t.Errorf("Globs = %v, want [internal/hook/** internal/cli/**]", got.Globs)
	}
	if got.TriggerFilePath != "internal/hook/instructions_loaded.go" {
		t.Errorf("TriggerFilePath = %q, want the trigger path", got.TriggerFilePath)
	}
	if got.SessionID != "sess-1" {
		t.Errorf("SessionID = %q, want sess-1", got.SessionID)
	}
	if got.Timestamp == "" {
		t.Error("Timestamp empty")
	}
}

// TestInstructionsLoaded_AuditOmitsAbsentHostFields is the negative control:
// when the host supplies none of the three fields, the row records their
// absence rather than inventing values. Without this the positive case could
// pass on hardcoded output.
func TestInstructionsLoaded_AuditOmitsAbsentHostFields(t *testing.T) {
	root := t.TempDir()
	h := NewInstructionsLoadedHandler()

	if _, err := h.Handle(context.Background(), &HookInput{
		SessionID:     "sess-2",
		HookEventName: "InstructionsLoaded",
		CWD:           root,
		FilePath:      filepath.Join(root, "CLAUDE.md"),
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	recs, exists := readRuleLoadAudit(t, root)
	if !exists {
		t.Fatal("audit log absent")
	}
	if len(recs) != 1 {
		t.Fatalf("rows = %d, want 1", len(recs))
	}
	if got := recs[0]; got.LoadReason != "" || len(got.Globs) != 0 || got.TriggerFilePath != "" {
		t.Errorf("absent host fields materialized: %+v", got)
	}
}

// TestInstructionsLoaded_AuditSkippedWithoutRoot keeps the audit write from
// guessing a destination when the event carries no cwd.
func TestInstructionsLoaded_AuditSkippedWithoutRoot(t *testing.T) {
	root := t.TempDir()
	h := NewInstructionsLoadedHandler()

	if _, err := h.Handle(context.Background(), &HookInput{
		SessionID:     "sess-3",
		HookEventName: "InstructionsLoaded",
		LoadReason:    "session_start",
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if _, exists := readRuleLoadAudit(t, root); exists {
		t.Error("audit row written despite empty CWD")
	}
}
