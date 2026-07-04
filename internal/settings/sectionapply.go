package settings

// 이 파일은 M2b 확장 필드의 영속화 디스패처다 (SPEC-WEB-CONSOLE-011 M2b).
//
// @MX:WARN: [AUTO] ApplySchemaEdits는 10섹션 확장 필드를 디스크에 쓰는 영속화
// 경계다. seam 섹션은 WriteSectionViaSeam(yamlpatch — 주석/미모델링 키 보존),
// typed 섹션(git_strategy/llm/quality)은 config.NewConfigManager LoadRaw →
// per-field apply → SetSection → Save 경로만 사용한다.
// @MX:REASON: [AUTO] workflow.yaml 등 8개 seam 섹션에 typed re-marshal을 적용하면
// 주석과 미모델링 키가 파괴된다 (REQ-WC11-005/017, AP-1/AP-11). 반대로
// git-strategy는 완전 typed + dirty-flag Save가 요구 계약이다 (REQ-WC11-010,
// SPEC-GITSTRATEGY-SAVE-ISOLATION-001). 필드별 라우팅은 FieldDef.Persist.Kind가
// SSOT이며, 여기 없는 키(read-only: llm.mode/team_mode, db system 5키)는 어떤
// 경로로도 기록되지 않는다 (REQ-WC11-013/019).

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// ApplySchemaEdits는 제출된 확장 필드 값(FieldDef.Name → 문자열 값)을 영속화한다.
// 빈 문자열 값은 호출자(웹 파서)가 이미 걸러낸다(empty=preserve, EC-1) — 여기서
// 받는 모든 값은 명시 제출 값이다. 알 수 없는 필드명은 오류다.
func ApplySchemaEdits(projectRoot string, edits map[string]string) error {
	if len(edits) == 0 {
		return nil
	}

	// 결정적 순서로 처리 (맵 순회 비결정성 제거).
	names := make([]string, 0, len(edits))
	for name := range edits {
		names = append(names, name)
	}
	sort.Strings(names)

	seamEdits := map[string][]yamlpatch.KeyEdit{} // 섹션 파일 → edits
	typedEdits := make([]FieldDef, 0)
	typedValues := make([]string, 0)

	for _, name := range names {
		f, ok := Field(name)
		if !ok {
			return fmt.Errorf("settings: unknown schema field %q", name)
		}
		switch f.Persist.Kind {
		case PersistSeam:
			seamEdits[f.Persist.Section] = append(seamEdits[f.Persist.Section],
				yamlpatch.KeyEdit{Path: f.Persist.Path, Value: edits[name]})
		case PersistTypedSection:
			typedEdits = append(typedEdits, f)
			typedValues = append(typedValues, edits[name])
		default:
			return fmt.Errorf("settings: field %q is not a schema-section field (kind %s)", name, f.Persist.Kind)
		}
	}

	// typed 섹션: 단일 LoadRaw → 전 필드 적용 → 변경 섹션만 SetSection → 단일 Save.
	if len(typedEdits) > 0 {
		if err := applyTypedEdits(projectRoot, typedEdits, typedValues); err != nil {
			return err
		}
	}

	// seam 섹션: 파일별 단일 PatchFile (원자적 기록).
	seamSections := make([]string, 0, len(seamEdits))
	for sec := range seamEdits {
		seamSections = append(seamSections, sec)
	}
	sort.Strings(seamSections)
	for _, sec := range seamSections {
		if err := WriteSectionViaSeam(projectRoot, sec, seamEdits[sec]); err != nil {
			return err
		}
	}
	return nil
}

// applyTypedEdits는 typed 섹션 필드를 config 매니저 경로로 영속화한다.
// git_strategy 필드가 하나라도 있으면 SetSection("git_strategy")이 dirty-flag를
// 세워 Save가 git-strategy.yaml을 재기록한다 (그 외에는 기존 파일 byte 보존 —
// SPEC-GITSTRATEGY-SAVE-ISOLATION-001).
func applyTypedEdits(projectRoot string, fields []FieldDef, values []string) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("settings: load project config: %w", err)
	}

	touched := map[string]bool{}
	for i, f := range fields {
		switch f.Persist.Section {
		case "git_strategy":
			if err := applyGitStrategyKey(&cfg.GitStrategy, f.Persist.Key, values[i]); err != nil {
				return err
			}
		case "llm":
			if err := applyLLMKey(&cfg.LLM, f.Persist.Key, values[i]); err != nil {
				return err
			}
		case "quality":
			if err := applyQualityKey(&cfg.Quality, f.Persist.Key, values[i]); err != nil {
				return err
			}
		default:
			return fmt.Errorf("settings: no typed applier for section %q", f.Persist.Section)
		}
		touched[f.Persist.Section] = true
	}

	for _, section := range []string{"git_strategy", "llm", "quality"} {
		if !touched[section] {
			continue
		}
		var value any
		switch section {
		case "git_strategy":
			value = cfg.GitStrategy
		case "llm":
			value = cfg.LLM
		case "quality":
			value = cfg.Quality
		}
		if err := mgr.SetSection(section, value); err != nil {
			return fmt.Errorf("settings: set %s section: %w", section, err)
		}
	}
	if err := mgr.Save(); err != nil {
		return fmt.Errorf("settings: save project config: %w", err)
	}
	return nil
}

// parseBoolValue / parseIntValue / parseFloatValue는 typed applier의 값 변환
// 가드다 (웹 파서가 1차 검증하지만 seam 없이 직접 호출되는 경우를 방어).
func parseBoolValue(key, v string) (bool, error) {
	switch v {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fmt.Errorf("settings: %s: %q is not a boolean", key, v)
}

func parseIntValue(key, v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("settings: %s: %q is not an integer", key, v)
	}
	return n, nil
}

// applyGitStrategyKey는 git_strategy.<key> 편집을 typed struct에 적용한다
// (REQ-WC11-010 — 전 필드 typed, Save dirty-flag 경로).
func applyGitStrategyKey(gs *config.GitStrategyConfig, key, v string) error {
	switch key {
	case "mode":
		gs.Mode = v
		return nil
	case "provider":
		gs.Provider = v
		return nil
	case "github_username":
		gs.GitHubUsername = v
		return nil
	case "gitlab.instance_url":
		gs.GitLab.InstanceURL = v
		return nil
	}

	profileName, rest, ok := strings.Cut(key, ".")
	if !ok {
		return fmt.Errorf("settings: unknown git_strategy key %q", key)
	}
	var p *config.ModeProfile
	switch profileName {
	case "manual":
		p = &gs.Manual
	case "personal":
		p = &gs.Personal
	case "team":
		p = &gs.Team
	default:
		return fmt.Errorf("settings: unknown git_strategy profile %q", profileName)
	}

	switch rest {
	case "workflow":
		p.Workflow = v
	case "environment":
		p.Environment = v
	case "auto_checkpoint":
		p.AutoCheckpoint = v
	case "branch_prefix":
		p.BranchPrefix = v
	case "main_branch":
		p.MainBranch = v
	case "commit_style.format":
		p.CommitStyle.Format = v
	case "hooks.pre_commit":
		p.Hooks.PreCommit = v
	case "hooks.pre_push":
		p.Hooks.PrePush = v
	case "hooks.commit_msg":
		p.Hooks.CommitMsg = v
	case "required_reviews":
		n, err := parseIntValue(key, v)
		if err != nil {
			return err
		}
		p.RequiredReviews = n
	default:
		b, err := parseBoolValue(key, v)
		if err != nil {
			return fmt.Errorf("settings: unknown git_strategy key %q", key)
		}
		switch rest {
		case "github_integration":
			p.GitHubIntegration = b
		case "push_to_remote":
			p.PushToRemote = b
		case "draft_pr":
			p.DraftPR = b
		case "branch_protection":
			p.BranchProtection = b
		case "automation.auto_branch":
			p.Automation.AutoBranch = b
		case "automation.auto_commit":
			p.Automation.AutoCommit = b
		case "automation.auto_pr":
			p.Automation.AutoPR = b
		case "automation.auto_push":
			p.Automation.AutoPush = b
		case "branch_creation.auto_enabled":
			p.BranchCreation.AutoEnabled = b
		case "branch_creation.prompt_always":
			p.BranchCreation.PromptAlways = b
		case "commit_style.scope_required":
			p.CommitStyle.ScopeRequired = b
		default:
			return fmt.Errorf("settings: unknown git_strategy key %q", key)
		}
	}
	return nil
}

// applyLLMKey는 llm.<key> 편집을 typed struct에 적용한다 (REQ-WC11-012 안전 키만;
// mode/team_mode는 read-only — 스키마에 편집 필드가 없어 여기 도달 불가하며,
// 도달 시 명시적으로 거부한다, REQ-WC11-013).
func applyLLMKey(l *config.LLMConfig, key, v string) error {
	switch key {
	case "mode", "team_mode":
		return fmt.Errorf("settings: llm.%s is read-only (runtime-managed, REQ-WC11-013)", key)
	case "performance_tier":
		l.PerformanceTier = v
	case "claude_models.high":
		l.ClaudeModels.High = v
	case "claude_models.medium":
		l.ClaudeModels.Medium = v
	case "claude_models.low":
		l.ClaudeModels.Low = v
	case "glm.models.high":
		l.GLM.Models.High = v
	case "glm.models.medium":
		l.GLM.Models.Medium = v
	case "glm.models.low":
		l.GLM.Models.Low = v
	case "glm.models.opus":
		l.GLM.Models.Opus = v
	case "glm.models.sonnet":
		l.GLM.Models.Sonnet = v
	case "glm.models.haiku":
		l.GLM.Models.Haiku = v
	default:
		return fmt.Errorf("settings: unknown llm key %q", key)
	}
	return nil
}

// applyQualityKey는 quality.<key> 확장 편집을 typed struct에 적용한다
// (REQ-WC11-011 — 기존 4필드 외 잔여 typed 키).
func applyQualityKey(q *models.QualityConfig, key, v string) error {
	switch key {
	case "coverage_threshold":
		n, err := parseIntValue(key, v)
		if err != nil {
			return err
		}
		q.CoverageThreshold = n
	case "coverage_exemptions.max_exempt_percentage":
		n, err := parseIntValue(key, v)
		if err != nil {
			return err
		}
		q.CoverageExemptions.MaxExemptPercentage = n
	case "ddd_settings.max_transformation_size":
		q.DDDSettings.MaxTransformationSize = v
	default:
		b, err := parseBoolValue(key, v)
		if err != nil {
			return fmt.Errorf("settings: unknown quality key %q", key)
		}
		switch key {
		case "ddd_settings.require_existing_tests":
			q.DDDSettings.RequireExistingTests = b
		case "ddd_settings.characterization_tests":
			q.DDDSettings.CharacterizationTests = b
		case "ddd_settings.behavior_snapshots":
			q.DDDSettings.BehaviorSnapshots = b
		case "ddd_settings.preserve_before_improve":
			q.DDDSettings.PreserveBeforeImprove = b
		case "tdd_settings.red_green_refactor":
			q.TDDSettings.RedGreenRefactor = b
		case "tdd_settings.test_first_required":
			q.TDDSettings.TestFirstRequired = b
		case "tdd_settings.mutation_testing_enabled":
			q.TDDSettings.MutationTestingEnabled = b
		case "coverage_exemptions.enabled":
			q.CoverageExemptions.Enabled = b
		case "coverage_exemptions.require_justification":
			q.CoverageExemptions.RequireJustification = b
		case "test_quality.specification_based":
			q.TestQuality.SpecificationBased = b
		case "test_quality.meaningful_assertions":
			q.TestQuality.MeaningfulAssertions = b
		case "test_quality.avoid_implementation_coupling":
			q.TestQuality.AvoidImplementationCoupling = b
		case "test_quality.mutation_testing_enabled":
			q.TestQuality.MutationTestingEnabled = b
		default:
			return fmt.Errorf("settings: unknown quality key %q", key)
		}
	}
	return nil
}
