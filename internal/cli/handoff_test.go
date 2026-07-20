package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// writeDecoySessionHandoffMD creates a session-handoff/pending.md decoy so tests
// can assert the reverse-handoff CLI never touches the SessionEnd flow's file.
func writeDecoySessionHandoffMD(t *testing.T, projectDir string) (string, time.Time) {
	t.Helper()
	dir := filepath.Join(projectDir, ".moai", "state", "session-handoff")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session-handoff: %v", err)
	}
	p := filepath.Join(dir, "pending.md")
	if err := os.WriteFile(p, []byte("decoy markdown handoff\n"), 0o644); err != nil {
		t.Fatalf("write decoy: %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat decoy: %v", err)
	}
	return p, info.ModTime()
}

// runHandoff executes an isolated `moai handoff` command tree with the given args.
func runHandoff(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	cmd := newHandoffCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if stdin != "" {
		cmd.SetIn(strings.NewReader(stdin))
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// TestHandoffSave_WritesJSONNotMarkdown verifies AC-005: save writes
// handoff/pending.json (valid JSON) and leaves session-handoff/pending.md
// untouched (decoy mtime unchanged).
func TestHandoffSave_WritesJSONNotMarkdown(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	decoyPath, decoyMTime := writeDecoySessionHandoffMD(t, pd)

	out, err := runHandoff(t, "", "save", "--project-dir", pd, "--body", "resume body 6-block")
	if err != nil {
		t.Fatalf("handoff save: %v (out: %s)", err, out)
	}

	data, err := os.ReadFile(handoff.PendingPath(pd))
	if err != nil {
		t.Fatalf("read pending.json: %v", err)
	}
	var parsed handoff.PendingRecord
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("pending.json invalid JSON: %v", err)
	}
	if parsed.Body != "resume body 6-block" {
		t.Errorf("body: got %q", parsed.Body)
	}

	info, err := os.Stat(decoyPath)
	if err != nil {
		t.Fatalf("stat decoy: %v", err)
	}
	if !info.ModTime().Equal(decoyMTime) {
		t.Error("session-handoff/pending.md modified by save — path isolation violated")
	}
}

// TestHandoffSave_Schema verifies AC-006: pending.json carries schema_version,
// body (verbatim), directives.ultrathink==true, conversation_language, saved_at.
func TestHandoffSave_Schema(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	body := "ultrathink. SPEC-X run 진입.\n실행: /moai run SPEC-X"

	out, err := runHandoff(t, "", "save", "--project-dir", pd, "--body", body, "--ultrathink", "--lang", "ko", "--spec", "SPEC-X")
	if err != nil {
		t.Fatalf("handoff save: %v (out: %s)", err, out)
	}

	data, err := os.ReadFile(handoff.PendingPath(pd))
	if err != nil {
		t.Fatalf("read pending.json: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	for _, key := range []string{"schema_version", "body", "directives", "conversation_language", "saved_at"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("pending.json missing required key %q", key)
		}
	}

	var rec handoff.PendingRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("unmarshal rec: %v", err)
	}
	if rec.SchemaVersion != handoff.PendingSchemaVersion {
		t.Errorf("schema_version: got %d, want %d", rec.SchemaVersion, handoff.PendingSchemaVersion)
	}
	if rec.Body != body {
		t.Errorf("body not verbatim: got %q, want %q", rec.Body, body)
	}
	if !rec.Directives.Ultrathink {
		t.Error("directives.ultrathink: got false, want true")
	}
	if rec.ConversationLanguage != "ko" {
		t.Errorf("conversation_language: got %q, want ko", rec.ConversationLanguage)
	}
	if rec.SavedAt.IsZero() {
		t.Error("saved_at not populated")
	}
}

// TestHandoffSave_Stdin verifies the --stdin body path.
func TestHandoffSave_Stdin(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	out, err := runHandoff(t, "piped resume body", "save", "--project-dir", pd, "--stdin")
	if err != nil {
		t.Fatalf("handoff save --stdin: %v (out: %s)", err, out)
	}
	rec, present, err := handoff.ReadPending(pd)
	if err != nil || !present {
		t.Fatalf("ReadPending: (%v, %v, %v)", rec, present, err)
	}
	if rec.Body != "piped resume body" {
		t.Errorf("stdin body: got %q", rec.Body)
	}
}

// TestHandoffSave_RequiresBody verifies an empty body is rejected.
func TestHandoffSave_RequiresBody(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	_, err := runHandoff(t, "", "save", "--project-dir", pd, "--body", "   ")
	if err == nil {
		t.Error("expected error for empty body, got nil")
	}
	if _, statErr := os.Stat(handoff.PendingPath(pd)); !os.IsNotExist(statErr) {
		t.Error("pending.json should not be written when body is empty")
	}
}

// TestHandoffClear verifies AC-007: clear removes handoff/pending.json and leaves
// session-handoff/pending.md untouched.
func TestHandoffClear(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	decoyPath, decoyMTime := writeDecoySessionHandoffMD(t, pd)

	if _, err := runHandoff(t, "", "save", "--project-dir", pd, "--body", "b"); err != nil {
		t.Fatalf("pre-save: %v", err)
	}
	if _, err := os.Stat(handoff.PendingPath(pd)); err != nil {
		t.Fatalf("pending.json should exist before clear: %v", err)
	}

	if _, err := runHandoff(t, "", "clear", "--project-dir", pd); err != nil {
		t.Fatalf("handoff clear: %v", err)
	}
	if _, err := os.Stat(handoff.PendingPath(pd)); !os.IsNotExist(err) {
		t.Errorf("pending.json should be removed after clear, err: %v", err)
	}

	info, err := os.Stat(decoyPath)
	if err != nil {
		t.Fatalf("stat decoy: %v", err)
	}
	if !info.ModTime().Equal(decoyMTime) {
		t.Error("session-handoff/pending.md modified by clear — path isolation violated")
	}
}

// TestHandoffCmdRegistered verifies the handoff command self-registers on rootCmd
// via init() (glm.go/doctor.go pattern).
func TestHandoffCmdRegistered(t *testing.T) {
	t.Parallel()

	var found *struct{}
	for _, c := range rootCmd.Commands() {
		if c.Name() == "handoff" {
			found = &struct{}{}
			// Verify save + clear subcommands are present.
			subs := map[string]bool{}
			for _, s := range c.Commands() {
				subs[s.Name()] = true
			}
			if !subs["save"] || !subs["clear"] {
				t.Errorf("handoff subcommands: got %v, want save+clear", subs)
			}
		}
	}
	if found == nil {
		t.Error("handoff command not registered on rootCmd")
	}
}
