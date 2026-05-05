package hook

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// subagentStartHandler processes SubagentStart events.
// It logs subagent startup for session tracking and optionally injects project context.
type subagentStartHandler struct {
	cfg        ConfigProvider
	projectDir string
}

// NewSubagentStartHandler creates a new SubagentStart event handler without config.
// This constructor preserves backward compatibility; no additionalContext is injected.
func NewSubagentStartHandler() Handler {
	return &subagentStartHandler{}
}

// NewSubagentStartHandlerWithConfig creates a new SubagentStart event handler with
// config access for project context injection.
// projectDir is resolved from CLAUDE_PROJECT_DIR env var or os.Getwd() as fallback.
func NewSubagentStartHandlerWithConfig(cfg ConfigProvider) Handler {
	dir := os.Getenv(config.EnvClaudeProjectDir)
	if dir == "" {
		if cwd, err := os.Getwd(); err == nil {
			dir = cwd
		}
	}
	return &subagentStartHandler{cfg: cfg, projectDir: dir}
}

// EventType returns EventSubagentStart.
func (h *subagentStartHandler) EventType() EventType {
	return EventSubagentStart
}

// Handle processes a SubagentStart event.
// If config is available, it injects project context via additionalContext.
// Errors are non-blocking; an empty output is returned on failure.
func (h *subagentStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("subagent started",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_transcript_path", input.AgentTranscriptPath,
	)

	contextStr := h.buildContext(input)
	if contextStr == "" {
		return &HookOutput{}, nil
	}

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     string(EventSubagentStart),
			AdditionalContext: contextStr,
		},
	}, nil
}

// buildContext constructs a concise project context string (under 200 chars).
// Returns empty string if no useful context is available.
func (h *subagentStartHandler) buildContext(input *HookInput) string {
	if h.cfg == nil {
		return ""
	}

	cfg := h.cfg.Get()
	if cfg == nil {
		return ""
	}

	var parts []string

	if cfg.Project.Name != "" {
		parts = append(parts, "project:"+cfg.Project.Name)
	}
	if string(cfg.Project.Type) != "" {
		parts = append(parts, "type:"+string(cfg.Project.Type))
	}
	if cfg.Project.Language != "" {
		parts = append(parts, "lang:"+cfg.Project.Language)
	}

	// Resolve project directory from input CWD, handler field, or env
	dir := input.CWD
	if dir == "" {
		dir = h.projectDir
	}

	if spec := h.detectActiveSpec(dir); spec != "" {
		parts = append(parts, "spec:"+spec)
	}

	if len(parts) == 0 {
		return ""
	}

	result := strings.Join(parts, " | ")
	// Truncate to stay under 200 chars
	if len(result) > 199 {
		result = result[:199]
	}
	return result
}

// detectActiveSpec returns the SPEC ID of the most recently modified spec.md under dir.
// Returns empty string if no SPEC is found.
func (h *subagentStartHandler) detectActiveSpec(dir string) string {
	if dir == "" {
		return ""
	}

	pattern := filepath.Join(dir, specFilePattern)
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}

	var latestMatch string
	var latestModTime int64
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		mt := info.ModTime().UnixNano()
		if mt > latestModTime {
			latestModTime = mt
			latestMatch = match
		}
	}
	if latestMatch == "" {
		return ""
	}

	// Extract SPEC ID from directory name (e.g., "SPEC-FOO-001")
	specDirName := filepath.Base(filepath.Dir(latestMatch))
	return specDirName
}

// ============================================================
// agentStartHandler: SubagentStart retired-rejection guard
// REQ-RA-004, REQ-RA-007, REQ-RA-008, REQ-RA-009, REQ-RA-012
// ============================================================

// agentFrontmatter는 에이전트 파일 YAML frontmatter에서 파싱하는 필드를 정의한다.
// retired: true 가 있을 때 spawn을 거부한다 (REQ-RA-007).
type agentFrontmatter struct {
	Retired           bool   `yaml:"retired"`
	RetiredReplacement string `yaml:"retired_replacement"`
	RetiredParamHint  string `yaml:"retired_param_hint"`
}

// @MX:NOTE: [AUTO] agentStartHandler는 SPEC-V3R3-RETIRED-AGENT-001 5-layer defect chain
// Layer 1 (retired stub frontmatter invalid) 차단의 핵심 진입점이다.
// retired:true frontmatter detect 시 block decision JSON + exit code 2 반환으로
// Claude Code Agent runtime의 worktree allocation 전 spawn을 거부한다 (≤500ms 응답 budget).
//
// agentStartHandler는 SubagentStart 이벤트에서 retired 에이전트 spawn을 거부한다.
// REQ-RA-004: SubagentStart handler 신규 등록
type agentStartHandler struct{}

// NewAgentStartHandler는 retired-rejection guard가 포함된
// SubagentStart 이벤트 핸들러를 생성한다.
// Option A: subagent_start.go에 통합 (plan.md agent_start.go 명목과 다름).
// 결정 사유: file 중복 없이 clean integration, EventSubagentStart는 이미 등록됨.
func NewAgentStartHandler() Handler {
	return &agentStartHandler{}
}

// EventType은 EventSubagentStart를 반환한다.
func (h *agentStartHandler) EventType() EventType {
	return EventSubagentStart
}

// Handle은 SubagentStart 이벤트를 처리한다.
// AgentName이 없거나 파일을 찾을 수 없으면 allow (REQ-RA-008).
// retired: true frontmatter가 있으면 block decision 반환 (REQ-RA-007).
// 성능: 단일 파일 read + YAML parse, 네트워크 I/O 없음 (REQ-RA-012 ≤500ms).
func (h *agentStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	agentName := input.AgentName
	if agentName == "" {
		// AgentName 미제공 시 pass-through (REQ-RA-008)
		return &HookOutput{}, nil
	}

	// path traversal 방지: 슬래시, 점 두 개 포함 시 거부
	if strings.Contains(agentName, "/") || strings.Contains(agentName, "..") {
		slog.Warn("agentStartHandler: 잠재적 경로 탐색 에이전트 이름 거부",
			"agent_name", agentName)
		return &HookOutput{}, nil
	}

	// projectDir 결정: CWD 우선, 환경변수 fallback
	projectDir := input.CWD
	if projectDir == "" {
		projectDir = os.Getenv(config.EnvClaudeProjectDir)
	}
	if projectDir == "" {
		if cwd, err := os.Getwd(); err == nil {
			projectDir = cwd
		}
	}

	fm, found, err := h.loadAgentFrontmatter(projectDir, agentName)
	if err != nil {
		// 파싱 오류 시 fail-safe: allow
		slog.Warn("agentStartHandler: frontmatter 파싱 오류, allow로 처리",
			"agent", agentName, "error", err)
		return &HookOutput{}, nil
	}
	if !found {
		// 파일 없음 → 비-MoAI 에이전트 bypass (REQ-RA-008)
		return &HookOutput{}, nil
	}

	if !fm.Retired {
		// 활성 에이전트: allow (REQ-RA-008)
		return &HookOutput{}, nil
	}

	// retired: true → block (REQ-RA-007)
	reason := fmt.Sprintf(
		"agent %s retired (SPEC-V3R3-RETIRED-AGENT-001), use %s",
		agentName, fm.RetiredReplacement,
	)
	if fm.RetiredParamHint != "" {
		reason += " with " + fm.RetiredParamHint
	}

	slog.Info("agentStartHandler: retired 에이전트 spawn 차단",
		"agent", agentName,
		"replacement", fm.RetiredReplacement,
	)

	return &HookOutput{
		Decision: DecisionBlock,
		Reason:   reason,
	}, nil
}

// loadAgentFrontmatter는 에이전트 파일을 찾아 YAML frontmatter를 파싱한다.
// 탐색 순서: .claude/agents/moai/<name>.md → .claude/agents/<name>.md
// found=false이면 파일이 존재하지 않는 것 (REQ-RA-008 bypass).
func (h *agentStartHandler) loadAgentFrontmatter(projectDir, agentName string) (*agentFrontmatter, bool, error) {
	candidates := []string{
		filepath.Join(projectDir, ".claude", "agents", "moai", agentName+".md"),
		filepath.Join(projectDir, ".claude", "agents", agentName+".md"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			// 읽기 오류는 found=true로 처리하되 parse error 반환
			return nil, true, fmt.Errorf("에이전트 파일 읽기 실패 %s: %w", path, err)
		}

		fm, err := parseAgentFrontmatter(data)
		if err != nil {
			return nil, true, fmt.Errorf("frontmatter 파싱 실패 %s: %w", path, err)
		}
		return fm, true, nil
	}

	return nil, false, nil
}

// parseAgentFrontmatter는 Markdown 파일에서 YAML frontmatter(--- 구분자 사이)를 추출하고 파싱한다.
// frontmatter가 없으면 빈 구조체를 반환한다 (오류 없음).
func parseAgentFrontmatter(data []byte) (*agentFrontmatter, error) {
	fm := &agentFrontmatter{}

	// --- 로 시작하는 frontmatter 추출
	if !bytes.HasPrefix(data, []byte("---")) {
		// frontmatter 없음 → 빈 구조체 반환 (활성 에이전트로 처리)
		return fm, nil
	}

	// 첫 번째 --- 이후 두 번째 --- 찾기
	rest := data[3:]
	// 줄바꿈 건너뜀
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 0 && rest[0] == '\r' && len(rest) > 1 && rest[1] == '\n' {
		rest = rest[2:]
	}

	yamlContent, _, ok := bytes.Cut(rest, []byte("\n---"))
	if !ok {
		// 닫는 --- 없음 → frontmatter 없는 것으로 처리
		return fm, nil
	}
	if err := yaml.Unmarshal(yamlContent, fm); err != nil {
		return nil, fmt.Errorf("YAML 언마샬 실패: %w", err)
	}

	return fm, nil
}
