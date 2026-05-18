package cli

// @MX:NOTE: [AUTO] HRN-001 harness routing CLI — SPEC-V3R2-HRN-001 run-phase
// @MX:NOTE: [AUTO] newHarnessRouterCmd()는 retired newHarnessCmd()와 별개의 팩토리입니다
// @MX:WARN: [AUTO] root.go 등록 시 TestHarnessRetirement CI 가드 준수 필수
// @MX:REASON: SPEC-V3R4-HARNESS-001 retirement guard — harness route/validate 동사는 허용됨

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// defaultHarnessConfigPath는 기본 harness.yaml 경로입니다.
// internal/cli/harness.go:41의 harnessConfigPath 상수와 동일한 경로를 참조합니다.
const defaultHarnessConfigPath = ".moai/config/sections/harness.yaml"

// harnessRouteJSONOutput은 --json 출력 스키마입니다.
// REQ-HRN-001-011, AC-HRN-001-06.
type harnessRouteJSONOutput struct {
	Level           string           `json:"level"`
	Rationale       router.Rationale `json:"rationale"`
	Effort          string           `json:"effort"`
	EvaluatorProfile string          `json:"evaluator_profile"`
	SprintContract  bool             `json:"sprint_contract"`
	PlanAudit       bool             `json:"plan_audit"`
}

// newHarnessRouterCmd는 `moai harness` 부모 커맨드 팩토리입니다.
// REQ-HRN-001-006: route + validate 서브커맨드를 등록합니다.
//
// CRITICAL: 이 팩토리는 내부 cli/harness.go:47의 newHarnessCmd()와
// 완전히 별개입니다. TestHarnessRetirement CI 가드는 newHarnessCmd()가
// root에 등록되지 않음을 검증합니다. newHarnessRouterCmd()는 별도의
// routing/validate 동사를 제공하는 새 엔트리입니다.
// @MX:ANCHOR: [AUTO] HRN-001 harness 커맨드 팩토리 (route + validate 동사)
// @MX:REASON: fan_in >= 3: root.go 등록, TestHarnessRouterCmd 테스트, CI integration에서 사용
func newHarnessRouterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harness",
		Short: "Harness routing and validation commands (HRN-001)",
		Long: `Harness routing and validation for SPEC complexity estimation.

Commands:
  route     Route a SPEC to minimal/standard/thorough harness level
  validate  Validate harness.yaml against schema and invariants

Note: This command provides routing/validation verbs (route, validate).
The legacy harness lifecycle verbs (status/apply/rollback/disable) are
retired per SPEC-V3R4-HARNESS-001 and accessible only via /moai:harness
slash command.`,
	}
	cmd.AddCommand(newHarnessRouteCmd())
	cmd.AddCommand(newHarnessValidateCmd())
	return cmd
}

// newHarnessRouteCmd는 `moai harness route` 서브커맨드 팩토리입니다.
// REQ-HRN-001-006/011, AC-HRN-001-02/03/06/09.
func newHarnessRouteCmd() *cobra.Command {
	var (
		specID     string
		jsonOutput bool
		cfgPath    string
		baseDir    string
	)

	cmd := &cobra.Command{
		Use:   "route",
		Short: "Route a SPEC to a harness level",
		Long: `Route a SPEC to minimal, standard, or thorough harness level
based on Complexity Estimator signals (file_count, domain_count, keywords, priority).

Examples:
  moai harness route --spec SPEC-V3R2-ORC-001
  moai harness route --spec SPEC-V3R2-ORC-001 --json
  moai harness route --spec SPEC-V3R2-ORC-001 --path /custom/harness.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHarnessRoute(cmd, specID, jsonOutput, cfgPath, baseDir)
		},
	}

	cmd.Flags().StringVar(&specID, "spec", "", "SPEC ID to route (e.g., SPEC-V3R2-ORC-001)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output routing decision as JSON")
	cmd.Flags().StringVar(&cfgPath, "path", "", "Path to harness.yaml (default: "+defaultHarnessConfigPath+")")
	cmd.Flags().StringVar(&baseDir, "base-dir", "", "Base directory for .moai/specs/ lookup (default: current dir)")

	if err := cmd.MarkFlagRequired("spec"); err != nil {
		panic(fmt.Sprintf("harness route: MarkFlagRequired: %v", err))
	}

	return cmd
}

// runHarnessRoute는 `moai harness route` 커맨드를 실행합니다.
func runHarnessRoute(cmd *cobra.Command, specID string, jsonOutput bool, cfgPath string, baseDir string) error {
	// harness.yaml 경로 결정
	harnessPath := cfgPath
	if harnessPath == "" {
		harnessPath = defaultHarnessConfigPath
	}

	// harness.yaml 로드
	cfg, err := config.LoadHarnessConfig(harnessPath)
	if err != nil {
		return fmt.Errorf("harness route: load config: %w", err)
	}

	// SPEC 파일 경로 해석: SPEC-ID → .moai/specs/{SPEC-ID}/spec.md
	specPath, err := resolveSpecPath(specID, baseDir)
	if err != nil {
		return fmt.Errorf("harness route: resolve spec path: %w", err)
	}

	// 라우팅 수행
	r := router.New(cfg)
	level, rationale, err := r.RouteFromFile(specPath, cfg)
	if err != nil {
		return fmt.Errorf("harness route: routing failed: %w", err)
	}

	// 노력 수준 및 evaluator 프로필 결정
	effort := router.EffortForLevel(level, cfg)
	evaluatorProfile := cfg.DefaultProfile
	sprintContract := false
	planAudit := true

	if levelCfg, ok := cfg.Levels[string(level)]; ok {
		if levelCfg.EvaluatorProfile != "" {
			evaluatorProfile = levelCfg.EvaluatorProfile
		}
		sprintContract = levelCfg.SprintContract
		planAudit = levelCfg.PlanAudit.Enabled
	}

	// 출력 포맷
	if jsonOutput {
		output := harnessRouteJSONOutput{
			Level:            string(level),
			Rationale:        rationale,
			Effort:           effort,
			EvaluatorProfile: evaluatorProfile,
			SprintContract:   sprintContract,
			PlanAudit:        planAudit,
		}
		data, err := json.Marshal(output)
		if err != nil {
			return fmt.Errorf("harness route: json marshal: %w", err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	} else {
		// plaintext 출력
		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "SPEC: %s\n", specID)
		_, _ = fmt.Fprintf(w, "Level: %s\n", level)
		_, _ = fmt.Fprintf(w, "Matched Rule: %s\n", rationale.MatchedRule)
		_, _ = fmt.Fprintf(w, "Effort: %s\n", effort)
		_, _ = fmt.Fprintf(w, "Evaluator Profile: %s\n", evaluatorProfile)
		_, _ = fmt.Fprintf(w, "Sprint Contract: %v\n", sprintContract)
		_, _ = fmt.Fprintf(w, "Plan Audit: %v\n", planAudit)
		if len(rationale.Keywords) > 0 {
			_, _ = fmt.Fprintf(w, "Matched Keywords: %v\n", rationale.Keywords)
		}
	}

	return nil
}

// resolveSpecPath는 SPEC-ID 문자열로부터 spec.md 파일 경로를 결정합니다.
// baseDir가 주어지면 그것을 기준으로 하고, 없으면 현재 작업 디렉토리를 기준으로 합니다.
func resolveSpecPath(specID string, baseDir string) (string, error) {
	// 기준 디렉토리 결정
	base := baseDir
	if base == "" {
		var err error
		base, err = os.Getwd()
		if err != nil {
			base = "."
		}
	}

	candidates := []string{
		filepath.Join(base, ".moai", "specs", specID, "spec.md"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("spec file not found for %q; tried: %v", specID, candidates)
}
