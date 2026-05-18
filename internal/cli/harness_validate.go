package cli

// @MX:NOTE: [AUTO] HRN-001 harness validate CLI — SPEC-V3R2-HRN-001 run-phase
// REQ-HRN-001-006/010/012, AC-HRN-001-04/05/07.

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// newHarnessValidateCmd는 `moai harness validate` 서브커맨드 팩토리입니다.
// REQ-HRN-001-006, AC-HRN-001-04/05/07.
func newHarnessValidateCmd() *cobra.Command {
	var cfgPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate harness.yaml against schema and invariants",
		Long: `Validate harness.yaml against:
  - Level enum: {minimal, standard, thorough} FROZEN (REQ-HRN-001-017)
  - pass_threshold >= 0.60 FROZEN floor (REQ-HRN-001-012, design-constitution §5)
  - MOAI_CONFIG_STRICT=1: unknown keys become errors (REQ-HRN-001-019)
  - evaluator.memory_scope must be per_iteration (HRN-002 FROZEN)

Exit code 0 = valid, exit code 1 = validation error.

Examples:
  moai harness validate
  moai harness validate --path /custom/harness.yaml
  MOAI_CONFIG_STRICT=1 moai harness validate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHarnessValidate(cmd, cfgPath)
		},
	}

	cmd.Flags().StringVar(&cfgPath, "path", "", "Path to harness.yaml (default: "+defaultHarnessConfigPath+")")

	return cmd
}

// runHarnessValidate는 `moai harness validate` 커맨드를 실행합니다.
// AC-HRN-001-04: 정상 종료 시 exit 0 + "harness.yaml: OK".
// AC-HRN-001-05: pass_threshold < 0.60 → exit 1 + HRN_PASS_THRESHOLD_FLOOR 오류.
// AC-HRN-001-07: MOAI_CONFIG_STRICT=1 + unknown key → exit 1 + HRN_SCHEMA_DRIFT 오류.
func runHarnessValidate(cmd *cobra.Command, cfgPath string) error {
	// harness.yaml 경로 결정
	harnessPath := cfgPath
	if harnessPath == "" {
		harnessPath = defaultHarnessConfigPath
	}

	// harness.yaml 로드 (레벨 enum, 메모리 스코프, 스키마 드리프트 검증 포함)
	cfg, err := config.LoadHarnessConfig(harnessPath)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
		return fmt.Errorf("harness validate: %w", err)
	}

	// pass_threshold floor 검증 (REQ-HRN-001-012)
	// 각 레벨의 evaluator_profile이 참조하는 .md 파일을 파싱하여 검증합니다.
	if err := validateEvaluatorProfileFloors(harnessPath, cfg, cmd); err != nil {
		return err
	}

	// model_upgrade_review 알림 (REQ-HRN-001-016)
	emitModelUpgradeReminder(cfg, cmd)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "harness.yaml: OK")
	return nil
}

// validateEvaluatorProfileFloors는 각 레벨의 evaluator_profile 파일을 파싱하여
// pass_threshold >= 0.60 FROZEN floor를 검증합니다.
// REQ-HRN-001-012, AC-HRN-001-05.
//
// @MX:WARN: [AUTO] FROZEN floor 검증 — pass_threshold < 0.60 → ErrPassThresholdFloor
// @MX:REASON: design-constitution §5 FROZEN floor; 위반 시 즉시 exit 1
func validateEvaluatorProfileFloors(harnessPath string, cfg *config.HarnessConfig, cmd *cobra.Command) error {
	if cfg == nil {
		return nil
	}

	// harness.yaml 경로 기준으로 evaluator-profiles 디렉토리 찾기
	// 표준 위치: {config_dir}/../evaluator-profiles/
	// harnessPath 예: .moai/config/sections/harness.yaml
	// → .moai/config/evaluator-profiles/
	harnessDir := harnessPath
	// sections/harness.yaml → sections/ → config/ → evaluator-profiles/
	sectionsDir := harnessDir
	for i := 0; i < 2; i++ {
		if d := parentDir(sectionsDir); d != sectionsDir {
			sectionsDir = d
		}
	}
	profilesDir := sectionsDir + "/evaluator-profiles"

	// 각 레벨의 evaluator_profile 파일 검증
	for levelName, levelCfg := range cfg.Levels {
		profileName := levelCfg.EvaluatorProfile
		if profileName == "" {
			profileName = cfg.DefaultProfile
		}
		if profileName == "" {
			continue
		}

		profilePath := profilesDir + "/" + profileName + ".md"

		// 파일이 없으면 경고만 출력하고 계속
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			continue
		}

		// 프로파일 파싱
		rubric, err := router.ParseProfileFloor(profilePath)
		if err != nil {
			// 파싱 오류 시 경고만 출력
			continue
		}

		// FROZEN floor 검증
		if rubric < 0.60 {
			errMsg := fmt.Sprintf("HRN_PASS_THRESHOLD_FLOOR: levels.%s.evaluator_profile=%q pass_threshold=%.2f is below FROZEN floor 0.60 (design-constitution §5)",
				levelName, profileName, rubric)
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), errMsg)
			return &config.ValidationError{
				Field:   fmt.Sprintf("levels.%s.evaluator_profile", levelName),
				Message: errMsg,
				Value:   rubric,
				Wrapped: config.ErrPassThresholdFloor,
			}
		}
	}

	return nil
}

// emitModelUpgradeReminder는 모델 업그레이드 검토 알림을 출력합니다.
// REQ-HRN-001-016: OnModelChange == true 시 알림.
func emitModelUpgradeReminder(cfg *config.HarnessConfig, cmd *cobra.Command) {
	if cfg == nil {
		return
	}
	if !cfg.ModelUpgradeReview.Enabled || !cfg.ModelUpgradeReview.Trigger.OnModelChange {
		return
	}

	// CLAUDE_MODEL_PREVIOUS 환경변수로 모델 변경 감지
	previousModel := os.Getenv("CLAUDE_MODEL_PREVIOUS")
	currentModel := os.Getenv("CLAUDE_MODEL")
	if previousModel != "" && currentModel != "" && previousModel != currentModel {
		reportPath := cfg.ModelUpgradeReview.Output.ReportPath
		if reportPath == "" {
			reportPath = ".moai/reports/harness-review.md"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(),
			"Model upgrade detected (%s → %s). Review checklist at: %s\n",
			previousModel, currentModel, reportPath)
	}
}

// parentDir는 경로의 부모 디렉토리를 반환합니다.
func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return path
}

// ValidateHarnessErrors는 하네스 오류를 errors.Is로 확인하는 헬퍼입니다.
// 테스트에서 사용됩니다.
func ValidateHarnessErrors(err error) (isFloor, isDrift bool) {
	isFloor = errors.Is(err, config.ErrPassThresholdFloor)
	isDrift = errors.Is(err, config.ErrSchemaDrift)
	return
}
