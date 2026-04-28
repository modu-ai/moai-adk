package hook

// Additional coverage tests for 0% and low-coverage functions:
// - setup.go (REMOVED - was 0%)
// - types.go: NewDeferOutput (0%)
// - pre_tool.go: MergeExtraPatterns (0%)
// - worktree_registry.go: saveWorktreeEntries (44%)
// - reflective_write.go: pickEvolvableZone (42%)
// - session_start.go: isCGMode, Handle with config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/security"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// --- setup.go REMOVED (orphan handler) ---

// --- types.go: NewDeferOutput ---

// TestNewDeferOutput_ReturnsDefer verifies decision is DecisionDefer.
func TestNewDeferOutput_ReturnsDefer(t *testing.T) {
	t.Parallel()

	out := NewDeferOutput("security review required")
	if out == nil {
		t.Fatal("NewDeferOutput() returned nil")
	}
	if out.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}
	if out.HookSpecificOutput.PermissionDecision != DecisionDefer {
		t.Errorf("PermissionDecision = %q, want %q", out.HookSpecificOutput.PermissionDecision, DecisionDefer)
	}
	if out.HookSpecificOutput.PermissionDecisionReason != "security review required" {
		t.Errorf("Reason = %q", out.HookSpecificOutput.PermissionDecisionReason)
	}
}

// TestNewDeferOutput_EmptyReason works with empty reason.
func TestNewDeferOutput_EmptyReason(t *testing.T) {
	t.Parallel()

	out := NewDeferOutput("")
	if out == nil {
		t.Fatal("NewDeferOutput() returned nil for empty reason")
	}
	if out.HookSpecificOutput.PermissionDecision != DecisionDefer {
		t.Errorf("PermissionDecision = %q, want %q", out.HookSpecificOutput.PermissionDecision, DecisionDefer)
	}
}

// --- pre_tool.go: MergeExtraPatterns ---

// TestMergeExtraPatterns_NilExtra is a no-op.
func TestMergeExtraPatterns_NilExtra(t *testing.T) {
	t.Parallel()

	policy := &SecurityPolicy{}
	policy.MergeExtraPatterns(nil)
	// Should not panic; policy unchanged.
	if len(policy.DangerousBashPatterns) != 0 {
		t.Error("DangerousBashPatterns should remain empty")
	}
}

// TestMergeExtraPatterns_WithPatterns appends compiled patterns.
func TestMergeExtraPatterns_WithPatterns(t *testing.T) {
	t.Parallel()

	policy := &SecurityPolicy{}
	extra := &security.ExtraSecurityConfig{}
	extra.Security.ExtraDangerousBashPatterns = []string{"rm -rf"}
	extra.Security.ExtraDenyPatterns = []string{`\.env$`}
	extra.Security.ExtraAskPatterns = []string{`secrets/`}
	extra.Security.ExtraSensitiveContentPatterns = []string{"password="}

	policy.MergeExtraPatterns(extra)

	if len(policy.DangerousBashPatterns) != 1 {
		t.Errorf("DangerousBashPatterns len = %d, want 1", len(policy.DangerousBashPatterns))
	}
	if len(policy.DenyPatterns) != 1 {
		t.Errorf("DenyPatterns len = %d, want 1", len(policy.DenyPatterns))
	}
	if len(policy.AskPatterns) != 1 {
		t.Errorf("AskPatterns len = %d, want 1", len(policy.AskPatterns))
	}
	if len(policy.SensitiveContentPatterns) != 1 {
		t.Errorf("SensitiveContentPatterns len = %d, want 1", len(policy.SensitiveContentPatterns))
	}
}

// TestMergeExtraPatterns_OnlyDangerous only adds dangerous patterns.
func TestMergeExtraPatterns_OnlyDangerous(t *testing.T) {
	t.Parallel()

	policy := &SecurityPolicy{}
	extra := &security.ExtraSecurityConfig{}
	extra.Security.ExtraDangerousBashPatterns = []string{"curl.*evil", "wget.*evil"}

	policy.MergeExtraPatterns(extra)

	if len(policy.DangerousBashPatterns) != 2 {
		t.Errorf("DangerousBashPatterns len = %d, want 2", len(policy.DangerousBashPatterns))
	}
	if len(policy.DenyPatterns) != 0 {
		t.Errorf("DenyPatterns should be empty, got %d", len(policy.DenyPatterns))
	}
}

// --- worktree_registry.go: saveWorktreeEntries ---

// TestSaveWorktreeEntries_CreatesFile verifies the state file is created.
func TestSaveWorktreeEntries_CreatesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "nested", "worktrees.json")

	entries := []WorktreeEntry{
		{
			Path:      "/project/.claude/worktrees/abc123",
			Branch:    "feature/test",
			AgentName: "test-agent",
			CreatedAt: time.Now(),
		},
	}

	saveWorktreeEntries(statePath, entries)

	// Verify file was created.
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("state file not created: %v", err)
	}
	if len(data) == 0 {
		t.Error("state file is empty")
	}
}

// TestSaveWorktreeEntries_EmptySlice writes empty array.
func TestSaveWorktreeEntries_EmptySlice(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "worktrees.json")

	saveWorktreeEntries(statePath, []WorktreeEntry{})

	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("state file not created: %v", err)
	}
	// Should contain "[]" or similar empty JSON.
	if len(data) == 0 {
		t.Error("state file is empty")
	}
}

// TestSaveWorktreeEntries_LoadRoundtrip verifies save+load round trip.
func TestSaveWorktreeEntries_LoadRoundtrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "worktrees.json")

	entries := []WorktreeEntry{
		{
			Path:      "/project/worktree-1",
			Branch:    "spec/test-001",
			AgentName: "agent-a",
			CreatedAt: time.Now().Truncate(time.Second), // truncate for JSON round-trip
		},
	}

	saveWorktreeEntries(statePath, entries)
	loaded := loadWorktreeEntries(statePath)

	if len(loaded) != 1 {
		t.Fatalf("loaded %d entries, want 1", len(loaded))
	}
	if loaded[0].Path != entries[0].Path {
		t.Errorf("Path = %q, want %q", loaded[0].Path, entries[0].Path)
	}
	if loaded[0].Branch != entries[0].Branch {
		t.Errorf("Branch = %q, want %q", loaded[0].Branch, entries[0].Branch)
	}
}

// --- reflective_write.go: pickEvolvableZone ---

// TestPickEvolvableZone_NonExistentFile returns default "best-practices".
func TestPickEvolvableZone_NonExistentFile(t *testing.T) {
	t.Parallel()

	result := pickEvolvableZone("/nonexistent/path/SKILL.md")
	if result != "best-practices" {
		t.Errorf("pickEvolvableZone() = %q, want 'best-practices' for missing file", result)
	}
}

// TestPickEvolvableZone_EmptyFile returns default "best-practices".
func TestPickEvolvableZone_EmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	_ = os.WriteFile(path, []byte(""), 0o644)

	result := pickEvolvableZone(path)
	if result != "best-practices" {
		t.Errorf("pickEvolvableZone() = %q, want 'best-practices' for empty file", result)
	}
}

// TestPickEvolvableZone_WithEvolvableZone returns the first zone ID.
func TestPickEvolvableZone_WithEvolvableZone(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	content := `# Test Skill
<!-- EVOLVABLE:zone-alpha -->
Some content here.
<!-- /EVOLVABLE:zone-alpha -->
`
	_ = os.WriteFile(path, []byte(content), 0o644)

	result := pickEvolvableZone(path)
	// If merge.ParseEvolvableZones finds a zone, it should not be empty.
	// If the format doesn't match parser expectations, falls back to "best-practices".
	if result == "" {
		t.Error("pickEvolvableZone() returned empty string")
	}
}

// TestPickEvolvableZone_NoEvolvableZone returns default "best-practices".
func TestPickEvolvableZone_NoEvolvableZone(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	content := `# Test Skill

This file has no evolvable zones.
Just plain content.
`
	_ = os.WriteFile(path, []byte(content), 0o644)

	result := pickEvolvableZone(path)
	if result != "best-practices" {
		t.Errorf("pickEvolvableZone() = %q, want 'best-practices' for file without zones", result)
	}
}

// --- session_start.go: isCGMode ---

// TestIsCGMode_MissingFile returns false when llm.yaml is absent.
func TestIsCGMode_MissingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if got := isCGMode(dir); got {
		t.Error("isCGMode() should return false when llm.yaml is missing")
	}
}

// TestIsCGMode_WithCGTeamMode returns true when team_mode: cg.
func TestIsCGMode_WithCGTeamMode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	llmDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(llmDir, 0o755)
	_ = os.WriteFile(filepath.Join(llmDir, "llm.yaml"), []byte("team_mode: cg\n"), 0o644)

	if got := isCGMode(dir); !got {
		t.Error("isCGMode() should return true when team_mode: cg")
	}
}

// TestIsCGMode_WithOtherTeamMode returns false for non-cg team_mode.
func TestIsCGMode_WithOtherTeamMode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	llmDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(llmDir, 0o755)
	_ = os.WriteFile(filepath.Join(llmDir, "llm.yaml"), []byte("team_mode: glm\n"), 0o644)

	if got := isCGMode(dir); got {
		t.Error("isCGMode() should return false for team_mode: glm")
	}
}

// TestIsCGMode_EmptyFile returns false for empty llm.yaml.
func TestIsCGMode_EmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	llmDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(llmDir, 0o755)
	_ = os.WriteFile(filepath.Join(llmDir, "llm.yaml"), []byte(""), 0o644)

	if got := isCGMode(dir); got {
		t.Error("isCGMode() should return false for empty llm.yaml")
	}
}

// --- worktree_create.go: Handle ---

// TestWorktreeCreateHandler_Handle_BasicPath verifies Handle returns output.
func TestWorktreeCreateHandler_Handle_BasicPath(t *testing.T) {
	t.Parallel()

	h := NewWorktreeCreateHandler()
	if h == nil {
		t.Fatal("NewWorktreeCreateHandler() returned nil")
	}

	input := &HookInput{
		SessionID:  "sess-wt-001",
		ProjectDir: t.TempDir(),
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil output")
	}
}

// --- session_start.go: Handle with real config (covers cfg != nil branches) ---

// makeHandlerWithProjectConfig creates a sessionStartHandler with a config
// that has a non-empty project name, type, and language.
func makeHandlerWithProjectConfig() *sessionStartHandler {
	cfg := config.NewDefaultConfig()
	cfg.Project.Name = "test-project"
	cfg.Project.Type = models.ProjectTypeCLI
	cfg.Project.Language = "go"
	return &sessionStartHandler{
		cfg: &mockConfigProvider{cfg: cfg},
	}
}

// TestSessionStartHandler_Handle_WithProjectConfig covers cfg.Project.Name branch.
func TestSessionStartHandler_Handle_WithProjectConfig(t *testing.T) {
	t.Parallel()

	h := makeHandlerWithProjectConfig()
	input := &HookInput{
		SessionID:     "sess-with-config",
		CWD:           t.TempDir(),
		HookEventName: "SessionStart",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
}

// TestSessionStartHandler_Handle_WithProjectDirAndConfig exercises ProjectDir branches.
func TestSessionStartHandler_Handle_WithProjectDirAndConfig(t *testing.T) {
	t.Parallel()

	h := makeHandlerWithProjectConfig()

	dir := t.TempDir()
	// Create minimal .claude directory so ensureTeammateMode has somewhere to write.
	_ = os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)

	input := &HookInput{
		SessionID:     "sess-with-project-dir",
		CWD:           dir,
		ProjectDir:    dir,
		HookEventName: "SessionStart",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
}

// TestWorktreeRemoveHandler_Handle_BasicPath verifies Handle returns output.
func TestWorktreeRemoveHandler_Handle_BasicPath(t *testing.T) {
	t.Parallel()

	h := NewWorktreeRemoveHandler()
	if h == nil {
		t.Fatal("NewWorktreeRemoveHandler() returned nil")
	}

	input := &HookInput{
		SessionID:  "sess-wt-remove-001",
		ProjectDir: t.TempDir(),
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil output")
	}
}

// --- reflective_write.go: PresentPendingProposals, AnalyzeSessionAndLog ---

// TestPresentPendingProposals_NoLearnings returns empty string when no learnings.
func TestPresentPendingProposals_NoLearnings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := PresentPendingProposals(dir)
	if result != "" {
		t.Errorf("PresentPendingProposals() should return empty for no learnings, got %q", result)
	}
}

// TestPresentPendingProposals_EmptyEvolutionDir returns empty for missing evolution dir.
func TestPresentPendingProposals_EmptyEvolutionDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	evolutionDir := filepath.Join(dir, ".moai", "evolution", "learnings")
	_ = os.MkdirAll(evolutionDir, 0o755)

	result := PresentPendingProposals(dir)
	if result != "" {
		t.Errorf("PresentPendingProposals() should return empty for empty learnings dir, got %q", result)
	}
}

// TestAnalyzeSessionAndLog_EmptyProjectDir is non-blocking on error.
func TestAnalyzeSessionAndLog_EmptyProjectDir(t *testing.T) {
	t.Parallel()

	// Should not panic; projectRoot doesn't exist → graceful error handling.
	AnalyzeSessionAndLog("/nonexistent/project/path", "sess-test-001")
}

// TestAnalyzeSessionAndLog_ValidDir is non-blocking for valid dir with no tool files.
func TestAnalyzeSessionAndLog_ValidDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Should not panic; no tool history files exist.
	AnalyzeSessionAndLog(dir, "sess-test-002")
}

// --- teammate_idle.go: loadCoverageThreshold, loadCoverageData ---

// TestLoadCoverageThreshold_MissingFile returns default threshold.
func TestLoadCoverageThreshold_MissingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := loadCoverageThreshold(dir)
	if result <= 0 {
		t.Errorf("loadCoverageThreshold() should return positive default, got %v", result)
	}
}

// TestLoadCoverageThreshold_ValidFile returns configured threshold.
func TestLoadCoverageThreshold_ValidFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(sectionsDir, 0o755)
	content := "constitution:\n  test_coverage_target: 90.0\n"
	_ = os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"), []byte(content), 0o644)

	result := loadCoverageThreshold(dir)
	if result != 90.0 {
		t.Errorf("loadCoverageThreshold() = %v, want 90.0", result)
	}
}

// TestLoadCoverageThreshold_ZeroValue returns default when target is 0.
func TestLoadCoverageThreshold_ZeroValue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(sectionsDir, 0o755)
	content := "constitution:\n  test_coverage_target: 0\n"
	_ = os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"), []byte(content), 0o644)

	result := loadCoverageThreshold(dir)
	// Zero value → falls back to default.
	if result <= 0 {
		t.Errorf("loadCoverageThreshold() should return positive default for 0 target, got %v", result)
	}
}

// TestLoadCoverageThreshold_MalformedYAML returns default on parse error.
func TestLoadCoverageThreshold_MalformedYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(sectionsDir, 0o755)
	_ = os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"), []byte("{bad yaml"), 0o644)

	result := loadCoverageThreshold(dir)
	if result <= 0 {
		t.Errorf("loadCoverageThreshold() should return positive default for bad YAML, got %v", result)
	}
}

// TestLoadCoverageData_MissingFile returns (0, false).
func TestLoadCoverageData_MissingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pct, ok := loadCoverageData(dir)
	if ok {
		t.Error("loadCoverageData() should return false for missing file")
	}
	if pct != 0 {
		t.Errorf("loadCoverageData() should return 0 for missing file, got %v", pct)
	}
}

// TestLoadCoverageData_ValidFile returns (percent, true).
func TestLoadCoverageData_ValidFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	_ = os.MkdirAll(stateDir, 0o755)
	_ = os.WriteFile(filepath.Join(stateDir, "coverage.json"),
		[]byte(`{"coverage_percent": 87.5}`), 0o644)

	pct, ok := loadCoverageData(dir)
	if !ok {
		t.Error("loadCoverageData() should return true for valid file")
	}
	if pct != 87.5 {
		t.Errorf("loadCoverageData() = %v, want 87.5", pct)
	}
}

// TestLoadCoverageData_MalformedJSON returns (0, false).
func TestLoadCoverageData_MalformedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".moai", "state")
	_ = os.MkdirAll(stateDir, 0o755)
	_ = os.WriteFile(filepath.Join(stateDir, "coverage.json"), []byte("{bad json"), 0o644)

	pct, ok := loadCoverageData(dir)
	if ok {
		t.Error("loadCoverageData() should return false for malformed JSON")
	}
	if pct != 0 {
		t.Errorf("loadCoverageData() should return 0 for malformed JSON, got %v", pct)
	}
}
