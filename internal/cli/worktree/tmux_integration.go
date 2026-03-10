package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/tmux"
)

// TmuxIntegration은 SPEC-WORKTREE-002의 R5: Tmux Integration 요구사항을 구현합니다.
// worktree 생성 후 자동으로 tmux 세션을 생성하고, 필요한 환경 변수를 주입합니다.
//
// @MX:NOTE: SPEC-WORKTREE-002 R5 구현 - tmux 세션 자동 생성 및 환경 변수 주입
// @MX:SPEC: SPEC-WORKTREE-002

// TmuxSessionConfig는 tmux 세션 생성을 위한 설정입니다.
type TmuxSessionConfig struct {
	// ProjectName은 프로젝트 이름입니다 (예: "moai-adk-go")
	ProjectName string

	// SpecID는 SPEC 식별자입니다 (예: "SPEC-WORKTREE-002")
	SpecID string

	// WorktreePath는 worktree의 절대 경로입니다
	WorktreePath string

	// ActiveMode는 현재 LLM 모드입니다 (cc, glm, cg)
	ActiveMode string

	// GLMEnvVars는 GLM/CG 모드에서 주입할 환경 변수입니다.
	// CC 모드에서는 비어있어야 합니다.
	GLMEnvVars map[string]string
}

// CreateTmuxSession은 worktree를 위한 tmux 세션을 생성합니다.
//
// R5.1: 세션 이름 패턴: moai-{ProjectName}-{SPEC-ID}
// R5.2-5.3: GLM/CG 모드에서는 환경 변수 주입, CC 모드에서는 주입 없음
// R5.4: 세션 생성 후 worktree로 cd하고 /moai run 명령어 실행
//
// @MX:ANCHOR: worktree 기반 개발 워크플로우의 핵심 진입점
// @MX:REASON: tmux 세션 자동화는 SPEC-WORKTREE-002의 핵심 기능으로, 여러 클라이언트에서 호출됨
// @MX:SPEC: SPEC-WORKTREE-002
func CreateTmuxSession(ctx context.Context, cfg *TmuxSessionConfig, tmuxMgr tmux.SessionManager) error {
	if cfg == nil {
		return fmt.Errorf("tmux session config is required")
	}

	if tmuxMgr == nil {
		return fmt.Errorf("tmux manager is required")
	}

	// R5.1: 세션 이름 생성
	sessionName := GenerateTmuxSessionName(cfg.ProjectName, cfg.SpecID)

	// R5.4: tmux 세션 생성 (detached 모드)
	sessionCfg := &tmux.SessionConfig{
		Name:       sessionName,
		MaxVisible: 1, // 단일 pane 사용
		Panes: []tmux.PaneConfig{
			{
				SpecID: cfg.SpecID,
				Command: buildTmuxInitialCommand(cfg),
			},
		},
	}

	result, err := tmuxMgr.Create(ctx, sessionCfg)
	if err != nil {
		return fmt.Errorf("create tmux session: %w", err)
	}

	// R5.2-5.3: GLM/CG 모드에서 환경 변수 주입
	if cfg.ActiveMode == "glm" || cfg.ActiveMode == "cg" {
		if len(cfg.GLMEnvVars) > 0 {
			if err := tmuxMgr.InjectEnv(ctx, cfg.GLMEnvVars); err != nil {
				return fmt.Errorf("inject GLM env: %w", err)
			}
		}
	}

	// 로그 출력
	fmt.Printf("Tmux session created: %s\n", result.SessionName)
	fmt.Printf("Panes created: %d\n", result.PaneCount)
	fmt.Printf("Attached: %v\n", result.Attached)
	fmt.Printf("Worktree path: %s\n", cfg.WorktreePath)
	fmt.Printf("To attach: tmux attach-session -t %s\n", sessionName)

	return nil
}

// buildTmuxInitialCommand는 tmux pane에서 실행할 초기 명령어를构建합니다.
// R5.4: cd to worktree + execute /moai run
func buildTmuxInitialCommand(cfg *TmuxSessionConfig) string {
	// worktree 경로로 cd
	cdCmd := fmt.Sprintf("cd %s", cfg.WorktreePath)

	// /moai run 명령어 실행
	moaiCmd := fmt.Sprintf("/moai run %s", cfg.SpecID)

	// 두 명령어를 연결 (;로 구분)
	return fmt.Sprintf("%s ; %s", cdCmd, moaiCmd)
}

// IsTmuxAvailable은 현재 환경에서 tmux를 사용할 수 있는지 확인합니다.
// R1: Execution Mode Selection Gate에서 tmux 가용성 확인에 사용
//
// @MX:NOTE: SPEC-WORKTREE-002 R1 구현 - tmux 가용성 감지
// @MX:SPEC: SPEC-WORKTREE-002
func IsTmuxAvailable() bool {
	// $TMUX 환경 변수 확인
	return os.Getenv("TMUX") != ""
}

// GetActiveMode는 .moai/config/sections/llm.yaml에서 현재 활성 모드를 읽어옵니다.
// R1.1: active mode 감지
//
// @MX:NOTE: SPEC-WORKTREE-002 R1.1 구현 - LLM 모드 감지
// @MX:SPEC: SPEC-WORKTREE-002
func GetActiveMode(projectRoot string) (string, error) {
	// llm.yaml 파일 경로
	llmConfigPath := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")

	// 파일 존재 확인
	if _, err := os.Stat(llmConfigPath); os.IsNotExist(err) {
		// 기본값: cc (Claude Only)
		return "cc", nil
	}

	// TODO: llm.yaml 파일 파싱 로직 구현
	// 현재는 기본값 반환
	return "cc", nil
}

// BuildTmuxSessionConfig는 worktree 정보로부터 tmux 세션 설정을构建합니다.
//
// @MX:NOTE: SPEC-WORKTREE-002 통합 함수 - tmux 설정 빌더
// @MX:SPEC: SPEC-WORKTREE-002
func BuildTmuxSessionConfig(projectName, specID, worktreePath, projectRoot string) (*TmuxSessionConfig, error) {
	activeMode, err := GetActiveMode(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("get active mode: %w", err)
	}

	cfg := &TmuxSessionConfig{
		ProjectName:  projectName,
		SpecID:       specID,
		WorktreePath: worktreePath,
		ActiveMode:   activeMode,
		GLMEnvVars:   make(map[string]string),
	}

	// R5.2-5.3: GLM/CG 모드에서 환경 변수 설정
	if activeMode == "glm" || activeMode == "cg" {
		// GLM 환경 변수 주입
		// TODO: 실제 GLM 환경 변수 로드 로직 구현
		cfg.GLMEnvVars["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = os.Getenv("ANTHROPIC_DEFAULT_HAIKU_MODEL")
		cfg.GLMEnvVars["ANTHROPIC_DEFAULT_SONNET_MODEL"] = os.Getenv("ANTHROPIC_DEFAULT_SONNET_MODEL")
		cfg.GLMEnvVars["ANTHROPIC_DEFAULT_OPUS_MODEL"] = os.Getenv("ANTHROPIC_DEFAULT_OPUS_MODEL")
	}

	return cfg, nil
}
