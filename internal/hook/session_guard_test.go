package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// writeRegistryFixture writes an active-sessions.json fixture at the canonical
// project-relative path under dir, so checkForeignSessionAdvisory reads it via
// filepath.Join(projectDir, session.DefaultRegistryPath).
func writeRegistryFixture(t *testing.T, dir string, entries []session.Entry) {
	t.Helper()
	regPath := filepath.Join(dir, session.DefaultRegistryPath)
	if err := os.MkdirAll(filepath.Dir(regPath), 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal registry entries: %v", err)
	}
	if err := os.WriteFile(regPath, data, 0o644); err != nil {
		t.Fatalf("write registry fixture: %v", err)
	}
}

// TestCheckForeignSessionAdvisory_FailOpen is the core behavior contract
// (SPEC-PREEDIT-PARALLEL-SESSION-GUARD-001 M4, REQ-PES-004): the advisory is
// FAIL-OPEN — it surfaces foreign sessions via stderr + log but ALWAYS returns
// an empty (allow) decision and never blocks the edit.
func TestCheckForeignSessionAdvisory_FailOpen(t *testing.T) {
	now := time.Now().UTC()
	own := "self-session-id"
	foreign1 := session.Entry{SessionID: "other-1", SpecID: "SPEC-A", Phase: "run", StartedAt: now, LastHeartbeat: now, PID: 111, Host: "h", CWD: "/x"}
	foreign2 := session.Entry{SessionID: "other-2", SpecID: "SPEC-B", Phase: "sync", StartedAt: now, LastHeartbeat: now, PID: 222, Host: "h", CWD: "/y"}

	t.Run("ForeignSessions_AllowAndLog", func(t *testing.T) {
		dir := t.TempDir()
		writeRegistryFixture(t, dir, []session.Entry{foreign1, foreign2})
		input := &HookInput{SessionID: own, ToolName: "Edit"}
		decision, reason := checkForeignSessionAdvisory(input, dir)
		if decision != "" {
			t.Fatalf("decision = %q, want \"\" (fail-open must never Deny)", decision)
		}
		if reason != "" {
			t.Fatalf("reason = %q, want \"\"", reason)
		}
		// Advisory side effect: the log file must exist (proves the advisory fired).
		logPath := filepath.Join(dir, preEditAdvisoryLogRelPath)
		if _, err := os.Stat(logPath); err != nil {
			t.Fatalf("advisory log not written for foreign sessions: %v", err)
		}
	})

	t.Run("NoForeign_NoLog", func(t *testing.T) {
		dir := t.TempDir()
		// Only the own session is present — filtered out, zero foreign.
		writeRegistryFixture(t, dir, []session.Entry{
			{SessionID: own, SpecID: "S", Phase: "run", StartedAt: now, LastHeartbeat: now, PID: 1, Host: "h", CWD: "/z"},
		})
		input := &HookInput{SessionID: own, ToolName: "Edit"}
		decision, reason := checkForeignSessionAdvisory(input, dir)
		if decision != "" || reason != "" {
			t.Fatalf("no-foreign: decision=%q reason=%q, want empty", decision, reason)
		}
		logPath := filepath.Join(dir, preEditAdvisoryLogRelPath)
		if _, err := os.Stat(logPath); !os.IsNotExist(err) {
			t.Fatalf("advisory log must NOT exist when no foreign session; stat err=%v", err)
		}
	})

	t.Run("MissingRegistry_FailOpen_NoPanic", func(t *testing.T) {
		dir := t.TempDir() // registry file intentionally absent
		input := &HookInput{SessionID: own, ToolName: "Write"}
		decision, reason := checkForeignSessionAdvisory(input, dir)
		if decision != "" || reason != "" {
			t.Fatalf("missing registry: decision=%q reason=%q, want fail-open empty", decision, reason)
		}
	})

	t.Run("NilInput_FailOpen", func(t *testing.T) {
		decision, reason := checkForeignSessionAdvisory(nil, t.TempDir())
		if decision != "" || reason != "" {
			t.Fatalf("nil input: decision=%q reason=%q, want fail-open empty", decision, reason)
		}
	})

	t.Run("EmptyProjectDir_FailOpen", func(t *testing.T) {
		input := &HookInput{SessionID: own, ToolName: "Edit"}
		decision, reason := checkForeignSessionAdvisory(input, "")
		if decision != "" || reason != "" {
			t.Fatalf("empty projectDir: decision=%q reason=%q, want fail-open empty", decision, reason)
		}
	})
}

// TestCheckForeignSessionAdvisory_Falsifiable is the falsifiability anchor: it
// proves the advisory actually fires on a foreign session and records the count.
// If the foreign filter were inverted, the log write removed, or the function
// short-circuited, this test would fail.
func TestCheckForeignSessionAdvisory_Falsifiable(t *testing.T) {
	now := time.Now().UTC()
	dir := t.TempDir()
	own := "self-id"
	writeRegistryFixture(t, dir, []session.Entry{
		{SessionID: "foreign-x", SpecID: "S", Phase: "run", StartedAt: now, LastHeartbeat: now, PID: 9, Host: "h", CWD: "/p"},
	})
	input := &HookInput{SessionID: own, ToolName: "Write"}
	checkForeignSessionAdvisory(input, dir)
	data, err := os.ReadFile(filepath.Join(dir, preEditAdvisoryLogRelPath))
	if err != nil {
		t.Fatalf("advisory log missing — advisory did not fire on a foreign session: %v", err)
	}
	if !strings.Contains(string(data), "foreign_count=1") {
		t.Fatalf("advisory log missing foreign_count=1 marker: %q", string(data))
	}
}
