package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// constitutionRegistryEnvKey는 registry 경로를 지정하는 환경 변수 이름이다.
const constitutionRegistryEnvKey = "MOAI_CONSTITUTION_REGISTRY"

// constitutionRegistryRelPath는 기본 registry 파일의 프로젝트 상대 경로이다.
const constitutionRegistryRelPath = ".claude/rules/moai/core/zone-registry.md"

// newConstitutionCmd는 `moai constitution` 루트 서브커맨드를 생성한다.
// research.go 패턴을 따른다.
func newConstitutionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "constitution",
		Short:   "Manage the zone registry (FROZEN/EVOLVABLE zone codification)",
		Long:    "Zone registry 조회 및 검증 커맨드. SPEC-V3R2-CON-001 구현.",
		GroupID: "tools",
	}
	cmd.AddCommand(newConstitutionListCmd())
	cmd.AddCommand(newConstitutionGuardCmd())
	cmd.AddCommand(newConstitutionAmendCmd())
	return cmd
}

// newConstitutionGuardCmd는 `moai constitution guard` 서브커맨드를 생성한다.
// --violations 플래그로 변경된 rule ID 목록을 받아 FROZEN zone 위반 여부를 반환한다.
// SPEC-V3R2-CON-001 AC-CON-001-003 구현.
func newConstitutionGuardCmd() *cobra.Command {
	var violationsFlag []string

	cmd := &cobra.Command{
		Use:   "guard",
		Short: "Check for FROZEN zone violations",
		Long:  "변경된 rule ID 목록을 받아 Frozen zone 위반 여부를 점검한다. CI 통합에 사용.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory 확인 오류: %w", err)
			}
			registryPath := resolveRegistryPath(cwd)
			return runConstitutionGuard(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, registryPath, violationsFlag)
		},
	}

	cmd.Flags().StringSliceVar(&violationsFlag, "violations", nil, "변경된 rule ID 목록 (쉼표 구분 또는 반복 플래그)")
	return cmd
}

// runConstitutionGuard는 변경된 rule ID 중 Frozen zone 위반을 탐지한다.
// violations: 변경된 rule ID 목록 (비어있으면 위반 없음으로 처리).
// 반환값: Frozen zone 위반 시 에러, 없으면 nil.
func runConstitutionGuard(w, wWarn io.Writer, projectDir, registryPath string, violations []string) error {
	reg, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		return fmt.Errorf("registry 로드 오류 %q: %w", registryPath, err)
	}

	// orphan 경고 출력 (stderr)
	for _, warn := range reg.Warnings {
		_, _ = fmt.Fprintf(wWarn, "경고: %s\n", warn)
	}

	// 변경된 ID 중 Frozen zone 위반 탐지
	var frozenViolations []string
	for _, id := range violations {
		rule, ok := reg.Get(id)
		if !ok {
			// registry에 없는 ID는 dangling ref - 경고만 출력
			_, _ = fmt.Fprintf(wWarn, "경고: dangling reference %q - registry에 없는 ID\n", id)
			continue
		}
		if rule.Zone == constitution.ZoneFrozen {
			frozenViolations = append(frozenViolations, id)
		}
	}

	if len(frozenViolations) > 0 {
		_, _ = fmt.Fprintf(w, "FROZEN zone 위반 탐지 (%d개): %s\n",
			len(frozenViolations), strings.Join(frozenViolations, ", "))
		return fmt.Errorf("FROZEN zone 위반: %s", strings.Join(frozenViolations, ", "))
	}

	_, _ = fmt.Fprintln(w, "constitution guard: OK - Frozen zone 위반 없음")
	return nil
}

// newConstitutionListCmd는 `moai constitution list` 서브커맨드를 생성한다.
func newConstitutionListCmd() *cobra.Command {
	var zoneFlag string
	var fileFlag string
	var formatFlag string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List zone registry entries",
		Long:  "zone registry 엔트리를 출력한다. --zone, --file, --format 플래그로 필터링 가능.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory 확인 오류: %w", err)
			}

			registryPath := resolveRegistryPath(cwd)

			var zoneFilter *constitution.Zone
			if zoneFlag != "" {
				z, parseErr := constitution.ParseZone(zoneFlag)
				if parseErr != nil {
					return fmt.Errorf("--zone 파싱 오류: %w", parseErr)
				}
				zoneFilter = &z
			}

			return runConstitutionList(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, registryPath, zoneFilter, fileFlag, formatFlag)
		},
	}

	cmd.Flags().StringVar(&zoneFlag, "zone", "", "Zone 필터 (frozen|evolvable)")
	cmd.Flags().StringVar(&fileFlag, "file", "", "파일 경로 필터 (부분 일치)")
	cmd.Flags().StringVar(&formatFlag, "format", "table", "출력 형식 (table|json)")

	return cmd
}

// resolveRegistryPath는 우선순위에 따라 registry 파일 경로를 결정한다.
// 우선순위: MOAI_CONSTITUTION_REGISTRY 환경변수 → CLAUDE_PROJECT_DIR 기준 경로 → cwd 기준 경로.
func resolveRegistryPath(cwd string) string {
	if envPath := os.Getenv(constitutionRegistryEnvKey); envPath != "" {
		return envPath
	}

	if projectDir := os.Getenv("CLAUDE_PROJECT_DIR"); projectDir != "" {
		return filepath.Join(projectDir, constitutionRegistryRelPath)
	}

	return filepath.Join(cwd, constitutionRegistryRelPath)
}

// runConstitutionList는 registry를 로드하고 w에 출력한다.
// 경고는 wWarn (stderr)에 출력하여 stdout 출력을 오염시키지 않는다.
// 테스트 친화적 순수 함수.
func runConstitutionList(w, wWarn io.Writer, projectDir, registryPath string, zoneFilter *constitution.Zone, fileFilter, format string) error {
	reg, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		return fmt.Errorf("registry 로드 오류 %q: %w", registryPath, err)
	}

	// 경고는 stderr(wWarn)에 출력
	for _, warn := range reg.Warnings {
		_, _ = fmt.Fprintf(wWarn, "경고: %s\n", warn)
	}

	// 필터 적용
	entries := reg.Entries
	if zoneFilter != nil {
		entries = reg.FilterByZone(*zoneFilter)
	}
	if fileFilter != "" {
		var filtered []constitution.Rule
		for _, e := range entries {
			if strings.Contains(e.File, fileFilter) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	switch format {
	case "json":
		return renderConstitutionJSON(w, entries)
	default:
		renderConstitutionTable(w, entries)
		return nil
	}
}

// constitutionJSONOutput은 JSON 형식 출력 구조체이다.
type constitutionJSONOutput struct {
	Entries []constitutionJSONEntry `json:"entries"`
}

// constitutionJSONEntry는 JSON 직렬화용 엔트리 구조체이다.
type constitutionJSONEntry struct {
	ID         string `json:"id"`
	Zone       string `json:"zone"`
	File       string `json:"file"`
	Anchor     string `json:"anchor"`
	Clause     string `json:"clause"`
	CanaryGate bool   `json:"canary_gate"`
}

// renderConstitutionJSON은 JSON 형식으로 엔트리를 출력한다.
func renderConstitutionJSON(w io.Writer, entries []constitution.Rule) error {
	jsonEntries := make([]constitutionJSONEntry, 0, len(entries))
	for _, e := range entries {
		jsonEntries = append(jsonEntries, constitutionJSONEntry{
			ID:         e.ID,
			Zone:       e.Zone.String(),
			File:       e.File,
			Anchor:     e.Anchor,
			Clause:     e.Clause,
			CanaryGate: e.CanaryGate,
		})
	}

	out := constitutionJSONOutput{Entries: jsonEntries}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 직렬화 오류: %w", err)
	}

	_, _ = fmt.Fprintln(w, string(data))
	return nil
}

// renderConstitutionTable은 table 형식으로 엔트리를 출력한다.
// Clause는 -v 옵션 없이는 40자로 잘린다.
func renderConstitutionTable(w io.Writer, entries []constitution.Rule) {
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(w, "엔트리 없음.")
		return
	}

	const idWidth = 18
	const zoneWidth = 10
	const fileWidth = 50
	const clauseWidth = 40

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s",
		idWidth, "ID",
		zoneWidth, "Zone",
		fileWidth, "File",
		clauseWidth, "Clause",
	)
	separator := strings.Repeat("-", idWidth+2+zoneWidth+2+fileWidth+2+clauseWidth)

	_, _ = fmt.Fprintln(w, header)
	_, _ = fmt.Fprintln(w, separator)

	for _, e := range entries {
		clause := e.Clause
		if len(clause) > clauseWidth {
			clause = clause[:clauseWidth-3] + "..."
		}
		fileStr := e.File
		if len(fileStr) > fileWidth {
			fileStr = "..." + fileStr[len(fileStr)-(fileWidth-3):]
		}

		line := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s",
			idWidth, e.ID,
			zoneWidth, e.Zone.String(),
			fileWidth, fileStr,
			clauseWidth, clause,
		)
		_, _ = fmt.Fprintln(w, line)
	}

	_, _ = fmt.Fprintf(w, "\n총 %d개 엔트리\n", len(entries))
}

// newConstitutionAmendCmd는 `moai constitution amend` 서브커맨드를 생성한다.
// SPEC-V3R2-CON-002 구현. 5-layer safety gate를 통한 constitutional amendment.
func newConstitutionAmendCmd() *cobra.Command {
	var (
		ruleIDFlag    string
		beforeFlag    string
		afterFlag     string
		evidenceFlag  string
		dryRunFlag    bool
		dryRunEnv     = os.Getenv("MOAI_CONSTITUTION_DRY_RUN") == "true"
	)

	cmd := &cobra.Command{
		Use:   "amend",
		Short: "Propose a constitutional amendment with 5-layer safety gate",
		Long: "Constitutional amendment proposal 실행. 5-layer safety gate (FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight)를 통과해야 적용됩니다.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("working directory 확인 오류: %w", err)
			}

			// 필수 플래그 검증
			if ruleIDFlag == "" {
				return fmt.Errorf("--rule 필수")
			}
			if beforeFlag == "" || afterFlag == "" {
				return fmt.Errorf("--before와 --after 필수")
			}

			// 환경변수 dry-run 우선
			dryRun := dryRunFlag || dryRunEnv

			return runConstitutionAmend(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, ruleIDFlag, beforeFlag, afterFlag, evidenceFlag, dryRun)
		},
	}

	cmd.Flags().StringVar(&ruleIDFlag, "rule", "", "Rule ID (CONST-V3R2-NNN) [필수]")
	cmd.Flags().StringVar(&beforeFlag, "before", "", "현재 clause 텍스트 [필수]")
	cmd.Flags().StringVar(&afterFlag, "after", "", "새로운 clause 텍스트 [필수]")
	cmd.Flags().StringVar(&evidenceFlag, "evidence", "", "Amendment justification (Frozen zone 필수)")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Dry-run 모드: 파일 수정 없이 시뮬레이션만")

	return cmd
}

// runConstitutionAmend는 constitutional amendment pipeline을 실행한다.
func runConstitutionAmend(w, wWarn io.Writer, projectDir, ruleID, before, after, evidence string, dryRun bool) error {
	// Registry 로드
	registryPath := resolveRegistryPath(projectDir)
	registry, err := constitution.LoadRegistry(registryPath, projectDir)
	if err != nil {
		return fmt.Errorf("registry 로드 오류: %w", err)
	}

	// 경고 출력
	for _, warn := range registry.Warnings {
		_, _ = fmt.Fprintf(wWarn, "경고: %s\n", warn)
	}

	// Rule 존재 확인
	rule, exists := registry.Get(ruleID)
	if !exists {
		return fmt.Errorf("rule %q을(를) 찾을 수 없음", ruleID)
	}

	// Before 검증 (현재 clause와 일치하는지 확인)
	if rule.Clause != before {
		return fmt.Errorf("clause 불일치: --before가 현재 clause와 다름\n현재: %s\n입력: %s", rule.Clause, before)
	}

	// Proposal 생성
	proposal := &constitution.AmendmentProposal{
		RuleID:   ruleID,
		Before:   before,
		After:    after,
		Evidence: evidence,
	}

	// Pipeline 실행
	pipeline := constitution.NewPipeline()
	log, err := pipeline.Execute(proposal, projectDir, dryRun)
	if err != nil {
		return fmt.Errorf("amendment 실패: %w", err)
	}

	// 결과 출력
	if dryRun {
		_, _ = fmt.Fprintln(w, "=== Dry-run Results ===")
		_, _ = fmt.Fprintf(w, "Rule ID: %s\n", log.RuleID)
		_, _ = fmt.Fprintf(w, "Zone: %s\n", log.ZoneAfter)
		_, _ = fmt.Fprintf(w, "Clause Before: %s\n", log.ClauseBefore)
		_, _ = fmt.Fprintf(w, "Clause After: %s\n", log.ClauseAfter)
		_, _ = fmt.Fprintf(w, "Canary Verdict: %s\n", log.CanaryVerdict)
		if len(log.Contradictions) > 0 {
			_, _ = fmt.Fprintln(w, "Contradictions:")
			for _, c := range log.Contradictions {
				_, _ = fmt.Fprintf(w, "  - %s\n", c)
			}
		}
		_, _ = fmt.Fprintln(w, "\nDry-run 성공: 파일이 수정되지 않았습니다.")
	} else {
		_, _ = fmt.Fprintf(w, "Amendment 성공: %s\n", log.ID)
		_, _ = fmt.Fprintf(w, "Rule %s가 업데이트되었습니다.\n", ruleID)
	}

	return nil
}
