package hook

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/workflow"
)

// specFilePattern은 SPEC 디렉토리 내 spec.md 파일을 탐색하는 glob 패턴이다.
const specFilePattern = ".moai/specs/*/spec.md"

// userPromptSubmitHandler는 UserPromptSubmit 이벤트를 처리한다.
// 사용자 프롬프트 제출 시 세션 타이틀을 자동으로 생성하여 Claude Code에 반환한다.
type userPromptSubmitHandler struct {
	cfg ConfigProvider
}

// NewUserPromptSubmitHandler는 UserPromptSubmit 이벤트 핸들러를 생성한다.
func NewUserPromptSubmitHandler(cfg ConfigProvider) Handler {
	return &userPromptSubmitHandler{cfg: cfg}
}

// EventType은 EventUserPromptSubmit을 반환한다.
func (h *userPromptSubmitHandler) EventType() EventType {
	return EventUserPromptSubmit
}

// workflowKeywords are prompt keywords that indicate an active MoAI workflow context.
var workflowKeywords = []string{"loop", "run", "plan"}

// detectWorkflowContext checks whether the prompt contains any workflow keywords
// and returns a non-empty additionalContext string if a match is found.
func detectWorkflowContext(prompt string) string {
	lower := strings.ToLower(prompt)
	for _, kw := range workflowKeywords {
		if strings.Contains(lower, kw) {
			return "workflow keyword '" + kw + "' detected — MoAI workflow context may be active"
		}
	}
	return ""
}

// Handle은 UserPromptSubmit 이벤트를 처리한다.
// CWD에서 활성 SPEC을 탐지하여 세션 타이틀을 생성하고 hookSpecificOutput으로 반환한다.
// 에러가 발생해도 빈 타이틀을 반환하며 프롬프트를 차단하지 않는다.
func (h *userPromptSubmitHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// 프롬프트 로깅 (감사 목적, 100자 초과 시 잘라냄)
	prompt := input.Prompt
	preview := prompt
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	slog.Info("user prompt submitted",
		"session_id", input.SessionID,
		"prompt_preview", preview,
	)

	// 세션 타이틀 생성 (에러는 무시하고 빈 타이틀로 폴백)
	title := h.buildSessionTitle(input.CWD)

	// 워크플로우 컨텍스트 감지
	additionalCtx := detectWorkflowContext(prompt)

	// 타이틀과 워크플로우 컨텍스트 모두 없으면 빈 응답
	if title == "" && additionalCtx == "" {
		return &HookOutput{}, nil
	}

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			SessionTitle:      title,
			AdditionalContext: additionalCtx,
		},
	}, nil
}

// buildSessionTitle은 CWD 기반으로 세션 타이틀을 생성한다.
// SPEC 발견 시: "SPEC-ID: 제목"
// SPEC 없을 시: "프로젝트명 / 브랜치명"
// 에러 발생 시: 빈 문자열 반환
func (h *userPromptSubmitHandler) buildSessionTitle(cwd string) string {
	if cwd == "" {
		return ""
	}

	// 활성 SPEC 탐지 시도
	if title := detectActiveSpec(cwd); title != "" {
		return title
	}

	// SPEC 없을 시 프로젝트명/브랜치명 조합
	return buildProjectBranchTitle(cwd)
}

// detectActiveSpec은 cwd/.moai/specs/*/spec.md를 탐색하여
// 가장 최근에 수정된 SPEC의 타이틀을 반환한다.
// SPEC이 없거나 읽기 실패 시 빈 문자열을 반환한다.
func detectActiveSpec(cwd string) string {
	pattern := filepath.Join(cwd, specFilePattern)
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}

	// 가장 최근 수정된 spec.md를 우선 선택
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

	// SPEC ID를 디렉토리 이름에서 추출
	specDirName := filepath.Base(filepath.Dir(latestMatch))
	specID := workflow.SpecIDPattern.FindString(specDirName)
	if specID == "" {
		return ""
	}

	// spec.md의 첫 번째 헤딩(# 으로 시작하는 줄)을 읽는다
	heading := readFirstHeading(latestMatch)
	if heading == "" {
		return specID
	}

	return fmt.Sprintf("%s: %s", specID, heading)
}

// readFirstHeading은 markdown 파일에서 첫 번째 # 헤딩 텍스트를 반환한다.
func readFirstHeading(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// buildProjectBranchTitle은 "프로젝트명 / 브랜치명" 형식의 타이틀을 생성한다.
// git 브랜치 조회 실패 시 "프로젝트명 / unknown"을 반환한다.
func buildProjectBranchTitle(cwd string) string {
	projectName := filepath.Base(cwd)
	if projectName == "" || projectName == "." {
		projectName = "unknown"
	}

	branch := getGitBranch(cwd)
	if branch == "" {
		branch = "unknown"
	}

	return fmt.Sprintf("%s / %s", projectName, branch)
}

// getGitBranch는 git rev-parse --abbrev-ref HEAD로 현재 브랜치명을 조회한다.
// 조회 실패 시 빈 문자열을 반환한다.
func getGitBranch(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cwd

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}
