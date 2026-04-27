// Package cli — /moai harness 서브커맨드.
// REQ-HL-009: status / apply / rollback <date> / disable 4개 verb 제공.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// harnessDefaultLogPath는 usage-log.jsonl 기본 경로 (projectRoot 기준 상대 경로).
const harnessDefaultLogPath = ".moai/harness/usage-log.jsonl"

// harnessDefaultSnapshotBase는 스냅샷 기본 디렉토리 (projectRoot 기준 상대 경로).
const harnessDefaultSnapshotBase = ".moai/harness/learning-history/snapshots"

// harnessDefaultProposalDir는 대기 중인 proposal 디렉토리 (projectRoot 기준 상대 경로).
const harnessDefaultProposalDir = ".moai/harness/proposals"

// harnessConfigPath는 harness.yaml 경로 (projectRoot 기준 상대 경로).
const harnessConfigPath = ".moai/config/sections/harness.yaml"

// newHarnessCmd는 /moai harness 루트 커맨드를 생성한다.
//
// @MX:ANCHOR: [AUTO] newHarnessCmd는 harness CLI 진입점이다.
// @MX:REASON: [AUTO] fan_in >= 3: harness_test.go, root.go init(), Phase 5 IT
func newHarnessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harness",
		Short: "Harness 학습 서브시스템 관리",
		Long: `Harness 학습 서브시스템을 관리하는 서브커맨드.

사용 가능한 verb:
  status    tier 분포, rate-limit 윈도우, 대기 중인 proposal 상태 출력
  apply     다음 대기 중인 proposal을 로드하고 orchestrator에게 payload 반환
  rollback  지정한 날짜의 스냅샷으로 파일 복원
  disable   learning.enabled: false 설정 (observer + learner 비활성화)`,
		GroupID: "tools",
	}

	// --project-root 플래그 (모든 subcommand 공유)
	cmd.PersistentFlags().String("project-root", "", "프로젝트 루트 경로 (기본: 현재 디렉토리)")

	// verb 등록
	cmd.AddCommand(newHarnessStatusCmd())
	cmd.AddCommand(newHarnessApplyCmd())
	cmd.AddCommand(newHarnessRollbackCmd())
	cmd.AddCommand(newHarnessDisableCmd())

	return cmd
}

// resolveProjectRoot는 --project-root 플래그 또는 현재 디렉토리를 반환한다.
func resolveProjectRoot(cmd *cobra.Command) (string, error) {
	root, _ := cmd.Flags().GetString("project-root")
	if root == "" {
		// 상속된 플래그(부모 커맨드의 --project-root) 탐색
		if f := cmd.InheritedFlags().Lookup("project-root"); f != nil {
			root = f.Value.String()
		}
	}
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("현재 디렉토리 확인 실패: %w", err)
		}
	}
	return root, nil
}

// ─────────────────────────────────────────────
// status verb (T-P4-02)
// ─────────────────────────────────────────────

func newHarnessStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "학습 서브시스템 상태 출력",
		Long: `tier 분포, 마지막 업데이트, rate-limit 윈도우,
대기 중인 proposal 수, observer 활성화 상태를 출력한다.`,
		RunE: runHarnessStatus,
	}
}

// runHarnessStatus는 status verb를 실행한다.
func runHarnessStatus(cmd *cobra.Command, _ []string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	// harness.yaml 읽기 (enabled 상태 확인)
	cfg, err := loadHarnessYAML(filepath.Join(root, harnessConfigPath))
	if err != nil {
		// 파일이 없어도 기본 상태 출력
		cfg = defaultLearningConfig()
	}

	// usage-log.jsonl에서 패턴 집계
	logPath := filepath.Join(root, harnessDefaultLogPath)
	patterns, _ := harness.AggregatePatterns(logPath)

	thresholds := cfg.TierThresholds
	if len(thresholds) == 0 {
		thresholds = []int{1, 3, 5, 10}
	}

	// Tier 분포 계산
	tierCounts := make(map[string]int)
	tierCounts["observation"] = 0
	tierCounts["heuristic"] = 0
	tierCounts["rule"] = 0
	tierCounts["auto_update"] = 0

	for _, p := range patterns {
		t := harness.ClassifyTier(p, thresholds)
		tierCounts[t.String()]++
	}

	// 대기 중인 proposal 수 계산
	proposalDir := filepath.Join(root, harnessDefaultProposalDir)
	pendingCount := countProposals(proposalDir)

	// 출력 (errcheck: fmt.Fprintf 반환값 무시는 CLI 출력에서 관례적으로 허용)
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "=== Harness 학습 서브시스템 상태 ===\n\n")
	_, _ = fmt.Fprintf(out, "학습 활성화 (enabled): %v\n", cfg.Enabled)
	_, _ = fmt.Fprintf(out, "자동 적용 (auto_apply): %v\n", cfg.AutoApply)
	_, _ = fmt.Fprintf(out, "로그 보존 기간: %d일\n", cfg.LogRetentionDays)
	_, _ = fmt.Fprintf(out, "\n--- Tier 분포 (총 %d 패턴) ---\n", len(patterns))
	_, _ = fmt.Fprintf(out, "  observation : %d\n", tierCounts["observation"])
	_, _ = fmt.Fprintf(out, "  heuristic   : %d\n", tierCounts["heuristic"])
	_, _ = fmt.Fprintf(out, "  rule        : %d\n", tierCounts["rule"])
	_, _ = fmt.Fprintf(out, "  auto_update : %d\n", tierCounts["auto_update"])
	_, _ = fmt.Fprintf(out, "\n--- Rate Limit 설정 ---\n")
	_, _ = fmt.Fprintf(out, "  주간 최대 횟수: %d회\n", cfg.RateLimit.MaxPerWeek)
	_, _ = fmt.Fprintf(out, "  cooldown    : %d시간\n", cfg.RateLimit.CooldownHours)
	_, _ = fmt.Fprintf(out, "\n대기 중인 proposal: %d건\n", pendingCount)

	return nil
}

// countProposals는 proposalDir의 .json 파일 수를 반환한다.
func countProposals(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			count++
		}
	}
	return count
}

// ─────────────────────────────────────────────
// apply verb (T-P4-03)
// ─────────────────────────────────────────────

func newHarnessApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "다음 대기 중인 proposal을 orchestrator에게 반환",
		Long: `대기 중인 proposal 중 가장 오래된 것을 로드하고
JSON payload를 stdout에 출력한다.

[HARD] 이 커맨드는 AskUserQuestion을 직접 호출하지 않는다.
orchestrator(moai-harness-learner skill)가 이 payload를 받아
사용자에게 AskUserQuestion으로 제시한다.`,
		RunE: runHarnessApply,
	}
}

// runHarnessApply는 apply verb를 실행한다.
// [HARD] Subagent boundary: payload만 반환하고 AskUserQuestion 미호출.
func runHarnessApply(cmd *cobra.Command, _ []string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	proposalDir := filepath.Join(root, harnessDefaultProposalDir)
	entries, err := os.ReadDir(proposalDir)
	if err != nil || len(entries) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "대기 중인 proposal 없음.")
		return nil
	}

	// 가장 오래된 proposal 선택 (파일명 정렬 기준)
	var oldest os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			oldest = e
			break
		}
	}
	if oldest == nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "대기 중인 proposal 없음.")
		return nil
	}

	// proposal 읽기
	propPath := filepath.Join(proposalDir, oldest.Name())
	data, err := os.ReadFile(propPath)
	if err != nil {
		return fmt.Errorf("apply: proposal 파일 읽기 실패: %w", err)
	}

	// JSON payload를 stdout에 출력 (orchestrator가 이 내용을 AskUserQuestion으로 사용자에게 제시)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "--- 다음 Proposal (orchestrator에게 반환) ---")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "---")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[HARD] 이 CLI는 직접 승인/거부를 묻지 않습니다.")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "moai-harness-learner skill이 이 payload로 AskUserQuestion을 호출합니다.")

	return nil
}

// ─────────────────────────────────────────────
// rollback verb (T-P4-04)
// ─────────────────────────────────────────────

func newHarnessRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <date>",
		Short: "지정한 날짜의 스냅샷으로 파일 복원",
		Long: `<date>는 스냅샷 디렉토리명이다 (예: 2026-04-27T00-00-00.000000000Z).
manifest.json을 읽어 원본 파일을 byte-identical하게 복원한다.

존재하지 않는 날짜를 지정하면 오류 메시지를 출력하고 exit 1로 종료한다.`,
		Args: cobra.ExactArgs(1),
		RunE: runHarnessRollback,
	}
}

// runHarnessRollback은 rollback verb를 실행한다.
func runHarnessRollback(cmd *cobra.Command, args []string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	date := args[0]
	snapshotDir := filepath.Join(root, harnessDefaultSnapshotBase, date)

	// 스냅샷 디렉토리 존재 여부 확인
	if _, statErr := os.Stat(snapshotDir); os.IsNotExist(statErr) {
		return fmt.Errorf("rollback: 스냅샷을 찾을 수 없음 (날짜: %s). 'moai harness status'로 사용 가능한 스냅샷을 확인하세요", date)
	}

	// RestoreSnapshot 호출 (harness.RestoreSnapshot)
	if err := harness.RestoreSnapshot(snapshotDir); err != nil {
		return fmt.Errorf("rollback: 복원 실패: %w", err)
	}

	// rollback 이벤트 로그 기록 (Observer를 통해 기록)
	logPath := filepath.Join(root, harnessDefaultLogPath)
	obs := harness.NewObserver(logPath)
	_ = obs.RecordEvent(harness.EventTypeFeedback, "harness rollback "+date, "")

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "rollback 완료: %s 스냅샷으로 복원되었습니다.\n", date)
	return nil
}

// ─────────────────────────────────────────────
// disable verb (T-P4-05)
// ─────────────────────────────────────────────

func newHarnessDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "학습 서브시스템 비활성화 (learning.enabled: false)",
		Long: `harness.yaml의 learning.enabled 키를 false로 설정한다.
YAML round-trip을 사용하여 주석과 키 순서를 보존한다.

비활성화 후에는 observer와 learner가 no-op으로 동작한다.
재활성화하려면 harness.yaml에서 learning.enabled: true로 변경하세요.`,
		RunE: runHarnessDisable,
	}
}

// runHarnessDisable은 disable verb를 실행한다.
// [HARD] YAML round-trip — 주석과 키 순서 보존.
func runHarnessDisable(cmd *cobra.Command, _ []string) error {
	root, err := resolveProjectRoot(cmd)
	if err != nil {
		return err
	}

	configPath := filepath.Join(root, harnessConfigPath)

	// YAML 읽기 (yaml.v3 Node API로 주석 보존)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("disable: harness.yaml 읽기 실패: %w", err)
	}

	// Node API로 파싱
	var root2 yaml.Node
	if err := yaml.Unmarshal(data, &root2); err != nil {
		return fmt.Errorf("disable: harness.yaml 파싱 실패: %w", err)
	}

	// learning.enabled 노드를 false로 변경
	if err := setYAMLNodeValue(&root2, []string{"learning", "enabled"}, "false"); err != nil {
		return fmt.Errorf("disable: learning.enabled 수정 실패: %w", err)
	}

	// 직렬화
	newData, err := yaml.Marshal(&root2)
	if err != nil {
		return fmt.Errorf("disable: YAML 직렬화 실패: %w", err)
	}

	if err := os.WriteFile(configPath, newData, 0o644); err != nil {
		return fmt.Errorf("disable: harness.yaml 쓰기 실패: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "학습 서브시스템 비활성화 완료. (learning.enabled: false)\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "재활성화: harness.yaml에서 learning.enabled: true로 변경하세요.\n")
	return nil
}

// setYAMLNodeValue는 yaml.v3 Node 트리에서 keyPath에 해당하는 scalar 값을 value로 설정한다.
// 주석과 키 순서를 보존한다.
func setYAMLNodeValue(node *yaml.Node, keyPath []string, value string) error {
	if len(keyPath) == 0 {
		return nil
	}

	// DocumentNode 처리
	target := node
	if target.Kind == yaml.DocumentNode && len(target.Content) > 0 {
		target = target.Content[0]
	}

	if target.Kind != yaml.MappingNode {
		return fmt.Errorf("YAML 노드가 MappingNode가 아님: kind=%d", target.Kind)
	}

	// key 탐색 (MappingNode.Content는 [key, value, key, value, ...] 쌍)
	for i := 0; i+1 < len(target.Content); i += 2 {
		keyNode := target.Content[i]
		valueNode := target.Content[i+1]

		if keyNode.Value == keyPath[0] {
			if len(keyPath) == 1 {
				// 마지막 키 — 값 수정
				valueNode.Kind = yaml.ScalarNode
				valueNode.Tag = "!!bool"
				valueNode.Value = value
				return nil
			}
			// 더 깊이 탐색
			return setYAMLNodeValue(valueNode, keyPath[1:], value)
		}
	}

	return fmt.Errorf("키 '%s'를 찾을 수 없음", keyPath[0])
}

// ─────────────────────────────────────────────
// harness.yaml 로딩 헬퍼
// ─────────────────────────────────────────────

// learningConfig는 harness.yaml의 learning: 섹션 구조이다.
type learningConfig struct {
	Enabled          bool        `yaml:"enabled"`
	AutoApply        bool        `yaml:"auto_apply"`
	TierThresholds   []int       `yaml:"tier_thresholds"`
	RateLimit        rateLimitCfg `yaml:"rate_limit"`
	LogRetentionDays int         `yaml:"log_retention_days"`
}

// rateLimitCfg는 rate_limit 하위 설정이다.
type rateLimitCfg struct {
	MaxPerWeek    int `yaml:"max_per_week"`
	CooldownHours int `yaml:"cooldown_hours"`
}

// harnessYAMLRoot는 harness.yaml 전체 구조이다.
type harnessYAMLRoot struct {
	Learning learningConfig `yaml:"learning"`
}

// loadHarnessYAML은 harness.yaml을 읽어 learningConfig를 반환한다.
func loadHarnessYAML(path string) (learningConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return learningConfig{}, fmt.Errorf("loadHarnessYAML: 파일 읽기 실패 %s: %w", path, err)
	}

	var root harnessYAMLRoot
	if err := yaml.Unmarshal(data, &root); err != nil {
		return learningConfig{}, fmt.Errorf("loadHarnessYAML: 파싱 실패: %w", err)
	}

	return root.Learning, nil
}

// defaultLearningConfig는 기본 학습 설정을 반환한다.
func defaultLearningConfig() learningConfig {
	return learningConfig{
		Enabled:          true,
		AutoApply:        false,
		TierThresholds:   []int{1, 3, 5, 10},
		RateLimit:        rateLimitCfg{MaxPerWeek: 3, CooldownHours: 24},
		LogRetentionDays: 90,
	}
}

