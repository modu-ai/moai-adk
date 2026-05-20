package cli

// @MX:NOTE: [AUTO] V3R5 harness CLI 통합 팩토리 — SPEC-V3R5-HARNESS-AUTONOMY-001 §6 + AC-HRA-009
// @MX:NOTE: [AUTO] newHarnessRouterCmd()는 V3R5에서 8개 lifecycle/proposal 동사를 추가로 통합합니다
// @MX:WARN: [AUTO] V3R5는 SPEC-V3R4-HARNESS-001의 CLI retirement를 supersede합니다
// @MX:REASON: plan.md §6.4 + AC-HRA-009 (`./moai harness --help | grep ... ≥6 matches`) 강제

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

// newHarnessRouterCmd는 `moai harness` 부모 커맨드 팩토리입니다 (V3R5 unified).
//
// ARCHITECTURE DECISION (Option A — merge into router):
// V3R5-HARNESS-AUTONOMY-001 §6.4 + AC-HRA-009는 `moai harness` 트리에서 다음 10개 동사가
// 모두 노출되어야 한다고 명시합니다:
//   - HRN-001 routing 동사: route, validate
//   - V3R5 lifecycle 동사 (un-retired): status, apply, rollback, disable
//   - V3R5 proposal-management 동사 (M4 신규): mute, mute-list, unmute, verify
//
// 이전 V3R4-HARNESS-001은 lifecycle 동사를 retirement했지만, V3R5는 명시적으로 이를
// supersede합니다. 본 팩토리는 단일 부모 커맨드 아래 10개 서브커맨드를 모두 등록하여
// AC-HRA-009 (`./moai harness --help | grep -E '(status|apply|rollback|disable|mute|verify)'`
// 최소 6개 매칭)를 충족합니다.
//
// 별도의 newHarnessCmd() (internal/cli/harness.go)는 SPEC-V3R4-HARNESS-001 §2.1의 deprecation
// marker 계약에 따라 보존되지만 root 트리에 등록되지 않습니다 (TestHarnessFactoryStillCompiles 참조).
// V3R5-supersedence 이후 TestHarnessRetirement는 lifecycle 동사 등록을 허용하도록 갱신되었습니다.
//
// @MX:ANCHOR: [AUTO] V3R5 harness 커맨드 팩토리 (route/validate + 8 lifecycle/proposal 동사)
// @MX:REASON: fan_in >= 4: root.go 등록, harness_route_test.go, harness_test.go, harness_mute_test.go, AC-HRA-009 verification
func newHarnessRouterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harness",
		Short: "Harness routing, validation, and learning subsystem management",
		Long: `Harness commands for SPEC complexity routing and learning subsystem management.

Routing verbs (SPEC-V3R2-HRN-001):
  route     Route a SPEC to minimal/standard/thorough harness level
  validate  Validate harness.yaml against schema and invariants

Lifecycle verbs (SPEC-V3R5-HARNESS-AUTONOMY-001 §6, un-retired):
  status    Show observation/tier/evolution summary
  apply     Manually trigger 5-Layer pipeline for queued proposal
  rollback  Revert applied evolution (snapshot restore)
  disable   Set learning.enabled: false

Proposal-management verbs (SPEC-V3R5-HARNESS-AUTONOMY-001 §6, new in M4):
  mute       Mute a proposal category (workflow.yaml)
  mute-list  Print current muted categories
  unmute     Remove a category from the mute list
  verify     Verify harness determinism (W4 placeholder)

Note: SPEC-V3R5-HARNESS-AUTONOMY-001 supersedes the lifecycle CLI retirement
that was previously declared by SPEC-V3R4-HARNESS-001. The unified Cobra tree
satisfies AC-HRA-009 (6+ verb surface).`,
	}

	// --project-root flag (shared by all lifecycle/proposal subcommands)
	cmd.PersistentFlags().String("project-root", "", "project root path (default: current directory)")

	// V3R2-HRN-001 routing verbs.
	cmd.AddCommand(newHarnessRouteCmd())
	cmd.AddCommand(newHarnessValidateCmd())

	// V3R5 lifecycle verbs (un-retired per plan.md §6.4).
	cmd.AddCommand(newHarnessStatusCmd())
	cmd.AddCommand(newHarnessApplyCmd())
	cmd.AddCommand(newHarnessRollbackCmd())
	cmd.AddCommand(newHarnessDisableCmd())

	// V3R5 M4 proposal-management verbs (REQ-HRA-033, REQ-HRA-036).
	cmd.AddCommand(newHarnessMuteCmd())
	cmd.AddCommand(newHarnessMuteListCmd())
	cmd.AddCommand(newHarnessUnmuteCmd())
	cmd.AddCommand(newHarnessVerifyCmd())

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
