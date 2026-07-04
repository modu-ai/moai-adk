package config

// SPEC-WEB-CONSOLE-011 M3: workflow_agents typed 읽기 표면 테스트 (AC-WC11-070).

import (
	"os"
	"path/filepath"
	"testing"
)

// seedWorkflowYAML은 임시 프로젝트 루트에 workflow.yaml을 시드한다.
func seedWorkflowYAML(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestWorkflowAgentsTypedLoad는 workflow_agents 블록의 typed 로더 round-trip을
// 검증한다: 7-purpose 맵 파싱 + model/effort 값 (REQ-WC11-070/071).
func TestWorkflowAgentsTypedLoad(t *testing.T) {
	t.Parallel()
	root := seedWorkflowYAML(t, `workflow:
    default_mode: ""
    workflow_agents:
        read-only-extract: { model: haiku, effort: low }
        mechanical-transform: { model: sonnet, effort: medium }
        synthesize: { model: sonnet, effort: high }
        research: { model: sonnet, effort: high }
        verify-judge: { model: sonnet, effort: xhigh }
        implement: { model: sonnet, effort: xhigh }
        design-architecture: { model: opus, effort: xhigh }
`)

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	wa := cfg.Workflow.WorkflowAgents
	if len(wa) != 7 {
		t.Fatalf("WorkflowAgents length = %d, want 7", len(wa))
	}
	if got := wa["read-only-extract"]; got.Model != "haiku" || got.Effort != "low" {
		t.Errorf("read-only-extract = %+v, want {haiku low}", got)
	}
	if got := wa["design-architecture"]; got.Model != "opus" || got.Effort != "xhigh" {
		t.Errorf("design-architecture = %+v, want {opus xhigh}", got)
	}
}

// TestWorkflowAgentsAbsentBlockZeroValue는 블록 부재 시 zero-value(nil 맵)로
// 무오류 로드됨을 검증한다 (AC-WC11-070 — 2026-07-03 grep 0 실측 상태 호환).
func TestWorkflowAgentsAbsentBlockZeroValue(t *testing.T) {
	t.Parallel()
	root := seedWorkflowYAML(t, "workflow:\n    default_mode: \"\"\n")

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if cfg.Workflow.WorkflowAgents != nil {
		t.Errorf("absent block should load as nil map, got %+v", cfg.Workflow.WorkflowAgents)
	}
}
