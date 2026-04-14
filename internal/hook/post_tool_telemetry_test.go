package hook

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// TestLogSkillUsage_WritesToTelemetryFile verifies that logSkillUsage correctly
// writes a telemetry record when a Skill tool is invoked with a valid project root.
func TestLogSkillUsage_WritesToTelemetryFile(t *testing.T) {
	// No t.Parallel() — t.Setenv requires sequential execution

	// Set up a fake project root with .moai directory
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Override CLAUDE_PROJECT_DIR to point to our temp dir
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	input := &HookInput{
		SessionID: "sess-telemetry-test",
		CWD:       dir,
		ToolName:  "Skill",
		ToolInput: json.RawMessage(`{"skill":"moai-workflow-tdd","args":"implement auth feature"}`),
	}

	logSkillUsage(input)

	// Verify a telemetry file was created today
	dayKey := time.Now().UTC().Format("2006-01-02")
	telPath := filepath.Join(dir, ".moai", "evolution", "telemetry", "usage-"+dayKey+".jsonl")

	data, err := os.ReadFile(telPath)
	if err != nil {
		t.Fatalf("expected telemetry file at %s, got error: %v", telPath, err)
	}

	// Parse and verify the record
	var rec telemetry.UsageRecord
	if err := json.Unmarshal(data[:len(data)-1], &rec); err != nil {
		t.Fatalf("failed to parse JSONL record: %v", err)
	}

	if rec.SkillID != "moai-workflow-tdd" {
		t.Errorf("SkillID = %q, want %q", rec.SkillID, "moai-workflow-tdd")
	}
	if rec.SessionID != "sess-telemetry-test" {
		t.Errorf("SessionID = %q, want %q", rec.SessionID, "sess-telemetry-test")
	}
	if rec.Outcome != telemetry.OutcomeUnknown {
		t.Errorf("Outcome = %q, want %q", rec.Outcome, telemetry.OutcomeUnknown)
	}
	if rec.Trigger != telemetry.TriggerExplicit {
		t.Errorf("Trigger = %q, want %q", rec.Trigger, telemetry.TriggerExplicit)
	}
	if len(rec.ContextHash) != 8 {
		t.Errorf("ContextHash = %q (len=%d), want 8 chars", rec.ContextHash, len(rec.ContextHash))
	}
}

// TestLogSkillUsage_SkipsWhenNoProjectRoot verifies that logSkillUsage silently
// skips recording when there is no MoAI project root.
func TestLogSkillUsage_SkipsWhenNoProjectRoot(t *testing.T) {
	// No t.Parallel() — t.Setenv requires sequential execution

	dir := t.TempDir()
	// Do NOT create .moai directory — so resolveProjectRoot returns ""

	// Clear CLAUDE_PROJECT_DIR to ensure we use CWD
	t.Setenv("CLAUDE_PROJECT_DIR", "")

	input := &HookInput{
		SessionID: "sess-no-root",
		CWD:       dir, // no .moai directory
		ToolName:  "Skill",
		ToolInput: json.RawMessage(`{"skill":"moai-workflow-tdd"}`),
	}

	// Should not panic or return error
	logSkillUsage(input)

	// Verify no telemetry file was created
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if _, err := os.Stat(telDir); !os.IsNotExist(err) {
		t.Errorf("expected no telemetry directory, but got: %v", err)
	}
}

// TestLogSkillUsage_SkipsWhenNoSkillID verifies that logSkillUsage silently
// skips when the skill tool input lacks a "skill" field.
func TestLogSkillUsage_SkipsWhenNoSkillID(t *testing.T) {
	// No t.Parallel() — t.Setenv requires sequential execution

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	input := &HookInput{
		SessionID: "sess-no-skill",
		CWD:       dir,
		ToolName:  "Skill",
		ToolInput: json.RawMessage(`{"args":"some args but no skill field"}`),
	}

	logSkillUsage(input)

	// No telemetry file should be created
	dayKey := time.Now().UTC().Format("2006-01-02")
	telPath := filepath.Join(dir, ".moai", "evolution", "telemetry", "usage-"+dayKey+".jsonl")
	if _, err := os.Stat(telPath); !os.IsNotExist(err) {
		t.Errorf("expected no telemetry file, but found one: %v", err)
	}
}

// TestPostToolHandler_Handle_SkillToolRecorded verifies that the PostToolUse
// handler triggers skill telemetry recording for Skill tool calls.
func TestPostToolHandler_Handle_SkillToolRecorded(t *testing.T) {
	// No t.Parallel() — t.Setenv requires sequential execution

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	h := NewPostToolHandler()
	input := &HookInput{
		SessionID:     "sess-skill-recorded",
		CWD:           dir,
		HookEventName: "PostToolUse",
		ToolName:      "Skill",
		ToolInput:     json.RawMessage(`{"skill":"moai-workflow-tdd"}`),
	}

	ctx := t.Context()
	out, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil output")
	}

	// Verify the telemetry file was created
	dayKey := time.Now().UTC().Format("2006-01-02")
	telPath := filepath.Join(dir, ".moai", "evolution", "telemetry", "usage-"+dayKey+".jsonl")

	f, err := os.Open(telPath)
	if err != nil {
		t.Fatalf("expected telemetry file at %s: %v", telPath, err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 telemetry record, got %d", count)
	}
}
