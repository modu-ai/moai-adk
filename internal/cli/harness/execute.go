// Package harness — `moai harness execute` CLI surface (SPEC-HARNESS-APPLY-EXECUTE-001).
//
// 이 파일은 `Applier.Apply()`의 첫 프로덕션 caller인 opt-in execute verb를 정의한다.
// Self-Harness 로드맵 P2 "observer/gate activation" 1차: 이전까지 dormant였던
// regression-gate + outcome-capture 파이프라인을 처음으로 live 경로에 배선하여
// 첫 apply-outcome telemetry를 생성한다.
//
// 정직한 가치 framing (spec.md §A.2): 현재 harness write surface는 markdown-only
// FROZEN allowlist이므로 regression gate의 측정 delta는 사실상 항상 Δ=0(always-pass)
// 이다. 본 verb는 회귀를 "방지"하지 않는다. 실질 가치는 `Applier.Apply()`의 첫
// 프로덕션 caller가 되어 첫 apply-outcome telemetry(usage-log.jsonl의 apply_outcome
// line)를 생성하는 것 — 이 telemetry가 Phase 5 분석의 입력 substrate가 된다.
//
// HARD subagent boundary (REQ-AEX-016): 이 패키지의 어떤 소스 파일도 AskUserQuestion을
// 호출하지 않는다. 오케스트레이터가 C-HRA-008 경계에서 사용자 상호작용(L5 승인 포함)을
// 소유한다. 이 verb는 positional/flag 입력을 받고 구조화된 에러를 emit한다.
package harness

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/safety"
)

// canonical harness 경로 (project root 상대, REQ-AEX-009). harness.go의 동일 const와
// 의미가 일치하지만 패키지 경계가 다르므로 여기서 별도 선언한다 (cli 패키지의
// unexported const는 cross-package 참조 불가).
const (
	execProposalDirRel = ".moai/harness/proposals"
	execSnapshotBaseRel = ".moai/harness/learning-history/snapshots"
	execManifestRel     = ".moai/harness/learning-history/manifest.jsonl"
	execBaselineRel     = ".moai/harness/measurements-baseline.yaml"
	execUsageLogRel     = ".moai/harness/usage-log.jsonl"
	execViolationLogRel = ".moai/harness/frozen-guard-violations.jsonl"
	execRateLimitRel    = ".moai/harness/rate-limit-state.json"
)

// ExecuteOptions는 execute verb의 입력을 담는다.
type ExecuteOptions struct {
	// ID는 .moai/harness/proposals/<ID>.json에서 로드할 proposal 식별자다 (필수).
	ID string
	// ProjectRoot는 프로젝트 root 절대 경로다. 모든 harness 경로가 이 기준으로
	// resolve된다. 빈 값이면 호출부(RunExecute)가 현재 디렉터리로 fallback한다.
	ProjectRoot string
}

// executePaths는 execute verb가 project root로부터 resolve한 canonical harness
// 경로 집합이다 (REQ-AEX-009, design.md §B Wiring Recipe). 테스트가 production
// 경로 구성을 non-vacuous하게 관측할 수 있도록 노출된 seam이다.
type executePaths struct {
	proposalDir  string
	snapshotBase string
	manifestPath string
	baselinePath string
	usageLogPath string
	violationLog string
	rateLimitPath string
}

// resolveExecutePaths는 project root를 기준으로 7개 harness 경로를 canonical하게
// resolve한다 (AC-AEX-009). 절대경로 규칙은 호출부(RunExecute)가 filepath.Abs로
// root를 절대화한 뒤 이 함수를 호출하는 것으로 보장된다.
func resolveExecutePaths(root string) executePaths {
	return executePaths{
		proposalDir:   filepath.Join(root, execProposalDirRel),
		snapshotBase:  filepath.Join(root, execSnapshotBaseRel),
		manifestPath:  filepath.Join(root, execManifestRel),
		baselinePath:  filepath.Join(root, execBaselineRel),
		usageLogPath:  filepath.Join(root, execUsageLogRel),
		violationLog:  filepath.Join(root, execViolationLogRel),
		rateLimitPath: filepath.Join(root, execRateLimitRel),
	}
}

// buildExecutePipelineConfig는 execute verb가 safety.NewPipeline에 전달하는
// PipelineConfig 값을 구성한다 (REQ-AEX-005). AutoApply: true가 이 값에 박혀
// 있음을 테스트가 직접 관측하여 autoApply contract production 배선을 non-vacuous하게
// 검증한다 (AC-AEX-007). PipelineConfig 필드는 exported이므로 cross-package에서
// 값 확인이 가능하다.
//
// @MX:NOTE: [AUTO] AutoApply: true는 in-memory PipelineConfig 값일 뿐 harness.yaml
// 디스크 값(auto_apply: false)을 변경하지 않는다 (spec.md §B.2 FROZEN 불변식, C1).
func buildExecutePipelineConfig(paths executePaths) safety.PipelineConfig {
	return safety.PipelineConfig{
		ViolationLogPath: paths.violationLog,
		RateLimitPath:    paths.rateLimitPath,
		AutoApply:        true,
	}
}

// RunExecute는 execute verb의 프로덕션 진입점이다 — `Applier.Apply()`의 첫 프로덕션
// caller. production safety Pipeline(AutoApply=true) + regression-gate Applier +
// outcome Observer를 구성한 뒤 injectable한 runExecuteWith로 위임한다 (design.md §F.1
// 테스트 seam: production 구성은 여기, 테스트는 stub evaluator/applier 주입).
//
// @MX:NOTE: [AUTO] AutoApply=true는 L5 재pending 회피 의도 — 오케스트레이터가 C-HRA-008
// 경계에서 이미 사용자 L5 승인을 획득한 상태에서만 이 verb를 호출한다. L1~L4는 여전히
// 강제되고 L5만 CLI 레벨에서 auto-approve된다 (spec.md §B.1 autoApply contract).
// @MX:WARN: [AUTO] AutoApply=true는 in-memory PipelineConfig 전용이다 — harness.yaml의
// 디스크 auto_apply: false를 절대 mutate하지 않는다.
// @MX:REASON: [AUTO] 디스크 값을 true로 바꾸면 이후 모든 harness 동작이 사람 승인 없이
// auto-apply되어 시스템 default 안전 정책이 무너진다. in-memory override는 단 1회
// 호출에만 국한되어 default를 보존한다 (spec.md §B.2 FROZEN 불변식, C1).
func RunExecute(opts ExecuteOptions) error {
	root := opts.ProjectRoot
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("harness execute: resolve project root: %w", err)
		}
		root = wd
	} else {
		// 절대경로 규칙 (internal/cli/CLAUDE.md): filepath.Abs로 절대화한다.
		// filepath.Join(cwd, userPath) 금지.
		abs, err := filepath.Abs(root)
		if err != nil {
			return fmt.Errorf("harness execute: resolve project root: %w", err)
		}
		root = abs
	}
	normalized := opts
	normalized.ProjectRoot = root

	// canonical 경로 resolve (REQ-AEX-009, design.md §B Wiring Recipe).
	paths := resolveExecutePaths(root)

	// autoApply contract: AutoApply=true (in-memory ONLY — harness.yaml 디스크 불변).
	// L1~L4 강제, L5 auto-approve (REQ-AEX-005).
	pipeline := safety.NewPipeline(buildExecutePipelineConfig(paths))

	// regression gate + outcome observer 배선 (REQ-AEX-008).
	applier := harness.NewApplierWithRegressionGate(paths.manifestPath, paths.baselinePath).
		WithOutcomeObserver(harness.NewObserver(paths.usageLogPath))

	return runExecuteWithBase(normalized, pipeline, applier, paths.snapshotBase)
}

// runExecuteWith는 design.md §F.1 T2 테스트 seam이다 — production이 아닌 stub
// evaluator/applier를 주입할 수 있도록 RunExecute의 핵심 로직을 분해한 내부 함수.
// snapshotBase는 ProjectRoot로부터 canonical하게 derive하므로, 테스트는 ProjectRoot만
// 지정하면 된다. applier.go/pipeline.go/FROZEN 파일을 일절 수정하지 않는다 (C2/C3 보존).
func runExecuteWith(opts ExecuteOptions, evaluator harness.SafetyEvaluator, applier *harness.Applier) error {
	root := opts.ProjectRoot
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("harness execute: resolve project root: %w", err)
		}
		root = wd
	}
	snapshotBase := filepath.Join(root, execSnapshotBaseRel)
	normalized := opts
	normalized.ProjectRoot = root
	return runExecuteWithBase(normalized, evaluator, applier, snapshotBase)
}

// runExecuteWithBase는 proposal 로드 → Apply 호출 → error→exit 분류 흐름의 공통
// 코어다. evaluator + applier + snapshotBase를 모두 명시적으로 받는다.
//
// L2 Canary용 sessions는 nil/empty로 전달한다 (REQ-AEX-011): first execute run에는
// recent-session metrics가 없으며, baselineScore([])=0 + defaultProjectedScorer가
// baseline+0.02를 반환하여 drop=0이므로 L2가 reject하지 않는다 (canary.go nil-safe).
func runExecuteWithBase(opts ExecuteOptions, evaluator harness.SafetyEvaluator, applier *harness.Applier, snapshotBase string) error {
	proposalPath, err := resolveProposalPath(opts.ProjectRoot, opts.ID)
	if err != nil {
		return err
	}

	proposal, err := loadProposalByID(proposalPath)
	if err != nil {
		return err
	}

	var sessions []harness.Session // nil — first run, REQ-AEX-011 (canary nil-safe)

	return applier.Apply(proposal, evaluator, snapshotBase, sessions)
}

// resolveProposalPath는 proposal ID를 .moai/harness/proposals/<id>.json 경로로
// 변환한다. 경로 traversal(../)을 방지하기 위해 ID가 단순 base name인지 검증한다
// (EC-1, 절대경로 규칙).
func resolveProposalPath(root, id string) (string, error) {
	if id == "" {
		return "", &userError{msg: "harness execute: empty proposal ID"}
	}
	// 경로 traversal 방지: ID는 단순 식별자여야 한다 (디렉터리 구분자/.. 금지).
	if strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return "", &userError{msg: fmt.Sprintf("harness execute: invalid proposal ID %q (path traversal not allowed)", id)}
	}
	return filepath.Join(root, execProposalDirRel, id+".json"), nil
}

// loadProposalByID는 proposal JSON 파일을 harness.Proposal로 로드한다 (REQ-AEX-004).
// 파일 부재는 user error(exit 1, REQ-AEX-012)로 분류된다.
func loadProposalByID(path string) (harness.Proposal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return harness.Proposal{}, &userError{msg: fmt.Sprintf("harness execute: proposal not found: %s", filepath.Base(path))}
		}
		return harness.Proposal{}, &userError{msg: fmt.Sprintf("harness execute: read proposal %s: %v", filepath.Base(path), err)}
	}
	var prop harness.Proposal
	if err := json.Unmarshal(data, &prop); err != nil {
		return harness.Proposal{}, &userError{msg: fmt.Sprintf("harness execute: parse proposal %s: %v", filepath.Base(path), err)}
	}
	return prop, nil
}

// userError는 exit 1(user error)로 분류되는 에러를 표시하는 sentinel 타입이다.
// proposal 부재/파싱 실패/traversal 같은 입력 측 오류에 사용된다.
type userError struct {
	msg string
}

func (e *userError) Error() string { return e.msg }

// ExitCodeForError는 Apply 반환 에러를 exit code로 분류한다 (design.md §E).
//
// 분기 순서 (errors.As 타입 분기 — errjoin walk 가능):
//  1. nil                       → 0 (성공)
//  2. *ApplyPendingError        → 2 (INVARIANT VIOLATION: AutoApply=true 하 Pending)
//  3. *ApplyRegressionError     → 1 (gate rolled back — user-actionable)
//  4. *userError (입력 오류)     → 1 (proposal 부재/파싱/traversal)
//  5. rejection (L1~L4)         → 1 (string "rejected" match)
//  6. 기타 (measurement-exec 등) → 2 (system error)
//
// @MX:NOTE: [AUTO] *ApplyPendingError가 exit 2인 이유는 AutoApply=true 하에서는
// L5가 절대 Pending을 반환하지 않아야 하기 때문이다 (spec.md §B.1 autoApply contract).
// 발생 시 그것은 contract 위반(invariant violation)이므로 system error로 분류한다.
func ExitCodeForError(err error) int {
	if err == nil {
		return 0
	}

	// (2) *ApplyPendingError — AutoApply=true 하 발생은 invariant 위반 → exit 2.
	var pendingErr *harness.ApplyPendingError
	if errors.As(err, &pendingErr) {
		return 2
	}

	// (3) *ApplyRegressionError — gate rolled back → exit 1 (user-actionable).
	var regErr *harness.ApplyRegressionError
	if errors.As(err, &regErr) {
		return 1
	}

	// (4) *userError — 입력 측 오류 (proposal 부재/파싱/traversal) → exit 1.
	var uErr *userError
	if errors.As(err, &uErr) {
		return 1
	}

	// (5) L1~L4 rejection — Apply가 "rejected" 문자열 에러를 반환 → exit 1.
	if strings.Contains(err.Error(), "rejected") {
		return 1
	}

	// (6) 기타 (measurement-exec failure / system error) → exit 2.
	return 2
}

// diagnosticForError는 stderr에 출력할 진단 메시지를 분류 결과에 맞춰 구성한다
// (design.md §E 메시지 열).
func diagnosticForError(err error) string {
	var pendingErr *harness.ApplyPendingError
	if errors.As(err, &pendingErr) {
		return "INVARIANT VIOLATION: autoApply contract — Pending under AutoApply=true"
	}
	var regErr *harness.ApplyRegressionError
	if errors.As(err, &regErr) {
		return fmt.Sprintf("regression gate rolled back: regressed=%v", regErr.Regressed)
	}
	return err.Error()
}

// NewExecuteCmd는 `moai harness execute` cobra 팩토리다 (REQ-AEX-002).
//
// propose.go / install.go를 미러링하여 internal/cli/harness/ 디렉터리에 위치하므로
// C-HRA-008 boundary guard(TestPropose_NoAskUserQuestion)가 자동으로 스캔한다.
// newHarnessRouterCmd()(harness_route.go)에서 등록된다.
//
// 이 verb는 AskUserQuestion을 절대 호출하지 않는다 — 사용자 상호작용(L5 승인 포함)은
// 오케스트레이터가 C-HRA-008 경계에서 이미 처리한 상태에서 opt-in 호출된다.
func NewExecuteCmd() *cobra.Command {
	var (
		id          string
		projectRoot string
	)

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Apply a pending proposal through the Go pipeline (opt-in — first production Applier.Apply caller)",
		Long: `Apply a pending harness proposal through the Go safety pipeline + regression
gate + outcome observer (Applier.Apply()).

This is the opt-in Go execute path (Path G). The default skill-workflow Edit path
(Path S) remains unchanged — this verb is invoked ONLY when the orchestrator has
already obtained the user's L5 approval at the C-HRA-008 boundary.

HONEST FRAMING: for the current markdown-only FROZEN write surface the regression
gate's measured delta is Δ=0 (always-pass). This verb does NOT prevent regressions.
Its value is being the first production caller of Applier.Apply() — generating the
first apply-outcome telemetry line in usage-log.jsonl, the input substrate for
downstream Phase 5 analysis.

This subcommand never invokes AskUserQuestion. It takes flags (--id, --project-root)
and emits structured errors. User interaction is owned exclusively by the
orchestrator per the subagent boundary HARD contract.

Examples:
  moai harness execute --id SPEC-PROJ-001
  moai harness execute --id SPEC-X --project-root /path/to/proj`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := runExecuteCommand(cmd, id, projectRoot)
			if err != nil {
				// exit code 분류 + 진단 메시지는 runExecuteCommand가 이미 stderr로 emit.
				cmd.SilenceUsage = true
				cmd.SilenceErrors = true
				os.Exit(ExitCodeForError(err))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "",
		"Proposal ID to apply (loads .moai/harness/proposals/<id>.json) (required)")
	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"Project root path (default: current directory)")

	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(fmt.Sprintf("harness execute: MarkFlagRequired: %v", err))
	}

	return cmd
}

// runExecuteCommand는 NewExecuteCmd의 RunE 본문을 테스트 가능하게 분해한 함수다.
// os.Exit를 직접 호출하지 않고 에러를 반환하여(success → nil), RunE가 exit code
// 분류만 담당한다. 성공 시 telemetry 기록 알림을 stdout으로, 실패 시 진단 메시지를
// stderr로 emit한다 (exit code는 호출부 RunE가 ExitCodeForError로 결정).
func runExecuteCommand(cmd *cobra.Command, id, projectRoot string) error {
	// --project-root 미지정 시 부모(persistent flag)에서 상속 시도.
	if projectRoot == "" {
		if f := cmd.InheritedFlags().Lookup("project-root"); f != nil {
			projectRoot = f.Value.String()
		}
	}
	if err := RunExecute(ExecuteOptions{ID: id, ProjectRoot: projectRoot}); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness execute: %s\n", diagnosticForError(err))
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(),
		"harness execute: proposal %s applied; apply-outcome telemetry recorded to %s\n",
		id, execUsageLogRel)
	return nil
}
