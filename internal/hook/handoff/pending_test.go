package handoff

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeDecoyPendingMD creates a session-handoff/pending.md decoy and returns its
// path + initial modtime, so tests can assert the reverse-handoff flow never
// touches the SessionEnd flow's file (path-isolation guard).
func writeDecoyPendingMD(t *testing.T, projectDir string) (string, time.Time) {
	t.Helper()
	dir := filepath.Join(projectDir, ".moai", "state", "session-handoff")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session-handoff: %v", err)
	}
	p := filepath.Join(dir, "pending.md")
	if err := os.WriteFile(p, []byte("---\nsprint: s\nspec: x\nstatus: c\nindex_line: l\n---\n## Next Session Entry Point\n```text\nx\n```\n"), 0o644); err != nil {
		t.Fatalf("write decoy pending.md: %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat decoy: %v", err)
	}
	return p, info.ModTime()
}

// TestSavePending_WritesJSONNotMarkdown verifies REQ-005: save writes
// handoff/pending.json (valid JSON) and does NOT create/modify
// session-handoff/pending.md.
func TestSavePending_WritesJSONNotMarkdown(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	decoyPath, decoyMTime := writeDecoyPendingMD(t, projectDir)

	rec := &PendingRecord{Body: "resume body"}
	if err := SavePending(projectDir, rec); err != nil {
		t.Fatalf("SavePending: %v", err)
	}

	// pending.json exists and is valid JSON.
	data, err := os.ReadFile(PendingPath(projectDir))
	if err != nil {
		t.Fatalf("read pending.json: %v", err)
	}
	var parsed PendingRecord
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("pending.json is not valid JSON: %v", err)
	}
	if parsed.Body != "resume body" {
		t.Errorf("body: got %q, want %q", parsed.Body, "resume body")
	}

	// session-handoff/pending.md untouched (mtime unchanged).
	info, err := os.Stat(decoyPath)
	if err != nil {
		t.Fatalf("stat decoy after save: %v", err)
	}
	if !info.ModTime().Equal(decoyMTime) {
		t.Errorf("session-handoff/pending.md was modified by save (mtime changed) — path isolation violated")
	}
}

// TestSavePending_Schema verifies REQ-006: the persisted record carries at least
// schema_version, body (verbatim), directives, conversation_language, saved_at.
func TestSavePending_Schema(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	before := time.Now().Add(-time.Second)

	rec := &PendingRecord{
		Body:                 "verbatim\nbody\nwith\nnewlines",
		ConversationLanguage: "ko",
		Directives:           Directives{Ultrathink: true},
	}
	if err := SavePending(projectDir, rec); err != nil {
		t.Fatalf("SavePending: %v", err)
	}

	data, err := os.ReadFile(PendingPath(projectDir))
	if err != nil {
		t.Fatalf("read pending.json: %v", err)
	}
	// Assert the raw JSON carries the required keys.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	for _, key := range []string{"schema_version", "body", "directives", "conversation_language", "saved_at"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("pending.json missing required key %q", key)
		}
	}

	var parsed PendingRecord
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal parsed: %v", err)
	}
	if parsed.SchemaVersion != PendingSchemaVersion {
		t.Errorf("schema_version: got %d, want %d", parsed.SchemaVersion, PendingSchemaVersion)
	}
	if parsed.Body != "verbatim\nbody\nwith\nnewlines" {
		t.Errorf("body not verbatim: got %q", parsed.Body)
	}
	if !parsed.Directives.Ultrathink {
		t.Errorf("directives.ultrathink: got false, want true")
	}
	if parsed.ConversationLanguage != "ko" {
		t.Errorf("conversation_language: got %q, want %q", parsed.ConversationLanguage, "ko")
	}
	if parsed.SavedAt.Before(before) {
		t.Errorf("saved_at not auto-populated: got %v", parsed.SavedAt)
	}
}

// TestClearPending verifies REQ-007: clear removes handoff/pending.json and does
// NOT touch session-handoff/pending.md. An absent file is a no-op.
func TestClearPending(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	decoyPath, decoyMTime := writeDecoyPendingMD(t, projectDir)

	if err := SavePending(projectDir, &PendingRecord{Body: "b"}); err != nil {
		t.Fatalf("SavePending: %v", err)
	}
	if _, err := os.Stat(PendingPath(projectDir)); err != nil {
		t.Fatalf("pending.json should exist before clear: %v", err)
	}

	if err := ClearPending(projectDir); err != nil {
		t.Fatalf("ClearPending: %v", err)
	}
	if _, err := os.Stat(PendingPath(projectDir)); !os.IsNotExist(err) {
		t.Errorf("pending.json should be removed after clear, stat err: %v", err)
	}

	// Decoy untouched.
	info, err := os.Stat(decoyPath)
	if err != nil {
		t.Fatalf("stat decoy after clear: %v", err)
	}
	if !info.ModTime().Equal(decoyMTime) {
		t.Errorf("session-handoff/pending.md was modified by clear — path isolation violated")
	}

	// Second clear (absent file) is a no-op.
	if err := ClearPending(projectDir); err != nil {
		t.Errorf("ClearPending on absent file should be a no-op, got: %v", err)
	}
}

// TestReadPending_States verifies the three ReadPending states: absent, valid,
// corrupt (the present bool drives the SessionStart handler's no-op vs warn).
func TestReadPending_States(t *testing.T) {
	t.Parallel()

	t.Run("absent", func(t *testing.T) {
		t.Parallel()
		rec, present, err := ReadPending(t.TempDir())
		if rec != nil || present || err != nil {
			t.Errorf("absent: got (%v, %v, %v), want (nil, false, nil)", rec, present, err)
		}
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		pd := t.TempDir()
		if err := SavePending(pd, &PendingRecord{Body: "b", ConversationLanguage: "en"}); err != nil {
			t.Fatalf("SavePending: %v", err)
		}
		rec, present, err := ReadPending(pd)
		if err != nil || !present || rec == nil {
			t.Fatalf("valid: got (%v, %v, %v), want (rec, true, nil)", rec, present, err)
		}
		if rec.Body != "b" {
			t.Errorf("body: got %q, want %q", rec.Body, "b")
		}
	})

	t.Run("corrupt", func(t *testing.T) {
		t.Parallel()
		pd := t.TempDir()
		dir := handoffStateDir(pd)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(PendingPath(pd), []byte("{not valid json"), 0o600); err != nil {
			t.Fatalf("write corrupt: %v", err)
		}
		rec, present, err := ReadPending(pd)
		if err == nil || !present || rec != nil {
			t.Errorf("corrupt: got (%v, %v, %v), want (nil, true, err)", rec, present, err)
		}
	})
}

// TestSavePending_FilePerm verifies pending.json is written 0o600 (the resume
// body may carry session context; mirror the persist.go 0o600 discipline).
func TestSavePending_FilePerm(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	if err := SavePending(projectDir, &PendingRecord{Body: "b"}); err != nil {
		t.Fatalf("SavePending: %v", err)
	}
	info, err := os.Stat(PendingPath(projectDir))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("pending.json perm: got %o, want 0600", perm)
	}
}
