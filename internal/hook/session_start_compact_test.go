package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAC007_CompactReinjectEmitsAllFourElements asserts the SessionStart(compact)
// re-inject emits (a) goal condition, (b) active SPEC-id, (c) last-verified
// mechanical state, (d) single next action — when a goal is armed and a SPEC is
// in-progress. SPEC-INFINITE-GOAL-001 REQ-5 / AC-007.
func TestAC007_CompactReinjectEmitsAllFourElements(t *testing.T) {
	tmp := t.TempDir()
	sessionID := "compact-session"

	// Armed goal.
	goalDir := filepath.Join(tmp, ".moai", "state", "goal")
	if err := os.MkdirAll(goalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	goalJSON := `{"session_id":"` + sessionID + `","goal":"all AC rows show PASS exits 0","conditions":[{"type":"mechanical","cmd":"go test ./...","expect_exit":0}],"ceiling":{"max_turns":0,"max_duration":3600},"turns_used":3,"progress":[{"turn":3,"note":"mechanical condition failed: cmd \"go test ./...\" exited 1","fingerprint":"x"}],"progression_mode":"autonomous","created_at":"","status":"armed"}`
	if err := os.WriteFile(filepath.Join(goalDir, sessionID+".json"), []byte(goalJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	// Active in-progress SPEC.
	specDir := filepath.Join(tmp, ".moai", "specs", "SPEC-TEST-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	specMD := "---\nid: SPEC-TEST-001\nstatus: in-progress\nupdated: 2026-08-03\n---\n# spec\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o600); err != nil {
		t.Fatal(err)
	}
	progressTail := "## §E.2 Run-phase Evidence\nM1 done. M2 in progress — last: statusline /clear suppression.\n"
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressTail), 0o600); err != nil {
		t.Fatal(err)
	}

	got := renderCompactReinject(tmp, sessionID)
	low := strings.ToLower(got)
	// (a) goal condition
	if !strings.Contains(low, "ac rows") && !strings.Contains(low, "exits 0") {
		t.Errorf("AC-007 (a): reinject must carry the goal condition; got %q", got)
	}
	// (b) active SPEC-id
	if !strings.Contains(got, "SPEC-TEST-001") {
		t.Errorf("AC-007 (b): reinject must carry the active SPEC-id; got %q", got)
	}
	// (c) last-verified mechanical state (the last progress note / failed cond)
	if !strings.Contains(low, "go test") && !strings.Contains(low, "mechanical") {
		t.Errorf("AC-007 (c): reinject must carry last-verified mechanical state; got %q", got)
	}
	// (d) single next action
	if !strings.Contains(low, "next") && !strings.Contains(low, "re-run") && !strings.Contains(low, "advance") {
		t.Errorf("AC-007 (d): reinject must carry a single next action; got %q", got)
	}
}

// TestAC007_CompactReinjectNoOpWhenNoGoal asserts the handler is a no-op (empty
// output) when no goal is armed for the session.
func TestAC007_CompactReinjectNoOpWhenNoGoal(t *testing.T) {
	tmp := t.TempDir()
	got := renderCompactReinject(tmp, "no-goal-session")
	if strings.TrimSpace(got) != "" {
		t.Errorf("AC-007 no-op: expected empty reinject with no armed goal; got %q", got)
	}
}

// TestAC007_CompactHandlerOnlyFiresOnCompactSource asserts the registered handler
// filters by source=="compact" (no-op on startup/resume/clear).
func TestAC007_CompactHandlerOnlyFiresOnCompactSource(t *testing.T) {
	tmp := t.TempDir()
	sessionID := "src-session"
	goalDir := filepath.Join(tmp, ".moai", "state", "goal")
	_ = os.MkdirAll(goalDir, 0o755)
	goalJSON := `{"session_id":"` + sessionID + `","goal":"g exits 0","conditions":[{"type":"mechanical","cmd":"g","expect_exit":0}],"ceiling":{"max_turns":0,"max_duration":3600},"turns_used":1,"progress":[],"progression_mode":"autonomous","created_at":"","status":"armed"}`
	_ = os.WriteFile(filepath.Join(goalDir, sessionID+".json"), []byte(goalJSON), 0o600)

	h := &sessionStartCompactHandler{}
	for _, src := range []string{"startup", "resume", "clear", ""} {
		out, err := h.handle(context.Background(), tmp, &HookInput{Source: src, SessionID: sessionID})
		if err != nil {
			t.Errorf("source %q: unexpected error %v", src, err)
		}
		if out != nil && out.HookSpecificOutput != nil && out.HookSpecificOutput.AdditionalContext != "" {
			t.Errorf("source %q: handler must no-op (only compact fires); got AdditionalContext %q", src, out.HookSpecificOutput.AdditionalContext)
		}
	}
	// compact source → emits.
	out, err := h.handle(context.Background(), tmp, &HookInput{Source: "compact", SessionID: sessionID})
	if err != nil {
		t.Fatalf("compact: unexpected error %v", err)
	}
	if out == nil || out.HookSpecificOutput == nil || out.HookSpecificOutput.AdditionalContext == "" {
		t.Errorf("compact: handler must emit AdditionalContext; got %+v", out)
	}
}
