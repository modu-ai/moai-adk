package hook

// user_decision_capture_coverage_test.go — focused coverage tests for the
// user_decision_capture subpipeline helper functions. These complement the
// AC-driven tests in user_decision_capture_test.go by exercising the domain
// classification branches, the warn-fail-open paths, and the tool_input-only
// selection fallback that the AC tests do not all reach.
//
// The goal is to lift the captureUserDecision entry-function coverage and
// its helpers above the acceptance.md §D.6 85% threshold for the extension
// branch.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestClassifyDomain_Branches exercises every classifyDomain switch arm so
// the domain classification has full branch coverage.
func TestClassifyDomain_Branches(t *testing.T) {
	cases := map[string]string{
		"Tier 선택":          "complexity_tier",
		"티어":               "complexity_tier",
		"언어":               "language",
		"lang":             "language",
		"effort":           "effort",
		"노력":               "effort",
		"진행 방향":            "workflow_direction",
		"direction":        "workflow_direction",
		"agent delegation": "agent_delegation",
		"에이전트":            "agent_delegation",
		"branch":          "git_strategy",
		"브랜치":             "git_strategy",
		"PR strategy":     "pr_strategy",
		"design":          "design",
		"디자인":             "design",
		"scope":           "scope",
		"범위":              "scope",
		"unknown thing":   "general",
		"":                "general",
	}
	for header, want := range cases {
		got := classifyDomain(header)
		if got != want {
			t.Errorf("classifyDomain(%q) = %q, want %q", header, got, want)
		}
	}
}

// TestCaptureUserDecision_WarnFailOpen_UpsertError exercises the warnFailOpen
// path by forcing the store upsert to fail (read-only user_decisions dir so
// the atomic-write temp creation fails inside the store).
func TestCaptureUserDecision_WarnFailOpen_UpsertError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission injection ineffective as root")
	}
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	memDir := resolveMemoryDirForTest(t, dir)
	udDir := filepath.Join(memDir, "user_decisions")
	if err := os.MkdirAll(filepath.Join(udDir, "archival"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Make udDir read-only so atomicWrite's temp-file creation fails. Touch the
	// tier files first so NewFileStore's Stat check passes.
	for _, n := range []string{"core.yaml", "recall.jsonl"} {
		if err := os.WriteFile(filepath.Join(udDir, n), nil, 0o644); err != nil {
			t.Fatalf("touch %s: %v", n, err)
		}
	}
	if err := os.Chmod(udDir, 0o500); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(udDir, 0o755) })

	input := &HookInput{
		ToolName:     askUserQuestionTool,
		ToolInput:    mustMarshal(t, map[string]any{"questions": []map[string]any{{"header": "direction", "question": "x"}}}),
		ToolResponse: mustMarshal(t, map[string]any{"selected_option_label": "proceed"}),
		SessionID:    "sess-warn-upsert",
		CWD:          dir,
	}
	captureUserDecision(input) // must not panic / block
}

// TestCaptureUserDecision_WarnFailOpen_StoreConstructionError exercises
// the store-construction error path by planting a file where a dir is
// expected (NewFileStore MkdirAll fails).
func TestCaptureUserDecision_WarnFailOpen_StoreConstructionError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	memDir := resolveMemoryDirForTest(t, dir)
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Plant a file where user_decisions dir should go — NewFileStore fails.
	if err := os.WriteFile(filepath.Join(memDir, "user_decisions"), []byte("x"), 0o644); err != nil {
		t.Fatalf("plant: %v", err)
	}

	input := &HookInput{
		ToolName:     askUserQuestionTool,
		ToolInput:    mustMarshal(t, map[string]any{"questions": []map[string]any{{"header": "tier"}}}),
		ToolResponse: mustMarshal(t, map[string]any{"selected_option_label": "M"}),
		SessionID:    "sess-warn-store",
		CWD:          dir,
	}
	captureUserDecision(input) // must not panic / block
}

// TestCaptureUserDecision_WarnFailOpen_ResolveError exercises the
// memory-dir resolution error path by unsetting HOME so os.UserHomeDir
// fails.
func TestCaptureUserDecision_WarnFailOpen_ResolveError(t *testing.T) {
	t.Setenv("HOME", "")
	input := &HookInput{
		ToolName:     askUserQuestionTool,
		ToolInput:    mustMarshal(t, map[string]any{"questions": []map[string]any{{"header": "x"}}}),
		ToolResponse: mustMarshal(t, map[string]any{"selected_option_label": "y"}),
		SessionID:    "sess-warn-resolve",
		CWD:          "/tmp",
	}
	captureUserDecision(input) // must not panic / block
}

// TestCaptureUserDecision_ToolInputOnlySelection exercises the tool_input
// fallback path where the selection lives in tool_input rather than
// tool_response.
func TestCaptureUserDecision_ToolInputOnlySelection(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	memDir := resolveMemoryDirForTest(t, dir)
	if err := os.MkdirAll(filepath.Join(memDir, "user_decisions", "archival"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// tool_response is empty; selection is embedded in tool_input.
	input := &HookInput{
		ToolName: askUserQuestionTool,
		ToolInput: mustMarshal(t, map[string]any{
			"questions":            []map[string]any{{"header": "lang", "question": "언어?"}},
			"selected_option_label": "ko",
		}),
		ToolResponse: nil,
		SessionID:    "sess-input-only",
		CWD:          dir,
	}
	captureUserDecision(input)

	recallPath := filepath.Join(memDir, "user_decisions", "recall.jsonl")
	data, err := os.ReadFile(recallPath)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected capture from tool_input fallback; recall.jsonl empty or unreadable: %v", err)
	}
	if !strings.Contains(string(data), `"ko"`) {
		t.Errorf("tool_input fallback did not capture the selected option; recall=%s", string(data))
	}
}

// TestCaptureUserDecision_CLAUDEProjectDirFallback exercises the
// CLAUDE_PROJECT_DIR fallback in resolveCaptureMemoryDir when input.CWD is
// empty.
func TestCaptureUserDecision_CLAUDEProjectDirFallback(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	memDir := resolveMemoryDirForTest(t, dir)
	if err := os.MkdirAll(filepath.Join(memDir, "user_decisions", "archival"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Empty CWD — resolver must fall back to CLAUDE_PROJECT_DIR.
	input := &HookInput{
		ToolName:     askUserQuestionTool,
		ToolInput:    mustMarshal(t, map[string]any{"questions": []map[string]any{{"header": "effort"}}}),
		ToolResponse: mustMarshal(t, map[string]any{"selected_option_label": "high"}),
		SessionID:    "sess-cpd-fallback",
		CWD:          "",
	}
	captureUserDecision(input)

	recallPath := filepath.Join(memDir, "user_decisions", "recall.jsonl")
	if data, _ := os.ReadFile(recallPath); len(data) == 0 {
		t.Errorf("CLAUDE_PROJECT_DIR fallback did not resolve; recall.jsonl empty")
	}
}

// TestCaptureUserDecision_NilInputSafe verifies the defensive nil check at
// the top of captureUserDecision does not panic on a nil input.
func TestCaptureUserDecision_NilInputSafe(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("captureUserDecision panicked on nil input: %v", r)
		}
	}()
	captureUserDecision(nil) // must no-op, no panic
}

// TestCaptureUserDecision_WrongToolNameSafe verifies the tool_name re-check
// returns early for a non-AskUserQuestion tool.
func TestCaptureUserDecision_WrongToolNameSafe(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	input := &HookInput{
		ToolName: "Write",
		CWD:      dir,
	}
	captureUserDecision(input) // must no-op

	// No artifacts should be created.
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
		t.Errorf("captureUserDecision created artifacts for non-target tool")
	}
}

// TestCaptureUserDecision_SourceCitationVariants exercises both the
// tool_use_id-present and tool_use_id-absent citation formats.
func TestCaptureUserDecision_SourceCitationVariants(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	memDir := resolveMemoryDirForTest(t, dir)
	_ = os.MkdirAll(filepath.Join(memDir, "user_decisions", "archival"), 0o755)

	// With tool_use_id.
	h := NewPostToolHandler()
	input := &HookInput{
		ToolName:     askUserQuestionTool,
		ToolInput:    mustMarshal(t, map[string]any{"questions": []map[string]any{{"header": "branch"}}}),
		ToolResponse: mustMarshal(t, map[string]any{"selected_option_label": "feat/x"}),
		SessionID:    "sess-cite",
		ToolUseID:    "tooluse-abc",
		CWD:          dir,
	}
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	recallPath := filepath.Join(memDir, "user_decisions", "recall.jsonl")
	data, err := os.ReadFile(recallPath)
	if err != nil || len(data) == 0 {
		t.Fatalf("recall.jsonl empty: %v", err)
	}
	if !strings.Contains(string(data), "tooluse-abc") {
		t.Errorf("SourceCitation does not carry tool_use_id; recall=%s", string(data))
	}
}

// TestParseCapturedResult_NoSelection verifies the parse returns ok=false
// when neither tool_response nor tool_input carries a selection.
func TestParseCapturedResult_NoSelection(t *testing.T) {
	input := &HookInput{
		ToolInput:    json.RawMessage(`{"questions":[{"header":"x","question":"y"}]}`),
		ToolResponse: json.RawMessage(`{}`),
	}
	selected, header, ok := parseCapturedResult(input)
	if ok {
		t.Errorf("expected ok=false when no selection present; got selected=%q header=%q", selected, header)
	}
}

// TestParseCapturedResult_HeaderFromToolInput verifies the header is
// extracted from tool_input when tool_response lacks it.
func TestParseCapturedResult_HeaderFromToolInput(t *testing.T) {
	input := &HookInput{
		ToolInput:    json.RawMessage(`{"questions":[{"header":"from_input","question":"y"}]}`),
		ToolResponse: json.RawMessage(`{"selected_option_label":"picked"}`),
	}
	selected, header, ok := parseCapturedResult(input)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if selected != "picked" {
		t.Errorf("selected=%q want %q", selected, "picked")
	}
	if header != "from_input" {
		t.Errorf("header=%q want %q", header, "from_input")
	}
}

